package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/rs/zerolog"
	"github.com/theblitlabs/gologger"
	"github.com/theblitlabs/parity-client/internal/config"
	"github.com/theblitlabs/parity-client/internal/docker/service"
	"github.com/theblitlabs/parity-client/internal/task"
	"github.com/theblitlabs/parity-client/internal/types"
)

type TaskHandler struct {
	config      *config.Config
	deviceID    string
	creatorAddr string
	docker      *service.DockerService
	logger      zerolog.Logger
}

func NewTaskHandler(cfg *config.Config, deviceID, creatorAddr string) *TaskHandler {
	return &TaskHandler{
		config:      cfg,
		deviceID:    deviceID,
		creatorAddr: creatorAddr,
		docker:      service.NewDockerService(),
		logger:      gologger.Get().With().Str("component", "task_handler").Logger(),
	}
}

func (h *TaskHandler) ValidateAndProcessTask(w http.ResponseWriter, req *task.Request) error {
	if req.Title == "" {
		return fmt.Errorf("title is required")
	}

	taskData := map[string]interface{}{
		"title":           req.Title,
		"description":     req.Description,
		"image":           req.Image,
		"command":         req.Command,
		"device_id":       h.deviceID,
		"creator_address": h.creatorAddr,
	}

	if req.Image == "" {
		return fmt.Errorf("image is required")
	}

	h.logger.Info().
		Str("image", req.Image).
		Msg("Processing Docker image request")

	if err := h.docker.EnsureImageExists(req.Image); err != nil {
		return fmt.Errorf("failed to ensure Docker image exists: %v", err)
	}

	tarFile, err := h.docker.SaveImage(req.Image)
	if err != nil {
		return fmt.Errorf("failed to save Docker image: %v", err)
	}

	uploadURL := fmt.Sprintf("%s/api/v1/tasks", strings.TrimSuffix(h.config.Runner.ServerURL, "/"))
	h.logger.Debug().Str("uploadURL", uploadURL).Msg("Uploading Docker image")

	if err := h.docker.UploadImage(tarFile, taskData, uploadURL); err != nil {
		return fmt.Errorf("failed to upload Docker image: %v", err)
	}

	h.logger.Info().Msg("Successfully processed and uploaded Docker image")
	return types.WriteJSON(w, http.StatusCreated, taskData)
}
