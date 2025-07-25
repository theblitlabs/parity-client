package cli

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/spf13/cobra"
	"github.com/theblitlabs/deviceid"
	walletsdk "github.com/theblitlabs/go-wallet-sdk"
	"github.com/theblitlabs/gologger"
	"github.com/theblitlabs/parity-client/internal/adapters/keystore"
	"github.com/theblitlabs/parity-client/internal/adapters/wallet"
	"github.com/theblitlabs/parity-client/internal/config"
	"github.com/theblitlabs/parity-client/internal/utils"
)

func RunStake(cmd *cobra.Command, args []string) {
	log := gologger.Get().With().Str("component", "stake").Logger()

	amount, _ := cmd.Flags().GetFloat64("amount")
	configPath, _ := cmd.Flags().GetString("config-path")

	if err := utils.ValidateAmount(amount, 0.000001); err != nil {
		log.Error().Err(err).Msg("Invalid stake amount")
		return
	}

	log.Info().
		Float64("amount", amount).
		Msg("Processing stake request")

	if err := executeStake(amount, configPath); err != nil {
		log.Error().Err(err).Msg("Stake operation failed")
	}
}

func executeStake(amount float64, configPath string) error {
	log := gologger.Get().With().Str("component", "stake").Logger()

	configManager := config.NewConfigManager(configPath)
	cfg, err := configManager.GetConfig()
	if err != nil {
		log.Error().
			Err(err).
			Msg("Failed to load configuration - please ensure config file exists")
		return err
	}

	keystoreAdapter, err := keystore.NewAdapter(nil)
	if err != nil {
		log.Error().
			Err(err).
			Msg("Failed to create keystore")
		return err
	}

	privateKey, err := keystoreAdapter.LoadPrivateKey()
	if err != nil {
		log.Error().
			Err(err).
			Msg("No private key found - please authenticate first using 'parity auth'")
		return err
	}

	walletAdapter, err := wallet.NewAdapter(walletsdk.ClientConfig{
		RPCURL:       cfg.BlockchainNetwork.RPC,
		ChainID:      cfg.BlockchainNetwork.ChainID,
		PrivateKey:   common.Bytes2Hex(crypto.FromECDSA(privateKey)),
		TokenAddress: common.HexToAddress(cfg.BlockchainNetwork.TokenAddress),
		StakeAddress: common.HexToAddress(cfg.BlockchainNetwork.StakeWalletAddress),
	})
	if err != nil {
		log.Error().
			Err(err).
			Str("rpc_endpoint", cfg.BlockchainNetwork.RPC).
			Int64("chain_id", cfg.BlockchainNetwork.ChainID).
			Msg("Failed to connect to blockchain - please check your network connection")
		return err
	}

	deviceIDManager := deviceid.NewManager(deviceid.Config{})
	deviceID, err := deviceIDManager.VerifyDeviceID()
	if err != nil {
		log.Error().
			Err(err).
			Msg("Failed to verify device - please ensure you have a valid device ID")
		return err
	}

	log.Info().
		Str("device_id", deviceID).
		Str("wallet", walletAdapter.GetAddress().Hex()).
		Msg("Device verified successfully")

	tokenAddr := common.HexToAddress(cfg.BlockchainNetwork.TokenAddress)
	stakeWalletAddr := common.HexToAddress(cfg.BlockchainNetwork.StakeWalletAddress)

	token, err := walletAdapter.NewParityToken(tokenAddr)
	if err != nil {
		log.Error().
			Err(err).
			Str("token_address", tokenAddr.Hex()).
			Str("wallet", walletAdapter.GetAddress().Hex()).
			Msg("Failed to create token contract - please try again")
		return err
	}

	balance, err := token.BalanceOf(nil, walletAdapter.GetAddress())
	if err != nil {
		log.Error().
			Err(err).
			Str("token_address", tokenAddr.Hex()).
			Str("wallet", walletAdapter.GetAddress().Hex()).
			Msg("Failed to check token balance - please try again")
		return err
	}

	amountToStake := amountWei(amount)
	if balance.Cmp(amountToStake) < 0 {
		log.Error().
			Str("current_balance", utils.FormatEther(balance)+" "+cfg.BlockchainNetwork.TokenSymbol).
			Str("required_amount", utils.FormatEther(amountToStake)+" "+cfg.BlockchainNetwork.TokenSymbol).
			Msg("Insufficient token balance - please ensure you have enough " + cfg.BlockchainNetwork.TokenSymbol + " tokens")
		return fmt.Errorf("insufficient token balance")
	}

	log.Info().
		Str("balance", utils.FormatEther(balance)+" "+cfg.BlockchainNetwork.TokenSymbol).
		Str("token_address", tokenAddr.Hex()).
		Msg("Current token balance verified")

	allowance, err := token.Allowance(nil, walletAdapter.GetAddress(), stakeWalletAddr)
	if err != nil {
		log.Error().
			Err(err).
			Str("token_address", tokenAddr.Hex()).
			Str("stake_wallet", stakeWalletAddr.Hex()).
			Msg("Failed to check token allowance - please try again")
		return err
	}

	if allowance.Cmp(amountToStake) < 0 {
		log.Info().
			Str("amount", utils.FormatEther(amountToStake)+" "+cfg.BlockchainNetwork.TokenSymbol).
			Msg("Approving token spending...")

		txOpts, err := walletAdapter.GetTransactOpts()
		if err != nil {
			log.Error().
				Err(err).
				Msg("Failed to prepare transaction - please try again")
			return err
		}

		tx, err := token.Approve(txOpts, stakeWalletAddr, amountToStake)
		if err != nil {
			log.Error().
				Err(err).
				Str("amount", utils.FormatEther(amountToStake)+" "+cfg.BlockchainNetwork.TokenSymbol).
				Msg("Failed to approve token spending - please try again")
			return err
		}

		log.Info().
			Str("tx_hash", tx.Hash().Hex()).
			Str("amount", utils.FormatEther(amountToStake)+" "+cfg.BlockchainNetwork.TokenSymbol).
			Msg("Token approval submitted - waiting for confirmation...")

		ctx := context.Background()
		receipt, err := bind.WaitMined(ctx, walletAdapter.GetClient(), tx)
		if err != nil {
			log.Error().
				Err(err).
				Str("tx_hash", tx.Hash().Hex()).
				Msg("Failed to confirm token approval - please check the transaction status")
			return err
		}

		if receipt.Status == 0 {
			log.Error().
				Str("tx_hash", tx.Hash().Hex()).
				Msg("Token approval failed - please check the transaction status")
			return fmt.Errorf("token approval transaction failed")
		}

		log.Info().
			Str("tx_hash", tx.Hash().Hex()).
			Msg("Token approval confirmed successfully")

		time.Sleep(5 * time.Second)
	}

	log.Info().
		Str("amount", utils.FormatEther(amountToStake)+" "+cfg.BlockchainNetwork.TokenSymbol).
		Str("device_id", deviceID).
		Msg("Submitting stake transaction...")

	stakeWallet, err := walletAdapter.NewStakeWallet(stakeWalletAddr, tokenAddr)
	if err != nil {
		log.Error().
			Err(err).
			Str("stake_wallet", stakeWalletAddr.Hex()).
			Msg("Failed to create stake wallet contract")
		return err
	}

	tx, err := walletAdapter.Stake(context.Background(), stakeWallet, amountToStake, deviceID)
	if err != nil {
		log.Error().
			Err(err).
			Str("amount", utils.FormatEther(amountToStake)+" "+cfg.BlockchainNetwork.TokenSymbol).
			Str("device_id", deviceID).
			Msg("Failed to submit stake transaction - please try again")
		return err
	}

	log.Info().
		Str("tx_hash", tx.Hash().Hex()).
		Str("amount", utils.FormatEther(amountToStake)+" "+cfg.BlockchainNetwork.TokenSymbol).
		Str("device_id", deviceID).
		Msg("Stake transaction submitted successfully")

	return nil
}

func amountWei(amount float64) *big.Int {
	floatStr := fmt.Sprintf("%.18f", amount)
	wei := new(big.Int)
	wei.SetString(strings.Replace(floatStr, ".", "", 1), 10)
	return wei
}
