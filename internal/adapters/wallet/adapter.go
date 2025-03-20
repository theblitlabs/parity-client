package wallet

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	walletsdk "github.com/theblitlabs/go-wallet-sdk"
)

// Adapter wraps the go-wallet-sdk package to provide a clean interface
type Adapter struct {
	client *walletsdk.Client
}

// NewAdapter creates a new wallet adapter
func NewAdapter(config walletsdk.ClientConfig) (*Adapter, error) {
	client, err := walletsdk.NewClient(config)
	if err != nil {
		return nil, err
	}

	return &Adapter{
		client: client,
	}, nil
}

// GetClient returns the underlying wallet SDK client
// This should be used sparingly and only when direct access to the client is necessary
func (a *Adapter) GetClient() *walletsdk.Client {
	return a.client
}

// GetAddress returns the wallet's address
func (a *Adapter) GetAddress() common.Address {
	return a.client.Address()
}

// GetBalance retrieves the balance for a given address
func (a *Adapter) GetBalance(ctx context.Context, address common.Address) (*big.Int, error) {
	return a.client.GetBalance(address)
}

// Transfer sends ETH to the specified address
func (a *Adapter) Transfer(to common.Address, amount *big.Int) (*types.Transaction, error) {
	return a.client.Transfer(to, amount)
}

// Stake tokens in the contract
func (a *Adapter) Stake(ctx context.Context, stakeWallet *walletsdk.StakeWallet, amount *big.Int, deviceID string) (*types.Transaction, error) {
	return stakeWallet.Stake(amount, deviceID)
}

// WithdrawStake withdraws tokens from the contract
func (a *Adapter) WithdrawStake(ctx context.Context, stakeWallet *walletsdk.StakeWallet, deviceID string, amount *big.Int) (*types.Transaction, error) {
	return stakeWallet.WithdrawStake(deviceID, amount)
}

// GetStakeInfo retrieves staking information for a device
func (a *Adapter) GetStakeInfo(ctx context.Context, stakeWallet *walletsdk.StakeWallet, deviceID string) (walletsdk.StakeInfo, error) {
	return stakeWallet.GetStakeInfo(deviceID)
}

// GetTokenBalance retrieves the token balance for a given token contract
func (a *Adapter) GetTokenBalance(ctx context.Context, token *walletsdk.ParityToken, address common.Address) (*big.Int, error) {
	opts := &bind.CallOpts{Context: ctx}
	return token.BalanceOf(opts, address)
}

// TransferToken sends tokens to the specified address
func (a *Adapter) TransferToken(ctx context.Context, token *walletsdk.ParityToken, to common.Address, amount *big.Int) (*types.Transaction, error) {
	opts, err := a.client.GetTransactOpts()
	if err != nil {
		return nil, err
	}
	return token.Transfer(opts, to, amount)
}

// GetTransactOpts gets transaction options for contract interactions
func (a *Adapter) GetTransactOpts() (*bind.TransactOpts, error) {
	return a.client.GetTransactOpts()
}

// NewParityToken creates a new instance of the Parity token contract
func (a *Adapter) NewParityToken(tokenAddress common.Address) (*walletsdk.ParityToken, error) {
	return walletsdk.NewParityToken(tokenAddress, a.client)
}

// NewStakeWallet creates a new instance of the stake wallet contract
func (a *Adapter) NewStakeWallet(stakeWalletAddr, tokenAddr common.Address) (*walletsdk.StakeWallet, error) {
	return walletsdk.NewStakeWallet(a.client, stakeWalletAddr, tokenAddr)
}
