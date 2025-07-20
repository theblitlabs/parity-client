package commands

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/theblitlabs/deviceid"
	"github.com/theblitlabs/parity-client/internal/config"
	"github.com/theblitlabs/parity-client/internal/utils"
)

type RunnerStatus struct {
	RunnerID        string                 `json:"runner_id"`
	WalletAddress   string                 `json:"wallet_address"`
	ReputationScore int                    `json:"reputation_score"`
	Level           string                 `json:"level"`
	Status          string                 `json:"status"`
	IsEligible      bool                   `json:"is_eligible"`
	IsBanned        bool                   `json:"is_banned"`
	TotalTasks      int                    `json:"total_tasks"`
	SuccessfulTasks int                    `json:"successful_tasks"`
	FailedTasks     int                    `json:"failed_tasks"`
	SuccessRate     float64                `json:"success_rate"`
	QualityScore    float64                `json:"quality_score"`
	NetworkHealth   string                 `json:"network_health,omitempty"`
	BanReason       string                 `json:"ban_reason,omitempty"`
	LastSeen        string                 `json:"last_seen"`
	Specializations map[string]float64     `json:"specializations,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

type LeaderboardEntry struct {
	RunnerID        string             `json:"runner_id"`
	ReputationScore int                `json:"reputation_score"`
	Level           string             `json:"level"`
	Status          string             `json:"status"`
	TotalTasks      int                `json:"total_tasks"`
	SuccessRate     float64            `json:"success_rate"`
	QualityScore    float64            `json:"quality_score"`
	Specializations map[string]float64 `json:"specializations,omitempty"`
}

type NetworkStats struct {
	TotalRunners      int     `json:"total_runners"`
	ActiveRunners     int     `json:"active_runners"`
	BannedRunners     int     `json:"banned_runners"`
	WarningRunners    int     `json:"warning_runners"`
	AverageReputation int     `json:"average_reputation"`
	NetworkHealth     string  `json:"network_health"`
	TotalTasks        int     `json:"total_tasks"`
	BanRate           float64 `json:"ban_rate"`
}

type ReputationEvent struct {
	EventType   string                 `json:"event_type"`
	ScoreDelta  int                    `json:"score_delta"`
	NewScore    int                    `json:"new_score"`
	Description string                 `json:"description"`
	Timestamp   string                 `json:"timestamp"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

var reputationCmd = &cobra.Command{
	Use:   "reputation",
	Short: "Network quality control and runner reputation management",
	Long: `Check runner eligibility, view network quality stats, and manage runner reputation for network participation control.
	
This system helps maintain network quality by:
- Tracking runner performance and reliability
- Identifying and banning malicious actors
- Ensuring only quality runners participate in tasks
- Providing transparency in network health`,
}

var checkEligibilityCmd = &cobra.Command{
	Use:   "check-eligibility [runner-id]",
	Short: "Check if a runner is eligible to participate in the network",
	Long: `Verify whether a runner can participate in network tasks based on their reputation score, 
ban status, and overall network standing. This is the primary command for network quality control.`,
	Args: cobra.MaximumNArgs(1),
	Example: `  # Check your own eligibility
  parity-client reputation check-eligibility
  
  # Check another runner's eligibility
  parity-client reputation check-eligibility runner-123`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.NewConfigManager("").GetConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		var runnerID string
		if len(args) > 0 {
			runnerID = args[0]
			if err := utils.ValidateDeviceID(runnerID); err != nil {
				return err
			}
		} else {
			deviceIDManager := deviceid.NewManager(deviceid.Config{})
			runnerID, err = deviceIDManager.VerifyDeviceID()
			if err != nil {
				return fmt.Errorf("failed to get device ID: %w", err)
			}
		}

		serverURL := fmt.Sprintf("http://%s:%d", cfg.Server.Host, cfg.Server.Port)
		url := fmt.Sprintf("%s/api/v1/reputation/eligibility/%s", serverURL, runnerID)

		resp, err := http.Get(url)
		if err != nil {
			return fmt.Errorf("failed to check eligibility: %w", err)
		}
		defer func() {
			if closeErr := resp.Body.Close(); closeErr != nil {
				log.Printf("Error closing response body: %v", closeErr)
			}
		}()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("eligibility check failed with status: %d", resp.StatusCode)
		}

		var status RunnerStatus
		if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}

		printEligibilityStatus(&status)
		return nil
	},
}

