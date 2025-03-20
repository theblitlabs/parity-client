package cli

import (
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	walletsdk "github.com/theblitlabs/go-wallet-sdk"
	"github.com/theblitlabs/parity-client/internal/adapters/keystore"
	"github.com/theblitlabs/parity-client/internal/adapters/wallet"
	"github.com/theblitlabs/parity-client/internal/config"
)

func RunAuth() {
	var privateKey string

	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Authenticate with the network",
		Run: func(cmd *cobra.Command, args []string) {
			if err := ExecuteAuth(privateKey, "config/config.yaml"); err != nil {
				log.Fatal().Err(err).Msg("Failed to authenticate")
			}
		},
	}

	cmd.Flags().StringVarP(&privateKey, "private-key", "k", "", "Private key in hex format")
	if err := cmd.MarkFlagRequired("private-key"); err != nil {
		log.Fatal().Err(err).Msg("Failed to mark flag as required")
	}

	if err := cmd.Execute(); err != nil {
		log.Fatal().Err(err).Msg("Failed to execute auth command")
	}
}

func ExecuteAuth(privateKey string, configPath string) error {
	log := log.With().Str("component", "auth").Logger()

	if privateKey == "" {
		return fmt.Errorf("private key is required")
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	privateKey = strings.TrimPrefix(privateKey, "0x")

	if len(privateKey) != 64 {
		return fmt.Errorf("invalid private key - must be 64 hex characters without 0x prefix")
	}

	_, err = crypto.HexToECDSA(privateKey)
	if err != nil {
		return fmt.Errorf("invalid private key format: %w", err)
	}

	keystoreAdapter, err := keystore.NewAdapter(nil) // Uses default config
	if err != nil {
		return fmt.Errorf("failed to create keystore: %w", err)
	}

	if err := keystoreAdapter.SavePrivateKey(privateKey); err != nil {
		return fmt.Errorf("failed to save private key: %w", err)
	}

	walletAdapter, err := wallet.NewAdapter(walletsdk.ClientConfig{
		RPCURL:       cfg.Ethereum.RPC,
		ChainID:      cfg.Ethereum.ChainID,
		PrivateKey:   privateKey,
		TokenAddress: common.HexToAddress(cfg.Ethereum.TokenAddress),
	})
	if err != nil {
		return fmt.Errorf("invalid private key: %w", err)
	}

	log.Info().
		Str("address", walletAdapter.GetAddress().Hex()).
		Str("keystore", fmt.Sprintf("%s/%s", keystore.DefaultDirName, keystore.DefaultFileName)).
		Msg("Wallet authenticated successfully")

	return nil
}
