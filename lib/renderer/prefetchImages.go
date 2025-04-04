package renderer

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	log "github.com/Zomato/espresso/lib/logger"
	"github.com/Zomato/espresso/lib/workerpool"
)

type stackItem struct {
	key  string
	data map[string]interface{}
}

// Prefetch images and replace their URLs with data URIs
func PrefetchImages(ctx context.Context, data map[string]interface{}) map[string]interface{} {

	startTime := time.Now()
	var wg sync.WaitGroup
	var mu sync.Mutex // to add lock on updating the json data

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
					var dataURI string
					var err error
					if strings.HasPrefix(v, "https://") {
						duration := time.Since(startTime)
						log.Logger.Info(ctx, "fetching image at", map[string]any{"name": v, "time": duration})

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
					log.Logger.Info(ctx, "fetched image data at", map[string]any{"time": duration, "image": v})

					mu.Lock()
					parentData[k] = dataURI
					mu.Unlock()
					log.Logger.Info(ctx, "replaced image data at", map[string]any{"time": duration, "key": k, "error": err})
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
	log.Logger.Info(ctx, "prefetching images completed at", map[string]any{"time": duration})

	wg.Wait()

	duration = time.Since(startTime)
	log.Logger.Info(ctx, "all worker pool tasks completed at", map[string]any{"time": duration})

	return data
}

// Fetch an image and convert it to a data URI
func fetchImageAsDataURIFromURL(url string) (string, error) {
	startTime := time.Now()
	ctx := context.Background()

	duration := time.Since(startTime)
	log.Logger.Info(context.Background(), "fetching image at", map[string]any{"time": duration, "url": url})

	resp, err := http.Get(url)
	if err != nil {
		log.Logger.Error(ctx, "failed to fetch image", err, nil)
		return "", fmt.Errorf("failed to fetch image: %v", err)
	}

	duration = time.Since(startTime)
	log.Logger.Info(context.Background(), "fetched image at", map[string]any{"time": duration, "url": url})

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Logger.Error(ctx, "failed to fetch image", nil, map[string]any{"status_code": resp.StatusCode})
		return "", fmt.Errorf("failed to fetch image, status code: %d", resp.StatusCode)
	}

	imageBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Logger.Error(ctx, "failed to read image bytes", err, nil)
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
	log.Logger.Info(context.Background(), "returning image at", map[string]any{"time": duration, "url": url})

	return dataURI, nil
}
