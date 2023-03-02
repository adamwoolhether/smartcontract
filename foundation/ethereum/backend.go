package ethereum

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
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

// SimulatedBackend represents a simulated connection to an ethereum node.
type SimulatedBackend struct {
	*backends.SimulatedBackend
	AutoCommit  bool
	PrivateKeys []*ecdsa.PrivateKey
	network     string
	chainID     *big.Int
}

// CreateSimulatedBackend constructs a simulated backend and set of private keys
// registered to the backend with a balance on 100 ETH. These private keys are
// used with the NewSimulation call to get an Ethereum API value.
func CreateSimulatedBackend(numAccounts int, autoCommit bool, accountBalance *big.Int) (*SimulatedBackend, error) {
	keys := make([]*ecdsa.PrivateKey, numAccounts)
	alloc := make(core.GenesisAlloc)

	for i := 0; i < numAccounts; i++ {
		privateKey, err := ecdsa.GenerateKey(crypto.S256(), rand.Reader)
		if err != nil {
			return nil, fmt.Errorf("unable to generate private key: %w", err)
		}

		keys[i] = privateKey

		alloc[crypto.PubkeyToAddress(privateKey.PublicKey)] = core.GenesisAccount{
			Balance: big.NewInt(0).Mul(accountBalance, big.NewInt(1e18)),
		}
	}

	maxLimit := uint64(9223372036854775807)
	client := backends.NewSimulatedBackend(alloc, maxLimit)

	// Set the clock to 5.15 minutes into the past to deal with bugs
	// the simulated clock. Transactions won't be committed otherwise.
	now := time.Since(time.Date(1970, time.January, 1, 0, 5, 15, 0, time.UTC))
	client.AdjustTime(now)

	client.Commit()

	b := SimulatedBackend{
		SimulatedBackend: client,
		AutoCommit:       autoCommit,
		PrivateKeys:      keys,
		network:          "simulated",
		chainID:          big.NewInt(1337),
	}

	return &b, nil
}

// Network returns the network that the backend is connected to.
func (sb *SimulatedBackend) Network() string {
	return sb.network
}

// ChainID returns the chain id that the backend is connected to.
func (sb *SimulatedBackend) ChainID() *big.Int {
	return sb.chainID
}

// SendTransaction pipes parameters to the embedded backend and
// also calls Commit() if sb.AutoCommit==true.
func (sb *SimulatedBackend) SendTransaction(ctx context.Context, tx *types.Transaction) error {
	if err := sb.SimulatedBackend.SendTransaction(ctx, tx); err != nil {
		return err
	}

	if sb.AutoCommit {
		sb.Commit()
	}

	return nil
}

// SetTime shifts the time of the simulated clock.
// It can only be called on empty blocks.
func (sb *SimulatedBackend) SetTime(t time.Time) {
	sb.AdjustTime(time.Since(t))
	sb.Commit()
}