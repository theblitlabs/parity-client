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

	if err := types.WriteJSON(w, http.StatusAccepted, taskData); err != nil {
		return fmt.Errorf("failed to write response: %v", err)
	}

	if req.Image != "" {
		go h.processDockerImage(req.Image, taskData)
	}

	return nil
}

func (h *TaskHandler) processDockerImage(imageName string, taskData map[string]interface{}) {
	h.logger.Info().
		Str("image", imageName).
		Msg("Processing Docker image request")

	if err := h.docker.EnsureImageExists(imageName); err != nil {
		h.logger.Error().Err(err).Msg("Failed to ensure Docker image exists")
		return
	}

	tarFile, err := h.docker.SaveImage(imageName)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to save Docker image")
		return
	}

	uploadURL := fmt.Sprintf("%s/tasks", strings.TrimSuffix(h.config.Runner.ServerURL, "/"))
	h.logger.Debug().Str("uploadURL", uploadURL).Msg("Uploading Docker image")

	if err := h.docker.UploadImage(tarFile, taskData, uploadURL); err != nil {
		h.logger.Error().Err(err).Msg("Failed to upload Docker image")
		return
	}

	h.logger.Info().Msg("Successfully processed and uploaded Docker image")
}