var getStatusCmd = &cobra.Command{
	Use:   "status [runner-id]",
	Short: "Get detailed runner reputation and network standing",
	Long: `Get comprehensive information about a runner's reputation, performance metrics,
and current standing in the network. Shows eligibility, ban status, and quality indicators.`,
	Args: cobra.MaximumNArgs(1),
	Example: `  # Get your own status
  parity-client reputation status
  
  # Get another runner's status
  parity-client reputation status runner-123`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.NewConfigManager("").GetConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		var runnerID string
		if len(args) > 0 {
			runnerID = args[0]
			if err := utils.ValidateDeviceID(runnerID); err != nil {
				return err
			}
		} else {
			deviceIDManager := deviceid.NewManager(deviceid.Config{})
			runnerID, err = deviceIDManager.VerifyDeviceID()
			if err != nil {
				return fmt.Errorf("failed to get device ID: %w", err)
			}
		}

		serverURL := fmt.Sprintf("http://%s:%d", cfg.Server.Host, cfg.Server.Port)
		url := fmt.Sprintf("%s/api/v1/reputation/runner/%s", serverURL, runnerID)

		resp, err := http.Get(url)
		if err != nil {
			return fmt.Errorf("failed to get reputation: %w", err)
		}
		defer func() {
			if closeErr := resp.Body.Close(); closeErr != nil {
				log.Printf("Error closing response body: %v", closeErr)
			}
		}()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("failed to get reputation with status: %d", resp.StatusCode)
		}

		var status RunnerStatus
		if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}

		printRunnerStatus(&status)
		return nil
	},
}

var networkStatsCmd = &cobra.Command{
	Use:   "network-stats",
	Short: "View network health and quality statistics",
	Long: `Display overall network health metrics including number of active runners,
banned runners, average reputation, and network quality indicators.`,
	Example: `  # View network statistics
  parity-client reputation network-stats`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.NewConfigManager("").GetConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		serverURL := fmt.Sprintf("http://%s:%d", cfg.Server.Host, cfg.Server.Port)
		url := fmt.Sprintf("%s/api/v1/reputation/network/stats", serverURL)

		resp, err := http.Get(url)
		if err != nil {
			return fmt.Errorf("failed to get network stats: %w", err)
		}
		defer func() {
			if closeErr := resp.Body.Close(); closeErr != nil {
				log.Printf("Error closing response body: %v", closeErr)
			}
		}()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("failed to get network stats with status: %d", resp.StatusCode)
		}

		var stats NetworkStats
		if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}

		printNetworkStats(&stats)
		return nil
	},
}

var leaderboardCmd = &cobra.Command{
	Use:   "leaderboard [type]",
	Short: "View top-performing runners by category",
	Long: `Display leaderboards of top-performing runners. Only shows eligible (non-banned) runners.
Available types: overall, docker, llm, federated-learning`,
	Args: cobra.MaximumNArgs(1),
	Example: `  # Overall leaderboard
  parity-client reputation leaderboard
  
  # Docker execution specialists
  parity-client reputation leaderboard docker
  
  # LLM inference specialists  
  parity-client reputation leaderboard llm
  
  # Federated learning specialists
  parity-client reputation leaderboard federated-learning`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.NewConfigManager("").GetConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		leaderboardType := "overall"
		if len(args) > 0 {
			leaderboardType = args[0]
		}

		limit, _ := cmd.Flags().GetInt("limit")

		serverURL := fmt.Sprintf("http://%s:%d", cfg.Server.Host, cfg.Server.Port)
		url := fmt.Sprintf("%s/api/v1/reputation/leaderboard/%s?limit=%d", serverURL, leaderboardType, limit)

		resp, err := http.Get(url)
		if err != nil {
			return fmt.Errorf("failed to get leaderboard: %w", err)
		}
		defer func() {
			if closeErr := resp.Body.Close(); closeErr != nil {
				log.Printf("Error closing response body: %v", closeErr)
			}
		}()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("failed to get leaderboard with status: %d", resp.StatusCode)
		}

		var leaderboard []LeaderboardEntry
		if err := json.NewDecoder(resp.Body).Decode(&leaderboard); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}

		printLeaderboard(leaderboardType, leaderboard)
		return nil
	},
}

