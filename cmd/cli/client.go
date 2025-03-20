package cli

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/theblitlabs/deviceid"
	"github.com/theblitlabs/gologger"
	"github.com/theblitlabs/parity-client/internal/adapters/keystore"
	"github.com/theblitlabs/parity-client/internal/config"
	"github.com/theblitlabs/parity-client/internal/proxy"
)

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
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Image       string   `json:"image"`
	Command     []string `json:"command"`
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
	ks, err := keystore.NewAdapter(nil)
	if err != nil {
		return "", fmt.Errorf("failed to create keystore: %v", err)
	}

	privateKey, err := ks.LoadPrivateKey()
	if err != nil {
		return "", fmt.Errorf("failed to load private key: %v", err)
	}

	address := crypto.PubkeyToAddress(privateKey.PublicKey)
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

	// Create and start proxy server
	server := proxy.NewServer(cfg, deviceID, creatorAddress, port)
	if err := server.Start(); err != nil {
		log.Fatal().Err(err).Msg("Failed to start chain proxy server")
	}
}

func NewMultipartWriter(body *bytes.Buffer) *multipart.Writer {
	return multipart.NewWriter(body)
}
