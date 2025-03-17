package browser_manager

import (
	"context"
	"fmt"
	"os"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
)

var (
	Browser *rod.Browser
)

func Init(ctx context.Context, tabPool int) error {
	fmt.Println("Initializing browser...")
	browserPath := os.Getenv("ROD_BROWSER_BIN")
	if browserPath == "" {
		return fmt.Errorf("ROD_BROWSER_BIN environment variable not set")
	}

	launcher := launcher.New().Bin(browserPath).
		Headless(true).
		Set("--disable-gpu").
		Set("--no-first-run").
		Set("--no-default-browser-check").
		Set("--disable-infobars").
		Set("--disable-dev-shm-usage").
		Set("--disable-accelerated-2d-canvas").
		Set("--disable-accelerated-video-decode").
		Set("--disable-background-networking").
		Set("--disable-background-timer-throttling").
		Set("--disable-translate").
		Set("--disable-sync").
		Set("--metrics-recording-only").
		Set("--mute-audio").
		Set("--user-data-dir", "/tmp/chrome-user-data").
		Set("--disable-web-security").
		Set("--no-startup-window").
		Set("--disable-renderer-backgrounding"). // Prevent background throttling
		Set("--force-fieldtrials", "SiteIsolationExtensions/Disable").
		Set("--disable-hyperlink-auditing").
		Set("--disable-site-isolation-trials").
		Set("--disable-host-resolver").
		Set("--dns-prefetch-disable").
		Set("--disable-logging").
		Set("--disable-breakpad").
		Set("--disable-devtools").
		Set("--disable-threaded-animation").
		Set("--disable-threaded-scrolling").
		Set("--disable-histogram-customizer").
		Set("--disable-notifications").
		Set("--disable-component-update").
		Set("--enable-low-end-device-mode").
		Set("--disable-partitioning").
		Set("--disable-backgrounding-occluded-windows").
		Set("--force-low-power-mode").
		Set("--disable-renderer-accessibility").
		Set("--disable-cache").
		Set("--disable-prompt-on-repost").
		Set("--disable-domain-reliability").
		Set("--disable-features", "NetworkService,OutOfBlinkCors,InterestGroupStorage,UserAgentClientHint").
		Set("--disable-extensions").
		Set("--disable-component-extensions-with-background-pages").
		Set("--blink-settings", "autoplayPolicy=document-user-activation-required").
		Set("--disable-blink-features", "AutomationControlled,BackgroundTimers,BackForwardCache,MediaStream").
		Set("--disable-software-rasterizer").
		Set("--disable-backgrounding-occluded-windows").
		Set("--disable-background-timer-throttling").
		Set("--disable-background-downloads")

	url, err := launcher.Launch()
	if err != nil {
		return fmt.Errorf("failed to launch browser: %v", err)
	}
	fmt.Printf("Browser launched at URL: %s\n", url)

	browser := rod.New().ControlURL(url)
	if err := browser.Connect(); err != nil {
		return fmt.Errorf("failed to connect to browser: %v", err)
	}
	Browser = browser

	fmt.Println("Browser connected successfully")

	InitializeTabManager(ctx, tabPool)

	return nil
}
