package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	shell "github.com/ipfs/go-ipfs-api"
	"github.com/rs/zerolog"
	"github.com/theblitlabs/gologger"
	"github.com/theblitlabs/parity-client/internal/config"
	"github.com/theblitlabs/parity-client/internal/version"
)

var (
	startTime = time.Now()
)

type HealthHandler struct {
	config *config.Config
	logger zerolog.Logger
}

type HealthStatus struct {
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Version   string            `json:"version"`
	Uptime    string            `json:"uptime"`
	Services  map[string]string `json:"services,omitempty"`
}

type DetailedHealthStatus struct {
	Status    string                 `json:"status"`
	Timestamp time.Time              `json:"timestamp"`
	Version   string                 `json:"version"`
	Uptime    string                 `json:"uptime"`
	Services  map[string]ServiceInfo `json:"services"`
	Config    ConfigInfo             `json:"config"`
}

type ServiceInfo struct {
	Status    string    `json:"status"`
	LastCheck time.Time `json:"last_check"`
	Error     string    `json:"error,omitempty"`
	Latency   string    `json:"latency,omitempty"`
}

type ConfigInfo struct {
	ServerHost    string `json:"server_host"`
	ServerPort    int    `json:"server_port"`
	BlockchainRPC string `json:"blockchain_rpc"`
	IPFSEndpoint  string `json:"ipfs_endpoint"`
	RunnerURL     string `json:"runner_url"`
}

func NewHealthHandler(cfg *config.Config) *HealthHandler {
	return &HealthHandler{
		config: cfg,
		logger: gologger.Get().With().Str("component", "health").Logger(),
	}
}

func (h *HealthHandler) getUptime() string {
	uptime := time.Since(startTime)
	return uptime.String()
}

func (h *HealthHandler) checkBlockchainHealth() ServiceInfo {
	start := time.Now()
	status := ServiceInfo{
		Status:    "unhealthy",
		LastCheck: time.Now(),
	}

	if h.config.BlockchainNetwork.RPC == "" {
		status.Error = "Blockchain RPC URL not configured"
		return status
	}

	client, err := ethclient.Dial(h.config.BlockchainNetwork.RPC)
	if err != nil {
		status.Error = fmt.Sprintf("Failed to connect to blockchain: %v", err)
		return status
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = client.BlockNumber(ctx)
	if err != nil {
		status.Error = fmt.Sprintf("Failed to get block number: %v", err)
		return status
	}

	status.Status = "healthy"
	status.Latency = time.Since(start).String()
	return status
}

func (h *HealthHandler) checkIPFSHealth() ServiceInfo {
	start := time.Now()
	status := ServiceInfo{
		Status:    "unhealthy",
		LastCheck: time.Now(),
	}

	if h.config.BlockchainNetwork.IPFSEndpoint == "" {
		status.Error = "IPFS endpoint not configured"
		return status
	}

	sh := shell.NewShell(h.config.BlockchainNetwork.IPFSEndpoint)

	// Test IPFS connection by getting version
	_, _, err := sh.Version()
	if err != nil {
		status.Error = fmt.Sprintf("Failed to connect to IPFS: %v", err)
		return status
	}

	status.Status = "healthy"
	status.Latency = time.Since(start).String()
	return status
}

func (h *HealthHandler) checkRunnerHealth() ServiceInfo {
	start := time.Now()
	status := ServiceInfo{
		Status:    "unhealthy",
		LastCheck: time.Now(),
	}

	if h.config.Runner.ServerURL == "" {
		status.Error = "Runner URL not configured"
		return status
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// Try to connect to runner health endpoint
	resp, err := client.Get(fmt.Sprintf("%s/health", h.config.Runner.ServerURL))
	if err != nil {
		status.Error = fmt.Sprintf("Failed to connect to runner: %v", err)
		return status
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		status.Error = fmt.Sprintf("Runner returned status: %d", resp.StatusCode)
		return status
	}

	status.Status = "healthy"
	status.Latency = time.Since(start).String()
	return status
}

func (h *HealthHandler) HandleHealthCheck(w http.ResponseWriter, r *http.Request) {
	// Quick health check - only check if services are configured
	services := make(map[string]string)

	if h.config.BlockchainNetwork.RPC != "" {
		services["blockchain"] = "configured"
	}
	if h.config.BlockchainNetwork.IPFSEndpoint != "" {
		services["ipfs"] = "configured"
	}
	if h.config.Runner.ServerURL != "" {
		services["runner"] = "configured"
	}

	status := HealthStatus{
		Status:    "healthy",
		Timestamp: time.Now(),
		Version:   version.GetShortVersion(),
		Uptime:    h.getUptime(),
		Services:  services,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(status); err != nil {
		h.logger.Error().Err(err).Msg("Failed to encode health status")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func (h *HealthHandler) HandleDetailedHealthCheck(w http.ResponseWriter, r *http.Request) {
	services := make(map[string]ServiceInfo)

	// Check blockchain connection
	services["blockchain"] = h.checkBlockchainHealth()

	// Check IPFS connection
	services["ipfs"] = h.checkIPFSHealth()

	// Check runner service
	services["runner"] = h.checkRunnerHealth()

	// Determine overall status
	overallStatus := "healthy"
	for _, service := range services {
		if service.Status != "healthy" {
			overallStatus = "degraded"
			break
		}
	}

	status := DetailedHealthStatus{
		Status:    overallStatus,
		Timestamp: time.Now(),
		Version:   version.GetShortVersion(),
		Uptime:    h.getUptime(),
		Services:  services,
		Config: ConfigInfo{
			ServerHost:    h.config.Server.Host,
			ServerPort:    h.config.Server.Port,
			BlockchainRPC: h.config.BlockchainNetwork.RPC,
			IPFSEndpoint:  h.config.BlockchainNetwork.IPFSEndpoint,
			RunnerURL:     h.config.Runner.ServerURL,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(status); err != nil {
		h.logger.Error().Err(err).Msg("Failed to encode detailed health status")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func (h *HealthHandler) HandleReadinessCheck(w http.ResponseWriter, r *http.Request) {
	// Readiness check - verify all required services are available
	blockchainHealth := h.checkBlockchainHealth()
	ipfsHealth := h.checkIPFSHealth()
	runnerHealth := h.checkRunnerHealth()

	ready := blockchainHealth.Status == "healthy" &&
		ipfsHealth.Status == "healthy" &&
		runnerHealth.Status == "healthy"

	status := HealthStatus{
		Status:    "ready",
		Timestamp: time.Now(),
		Version:   version.GetShortVersion(),
		Uptime:    h.getUptime(),
	}

	if !ready {
		status.Status = "not_ready"
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(status); err != nil {
		h.logger.Error().Err(err).Msg("Failed to encode readiness status")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func (h *HealthHandler) HandleLivenessCheck(w http.ResponseWriter, r *http.Request) {
	// Liveness check - just verify the application is running
	status := HealthStatus{
		Status:    "alive",
		Timestamp: time.Now(),
		Version:   version.GetShortVersion(),
		Uptime:    h.getUptime(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(status); err != nil {
		h.logger.Error().Err(err).Msg("Failed to encode liveness status")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
