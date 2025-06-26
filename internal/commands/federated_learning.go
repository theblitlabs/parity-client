package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/theblitlabs/gologger"
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
	Long:  `Create a new federated learning session with specified configuration`,
	RunE: func(cmd *cobra.Command, args []string) error {
		configPath := utils.GetDefaultConfigPath()
		configManager := config.NewConfigManager(configPath)
		cfg, err := configManager.GetConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		name, _ := cmd.Flags().GetString("name")
		description, _ := cmd.Flags().GetString("description")
		modelType, _ := cmd.Flags().GetString("model-type")
		totalRounds, _ := cmd.Flags().GetInt("total-rounds")
		minParticipants, _ := cmd.Flags().GetInt("min-participants")
		creatorAddress, _ := cmd.Flags().GetString("creator-address")
		aggregationMethod, _ := cmd.Flags().GetString("aggregation-method")
		learningRate, _ := cmd.Flags().GetFloat64("learning-rate")
		batchSize, _ := cmd.Flags().GetInt("batch-size")
		localEpochs, _ := cmd.Flags().GetInt("local-epochs")
		datasetCID, _ := cmd.Flags().GetString("dataset-cid")
		dataFormat, _ := cmd.Flags().GetString("data-format")
		splitStrategy, _ := cmd.Flags().GetString("split-strategy")
		configFile, _ := cmd.Flags().GetString("config-file")
		enableDP, _ := cmd.Flags().GetBool("enable-differential-privacy")
		noiseMultiplier, _ := cmd.Flags().GetFloat64("noise-multiplier")
		l2NormClip, _ := cmd.Flags().GetFloat64("l2-norm-clip")

		if name == "" {
			return fmt.Errorf("session name is required")
		}
		if modelType == "" {
			return fmt.Errorf("model type is required")
		}
		if creatorAddress == "" {
			return fmt.Errorf("creator address is required")
		}
		if datasetCID == "" {
			return fmt.Errorf("dataset CID is required")
		}

		serverURL := fmt.Sprintf("http://%s:%d", cfg.Server.Host, cfg.Server.Port)
		flClient := client.NewFederatedLearningClient(serverURL)

		// Load custom config if provided
		modelConfig := map[string]interface{}{
			"input_size":  784,
			"output_size": 10,
			"hidden_size": 128,
		}
		if configFile != "" {
			if data, err := os.ReadFile(configFile); err == nil {
				json.Unmarshal(data, &modelConfig)
				fmt.Printf("📋 Custom model config loaded from: %s\n", configFile)
			} else {
				fmt.Printf("⚠️ Failed to load custom config from %s, using defaults\n", configFile)
			}
		}

		// Add differential privacy settings if enabled
		if enableDP {
			modelConfig["differential_privacy"] = map[string]interface{}{
				"enabled":          true,
				"noise_multiplier": noiseMultiplier,
				"l2_norm_clip":     l2NormClip,
			}
			fmt.Printf("🔒 Differential privacy enabled (noise: %.2f, clip: %.2f)\n", noiseMultiplier, l2NormClip)
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
			return fmt.Errorf("failed to create FL session: %w", err)
		}

		fmt.Printf("✅ Federated Learning Session Created\n\n")
		fmt.Printf("Session ID: %s\n", session.ID)
		fmt.Printf("Name: %s\n", session.Name)
		fmt.Printf("Description: %s\n", session.Description)
		fmt.Printf("Model Type: %s\n", session.ModelType)
		fmt.Printf("Status: %s\n", session.Status)
		fmt.Printf("Total Rounds: %d\n", session.TotalRounds)
		fmt.Printf("Min Participants: %d\n", session.MinParticipants)
		fmt.Printf("Creator: %s\n", session.CreatorAddress)
		fmt.Printf("Created At: %s\n", session.CreatedAt)

		return nil
	},
}

