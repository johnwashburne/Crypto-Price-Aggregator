// A symbol manager to account for variations in how various cryptocurrency are referenced
// across different exchanges
// Currently using a naive local json implementation

package symbol

import (
	"encoding/json"
	"io"
	"os"
)

type SymbolManager interface {
	GetCurrencyPair(baseCurrency string, quoteCurrency string) CurrencyPair
}

type CurrencyPair struct {
	BinanceUS string `json:"Binance.US"`
	Bitstamp  string `json:"Bitstamp"`
	Coinbase  string `json:"Coinbase"`
	CryptoCom string `json:"Crypto.com"`
	Gemini    string `json:"Gemini"`
	Kraken    string `json:"Kraken"`
	Kucoin    string `json:"Kucoin"`
}

type JsonManager struct {
	data map[string]map[string]CurrencyPair
}

// Load symbol data from the local json file
func LoadJsonSymbolData() (*JsonManager, error) {
	jsonFile, err := os.Open("./pkg/symbol/symbol_database.json")
	if err != nil {
		return nil, err
	}
	defer jsonFile.Close()
	bytes, err := io.ReadAll(jsonFile)
	if err != nil {
		return nil, err
	}

	var data map[string]map[string]CurrencyPair
	json.Unmarshal(bytes, &data)

	return &JsonManager{
		data: data,
	}, nil
}

// Get a currency pair from the json SymbolManager implementation
func (j *JsonManager) GetCurrencyPair(baseCurrency string, quoteCurrency string) CurrencyPair {
	return j.data[baseCurrency][quoteCurrency]
}
