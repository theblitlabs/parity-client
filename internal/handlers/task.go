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
)

type TaskHandler struct {
	config      *config.Config
	deviceID    string
	creatorAddr string
	docker      *service.DockerService
	logger      zerolog.Logger
}

type TaskProcessingError struct {
	StatusCode int
	Message    string
	Hint       string
	NextSteps  []string
	Cause      error
}

func (e *TaskProcessingError) Error() string {
	if e.Cause == nil {
		return e.Message
	}

	return fmt.Sprintf("%s: %v", e.Message, e.Cause)
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
		return &TaskProcessingError{
			StatusCode: http.StatusBadRequest,
			Message:    "task title is required",
			Hint:       "Set a short title so the task is easy to identify in listings and logs.",
			NextSteps: []string{
				"Add a non-empty title field and retry the request.",
			},
		}
	}

	taskData := map[string]interface{}{
		"type":            "docker",
		"title":           req.Title,
		"description":     req.Description,
		"image":           req.Image,
		"command":         req.Command,
		"device_id":       h.deviceID,
		"creator_address": h.creatorAddr,
	}

	if req.Image == "" {
		return &TaskProcessingError{
			StatusCode: http.StatusBadRequest,
			Message:    "docker image is required for local image upload",
			Hint:       "Use /api/v1/tasks/local only when the client should package a Docker image from this machine. Use /api/v1/tasks for registry-backed tasks instead.",
			NextSteps: []string{
				"Set image to a local Docker image name such as my-image:latest.",
				"Or send the same JSON to /api/v1/tasks if the runner should pull the image from a registry.",
			},
		}
	}

	h.logger.Info().
		Str("image", req.Image).
		Msg("Processing Docker image request")

	if err := h.docker.EnsureImageExists(req.Image); err != nil {
		return &TaskProcessingError{
			StatusCode: http.StatusBadRequest,
			Message:    fmt.Sprintf("unable to prepare Docker image %q for local upload", req.Image),
			Hint:       "The client checks the image locally and then tries docker pull if it is missing. If both fail, the image cannot be packaged for upload.",
			NextSteps: []string{
				fmt.Sprintf("Run docker pull %s or docker build -t %s . and retry.", req.Image, req.Image),
				"Or submit the task to /api/v1/tasks if the runner should pull from a registry directly.",
			},
			Cause: err,
		}
	}

	tarFile, err := h.docker.SaveImage(req.Image)
	if err != nil {
		return &TaskProcessingError{
			StatusCode: http.StatusBadRequest,
			Message:    fmt.Sprintf("failed to export Docker image %q for upload", req.Image),
			Hint:       "The image exists, but docker save could not create a tarball. This is usually a local Docker or disk-space issue.",
			NextSteps: []string{
				"Check that the Docker daemon is running and the current user can access it.",
				"Verify there is enough free disk space for the image tarball.",
			},
			Cause: err,
		}
	}

	uploadURL := fmt.Sprintf("%s/api/v1/tasks", normalizeServerBaseURL(h.config.Runner.ServerURL))
	h.logger.Debug().Str("uploadURL", uploadURL).Msg("Uploading Docker image")

	statusCode, responseBody, err := h.docker.UploadImage(tarFile, taskData, uploadURL)
	if err != nil {
		hint := "The client packaged the local image, but the upstream server rejected or failed the upload."
		nextSteps := []string{
			"Check the client /health and server /health endpoints, then retry.",
			"Use /api/v1/tasks for registry-backed tasks if the image does not need to be uploaded from this machine.",
		}
		if strings.Contains(err.Error(), "server returned error:") {
			hint = "The upstream server returned an application error for the uploaded task."
			nextSteps = []string{
				"Inspect the server response in the error details and correct the task payload if needed.",
				"Retry once the upstream server is healthy and reachable.",
			}
		}

		return &TaskProcessingError{
			StatusCode: taskUploadStatusCode(err),
			Message:    "failed to upload local Docker image to the Parity server",
			Hint:       hint,
			NextSteps:  nextSteps,
			Cause:      err,
		}
	}

	h.logger.Info().Msg("Successfully processed and uploaded Docker image")
	w.Header().Set("X-Parity-Proxy-Mode", "local-image-upload")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if _, err := w.Write(responseBody); err != nil {
		return fmt.Errorf("failed to write upload response: %w", err)
	}
	return nil
}

func taskUploadStatusCode(err error) int {
	message := err.Error()
	if !strings.Contains(message, "server returned error: status=") {
		return http.StatusBadGateway
	}

	var statusCode int
	if _, scanErr := fmt.Sscanf(message, "server returned error: status=%d,", &statusCode); scanErr == nil && statusCode >= 400 && statusCode <= 599 {
		return statusCode
	}

	return http.StatusBadGateway
}
