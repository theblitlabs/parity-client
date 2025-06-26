package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

type FederatedLearningClient struct {
	baseURL string
	client  *http.Client
}

type CreateFLSessionRequest struct {
	Name            string           `json:"name"`
	Description     string           `json:"description"`
	ModelType       string           `json:"model_type"`
	TotalRounds     int              `json:"total_rounds"`
	MinParticipants int              `json:"min_participants"`
	CreatorAddress  string           `json:"creator_address"`
	TrainingData    TrainingDataInfo `json:"training_data"`
	Config          FLConfigRequest  `json:"config"`
}

type TrainingDataInfo struct {
	DatasetCID    string   `json:"dataset_cid"`
	DatasetSize   int64    `json:"dataset_size,omitempty"`
	DataFormat    string   `json:"data_format"`
	SplitStrategy string   `json:"split_strategy"`
	Features      []string `json:"features,omitempty"`
	Labels        []string `json:"labels,omitempty"`
}

type FLConfigRequest struct {
	AggregationMethod string                 `json:"aggregation_method"`
	LearningRate      float64                `json:"learning_rate"`
	BatchSize         int                    `json:"batch_size"`
	LocalEpochs       int                    `json:"local_epochs"`
	ClientSelection   string                 `json:"client_selection"`
	ModelConfig       map[string]interface{} `json:"model_config"`
}

type FLSession struct {
	ID               string   `json:"id"`
	Name             string   `json:"name"`
	Description      string   `json:"description"`
	ModelType        string   `json:"model_type"`
	Status           string   `json:"status"`
	TotalRounds      int      `json:"total_rounds"`
	CurrentRound     int      `json:"current_round"`
	MinParticipants  int      `json:"min_participants"`
	ParticipantCount int      `json:"participant_count"`
	CreatorAddress   string   `json:"creator_address"`
	CreatedAt        string   `json:"created_at"`
	UpdatedAt        string   `json:"updated_at"`
	CompletedAt      *string  `json:"completed_at,omitempty"`
	Config           FLConfig `json:"config"`
}

type FLConfig struct {
	AggregationMethod string                 `json:"aggregation_method"`
	LearningRate      float64                `json:"learning_rate"`
	BatchSize         int                    `json:"batch_size"`
	LocalEpochs       int                    `json:"local_epochs"`
	ClientSelection   string                 `json:"client_selection"`
	ModelConfig       map[string]interface{} `json:"model_config"`
}

type ListSessionsResponse struct {
	Sessions []FLSession `json:"sessions"`
	Count    int         `json:"count"`
}

type SubmitModelUpdateRequest struct {
	SessionID    string                 `json:"session_id"`
	RoundID      string                 `json:"round_id"`
	RunnerID     string                 `json:"runner_id"`
	Gradients    map[string][]float64   `json:"gradients"`
	UpdateType   string                 `json:"update_type"`
	DataSize     int                    `json:"data_size"`
	Loss         float64                `json:"loss"`
	Accuracy     float64                `json:"accuracy"`
	TrainingTime int                    `json:"training_time"`
	Metadata     map[string]interface{} `json:"metadata"`
}

func NewFederatedLearningClient(baseURL string) *FederatedLearningClient {
	return &FederatedLearningClient{
		baseURL: baseURL,
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *FederatedLearningClient) CreateSession(ctx context.Context, req *CreateFLSessionRequest) (*FLSession, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	baseURL := strings.TrimSuffix(c.baseURL, "/api")
	resp, err := c.client.Post(baseURL+"/api/v1/federated-learning/sessions", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
	}

	var session FLSession
	if err := json.NewDecoder(resp.Body).Decode(&session); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return &session, nil
}

func (c *FederatedLearningClient) ListSessions(ctx context.Context, creator string) (*ListSessionsResponse, error) {
	baseURL := strings.TrimSuffix(c.baseURL, "/api")
	url := baseURL + "/api/v1/federated-learning/sessions"
	if creator != "" {
		url += "?creator=" + creator
	}

	resp, err := c.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to list sessions: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
	}

	var response ListSessionsResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return &response, nil
}

func (c *FederatedLearningClient) GetSession(ctx context.Context, sessionID string) (*FLSession, error) {
	baseURL := strings.TrimSuffix(c.baseURL, "/api")
	resp, err := c.client.Get(baseURL + "/api/v1/federated-learning/sessions/" + sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("session not found")
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
	}

	var session FLSession
	if err := json.NewDecoder(resp.Body).Decode(&session); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return &session, nil
}

func (c *FederatedLearningClient) StartSession(ctx context.Context, sessionID string) error {
	url := strings.TrimSuffix(c.baseURL, "/api")
	resp, err := c.client.Post(url+"/api/v1/federated-learning/sessions/"+sessionID+"/start", "application/json", nil)
	log.Println("baseURL", url+"/api/v1/federated-learning/sessions/"+sessionID+"/start")
	if err != nil {
		return fmt.Errorf("failed to start session: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func (c *FederatedLearningClient) SubmitModelUpdate(ctx context.Context, req *SubmitModelUpdateRequest) error {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	baseURL := strings.TrimSuffix(c.baseURL, "/api")
	resp, err := c.client.Post(baseURL+"/api/v1/federated-learning/model-updates", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to submit model update: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func (c *FederatedLearningClient) GetTrainedModel(ctx context.Context, sessionID string) (map[string]interface{}, error) {
	baseURL := strings.TrimSuffix(c.baseURL, "/api")
	resp, err := c.client.Get(baseURL + "/api/v1/federated-learning/sessions/" + sessionID + "/model")
	if err != nil {
		return nil, fmt.Errorf("failed to get trained model: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("model not found: %s", string(body))
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
	}

	var model map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&model); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return model, nil
}
