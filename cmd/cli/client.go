package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/theblitlabs/deviceid"
	"github.com/theblitlabs/gologger"
	"github.com/theblitlabs/parity-client/internal/config"
	"github.com/theblitlabs/parity-client/internal/models"
	"github.com/theblitlabs/parity-client/internal/task"
)

type KeyStore struct {
	PrivateKey string `json:"private_key"`
}

type DockerConfig struct {
	Image   string   `json:"image"`
	Workdir string   `json:"workdir"`
	Command []string `json:"command,omitempty"`
}

type TaskConfig struct {
	Command []string     `json:"command"`
	Config  DockerConfig `json:"config,omitempty"`
}

type TaskEnvironment struct {
	Type   string       `json:"type"`
	Config DockerConfig `json:"config"`
}

type DockerTask struct {
	Image   string   `json:"image"`
	Command []string `json:"command,omitempty"`
}

type ResourceConfig struct {
	Memory    string `json:"memory,omitempty"`
	CPUShares int64  `json:"cpu_shares,omitempty"`
	Timeout   string `json:"timeout,omitempty"`
}

type TaskRequest struct {
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Image       string            `json:"image"`
	Command     []string          `json:"command"`
	Env         map[string]string `json:"env,omitempty"`
	Resources   ResourceConfig    `json:"resources,omitempty"`
}

func isPortAvailable(port int) error {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("port %d is not available: %w", port, err)
	}
	ln.Close()
	return nil
}

func saveDockerImage(imageName string) (string, error) {
	log := gologger.Get().With().Str("component", "docker").Logger()

	log.Info().
		Str("image", imageName).
		Msg("Starting Docker image save operation")

	tmpDir, err := os.MkdirTemp("", "docker-images")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}
	log.Debug().
		Str("tmpDir", tmpDir).
		Msg("Created temporary directory for Docker image")

	tarFileName := filepath.Join(tmpDir, strings.ReplaceAll(imageName, "/", "_")+".tar")
	log.Debug().
		Str("tarFile", tarFileName).
		Msg("Generated tar filename")

	log.Info().
		Str("image", imageName).
		Str("tarFile", tarFileName).
		Msg("Saving Docker image to tar file")

	cmd := exec.Command("docker", "save", "-o", tarFileName, imageName)
	if output, err := cmd.CombinedOutput(); err != nil {
		log.Error().
			Err(err).
			Str("image", imageName).
			Str("output", string(output)).
			Msg("Failed to save Docker image")
		if rmErr := os.RemoveAll(tmpDir); rmErr != nil {
			log.Error().Err(rmErr).Str("tmpDir", tmpDir).Msg("Failed to clean up temporary directory")
		}
		return "", fmt.Errorf("failed to save docker image: %w", err)
	}

	fileInfo, err := os.Stat(tarFileName)
	if err == nil {
		log.Info().
			Str("image", imageName).
			Str("tarFile", tarFileName).
			Int64("sizeBytes", fileInfo.Size()).
			Msg("Successfully saved Docker image to tar file")
	}

	return tarFileName, nil
}

func getCreatorAddress() (string, error) {
	keystoreDir := filepath.Join(os.Getenv("HOME"), ".parity")
	keystorePath := filepath.Join(keystoreDir, "keystore.json")

	data, err := os.ReadFile(keystorePath)
	if err != nil {
		return "", fmt.Errorf("failed to read keystore: %w", err)
	}

	var keystore KeyStore
	if err := json.Unmarshal(data, &keystore); err != nil {
		return "", fmt.Errorf("failed to parse keystore: %w", err)
	}

	privateKey := keystore.PrivateKey
	ecdsaKey, err := crypto.HexToECDSA(privateKey)
	if err != nil {
		return "", fmt.Errorf("invalid private key in keystore: %w", err)
	}

	address := crypto.PubkeyToAddress(ecdsaKey.PublicKey)
	return address.Hex(), nil
}