var listSessionsCmd = &cobra.Command{
	Use:   "list-sessions",
	Short: "List federated learning sessions",
	Long:  `List all federated learning sessions or filter by creator address`,
	RunE: func(cmd *cobra.Command, args []string) error {
		configPath := utils.GetDefaultConfigPath()
		configManager := config.NewConfigManager(configPath)
		cfg, err := configManager.GetConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		creator, _ := cmd.Flags().GetString("creator-address")

		serverURL := fmt.Sprintf("http://%s:%d", cfg.Server.Host, cfg.Server.Port)
		flClient := client.NewFederatedLearningClient(serverURL)

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		response, err := flClient.ListSessions(ctx, creator)
		if err != nil {
			return fmt.Errorf("failed to list FL sessions: %w", err)
		}

		if len(response.Sessions) == 0 {
			fmt.Println("No federated learning sessions found")
			return nil
		}

		fmt.Printf("🔍 Federated Learning Sessions (%d):\n\n", response.Count)
		for i, session := range response.Sessions {
			fmt.Printf("%d. %s\n", i+1, session.Name)
			fmt.Printf("   ID: %s\n", session.ID)
			fmt.Printf("   Model Type: %s\n", session.ModelType)
			fmt.Printf("   Status: %s\n", session.Status)
			fmt.Printf("   Round: %d/%d\n", session.CurrentRound, session.TotalRounds)
			fmt.Printf("   Min Participants: %d\n", session.MinParticipants)
			fmt.Printf("   Creator: %s\n", session.CreatorAddress)
			fmt.Printf("   Created: %s\n", session.CreatedAt)
			if session.CompletedAt != nil {
				fmt.Printf("   Completed: %s\n", *session.CompletedAt)
			}
			fmt.Println()
		}

		return nil
	},
}

var getSessionCmd = &cobra.Command{
	Use:   "get-session",
	Short: "Get details of a federated learning session",
	Long:  `Get detailed information about a specific federated learning session`,
	RunE: func(cmd *cobra.Command, args []string) error {
		configPath := utils.GetDefaultConfigPath()
		configManager := config.NewConfigManager(configPath)
		cfg, err := configManager.GetConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		sessionID, _ := cmd.Flags().GetString("session-id")
		if sessionID == "" {
			return fmt.Errorf("session ID is required")
		}

		serverURL := fmt.Sprintf("http://%s:%d", cfg.Server.Host, cfg.Server.Port)
		flClient := client.NewFederatedLearningClient(serverURL)

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		session, err := flClient.GetSession(ctx, sessionID)
		if err != nil {
			return fmt.Errorf("failed to get FL session: %w", err)
		}

		fmt.Printf("📊 Federated Learning Session Details\n\n")
		fmt.Printf("Session ID: %s\n", session.ID)
		fmt.Printf("Name: %s\n", session.Name)
		fmt.Printf("Description: %s\n", session.Description)
		fmt.Printf("Model Type: %s\n", session.ModelType)
		fmt.Printf("Status: %s\n", session.Status)
		fmt.Printf("Current Round: %d/%d\n", session.CurrentRound, session.TotalRounds)
		fmt.Printf("Min Participants: %d\n", session.MinParticipants)
		fmt.Printf("Participant Count: %d\n", session.ParticipantCount)
		fmt.Printf("Creator: %s\n", session.CreatorAddress)
		fmt.Printf("Created At: %s\n", session.CreatedAt)
		fmt.Printf("Updated At: %s\n", session.UpdatedAt)

		if session.CompletedAt != nil {
			fmt.Printf("Completed At: %s\n", *session.CompletedAt)
		}

		// Display configuration
		fmt.Printf("\nConfiguration:\n")
		fmt.Printf("  Aggregation Method: %s\n", session.Config.AggregationMethod)
		fmt.Printf("  Learning Rate: %f\n", session.Config.LearningRate)
		fmt.Printf("  Batch Size: %d\n", session.Config.BatchSize)
		fmt.Printf("  Local Epochs: %d\n", session.Config.LocalEpochs)
		fmt.Printf("  Client Selection: %s\n", session.Config.ClientSelection)

		return nil
	},
}

var startSessionCmd = &cobra.Command{
	Use:   "start-session [session-id]",
	Short: "Start a federated learning session",
	Args:  cobra.ExactArgs(1),
	Long:  `Start a federated learning session and begin the first training round`,
	RunE: func(cmd *cobra.Command, args []string) error {
		configPath := utils.GetDefaultConfigPath()
		configManager := config.NewConfigManager(configPath)
		cfg, err := configManager.GetConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		sessionID := args[0]
		if sessionID == "" {
			return fmt.Errorf("session ID is required")
		}

		serverURL := fmt.Sprintf("http://%s:%d", cfg.Server.Host, cfg.Server.Port)
		flClient := client.NewFederatedLearningClient(serverURL)

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := flClient.StartSession(ctx, sessionID); err != nil {
			return fmt.Errorf("failed to start FL session: %w", err)
		}

		fmt.Printf("✅ Federated learning session started successfully\n")
		fmt.Printf("Session ID: %s\n", sessionID)
		fmt.Printf("Training rounds will begin automatically...\n")

		return nil
	},
}

