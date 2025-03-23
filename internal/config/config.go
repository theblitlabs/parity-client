package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Ethereum EthereumConfig `mapstructure:"ethereum"`
	Runner   RunnerConfig   `mapstructure:"runner"`
}

type ServerConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Endpoint string `mapstructure:"endpoint"`
}

type EthereumConfig struct {
	RPC                string `mapstructure:"rpc"`
	ChainID            int64  `mapstructure:"chain_id"`
	TokenAddress       string `mapstructure:"token_address"`
	StakeWalletAddress string `mapstructure:"stake_wallet_address"`
}

type RunnerConfig struct {
	ServerURL   string `mapstructure:"server_url"`
	WebhookPort int    `mapstructure:"webhook_port"`
	APIPrefix   string `mapstructure:"api_prefix"`
}

func LoadConfig(path string) (*Config, error) {
	viper.SetConfigFile(path)
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
