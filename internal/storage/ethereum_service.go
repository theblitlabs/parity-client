package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	shell "github.com/ipfs/go-ipfs-api"
	"github.com/theblitlabs/gologger"
	"github.com/theblitlabs/parity-client/internal/config"
)

type BlockchainService struct {
	ipfsClient *shell.Shell
	gatewayURL string
}

func NewBlockchainService(cfg *config.Config) (*BlockchainService, error) {
	ipfsEndpoint := cfg.BlockchainNetwork.IPFSEndpoint
	if ipfsEndpoint == "" {
		ipfsEndpoint = "localhost:5001" // fallback
	}

	gatewayURL := cfg.BlockchainNetwork.GatewayURL
	if gatewayURL == "" {
		gatewayURL = "https://ipfs.io" // fallback
	}

	ipfsClient := shell.NewShell(ipfsEndpoint)

	return &BlockchainService{
		ipfsClient: ipfsClient,
		gatewayURL: gatewayURL,
	}, nil
}

func (f *BlockchainService) UploadFile(ctx context.Context, filePath string) (string, error) {
	log := gologger.Get()

	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			log.Printf("Error closing file: %v", closeErr)
		}
	}()

	filename := filepath.Base(filePath)
	uniqueFilename := fmt.Sprintf("%s-%s%s",
		filename[:len(filename)-len(filepath.Ext(filename))],
		uuid.New().String()[:8],
		filepath.Ext(filename))

	log.Info().
		Str("file", filePath).
		Str("unique_name", uniqueFilename).
		Msg("Uploading file to IPFS")

	cid, err := f.ipfsClient.Add(file, shell.Pin(true))
	if err != nil {
		log.Error().Err(err).
			Str("file", filePath).
			Msg("Failed to upload file to IPFS")
		return "", fmt.Errorf("failed to upload file to IPFS: %w", err)
	}

	log.Info().
		Str("file", filePath).
		Str("cid", cid).
		Str("gateway_url", f.GetFileURL(cid)).
		Msg("Successfully uploaded file to IPFS")

	return cid, nil
}

func (f *BlockchainService) UploadData(ctx context.Context, data []byte, filename string) (string, error) {
	log := gologger.Get()

	uniqueFilename := fmt.Sprintf("%s-%s", filename, uuid.New().String()[:8])

	log.Info().
		Str("filename", uniqueFilename).
		Int("size_bytes", len(data)).
		Msg("Uploading data to IPFS")

	reader := bytes.NewReader(data)

	cid, err := f.ipfsClient.Add(reader, shell.Pin(true))
	if err != nil {
		log.Error().Err(err).
			Str("filename", uniqueFilename).
			Msg("Failed to upload data to IPFS")
		return "", fmt.Errorf("failed to upload data to IPFS: %w", err)
	}

	log.Info().
		Str("filename", uniqueFilename).
		Str("cid", cid).
		Str("gateway_url", f.GetFileURL(cid)).
		Msg("Successfully uploaded data to IPFS")

	return cid, nil
}

func (f *BlockchainService) UploadDirectory(ctx context.Context, dirPath string) (string, error) {
	log := gologger.Get()

	log.Info().
		Str("directory", dirPath).
		Msg("Uploading directory to IPFS")

	cid, err := f.ipfsClient.AddDir(dirPath)
	if err != nil {
		log.Error().Err(err).
			Str("directory", dirPath).
			Msg("Failed to upload directory to IPFS")
		return "", fmt.Errorf("failed to upload directory to IPFS: %w", err)
	}

	log.Info().
		Str("directory", dirPath).
		Str("cid", cid).
		Str("gateway_url", f.GetFileURL(cid)).
		Msg("Successfully uploaded directory to IPFS")

	return cid, nil
}

func (f *BlockchainService) DownloadFile(ctx context.Context, cid string, outputPath string) error {
	log := gologger.Get()

	log.Info().
		Str("cid", cid).
		Str("output", outputPath).
		Msg("Downloading file from IPFS")

	reader, err := f.ipfsClient.Cat(cid)
	if err != nil {
		log.Error().Err(err).
			Str("cid", cid).
			Msg("Failed to download file from IPFS")
		return fmt.Errorf("failed to download file from IPFS: %w", err)
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			log.Error().Err(closeErr).Msg("Error closing reader")
		}
	}()

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			log.Error().Err(closeErr).Msg("Error closing file")
		}
	}()

	_, err = io.Copy(file, reader)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	log.Info().
		Str("cid", cid).
		Str("output", outputPath).
		Msg("Successfully downloaded file from IPFS")

	return nil
}

func (f *BlockchainService) GetFileURL(cid string) string {
	return fmt.Sprintf("%s/ipfs/%s", f.gatewayURL, cid)
}

func (f *BlockchainService) PinFile(cid string) error {
	return f.ipfsClient.Pin(cid)
}

func (f *BlockchainService) UnpinFile(cid string) error {
	return f.ipfsClient.Unpin(cid)
}

func (f *BlockchainService) GetFileInfo(cid string) (*shell.ObjectStats, error) {
	return f.ipfsClient.ObjectStat(cid)
}