var submitUpdateCmd = &cobra.Command{
	Use:   "submit-update",
	Short: "Submit model update to federated learning session",
	Long:  `Submit local training results as model updates to a federated learning session`,
	RunE: func(cmd *cobra.Command, args []string) error {
		configPath := utils.GetDefaultConfigPath()
		configManager := config.NewConfigManager(configPath)
		cfg, err := configManager.GetConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		sessionID, _ := cmd.Flags().GetString("session-id")
		roundID, _ := cmd.Flags().GetString("round-id")
		runnerID, _ := cmd.Flags().GetString("runner-id")
		gradientsFile, _ := cmd.Flags().GetString("gradients-file")
		dataSize, _ := cmd.Flags().GetInt("data-size")
		loss, _ := cmd.Flags().GetFloat64("loss")
		accuracy, _ := cmd.Flags().GetFloat64("accuracy")

		if sessionID == "" {
			return fmt.Errorf("session ID is required")
		}
		if roundID == "" {
			return fmt.Errorf("round ID is required")
		}
		if runnerID == "" {
			return fmt.Errorf("runner ID is required")
		}

		var gradients map[string][]float64

		if gradientsFile != "" {
			// Load gradients from file
			data, err := os.ReadFile(gradientsFile)
			if err != nil {
				return fmt.Errorf("failed to read gradients file: %w", err)
			}

			if err := json.Unmarshal(data, &gradients); err != nil {
				return fmt.Errorf("failed to parse gradients file: %w", err)
			}
		} else {
			// Use mock gradients for demonstration
			gradients = map[string][]float64{
				"layer1_weights": {0.1, -0.05, 0.02, 0.08, -0.03},
				"layer1_bias":    {0.01, -0.02},
				"layer2_weights": {-0.02, 0.04, -0.01, 0.03},
				"layer2_bias":    {0.005},
			}
		}

		serverURL := fmt.Sprintf("http://%s:%d", cfg.Server.Host, cfg.Server.Port)
		flClient := client.NewFederatedLearningClient(serverURL)

		req := &client.SubmitModelUpdateRequest{
			SessionID:    sessionID,
			RoundID:      roundID,
			RunnerID:     runnerID,
			Gradients:    gradients,
			UpdateType:   "gradients",
			DataSize:     dataSize,
			Loss:         loss,
			Accuracy:     accuracy,
			TrainingTime: 2000, // Mock training time in ms
			Metadata: map[string]interface{}{
				"local_epochs":  3,
				"batch_size":    32,
				"learning_rate": 0.001,
			},
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := flClient.SubmitModelUpdate(ctx, req); err != nil {
			return fmt.Errorf("failed to submit model update: %w", err)
		}

		fmt.Printf("✅ Model update submitted successfully\n")
		fmt.Printf("Session ID: %s\n", sessionID)
		fmt.Printf("Round ID: %s\n", roundID)
		fmt.Printf("Runner ID: %s\n", runnerID)
		fmt.Printf("Data Size: %d\n", dataSize)
		fmt.Printf("Loss: %.4f\n", loss)
		fmt.Printf("Accuracy: %.4f\n", accuracy)

		return nil
	},
}

var createSessionWithDataCmd = &cobra.Command{
	Use:   "create-session-with-data [data-path]",
	Short: "Upload training data and create federated learning session",
	Args:  cobra.ExactArgs(1),
	Example: `  # Upload dataset file and create session
  parity-client fl create-session-with-data ./dataset.csv --name "Image Classification" --model-type cnn --total-rounds 5
  
  # Upload dataset directory and create session  
  parity-client fl create-session-with-data ./training_data/ --name "NLP Training" --model-type transformer`,
	Run: func(cmd *cobra.Command, args []string) {
		log := gologger.Get()

		dataPath := args[0]

		// Check if path exists
		if _, err := os.Stat(dataPath); os.IsNotExist(err) {
			log.Fatal().Str("path", dataPath).Msg("Data path does not exist")
		}

		configPath := utils.GetDefaultConfigPath()
		configManager := config.NewConfigManager(configPath)
		cfg, err := configManager.GetConfig()
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to load config")
		}

		// Initialize storage service
		filecoinService, err := storage.NewFilecoinService(cfg)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to initialize Filecoin service")
		}

		// Upload data to IPFS/Filecoin
		log.Info().Str("path", dataPath).Msg("Uploading training data to IPFS/Filecoin")

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
		defer cancel()

		var cid string
		if info, err := os.Stat(dataPath); err == nil && info.IsDir() {
			cid, err = filecoinService.UploadDirectory(ctx, dataPath)
			if err != nil {
				log.Fatal().Err(err).Msg("Failed to upload directory")
			}
			log.Info().Str("directory", dataPath).Str("cid", cid).Msg("Directory uploaded successfully")
		} else {
			cid, err = filecoinService.UploadFile(ctx, dataPath)
			if err != nil {
				log.Fatal().Err(err).Msg("Failed to upload file")
			}
			log.Info().Str("file", dataPath).Str("cid", cid).Msg("File uploaded successfully")
		}

		// Get session parameters
		name, _ := cmd.Flags().GetString("name")
		description, _ := cmd.Flags().GetString("description")
		modelType, _ := cmd.Flags().GetString("model-type")
		totalRounds, _ := cmd.Flags().GetInt("total-rounds")
		minParticipants, _ := cmd.Flags().GetInt("min-participants")
		creatorAddress, _ := cmd.Flags().GetString("creator-address")
		dataFormat, _ := cmd.Flags().GetString("data-format")
		splitStrategy, _ := cmd.Flags().GetString("split-strategy")
		configFile, _ := cmd.Flags().GetString("config-file")
		enableDP, _ := cmd.Flags().GetBool("enable-differential-privacy")
		noiseMultiplier, _ := cmd.Flags().GetFloat64("noise-multiplier")
		l2NormClip, _ := cmd.Flags().GetFloat64("l2-norm-clip")

		// Use default creator address if not provided
		if creatorAddress == "" {
			creatorAddress = "0x0000000000000000000000000000000000000000" // Default placeholder
		}

		// Set dataset size (estimate based on upload)
		var datasetSize int64
		if info, err := os.Stat(dataPath); err == nil {
			if info.IsDir() {
				// For directories, walk through and sum file sizes
				filepath.Walk(dataPath, func(path string, info os.FileInfo, err error) error {
					if err == nil && !info.IsDir() {
						datasetSize += info.Size()
					}
					return nil
				})
			} else {
				datasetSize = info.Size()
			}
		}

		// Initialize FL client
		serverURL := fmt.Sprintf("http://%s:%d", cfg.Server.Host, cfg.Server.Port)
		flClient := client.NewFederatedLearningClient(serverURL)

		// Load custom config if provided
		modelConfig := map[string]interface{}{
			"input_size":  784,
			"output_size": 10,
			"hidden_size": 128,
		}
		if configFile != "" {
			if data, err := os.ReadFile(configFile); err == nil {
				json.Unmarshal(data, &modelConfig)
				log.Info().Str("config_file", configFile).Msg("Custom model config loaded")
			} else {
				log.Warn().Err(err).Str("config_file", configFile).Msg("Failed to load custom config, using defaults")
			}
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
			},
			Config: client.FLConfigRequest{
				AggregationMethod: "federated_averaging",
				LearningRate:      0.001,
				BatchSize:         32,
				LocalEpochs:       3,
				ClientSelection:   "random",
				ModelConfig:       modelConfig,
			},
		}

		log.Info().Str("cid", cid).Msg("Creating federated learning session with uploaded data")

		session, err := flClient.CreateSession(context.Background(), request)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to create federated learning session")
		}

		fmt.Printf("✅ Training data uploaded and federated learning session created successfully!\n")
		fmt.Printf("📄 Data CID: %s\n", cid)
		fmt.Printf("📄 Data URL: %s\n", filecoinService.GetFileURL(cid))
		fmt.Printf("🔧 Session ID: %s\n", session.ID)
		fmt.Printf("📝 Session Name: %s\n", session.Name)
		fmt.Printf("🤖 Model Type: %s\n", session.ModelType)
		fmt.Printf("🔄 Total Rounds: %d\n", session.TotalRounds)
		fmt.Printf("👥 Min Participants: %d\n", session.MinParticipants)
		fmt.Printf("📊 Status: %s\n", session.Status)
		fmt.Printf("\n💡 Start the session when ready:\n")
		fmt.Printf("parity-client fl start-session %s\n", session.ID)
	},
}

