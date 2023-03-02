package ethereum

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/ethclient"
)

// DialedBackend represents a dialed connect to an Ethereum node.
type DialedBackend struct {
	*ethclient.Client
	network string
	chainID *big.Int
}

// CreateDialedBackend constructs and ethereum client value
// for the given network and establishes a connection.
func CreateDialedBackend(ctx context.Context, network string) (*DialedBackend, error) {
	client, err := ethclient.Dial(network)
	if err != nil {
		return nil, err
	}

	chaindID, err := client.ChainID(ctx)
	if err != nil {
		return nil, err
	}

	db := DialedBackend{
		Client:  client,
		network: network,
		chainID: chaindID,
	}

	return &db, nil
}

// Network returns the network that the backend is connected to.
func (db *DialedBackend) Network() string {
	return db.network
}

// ChainID returns the chain id that the backend is connected to.
func (db *DialedBackend) ChainID() *big.Int {
	return db.chainID
}

// /////////////////////////////////////////////////////////////////
