package bank_test

import (
	"context"
	"math/big"
	"testing"

	"github.com/adamwoolhether/smartcontract/app/bank/single/contract/go/bank"
	"github.com/adamwoolhether/smartcontract/foundation/ethereum"
)

const (
	deployerAcct = iota
	winnerAcc
	loser1Acc
	loser2Acc
	numAccounts
)

func TestBankSingle(t *testing.T) {
	ctx := context.Background()

	backend, err := ethereum.CreateSimulatedBackend(numAccounts, true, big.NewInt(100))
	if err != nil {
		t.Fatalf("unable to create simulated backend: %s", err)
	}
	defer backend.Close()

	// /////////////////////////////////////////////////////////////

	deployer, err := ethereum.NewClient(backend, backend.PrivateKeys[deployerAcct])
	if err != nil {
		t.Fatalf("unable to create deplayerAcct: %s", err)
	}

	callOpts, err := deployer.NewCallOpts(ctx)
	if err != nil {
		t.Fatalf("unable to create call opts: %s", err)
	}

	// /////////////////////////////////////////////////////////////

	const gasLimit = 1_700_000
	const valueGwei = 0.0

	var testBank *bank.Bank

	// /////////////////////////////////////////////////////////////

	t.Run("deploy bank", func(t *testing.T) {
		deployTxOpts, err := deployer.NewTransactOpts(ctx, gasLimit, big.NewInt(0), big.NewFloat(valueGwei))
		if err != nil {
			t.Fatalf("unable to create transaction opts for deploy: %s", err)
		}

		contractID, tx, _, err := bank.DeployBank(deployTxOpts, deployer.Backend)
		if err != nil {
			t.Fatalf("unable to deploy bank: %s", err)
		}

		if _, err := deployer.WaitMined(ctx, tx); err != nil {
			t.Fatalf("waiting for deploy: %s", err)
		}

		testBank, err = bank.NewBank(contractID, deployer.Backend)
		if err != nil {
			t.Fatalf("unable to create bank: %s", err)
		}

		// /////////////////////////////////////////////////////////////

		t.Run("check owner matches", func(t *testing.T) {
			owner, err := testBank.Owner(callOpts)
			if err != nil {
				t.Fatalf("unable to get account owner: %s", err)
			}

			if owner != deployer.Address() {
				t.Fatalf("retrieved owner doesn't match expectation: %v != %v", owner, deployer.Address())
			}
		})

		// /////////////////////////////////////////////////////////////

		t.Run("check deposit", func(t *testing.T) {
			initialBalance, err := testBank.Balance(callOpts)
			if err != nil {
				t.Fatalf("should be able to get the initial balance: %s", err)
			}

			depositTxOpts, err := deployer.NewTransactOpts(ctx, gasLimit, big.NewInt(0), big.NewFloat(valueGwei))
			if err != nil {
				t.Fatalf("unable to create transaction opts for deposit: %s", err)
			}

			depositTxOpts.Value = big.NewInt(10)
			tx, err := testBank.Deposit(depositTxOpts)
			if err != nil {
				t.Fatalf("should be able to deposit money: %s", err)
			}

			if _, err := deployer.WaitMined(ctx, tx); err != nil {
				t.Fatalf("waitinf for deposit: %s", err)
			}

			postDepositBalance, err := testBank.Balance(callOpts)
			if err != nil {
				t.Fatalf("unable to get balance after deposit: %s", err)
			}

			expectedBalance := initialBalance.Add(initialBalance, depositTxOpts.Value)
			if postDepositBalance.Cmp(expectedBalance) != 0 {
				t.Fatalf("wrong balance, got %v, exp %v", postDepositBalance, expectedBalance)
			}
		})

		// /////////////////////////////////////////////////////////////

		t.Run("check withdraw", func(t *testing.T) {
			initialBalance, err := testBank.Balance(callOpts)
			if err != nil {
				t.Fatalf("should be able to get the initial balacne: %s", err)
			}

			withdrawTxOpts, err := deployer.NewTransactOpts(ctx, gasLimit, big.NewInt(0), big.NewFloat(valueGwei))
			if err != nil {
				t.Fatalf("unable to create transaction opts for withdraw: %s", err)
			}

			withdrawTxOpts.Value = big.NewInt(10)
			tx, err := testBank.Withdraw(withdrawTxOpts)
			if err != nil {
				t.Fatalf("unable to withdraw money: %s", err)
			}

			if _, err := deployer.WaitMined(ctx, tx); err != nil {
				t.Fatalf("waiting for withdraw: %s", err)
			}

			postWithdrawBalance, err := testBank.Balance(callOpts)
			if err != nil {
				t.Fatalf("should be able to get balance after withdraw: %s", err)
			}

			expectedBalance := initialBalance.Sub(initialBalance, withdrawTxOpts.Value)
			if postWithdrawBalance.Cmp(expectedBalance) != 0 {
				t.Fatalf("wrong balance, got %v  exp %v", postWithdrawBalance, expectedBalance)
			}
		})

		// /////////////////////////////////////////////////////////////

		t.Run("check version", func(t *testing.T) {
			version, err := testBank.Version(callOpts)
			if err != nil {
				t.Fatalf("error getting version: %s", err)
			}

			const expectedVersion = "0.1.0"
			if version != expectedVersion {
				t.Fatalf("wrong version. got %s, exp %s\n", version, expectedVersion)
			}
		})
	})
}