func RunChain(port int) {
	log := gologger.Get().With().Str("component", "chain").Logger()

	// Load config
	cfg, err := config.LoadConfig("config/config.yaml")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load config")
	}

	// Get or generate device ID
	deviceIDManager := deviceid.NewManager(deviceid.Config{})
	deviceID, err := deviceIDManager.VerifyDeviceID()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to verify device ID")
	}

	// Get creator address from keystore
	creatorAddress, err := getCreatorAddress()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to get creator address. Please authenticate first using 'auth' command")
	}

	// Proxy request to the server
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Debug().
			Str("original_path", r.URL.Path).
			Str("method", r.Method).
			Str("content_type", r.Header.Get("Content-Type")).
			Msg("Received request")

		path := strings.TrimPrefix(r.URL.Path, "/")
		path = strings.TrimPrefix(path, "api/")

		targetURL := fmt.Sprintf("%s/%s", strings.TrimSuffix(cfg.Runner.ServerURL, "/"), path)

		var proxyReq *http.Request
		var err error

		if r.Method == "POST" && strings.Contains(r.Header.Get("Content-Type"), "application/json") {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "Error reading request body", http.StatusInternalServerError)
				return
			}
			r.Body.Close()

			var taskRequest TaskRequest
			if err := json.Unmarshal(body, &taskRequest); err != nil {
				log.Error().Err(err).Msg("Failed to decode request body")
				http.Error(w, "Invalid request body", http.StatusBadRequest)
				return
			}

			validatedTask, err := task.ValidateAndTransform(task.Request{
				Title:       taskRequest.Title,
				Description: taskRequest.Description,
				Image:       taskRequest.Image,
				Command:     taskRequest.Command,
				Env:         taskRequest.Env,
				Resources:   task.ResourceConfig(taskRequest.Resources),
			}, creatorAddress, deviceID)
			if err != nil {
				log.Error().Err(err).Msg("Task validation failed")
				http.Error(w, fmt.Sprintf("Task validation failed: %v", err), http.StatusBadRequest)
				return
			}

			responseBody, err := json.Marshal(validatedTask)
			if err != nil {
				log.Error().Err(err).Msg("Failed to marshal response")
				http.Error(w, "Error processing request", http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusAccepted)
			w.Write(responseBody)

			if taskRequest.Image != "" {
				go func() {
					log := log.With().
						Str("image", taskRequest.Image).
						Str("path", path).
						Logger()

					log.Info().Msg("Processing Docker image request")

					imageName := taskRequest.Image
					tarFile, err := saveDockerImage(imageName)
					if err != nil {
						log.Error().Err(err).Msg("Failed to save Docker image")
						return
					}
					defer func() {
						log.Debug().
							Str("tarFile", tarFile).
							Msg("Cleaning up temporary tar file")
						if err := os.Remove(tarFile); err != nil {
							log.Error().Err(err).Str("tarFile", tarFile).Msg("Failed to clean up temporary tar file")
						}
					}()

					imageData, err := os.ReadFile(tarFile)
					if err != nil {
						log.Error().
							Err(err).
							Str("tarFile", tarFile).
							Msg("Failed to read Docker image tar")
						return
					}

					body := &bytes.Buffer{}
					writer := NewMultipartWriter(body)

					jsonPart, err := writer.CreateFormField("task")
					if err != nil {
						log.Error().Err(err).Msg("Failed to create form field")
						return
					}

					// Unmarshal the config
					var taskConfig models.TaskConfig
					if err := json.Unmarshal(validatedTask.Config, &taskConfig); err != nil {
						log.Error().Err(err).Msg("Failed to unmarshal task config")
						return
					}

					taskData := map[string]interface{}{
						"title":             validatedTask.Title,
						"description":       validatedTask.Description,
						"image":             taskRequest.Image,
						"command":           taskRequest.Command,
						"creator_device_id": validatedTask.CreatorDeviceID,
						"creator_address":   validatedTask.CreatorAddress,
					}

					if err := json.NewEncoder(jsonPart).Encode(taskData); err != nil {
						log.Error().Err(err).Msg("Failed to encode task request")
						return
					}

					// Add the image tar part
					imagePart, err := writer.CreateFormFile("image", filepath.Base(tarFile))
					if err != nil {
						log.Error().Err(err).Msg("Failed to create form file")
						return
					}

					log.Debug().
						Str("imageSize", fmt.Sprintf("%d bytes", len(imageData))).
						Msg("Writing image data to request")

					if _, err := imagePart.Write(imageData); err != nil {
						log.Error().Err(err).Msg("Failed to write image data")
						return
					}

					writer.Close()

					log.Debug().
						Str("contentType", writer.FormDataContentType()).
						Int("bodySize", body.Len()).
						Msg("Prepared multipart request")

					// Create and send multipart request to server
					req, err := http.NewRequest("POST", targetURL, body)
					if err != nil {
						log.Error().Err(err).Msg("Failed to create server request")
						return
					}
					req.Header.Set("Content-Type", writer.FormDataContentType())
					req.Header.Set("X-Device-ID", deviceID)
					req.Header.Set("X-Creator-Address", creatorAddress)

					client := &http.Client{}
					resp, err := client.Do(req)
					if err != nil {
						log.Error().Err(err).Msg("Failed to send request to server")
						return
					}
					defer resp.Body.Close()

					if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
						respBody, err := io.ReadAll(resp.Body)
						if err != nil {
							log.Error().
								Err(err).
								Int("statusCode", resp.StatusCode).
								Msg("Failed to read error response from server")
							return
						}

						log.Error().
							Int("statusCode", resp.StatusCode).
							Str("response", string(respBody)).
							Str("contentType", resp.Header.Get("Content-Type")).
							Msg("Server returned error response")
						return
					}

					log.Info().
						Int("statusCode", resp.StatusCode).
						Msg("Successfully processed and uploaded Docker image")
				}()
				return
			}

			proxyReq, err = http.NewRequest(r.Method, targetURL, bytes.NewBuffer(responseBody))
			if err != nil {
				log.Error().Err(err).Msg("Failed to create server request")
				return
			}
			proxyReq.Header.Set("Content-Type", "application/json")
			proxyReq.Header.Set("X-Device-ID", deviceID)
			proxyReq.Header.Set("X-Creator-Address", creatorAddress)
		} else {
			proxyReq, err = http.NewRequest(r.Method, targetURL, r.Body)
			if err != nil {
				http.Error(w, "Error creating proxy request", http.StatusInternalServerError)
				return
			}
		}

		// Copy headers
		for header, values := range r.Header {
			for _, value := range values {
				proxyReq.Header.Add(header, value)
			}
		}

		// Forward the request
		client := &http.Client{}
		resp, err := client.Do(proxyReq)
		if err != nil {
			log.Error().Err(err).Msg("Error forwarding request")
			http.Error(w, "Error forwarding request", http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		// Copy response headers
		for header, values := range resp.Header {
			for _, value := range values {
				w.Header().Add(header, value)
			}
		}

		// Set response status code
		w.WriteHeader(resp.StatusCode)

		// Copy response body
		if _, err := io.Copy(w, resp.Body); err != nil {
			log.Error().Err(err).Msg("Failed to copy response body")
		}
	})

	// Start local proxy server
	localAddr := fmt.Sprintf("%s:%d", cfg.Server.Host, port)

	// Check if specified port is available
	if err := isPortAvailable(port); err != nil {
		log.Fatal().Err(err).Int("port", port).Msg("Chain proxy port is not available")
	}

	log.Info().
		Str("address", localAddr).
		Str("device_id", deviceID).
		Str("creator_address", creatorAddress).
		Int("port", port).
		Msg("Starting chain proxy server")

	if err := http.ListenAndServe(localAddr, nil); err != nil {
		log.Fatal().Err(err).Msg("Failed to start chain proxy server")
	}
}

func NewMultipartWriter(body *bytes.Buffer) *multipart.Writer {
	return multipart.NewWriter(body)
}
