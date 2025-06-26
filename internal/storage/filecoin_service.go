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

type FilecoinService struct {
	ipfsClient  *shell.Shell
	gatewayURL  string
	storageDeal bool
}

func NewFilecoinService(cfg *config.Config) (*FilecoinService, error) {
	ipfsEndpoint := cfg.FilecoinNetwork.IPFSEndpoint
	if ipfsEndpoint == "" {
		ipfsEndpoint = "localhost:5001" // fallback
	}

	gatewayURL := cfg.FilecoinNetwork.GatewayURL
	if gatewayURL == "" {
		gatewayURL = "https://ipfs.io" // fallback
	}

	ipfsClient := shell.NewShell(ipfsEndpoint)

	return &FilecoinService{
		ipfsClient:  ipfsClient,
		gatewayURL:  gatewayURL,
		storageDeal: cfg.FilecoinNetwork.CreateStorageDeals,
	}, nil
}

func (f *FilecoinService) UploadFile(ctx context.Context, filePath string) (string, error) {
	log := gologger.Get()

	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer file.Close()

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

	if f.storageDeal {
		if err := f.createStorageDeal(ctx, cid); err != nil {
			log.Warn().Err(err).
				Str("cid", cid).
				Msg("Failed to create storage deal, file stored on IPFS only")
		}
	}

	log.Info().
		Str("file", filePath).
		Str("cid", cid).
		Str("gateway_url", f.GetFileURL(cid)).
		Msg("Successfully uploaded file to IPFS/Filecoin")

	return cid, nil
}

func (f *FilecoinService) UploadData(ctx context.Context, data []byte, filename string) (string, error) {
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

	if f.storageDeal {
		if err := f.createStorageDeal(ctx, cid); err != nil {
			log.Warn().Err(err).
				Str("cid", cid).
				Msg("Failed to create storage deal, data stored on IPFS only")
		}
	}

	log.Info().
		Str("filename", uniqueFilename).
		Str("cid", cid).
		Str("gateway_url", f.GetFileURL(cid)).
		Msg("Successfully uploaded data to IPFS/Filecoin")

	return cid, nil
}

func (f *FilecoinService) UploadDirectory(ctx context.Context, dirPath string) (string, error) {
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

	if f.storageDeal {
		if err := f.createStorageDeal(ctx, cid); err != nil {
			log.Warn().Err(err).
				Str("cid", cid).
				Msg("Failed to create storage deal, directory stored on IPFS only")
		}
	}

	log.Info().
		Str("directory", dirPath).
		Str("cid", cid).
		Str("gateway_url", f.GetFileURL(cid)).
		Msg("Successfully uploaded directory to IPFS/Filecoin")

	return cid, nil
}

func (f *FilecoinService) DownloadFile(ctx context.Context, cid string, outputPath string) error {
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
	defer reader.Close()

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

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

func (f *FilecoinService) GetFileURL(cid string) string {
	return fmt.Sprintf("%s/ipfs/%s", f.gatewayURL, cid)
}

func (f *FilecoinService) PinFile(cid string) error {
	return f.ipfsClient.Pin(cid)
}

func (f *FilecoinService) UnpinFile(cid string) error {
	return f.ipfsClient.Unpin(cid)
}

func (f *FilecoinService) GetFileInfo(cid string) (*shell.ObjectStats, error) {
	return f.ipfsClient.ObjectStat(cid)
}

func (f *FilecoinService) createStorageDeal(ctx context.Context, cid string) error {
	log := gologger.Get()

	log.Info().
		Str("cid", cid).
		Msg("Creating storage deal for file")

	// This is a placeholder for actual Filecoin storage deal creation
	// In a real implementation, this would interact with Filecoin storage providers
	// to create storage deals for the uploaded content

	return nil
}
