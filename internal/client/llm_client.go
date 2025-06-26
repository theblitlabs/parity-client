package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/theblitlabs/gologger"
)

type LLMClient struct {
	serverURL string
	clientID  string
	client    *http.Client
}

type PromptRequest struct {
	Prompt    string `json:"prompt"`
	ModelName string `json:"model_name"`
}

type PromptResponse struct {
	ID          string  `json:"id"`
	Response    string  `json:"response"`
	Status      string  `json:"status"`
	ModelName   string  `json:"model_name"`
	CreatedAt   string  `json:"created_at"`
	CompletedAt *string `json:"completed_at,omitempty"`
}

type BillingMetricsResponse struct {
	TotalRequests    int     `json:"total_requests"`
	TotalTokens      int     `json:"total_tokens"`
	TotalCost        float64 `json:"total_cost"`
	AvgInferenceTime float64 `json:"avg_inference_time_ms"`
}

type ModelInfo struct {
	ModelName string `json:"model_name"`
	MaxTokens int    `json:"max_tokens"`
	IsLoaded  bool   `json:"is_loaded"`
}

type ModelsResponse struct {
	Models []ModelInfo `json:"models"`
	Count  int         `json:"count"`
}

func NewLLMClient(serverURL, clientID string) *LLMClient {
	return &LLMClient{
		serverURL: serverURL,
		clientID:  clientID,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *LLMClient) SubmitPrompt(ctx context.Context, prompt, modelName string) (*PromptResponse, error) {
	log := gologger.WithComponent("llm_client")

	req := PromptRequest{
		Prompt:    prompt,
		ModelName: modelName,
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/llm/prompts", c.serverURL)
	log.Info().Str("full_url", url).Msg("Making request to URL")
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Client-ID", c.clientID)

	log.Info().
		Str("model_name", modelName).
		Str("prompt_preview", truncateString(prompt, 100)).
		Msg("Submitting prompt request")

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to submit prompt: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		return nil, fmt.Errorf("prompt submission failed with status: %d", resp.StatusCode)
	}

	var response PromptResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	log.Info().
		Str("prompt_id", response.ID).
		Str("status", response.Status).
		Msg("Prompt submitted successfully")

	return &response, nil
}

func (c *LLMClient) GetPrompt(ctx context.Context, promptID string) (*PromptResponse, error) {
	url := fmt.Sprintf("%s/api/llm/prompts/%s", c.serverURL, promptID)
	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("X-Client-ID", c.clientID)

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to get prompt: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get prompt failed with status: %d", resp.StatusCode)
	}

	var response PromptResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &response, nil
}

func (c *LLMClient) ListPrompts(ctx context.Context, limit, offset int) ([]*PromptResponse, error) {
	url := fmt.Sprintf("%s/api/llm/prompts?limit=%d&offset=%d", c.serverURL, limit, offset)
	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("X-Client-ID", c.clientID)

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to list prompts: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("list prompts failed with status: %d", resp.StatusCode)
	}

	var response struct {
		Prompts []*PromptResponse `json:"prompts"`
		Limit   int               `json:"limit"`
		Offset  int               `json:"offset"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return response.Prompts, nil
}

func (c *LLMClient) WaitForCompletion(ctx context.Context, promptID string, pollInterval time.Duration) (*PromptResponse, error) {
	log := gologger.WithComponent("llm_client")

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			prompt, err := c.GetPrompt(ctx, promptID)
			if err != nil {
				log.Error().Err(err).Str("prompt_id", promptID).Msg("Failed to check prompt status")
				continue
			}

			if prompt.Status == "completed" || prompt.Status == "failed" {
				return prompt, nil
			}
		}
	}
}

func (c *LLMClient) GetBillingMetrics(ctx context.Context) (*BillingMetricsResponse, error) {
	url := fmt.Sprintf("%s/api/llm/billing/metrics", c.serverURL)
	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("X-Client-ID", c.clientID)

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to get billing metrics: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get billing metrics failed with status: %d", resp.StatusCode)
	}

	var response BillingMetricsResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &response, nil
}

func (c *LLMClient) GetAvailableModels(ctx context.Context) (*ModelsResponse, error) {
	url := fmt.Sprintf("%s/api/llm/models", c.serverURL)
	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to get available models: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get available models failed with status: %d", resp.StatusCode)
	}

	var response ModelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &response, nil
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
