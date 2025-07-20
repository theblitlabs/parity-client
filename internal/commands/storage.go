package commands

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/theblitlabs/gologger"
	"github.com/theblitlabs/parity-client/internal/config"
	"github.com/theblitlabs/parity-client/internal/storage"
	"github.com/theblitlabs/parity-client/internal/utils"
)

var storageCmd = &cobra.Command{
	Use:   "storage",
	Short: "Manage IPFS storage operations",
	Long:  `Upload, download, and manage files on IPFS for federated learning training data`,
}

var uploadFileCmd = &cobra.Command{
	Use:   "upload-file [file-path]",
	Short: "Upload a file to IPFS",
	Args:  cobra.ExactArgs(1),
	Example: `  # Upload a dataset file
  parity-client storage upload-file /path/to/dataset.csv
  
  # Upload with custom filename
  parity-client storage upload-file /path/to/data.zip --filename training-data.zip`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runUploadFile(cmd, args); err != nil {
			log := gologger.Get()
			log.Error().Err(err).Msg("Upload file failed")
			os.Exit(1)
		}
	},
}

func runUploadFile(cmd *cobra.Command, args []string) error {
	log := gologger.Get()

	filePath := args[0]
	customName, _ := cmd.Flags().GetString("name")
	shouldPin, _ := cmd.Flags().GetBool("pin")

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", filePath)
	}

	if filePath == "" {
		return fmt.Errorf("file path is required")
	}

	configPath := utils.GetDefaultConfigPath()
	configManager := config.NewConfigManager(configPath)
	cfg, err := configManager.GetConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	blockchainService, err := storage.NewBlockchainService(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize blockchain storage service: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	if customName != "" {
		log.Info().Str("file", filePath).Str("name", customName).Msg("Starting file upload to IPFS with custom name")
	} else {
		log.Info().Str("file", filePath).Msg("Starting file upload to IPFS")
	}

	cid, err := blockchainService.UploadFile(ctx, filePath)
	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}

	if shouldPin {
		if err := blockchainService.PinFile(cid); err != nil {
			log.Warn().Err(err).Msg("Failed to pin file, but upload was successful")
		} else {
			log.Info().Str("cid", cid).Msg("File pinned successfully")
		}
	}

	fileUrl := blockchainService.GetFileURL(cid)

	log.Info().
		Str("file", filePath).
		Str("cid", cid).
		Str("url", fileUrl).
		Msg("File uploaded successfully")

	fmt.Printf("File uploaded successfully!\n")
	if customName != "" {
		fmt.Printf("Name: %s\n", customName)
	}
	fmt.Printf("CID: %s\n", cid)
	fmt.Printf("URL: %s\n", fileUrl)
	if shouldPin {
		fmt.Printf("File pinned to IPFS\n")
	}
	fmt.Printf("\nUse this CID when creating federated learning sessions:\n")
	fmt.Printf("parity-client fl create-session --dataset-cid %s [other options]\n", cid)

	return nil
}

var uploadDirectoryCmd = &cobra.Command{
	Use:   "upload-dir [directory-path]",
	Short: "Upload a directory to IPFS",
	Args:  cobra.ExactArgs(1),
	Example: `  # Upload a dataset directory
  parity-client storage upload-dir /path/to/dataset/
  
  # Upload training data directory
  parity-client storage upload-dir ./training_data/`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runUploadDirectory(cmd, args); err != nil {
			log := gologger.Get()
			log.Error().Err(err).Msg("Upload directory failed")
			os.Exit(1)
		}
	},
}

func runUploadDirectory(cmd *cobra.Command, args []string) error {
	log := gologger.Get()

	dirPath := args[0]
	shouldCompress, _ := cmd.Flags().GetBool("compress")

	if info, err := os.Stat(dirPath); os.IsNotExist(err) || !info.IsDir() {
		return fmt.Errorf("directory does not exist: %s", dirPath)
	}

	if dirPath == "" {
		return fmt.Errorf("directory path is required")
	}

	configPath := utils.GetDefaultConfigPath()
	configManager := config.NewConfigManager(configPath)
	cfg, err := configManager.GetConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	blockchainService, err := storage.NewBlockchainService(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize blockchain storage service: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	if shouldCompress {
		log.Info().Str("directory", dirPath).Msg("Starting directory compression and upload to IPFS")
	} else {
		log.Info().Str("directory", dirPath).Msg("Starting directory upload to IPFS")
	}

	cid, err := blockchainService.UploadDirectory(ctx, dirPath)
	if err != nil {
		return fmt.Errorf("failed to upload directory: %w", err)
	}

	dirUrl := blockchainService.GetFileURL(cid)

	log.Info().
		Str("directory", dirPath).
		Str("cid", cid).
		Str("url", dirUrl).
		Msg("Directory uploaded successfully")

	fmt.Printf("Directory uploaded successfully!\n")
	if shouldCompress {
		fmt.Printf("Directory was compressed before upload\n")
	}
	fmt.Printf("CID: %s\n", cid)
	fmt.Printf("URL: %s\n", dirUrl)
	fmt.Printf("\nUse this CID when creating federated learning sessions:\n")
	fmt.Printf("parity-client fl create-session --dataset-cid %s [other options]\n", cid)

	return nil
}

var downloadFileCmd = &cobra.Command{
	Use:   "download [cid] [output-path]",
	Short: "Download a file from IPFS",
	Args:  cobra.ExactArgs(2),
	Example: `  # Download a file by CID
  parity-client storage download QmX5Y... ./downloaded-file.csv
  
  # Download to specific directory
  parity-client storage download QmX5Y... ./data/dataset.zip`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runDownloadFile(cmd, args); err != nil {
			log := gologger.Get()
			log.Error().Err(err).Msg("Download file failed")
			os.Exit(1)
		}
	},
}

