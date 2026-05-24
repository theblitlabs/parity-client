package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/theblitlabs/parity-client/internal/config"
)

func TestShouldHandleLocalDockerUpload(t *testing.T) {
	router := &RequestRouter{}

	tests := []struct {
		name        string
		path        string
		contentType string
		headerValue string
		want        bool
	}{
		{
			name:        "explicit task upload route is handled",
			path:        "/api/v1/tasks",
			contentType: "application/json",
			headerValue: "true",
			want:        true,
		},
		{
			name:        "dedicated local task upload route is handled",
			path:        "/api/v1/tasks/local",
			contentType: "application/json",
			want:        true,
		},
		{
			name:        "task creation without upload header is proxied",
			path:        "/api/v1/tasks",
			contentType: "application/json",
			want:        false,
		},
		{
			name:        "non task json route is proxied",
			path:        "/api/v1/llm/prompts",
			contentType: "application/json",
			headerValue: "true",
			want:        false,
		},
		{
			name:        "multipart requests are proxied",
			path:        "/api/v1/tasks",
			contentType: "multipart/form-data; boundary=test",
			headerValue: "true",
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, tt.path, nil)
			req.Header.Set("Content-Type", tt.contentType)
			if tt.headerValue != "" {
				req.Header.Set("X-Parity-Upload-Image", tt.headerValue)
			}

			got := router.shouldHandleLocalDockerUpload(req, req.URL.Path[len("/api/"):])
			if got != tt.want {
				t.Fatalf("shouldHandleLocalDockerUpload() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHandleRequestForwardsNonTaskJSONPost(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/llm/prompts" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method %q", r.Method)
		}

		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"Client ID is required"}`))
	}))
	defer backend.Close()

	router := NewRequestRouter(&config.Config{
		Runner: config.RunnerConfig{
			ServerURL: backend.URL,
		},
	}, "device-1", "creator-1")

	req := httptest.NewRequest(http.MethodPost, "/api/v1/llm/prompts", strings.NewReader(`{"prompt":"hi","model_name":"qwen","creator_address":"0xabc"}`))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	router.HandleRequest(rec, req)

	resp := rec.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status code = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}

	if !strings.Contains(string(body), "Client ID is required") {
		t.Fatalf("response body = %q, want proxied backend error", string(body))
	}
}

func TestHandleRequestServesProxyInfoAtRoot(t *testing.T) {
	router := NewRequestRouter(&config.Config{
		Server: config.ServerConfig{
			Host: "127.0.0.1",
			Port: 8090,
		},
		Runner: config.RunnerConfig{
			ServerURL: "http://127.0.0.1:8080",
		},
	}, "device-1", "creator-1")

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	router.HandleRequest(rec, req)

	resp := rec.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status code = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	if got := resp.Header.Get("X-Parity-Proxy-Mode"); got != "local-info" {
		t.Fatalf("X-Parity-Proxy-Mode = %q, want %q", got, "local-info")
	}

	var payload map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}

	if payload["local_proxy_url"] != "http://127.0.0.1:8090" {
		t.Fatalf("local_proxy_url = %v, want %q", payload["local_proxy_url"], "http://127.0.0.1:8090")
	}

	routes, ok := payload["routes"].([]any)
	if !ok || len(routes) == 0 {
		t.Fatalf("routes missing from proxy info response: %#v", payload["routes"])
	}
}

func TestHandleRequestReturnsActionableErrorForMissingLocalImage(t *testing.T) {
	router := NewRequestRouter(&config.Config{
		Runner: config.RunnerConfig{
			ServerURL: "http://127.0.0.1:8080",
		},
	}, "device-1", "creator-1")

	req := httptest.NewRequest(http.MethodPost, "/api/v1/tasks/local", strings.NewReader(`{"title":"missing-image"}`))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	router.HandleRequest(rec, req)

	resp := rec.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status code = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}

	var payload map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}

	if payload["message"] != "docker image is required for local image upload" {
		t.Fatalf("message = %v", payload["message"])
	}

	if _, ok := payload["hint"].(string); !ok {
		t.Fatalf("hint missing from error payload: %#v", payload)
	}

	nextSteps, ok := payload["next_steps"].([]any)
	if !ok || len(nextSteps) == 0 {
		t.Fatalf("next_steps missing from error payload: %#v", payload)
	}
}

func TestHandleRequestServesDashboard(t *testing.T) {
	router := NewRequestRouter(&config.Config{
		Server: config.ServerConfig{
			Host: "127.0.0.1",
			Port: 8090,
		},
		Runner: config.RunnerConfig{
			ServerURL: "http://127.0.0.1:8080",
		},
	}, "device-1", "creator-1")

	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	rec := httptest.NewRecorder()

	router.HandleRequest(rec, req)

	resp := rec.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status code = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	if got := resp.Header.Get("X-Parity-Proxy-Mode"); got != "local-dashboard" {
		t.Fatalf("X-Parity-Proxy-Mode = %q, want %q", got, "local-dashboard")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}

	if !strings.Contains(string(body), "Parity Operations Dashboard") {
		t.Fatalf("dashboard response missing title: %q", string(body))
	}
}
