package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog"
	"github.com/theblitlabs/gologger"
)

type DockerService struct {
	log zerolog.Logger
}

func NewDockerService() *DockerService {
	return &DockerService{
		log: gologger.Get().With().Str("component", "docker").Logger(),
	}
}

func (s *DockerService) SaveImage(imageName string) (string, error) {
	s.log.Info().
		Str("image", imageName).
		Msg("Starting Docker image save operation")

	tmpDir, err := os.MkdirTemp("", "docker-images")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %v", err)
	}

	s.log.Debug().
		Str("tmpDir", tmpDir).
		Msg("Created temporary directory for Docker image")

	tarFileName := filepath.Join(tmpDir, strings.ReplaceAll(imageName, "/", "_")+".tar")
	s.log.Debug().
		Str("tarFile", tarFileName).
		Msg("Generated tar filename")

	if err := s.saveImageToTar(imageName, tarFileName); err != nil {
		if removeErr := os.RemoveAll(tmpDir); removeErr != nil {
			s.log.Error().Err(removeErr).Str("tmpDir", tmpDir).Msg("Failed to clean up temporary directory")
		}
		return "", err
	}

	fileInfo, err := os.Stat(tarFileName)
	if err == nil {
		s.log.Info().
			Str("image", imageName).
			Str("tarFile", tarFileName).
			Int64("sizeBytes", fileInfo.Size()).
			Msg("Successfully saved Docker image to tar file")
	}

	return tarFileName, nil
}

func (s *DockerService) UploadImage(tarFile string, taskData map[string]interface{}, serverURL string) error {
	defer func() {
		s.log.Debug().
			Str("tarFile", tarFile).
			Msg("Cleaning up temporary tar file")
		if err := os.Remove(tarFile); err != nil {
			s.log.Error().Err(err).Str("tarFile", tarFile).Msg("Failed to clean up temporary tar file")
		}
	}()

	imageData, err := os.ReadFile(tarFile)
	if err != nil {
		return fmt.Errorf("failed to read Docker image tar: %v", err)
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	jsonPart, err := writer.CreateFormField("task")
	if err != nil {
		return fmt.Errorf("failed to create form field: %v", err)
	}

	if err := json.NewEncoder(jsonPart).Encode(taskData); err != nil {
		return fmt.Errorf("failed to encode task request: %v", err)
	}

	imagePart, err := writer.CreateFormFile("image", filepath.Base(tarFile))
	if err != nil {
		return fmt.Errorf("failed to create form file: %v", err)
	}

	s.log.Debug().
		Str("imageSize", fmt.Sprintf("%d bytes", len(imageData))).
		Msg("Writing image data to request")

	if _, err := imagePart.Write(imageData); err != nil {
		return fmt.Errorf("failed to write image data: %v", err)
	}

	writer.Close()

	s.log.Debug().
		Str("contentType", writer.FormDataContentType()).
		Int("bodySize", body.Len()).
		Msg("Prepared multipart request")

	req, err := http.NewRequest("POST", serverURL, body)
	if err != nil {
		return fmt.Errorf("failed to create server request: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	deviceID, ok := taskData["device_id"].(string)
	if ok {
		req.Header.Set("X-Device-ID", deviceID)
	}
	creatorAddr, ok := taskData["creator_address"].(string)
	if ok {
		req.Header.Set("X-Creator-Address", creatorAddr)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request to server: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read error response from server: %v", err)
		}
		return fmt.Errorf("server returned error: status=%d, response=%s", resp.StatusCode, string(respBody))
	}

	return nil
}

func (s *DockerService) saveImageToTar(imageName, tarFileName string) error {
	s.log.Info().
		Str("image", imageName).
		Str("tarFile", tarFileName).
		Msg("Saving Docker image to tar file")

	cmd := exec.Command("docker", "save", "-o", tarFileName, imageName)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		s.log.Error().
			Err(err).
			Str("image", imageName).
			Str("stderr", stderr.String()).
			Msg("Failed to save Docker image")
		return fmt.Errorf("failed to save docker image: %v, stderr: %s", err, stderr.String())
	}

	return nil
}

func (s *DockerService) EnsureImageExists(imageName string) error {
	s.log.Info().
		Str("image", imageName).
		Msg("Checking if Docker image exists locally")

	cmd := exec.Command("docker", "image", "inspect", imageName)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		s.log.Info().
			Str("image", imageName).
			Msg("Docker image not found locally, pulling from registry")

		pullCmd := exec.Command("docker", "pull", imageName)
		var pullStderr bytes.Buffer
		pullCmd.Stderr = &pullStderr

		if err := pullCmd.Run(); err != nil {
			s.log.Error().
				Err(err).
				Str("image", imageName).
				Str("stderr", pullStderr.String()).
				Msg("Failed to pull Docker image")
			return fmt.Errorf("failed to pull Docker image: %v, stderr: %s", err, pullStderr.String())
		}

		s.log.Info().
			Str("image", imageName).
			Msg("Successfully pulled Docker image")
	} else {
		s.log.Info().
			Str("image", imageName).
			Msg("Docker image already exists locally")
	}

	return nil
}
