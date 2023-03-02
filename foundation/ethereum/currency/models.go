package currency

// Status represents information about the http call.
type Status struct {
	Timestamp    string `json:"timestamp"`
	ErrorCode    int    `json:"error_code"`
	ErrorMessage string `json:"error_message"`
	Elapsed      int    `json:"elapsed"`
	CreditCount  int    `json:"credit_count"`
}

// DataUSD represents data returned when requesting USD.
type DataUSD struct {
	Symbol      string `json:"symbol"`
	Amount      int    `json:"amount"`
	LastUpdated string `json:"last_updated"`
	Quote       struct {
		USD struct {
			Price       float64 `json:"price"`
			LastUpdated string  `json:"last_updated"`
		} `json:"usd"`
	} `json:"quote"`
}

// DataETH represents data returned when requesting ETH.
type DataETH struct {
	Symbol      string `json:"symbol"`
	Amount      int    `json:"amount"`
	LastUpdated string `json:"last_updated"`
	Quote       struct {
		ETH struct {
			Price       float64 `json:"price"`
			LastUpdated string  `json:"last_updated"`
		} `json:"eth"`
	} `json:"quote"`
}

// ResponseETH2USD represents a response for converting ETH to USD.
type ResponseETH2USD struct {
	Status Status    `json:"status"`
	Data   []DataUSD `json:"data"`
}

// ResponseUSD2ETH represents a response for converting USD to ETH.
type ResponseUSD2ETH struct {
	Status Status    `json:"status"`
	Data   []DataETH `json:"data"`
}

// TransactionDetails holds details about a transaction and its cost.
type TransactionDetails struct {
	Hash              string
	Nonce             uint64
	GasLimit          uint64
	GasOfferPriceGWei string
	Value             string
	MaxGasPriceGWei   string
	MaxGasPriceUSD    string
}

// ReceiptDetails holds details about a receipt and its cost.
type ReceiptDetails struct {
	Status        uint64
	GasUsed       uint64
	GasPriceGWei  string
	GasPriceUSD   string
	FinalCostGWei string
	FinalCostUSD  string
}

// BalanceDiff performs calculations on the starting and ending balance.
type BalanceDiff struct {
	BeforeGWei string
	AfterGWei  string
	DiffGWei   string
	DiffUSD    string
}

// LogData represents data we can pull from events in the receipt logs.
type LogData struct {
	EventName string
	Data      map[string]any
}
