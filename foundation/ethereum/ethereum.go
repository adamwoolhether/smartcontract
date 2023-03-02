// Package ethereum provides access to core geth functions for convenience.
// This was taken from ardanlabs/ethereum.
package ethereum

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

// Set of networks supported by the package.
const (
	NetworkHTTPLocalhost = "http://localhost:8545"
	NetworkLocalhost     = "zarf/ethereum/geth.ipc"
)

type Backend interface {
	bind.ContractBackend
	bind.DeployBackend
	TransactionByHash(ctx context.Context, txHash common.Hash) (*types.Transaction, bool, error)
	TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error)
	BalanceAt(ctx context.Context, contract common.Address, blockNumber *big.Int) (*big.Int, error)
	Network() string
	ChainID() *big.Int
}

// Client enables API interaction with smart contracts.
type Client struct {
	Backend
	address    common.Address
	privateKey *ecdsa.PrivateKey
}

// NewClient provides an API for accessing an ethereum node for performing
// blockchain operations. The private key will determine "who" will be
// interacting with the node vai the returned *Client.
func NewClient(backend Backend, privateKey *ecdsa.PrivateKey) (*Client, error) {
	client := Client{
		Backend:    backend,
		address:    crypto.PubkeyToAddress(privateKey.PublicKey),
		privateKey: privateKey,
	}

	return &client, nil
}

// Address returns the current address calculated from the private key.
func (c *Client) Address() common.Address {
	return c.address
}

// Network returns network info for the connected network.
func (c *Client) Network() string {
	return c.Backend.Network()
}

// ChainID returns network info for the connected network.
func (c *Client) ChainID() int {
	return int(c.Backend.ChainID().Int64())
}

// PrivateKey returns network info for the connected network.
func (c *Client) PrivateKey() *ecdsa.PrivateKey {
	return c.privateKey
}

// Balance retrieves the current balances for the client's account.
func (c *Client) Balance(ctx context.Context) (wei *big.Int, err error) {
	return c.BalanceAt(ctx, c.address, nil)
}

// /////////////////////////////////////////////////////////////////

// NewTransactOpts constructs a new TransactOpts - a collection of authorization data
// required for creating a valid Ethereum transaction. If gasLimit is set to 0, then
// the amount of gas needed will be estimated. If gasPrice is set to 0, then the
// connected geth service is consulted for the suggested gas price.
func (c *Client) NewTransactOpts(ctx context.Context, gasLimit uint64, gasPrice *big.Int, valueGWei *big.Float) (*bind.TransactOpts, error) {
	nonce, err := c.PendingNonceAt(ctx, c.address)
	if err != nil {
		return nil, err
	}

	if gasPrice == nil || gasPrice.Cmp(big.NewInt(0)) == 0 {
		gasPrice, err = c.SuggestGasPrice(ctx)
		if err != nil {
			return nil, fmt.Errorf("retrieving suggested gas price: %w", err)
		}
	}

	txOpts, err := bind.NewKeyedTransactorWithChainID(c.privateKey, c.Backend.ChainID())
	if err != nil {
		return nil, fmt.Errorf("keying transaction: %w", err)
	}

	// Convert the GWei value to Wei.
	gWei2Wei := big.NewInt(0)
	big.NewFloat(0).SetPrec(1024).Mul(valueGWei, big.NewFloat(1e9)).Int(gWei2Wei)

	txOpts.Nonce = big.NewInt(0).SetUint64(nonce)
	txOpts.Value = gWei2Wei
	txOpts.GasLimit = gasLimit // Maximum amount of gas we're willing to pay for.
	txOpts.GasPrice = gasPrice // Amount agree on to pay per unit of gas.

	return txOpts, nil
}

// WaitMined waits for the transaction to be mined before returning a receipt.
func (c *Client) WaitMined(ctx context.Context, tx *types.Transaction) (*types.Receipt, error) {
	receipt, err := bind.WaitMined(ctx, c.Backend, tx)
	if err != nil {
		return nil, fmt.Errorf("waiting for tx to be mined: %w", err)
	}

	if receipt.Status == 0 {
		if err := c.extractError(ctx, tx); err != nil {
			return nil, fmt.Errorf("extracting tx error: %w", err)
		}
	}

	return receipt, nil
}

// extractError checks retrieves the error message from a failed transaction.
func (c *Client) extractError(ctx context.Context, tx *types.Transaction) error {
	msg := ethereum.CallMsg{
		From:     c.address,
		To:       tx.To(),
		Gas:      tx.Gas(),
		GasPrice: tx.GasPrice(),
		Value:    tx.Value(),
		Data:     tx.Data(),
	}

	_, err := c.CallContract(ctx, msg, nil)
	return err
}