var eventsCmd = &cobra.Command{
	Use:   "events [runner-id]",
	Short: "View reputation events and changes for a runner",
	Long: `Display the history of reputation events for a runner, including task completions,
failures, quality changes, and any malicious behavior reports.`,
	Args: cobra.MaximumNArgs(1),
	Example: `  # View your own events
  parity-client reputation events
  
  # View another runner's events
  parity-client reputation events runner-123`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.NewConfigManager("").GetConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		var runnerID string
		if len(args) > 0 {
			runnerID = args[0]
		} else {
			deviceIDManager := deviceid.NewManager(deviceid.Config{})
			runnerID, err = deviceIDManager.VerifyDeviceID()
			if err != nil {
				return fmt.Errorf("failed to get device ID: %w", err)
			}
		}

		limit, _ := cmd.Flags().GetInt("limit")

		serverURL := fmt.Sprintf("http://%s:%d", cfg.Server.Host, cfg.Server.Port)
		url := fmt.Sprintf("%s/api/v1/reputation/events/%s?limit=%d", serverURL, runnerID, limit)

		resp, err := http.Get(url)
		if err != nil {
			return fmt.Errorf("failed to get events: %w", err)
		}
		defer func() {
			if closeErr := resp.Body.Close(); closeErr != nil {
				log.Printf("Error closing response body: %v", closeErr)
			}
		}()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("failed to get events with status: %d", resp.StatusCode)
		}

		var events []ReputationEvent
		if err := json.NewDecoder(resp.Body).Decode(&events); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}

		printReputationEvents(runnerID, events)
		return nil
	},
}

// Printing functions
func printEligibilityStatus(status *RunnerStatus) {
	fmt.Printf("Network Eligibility Check\n")
	fmt.Printf("========================\n\n")
	fmt.Printf("Runner ID:        %s\n", status.RunnerID)

	if status.IsEligible {
		fmt.Printf("✅ ELIGIBLE:      Yes - Can participate in network tasks\n")
	} else {
		fmt.Printf("❌ ELIGIBLE:      No - Cannot participate in network tasks\n")
	}

	if status.IsBanned {
		fmt.Printf("🚫 BANNED:        Yes - %s\n", status.BanReason)
	} else {
		fmt.Printf("🟢 BANNED:        No\n")
	}

	fmt.Printf("Status:           %s\n", status.Status)
	fmt.Printf("Reputation Score: %d\n", status.ReputationScore)
	fmt.Printf("Level:            %s\n", status.Level)

	if status.TotalTasks > 0 {
		fmt.Printf("Success Rate:     %.1f%% (%d/%d tasks)\n", status.SuccessRate, status.SuccessfulTasks, status.TotalTasks)
	}

	if status.Status == "warning" {
		fmt.Printf("\n⚠️  WARNING: Low reputation score. Improve performance to maintain network access.\n")
	}

	if status.IsBanned {
		fmt.Printf("\n🚫 BANNED: This runner is banned from network participation.\n")
		fmt.Printf("Reason: %s\n", status.BanReason)
	}
}

func printRunnerStatus(status *RunnerStatus) {
	fmt.Printf("Runner Network Status\n")
	fmt.Printf("====================\n\n")
	fmt.Printf("Runner ID:        %s\n", status.RunnerID)
	fmt.Printf("Wallet Address:   %s\n", status.WalletAddress)
	fmt.Printf("Reputation Score: %d\n", status.ReputationScore)
	fmt.Printf("Level:            %s\n", status.Level)
	fmt.Printf("Status:           %s\n", status.Status)
	fmt.Printf("Last Seen:        %s\n", status.LastSeen)

	fmt.Printf("\nNetwork Participation:\n")
	if status.IsEligible {
		fmt.Printf("  ✅ Eligible:    Yes\n")
	} else {
		fmt.Printf("  ❌ Eligible:    No\n")
	}

	if status.IsBanned {
		fmt.Printf("  🚫 Banned:      Yes - %s\n", status.BanReason)
	} else {
		fmt.Printf("  🟢 Banned:      No\n")
	}

	fmt.Printf("\nPerformance Metrics:\n")
	fmt.Printf("  Total Tasks:    %d\n", status.TotalTasks)
	fmt.Printf("  Successful:     %d\n", status.SuccessfulTasks)
	fmt.Printf("  Failed:         %d\n", status.FailedTasks)
	if status.TotalTasks > 0 {
		fmt.Printf("  Success Rate:   %.1f%%\n", status.SuccessRate)
	}
	fmt.Printf("  Quality Score:  %.1f\n", status.QualityScore)

	if len(status.Specializations) > 0 {
		fmt.Printf("\nSpecializations:\n")
		for spec, score := range status.Specializations {
			fmt.Printf("  %-18s %.1f\n", spec+":", score)
		}
	}
}

