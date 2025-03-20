package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/theblitlabs/gologger"
	"github.com/theblitlabs/parity-client/internal/config"
	"github.com/theblitlabs/parity-client/internal/docker/service"
	"github.com/theblitlabs/parity-client/internal/task"
)

// TaskHandler handles task-related operations
type TaskHandler struct {
	config      *config.Config
	deviceID    string
	creatorAddr string
	docker      *service.DockerService
}

// NewTaskHandler creates a new task handler
func NewTaskHandler(cfg *config.Config, deviceID, creatorAddr string) *TaskHandler {
	return &TaskHandler{
		config:      cfg,
		deviceID:    deviceID,
		creatorAddr: creatorAddr,
		docker:      service.NewDockerService(),
	}
}

// ValidateAndProcessTask validates and processes a task request
func (h *TaskHandler) ValidateAndProcessTask(w http.ResponseWriter, req *task.Request) error {
	if req.Title == "" {
		return fmt.Errorf("title is required")
	}

	if len(req.Command) == 0 {
		return fmt.Errorf("command is required")
	}

	taskData := map[string]interface{}{
		"title":           req.Title,
		"description":     req.Description,
		"image":           req.Image,
		"command":         req.Command,
		"device_id":       h.deviceID,
		"creator_address": h.creatorAddr,
	}

	responseBody, err := json.Marshal(taskData)
	if err != nil {
		return fmt.Errorf("failed to marshal response: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	if _, err := w.Write(responseBody); err != nil {
		return fmt.Errorf("failed to write response: %v", err)
	}

	if req.Image != "" {
		go h.processDockerImage(req.Image, taskData)
	}

	return nil
}

func (h *TaskHandler) processDockerImage(imageName string, taskData map[string]interface{}) {
	log := gologger.Get().With().
		Str("component", "task_handler").
		Str("image", imageName).
		Logger()

	log.Info().Msg("Processing Docker image request")

	tarFile, err := h.docker.SaveImage(imageName)
	if err != nil {
		log.Error().Err(err).Msg("Failed to save Docker image")
		return
	}

	// Construct the correct upload URL
	uploadURL := fmt.Sprintf("%s/tasks", strings.TrimSuffix(h.config.Runner.ServerURL, "/"))
	log.Debug().Str("uploadURL", uploadURL).Msg("Uploading Docker image")

	if err := h.docker.UploadImage(tarFile, taskData, uploadURL); err != nil {
		log.Error().Err(err).Msg("Failed to upload Docker image")
		return
	}

	log.Info().Msg("Successfully processed and uploaded Docker image")
}
