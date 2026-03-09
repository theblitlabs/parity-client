package commands

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/theblitlabs/gologger"
	"github.com/theblitlabs/parity-client/internal/config"
	"github.com/theblitlabs/parity-client/internal/utils"
)

var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "Check health status of the parity client",
	Long:  `Check the health status of the parity client and its connected services`,
	Run:   runHealthCheck,
}

var (
	healthDetailed bool
	healthEndpoint string
	healthTimeout  time.Duration
)

func init() {
	healthCmd.Flags().BoolVar(&healthDetailed, "detailed", false, "Get detailed health information")
	healthCmd.Flags().StringVar(&healthEndpoint, "endpoint", "", "Health check endpoint URL (defaults to the configured local client address)")
	healthCmd.Flags().DurationVar(&healthTimeout, "timeout", 10*time.Second, "Timeout for health check request")
}

func runHealthCheck(cmd *cobra.Command, args []string) {
	logger := gologger.Get().With().Str("component", "health-cmd").Logger()
	endpoint := resolveHealthEndpoint(cmd)

	var url string
	if healthDetailed {
		url = fmt.Sprintf("%s/health/detailed", endpoint)
	} else {
		url = fmt.Sprintf("%s/health", endpoint)
	}

	logger.Info().Str("url", url).Msg("Checking health status")

	client := &http.Client{
		Timeout: healthTimeout,
	}

	resp, err := client.Get(url)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to connect to health endpoint")
		fmt.Printf("❌ Health check failed: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Error().Int("status_code", resp.StatusCode).Msg("Health check returned non-OK status")
		fmt.Printf("❌ Health check failed with status: %d\n", resp.StatusCode)
		return
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		logger.Error().Err(err).Msg("Failed to decode health response")
		fmt.Printf("❌ Failed to decode health response: %v\n", err)
		return
	}

	// Pretty print the JSON response
	prettyJSON, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		logger.Error().Err(err).Msg("Failed to marshal health response")
		fmt.Printf("❌ Failed to format health response: %v\n", err)
		return
	}

	fmt.Printf("✅ Health check successful\n")
	fmt.Printf("Endpoint: %s\n", endpoint)
	fmt.Printf("Status: %s\n", result["status"])
	fmt.Printf("Timestamp: %s\n", result["timestamp"])
	fmt.Printf("Version: %s\n", result["version"])

	if healthDetailed {
		fmt.Printf("\nDetailed Information:\n")
		fmt.Println(string(prettyJSON))
	}
}

func resolveHealthEndpoint(cmd *cobra.Command) string {
	if cmd != nil && cmd.Flags().Changed("endpoint") && strings.TrimSpace(healthEndpoint) != "" {
		return strings.TrimRight(strings.TrimSpace(healthEndpoint), "/")
	}

	configPath := ""
	if cmd != nil {
		if flagValue, err := cmd.Flags().GetString("config-path"); err == nil {
			configPath = flagValue
		}
	}
	if configPath == "" {
		configPath = utils.GetDefaultConfigPath()
	}

	cfg, err := config.NewConfigManager(configPath).GetConfig()
	if err != nil || cfg == nil {
		return "http://127.0.0.1:8090"
	}

	return strings.TrimRight(localHealthBaseURL(cfg), "/")
}

func localHealthBaseURL(cfg *config.Config) string {
	host := cfg.Server.Host
	switch host {
	case "", "0.0.0.0", "::":
		host = "127.0.0.1"
	}

	port := cfg.Server.Port
	if port == 0 {
		port = 8090
	}

	return fmt.Sprintf("http://%s:%d", host, port)
}
