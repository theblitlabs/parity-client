package task

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/theblitlabs/parity-client/internal/models"
)

// Request represents the incoming task request from the client
type Request struct {
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Image       string            `json:"image"`
	Command     []string          `json:"command"`
	Env         map[string]string `json:"env,omitempty"`
	Resources   ResourceConfig    `json:"resources,omitempty"`
}

// ResourceConfig represents resource constraints for the task
type ResourceConfig struct {
	Memory    string `json:"memory,omitempty"`
	CPUShares int64  `json:"cpu_shares,omitempty"`
	Timeout   string `json:"timeout,omitempty"`
}

// ValidateAndTransform validates the incoming request and transforms it into a server Task
func ValidateAndTransform(req Request, creatorAddress, creatorDeviceID string) (*models.Task, error) {
	// Basic validation
	if req.Title == "" {
		return nil, fmt.Errorf("title is required")
	}

	if len(req.Command) == 0 {
		return nil, fmt.Errorf("command is required")
	}

	// Create task config based on request type
	taskConfig := models.TaskConfig{
		Command:   req.Command,
		Env:       req.Env,
		Resources: models.ResourceConfig(req.Resources),
	}

	// Determine task type and environment config
	var taskType models.TaskType
	var envConfig *models.EnvironmentConfig

	if req.Image != "" {
		taskType = models.TaskTypeDocker
		envConfig = &models.EnvironmentConfig{
			Type: "docker",
			Config: map[string]interface{}{
				"image":   req.Image,
				"command": req.Command,
			},
		}
	} else {
		taskType = models.TaskTypeCommand
	}

	// Marshal task config to JSON
	configJSON, err := json.Marshal(taskConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal task config: %w", err)
	}

	// Create new task
	task := models.NewTask()
	task.Title = req.Title
	task.Description = req.Description
	task.Type = taskType
	task.Config = configJSON
	task.Environment = envConfig
	task.CreatorAddress = creatorAddress
	task.CreatorDeviceID = creatorDeviceID
	task.Status = models.TaskStatusPending
	task.CreatedAt = time.Now()
	task.UpdatedAt = time.Now()

	// Validate the complete task
	if err := task.Validate(); err != nil {
		return nil, fmt.Errorf("task validation failed: %w", err)
	}

	return task, nil
}
