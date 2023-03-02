package main

import (
	"context"
	"fmt"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/log"

	"github.com/adamwoolhether/smartcontract/app/basic/contract/go/basic"
	"github.com/adamwoolhether/smartcontract/foundation/ethereum"
	"github.com/adamwoolhether/smartcontract/foundation/ethereum/currency"
)

const (
	keyStoreFile = "zarf/ethereum/keystore/UTC--2022-05-12T14-47-50.112225000Z--6327a38415c53ffb36c11db55ea74cc9cb4976fd"
	passPhrase   = "123"
)

var coinMarketCapKey = os.Getenv("CMC_API_KEY")

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() (err error) {
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

	converter, err := currency.NewConverter(basic.BasicMetaData.ABI, coinMarketCapKey)
	if err != nil {
		return err
	}
	oneETHtoUSD, oneUSDtoETH := converter.Values()

	fmt.Println("oneETHtoUSD:", oneETHtoUSD)
	fmt.Println("oneUSDtoETH:", oneUSDtoETH)

	// /////////////////////////////////////////////////////////////

	startingBalance, err := client.Balance(ctx)
	if err != nil {
		return err
	}
	defer func() {
		endingBalance, dErr := client.Balance(ctx)
		if dErr != nil {
			err = dErr
			return
		}
		fmt.Println(converter.FmtBalanceSheet(startingBalance, endingBalance))
	}()

	// /////////////////////////////////////////////////////////////

	const gasLimit = 1_600_000
	const gasPriceWei = 39.576 // Current price as of feb 18
	const valueGwei = 0.0
	txOpts, err := client.NewTransactOpts(ctx, gasLimit, currency.GWei2Wei(big.NewFloat(gasPriceWei)), big.NewFloat(valueGwei))
	if err != nil {
		return err
	}

	// /////////////////////////////////////////////////////////////

	address, tx, _, err := basic.DeployBasic(txOpts, client.Backend)
	if err != nil {
		return err
	}
	fmt.Println(converter.FmtTransaction(tx))

	fmt.Println("\nContract Details")
	fmt.Println("------------------------------------------------")
	fmt.Println("contract id      :", address.Hex())

	// Save the contract ID! We need this to make API calls.
	if err := os.WriteFile("zarf/ethereum/basic.cid", []byte(address.Hex()), 0644); err != nil {
		return fmt.Errorf("exporting basic.cid file: %w", err)
	}

	// /////////////////////////////////////////////////////////////

	fmt.Println("\nWaiting Logs")
	fmt.Println("------------------------------------------------")
	log.Root().SetHandler(log.StdoutHandler)

	// Wait for the
	receipt, err := client.WaitMined(ctx, tx)
	if err != nil {
		return err
	}
	fmt.Println(converter.FmtTransactionReceipt(receipt, tx.GasPrice()))

	return nil
}
