package symbol

import (
	"encoding/json"
	"io"
	"os"
)

type CurrencyPair struct {
	Gemini    string `json:"Gemini"`
	CryptoCom string `json:"Crypto.com"`
}

type SymbolManager struct {
	data map[string]map[string]CurrencyPair
}

func LoadData() (*SymbolManager, error) {
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

	return &SymbolManager{
		data: data,
	}, nil
}

func (s *SymbolManager) GetCurrencyPair(baseCurrency string, quoteCurrency string) CurrencyPair {
	return s.data[baseCurrency][quoteCurrency]
}