func init() {
	// Create session flags
	createSessionCmd.Flags().StringP("name", "n", "", "Session name (required)")
	createSessionCmd.Flags().StringP("description", "d", "", "Session description")
	createSessionCmd.Flags().StringP("model-type", "m", "", "Model type (required)")
	createSessionCmd.Flags().IntP("total-rounds", "r", 10, "Total training rounds")
	createSessionCmd.Flags().IntP("min-participants", "p", 2, "Minimum participants")
	createSessionCmd.Flags().StringP("creator-address", "c", "", "Creator wallet address (required)")
	createSessionCmd.Flags().StringP("aggregation-method", "a", "federated_averaging", "Aggregation method")
	createSessionCmd.Flags().Float64P("learning-rate", "l", 0.001, "Learning rate")
	createSessionCmd.Flags().IntP("batch-size", "b", 32, "Batch size")
	createSessionCmd.Flags().IntP("local-epochs", "e", 3, "Local epochs")
	// Training data flags
	createSessionCmd.Flags().StringP("dataset-cid", "", "", "IPFS/Filecoin dataset CID (required)")
	createSessionCmd.Flags().StringP("data-format", "", "csv", "Data format (csv, json, parquet)")
	createSessionCmd.Flags().StringP("split-strategy", "", "random", "Data split strategy (random, sequential, stratified)")
	// Advanced configuration flags
	createSessionCmd.Flags().StringP("config-file", "", "", "Custom model configuration file")
	createSessionCmd.Flags().BoolP("enable-differential-privacy", "", false, "Enable differential privacy")
	createSessionCmd.Flags().Float64P("noise-multiplier", "", 0.1, "Noise multiplier for differential privacy")
	createSessionCmd.Flags().Float64P("l2-norm-clip", "", 1.0, "L2 norm clipping for differential privacy")

	// List sessions flags
	listSessionsCmd.Flags().StringP("creator-address", "c", "", "Filter by creator address")

	// Get session flags
	getSessionCmd.Flags().StringP("session-id", "s", "", "Session ID (required)")

	// Start session flags (no flags needed - uses positional argument)

	// Submit update flags
	submitUpdateCmd.Flags().StringP("session-id", "s", "", "Session ID (required)")
	submitUpdateCmd.Flags().StringP("round-id", "r", "", "Round ID (required)")
	submitUpdateCmd.Flags().StringP("runner-id", "u", "", "Runner ID (required)")
	submitUpdateCmd.Flags().StringP("gradients-file", "g", "", "Path to gradients JSON file")
	submitUpdateCmd.Flags().IntP("data-size", "d", 1000, "Training data size")
	submitUpdateCmd.Flags().Float64P("loss", "l", 0.0, "Training loss")
	submitUpdateCmd.Flags().Float64P("accuracy", "a", 0.0, "Training accuracy")

	// Add create-session-with-data flags
	createSessionWithDataCmd.Flags().StringP("name", "n", "", "Session name (required)")
	createSessionWithDataCmd.Flags().StringP("description", "d", "", "Session description")
	createSessionWithDataCmd.Flags().StringP("model-type", "m", "", "Model type (required)")
	createSessionWithDataCmd.Flags().IntP("total-rounds", "r", 0, "Total training rounds (required)")
	createSessionWithDataCmd.Flags().IntP("min-participants", "p", 2, "Minimum participants")
	createSessionWithDataCmd.Flags().StringP("creator-address", "c", "", "Creator wallet address")
	createSessionWithDataCmd.Flags().StringP("data-format", "", "csv", "Data format (csv, json, parquet)")
	createSessionWithDataCmd.Flags().StringP("split-strategy", "", "random", "Data split strategy (random, stratified, sequential)")
	createSessionWithDataCmd.Flags().StringP("config-file", "", "", "Custom model configuration file")
	createSessionWithDataCmd.Flags().BoolP("enable-differential-privacy", "", false, "Enable differential privacy")
	createSessionWithDataCmd.Flags().Float64P("noise-multiplier", "", 0.1, "Noise multiplier for differential privacy")
	createSessionWithDataCmd.Flags().Float64P("l2-norm-clip", "", 1.0, "L2 norm clipping for differential privacy")

	createSessionWithDataCmd.MarkFlagRequired("name")
	createSessionWithDataCmd.MarkFlagRequired("model-type")
	createSessionWithDataCmd.MarkFlagRequired("total-rounds")

	// Add subcommands
	flCmd.AddCommand(createSessionCmd)
	flCmd.AddCommand(listSessionsCmd)
	flCmd.AddCommand(getSessionCmd)
	flCmd.AddCommand(startSessionCmd)
	flCmd.AddCommand(submitUpdateCmd)
	flCmd.AddCommand(createSessionWithDataCmd)

}

// GetFLCommand returns the federated learning command for integration into the main CLI
func GetFLCommand() *cobra.Command {
	return flCmd
}
