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
	Gemini    string `json:"Gemini"`
	CryptoCom string `json:"Crypto.com"`
	Coinbase  string `json:"Coinbase"`
}

type JsonManager struct {
	data map[string]map[string]CurrencyPair
}

func LoadJsonSymbolData() (*JsonManager, error) {
	jsonFile, err := os.Open("./symbol/symbol_database.json")
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

func (j *JsonManager) GetCurrencyPair(baseCurrency string, quoteCurrency string) CurrencyPair {
	return j.data[baseCurrency][quoteCurrency]
}
