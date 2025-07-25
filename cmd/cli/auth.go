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
	"github.com/theblitlabs/parity-client/internal/utils"
)

func RunAuth(cmd *cobra.Command, args []string) {
	log := log.With().Str("component", "auth").Logger()

	privateKey, _ := cmd.Flags().GetString("private-key")
	configPath, _ := cmd.Flags().GetString("config-path")

	if err := ExecuteAuth(privateKey, configPath); err != nil {
		log.Fatal().Err(err).Msg("Failed to authenticate")
	}
}

func ExecuteAuth(privateKey string, configPath string) error {
	log := log.With().Str("component", "auth").Logger()

	if err := utils.ValidatePrivateKey(privateKey); err != nil {
		return err
	}

	configManager := config.NewConfigManager(configPath)
	cfg, err := configManager.GetConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	privateKey = strings.TrimPrefix(privateKey, "0x")

	_, err = crypto.HexToECDSA(privateKey)
	if err != nil {
		return fmt.Errorf("invalid private key format: %w", err)
	}

	keystoreAdapter, err := keystore.NewAdapter(nil)
	if err != nil {
		return fmt.Errorf("failed to create keystore: %w", err)
	}

	if err := keystoreAdapter.SavePrivateKey(privateKey); err != nil {
		return fmt.Errorf("failed to save private key: %w", err)
	}

	walletAdapter, err := wallet.NewAdapter(walletsdk.ClientConfig{
		RPCURL:       cfg.BlockchainNetwork.RPC,
		ChainID:      cfg.BlockchainNetwork.ChainID,
		PrivateKey:   privateKey,
		TokenAddress: common.HexToAddress(cfg.BlockchainNetwork.TokenAddress),
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
