package cli

import (
	"context"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/theblitlabs/deviceid"
	walletsdk "github.com/theblitlabs/go-wallet-sdk"
	"github.com/theblitlabs/gologger"
	"github.com/theblitlabs/parity-client/internal/adapters/keystore"
	"github.com/theblitlabs/parity-client/internal/adapters/wallet"
	"github.com/theblitlabs/parity-client/internal/config"
)

func RunBalance() {
	log := gologger.Get().With().Str("component", "balance").Logger()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cfg, err := config.LoadConfig("config/config.yaml")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load config")
	}

	// Create keystore adapter
	keystoreAdapter, err := keystore.NewAdapter(nil)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create keystore")
	}

	privateKey, err := keystoreAdapter.LoadPrivateKey()
	if err != nil {
		log.Fatal().Err(err).Msg("No private key found - please authenticate first using 'parity auth'")
	}

	// Create wallet adapter
	walletAdapter, err := wallet.NewAdapter(walletsdk.ClientConfig{
		RPCURL:       cfg.Ethereum.RPC,
		ChainID:      cfg.Ethereum.ChainID,
		PrivateKey:   common.Bytes2Hex(crypto.FromECDSA(privateKey)),
		TokenAddress: common.HexToAddress(cfg.Ethereum.TokenAddress),
		StakeAddress: common.HexToAddress(cfg.Ethereum.StakeWalletAddress),
	})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create Ethereum client")
	}

	// Create token contract
	token, err := walletAdapter.NewParityToken(common.HexToAddress(cfg.Ethereum.TokenAddress))
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create token contract")
	}

	// Get token balance
	tokenBalance, err := walletAdapter.GetTokenBalance(ctx, token, walletAdapter.GetAddress())
	if err != nil {
		select {
		case <-ctx.Done():
			log.Fatal().Err(ctx.Err()).Msg("Operation timed out while getting token balance")
		default:
			log.Fatal().Err(err).Msg("Failed to get token balance")
		}
	}

	log.Info().
		Str("wallet_address", walletAdapter.GetAddress().Hex()).
		Str("balance", tokenBalance.String()+" PRTY").
		Msg("Wallet token balance")

	deviceIDManager := deviceid.NewManager(deviceid.Config{})
	deviceID, err := deviceIDManager.VerifyDeviceID()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to get device ID")
	}

	// Create stake wallet contract
	stakeWallet, err := walletAdapter.NewStakeWallet(
		common.HexToAddress(cfg.Ethereum.StakeWalletAddress),
		common.HexToAddress(cfg.Ethereum.TokenAddress),
	)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create stake wallet contract")
	}

	// Get stake info
	stakeInfo, err := walletAdapter.GetStakeInfo(ctx, stakeWallet, deviceID)
	if err != nil {
		select {
		case <-ctx.Done():
			log.Fatal().Err(ctx.Err()).Msg("Operation timed out while getting stake info")
		default:
			log.Fatal().Err(err).Msg("Failed to get stake info")
		}
	}

	if stakeInfo.Exists {
		log.Info().
			Str("amount", stakeInfo.Amount.String()+" PRTY").
			Str("device_id", stakeInfo.DeviceID).
			Str("wallet_address", stakeInfo.WalletAddress.Hex()).
			Msg("Current stake info")

		// Get contract token balance
		contractBalance, err := walletAdapter.GetTokenBalance(ctx, token, common.HexToAddress(cfg.Ethereum.StakeWalletAddress))
		if err != nil {
			select {
			case <-ctx.Done():
				log.Fatal().Err(ctx.Err()).Msg("Operation timed out while getting contract balance")
			default:
				log.Fatal().Err(err).Msg("Failed to get contract balance")
			}
		}
		log.Info().
			Str("balance", contractBalance.String()+" PRTY").
			Str("contract_address", cfg.Ethereum.StakeWalletAddress).
			Msg("Contract token balance")
	} else {
		log.Info().Msg("No active stake found")
	}
}
