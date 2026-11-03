package renderer

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Zomato/espresso/lib/browser_manager"
	"github.com/Zomato/espresso/lib/workerpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var validPNG = func() []byte {
	var buf bytes.Buffer
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.White)
	if err := png.Encode(&buf, img); err != nil {
		panic("failed to encode test PNG fixture: " + err.Error())
	}
	return buf.Bytes()
}()

func initTestPool(t *testing.T, size int) {
	t.Helper()
	workerpool.Initialize(size, 5*time.Second)
	t.Cleanup(func() {
		workerpool.Pool().Release()
	})
}

func startTestServer(t *testing.T, handler http.Handler) *httptest.Server {
	t.Helper()
	srv := httptest.NewTLSServer(handler)
	t.Cleanup(srv.Close)
	return srv
}

func allowTestHost(t *testing.T, srvURL string) {
	t.Helper()
	u, err := url.Parse(srvURL)
	require.NoError(t, err)
	// IsURLAllowed compares against parsedURL.Host which includes the port,
	// so we must allowlist the host:port form.
	browser_manager.SetAllowedDomains([]string{u.Host})
	t.Cleanup(func() { browser_manager.SetAllowedDomains(nil) })
}

func swapHTTPClient(t *testing.T, client *http.Client) {
	t.Helper()
	orig := imageFetchClient
	imageFetchClient = client
	t.Cleanup(func() { imageFetchClient = orig })
}

func TestPrefetchImages_MultipleURLs(t *testing.T) {
	srv := startTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		w.Write(validPNG)
	}))

	initTestPool(t, 4)
	allowTestHost(t, srv.URL)

	client := srv.Client()
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	swapHTTPClient(t, client)

	const numImages = 20
	data := make(map[string]interface{})
	for i := 0; i < numImages; i++ {
		data[fmt.Sprintf("img_%d", i)] = fmt.Sprintf("%s/image_%d.png", srv.URL, i)
	}

	result := PrefetchImages(context.Background(), data)

	assert.Len(t, result, numImages)
	for i := 0; i < numImages; i++ {
		key := fmt.Sprintf("img_%d", i)
		val, ok := result[key].(string)
		require.True(t, ok, "key %s should be a string", key)
		assert.True(t, strings.HasPrefix(val, "data:image/png;base64,"),
			"key %s should be a data URI, got: %s", key, val[:min(len(val), 40)])
	}
}

func TestPrefetchImages_NestedMaps(t *testing.T) {
	srv := startTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		w.Write(validPNG)
	}))

	initTestPool(t, 4)
	allowTestHost(t, srv.URL)

	client := srv.Client()
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	swapHTTPClient(t, client)

	data := map[string]interface{}{
		"title": "test",
		"nested": map[string]interface{}{
			"logo":  fmt.Sprintf("%s/logo.png", srv.URL),
			"label": "some text",
			"deep": map[string]interface{}{
				"icon": fmt.Sprintf("%s/icon.png", srv.URL),
			},
		},
		"top_image": fmt.Sprintf("%s/top.png", srv.URL),
	}

	result := PrefetchImages(context.Background(), data)

	assert.Equal(t, "test", result["title"])

	nested, ok := result["nested"].(map[string]interface{})
	require.True(t, ok)
	logoVal, ok := nested["logo"].(string)
	require.True(t, ok, "nested[logo] should be a string")
	assert.True(t, strings.HasPrefix(logoVal, "data:image/png;base64,"))
	assert.Equal(t, "some text", nested["label"])

	deep, ok := nested["deep"].(map[string]interface{})
	require.True(t, ok)
	iconVal, ok := deep["icon"].(string)
	require.True(t, ok, "deep[icon] should be a string")
	assert.True(t, strings.HasPrefix(iconVal, "data:image/png;base64,"))

	topVal, ok := result["top_image"].(string)
	require.True(t, ok, "result[top_image] should be a string")
	assert.True(t, strings.HasPrefix(topVal, "data:image/png;base64,"))
}

