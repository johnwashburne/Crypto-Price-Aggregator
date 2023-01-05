# python util file for building json symbol database

import requests
from typing import List
import json

def get_gemini_symbols() -> List[dict]:
    response = requests.get("https://api.gemini.com/v1/symbols")
    symbols = response.json()
    
    res = []

    for symbol in symbols:
        response = requests.get(f"https://api.gemini.com/v1/symbols/details/{symbol}").json()
        item = {}
        item['symbol'] = response['symbol']
        item['base_currency'] = response['base_currency']
        item['quote_currency'] = response['quote_currency']
        res.append(item)

    return res

def get_crypto_com_symbols() -> List[dict]:
    response = requests.get("https://api.crypto.com/v2/public/get-instruments")

    res = []

    for instrument in response.json()['result']['instruments']:
        item = {}
        item['symbol'] = instrument['instrument_name']
        item['base_currency'] = instrument['base_currency']
        item['quote_currency'] = instrument['quote_currency']
        res.append(item)

    return res
    
def add_symbols(curr_symbols: dict, new_symbols: List[dict], name: str) -> dict:
    for s in new_symbols:
        if s['base_currency'] not in curr_symbols.keys():
            curr_symbols[s['base_currency']] = {}

        if s['quote_currency'] not in curr_symbols[s['base_currency']].keys():
            curr_symbols[s['base_currency']][s['quote_currency']] = {}

        curr_symbols[s['base_currency']][s['quote_currency']][name] = s['symbol']

    return curr_symbols

if __name__ == "__main__":
    symbols = {}
    add_symbols(symbols, get_gemini_symbols(), "Gemini")
    add_symbols(symbols, get_crypto_com_symbols(), "Crypto.com")
    with open("../symbol/symbol_database.json", 'w') as f:
        json.dump(symbols, f)

