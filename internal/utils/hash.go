package utils

import (
	"crypto/sha256"
	"fmt"
	"os/exec"
	"strings"
)

func ComputeImageHash(imageName string) (string, error) {
	cmd := exec.Command("docker", "save", imageName)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to save image for hashing: %w", err)
	}

	hash := sha256.Sum256(output)
	return fmt.Sprintf("%x", hash), nil
}

func ComputeCommandHash(command []string) string {
	commandStr := strings.Join(command, " ")
	hash := sha256.Sum256([]byte(commandStr))
	return fmt.Sprintf("%x", hash)
}

func ComputeResultHash(stdout, stderr string, exitCode int) string {
	combined := fmt.Sprintf("%s%s%d", stdout, stderr, exitCode)
	hash := sha256.Sum256([]byte(combined))
	return fmt.Sprintf("%x", hash)
}
