package commands

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/spf13/cobra"
	"github.com/theblitlabs/gologger"
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
	healthCmd.Flags().StringVar(&healthEndpoint, "endpoint", "http://localhost:3000", "Health check endpoint URL")
	healthCmd.Flags().DurationVar(&healthTimeout, "timeout", 10*time.Second, "Timeout for health check request")
}

func runHealthCheck(cmd *cobra.Command, args []string) {
	logger := gologger.Get().With().Str("component", "health-cmd").Logger()

	var url string
	if healthDetailed {
		url = fmt.Sprintf("%s/health/detailed", healthEndpoint)
	} else {
		url = fmt.Sprintf("%s/health", healthEndpoint)
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
	fmt.Printf("Status: %s\n", result["status"])
	fmt.Printf("Timestamp: %s\n", result["timestamp"])
	fmt.Printf("Version: %s\n", result["version"])

	if healthDetailed {
		fmt.Printf("\nDetailed Information:\n")
		fmt.Println(string(prettyJSON))
	}
}
