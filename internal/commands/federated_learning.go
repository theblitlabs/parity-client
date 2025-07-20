package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	walletsdk "github.com/theblitlabs/go-wallet-sdk"
	"github.com/theblitlabs/gologger"
	"github.com/theblitlabs/parity-client/internal/adapters/keystore"
	"github.com/theblitlabs/parity-client/internal/adapters/wallet"
	"github.com/theblitlabs/parity-client/internal/client"
	"github.com/theblitlabs/parity-client/internal/config"
	"github.com/theblitlabs/parity-client/internal/storage"
	"github.com/theblitlabs/parity-client/internal/utils"
)

var flCmd = &cobra.Command{
	Use:   "fl",
	Short: "Federated Learning operations",
	Long:  `Manage federated learning sessions, join sessions, and coordinate distributed training`,
}

var createSessionCmd = &cobra.Command{
	Use:   "create-session",
	Short: "Create a new federated learning session",
	Long:  `Create a new federated learning session with specified configuration and training parameters`,
	Example: `  # Create a basic session
  parity-client fl create-session --name "Image Classification" --model-type cnn --dataset-cid QmXXX...
  
  # Create session with custom configuration
  parity-client fl create-session --name "Advanced Training" --model-type transformer --total-rounds 10 --min-participants 5`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Get required flags
		name, _ := cmd.Flags().GetString("name")
		modelType, _ := cmd.Flags().GetString("model-type")
		datasetCID, _ := cmd.Flags().GetString("dataset-cid")

		// Get creator address from wallet
		creatorAddress, err := getCreatorAddress(cfg)
		if err != nil {
			return err
		}

		// Validate required fields
		if name == "" {
			return fmt.Errorf("--name is required")
		}
		if modelType == "" {
			return fmt.Errorf("--model-type is required")
		}
		if datasetCID == "" {
			return fmt.Errorf("--dataset-cid is required")
		}

		// Get optional flags
		description, _ := cmd.Flags().GetString("description")
		totalRounds, _ := cmd.Flags().GetInt("total-rounds")
		minParticipants, _ := cmd.Flags().GetInt("min-participants")
		aggregationMethod, _ := cmd.Flags().GetString("aggregation-method")
		learningRate, _ := cmd.Flags().GetFloat64("learning-rate")
		batchSize, _ := cmd.Flags().GetInt("batch-size")
		localEpochs, _ := cmd.Flags().GetInt("local-epochs")
		dataFormat, _ := cmd.Flags().GetString("data-format")
		splitStrategy, _ := cmd.Flags().GetString("split-strategy")
		configFile, _ := cmd.Flags().GetString("config-file")
		enableDP, _ := cmd.Flags().GetBool("enable-differential-privacy")
		noiseMultiplier, _ := cmd.Flags().GetFloat64("noise-multiplier")
		l2NormClip, _ := cmd.Flags().GetFloat64("l2-norm-clip")

		// Get data partitioning parameters
		alpha, _ := cmd.Flags().GetFloat64("alpha")
		minSamples, _ := cmd.Flags().GetInt("min-samples")
		overlapRatio, _ := cmd.Flags().GetFloat64("overlap-ratio")

		// Initialize client
		flClient, err := createFLClient(cfg)
		if err != nil {
			return err
		}

		// Load custom config if provided
		modelConfig := map[string]interface{}{}
		if configFile != "" {
			if data, err := os.ReadFile(configFile); err == nil {
				if err := json.Unmarshal(data, &modelConfig); err != nil {
					fmt.Printf("Failed to parse config file %s: %v\n", configFile, err)
				} else {
					fmt.Printf("Custom model config loaded from: %s\n", configFile)
				}
			} else {
				fmt.Printf("Failed to load custom config from %s\n", configFile)
			}
		}

		// Only set model config from what's explicitly provided
		if len(modelConfig) == 0 {
			return fmt.Errorf("model configuration is required - please provide via --config-file flag")
		}

		// Add differential privacy settings if enabled
		if enableDP {
			modelConfig["differential_privacy"] = map[string]interface{}{
				"enabled":          true,
				"noise_multiplier": noiseMultiplier,
				"l2_norm_clip":     l2NormClip,
			}
			fmt.Printf("Differential privacy enabled (noise: %.2f, clip: %.2f)\n", noiseMultiplier, l2NormClip)
		}

		req := &client.CreateFLSessionRequest{
			Name:            name,
			Description:     description,
			ModelType:       modelType,
			TotalRounds:     totalRounds,
			MinParticipants: minParticipants,
			CreatorAddress:  creatorAddress,
			TrainingData: client.TrainingDataInfo{
				DatasetCID:    datasetCID,
				DataFormat:    dataFormat,
				SplitStrategy: splitStrategy,
				Features:      []string{"feature1", "feature2", "feature3"},
				Labels:        []string{"label"},
				Metadata: map[string]interface{}{
					"alpha":         alpha,
					"min_samples":   minSamples,
					"overlap_ratio": overlapRatio,
				},
			},
			Config: client.FLConfigRequest{
				AggregationMethod: aggregationMethod,
				LearningRate:      learningRate,
				BatchSize:         batchSize,
				LocalEpochs:       localEpochs,
				ClientSelection:   "random",
				ModelConfig:       modelConfig,
			},
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		session, err := flClient.CreateSession(ctx, req)
		if err != nil {
			return fmt.Errorf("failed to create session: %w", err)
		}

		printSessionCreated(session)
		return nil
	},
}

