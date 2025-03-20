package docker

type Config struct {
	Image   string   `json:"image"`
	Workdir string   `json:"workdir"`
	Command []string `json:"command,omitempty"`
}

type Task struct {
	Image   string   `json:"image"`
	Command []string `json:"command,omitempty"`
}

type ResourceConfig struct {
	Memory    string `json:"memory,omitempty"`
	CPUShares int64  `json:"cpu_shares,omitempty"`
	Timeout   string `json:"timeout,omitempty"`
}