func printNetworkStats(stats *NetworkStats) {
	fmt.Printf("PLGenesis Network Health\n")
	fmt.Printf("=======================\n\n")

	// Network health indicator
	switch stats.NetworkHealth {
	case "healthy":
		fmt.Printf("🟢 Network Health: HEALTHY\n")
	case "degraded":
		fmt.Printf("🟡 Network Health: DEGRADED\n")
	case "critical":
		fmt.Printf("🔴 Network Health: CRITICAL\n")
	default:
		fmt.Printf("⚪ Network Health: %s\n", stats.NetworkHealth)
	}

	fmt.Printf("\nRunner Statistics:\n")
	fmt.Printf("  Total Runners:     %d\n", stats.TotalRunners)
	fmt.Printf("  Active Runners:    %d\n", stats.ActiveRunners)
	fmt.Printf("  Warning Runners:   %d\n", stats.WarningRunners)
	fmt.Printf("  Banned Runners:    %d\n", stats.BannedRunners)
	fmt.Printf("  Ban Rate:          %.1f%%\n", stats.BanRate)

	fmt.Printf("\nQuality Metrics:\n")
	fmt.Printf("  Average Reputation: %d\n", stats.AverageReputation)
	fmt.Printf("  Total Tasks:        %d\n", stats.TotalTasks)

	if stats.BanRate > 20 {
		fmt.Printf("\n⚠️  WARNING: High ban rate detected. Network quality may be compromised.\n")
	}

	if stats.ActiveRunners < 10 {
		fmt.Printf("\n⚠️  WARNING: Low number of active runners. Network capacity may be limited.\n")
	}
}

func printLeaderboard(leaderboardType string, leaderboard []LeaderboardEntry) {
	fmt.Printf("Top Performers - %s\n", strings.ToUpper(leaderboardType))
	fmt.Printf("=========================\n\n")

	if len(leaderboard) == 0 {
		fmt.Printf("No eligible runners found\n")
		return
	}

	for i, entry := range leaderboard {
		fmt.Printf("%d. Runner: %s\n", i+1, entry.RunnerID)
		fmt.Printf("   Score:      %d (%s)\n", entry.ReputationScore, entry.Level)
		fmt.Printf("   Status:     %s\n", entry.Status)
		if entry.TotalTasks > 0 {
			fmt.Printf("   Success:    %.1f%% (%d tasks)\n", entry.SuccessRate, entry.TotalTasks)
		}
		fmt.Printf("   Quality:    %.1f\n", entry.QualityScore)

		if len(entry.Specializations) > 0 && leaderboardType != "overall" {
			for spec, score := range entry.Specializations {
				if strings.Contains(spec, leaderboardType) || (leaderboardType == "fl" && strings.Contains(spec, "federated")) {
					fmt.Printf("   Specialty:  %.1f (%s)\n", score, spec)
					break
				}
			}
		}
		fmt.Println()
	}
}

func printReputationEvents(runnerID string, events []ReputationEvent) {
	fmt.Printf("Reputation Events for %s\n", runnerID)
	fmt.Printf("================================\n\n")

	if len(events) == 0 {
		fmt.Printf("No reputation events found\n")
		return
	}

	for _, event := range events {
		timestamp, _ := time.Parse(time.RFC3339, event.Timestamp)

		// Color code based on event type
		var icon string
		switch event.EventType {
		case "task_completed":
			icon = "✅"
		case "task_failed":
			icon = "❌"
		case "malicious_behavior":
			icon = "🚫"
		case "quality_bonus":
			icon = "⭐"
		case "performance_penalty":
			icon = "⚠️"
		default:
			icon = "📊"
		}

		fmt.Printf("%s %s\n", icon, event.Description)
		fmt.Printf("   Score Change: %+d (New Total: %d)\n", event.ScoreDelta, event.NewScore)
		fmt.Printf("   Time:         %s\n", timestamp.Format("2006-01-02 15:04:05"))

		if len(event.Metadata) > 0 {
			fmt.Printf("   Details:      ")
			for key, value := range event.Metadata {
				fmt.Printf("%s=%v ", key, value)
			}
			fmt.Println()
		}
		fmt.Println()
	}
}