func TestPrefetchImages_StringMapConversion(t *testing.T) {
	srv := startTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		w.Write(validPNG)
	}))

	initTestPool(t, 4)
	allowTestHost(t, srv.URL)

	client := srv.Client()
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	swapHTTPClient(t, client)

	data := map[string]interface{}{
		"images": map[string]string{
			"avatar":     fmt.Sprintf("%s/avatar.png", srv.URL),
			"background": fmt.Sprintf("%s/bg.png", srv.URL),
			"plain_text": "not a url",
		},
	}

	result := PrefetchImages(context.Background(), data)

	converted, ok := result["images"].(map[string]interface{})
	require.True(t, ok, "map[string]string should have been converted to map[string]interface{}")

	avatarVal, ok := converted["avatar"].(string)
	require.True(t, ok)
	assert.True(t, strings.HasPrefix(avatarVal, "data:image/png;base64,"))

	bgVal, ok := converted["background"].(string)
	require.True(t, ok)
	assert.True(t, strings.HasPrefix(bgVal, "data:image/png;base64,"))

	assert.Equal(t, "not a url", converted["plain_text"])
}

func TestPrefetchImages_DisallowedURL(t *testing.T) {
	initTestPool(t, 2)

	browser_manager.SetAllowedDomains([]string{"allowed.example.com"})
	t.Cleanup(func() { browser_manager.SetAllowedDomains(nil) })

	data := map[string]interface{}{
		"safe":     "not a url",
		"bad_img":  "https://evil.example.com/malicious.png",
		"bad_link": "https://evil.example.com/page",
	}

	result := PrefetchImages(context.Background(), data)

	assert.Equal(t, "not a url", result["safe"])
	_, hasBadImg := result["bad_img"]
	assert.False(t, hasBadImg, "disallowed image URL should be deleted from map")
	_, hasBadLink := result["bad_link"]
	assert.False(t, hasBadLink, "disallowed non-image URL should be deleted from map")
}

func TestPrefetchImages_EmptyMap(t *testing.T) {
	initTestPool(t, 2)
	browser_manager.SetAllowedDomains(nil)
	t.Cleanup(func() { browser_manager.SetAllowedDomains(nil) })

	data := map[string]interface{}{}
	result := PrefetchImages(context.Background(), data)
	assert.Empty(t, result)
}

func TestPrefetchImages_NoURLs(t *testing.T) {
	initTestPool(t, 2)
	browser_manager.SetAllowedDomains(nil)
	t.Cleanup(func() { browser_manager.SetAllowedDomains(nil) })

	data := map[string]interface{}{
		"title":  "hello",
		"count":  42,
		"nested": map[string]interface{}{"key": "value"},
	}
	result := PrefetchImages(context.Background(), data)
	assert.Equal(t, "hello", result["title"])
	nested, ok := result["nested"].(map[string]interface{})
	require.True(t, ok, "result[nested] should be a map[string]interface{}")
	assert.Equal(t, "value", nested["key"])
}

func TestPrefetchImages_ConcurrentSafety(t *testing.T) {
	var mu sync.Mutex
	var requestCount int

	srv := startTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		requestCount++
		mu.Unlock()
		time.Sleep(5 * time.Millisecond)
		w.Header().Set("Content-Type", "image/png")
		w.Write(validPNG)
	}))

	initTestPool(t, 8)
	allowTestHost(t, srv.URL)

	client := srv.Client()
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	swapHTTPClient(t, client)

	const numImages = 30
	data := make(map[string]interface{})
	for i := 0; i < numImages; i++ {
		data[fmt.Sprintf("img_%d", i)] = fmt.Sprintf("%s/image_%d.png", srv.URL, i)
	}

	result := PrefetchImages(context.Background(), data)

	assert.Len(t, result, numImages)
	for i := 0; i < numImages; i++ {
		key := fmt.Sprintf("img_%d", i)
		val, ok := result[key].(string)
		require.True(t, ok, "key %s missing or not a string", key)
		assert.True(t, strings.HasPrefix(val, "data:image/png;base64,"),
			"key %s should be a data URI", key)
	}

	mu.Lock()
	assert.Equal(t, numImages, requestCount, "all images should have been fetched")
	mu.Unlock()
}

func TestPrefetchImages_FetchError(t *testing.T) {
	srv := startTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))

	initTestPool(t, 4)
	allowTestHost(t, srv.URL)

	client := srv.Client()
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	swapHTTPClient(t, client)

	data := map[string]interface{}{
		"img1":  fmt.Sprintf("%s/fail.png", srv.URL),
		"title": "keep me",
	}

	result := PrefetchImages(context.Background(), data)

	assert.Equal(t, "keep me", result["title"])
	// When fetchImageAsDataURIFromURL returns an error (non-200),
	// the worker returns early without modifying the map entry.
	val, ok := result["img1"].(string)
	assert.True(t, ok)
	assert.True(t, strings.HasPrefix(val, "https://"),
		"failed fetch leaves the original URL unchanged")
}
