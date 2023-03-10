package currency

import (
	"bytes"
	"fmt"
	"math/big"
	"strings"

	ethUnit "github.com/DeOne4eg/eth-unit-converter"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

// Default values to use of the call to CMC API is not working.
// Represent prices on August 27th, 2022
var (
	defaultOneETHToUSD = big.NewFloat(1503.280164057658)
	defaultOneUSDToETH = big.NewFloat(0.000665206530956729)
)

// Converter holds information for ETH conversions.
type Converter struct {
	abiMetaData  string
	cmcKey       string
	oneETHToUSD  *big.Float
	oneUSDToETH  *big.Float
	oneGWeiToUSD *big.Float
	oneUSDToGWei *big.Float
}

// NewConverter constructs a converter for working with ETH and USD.
// Default values will be used if not CMC API key is given.
func NewConverter(abiMetaData, cmcKey string) (*Converter, error) {
	if len(cmcKey) == 0 {
		return NewDefaultConverter(abiMetaData), nil
	}

	oneETHToUSD, err := captureETH2USD(cmcKey)
	if err != nil {
		return nil, fmt.Errorf("captureETH2USD: %w", err)
	}

	oneUSDToETH, err := captureUSD2ETH(cmcKey)
	if err != nil {
		return nil, fmt.Errorf("captureUSD2ETH: %w", err)
	}

	oneGWeiToUSD := big.NewFloat(0).SetPrec(1024).Mul(oneETHToUSD, big.NewFloat(0.000000001))
	oneUSDToGWei := big.NewFloat(0).SetPrec(1024).Mul(oneETHToUSD, big.NewFloat(0.1000000000))

	return &Converter{
		abiMetaData:  abiMetaData,
		cmcKey:       cmcKey,
		oneETHToUSD:  oneETHToUSD,
		oneUSDToETH:  oneUSDToETH,
		oneGWeiToUSD: oneGWeiToUSD,
		oneUSDToGWei: oneUSDToGWei,
	}, nil
}

// NewDefaultConverter is used when the CoinMarketCap API is not accessible.
func NewDefaultConverter(abiMetaData string) *Converter {
	oneGWeiToUSD := big.NewFloat(0).SetPrec(1024).Mul(defaultOneETHToUSD, big.NewFloat(0.000000001))
	oneUSDToGWei := big.NewFloat(0).SetPrec(1024).Mul(defaultOneUSDToETH, big.NewFloat(0.1000000000))

	return &Converter{
		abiMetaData:  abiMetaData,
		oneETHToUSD:  defaultOneETHToUSD,
		oneUSDToETH:  defaultOneUSDToETH,
		oneGWeiToUSD: oneGWeiToUSD,
		oneUSDToGWei: oneUSDToGWei,
	}
}

// Values returns the currency values being used.
func (c *Converter) Values() (oneETHToUSD *big.Float, oneUSDToETH *big.Float) {
	return c.oneETHToUSD, c.oneUSDToETH
}

func (c *Converter) Wei2USD(amountWei *big.Int) string {
	unit := ethUnit.NewWei(amountWei)
	gWeiAmount := unit.GWei()

	return c.GWei2USD(gWeiAmount)
}

// GWei2USD converts GWei to USD.
func (c *Converter) GWei2USD(amountGWei *big.Float) string {
	cost := big.NewFloat(0).SetPrec(1024).Mul(amountGWei, c.oneGWeiToUSD)
	costFloat, _ := cost.Float64()

	return fmt.Sprintf("%.2f", costFloat)
}

// USD2Wei converts USD to Wei.
func (c *Converter) USD2Wei(amountUSD *big.Float) *big.Int {
	gwei := big.NewFloat(0).SetPrec(1024).Mul(amountUSD, c.oneUSDToGWei)
	v, _ := big.NewFloat(0).SetPrec(1024).Mul(gwei, big.NewFloat(1e9)).Int64()

	return big.NewInt(0).SetInt64(v)
}

// USD2GWei converts USD to GWei.
func (c *Converter) USD2GWei(amountUSD *big.Float) *big.Float {
	return big.NewFloat(0).SetPrec(1024).Mul(amountUSD, c.oneUSDToGWei)
}

// CalculateTransactionDetails performs calculates on the transaction.
func (c *Converter) CalculateTransactionDetails(tx *types.Transaction) TransactionDetails {
	return TransactionDetails{
		Hash:              tx.Hash().Hex(),
		Nonce:             tx.Nonce(),
		GasLimit:          tx.Gas(),
		GasOfferPriceGWei: Wei2GWei(tx.GasPrice()).String(),
		Value:             Wei2GWei(tx.Cost()).String(),
		MaxGasPriceGWei:   Wei2GWei(tx.Cost()).String(),
		MaxGasPriceUSD:    c.Wei2USD(tx.Cost()),
	}
}

// CalculateReceiptDetails performs calculations on the receipt.
func (c *Converter) CalculateReceiptDetails(receipt *types.Receipt, gasPrice *big.Int) ReceiptDetails {
	cost := big.NewInt(0).Mul(big.NewInt(int64(receipt.GasUsed)), gasPrice)

	return ReceiptDetails{
		Status:        receipt.Status,
		GasUsed:       receipt.GasUsed,
		GasPriceGWei:  Wei2GWei(gasPrice).String(),
		GasPriceUSD:   c.Wei2USD(gasPrice),
		FinalCostGWei: Wei2GWei(cost).String(),
		FinalCostUSD:  c.Wei2USD(cost),
	}
}

// CalculateBalanceDiff performs calculations on the starting and ending balance.
func (c *Converter) CalculateBalanceDiff(startingBalance *big.Int, endingBalance *big.Int) (BalanceDiff, error) {
	cost := big.NewInt(0).Sub(startingBalance, endingBalance)

	return BalanceDiff{
		BeforeGWei: Wei2GWei(startingBalance).String(),
		AfterGWei:  Wei2GWei(endingBalance).String(),
		DiffGWei:   Wei2GWei(cost).String(),
		DiffUSD:    c.Wei2USD(cost),
	}, nil
}

// ExtractLogData pulls extra information from the receipt's logs.
func ExtractLogData(abiMetaData string, receipt *types.Receipt) ([]LogData, error) {
	type topic struct {
		Name string
		Func string
	}

	codeABI, err := abi.JSON(strings.NewReader(abiMetaData))
	if err != nil {
		return nil, err
	}

	events := make(map[common.Hash]topic)

	for _, event := range codeABI.Events {
		f := event.Name + "("
		for _, input := range event.Inputs {
			f = f + input.Type.String() + ","
		}
		f = f[:len(f)-1] + ")"

		events[crypto.Keccak256Hash([]byte(f))] = topic{
			Name: event.Name,
			Func: f,
		}
	}

	var logData []LogData

	for _, log := range receipt.Logs {
		for _, topic := range log.Topics {
			event, found := events[topic]
			if !found {
				continue
			}

			data := map[string]interface{}{}
			if err := codeABI.UnpackIntoMap(data, event.Name, log.Data); err != nil {
				return nil, err
			}

			logData = append(logData, LogData{
				EventName: event.Name,
				Data:      data,
			})
		}
	}

	return logData, nil
}

// FmtBalanceSheet produces a easy to read format of the starting and ending
// balance for the connected account.
func (c *Converter) FmtBalanceSheet(startingBalance *big.Int, endingBalance *big.Int) string {
	diff, err := c.CalculateBalanceDiff(startingBalance, endingBalance)
	if err != nil {
		return ""
	}

	return formatBalanceDiff(diff)
}

// FmtTransaction returns a human-readable format of the given transaction.
func (c *Converter) FmtTransaction(tx *types.Transaction) string {
	txDetails := c.CalculateTransactionDetails(tx)

	return formatTxCostDetails(txDetails)
}

// FmtTransactionReceipt produces a easy to read format of the specified receipt.
func (c *Converter) FmtTransactionReceipt(receipt *types.Receipt, gasPrice *big.Int) string {
	rcd := c.CalculateReceiptDetails(receipt, gasPrice)

	var b bytes.Buffer

	b.WriteString(formatReceiptCostDetails(rcd))
	logData, err := ExtractLogData(c.abiMetaData, receipt)
	if err == nil {
		b.WriteString(formatLogs(logData))
	}

	return b.String()
}
