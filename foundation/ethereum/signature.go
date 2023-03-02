package ethereum

import (
	"crypto/ecdsa"
	"fmt"
	"os"

	"github.com/ethereum/go-ethereum/accounts/keystore"
)

// ZeroHash represents a 0 value hashcode.
const ZeroHash = "0x0000000000000000000000000000000000000000000000000000000000000000"

// ethID is an arbitrary number for signing messages. It makes clear
// that the signature comes from the ethereum blockchain.
const ethID = 27

// /////////////////////////////////////////////////////////////////

func PrivateKeyByKeyFile(keyFile, passphrase string) (*ecdsa.PrivateKey, error) {
	data, err := os.ReadFile(keyFile)
	if err != nil {
		return nil, fmt.Errorf("read key file: %w", err)
	}

	key, err := keystore.DecryptKey(data, passphrase)
	if err != nil {
		return nil, err
	}

	return key.PrivateKey, nil
}
