package renderer

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	log "github.com/Zomato/espresso/lib/logger"
	"github.com/Zomato/espresso/lib/workerpool"
)

var (
	imageExtURLRegex = regexp.MustCompile(`(?i)\.(png|jpe?g|gif|webp|bmp|svg|tiff?)(\?.*)?$`)
	subdomainRegex   = regexp.MustCompile(`^[a-zA-Z0-9]+$`)

	allowedDomainsMu sync.RWMutex
	allowedDomains   []string
)

// SetAllowedImageDomains configures the allowlist of hosts (and their
// single-label subdomains) that PrefetchImages is permitted to fetch from.
// Callers are expected to invoke this during startup
// using values sourced from their own config system.
func SetAllowedImageDomains(domains []string) {
	allowedDomainsMu.Lock()
	defer allowedDomainsMu.Unlock()
	allowedDomains = append(allowedDomains[:0:0], domains...)
}

func getAllowedImageDomains() []string {
	allowedDomainsMu.RLock()
	defer allowedDomainsMu.RUnlock()
	return allowedDomains
}

type stackItem struct {
	key  string
	data map[string]interface{}
}

// Prefetch images and replace their URLs with data URIs
func PrefetchImages(ctx context.Context, data map[string]interface{}) map[string]interface{} {

	startTime := time.Now()
	var wg sync.WaitGroup
	var mu sync.Mutex // to add lock on updating the json data

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	errChan := make(chan error, 1)

	stack := []stackItem{{key: "", data: data}}

	for len(stack) > 0 {
		current := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		for key, value := range current.data {
			strValue, ok := value.(string)
			if ok && (strings.HasPrefix(strValue, "https://")) {
				wg.Add(1)
				err := workerpool.Pool().SubmitTask(func(args ...interface{}) {
					k := args[0].(string)
					v := args[1].(string)
					parentData := args[2].(map[string]interface{})

					defer func() {
						wg.Done()
						if r := recover(); r != nil {
							err := fmt.Errorf("panic: %v and stacktrace %s", r, string(debug.Stack()))
							log.Logger.Info(ctx, "recovered from panic", map[string]any{"error": err})
						}
					}()

					allowed, reason := IsURLAllowed(ctx, v)
					if !allowed {
						if !imageExtURLRegex.MatchString(v) {
							log.Logger.Info(ctx, "URL not allowed and has non-image extension", map[string]any{"url": v, "reason": reason})
							return
						}
						err := fmt.Errorf("image URL not allowed: %s. Reason: %s", v, reason)
						log.Logger.Error(ctx, "error while prefetching images", err, map[string]any{"url": v})
						select {
						case errChan <- err:
							cancel()
						default:
						}
						return
					}

					var dataURI string
					var err error
					if strings.HasPrefix(v, "https://") {
						duration := time.Since(startTime)
						log.Logger.Info(ctx, "fetching image at", map[string]any{"name": v, "duration": duration})
						dataURI, err = fetchImageAsDataURIFromURL(v)
						if err != nil {
							log.Logger.Error(ctx, "failed to download image", err, map[string]any{"key": k})
							return
						}
					}

					if dataURI == "" {
						log.Logger.Error(ctx, "failed to download image. data uri is empty", nil, map[string]any{"key": k})
						mu.Lock()
						parentData[k] = ""
						mu.Unlock()
						return
					}

					duration := time.Since(startTime)
					log.Logger.Info(ctx, "fetched image data at", map[string]any{"duration": duration, "image": v})

					mu.Lock()
					parentData[k] = dataURI
					mu.Unlock()
					log.Logger.Info(ctx, "replaced image data at", map[string]any{"duration": duration, "key": k, "error": err})
				}, key, strValue, current.data)
				if err != nil {
					log.Logger.Error(ctx, "failed to submit task to worker pool", err, nil)
				}
			} else if nestedMap, ok := value.(map[string]interface{}); ok {
				stack = append(stack, stackItem{key: key, data: nestedMap})
			} else if stringMap, ok := value.(map[string]string); ok {
				interfaceMap := make(map[string]interface{})
				for k, v := range stringMap {
					interfaceMap[k] = v
				}

				current.data[key] = interfaceMap
				stack = append(stack, stackItem{key: key, data: interfaceMap})
			}
		}
	}

	duration := time.Since(startTime)
	log.Logger.Info(ctx, "prefetching images completed at", map[string]any{"duration": duration})

	wg.Wait()

	duration = time.Since(startTime)
	log.Logger.Info(ctx, "all worker pool tasks completed at", map[string]any{"duration": duration})

	return data
}

var imageFetchClient = &http.Client{
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	},
}

// Fetch an image and convert it to a data URI
func fetchImageAsDataURIFromURL(url string) (string, error) {
	startTime := time.Now()

	duration := time.Since(startTime)
	log.Logger.Info(context.Background(), "fetching image at", map[string]any{"duration": duration, "url": url})

	resp, err := imageFetchClient.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to fetch image: %v", err)
	}

	duration = time.Since(startTime)
	log.Logger.Info(context.Background(), "fetched image at", map[string]any{"duration": duration, "url": url})

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch image, status code: %d", resp.StatusCode)
	}

	imageBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read image bytes: %v", err)
	}

	// Determine the content type of the image
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = http.DetectContentType(imageBytes)
	}

	// Encode the image as a data URI
	dataURI := fmt.Sprintf("data:%s;base64,%s", contentType, base64.StdEncoding.EncodeToString(imageBytes))

	duration = time.Since(startTime)
	log.Logger.Info(context.Background(), "returning image at", map[string]any{"duration": duration, "url": url})
	return dataURI, nil
}

// IsURLAllowed checks whether a URL is permitted for prefetching based on
// the allowlist configured using renderer.SetAllowedImageDomains
func IsURLAllowed(ctx context.Context, urlStr string) (bool, string) {
	if !strings.HasPrefix(urlStr, "https://") {
		return false, "URL does not start with https://"
	}

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return false, fmt.Sprintf("invalid URL format: %v", err)
	}

	parsedHost := strings.ToLower(parsedURL.Host)
	allowedDomains := getAllowedImageDomains()
	for _, domain := range allowedDomains {
		domain = strings.ToLower(domain)
		if parsedHost == domain {
			return true, ""
		}

		if strings.HasSuffix(parsedHost, "."+domain) {
			subdomainPart := strings.TrimSuffix(parsedHost, "."+domain)
			if subdomainPart != "" &&
				!strings.Contains(subdomainPart, ".") &&
				!strings.Contains(subdomainPart, "-") &&
				subdomainRegex.MatchString(subdomainPart) {
				return true, ""
			}
		}
	}

	return false, fmt.Sprintf("domain not in whitelist: %s", parsedURL.Host)
}
