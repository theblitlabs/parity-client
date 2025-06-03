package utils

import (
	"os"
	"path/filepath"
)

func GetParityConfigDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ".parity"
	}
	return filepath.Join(homeDir, ".parity")
}

func GetDefaultConfigPath() string {
	configDir := GetParityConfigDir()
	configPath := filepath.Join(configDir, ".env")

	if _, err := os.Stat(configPath); err == nil {
		return configPath
	}

	return ".env"
}

func EnsureConfigDir() error {
	configDir := GetParityConfigDir()
	return os.MkdirAll(configDir, 0755)
}
