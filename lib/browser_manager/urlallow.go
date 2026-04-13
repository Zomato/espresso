package browser_manager

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"sync"
)

var subdomainRegex = regexp.MustCompile(`^[a-zA-Z0-9]+$`)

var (
	mu             sync.RWMutex
	allowedDomains []string
)

// SetAllowedDomains replaces the allowlist. Safe for concurrent use with
// IsURLAllowed so the embedder (typically the service) can push updates at
// runtime — e.g. from a config hot-reload callback.
func SetAllowedDomains(domains []string) {
	mu.Lock()
	defer mu.Unlock()
	allowedDomains = append(allowedDomains[:0:0], domains...)
}

func getAllowedDomains() []string {
	mu.RLock()
	defer mu.RUnlock()
	return allowedDomains
}

// IsURLAllowed reports whether the URL may be fetched during PDF generation.
// Rules:
//   - Only https:// URLs are eligible.
//   - The host must exactly equal an allowlisted entry, or be a single-label
//     alphanumeric subdomain of one (no dots or hyphens in the subdomain part).
func IsURLAllowed(urlStr string) (bool, string) {
	if !strings.HasPrefix(urlStr, "https://") {
		return false, "URL does not start with https://"
	}

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return false, fmt.Sprintf("invalid URL format: %v", err)
	}

	parsedHost := strings.ToLower(parsedURL.Host)
	for _, domain := range getAllowedDomains() {
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
