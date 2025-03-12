package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/theblitlabs/deviceid"
	"github.com/theblitlabs/gologger"
	"github.com/theblitlabs/parity-client/internal/config"
)

// isPortAvailable verifies if a port is available for use
func isPortAvailable(port int) error {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("port %d is not available: %w", port, err)
	}
	ln.Close()
	return nil
}

func getCreatorAddress() (string, error) {
	keystoreDir := filepath.Join(os.Getenv("HOME"), ".parity")
	keystorePath := filepath.Join(keystoreDir, "keystore.json")

	data, err := os.ReadFile(keystorePath)
	if err != nil {
		return "", fmt.Errorf("failed to read keystore: %w", err)
	}

	var keystore KeyStore
	if err := json.Unmarshal(data, &keystore); err != nil {
		return "", fmt.Errorf("failed to parse keystore: %w", err)
	}

	privateKey := keystore.PrivateKey
	ecdsaKey, err := crypto.HexToECDSA(privateKey)
	if err != nil {
		return "", fmt.Errorf("invalid private key in keystore: %w", err)
	}

	address := crypto.PubkeyToAddress(ecdsaKey.PublicKey)
	return address.Hex(), nil
}

func RunChain(port int) {
	log := gologger.Get().With().Str("component", "chain").Logger()

	// Load config
	cfg, err := config.LoadConfig("config/config.yaml")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load config")
	}

	// Get or generate device ID
	deviceIDManager := deviceid.NewManager(deviceid.Config{})
	deviceID, err := deviceIDManager.VerifyDeviceID()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to verify device ID")
	}

	// Get creator address from keystore
	creatorAddress, err := getCreatorAddress()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to get creator address. Please authenticate first using 'auth' command")
	}

	// Proxy request to the server
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Debug().
			Str("original_path", r.URL.Path).
			Str("method", r.Method).
			Str("content_type", r.Header.Get("Content-Type")).
			Msg("Received request")

		// Get the path and ensure it doesn't have a leading slash
		path := strings.TrimPrefix(r.URL.Path, "/")
		path = strings.TrimPrefix(path, "api/")

		// Create new request to forward to the server
		targetURL := fmt.Sprintf("%s/%s", strings.TrimSuffix(cfg.Runner.ServerURL, "/"), path)
		log.Info().
			Str("method", r.Method).
			Str("original_path", r.URL.Path).
			Str("modified_path", path).
			Str("target_url", targetURL).
			Str("server_url", cfg.Runner.ServerURL).
			Str("device_id", deviceID).
			Str("creator_address", creatorAddress).
			Msg("Forwarding request")

		var proxyReq *http.Request
		var err error

		// Only modify body for POST/PUT requests with JSON content
		if (r.Method == "POST" || r.Method == "PUT") && strings.Contains(r.Header.Get("Content-Type"), "application/json") {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "Error reading request body", http.StatusInternalServerError)
				return
			}
			r.Body.Close()

			// Try to decode and modify JSON body
			var requestData map[string]interface{}
			if err := json.NewDecoder(bytes.NewBuffer(body)).Decode(&requestData); err != nil {
				log.Error().Err(err).Msg("Failed to decode request body")
				http.Error(w, "Invalid request body", http.StatusBadRequest)
				return
			}

			// Add device ID and creator address to request body
			requestData["creator_device_id"] = deviceID
			requestData["creator_address"] = creatorAddress

			// Marshal modified body
			modifiedBody, err := json.Marshal(requestData)
			if err != nil {
				log.Error().Err(err).Msg("Failed to marshal modified request body")
				http.Error(w, "Error processing request", http.StatusInternalServerError)
				return
			}

			proxyReq, err = http.NewRequest(r.Method, targetURL, bytes.NewBuffer(modifiedBody))
			if err != nil {
				http.Error(w, "Error creating proxy request", http.StatusInternalServerError)
				return
			}
		} else {
			// For other requests, forward the body as-is
			proxyReq, err = http.NewRequest(r.Method, targetURL, r.Body)
			if err != nil {
				http.Error(w, "Error creating proxy request", http.StatusInternalServerError)
				return
			}
		}

		// Copy headers
		for header, values := range r.Header {
			for _, value := range values {
				proxyReq.Header.Add(header, value)
			}
		}

		// Always add device ID and creator address headers
		proxyReq.Header.Set("X-Device-ID", deviceID)
		proxyReq.Header.Set("X-Creator-Address", creatorAddress)

		// Forward the request
		client := &http.Client{}
		resp, err := client.Do(proxyReq)
		if err != nil {
			log.Error().Err(err).Msg("Error forwarding request")
			http.Error(w, "Error forwarding request", http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		// Copy response headers
		for header, values := range resp.Header {
			for _, value := range values {
				w.Header().Add(header, value)
			}
		}

		// Set response status code
		w.WriteHeader(resp.StatusCode)

		// Copy response body
		if _, err := io.Copy(w, resp.Body); err != nil {
			log.Error().Err(err).Msg("Failed to copy response body")
		}
	})

	// Start local proxy server
	localAddr := fmt.Sprintf("%s:%d", cfg.Server.Host, port)

	// Check if specified port is available
	if err := isPortAvailable(port); err != nil {
		log.Fatal().Err(err).Int("port", port).Msg("Chain proxy port is not available")
	}

	log.Info().
		Str("address", localAddr).
		Str("device_id", deviceID).
		Str("creator_address", creatorAddress).
		Int("port", port).
		Msg("Starting chain proxy server")

	if err := http.ListenAndServe(localAddr, nil); err != nil {
		log.Fatal().Err(err).Msg("Failed to start chain proxy server")
	}
}