var listSessionsCmd = &cobra.Command{
	Use:   "list-sessions",
	Short: "List federated learning sessions",
	Long:  `List all federated learning sessions, optionally filtered by creator address`,
	Example: `  # List all sessions
  parity-client fl list-sessions
  
  # List sessions by creator
  parity-client fl list-sessions --creator-address 0x123...`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		creator, _ := cmd.Flags().GetString("creator-address")

		flClient, err := createFLClient(cfg)
		if err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		response, err := flClient.ListSessions(ctx, creator)
		if err != nil {
			return fmt.Errorf("failed to list sessions: %w", err)
		}

		printSessionsList(response)
		return nil
	},
}

var getSessionCmd = &cobra.Command{
	Use:   "get-session [session-id]",
	Short: "Get details of a federated learning session",
	Long:  `Get detailed information about a specific federated learning session including status, rounds, and configuration`,
	Args:  cobra.ExactArgs(1),
	Example: `  # Get session details
  parity-client fl get-session 3fe77346-c2dd-4759-b6dc-a8daa29b0991`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		sessionID := args[0]
		if sessionID == "" {
			return fmt.Errorf("session ID is required")
		}

		flClient, err := createFLClient(cfg)
		if err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		session, err := flClient.GetSession(ctx, sessionID)
		if err != nil {
			return fmt.Errorf("failed to get session: %w", err)
		}

		printSessionDetails(session)
		return nil
	},
}

var startSessionCmd = &cobra.Command{
	Use:   "start-session [session-id]",
	Short: "Start a federated learning session",
	Long:  `Start a federated learning session and begin the first training round`,
	Args:  cobra.ExactArgs(1),
	Example: `  # Start a session
  parity-client fl start-session 3fe77346-c2dd-4759-b6dc-a8daa29b0991`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		sessionID := args[0]
		if sessionID == "" {
			return fmt.Errorf("session ID is required")
		}

		flClient, err := createFLClient(cfg)
		if err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := flClient.StartSession(ctx, sessionID); err != nil {
			return fmt.Errorf("failed to start session: %w", err)
		}

		printSessionStarted(sessionID)
		return nil
	},
}

