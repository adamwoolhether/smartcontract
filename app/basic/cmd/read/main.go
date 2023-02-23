package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/ardanlabs/ethereum"
	"github.com/ethereum/go-ethereum/common"

	"github.com/adamwoolhether/smartcontract/app/basic/contract/go/basic"
)

const (
	keyStoreFile = "zarf/ethereum/keystore/UTC--2022-05-12T14-47-50.112225000Z--6327a38415c53ffb36c11db55ea74cc9cb4976fd"
	passPhrase   = "123"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	ctx := context.Background()

	backend, err := ethereum.CreateDialedBackend(ctx, ethereum.NetworkHTTPLocalhost)
	if err != nil {
		return err
	}
	defer backend.Close()

	privateKey, err := ethereum.PrivateKeyByKeyFile(keyStoreFile, passPhrase)
	if err != nil {
		return err
	}

	// NOTE that this client is for the account with the associated
	// private key. Multiple clients may be needed for conducting
	// transactions on from other accounts.
	client, err := ethereum.NewClient(backend, privateKey)
	if err != nil {
		return err
	}

	fmt.Println("\nInput Values")
	fmt.Println("------------------------------------------------")
	fmt.Println("fromAddress:", client.Address())

	// /////////////////////////////////////////////////////////////

	contractIDBytes, err := os.ReadFile("zarf/ethereum/basic.cid")
	if err != nil {
		return fmt.Errorf("importing basic.cid file: %w", err)
	}

	contractID := string(contractIDBytes)
	if contractID == "" {
		return errors.New("need to export the basic.cid file")
	}
	fmt.Println("contractID:", contractID)

	// Retrieve a value that contains our contract API.
	contract, err := basic.NewBasic(common.HexToAddress(contractID), client.Backend)
	if err != nil {
		return fmt.Errorf("new contract: %w", err)
	}

	version, err := contract.Version(nil) // We can pass nil CallOpts. there is no cost.
	if err != nil {
		return err
	}
	fmt.Println("version:", version)

	// /////////////////////////////////////////////////////////////

	key := "adam"
	result, err := contract.Items(nil, key)
	if err != nil {
		log.Fatal("SetItem ERROR:", err)
	}

	fmt.Println("\nRead Value")
	fmt.Println("------------------------------------------------")
	fmt.Println("value:", result)

	return nil
}
