package docker

// Config represents Docker container configuration
type Config struct {
	Image   string   `json:"image"`
	Workdir string   `json:"workdir"`
	Command []string `json:"command,omitempty"`
}

// Task represents a Docker task configuration
type Task struct {
	Image   string   `json:"image"`
	Command []string `json:"command,omitempty"`
}

// ResourceConfig represents resource limits for Docker containers
type ResourceConfig struct {
	Memory    string `json:"memory,omitempty"`
	CPUShares int64  `json:"cpu_shares,omitempty"`
	Timeout   string `json:"timeout,omitempty"`
}