var getModelCmd = &cobra.Command{
	Use:   "get-model [session-id]",
	Short: "Get the trained model from a federated learning session",
	Long:  `Retrieve the trained model from a completed federated learning session and optionally save to file`,
	Args:  cobra.ExactArgs(1),
	Example: `  # Display model in console
  parity-client fl get-model 3fe77346-c2dd-4759-b6dc-a8daa29b0991
  
  # Save model to file
  parity-client fl get-model 3fe77346-c2dd-4759-b6dc-a8daa29b0991 --output model.json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		sessionID := args[0]
		if sessionID == "" {
			return fmt.Errorf("session ID is required")
		}

		outputFile, _ := cmd.Flags().GetString("output")
		format, _ := cmd.Flags().GetString("format")

		flClient, err := createFLClient(cfg)
		if err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		model, err := flClient.GetTrainedModel(ctx, sessionID)
		if err != nil {
			return fmt.Errorf("failed to get trained model: %w", err)
		}

		if outputFile != "" {
			if err := saveModelToFile(model, outputFile, format); err != nil {
				return err
			}
			fmt.Printf("Trained model saved to: %s\n", outputFile)
		} else {
			printModelDetails(model, sessionID)
		}

		return nil
	},
}

var submitUpdateCmd = &cobra.Command{
	Use:   "submit-update",
	Short: "Submit model update to federated learning session",
	Long:  `Submit local training results as model updates to a federated learning session`,
	Example: `  # Submit update with gradients file
  parity-client fl submit-update --session-id abc123 --round-id round1 --runner-id runner1 --gradients-file gradients.json
  
  # Submit update with inline data
  parity-client fl submit-update --session-id abc123 --round-id round1 --runner-id runner1 --data-size 1000 --loss 0.5 --accuracy 0.85`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Get required flags
		sessionID, _ := cmd.Flags().GetString("session-id")
		roundID, _ := cmd.Flags().GetString("round-id")
		runnerID, _ := cmd.Flags().GetString("runner-id")
		gradientsFile, _ := cmd.Flags().GetString("gradients-file")
		dataSize, _ := cmd.Flags().GetInt("data-size")
		loss, _ := cmd.Flags().GetFloat64("loss")
		accuracy, _ := cmd.Flags().GetFloat64("accuracy")

		// Validate required fields
		if sessionID == "" {
			return fmt.Errorf("--session-id is required")
		}
		if roundID == "" {
			return fmt.Errorf("--round-id is required")
		}
		if runnerID == "" {
			return fmt.Errorf("--runner-id is required")
		}
		if gradientsFile == "" {
			return fmt.Errorf("--gradients-file is required")
		}

		var gradients map[string][]float64
		data, err := os.ReadFile(gradientsFile)
		if err != nil {
			return fmt.Errorf("failed to read gradients file: %w", err)
		}
		if err := json.Unmarshal(data, &gradients); err != nil {
			return fmt.Errorf("failed to parse gradients file: %w", err)
		}

		flClient, err := createFLClient(cfg)
		if err != nil {
			return err
		}

		req := &client.SubmitModelUpdateRequest{
			SessionID:    sessionID,
			RoundID:      roundID,
			RunnerID:     runnerID,
			Gradients:    gradients,
			UpdateType:   "gradients",
			DataSize:     dataSize,
			Loss:         loss,
			Accuracy:     accuracy,
			TrainingTime: 2000,
			Metadata: map[string]interface{}{
				"submission_time": time.Now().Unix(),
			},
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := flClient.SubmitModelUpdate(ctx, req); err != nil {
			return fmt.Errorf("failed to submit model update: %w", err)
		}

		printUpdateSubmitted(req)
		return nil
	},
}

var createSessionWithDataCmd = &cobra.Command{
	Use:   "create-session-with-data [data-path]",
	Short: "Upload training data and create federated learning session",
	Long:  `Upload training data to IPFS and create a federated learning session in one step`,
	Args:  cobra.ExactArgs(1),
	Example: `  # Upload dataset file and create session
  parity-client fl create-session-with-data ./dataset.csv --name "Image Classification" --model-type cnn --total-rounds 5
  
  # Upload dataset directory and create session  
  parity-client fl create-session-with-data ./training_data/ --name "NLP Training" --model-type transformer`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		dataPath := args[0]
		if _, err := os.Stat(dataPath); os.IsNotExist(err) {
			return fmt.Errorf("data path does not exist: %s", dataPath)
		}

		// Get required flags
		name, _ := cmd.Flags().GetString("name")
		modelType, _ := cmd.Flags().GetString("model-type")
		totalRounds, _ := cmd.Flags().GetInt("total-rounds")

		// Get creator address from wallet
		creatorAddress, err := getCreatorAddress(cfg)
		if err != nil {
			return err
		}

		// Validate required fields
		if name == "" {
			return fmt.Errorf("--name is required")
		}
		if modelType == "" {
			return fmt.Errorf("--model-type is required")
		}
		if totalRounds == 0 {
			return fmt.Errorf("--total-rounds is required and must be > 0")
		}

		// Get optional flags
		description, _ := cmd.Flags().GetString("description")
		minParticipants, _ := cmd.Flags().GetInt("min-participants")
		dataFormat, _ := cmd.Flags().GetString("data-format")
		splitStrategy, _ := cmd.Flags().GetString("split-strategy")
		configFile, _ := cmd.Flags().GetString("config-file")
		enableDP, _ := cmd.Flags().GetBool("enable-differential-privacy")
		noiseMultiplier, _ := cmd.Flags().GetFloat64("noise-multiplier")
		l2NormClip, _ := cmd.Flags().GetFloat64("l2-norm-clip")

		// Get data partitioning parameters
		alpha, _ := cmd.Flags().GetFloat64("alpha")
		minSamples, _ := cmd.Flags().GetInt("min-samples")
		overlapRatio, _ := cmd.Flags().GetFloat64("overlap-ratio")

		// Initialize storage service
		filecoinService, err := storage.NewBlockchainService(cfg)
		if err != nil {
			return fmt.Errorf("failed to initialize blockchain storage service: %w", err)
		}

		// Upload data to IPFS
		log := gologger.Get()
		log.Info().Str("path", dataPath).Msg("Uploading training data to IPFS")

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
		defer cancel()

		var cid string
		var datasetSize int64
		if info, err := os.Stat(dataPath); err == nil && info.IsDir() {
			cid, err = filecoinService.UploadDirectory(ctx, dataPath)
			if err != nil {
				return fmt.Errorf("failed to upload directory: %w", err)
			}
			log.Info().Str("directory", dataPath).Str("cid", cid).Msg("Directory uploaded successfully")

			// Calculate directory size
			if err := filepath.Walk(dataPath, func(path string, info os.FileInfo, err error) error {
				if err == nil && !info.IsDir() {
					datasetSize += info.Size()
				}
				return nil
			}); err != nil {
				log.Warn().Err(err).Str("path", dataPath).Msg("Failed to calculate directory size")
			}
		} else {
			cid, err = filecoinService.UploadFile(ctx, dataPath)
			if err != nil {
				return fmt.Errorf("failed to upload file: %w", err)
			}
			log.Info().Str("file", dataPath).Str("cid", cid).Msg("File uploaded successfully")
			datasetSize = info.Size()
		}

		// Initialize FL client
		flClient, err := createFLClient(cfg)
		if err != nil {
			return err
		}

		// Load custom config if provided
		modelConfig := map[string]interface{}{}
		if configFile != "" {
			if data, err := os.ReadFile(configFile); err == nil {
				if err := json.Unmarshal(data, &modelConfig); err != nil {
					log.Warn().Err(err).Str("config_file", configFile).Msg("Failed to parse config file")
				} else {
					log.Info().Str("config_file", configFile).Msg("Custom model config loaded")
				}
			} else {
				log.Warn().Err(err).Str("config_file", configFile).Msg("Failed to load custom config, using defaults")
			}
		}

		// Only set model config from what's explicitly provided
		if len(modelConfig) == 0 {
			return fmt.Errorf("model configuration is required - please provide via --config-file flag")
		}

		// Add differential privacy settings if enabled
		if enableDP {
			modelConfig["differential_privacy"] = map[string]interface{}{
				"enabled":          true,
				"noise_multiplier": noiseMultiplier,
				"l2_norm_clip":     l2NormClip,
			}
			log.Info().Float64("noise_multiplier", noiseMultiplier).Float64("l2_norm_clip", l2NormClip).Msg("Differential privacy enabled")
		}

		// Get aggregation method from flags
		aggregationMethod, _ := cmd.Flags().GetString("aggregation-method")
		learningRate, _ := cmd.Flags().GetFloat64("learning-rate")
		batchSize, _ := cmd.Flags().GetInt("batch-size")
		localEpochs, _ := cmd.Flags().GetInt("local-epochs")

		// Create session request with uploaded data CID
		request := &client.CreateFLSessionRequest{
			Name:            name,
			Description:     description,
			ModelType:       modelType,
			TotalRounds:     totalRounds,
			MinParticipants: minParticipants,
			CreatorAddress:  creatorAddress,
			TrainingData: client.TrainingDataInfo{
				DatasetCID:    cid,
				DatasetSize:   datasetSize,
				DataFormat:    dataFormat,
				SplitStrategy: splitStrategy,
				Metadata: map[string]interface{}{
					"alpha":         alpha,
					"min_samples":   minSamples,
					"overlap_ratio": overlapRatio,
				},
			},
			Config: client.FLConfigRequest{
				AggregationMethod: aggregationMethod,
				LearningRate:      learningRate,
				BatchSize:         batchSize,
				LocalEpochs:       localEpochs,
				ClientSelection:   "random",
				ModelConfig:       modelConfig,
			},
		}

		log.Info().Str("cid", cid).Msg("Creating federated learning session with uploaded data")

		session, err := flClient.CreateSession(context.Background(), request)
		if err != nil {
			return fmt.Errorf("failed to create federated learning session: %w", err)
		}

		printSessionWithDataCreated(session, cid, filecoinService.GetFileURL(cid))
		return nil
	},
}

// Helper functions for consistent behavior
func loadConfig() (*config.Config, error) {
	configPath := utils.GetDefaultConfigPath()
	configManager := config.NewConfigManager(configPath)
	return configManager.GetConfig()
}

func createFLClient(cfg *config.Config) (*client.FederatedLearningClient, error) {
	serverURL := cfg.FederatedLearning.ServerURL
	if serverURL == "" {
		return nil, fmt.Errorf("federated learning server URL not configured")
	}
	return client.NewFederatedLearningClient(serverURL), nil
}

func getCreatorAddress(cfg *config.Config) (string, error) {
	keystoreAdapter, err := keystore.NewAdapter(nil) // Uses default config
	if err != nil {
		return "", fmt.Errorf("failed to create keystore: %w", err)
	}

	privateKey, err := keystoreAdapter.GetPrivateKeyHex()
	if err != nil {
		return "", fmt.Errorf("failed to load private key: %w", err)
	}

	walletAdapter, err := wallet.NewAdapter(walletsdk.ClientConfig{
		RPCURL:       cfg.BlockchainNetwork.RPC,
		ChainID:      cfg.BlockchainNetwork.ChainID,
		PrivateKey:   privateKey,
		TokenAddress: ethcommon.HexToAddress(cfg.BlockchainNetwork.TokenAddress),
	})
	if err != nil {
		return "", fmt.Errorf("failed to create wallet adapter: %w", err)
	}

	address := walletAdapter.GetAddress()
	if (address == ethcommon.Address{}) {
		return "", fmt.Errorf("no address found in wallet. Please authenticate first using 'parity-client auth'")
	}

	return address.Hex(), nil
}

func printSessionCreated(session *client.FLSession) {
	fmt.Printf("Federated Learning Session Created\n\n")
	fmt.Printf("Session ID:      %s\n", session.ID)
	fmt.Printf("Name:            %s\n", session.Name)
	fmt.Printf("Description:     %s\n", session.Description)
	fmt.Printf("Model Type:      %s\n", session.ModelType)
	fmt.Printf("Status:          %s\n", session.Status)
	fmt.Printf("Total Rounds:    %d\n", session.TotalRounds)
	fmt.Printf("Min Participants: %d\n", session.MinParticipants)
	fmt.Printf("Creator:         %s\n", session.CreatorAddress)
	fmt.Printf("Created At:      %s\n", session.CreatedAt)
}

func printSessionsList(response *client.ListSessionsResponse) {
	if len(response.Sessions) == 0 {
		fmt.Println("No federated learning sessions found")
		return
	}

	fmt.Printf("Federated Learning Sessions (%d):\n\n", response.Count)
	for i, session := range response.Sessions {
		fmt.Printf("%d. %s\n", i+1, session.Name)
		fmt.Printf("   ID:              %s\n", session.ID)
		fmt.Printf("   Model Type:      %s\n", session.ModelType)
		fmt.Printf("   Status:          %s\n", session.Status)
		fmt.Printf("   Round:           %d/%d\n", session.CurrentRound, session.TotalRounds)
		fmt.Printf("   Min Participants: %d\n", session.MinParticipants)
		fmt.Printf("   Creator:         %s\n", session.CreatorAddress)
		fmt.Printf("   Created:         %s\n", session.CreatedAt)
		if session.CompletedAt != nil {
			fmt.Printf("   Completed:       %s\n", *session.CompletedAt)
		}
		fmt.Println()
	}
}

func printSessionDetails(session *client.FLSession) {
	fmt.Printf("Federated Learning Session Details\n\n")
	fmt.Printf("Session ID:       %s\n", session.ID)
	fmt.Printf("Name:             %s\n", session.Name)
	fmt.Printf("Description:      %s\n", session.Description)
	fmt.Printf("Model Type:       %s\n", session.ModelType)
	fmt.Printf("Status:           %s\n", session.Status)
	fmt.Printf("Current Round:    %d/%d\n", session.CurrentRound, session.TotalRounds)
	fmt.Printf("Min Participants: %d\n", session.MinParticipants)
	fmt.Printf("Participant Count: %d\n", session.ParticipantCount)
	fmt.Printf("Creator:          %s\n", session.CreatorAddress)
	fmt.Printf("Created At:       %s\n", session.CreatedAt)
	fmt.Printf("Updated At:       %s\n", session.UpdatedAt)

	if session.CompletedAt != nil {
		fmt.Printf("Completed At:     %s\n", *session.CompletedAt)
	}

	fmt.Printf("\nConfiguration:\n")
	fmt.Printf("  Aggregation Method: %s\n", session.Config.AggregationMethod)
	fmt.Printf("  Learning Rate:      %f\n", session.Config.LearningRate)
	fmt.Printf("  Batch Size:         %d\n", session.Config.BatchSize)
	fmt.Printf("  Local Epochs:       %d\n", session.Config.LocalEpochs)
	fmt.Printf("  Client Selection:   %s\n", session.Config.ClientSelection)
}

func printSessionStarted(sessionID string) {
	fmt.Printf("Federated learning session started successfully\n")
	fmt.Printf("Session ID: %s\n", sessionID)
	fmt.Printf("Training rounds will begin automatically...\n")
}

func printModelDetails(model map[string]interface{}, sessionID string) {
	fmt.Printf("Trained Model for Session: %s\n\n", sessionID)
	if sessionName, ok := model["session_name"].(string); ok {
		fmt.Printf("Session Name:  %s\n", sessionName)
	}
	if modelType, ok := model["model_type"].(string); ok {
		fmt.Printf("Model Type:    %s\n", modelType)
	}
	if status, ok := model["status"].(string); ok {
		fmt.Printf("Status:        %s\n", status)
	}
	if totalRounds, ok := model["total_rounds"].(float64); ok {
		fmt.Printf("Total Rounds:  %.0f\n", totalRounds)
	}
	if completedAt, ok := model["completed_at"].(string); ok && completedAt != "" {
		fmt.Printf("Completed At:  %s\n", completedAt)
	}

	fmt.Printf("\nModel Data:\n")
	modelData, err := json.MarshalIndent(model["model_data"], "", "  ")
	if err != nil {
		fmt.Printf("Error formatting model data: %v\n", err)
		return
	}
	fmt.Println(string(modelData))
}

func saveModelToFile(model map[string]interface{}, outputFile, format string) error {
	var data []byte
	var err error

	switch format {
	case "json":
		data, err = json.MarshalIndent(model, "", "  ")
	case "yaml":
		return fmt.Errorf("YAML format not yet supported")
	default:
		data, err = json.MarshalIndent(model, "", "  ")
	}

	if err != nil {
		return fmt.Errorf("failed to marshal model data: %w", err)
	}

	if err := os.WriteFile(outputFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}

	return nil
}

func printUpdateSubmitted(req *client.SubmitModelUpdateRequest) {
	fmt.Printf("Model update submitted successfully\n")
	fmt.Printf("Session ID:    %s\n", req.SessionID)
	fmt.Printf("Round ID:      %s\n", req.RoundID)
	fmt.Printf("Runner ID:     %s\n", req.RunnerID)
	fmt.Printf("Data Size:     %d\n", req.DataSize)
	fmt.Printf("Loss:          %.4f\n", req.Loss)
	fmt.Printf("Accuracy:      %.4f\n", req.Accuracy)
}

func printSessionWithDataCreated(session *client.FLSession, cid, fileURL string) {
	fmt.Printf("Training data uploaded and federated learning session created successfully!\n\n")
	fmt.Printf("Data CID:      %s\n", cid)
	fmt.Printf("Data URL:      %s\n", fileURL)
	fmt.Printf("Session ID:    %s\n", session.ID)
	fmt.Printf("Session Name:  %s\n", session.Name)
	fmt.Printf("Model Type:    %s\n", session.ModelType)
	fmt.Printf("Total Rounds:  %d\n", session.TotalRounds)
	fmt.Printf("Min Participants: %d\n", session.MinParticipants)
	fmt.Printf("Status:        %s\n", session.Status)
	fmt.Printf("\nStart the session when ready:\n")
	fmt.Printf("parity-client fl start-session %s\n", session.ID)
}

func init() {
	// Create session flags
	createSessionCmd.Flags().StringP("name", "n", "", "Session name (required)")
	createSessionCmd.Flags().StringP("description", "d", "", "Session description")
	createSessionCmd.Flags().StringP("model-type", "m", "", "Model type (required)")
	createSessionCmd.Flags().IntP("total-rounds", "r", 10, "Total training rounds")
	createSessionCmd.Flags().IntP("min-participants", "p", 1, "Minimum participants")
	createSessionCmd.Flags().StringP("aggregation-method", "a", "", "Aggregation method (required)")
	createSessionCmd.Flags().Float64P("learning-rate", "l", 0, "Learning rate (required)")
	createSessionCmd.Flags().IntP("batch-size", "b", 0, "Batch size (required)")
	createSessionCmd.Flags().IntP("local-epochs", "e", 0, "Local epochs (required)")
	createSessionCmd.Flags().String("dataset-cid", "", "IPFS dataset CID (required)")
	createSessionCmd.Flags().String("data-format", "csv", "Data format (csv, json, parquet)")
	createSessionCmd.Flags().String("split-strategy", "random", "Data split strategy (random, stratified, sequential, non_iid, label_skew)")
	createSessionCmd.Flags().String("config-file", "", "Custom model configuration file")
	createSessionCmd.Flags().Bool("enable-differential-privacy", false, "Enable differential privacy")
	createSessionCmd.Flags().Float64("noise-multiplier", 0.1, "Noise multiplier for differential privacy")
	createSessionCmd.Flags().Float64("l2-norm-clip", 1.0, "L2 norm clipping for differential privacy")

	// Data partitioning flags for truly distributed FL
	createSessionCmd.Flags().Float64("alpha", 0, "Dirichlet distribution parameter for non-IID partitioning (required for non_iid strategy)")
	createSessionCmd.Flags().Int("min-samples", 0, "Minimum samples per participant (required)")
	createSessionCmd.Flags().Float64("overlap-ratio", 0, "Data overlap ratio between participants (0.0 = no overlap, 0.1 = 10% overlap)")

	// List sessions flags
	listSessionsCmd.Flags().StringP("creator-address", "c", "", "Filter by creator address")

	// Get model flags
	getModelCmd.Flags().StringP("output", "o", "", "Output file path (optional - prints to console if not specified)")
	getModelCmd.Flags().StringP("format", "f", "json", "Output format (json)")

	// Submit update flags
	submitUpdateCmd.Flags().StringP("session-id", "s", "", "Session ID (required)")
	submitUpdateCmd.Flags().StringP("round-id", "r", "", "Round ID (required)")
	submitUpdateCmd.Flags().StringP("runner-id", "u", "", "Runner ID (required)")
	submitUpdateCmd.Flags().StringP("gradients-file", "g", "", "Path to gradients JSON file")
	submitUpdateCmd.Flags().IntP("data-size", "d", 1000, "Training data size")
	submitUpdateCmd.Flags().Float64P("loss", "l", 0.0, "Training loss")
	submitUpdateCmd.Flags().Float64P("accuracy", "a", 0.0, "Training accuracy")

	// Create session with data flags
	createSessionWithDataCmd.Flags().StringP("name", "n", "", "Session name (required)")
	createSessionWithDataCmd.Flags().StringP("description", "d", "", "Session description")
	createSessionWithDataCmd.Flags().StringP("model-type", "m", "", "Model type (required)")
	createSessionWithDataCmd.Flags().IntP("total-rounds", "r", 0, "Total training rounds (required)")
	createSessionWithDataCmd.Flags().IntP("min-participants", "p", 1, "Minimum participants")
	createSessionWithDataCmd.Flags().StringP("aggregation-method", "a", "", "Aggregation method (required)")
	createSessionWithDataCmd.Flags().Float64P("learning-rate", "l", 0, "Learning rate (required)")
	createSessionWithDataCmd.Flags().IntP("batch-size", "b", 0, "Batch size (required)")
	createSessionWithDataCmd.Flags().IntP("local-epochs", "e", 0, "Local epochs (required)")
	createSessionWithDataCmd.Flags().String("data-format", "csv", "Data format (csv, json, parquet)")
	createSessionWithDataCmd.Flags().String("split-strategy", "random", "Data split strategy (random, stratified, sequential, non_iid, label_skew)")
	createSessionWithDataCmd.Flags().String("config-file", "", "Custom model configuration file (required)")
	createSessionWithDataCmd.Flags().Bool("enable-differential-privacy", false, "Enable differential privacy")
	createSessionWithDataCmd.Flags().Float64("noise-multiplier", 0.1, "Noise multiplier for differential privacy")
	createSessionWithDataCmd.Flags().Float64("l2-norm-clip", 1.0, "L2 norm clipping for differential privacy")

	// Data partitioning flags for session with data
	createSessionWithDataCmd.Flags().Float64("alpha", 0, "Dirichlet distribution parameter for non-IID partitioning (required for non_iid strategy)")
	createSessionWithDataCmd.Flags().Int("min-samples", 0, "Minimum samples per participant (required)")
	createSessionWithDataCmd.Flags().Float64("overlap-ratio", 0, "Data overlap ratio between participants")

	// Mark required flags
	if err := createSessionCmd.MarkFlagRequired("name"); err != nil {
		log.Error().Err(err).Msg("Failed to mark name flag as required")
	}
	if err := createSessionCmd.MarkFlagRequired("model-type"); err != nil {
		log.Error().Err(err).Msg("Failed to mark model-type flag as required")
	}
	if err := createSessionCmd.MarkFlagRequired("dataset-cid"); err != nil {
		log.Error().Err(err).Msg("Failed to mark dataset-cid flag as required")
	}
	if err := createSessionCmd.MarkFlagRequired("aggregation-method"); err != nil {
		log.Error().Err(err).Msg("Failed to mark aggregation-method flag as required")
	}
	if err := createSessionCmd.MarkFlagRequired("learning-rate"); err != nil {
		log.Error().Err(err).Msg("Failed to mark learning-rate flag as required")
	}
	if err := createSessionCmd.MarkFlagRequired("batch-size"); err != nil {
		log.Error().Err(err).Msg("Failed to mark batch-size flag as required")
	}
	if err := createSessionCmd.MarkFlagRequired("local-epochs"); err != nil {
		log.Error().Err(err).Msg("Failed to mark local-epochs flag as required")
	}
	if err := createSessionCmd.MarkFlagRequired("config-file"); err != nil {
		log.Error().Err(err).Msg("Failed to mark config-file flag as required")
	}
	if err := createSessionCmd.MarkFlagRequired("min-samples"); err != nil {
		log.Error().Err(err).Msg("Failed to mark min-samples flag as required")
	}

	if err := submitUpdateCmd.MarkFlagRequired("session-id"); err != nil {
		log.Error().Err(err).Msg("Failed to mark session-id flag as required")
	}
	if err := submitUpdateCmd.MarkFlagRequired("round-id"); err != nil {
		log.Error().Err(err).Msg("Failed to mark round-id flag as required")
	}
	if err := submitUpdateCmd.MarkFlagRequired("runner-id"); err != nil {
		log.Error().Err(err).Msg("Failed to mark runner-id flag as required")
	}

	if err := createSessionWithDataCmd.MarkFlagRequired("name"); err != nil {
		log.Error().Err(err).Msg("Failed to mark name flag as required")
	}
	if err := createSessionWithDataCmd.MarkFlagRequired("model-type"); err != nil {
		log.Error().Err(err).Msg("Failed to mark model-type flag as required")
	}
	if err := createSessionWithDataCmd.MarkFlagRequired("total-rounds"); err != nil {
		log.Error().Err(err).Msg("Failed to mark total-rounds flag as required")
	}
	if err := createSessionWithDataCmd.MarkFlagRequired("aggregation-method"); err != nil {
		log.Error().Err(err).Msg("Failed to mark aggregation-method flag as required")
	}
	if err := createSessionWithDataCmd.MarkFlagRequired("learning-rate"); err != nil {
		log.Error().Err(err).Msg("Failed to mark learning-rate flag as required")
	}
	if err := createSessionWithDataCmd.MarkFlagRequired("batch-size"); err != nil {
		log.Error().Err(err).Msg("Failed to mark batch-size flag as required")
	}
	if err := createSessionWithDataCmd.MarkFlagRequired("local-epochs"); err != nil {
		log.Error().Err(err).Msg("Failed to mark local-epochs flag as required")
	}
	if err := createSessionWithDataCmd.MarkFlagRequired("config-file"); err != nil {
		log.Error().Err(err).Msg("Failed to mark config-file flag as required")
	}
	if err := createSessionWithDataCmd.MarkFlagRequired("min-samples"); err != nil {
		log.Error().Err(err).Msg("Failed to mark min-samples flag as required")
	}

	// Add subcommands
	flCmd.AddCommand(createSessionCmd)
	flCmd.AddCommand(listSessionsCmd)
	flCmd.AddCommand(getSessionCmd)
	flCmd.AddCommand(startSessionCmd)
	flCmd.AddCommand(getModelCmd)
	flCmd.AddCommand(submitUpdateCmd)
	flCmd.AddCommand(createSessionWithDataCmd)
}

// GetFLCommand returns the federated learning command for integration into the main CLI
func GetFLCommand() *cobra.Command {
	return flCmd
}
