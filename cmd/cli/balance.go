package cli

import (
	"context"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/theblitlabs/deviceid"
	walletsdk "github.com/theblitlabs/go-wallet-sdk"
	"github.com/theblitlabs/gologger"
	"github.com/theblitlabs/parity-client/internal/adapters/keystore"
	"github.com/theblitlabs/parity-client/internal/adapters/wallet"
	"github.com/theblitlabs/parity-client/internal/config"
)

func RunBalance(cmd *cobra.Command, args []string) {
	configPath, _ := cmd.Flags().GetString("config-path")
	if err := executeBalance(configPath); err != nil {
		log := gologger.Get()
		log.Error().Err(err).Msg("Balance check failed")
	}
}

func executeBalance(configPath string) error {
	log := gologger.Get().With().Str("component", "balance").Logger()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cfg, walletAdapter, err := initializeWallet(configPath, log)
	if err != nil {
		return err
	}

	if err := displayTokenBalance(ctx, walletAdapter, cfg, log); err != nil {
		return err
	}

	deviceID, err := getDeviceID(log)
	if err != nil {
		return err
	}

	return displayStakeInfo(ctx, walletAdapter, cfg, deviceID, log)
}

func initializeWallet(configPath string, log zerolog.Logger) (*config.Config, *wallet.Adapter, error) {
	configManager := config.NewConfigManager(configPath)
	cfg, err := configManager.GetConfig()
	if err != nil {
		log.Error().Err(err).Msg("Failed to load config")
		return nil, nil, err
	}

	keystoreAdapter, err := keystore.NewAdapter(nil)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create keystore")
		return nil, nil, err
	}

	privateKey, err := keystoreAdapter.LoadPrivateKey()
	if err != nil {
		log.Error().Err(err).Msg("No private key found - please authenticate first using 'parity auth'")
		return nil, nil, err
	}

	walletAdapter, err := wallet.NewAdapter(walletsdk.ClientConfig{
		RPCURL:       cfg.BlockchainNetwork.RPC,
		ChainID:      cfg.BlockchainNetwork.ChainID,
		PrivateKey:   common.Bytes2Hex(crypto.FromECDSA(privateKey)),
		TokenAddress: common.HexToAddress(cfg.BlockchainNetwork.TokenAddress),
		StakeAddress: common.HexToAddress(cfg.BlockchainNetwork.StakeWalletAddress),
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to create Ethereum client")
		return nil, nil, err
	}

	return cfg, walletAdapter, nil
}

func displayTokenBalance(ctx context.Context, walletAdapter *wallet.Adapter, cfg *config.Config, log zerolog.Logger) error {
	token, err := walletAdapter.NewParityToken(common.HexToAddress(cfg.BlockchainNetwork.TokenAddress))
	if err != nil {
		log.Error().Err(err).Msg("Failed to create token contract")
		return err
	}

	tokenBalance, err := walletAdapter.GetTokenBalance(ctx, token, walletAdapter.GetAddress())
	if err != nil {
		select {
		case <-ctx.Done():
			log.Error().Err(ctx.Err()).Msg("Operation timed out while getting token balance")
			return ctx.Err()
		default:
			log.Error().Err(err).Msg("Failed to get token balance")
			return err
		}
	}

	log.Info().
		Str("wallet_address", walletAdapter.GetAddress().Hex()).
		Str("balance", tokenBalance.String()+" "+cfg.BlockchainNetwork.TokenSymbol).
		Msg("Wallet token balance")

	return nil
}

func getDeviceID(log zerolog.Logger) (string, error) {
	deviceIDManager := deviceid.NewManager(deviceid.Config{})
	deviceID, err := deviceIDManager.VerifyDeviceID()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get device ID")
		return "", err
	}
	return deviceID, nil
}

func displayStakeInfo(ctx context.Context, walletAdapter *wallet.Adapter, cfg *config.Config, deviceID string, log zerolog.Logger) error {
	stakeWallet, err := walletAdapter.NewStakeWallet(
		common.HexToAddress(cfg.BlockchainNetwork.StakeWalletAddress),
		common.HexToAddress(cfg.BlockchainNetwork.TokenAddress),
	)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create stake wallet contract")
		return err
	}

	stakeInfo, err := walletAdapter.GetStakeInfo(ctx, stakeWallet, deviceID)
	if err != nil {
		select {
		case <-ctx.Done():
			log.Error().Err(ctx.Err()).Msg("Operation timed out while getting stake info")
			return ctx.Err()
		default:
			log.Error().Err(err).Msg("Failed to get stake info")
			return err
		}
	}

	if stakeInfo.Exists {
		return displayExistingStakeInfo(ctx, walletAdapter, cfg, stakeInfo, log)
	} else {
		log.Info().Msg("No active stake found")
		return nil
	}
}

func displayExistingStakeInfo(ctx context.Context, walletAdapter *wallet.Adapter, cfg *config.Config, stakeInfo walletsdk.StakeInfo, log zerolog.Logger) error {
	log.Info().
		Str("amount", stakeInfo.Amount.String()+" PRTY").
		Str("device_id", stakeInfo.DeviceID).
		Str("wallet_address", stakeInfo.WalletAddress.Hex()).
		Msg("Current stake info")

	token, err := walletAdapter.NewParityToken(common.HexToAddress(cfg.BlockchainNetwork.TokenAddress))
	if err != nil {
		log.Error().Err(err).Msg("Failed to create token contract")
		return err
	}

	contractBalance, err := walletAdapter.GetTokenBalance(ctx, token, common.HexToAddress(cfg.BlockchainNetwork.StakeWalletAddress))
	if err != nil {
		select {
		case <-ctx.Done():
			log.Error().Err(ctx.Err()).Msg("Operation timed out while getting contract balance")
			return ctx.Err()
		default:
			log.Error().Err(err).Msg("Failed to get contract balance")
			return err
		}
	}

	log.Info().
		Str("balance", contractBalance.String()+" PRTY").
		Str("contract_address", cfg.BlockchainNetwork.StakeWalletAddress).
		Msg("Contract token balance")

	return nil
}
