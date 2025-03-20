package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/theblitlabs/gologger"
	"github.com/theblitlabs/parity-client/internal/config"
	"github.com/theblitlabs/parity-client/internal/task"
)

// RequestRouter handles routing of incoming requests
type RequestRouter struct {
	config      *config.Config
	deviceID    string
	creatorAddr string
	taskHandler *TaskHandler
}

// NewRequestRouter creates a new request router
func NewRequestRouter(cfg *config.Config, deviceID, creatorAddr string) *RequestRouter {
	return &RequestRouter{
		config:      cfg,
		deviceID:    deviceID,
		creatorAddr: creatorAddr,
		taskHandler: NewTaskHandler(cfg, deviceID, creatorAddr),
	}
}

// HandleRequest handles incoming HTTP requests
func (r *RequestRouter) HandleRequest(w http.ResponseWriter, req *http.Request) {
	log := gologger.Get().With().Str("component", "router").Logger()

	log.Debug().
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

	r.proxyRequest(w, req, path)
}

func (r *RequestRouter) handleJSONRequest(w http.ResponseWriter, req *http.Request, path string) {
	log := gologger.Get().With().Str("component", "router").Logger()

	body, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}
	req.Body.Close()

	var taskRequest task.Request
	if err := json.Unmarshal(body, &taskRequest); err != nil {
		log.Error().Err(err).Msg("Failed to decode request body")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := r.taskHandler.ValidateAndProcessTask(w, &taskRequest); err != nil {
		log.Error().Err(err).Msg("Failed to process task")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func (r *RequestRouter) proxyRequest(w http.ResponseWriter, req *http.Request, path string) {
	logger := gologger.Get().With().Str("component", "router").Logger()
	targetURL := fmt.Sprintf("%s/%s", strings.TrimSuffix(r.config.Runner.ServerURL, "/"), path)

	proxyReq, err := http.NewRequest(req.Method, targetURL, req.Body)
	if err != nil {
		http.Error(w, "Error creating proxy request", http.StatusInternalServerError)
		return
	}

	// Copy headers
	for header, values := range req.Header {
		for _, value := range values {
			proxyReq.Header.Add(header, value)
		}
	}

	// Add custom headers
	proxyReq.Header.Set("X-Device-ID", r.deviceID)
	proxyReq.Header.Set("X-Creator-Address", r.creatorAddr)

	// Forward the request
	client := &http.Client{}
	resp, err := client.Do(proxyReq)
	if err != nil {
		http.Error(w, "Error forwarding request", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Copy response headers
	for header, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(header, value)
		}
	}

	// Set response status code
	w.WriteHeader(resp.StatusCode)

	// Copy response body
	if _, err := io.Copy(w, resp.Body); err != nil {
		logger.Error().Err(err).Msg("Failed to copy response body")
	}
}