func runDownloadFile(cmd *cobra.Command, args []string) error {
	log := gologger.Get()

	cid := args[0]
	outputPath := args[1]

	if err := utils.ValidateDatasetCID(cid); err != nil {
		return fmt.Errorf("invalid CID: %w", err)
	}

	if outputPath == "" {
		return fmt.Errorf("output path is required")
	}

	configPath := utils.GetDefaultConfigPath()
	configManager := config.NewConfigManager(configPath)
	cfg, err := configManager.GetConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	blockchainService, err := storage.NewBlockchainService(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize blockchain storage service: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	log.Info().
		Str("cid", cid).
		Str("output", outputPath).
		Msg("Starting file download from IPFS")

	err = blockchainService.DownloadFile(ctx, cid, outputPath)
	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}

	log.Info().
		Str("cid", cid).
		Str("output", outputPath).
		Msg("File downloaded successfully")

	fmt.Printf("File downloaded successfully to: %s\n", outputPath)
	return nil
}

var fileInfoCmd = &cobra.Command{
	Use:   "info [cid]",
	Short: "Get information about a file stored on IPFS",
	Args:  cobra.ExactArgs(1),
	Example: `  # Get file information
  parity-client storage info QmX5Y...`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runFileInfo(cmd, args); err != nil {
			log := gologger.Get()
			log.Error().Err(err).Msg("Get file info failed")
			os.Exit(1)
		}
	},
}

func runFileInfo(cmd *cobra.Command, args []string) error {
	cid := args[0]

	configPath := utils.GetDefaultConfigPath()
	configManager := config.NewConfigManager(configPath)
	cfg, err := configManager.GetConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	blockchainService, err := storage.NewBlockchainService(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize blockchain storage service: %w", err)
	}

	info, err := blockchainService.GetFileInfo(cid)
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	fmt.Printf("File Information\n")
	fmt.Printf("CID: %s\n", cid)
	fmt.Printf("Size: %d bytes\n", info.CumulativeSize)
	fmt.Printf("Blocks: %d\n", info.NumLinks)
	fmt.Printf("URL: %s\n", blockchainService.GetFileURL(cid))

	return nil
}

var pinFileCmd = &cobra.Command{
	Use:   "pin [cid]",
	Short: "Pin a file to keep it available on IPFS",
	Args:  cobra.ExactArgs(1),
	Example: `  # Pin a file
  parity-client storage pin QmX5Y...`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runPinFile(cmd, args); err != nil {
			log := gologger.Get()
			log.Error().Err(err).Msg("Pin file failed")
			os.Exit(1)
		}
	},
}

func runPinFile(cmd *cobra.Command, args []string) error {
	cid := args[0]

	configPath := utils.GetDefaultConfigPath()
	configManager := config.NewConfigManager(configPath)
	cfg, err := configManager.GetConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	blockchainService, err := storage.NewBlockchainService(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize blockchain storage service: %w", err)
	}

	err = blockchainService.PinFile(cid)
	if err != nil {
		return fmt.Errorf("failed to pin file: %w", err)
	}

	fmt.Printf("File pinned successfully: %s\n", cid)
	return nil
}

var unpinFileCmd = &cobra.Command{
	Use:   "unpin [cid]",
	Short: "Unpin a file from IPFS",
	Args:  cobra.ExactArgs(1),
	Example: `  # Unpin a file
  parity-client storage unpin QmX5Y...`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runUnpinFile(cmd, args); err != nil {
			log := gologger.Get()
			log.Error().Err(err).Msg("Unpin file failed")
			os.Exit(1)
		}
	},
}

func runUnpinFile(cmd *cobra.Command, args []string) error {
	cid := args[0]

	configPath := utils.GetDefaultConfigPath()
	configManager := config.NewConfigManager(configPath)
	cfg, err := configManager.GetConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	blockchainService, err := storage.NewBlockchainService(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize blockchain storage service: %w", err)
	}

	err = blockchainService.UnpinFile(cid)
	if err != nil {
		return fmt.Errorf("failed to unpin file: %w", err)
	}

	fmt.Printf("File unpinned successfully: %s\n", cid)
	return nil
}

func init() {
	uploadFileCmd.Flags().StringP("filename", "n", "", "Custom filename for the uploaded file")
	uploadFileCmd.Flags().StringP("name", "", "", "Custom name for the uploaded file")
	uploadFileCmd.Flags().BoolP("pin", "", false, "Pin the file to IPFS after upload")

	uploadDirectoryCmd.Flags().BoolP("compress", "", false, "Compress directory before upload")

	storageCmd.AddCommand(uploadFileCmd)
	storageCmd.AddCommand(uploadDirectoryCmd)
	storageCmd.AddCommand(downloadFileCmd)
	storageCmd.AddCommand(fileInfoCmd)
	storageCmd.AddCommand(pinFileCmd)
	storageCmd.AddCommand(unpinFileCmd)
}
