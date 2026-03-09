package handlers

import (
	"fmt"
	"net/http"

	"github.com/theblitlabs/parity-client/internal/config"
	"github.com/theblitlabs/parity-client/internal/types"
	"github.com/theblitlabs/parity-client/internal/version"
)

type InfoHandler struct {
	config      *config.Config
	deviceID    string
	creatorAddr string
}

type ProxyRouteInfo struct {
	Method      string `json:"method"`
	Path        string `json:"path"`
	Description string `json:"description"`
}

type ProxyExample struct {
	Name    string `json:"name"`
	Command string `json:"command"`
}

type ProxyInfoResponse struct {
	Service        string           `json:"service"`
	Status         string           `json:"status"`
	Version        string           `json:"version"`
	LocalProxyURL  string           `json:"local_proxy_url"`
	UpstreamAPIURL string           `json:"upstream_api_url"`
	DeviceID       string           `json:"device_id,omitempty"`
	CreatorAddress string           `json:"creator_address,omitempty"`
	Routes         []ProxyRouteInfo `json:"routes"`
	Examples       []ProxyExample   `json:"examples"`
	ProxyBehavior  []string         `json:"proxy_behavior"`
}

func NewInfoHandler(cfg *config.Config, deviceID, creatorAddr string) *InfoHandler {
	return &InfoHandler{
		config:      cfg,
		deviceID:    deviceID,
		creatorAddr: creatorAddr,
	}
}

func (h *InfoHandler) HandleInfo(w http.ResponseWriter, r *http.Request) {
	localURL := localProxyBaseURL(h.config.Server.Host, h.config.Server.Port)
	upstreamURL := ""
	if h.config.Runner.ServerURL != "" {
		upstreamURL = fmt.Sprintf("%s/api/v1", normalizeServerBaseURL(h.config.Runner.ServerURL))
	}

	response := ProxyInfoResponse{
		Service:        "parity-client",
		Status:         "ready",
		Version:        version.GetShortVersion(),
		LocalProxyURL:  localURL,
		UpstreamAPIURL: upstreamURL,
		DeviceID:       h.deviceID,
		CreatorAddress: h.creatorAddr,
		Routes: []ProxyRouteInfo{
			{Method: http.MethodGet, Path: "/", Description: "Show local proxy routes, examples, and upstream target"},
			{Method: http.MethodGet, Path: "/health", Description: "Quick local health summary for the client proxy"},
			{Method: http.MethodPost, Path: "/api/v1/tasks", Description: "Proxy a normal registry-backed task submission to the server"},
			{Method: http.MethodPost, Path: "/api/v1/tasks/local", Description: "Package a local Docker image and upload it through the client"},
		},
		Examples: []ProxyExample{
			{
				Name: "Registry-backed task",
				Command: fmt.Sprintf(
					"curl -X POST %s/api/v1/tasks -H 'Content-Type: application/json' -d '{\"title\":\"Registry Task\",\"description\":\"Runner pulls from registry\",\"image\":\"python:3.11-alpine\",\"command\":[\"python\",\"-c\",\"print(42)\"]}'",
					localURL,
				),
			},
			{
				Name: "Local Docker image upload",
				Command: fmt.Sprintf(
					"curl -X POST %s/api/v1/tasks/local -H 'Content-Type: application/json' -d '{\"title\":\"Local Image Task\",\"description\":\"Client uploads a local image tar\",\"image\":\"my-local-image:latest\",\"command\":[\"/bin/sh\",\"-lc\",\"echo ready\"]}'",
					localURL,
				),
			},
		},
		ProxyBehavior: []string{
			"All other /api/v1/* requests are forwarded upstream unchanged.",
			"Use /api/v1/tasks/local only when the image already exists on this machine and needs to be uploaded.",
			"Use /api/v1/tasks for registry-backed tasks so runners can pull the image directly.",
			"POST /api/v1/tasks also supports X-Parity-Upload-Image: true if you want local upload behavior without changing the path.",
		},
	}

	w.Header().Set("X-Parity-Proxy-Mode", "local-info")
	if err := types.WriteJSON(w, http.StatusOK, response); err != nil {
		http.Error(w, "failed to encode proxy info", http.StatusInternalServerError)
	}
}
