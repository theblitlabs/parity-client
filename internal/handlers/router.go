package handlers

import (
	"net/http"
	"strings"

	"github.com/rs/zerolog"
	"github.com/theblitlabs/gologger"
	"github.com/theblitlabs/parity-client/internal/config"
	"github.com/theblitlabs/parity-client/internal/task"
	"github.com/theblitlabs/parity-client/internal/types"
)

// RequestRouter handles routing of incoming requests
type RequestRouter struct {
	config      *config.Config
	deviceID    string
	creatorAddr string
	taskHandler *TaskHandler
	proxy       *proxyHandler
	logger      zerolog.Logger
}

// NewRequestRouter creates a new request router
func NewRequestRouter(cfg *config.Config, deviceID, creatorAddr string) *RequestRouter {
	return &RequestRouter{
		config:      cfg,
		deviceID:    deviceID,
		creatorAddr: creatorAddr,
		taskHandler: NewTaskHandler(cfg, deviceID, creatorAddr),
		proxy:       newProxyHandler(cfg.Runner.ServerURL, deviceID, creatorAddr),
		logger:      gologger.Get().With().Str("component", "router").Logger(),
	}
}

// HandleRequest handles incoming HTTP requests
func (r *RequestRouter) HandleRequest(w http.ResponseWriter, req *http.Request) {
	r.logger.Debug().
		Str("original_path", req.URL.Path).
		Str("method", req.Method).
		Str("content_type", req.Header.Get("Content-Type")).
		Msg("Received request")

	path := strings.TrimPrefix(req.URL.Path, "/")
	path = strings.TrimPrefix(path, "api/")

	if req.Method == "POST" && strings.Contains(req.Header.Get("Content-Type"), "application/json") {
		r.handleJSONRequest(w, req, path)
		return
	}

	if err := r.proxy.forwardRequest(w, req, path); err != nil {
		if writeErr := types.WriteError(w, http.StatusBadGateway, err.Error()); writeErr != nil {
			r.logger.Error().Err(writeErr).Msg("Failed to write error response")
		}
	}
}

func (r *RequestRouter) handleJSONRequest(w http.ResponseWriter, req *http.Request, path string) {
	var taskRequest task.Request
	if err := types.ReadJSONBody(req.Body, &taskRequest); err != nil {
		r.logger.Error().Err(err).Msg("Failed to decode request body")
		if writeErr := types.WriteError(w, http.StatusBadRequest, "Invalid request body"); writeErr != nil {
			r.logger.Error().Err(writeErr).Msg("Failed to write error response")
		}
		return
	}

	if err := r.taskHandler.ValidateAndProcessTask(w, &taskRequest); err != nil {
		r.logger.Error().Err(err).Msg("Failed to process task")
		if writeErr := types.WriteError(w, http.StatusBadRequest, err.Error()); writeErr != nil {
			r.logger.Error().Err(writeErr).Msg("Failed to write error response")
		}
		return
	}
}
