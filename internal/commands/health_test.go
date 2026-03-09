package commands

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
)

func TestResolveHealthEndpointUsesExplicitFlag(t *testing.T) {
	t.Cleanup(func() {
		healthEndpoint = ""
	})

	cmd := &cobra.Command{Use: "health"}
	cmd.Flags().String("endpoint", "", "")
	if err := cmd.Flags().Set("endpoint", "http://example.com:9000/"); err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	healthEndpoint = "http://example.com:9000/"

	if got := resolveHealthEndpoint(cmd); got != "http://example.com:9000" {
		t.Fatalf("resolveHealthEndpoint() = %q, want %q", got, "http://example.com:9000")
	}
}

func TestResolveHealthEndpointUsesConfigPath(t *testing.T) {
	t.Cleanup(func() {
		healthEndpoint = ""
	})

	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")
	configBody := []byte("SERVER:\n  HOST: 0.0.0.0\n  PORT: 8123\n")
	if err := os.WriteFile(configPath, configBody, 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	cmd := &cobra.Command{Use: "health"}
	cmd.Flags().String("endpoint", "", "")
	cmd.Flags().String("config-path", "", "")
	if err := cmd.Flags().Set("config-path", configPath); err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	if got := resolveHealthEndpoint(cmd); got != "http://127.0.0.1:8123" {
		t.Fatalf("resolveHealthEndpoint() = %q, want %q", got, "http://127.0.0.1:8123")
	}
}

func TestResolveHealthEndpointFallsBackToLocalDefault(t *testing.T) {
	t.Cleanup(func() {
		healthEndpoint = ""
	})

	cmd := &cobra.Command{Use: "health"}
	cmd.Flags().String("endpoint", "", "")
	cmd.Flags().String("config-path", "", "")

	if got := resolveHealthEndpoint(cmd); got != "http://127.0.0.1:8090" {
		t.Fatalf("resolveHealthEndpoint() = %q, want %q", got, "http://127.0.0.1:8090")
	}
}
