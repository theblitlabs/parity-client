package cli

import (
	"bytes"
	"fmt"
	"mime/multipart"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/spf13/cobra"
	"github.com/theblitlabs/deviceid"
	"github.com/theblitlabs/gologger"
	"github.com/theblitlabs/parity-client/internal/adapters/keystore"
	"github.com/theblitlabs/parity-client/internal/client"
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

func RunChain(port int, cmd *cobra.Command) {
	log := gologger.Get().With().Str("component", "chain").Logger()

	if err := client.IsPortAvailable(port); err != nil {
		log.Error().Err(err).Int("port", port).Msg("Port is not available")
		return
	}

	configPath, _ := cmd.Flags().GetString("config-path")
	configManager := config.NewConfigManager(configPath)
	cfg, err := configManager.GetConfig()
	if err != nil {
		log.Error().Err(err).Msg("Failed to load config")
		return
	}

	deviceIDManager := deviceid.NewManager(deviceid.Config{})
	deviceID, err := deviceIDManager.VerifyDeviceID()
	if err != nil {
		log.Error().Err(err).Msg("Failed to verify device ID")
		return
	}

	creatorAddress, err := getCreatorAddress()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get creator address. Please authenticate first using 'auth' command")
		return
	}

	server := proxy.NewServer(cfg, deviceID, creatorAddress, port)
	if err := server.Start(); err != nil {
		log.Error().Err(err).Msg("Failed to start chain proxy server")
		return
	}
}

func NewMultipartWriter(body *bytes.Buffer) *multipart.Writer {
	return multipart.NewWriter(body)
}
