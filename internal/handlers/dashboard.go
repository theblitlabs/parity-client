package handlers

import (
	_ "embed"
	"fmt"
	"net/http"
	"strings"

	"github.com/theblitlabs/parity-client/internal/config"
)

//go:embed dashboard.html
var dashboardHTML string

type DashboardHandler struct {
	config *config.Config
}

func NewDashboardHandler(cfg *config.Config) *DashboardHandler {
	return &DashboardHandler{config: cfg}
}

func (h *DashboardHandler) HandleDashboard(w http.ResponseWriter, _ *http.Request) {
	localURL := localProxyBaseURL(h.config.Server.Host, h.config.Server.Port)
	page := strings.ReplaceAll(dashboardHTML, "__LOCAL_PROXY_URL__", localURL)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Parity-Proxy-Mode", "local-dashboard")
	_, _ = fmt.Fprint(w, page)
}
