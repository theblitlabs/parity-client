package task

import "github.com/theblitlabs/parity-client/internal/docker/service"

type Config struct {
	Command []string       `json:"command"`
	Config  service.Config `json:"config,omitempty"`
}

type Environment struct {
	Type   string         `json:"type"`
	Config service.Config `json:"config"`
}

type Request struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Image       string   `json:"image"`
	Command     []string `json:"command"`
}
