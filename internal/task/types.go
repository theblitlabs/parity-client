package task

import "github.com/theblitlabs/parity-client/internal/docker"

// Config represents task configuration
type Config struct {
	Command []string      `json:"command"`
	Config  docker.Config `json:"config,omitempty"`
}

// Environment represents task environment configuration
type Environment struct {
	Type   string        `json:"type"`
	Config docker.Config `json:"config"`
}

// Request represents a task request
type Request struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Image       string   `json:"image"`
	Command     []string `json:"command"`
}