func init() {
	// Leaderboard flags
	leaderboardCmd.Flags().IntP("limit", "l", 10, "Number of top runners to display")

	// Events flags
	eventsCmd.Flags().IntP("limit", "l", 20, "Number of recent events to display")

	// Add subcommands
	reputationCmd.AddCommand(checkEligibilityCmd)
	reputationCmd.AddCommand(getStatusCmd)
	reputationCmd.AddCommand(networkStatsCmd)
	reputationCmd.AddCommand(leaderboardCmd)
	reputationCmd.AddCommand(eventsCmd)
	reputationCmd.AddCommand(monitoringCmd)
}

// GetReputationCommand returns the reputation command for integration into the main CLI
func GetReputationCommand() *cobra.Command {
	return reputationCmd
}

var monitoringCmd = &cobra.Command{
	Use:   "monitoring",
	Short: "Monitor peer runner behavior and assignments",
	Long:  "View active monitoring assignments and peer monitoring statistics",
}

func init() {
	monitoringCmd.AddCommand(
		monitoringAssignmentsCmd,
		monitoringStatsCmd,
		monitoringMetricsCmd,
	)
}

var monitoringAssignmentsCmd = &cobra.Command{
	Use:   "assignments",
	Short: "View active monitoring assignments",
	Long:  "Display all current peer monitoring assignments in the network",
	RunE: func(cmd *cobra.Command, args []string) error {
		serverURL := getServerURL()
		resp, err := http.Get(fmt.Sprintf("%s/api/v1/monitoring/assignments", serverURL))
		if err != nil {
			return fmt.Errorf("failed to get monitoring assignments: %w", err)
		}
		defer func() {
			if closeErr := resp.Body.Close(); closeErr != nil {
				log.Printf("Error closing response body: %v", closeErr)
			}
		}()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("server returned status %d", resp.StatusCode)
		}

		var assignments []map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&assignments); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}

		if len(assignments) == 0 {
			fmt.Println("No active monitoring assignments")
			return nil
		}

		fmt.Printf("📊 Active Monitoring Assignments (%d)\n\n", len(assignments))

		for _, assignment := range assignments {
			fmt.Printf("Assignment ID: %s\n", assignment["id"])
			fmt.Printf("Monitor: %s → Target: %s\n", assignment["monitor_id"], assignment["target_id"])

			if startTime, ok := assignment["start_time"].(string); ok {
				fmt.Printf("Started: %s\n", startTime)
			}
			if duration, ok := assignment["duration"].(string); ok {
				fmt.Printf("Duration: %s\n", duration)
			}

			status := "Active"
			if active, ok := assignment["is_active"].(bool); ok && !active {
				status = "Completed"
			}
			fmt.Printf("Status: %s\n", status)

			if reportType, ok := assignment["report_type"].(string); ok && reportType != "" {
				fmt.Printf("Report: %s\n", reportType)
			}

			fmt.Println(strings.Repeat("-", 50))
		}

		return nil
	},
}

var monitoringStatsCmd = &cobra.Command{
	Use:   "stats",
	Short: "View monitoring network statistics",
	Long:  "Display overall peer monitoring network statistics and health",
	RunE: func(cmd *cobra.Command, args []string) error {
		serverURL := getServerURL()
		resp, err := http.Get(fmt.Sprintf("%s/api/v1/monitoring/stats", serverURL))
		if err != nil {
			return fmt.Errorf("failed to get monitoring stats: %w", err)
		}
		defer func() {
			if closeErr := resp.Body.Close(); closeErr != nil {
				log.Printf("Error closing response body: %v", closeErr)
			}
		}()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("server returned status %d", resp.StatusCode)
		}

		var stats map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}

		fmt.Println("🔍 Peer Monitoring Network Statistics")
		fmt.Println(strings.Repeat("=", 50))

		if activeAssignments, ok := stats["active_assignments"].(float64); ok {
			fmt.Printf("Active Assignments: %.0f\n", activeAssignments)
		}
		if totalAssignments, ok := stats["total_assignments"].(float64); ok {
			fmt.Printf("Total Assignments: %.0f\n", totalAssignments)
		}
		if monitoredRunners, ok := stats["monitored_runners"].(float64); ok {
			fmt.Printf("Monitored Runners: %.0f\n", monitoredRunners)
		}
		if interval, ok := stats["monitoring_interval"].(string); ok {
			fmt.Printf("Monitoring Interval: %s\n", interval)
		}

		return nil
	},
}

