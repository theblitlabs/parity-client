package config

import (
	"fmt"
	"strings"
	"sync"

	"github.com/spf13/viper"
)

type Config struct {
	Server            ServerConfig            `mapstructure:"SERVER"`
	BlockchainNetwork BlockchainNetworkConfig `mapstructure:"BLOCKCHAIN_NETWORK"`
	Runner            RunnerConfig            `mapstructure:"RUNNER"`
	FederatedLearning FederatedLearningConfig `mapstructure:"FL"`
}

type ServerConfig struct {
	Host     string `mapstructure:"HOST"`
	Port     int    `mapstructure:"PORT"`
	Endpoint string `mapstructure:"ENDPOINT"`
}

type BlockchainNetworkConfig struct {
	RPC                string `mapstructure:"RPC"`
	ChainID            int64  `mapstructure:"CHAIN_ID"`
	TokenAddress       string `mapstructure:"TOKEN_ADDRESS"`
	TokenSymbol        string `mapstructure:"TOKEN_SYMBOL"`
	StakeWalletAddress string `mapstructure:"STAKE_WALLET_ADDRESS"`
	IPFSEndpoint       string `mapstructure:"IPFS_ENDPOINT"`
	GatewayURL         string `mapstructure:"GATEWAY_URL"`
}

type FederatedLearningConfig struct {
	ServerURL      string `mapstructure:"SERVER_URL"`
	DefaultTimeout string `mapstructure:"DEFAULT_TIMEOUT"`
	RetryAttempts  int    `mapstructure:"RETRY_ATTEMPTS"`
	LogLevel       string `mapstructure:"LOG_LEVEL"`
}

type RunnerConfig struct {
	ServerURL   string `mapstructure:"SERVER_URL"`
	WebhookPort int    `mapstructure:"WEBHOOK_PORT"`
	APIPrefix   string `mapstructure:"API_PREFIX"`
}

type ConfigManager struct {
	config     *Config
	configPath string
	mutex      sync.RWMutex
}

func NewConfigManager(configPath string) *ConfigManager {
	return &ConfigManager{
		configPath: configPath,
	}
}

func (cm *ConfigManager) SetConfigPath(path string) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	cm.configPath = path
	cm.config = nil
}

func (cm *ConfigManager) GetConfig() (*Config, error) {
	cm.mutex.RLock()
	if cm.config != nil {
		defer cm.mutex.RUnlock()
		return cm.config, nil
	}
	cm.mutex.RUnlock()

	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	if cm.config != nil {
		return cm.config, nil
	}

	var err error
	cm.config, err = loadConfigFile(cm.configPath)
	return cm.config, err
}

func loadConfigFile(path string) (*Config, error) {
	v := viper.New()

	v.SetConfigFile(path)
	v.SetEnvPrefix("")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	v.SetDefault("SERVER", map[string]interface{}{
		"HOST":     v.GetString("SERVER_HOST"),
		"PORT":     v.GetInt("SERVER_PORT"),
		"ENDPOINT": v.GetString("SERVER_ENDPOINT"),
	})

	v.SetDefault("BLOCKCHAIN_NETWORK", map[string]interface{}{
		"RPC":                  v.GetString("BLOCKCHAIN_RPC"),
		"CHAIN_ID":             v.GetInt64("BLOCKCHAIN_CHAIN_ID"),
		"TOKEN_ADDRESS":        v.GetString("TOKEN_ADDRESS"),
		"TOKEN_SYMBOL":         v.GetString("TOKEN_SYMBOL"),
		"STAKE_WALLET_ADDRESS": v.GetString("STAKE_WALLET_ADDRESS"),
		"IPFS_ENDPOINT":        v.GetString("IPFS_ENDPOINT"),
		"GATEWAY_URL":          v.GetString("GATEWAY_URL"),
	})

	v.SetDefault("RUNNER", map[string]interface{}{
		"SERVER_URL":   v.GetString("RUNNER_SERVER_URL"),
		"WEBHOOK_PORT": v.GetInt("RUNNER_WEBHOOK_PORT"),
		"API_PREFIX":   v.GetString("RUNNER_API_PREFIX"),
	})

	v.SetDefault("FL", map[string]interface{}{
		"SERVER_URL":      v.GetString("FL_SERVER_URL"),
		"DEFAULT_TIMEOUT": v.GetString("FL_DEFAULT_TIMEOUT"),
		"RETRY_ATTEMPTS":  v.GetInt("FL_RETRY_ATTEMPTS"),
		"LOG_LEVEL":       v.GetString("FL_LOG_LEVEL"),
	})

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("unable to decode into config struct: %w", err)
	}

	return &config, nil
}

func (cm *ConfigManager) GetConfigPath() string {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	return cm.configPath
}
