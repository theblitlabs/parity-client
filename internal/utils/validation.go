package utils

import (
	"fmt"
	"math/big"
	"regexp"
	"strings"

	"github.com/ethereum/go-ethereum/common"
)

var (
	ethereumAddressRegex = regexp.MustCompile(`^0x[a-fA-F0-9]{40}$`)
	privateKeyRegex      = regexp.MustCompile(`^[a-fA-F0-9]{64}$`)
	urlRegex             = regexp.MustCompile(`^https?://[^\s/$.?#].[^\s]*$`)
)

type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

func ValidateEthereumAddress(address string) error {
	if address == "" {
		return ValidationError{Field: "address", Message: "address is required"}
	}

	if !ethereumAddressRegex.MatchString(address) {
		return ValidationError{Field: "address", Message: "invalid Ethereum address format"}
	}

	if !common.IsHexAddress(address) {
		return ValidationError{Field: "address", Message: "invalid Ethereum address checksum"}
	}

	return nil
}

func ValidatePrivateKey(privateKey string) error {
	if privateKey == "" {
		return ValidationError{Field: "private_key", Message: "private key is required"}
	}

	privateKey = strings.TrimPrefix(privateKey, "0x")

	if len(privateKey) != 64 {
		return ValidationError{Field: "private_key", Message: "private key must be 64 hex characters"}
	}

	if !privateKeyRegex.MatchString(privateKey) {
		return ValidationError{Field: "private_key", Message: "invalid private key format"}
	}

	return nil
}

func ValidateAmount(amount float64, minAmount float64) error {
	if amount <= 0 {
		return ValidationError{Field: "amount", Message: "amount must be greater than 0"}
	}

	if amount < minAmount {
		return ValidationError{Field: "amount", Message: fmt.Sprintf("amount must be at least %f", minAmount)}
	}

	return nil
}

func ValidateURL(url string) error {
	if url == "" {
		return ValidationError{Field: "url", Message: "URL is required"}
	}

	if !urlRegex.MatchString(url) {
		return ValidationError{Field: "url", Message: "invalid URL format"}
	}

	return nil
}

func ValidateDeviceID(deviceID string) error {
	if deviceID == "" {
		return ValidationError{Field: "device_id", Message: "device ID is required"}
	}

	if len(deviceID) < 10 {
		return ValidationError{Field: "device_id", Message: "device ID is too short"}
	}

	return nil
}

func ValidateTaskTitle(title string) error {
	if title == "" {
		return ValidationError{Field: "title", Message: "task title is required"}
	}

	if len(title) > 255 {
		return ValidationError{Field: "title", Message: "task title is too long (max 255 characters)"}
	}

	return nil
}

func ValidateTaskDescription(description string) error {
	if description == "" {
		return ValidationError{Field: "description", Message: "task description is required"}
	}

	if len(description) > 1000 {
		return ValidationError{Field: "description", Message: "task description is too long (max 1000 characters)"}
	}

	return nil
}

func ValidateDockerImage(image string) error {
	if image == "" {
		return ValidationError{Field: "image", Message: "Docker image is required"}
	}

	if !strings.Contains(image, "/") && !strings.Contains(image, ":") {
		return ValidationError{Field: "image", Message: "invalid Docker image format"}
	}

	return nil
}

func ValidateReward(reward float64) error {
	if reward <= 0 {
		return ValidationError{Field: "reward", Message: "reward must be greater than 0"}
	}

	maxReward := new(big.Float).SetFloat64(1e18)
	if new(big.Float).SetFloat64(reward).Cmp(maxReward) > 0 {
		return ValidationError{Field: "reward", Message: "reward is too large"}
	}

	return nil
}

func ValidateModelName(modelName string) error {
	if modelName == "" {
		return ValidationError{Field: "model_name", Message: "model name is required"}
	}

	if len(modelName) > 100 {
		return ValidationError{Field: "model_name", Message: "model name is too long (max 100 characters)"}
	}

	return nil
}

func ValidatePrompt(prompt string) error {
	if prompt == "" {
		return ValidationError{Field: "prompt", Message: "prompt is required"}
	}

	if len(prompt) > 10000 {
		return ValidationError{Field: "prompt", Message: "prompt is too long (max 10000 characters)"}
	}

	return nil
}

func ValidateMaxTokens(maxTokens int) error {
	if maxTokens <= 0 {
		return ValidationError{Field: "max_tokens", Message: "max tokens must be greater than 0"}
	}

	if maxTokens > 100000 {
		return ValidationError{Field: "max_tokens", Message: "max tokens is too large (max 100000)"}
	}

	return nil
}

func ValidateFLSessionName(name string) error {
	if name == "" {
		return ValidationError{Field: "name", Message: "session name is required"}
	}

	if len(name) > 255 {
		return ValidationError{Field: "name", Message: "session name is too long (max 255 characters)"}
	}

	return nil
}

func ValidateTotalRounds(totalRounds int) error {
	if totalRounds <= 0 {
		return ValidationError{Field: "total_rounds", Message: "total rounds must be greater than 0"}
	}

	if totalRounds > 1000 {
		return ValidationError{Field: "total_rounds", Message: "total rounds is too large (max 1000)"}
	}

	return nil
}

func ValidateMinParticipants(minParticipants int) error {
	if minParticipants <= 0 {
		return ValidationError{Field: "min_participants", Message: "minimum participants must be greater than 0"}
	}

	if minParticipants > 100 {
		return ValidationError{Field: "min_participants", Message: "minimum participants is too large (max 100)"}
	}

	return nil
}

func ValidateDatasetCID(datasetCID string) error {
	if datasetCID == "" {
		return ValidationError{Field: "dataset_cid", Message: "dataset CID is required"}
	}

	if len(datasetCID) < 10 {
		return ValidationError{Field: "dataset_cid", Message: "invalid dataset CID format"}
	}

	return nil
}

func ValidateLearningRate(learningRate float64) error {
	if learningRate <= 0 {
		return ValidationError{Field: "learning_rate", Message: "learning rate must be greater than 0"}
	}

	if learningRate > 1.0 {
		return ValidationError{Field: "learning_rate", Message: "learning rate is too large (max 1.0)"}
	}

	return nil
}

func ValidateBatchSize(batchSize int) error {
	if batchSize <= 0 {
		return ValidationError{Field: "batch_size", Message: "batch size must be greater than 0"}
	}

	if batchSize > 10000 {
		return ValidationError{Field: "batch_size", Message: "batch size is too large (max 10000)"}
	}

	return nil
}

func ValidateLocalEpochs(localEpochs int) error {
	if localEpochs <= 0 {
		return ValidationError{Field: "local_epochs", Message: "local epochs must be greater than 0"}
	}

	if localEpochs > 1000 {
		return ValidationError{Field: "local_epochs", Message: "local epochs is too large (max 1000)"}
	}

	return nil
}