var monitoringMetricsCmd = &cobra.Command{
	Use:   "metrics [runner-id]",
	Short: "View monitoring metrics for a specific runner",
	Long:  "Display detailed peer monitoring metrics for a runner",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		runnerID := args[0]
		serverURL := getServerURL()
		resp, err := http.Get(fmt.Sprintf("%s/api/v1/monitoring/metrics/%s", serverURL, runnerID))
		if err != nil {
			return fmt.Errorf("failed to get monitoring metrics: %w", err)
		}
		defer func() {
			if closeErr := resp.Body.Close(); closeErr != nil {
				log.Printf("Error closing response body: %v", closeErr)
			}
		}()

		if resp.StatusCode == http.StatusNotFound {
			fmt.Printf("No monitoring metrics found for runner: %s\n", runnerID)
			return nil
		}

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("server returned status %d", resp.StatusCode)
		}

		var metrics map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&metrics); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}

		fmt.Printf("📈 Monitoring Metrics for Runner: %s\n", runnerID)
		fmt.Println(strings.Repeat("=", 50))

		if tasksObserved, ok := metrics["tasks_observed"].(float64); ok {
			fmt.Printf("Tasks Observed: %.0f\n", tasksObserved)
		}
		if tasksCompleted, ok := metrics["tasks_completed"].(float64); ok {
			fmt.Printf("Tasks Completed: %.0f\n", tasksCompleted)
		}
		if tasksFailed, ok := metrics["tasks_failed"].(float64); ok {
			fmt.Printf("Tasks Failed: %.0f\n", tasksFailed)
		}
		if avgResponseTime, ok := metrics["avg_response_time"].(string); ok {
			fmt.Printf("Avg Response Time: %s\n", avgResponseTime)
		}
		if qualityScore, ok := metrics["quality_score"].(float64); ok {
			fmt.Printf("Quality Score: %.1f/100\n", qualityScore)
		}
		if reliabilityScore, ok := metrics["reliability_score"].(float64); ok {
			fmt.Printf("Reliability Score: %.1f/100\n", reliabilityScore)
		}
		if offlineDuration, ok := metrics["offline_duration"].(string); ok && offlineDuration != "0s" {
			fmt.Printf("⚠️  Offline Duration: %s\n", offlineDuration)
		}

		if suspiciousPatterns, ok := metrics["suspicious_patterns"].([]interface{}); ok && len(suspiciousPatterns) > 0 {
			fmt.Printf("\n🚨 Suspicious Patterns Detected:\n")
			for _, pattern := range suspiciousPatterns {
				if patternStr, ok := pattern.(string); ok {
					fmt.Printf("  • %s\n", patternStr)
				}
			}
		}

		if lastActivity, ok := metrics["last_activity"].(string); ok {
			fmt.Printf("\nLast Activity: %s\n", lastActivity)
		}

		return nil
	},
}

func getServerURL() string {
	// Try to load configuration
	cfg, err := loadReputationConfig()
	if err != nil {
		// Fallback to environment variables or default
		return "http://localhost:8082" // Default server URL
	}

	// Use runner server URL if available, otherwise server config
	if cfg.Runner.ServerURL != "" {
		return cfg.Runner.ServerURL
	}

	if cfg.Server.Host != "" {
		return fmt.Sprintf("http://%s:%d", cfg.Server.Host, cfg.Server.Port)
	}

	return "http://localhost:8082" // Default fallback
}

func loadReputationConfig() (*config.Config, error) {
	configPath := ".env"
	configManager := config.NewConfigManager(configPath)
	return configManager.GetConfig()
}
