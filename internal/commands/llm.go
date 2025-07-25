package commands

import (
	"context"
	"fmt"
	"strconv"
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
	"github.com/theblitlabs/parity-client/internal/utils"
)

var llmCmd = &cobra.Command{
	Use:   "llm",
	Short: "LLM operations (submit prompts, check status, list prompts)",
	Long:  `Submit prompts to LLM models, check their status, and list previous prompts`,
}

var submitCmd = &cobra.Command{
	Use:   "submit",
	Short: "Submit a prompt to an LLM model",
	Long:  `Submit a prompt to an LLM model and optionally wait for completion`,
	Example: `  # Submit prompt and wait for response
  parity-client llm submit --model gpt-4 --prompt "What is the capital of France?" --wait

  # Submit prompt without waiting
  parity-client llm submit --model llama2 --prompt "Explain quantum computing"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		log := gologger.WithComponent("llm_submit")

		model, _ := cmd.Flags().GetString("model")
		prompt, _ := cmd.Flags().GetString("prompt")
		wait, _ := cmd.Flags().GetBool("wait")
		timeout, _ := cmd.Flags().GetDuration("timeout")

		if err := utils.ValidateModelName(model); err != nil {
			return err
		}
		if err := utils.ValidatePrompt(prompt); err != nil {
			return err
		}

		configPath, _ := cmd.Flags().GetString("config-path")
		if configPath == "" {
			configPath = utils.GetDefaultConfigPath()
		}

		configManager := config.NewConfigManager(configPath)
		cfg, err := configManager.GetConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Get creator address from wallet
		creatorAddress, err := getLLMCreatorAddress(cfg)
		if err != nil {
			return err
		}

		if cfg.Runner.ServerURL == "" {
			return fmt.Errorf("runner server URL not configured. Please set RUNNER_SERVER_URL in your config")
		}

		log.Info().Str("server_url", cfg.Runner.ServerURL).Msg("Using server URL")

		clientID := "parity-client-" + strconv.FormatInt(time.Now().Unix(), 10)
		llmClient := client.NewLLMClient(cfg.Runner.ServerURL, clientID)

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		log.Info().
			Str("model", model).
			Str("creator_address", creatorAddress).
			Str("prompt_preview", truncatePrompt(prompt, 50)).
			Bool("wait", wait).
			Msg("Submitting prompt")

		response, err := llmClient.SubmitPrompt(ctx, prompt, model, creatorAddress)
		if err != nil {
			return fmt.Errorf("failed to submit prompt: %w", err)
		}

		fmt.Printf("Prompt submitted successfully\n")
		fmt.Printf("ID: %s\n", response.ID)
		fmt.Printf("Status: %s\n", response.Status)
		fmt.Printf("Model: %s\n", response.ModelName)
		fmt.Printf("Created: %s\n", response.CreatedAt)

		if wait {
			log.Info().Str("prompt_id", response.ID).Msg("Waiting for completion...")
			fmt.Printf("\nWaiting for completion...\n")

			completed, err := llmClient.WaitForCompletion(ctx, response.ID, 2*time.Second)
			if err != nil {
				return fmt.Errorf("failed while waiting for completion: %w", err)
			}

			fmt.Printf("\nTask completed!\n")
			fmt.Printf("Status: %s\n", completed.Status)
			if completed.CompletedAt != nil {
				fmt.Printf("Completed: %s\n", *completed.CompletedAt)
			}
			if completed.Response != "" {
				fmt.Printf("\nResponse:\n%s\n", completed.Response)
			}
		}

		return nil
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List recent prompts",
	Long:  `List recent prompts submitted to the LLM service`,
	RunE: func(cmd *cobra.Command, args []string) error {
		limit, _ := cmd.Flags().GetInt("limit")
		offset, _ := cmd.Flags().GetInt("offset")

		configPath, _ := cmd.Flags().GetString("config-path")
		if configPath == "" {
			configPath = utils.GetDefaultConfigPath()
		}

		configManager := config.NewConfigManager(configPath)
		cfg, err := configManager.GetConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		if cfg.Runner.ServerURL == "" {
			return fmt.Errorf("runner server URL not configured. Please set RUNNER_SERVER_URL in your config")
		}

		clientID := "parity-client-" + strconv.FormatInt(time.Now().Unix(), 10)
		llmClient := client.NewLLMClient(cfg.Runner.ServerURL, clientID)

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		prompts, err := llmClient.ListPrompts(ctx, limit, offset)
		if err != nil {
			return fmt.Errorf("failed to list prompts: %w", err)
		}

		if len(prompts) == 0 {
			fmt.Println("No prompts found")
			return nil
		}

		fmt.Printf("Found %d prompts:\n\n", len(prompts))
		for i, p := range prompts {
			fmt.Printf("%d. ID: %s\n", i+1, p.ID)
			fmt.Printf("   Model: %s\n", p.ModelName)
			fmt.Printf("   Status: %s\n", p.Status)
			fmt.Printf("   Created: %s\n", p.CreatedAt)
			if p.CompletedAt != nil {
				fmt.Printf("   Completed: %s\n", *p.CompletedAt)
			}
			if p.Response != "" {
				fmt.Printf("   Response: %s\n", truncatePrompt(p.Response, 100))
			}
			fmt.Println()
		}

		return nil
	},
}

var statusCmd = &cobra.Command{
	Use:   "status [prompt-id]",
	Short: "Check the status of a specific prompt",
	Long:  `Check the status and response of a specific prompt by ID`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		promptID := args[0]

		configPath, _ := cmd.Flags().GetString("config-path")
		if configPath == "" {
			configPath = utils.GetDefaultConfigPath()
		}

		configManager := config.NewConfigManager(configPath)
		cfg, err := configManager.GetConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		if cfg.Runner.ServerURL == "" {
			return fmt.Errorf("runner server URL not configured. Please set RUNNER_SERVER_URL in your config")
		}

		clientID := "parity-client-" + strconv.FormatInt(time.Now().Unix(), 10)
		llmClient := client.NewLLMClient(cfg.Runner.ServerURL, clientID)

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		prompt, err := llmClient.GetPrompt(ctx, promptID)
		if err != nil {
			return fmt.Errorf("failed to get prompt status: %w", err)
		}

		fmt.Printf("Prompt Status\n\n")
		fmt.Printf("ID: %s\n", prompt.ID)
		fmt.Printf("Model: %s\n", prompt.ModelName)
		fmt.Printf("Status: %s\n", prompt.Status)
		fmt.Printf("Created: %s\n", prompt.CreatedAt)
		if prompt.CompletedAt != nil {
			fmt.Printf("Completed: %s\n", *prompt.CompletedAt)
		}

		if prompt.Response != "" {
			fmt.Printf("\nResponse:\n%s\n", prompt.Response)
		}

		return nil
	},
}

var listModelsCmd = &cobra.Command{
	Use:   "list-models",
	Short: "List available LLM models",
	Long:  `List all available LLM models that are currently loaded and ready to use`,
	RunE: func(cmd *cobra.Command, args []string) error {
		configPath, _ := cmd.Flags().GetString("config-path")
		if configPath == "" {
			configPath = utils.GetDefaultConfigPath()
		}

		configManager := config.NewConfigManager(configPath)
		cfg, err := configManager.GetConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		if cfg.Runner.ServerURL == "" {
			return fmt.Errorf("runner server URL not configured. Please set RUNNER_SERVER_URL in your config")
		}

		clientID := "parity-client-" + strconv.FormatInt(time.Now().Unix(), 10)
		llmClient := client.NewLLMClient(cfg.Runner.ServerURL, clientID)

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		modelsResp, err := llmClient.GetAvailableModels(ctx)
		if err != nil {
			return fmt.Errorf("failed to get available models: %w", err)
		}

		if len(modelsResp.Models) == 0 {
			fmt.Println("No models are currently available")
			return nil
		}

		fmt.Printf("Available Models (%d):\n\n", modelsResp.Count)
		for i, model := range modelsResp.Models {
			fmt.Printf("%d. %s\n", i+1, model.ModelName)
			if model.MaxTokens > 0 {
				fmt.Printf("   Max Tokens: %d\n", model.MaxTokens)
			}
			fmt.Printf("   Status: %s\n", func() string {
				if model.IsLoaded {
					return "✅ Loaded"
				}
				return "⏳ Loading"
			}())
			fmt.Println()
		}

		return nil
	},
}

func init() {
	// Submit command flags
	submitCmd.Flags().StringP("model", "m", "", "Model name (required)")
	submitCmd.Flags().StringP("prompt", "p", "", "Prompt text (required)")
	submitCmd.Flags().BoolP("wait", "w", false, "Wait for completion")
	submitCmd.Flags().DurationP("timeout", "t", 5*time.Minute, "Timeout duration")
	submitCmd.Flags().String("config-path", "", "Path to config file")

	// Mark required flags
	if err := submitCmd.MarkFlagRequired("model"); err != nil {
		log.Error().Err(err).Msg("Failed to mark model flag as required")
	}
	if err := submitCmd.MarkFlagRequired("prompt"); err != nil {
		log.Error().Err(err).Msg("Failed to mark prompt flag as required")
	}

	// List command flags
	listCmd.Flags().IntP("limit", "l", 10, "Number of prompts to list")
	listCmd.Flags().IntP("offset", "o", 0, "Offset for pagination")
	listCmd.Flags().String("config-path", "", "Path to config file")

	// Status command flags
	statusCmd.Flags().String("config-path", "", "Path to config file")

	// Add subcommands
	llmCmd.AddCommand(submitCmd)
	llmCmd.AddCommand(listCmd)
	llmCmd.AddCommand(listModelsCmd)
	llmCmd.AddCommand(statusCmd)
}

func truncatePrompt(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func getLLMCreatorAddress(cfg *config.Config) (string, error) {
	// First load the private key from keystore
	keystoreAdapter, err := keystore.NewAdapter(nil)
	if err != nil {
		return "", fmt.Errorf("failed to create keystore adapter: %w", err)
	}

	privateKey, err := keystoreAdapter.GetPrivateKeyHex()
	if err != nil {
		return "", fmt.Errorf("no private key found in keystore. Please authenticate first using 'parity-client auth'")
	}

	walletAdapter, err := wallet.NewAdapter(walletsdk.ClientConfig{
		RPCURL:       cfg.BlockchainNetwork.RPC,
		ChainID:      cfg.BlockchainNetwork.ChainID,
		TokenAddress: ethcommon.HexToAddress(cfg.BlockchainNetwork.TokenAddress),
		PrivateKey:   privateKey,
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
