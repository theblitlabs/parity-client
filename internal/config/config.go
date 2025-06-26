package config

import (
	"fmt"
	"strings"
	"sync"

	"github.com/spf13/viper"
)

type Config struct {
	Server            ServerConfig            `mapstructure:"SERVER"`
	FilecoinNetwork   FilecoinNetworkConfig   `mapstructure:"FILECOIN_NETWORK"`
	Runner            RunnerConfig            `mapstructure:"RUNNER"`
	FederatedLearning FederatedLearningConfig `mapstructure:"FL"`
}

type ServerConfig struct {
	Host     string `mapstructure:"HOST"`
	Port     int    `mapstructure:"PORT"`
	Endpoint string `mapstructure:"ENDPOINT"`
}

type FilecoinNetworkConfig struct {
	RPC                string `mapstructure:"RPC"`
	ChainID            int64  `mapstructure:"CHAIN_ID"`
	TokenAddress       string `mapstructure:"TOKEN_ADDRESS"`
	StakeWalletAddress string `mapstructure:"STAKE_WALLET_ADDRESS"`
	IPFSEndpoint       string `mapstructure:"IPFS_ENDPOINT"`
	GatewayURL         string `mapstructure:"GATEWAY_URL"`
	CreateStorageDeals bool   `mapstructure:"CREATE_STORAGE_DEALS"`
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

	v.SetDefault("FILECOIN_NETWORK", map[string]interface{}{
		"RPC":                  v.GetString("FILECOIN_RPC"),
		"CHAIN_ID":             v.GetInt64("FILECOIN_CHAIN_ID"),
		"TOKEN_ADDRESS":        v.GetString("FILECOIN_TOKEN_ADDRESS"),
		"STAKE_WALLET_ADDRESS": v.GetString("FILECOIN_STAKE_WALLET_ADDRESS"),
		"IPFS_ENDPOINT":        v.GetString("FILECOIN_IPFS_ENDPOINT"),
		"GATEWAY_URL":          v.GetString("FILECOIN_GATEWAY_URL"),
		"CREATE_STORAGE_DEALS": v.GetBool("FILECOIN_CREATE_STORAGE_DEALS"),
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
