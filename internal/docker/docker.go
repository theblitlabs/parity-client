package docker

import (
	"bytes"
	"fmt"
	"os/exec"
)

// SaveImage saves a Docker image to a tar file
func SaveImage(imageName string) (string, error) {
	outputFile := fmt.Sprintf("%s.tar", imageName)
	cmd := exec.Command("docker", "save", "-o", outputFile, imageName)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to save docker image: %v, stderr: %s", err, stderr.String())
	}

	return outputFile, nil
}
