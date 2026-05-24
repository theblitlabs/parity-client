package handlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/rs/zerolog"
	"github.com/theblitlabs/gologger"
	"github.com/theblitlabs/parity-client/internal/config"
	"github.com/theblitlabs/parity-client/internal/task"
	"github.com/theblitlabs/parity-client/internal/types"
)

type RequestRouter struct {
	config        *config.Config
	deviceID      string
	creatorAddr   string
	taskHandler   *TaskHandler
	proxy         *proxyHandler
	healthHandler *HealthHandler
	infoHandler   *InfoHandler
	dashboard     *DashboardHandler
	logger        zerolog.Logger
}

func NewRequestRouter(cfg *config.Config, deviceID, creatorAddr string) *RequestRouter {
	return &RequestRouter{
		config:        cfg,
		deviceID:      deviceID,
		creatorAddr:   creatorAddr,
		taskHandler:   NewTaskHandler(cfg, deviceID, creatorAddr),
		proxy:         newProxyHandler(cfg.Runner.ServerURL, deviceID, creatorAddr),
		healthHandler: NewHealthHandler(cfg),
		infoHandler:   NewInfoHandler(cfg, deviceID, creatorAddr),
		dashboard:     NewDashboardHandler(cfg),
		logger:        gologger.Get().With().Str("component", "router").Logger(),
	}
}

func (r *RequestRouter) HandleRequest(w http.ResponseWriter, req *http.Request) {
	r.logger.Debug().
		Str("original_path", req.URL.Path).
		Str("method", req.Method).
		Str("content_type", req.Header.Get("Content-Type")).
		Msg("Received request")

	path := strings.TrimPrefix(req.URL.Path, "/")
	path = strings.TrimPrefix(path, "api/")

	// Handle health check endpoints
	if r.handleHealthEndpoints(w, req, path) {
		return
	}

	if r.handleInfoEndpoints(w, req, path) {
		return
	}

	if r.handleDashboardEndpoints(w, req) {
		return
	}

	if r.shouldHandleLocalDockerUpload(req, path) {
		r.handleJSONRequest(w, req, path)
		return
	}

	if err := r.proxy.forwardRequest(w, req, path); err != nil {
		if writeErr := types.WriteError(w, http.StatusBadGateway, err.Error()); writeErr != nil {
			r.logger.Error().Err(writeErr).Msg("Failed to write error response")
		}
	}
}

func (r *RequestRouter) shouldHandleLocalDockerUpload(req *http.Request, path string) bool {
	if req.Method != http.MethodPost {
		return false
	}

	if !strings.Contains(req.Header.Get("Content-Type"), "application/json") {
		return false
	}

	normalizedPath := strings.Trim(path, "/")
	switch normalizedPath {
	case "tasks/local", "v1/tasks/local":
		return true
	case "tasks", "v1/tasks":
		return strings.EqualFold(req.Header.Get("X-Parity-Upload-Image"), "true")
	default:
		return false
	}
}

func (r *RequestRouter) handleJSONRequest(w http.ResponseWriter, req *http.Request, path string) {
	var taskRequest task.Request
	if err := types.ReadJSONBody(req.Body, &taskRequest); err != nil {
		r.logger.Error().Err(err).Msg("Failed to decode request body")
		if writeErr := types.WriteDetailedError(
			w,
			http.StatusBadRequest,
			"invalid JSON request body",
			"Send a JSON object with title, image, and command fields when using local image upload routes.",
			[]string{
				"Set Content-Type: application/json.",
				"Use /api/v1/tasks/local for local Docker image upload or /api/v1/tasks for normal proxying.",
			},
		); writeErr != nil {
			r.logger.Error().Err(writeErr).Msg("Failed to write error response")
		}
		return
	}

	if err := r.taskHandler.ValidateAndProcessTask(w, &taskRequest); err != nil {
		r.logger.Error().Err(err).Msg("Failed to process task")
		var taskErr *TaskProcessingError
		if errors.As(err, &taskErr) {
			if writeErr := types.WriteDetailedError(w, taskErr.StatusCode, taskErr.Message, taskErr.Hint, taskErr.NextSteps); writeErr != nil {
				r.logger.Error().Err(writeErr).Msg("Failed to write error response")
			}
			return
		}

		if writeErr := types.WriteError(w, http.StatusBadRequest, err.Error()); writeErr != nil {
			r.logger.Error().Err(writeErr).Msg("Failed to write error response")
		}
		return
	}
}

func (r *RequestRouter) handleInfoEndpoints(w http.ResponseWriter, req *http.Request, path string) bool {
	if req.Method != http.MethodGet {
		return false
	}

	normalizedRequestPath := strings.Trim(req.URL.Path, "/")
	switch normalizedRequestPath {
	case "", "api", "help", "api/help":
		r.infoHandler.HandleInfo(w, req)
		return true
	}

	switch strings.Trim(path, "/") {
	case "", "help":
		r.infoHandler.HandleInfo(w, req)
		return true
	}

	return false
}

func (r *RequestRouter) handleHealthEndpoints(w http.ResponseWriter, req *http.Request, path string) bool {
	switch path {
	case "health":
		if req.Method == "GET" {
			r.healthHandler.HandleHealthCheck(w, req)
			return true
		}
	case "health/detailed":
		if req.Method == "GET" {
			r.healthHandler.HandleDetailedHealthCheck(w, req)
			return true
		}
	case "health/ready":
		if req.Method == "GET" {
			r.healthHandler.HandleReadinessCheck(w, req)
			return true
		}
	case "health/live":
		if req.Method == "GET" {
			r.healthHandler.HandleLivenessCheck(w, req)
			return true
		}
	}
	return false
}

func (r *RequestRouter) handleDashboardEndpoints(w http.ResponseWriter, req *http.Request) bool {
	if req.Method != http.MethodGet {
		return false
	}

	switch strings.Trim(req.URL.Path, "/") {
	case "dashboard":
		r.dashboard.HandleDashboard(w, req)
		return true
	}

	return false
}
