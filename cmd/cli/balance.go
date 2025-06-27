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
	executeBalance(configPath)
}

func executeBalance(configPath string) {
	log := gologger.Get().With().Str("component", "balance").Logger()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cfg, walletAdapter, err := initializeWallet(configPath, log)
	if err != nil {
		return // Error already logged in initializeWallet
	}

	displayTokenBalance(ctx, walletAdapter, cfg, log)

	deviceID, err := getDeviceID(log)
	if err != nil {
		return
	}

	displayStakeInfo(ctx, walletAdapter, cfg, deviceID, log)
}

func initializeWallet(configPath string, log zerolog.Logger) (*config.Config, *wallet.Adapter, error) {
	configManager := config.NewConfigManager(configPath)
	cfg, err := configManager.GetConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load config")
		return nil, nil, err
	}

	keystoreAdapter, err := keystore.NewAdapter(nil)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create keystore")
		return nil, nil, err
	}

	privateKey, err := keystoreAdapter.LoadPrivateKey()
	if err != nil {
		log.Fatal().Err(err).Msg("No private key found - please authenticate first using 'parity auth'")
		return nil, nil, err
	}

	walletAdapter, err := wallet.NewAdapter(walletsdk.ClientConfig{
		RPCURL:       cfg.FilecoinNetwork.RPC,
		ChainID:      cfg.FilecoinNetwork.ChainID,
		PrivateKey:   common.Bytes2Hex(crypto.FromECDSA(privateKey)),
		TokenAddress: common.HexToAddress(cfg.FilecoinNetwork.TokenAddress),
		StakeAddress: common.HexToAddress(cfg.FilecoinNetwork.StakeWalletAddress),
	})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create Filecoin client")
		return nil, nil, err
	}

	return cfg, walletAdapter, nil
}

func displayTokenBalance(ctx context.Context, walletAdapter *wallet.Adapter, cfg *config.Config, log zerolog.Logger) {
	token, err := walletAdapter.NewParityToken(common.HexToAddress(cfg.FilecoinNetwork.TokenAddress))
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create token contract")
		return
	}

	tokenBalance, err := walletAdapter.GetTokenBalance(ctx, token, walletAdapter.GetAddress())
	if err != nil {
		select {
		case <-ctx.Done():
			log.Fatal().Err(ctx.Err()).Msg("Operation timed out while getting token balance")
		default:
			log.Fatal().Err(err).Msg("Failed to get token balance")
		}
		return
	}

	log.Info().
		Str("wallet_address", walletAdapter.GetAddress().Hex()).
		Str("balance", tokenBalance.String()+" USDFC").
		Msg("Wallet token balance")
}

func getDeviceID(log zerolog.Logger) (string, error) {
	deviceIDManager := deviceid.NewManager(deviceid.Config{})
	deviceID, err := deviceIDManager.VerifyDeviceID()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to get device ID")
		return "", err
	}
	return deviceID, nil
}

func displayStakeInfo(ctx context.Context, walletAdapter *wallet.Adapter, cfg *config.Config, deviceID string, log zerolog.Logger) {
	stakeWallet, err := walletAdapter.NewStakeWallet(
		common.HexToAddress(cfg.FilecoinNetwork.StakeWalletAddress),
		common.HexToAddress(cfg.FilecoinNetwork.TokenAddress),
	)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create stake wallet contract")
		return
	}

	stakeInfo, err := walletAdapter.GetStakeInfo(ctx, stakeWallet, deviceID)
	if err != nil {
		select {
		case <-ctx.Done():
			log.Fatal().Err(ctx.Err()).Msg("Operation timed out while getting stake info")
		default:
			log.Fatal().Err(err).Msg("Failed to get stake info")
		}
		return
	}

	if stakeInfo.Exists {
		displayExistingStakeInfo(ctx, walletAdapter, cfg, stakeInfo, log)
	} else {
		log.Info().Msg("No active stake found")
	}
}

func displayExistingStakeInfo(ctx context.Context, walletAdapter *wallet.Adapter, cfg *config.Config, stakeInfo walletsdk.StakeInfo, log zerolog.Logger) {
	log.Info().
		Str("amount", stakeInfo.Amount.String()+" USDFC").
		Str("device_id", stakeInfo.DeviceID).
		Str("wallet_address", stakeInfo.WalletAddress.Hex()).
		Msg("Current stake info")

	token, err := walletAdapter.NewParityToken(common.HexToAddress(cfg.FilecoinNetwork.TokenAddress))
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create token contract")
		return
	}

	contractBalance, err := walletAdapter.GetTokenBalance(ctx, token, common.HexToAddress(cfg.FilecoinNetwork.StakeWalletAddress))
	if err != nil {
		select {
		case <-ctx.Done():
			log.Fatal().Err(ctx.Err()).Msg("Operation timed out while getting contract balance")
		default:
			log.Fatal().Err(err).Msg("Failed to get contract balance")
		}
		return
	}

	log.Info().
		Str("balance", contractBalance.String()+" USDFC").
		Str("contract_address", cfg.FilecoinNetwork.StakeWalletAddress).
		Msg("Contract token balance")
}
