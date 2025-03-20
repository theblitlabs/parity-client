package wallet

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	walletsdk "github.com/theblitlabs/go-wallet-sdk"
)

type Adapter struct {
	client *walletsdk.Client
}

func NewAdapter(config walletsdk.ClientConfig) (*Adapter, error) {
	client, err := walletsdk.NewClient(config)
	if err != nil {
		return nil, err
	}

	return &Adapter{
		client: client,
	}, nil
}

func (a *Adapter) GetClient() *walletsdk.Client {
	return a.client
}

func (a *Adapter) GetAddress() common.Address {
	return a.client.Address()
}

func (a *Adapter) GetBalance(ctx context.Context, address common.Address) (*big.Int, error) {
	return a.client.GetBalance(address)
}

func (a *Adapter) Transfer(to common.Address, amount *big.Int) (*types.Transaction, error) {
	return a.client.Transfer(to, amount)
}

func (a *Adapter) Stake(ctx context.Context, stakeWallet *walletsdk.StakeWallet, amount *big.Int, deviceID string) (*types.Transaction, error) {
	return stakeWallet.Stake(amount, deviceID)
}

func (a *Adapter) WithdrawStake(ctx context.Context, stakeWallet *walletsdk.StakeWallet, deviceID string, amount *big.Int) (*types.Transaction, error) {
	return stakeWallet.WithdrawStake(deviceID, amount)
}

func (a *Adapter) GetStakeInfo(ctx context.Context, stakeWallet *walletsdk.StakeWallet, deviceID string) (walletsdk.StakeInfo, error) {
	return stakeWallet.GetStakeInfo(deviceID)
}

func (a *Adapter) GetTokenBalance(ctx context.Context, token *walletsdk.ParityToken, address common.Address) (*big.Int, error) {
	opts := &bind.CallOpts{Context: ctx}
	return token.BalanceOf(opts, address)
}

func (a *Adapter) TransferToken(ctx context.Context, token *walletsdk.ParityToken, to common.Address, amount *big.Int) (*types.Transaction, error) {
	opts, err := a.client.GetTransactOpts()
	if err != nil {
		return nil, err
	}
	return token.Transfer(opts, to, amount)
}

func (a *Adapter) GetTransactOpts() (*bind.TransactOpts, error) {
	return a.client.GetTransactOpts()
}

func (a *Adapter) NewParityToken(tokenAddress common.Address) (*walletsdk.ParityToken, error) {
	return walletsdk.NewParityToken(tokenAddress, a.client)
}

func (a *Adapter) NewStakeWallet(stakeWalletAddr, tokenAddr common.Address) (*walletsdk.StakeWallet, error) {
	return walletsdk.NewStakeWallet(a.client, stakeWalletAddr, tokenAddr)
}
