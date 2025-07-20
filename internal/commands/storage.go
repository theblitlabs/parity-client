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
		log := gologger.Get()

		filePath := args[0]
		customName, _ := cmd.Flags().GetString("name")
		shouldPin, _ := cmd.Flags().GetBool("pin")

		// Check if file exists
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			log.Fatal().Str("file", filePath).Msg("File does not exist")
		}

		configPath := utils.GetDefaultConfigPath()
		configManager := config.NewConfigManager(configPath)
		cfg, err := configManager.GetConfig()
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to load config")
		}

		blockchainService, err := storage.NewBlockchainService(cfg)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to initialize blockchain storage service")
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
			log.Fatal().Err(err).Msg("Failed to upload file")
		}

		// Pin file if requested
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
	},
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
		log := gologger.Get()

		dirPath := args[0]
		shouldCompress, _ := cmd.Flags().GetBool("compress")

		// Check if directory exists
		if info, err := os.Stat(dirPath); os.IsNotExist(err) || !info.IsDir() {
			log.Fatal().Str("directory", dirPath).Msg("Directory does not exist")
		}

		configPath := utils.GetDefaultConfigPath()
		configManager := config.NewConfigManager(configPath)
		cfg, err := configManager.GetConfig()
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to load config")
		}

		blockchainService, err := storage.NewBlockchainService(cfg)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to initialize blockchain storage service")
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
			log.Fatal().Err(err).Msg("Failed to upload directory")
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
	},
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
		log := gologger.Get()

		cid := args[0]
		outputPath := args[1]

		configPath := utils.GetDefaultConfigPath()
		configManager := config.NewConfigManager(configPath)
		cfg, err := configManager.GetConfig()
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to load config")
		}

		blockchainService, err := storage.NewBlockchainService(cfg)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to initialize blockchain storage service")
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()

		log.Info().
			Str("cid", cid).
			Str("output", outputPath).
			Msg("Starting file download from IPFS")

		err = blockchainService.DownloadFile(ctx, cid, outputPath)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to download file")
		}

		log.Info().
			Str("cid", cid).
			Str("output", outputPath).
			Msg("File downloaded successfully")

		fmt.Printf("File downloaded successfully to: %s\n", outputPath)
	},
}

var fileInfoCmd = &cobra.Command{
	Use:   "info [cid]",
	Short: "Get information about a file stored on IPFS",
	Args:  cobra.ExactArgs(1),
	Example: `  # Get file information
  parity-client storage info QmX5Y...`,
	Run: func(cmd *cobra.Command, args []string) {
		log := gologger.Get()

		cid := args[0]

		configPath := utils.GetDefaultConfigPath()
		configManager := config.NewConfigManager(configPath)
		cfg, err := configManager.GetConfig()
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to load config")
		}

		blockchainService, err := storage.NewBlockchainService(cfg)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to initialize blockchain storage service")
		}

		info, err := blockchainService.GetFileInfo(cid)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to get file info")
		}

		fmt.Printf("File Information\n")
		fmt.Printf("CID: %s\n", cid)
		fmt.Printf("Size: %d bytes\n", info.CumulativeSize)
		fmt.Printf("Blocks: %d\n", info.NumLinks)
		fmt.Printf("URL: %s\n", blockchainService.GetFileURL(cid))
	},
}

var pinFileCmd = &cobra.Command{
	Use:   "pin [cid]",
	Short: "Pin a file to keep it available on IPFS",
	Args:  cobra.ExactArgs(1),
	Example: `  # Pin a file
  parity-client storage pin QmX5Y...`,
	Run: func(cmd *cobra.Command, args []string) {
		log := gologger.Get()

		cid := args[0]

		configPath := utils.GetDefaultConfigPath()
		configManager := config.NewConfigManager(configPath)
		cfg, err := configManager.GetConfig()
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to load config")
		}

		blockchainService, err := storage.NewBlockchainService(cfg)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to initialize blockchain storage service")
		}

		err = blockchainService.PinFile(cid)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to pin file")
		}

		fmt.Printf("File pinned successfully: %s\n", cid)
	},
}

var unpinFileCmd = &cobra.Command{
	Use:   "unpin [cid]",
	Short: "Unpin a file from IPFS",
	Args:  cobra.ExactArgs(1),
	Example: `  # Unpin a file
  parity-client storage unpin QmX5Y...`,
	Run: func(cmd *cobra.Command, args []string) {
		log := gologger.Get()

		cid := args[0]

		configPath := utils.GetDefaultConfigPath()
		configManager := config.NewConfigManager(configPath)
		cfg, err := configManager.GetConfig()
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to load config")
		}

		blockchainService, err := storage.NewBlockchainService(cfg)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to initialize blockchain storage service")
		}

		err = blockchainService.UnpinFile(cid)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to unpin file")
		}

		fmt.Printf("File unpinned successfully: %s\n", cid)
	},
}

func init() {
	// Add upload-file flags
	uploadFileCmd.Flags().StringP("filename", "n", "", "Custom filename for the uploaded file")
	uploadFileCmd.Flags().StringP("name", "", "", "Custom name for the uploaded file")
	uploadFileCmd.Flags().BoolP("pin", "", false, "Pin the file to IPFS after upload")

	// Add upload-dir flags
	uploadDirectoryCmd.Flags().BoolP("compress", "", false, "Compress directory before upload")

	// Add subcommands
	storageCmd.AddCommand(uploadFileCmd)
	storageCmd.AddCommand(uploadDirectoryCmd)
	storageCmd.AddCommand(downloadFileCmd)
	storageCmd.AddCommand(fileInfoCmd)
	storageCmd.AddCommand(pinFileCmd)
	storageCmd.AddCommand(unpinFileCmd)
}
