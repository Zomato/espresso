package browser_manager

import (
	"context"

	log "github.com/Zomato/espresso/lib/logger"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

func isLocalScheme(s string) bool {
	switch s {
	case "data", "blob", "about", "chrome", "chrome-extension", "devtools", "file":
		return true
	}
	return false
}

// attachRequestFilter installs a request hijacker on the page that blocks any
// outbound request that fails the shared allowlist check in lib/common. The
// allowlist is read live on every request, so hot-reloading via
// common.SetAllowedDomains takes effect without re-initialising the pool.
func attachRequestFilter(page *rod.Page) {
	router := page.HijackRequests()
	err := router.Add("*", "", func(h *rod.Hijack) {
		u := h.Request.URL()
		scheme := u.Scheme

		if isLocalScheme(scheme) {
			h.ContinueRequest(&proto.FetchContinueRequest{})
			return
		}

		if scheme != "https" {
			log.Logger.Info(context.Background(), "blocked non-https scheme", map[string]any{"url": u.String(), "scheme": scheme})
			h.Response.Fail(proto.NetworkErrorReasonBlockedByClient)
			return
		}

		if ok, reason := IsURLAllowed(u.String()); !ok {
			log.Logger.Info(context.Background(), "blocked non-allowlisted url", map[string]any{"url": u.String(), "reason": reason})
			h.Response.Fail(proto.NetworkErrorReasonBlockedByClient)
			return
		}

		h.ContinueRequest(&proto.FetchContinueRequest{})
	})
	if err != nil {
		log.Logger.Error(context.Background(), "failed to register request hijack", err, nil)
		return
	}
	go router.Run()
}
