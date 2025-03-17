package browser_manager

import (
	"context"
	"fmt"
	"sync"

	"github.com/go-rod/rod"
)

// TabPool manages a pool of browser tabs using a channel.
type TabPool struct {
	pool      chan *rod.Page
	initOnce  sync.Once
	totalTabs int
}

var (
	tabManagerInstance *TabPool
	numTabs            int
)

func InitializeTabManager(ctx context.Context, tabPool int) {
	tabManagerInstance = NewTabPool(ctx, Browser, tabPool)
}

func NewTabPool(ctx context.Context, browser *rod.Browser, tabPool int) *TabPool {
	fmt.Println("Initializing tab pool with ", tabPool, " tabs")
	numTabs = tabPool
	if tabPool == 0 {
		return nil
	}

	pool := &TabPool{
		pool:      make(chan *rod.Page, tabPool),
		totalTabs: tabPool,
	}

	pool.initOnce.Do(func() {
		for i := 0; i < tabPool; i++ {
			page := browser.MustPage("about:blank")
			pool.pool <- page
		}
	})

	return pool
}

// if the `browser.tabs` is 0 then we are creating a new tab on each request
func GetTab() *rod.Page {
	fmt.Println("Getting tab")
	if numTabs == 0 {
		return Browser.MustPage("about:blank")
	}

	page := <-tabManagerInstance.pool
	return page
}

func ReleaseTab(page *rod.Page) {
	fmt.Println("Releasing tab")
	if numTabs == 0 {
		page.MustClose()
		return
	}
	err := page.SetDocumentContent("")
	if err != nil {
		fmt.Println("Error releasing tab", err)
		panic(err)
	}

	tabManagerInstance.pool <- page
}

func ClearAllBlobs(page *rod.Page, dynaminData map[string]interface{}) {
	for _, value := range dynaminData {
		strValue, ok := value.(string)
		if ok && strValue != "" {
			if err := removeBlobURL(page, strValue); err != nil {
				fmt.Println("Error clearing blob URL", err)
				panic(err)
			}
		}
	}
}

func removeBlobURL(page *rod.Page, blobURL string) error {
	_, err := page.Eval(`(blobURL) => {
		URL.revokeObjectURL(blobURL);
	}`, blobURL)

	if err != nil {
		return fmt.Errorf("failed to execute JavaScript for blob URL: %v", err)
	}

	return nil
}
