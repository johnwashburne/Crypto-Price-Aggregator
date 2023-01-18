# python util file for building json symbol database

import requests
from typing import List
import json
from tqdm import tqdm


class Symbol:

    def __init__(self, symbol: str, base_currency: str, quote_currency: str):
        self.symbol = symbol
        self.base_currency = base_currency
        self.quote_currency = quote_currency


def get_gemini_symbols() -> List[Symbol]:
    response = requests.get("https://api.gemini.com/v1/symbols")
    instruments = response.json()

    res = []

    for instrument in tqdm(instruments):
        response = requests.get(
            f"https://api.gemini.com/v1/symbols/details/{instrument}"
        ).json()

        res.append(Symbol(
            response['symbol'],
            response['base_currency'],
            response['quote_currency']
        ))

    return res

def get_crypto_com_symbols() -> List[Symbol]:
    response = requests.get("https://api.crypto.com/v2/public/get-instruments")

    res = []

    for instrument in tqdm(response.json()['result']['instruments']):
        res.append(Symbol(
            instrument['instrument_name'],
            instrument['base_currency'],
            instrument['quote_currency']
        ))

    return res

def get_coinbase_symbols() -> List[Symbol]:
    response = requests.get("https://api.exchange.coinbase.com/products/")

    res = []

    for instrument in tqdm(response.json()):
        res.append(Symbol(
            instrument['id'],
            instrument['base_currency'],
            instrument['quote_currency']
        ))

    return res

def get_kraken_symbols() -> List[Symbol]:
    response = requests.get("https://api.kraken.com/0/public/AssetPairs")
    res = []

    for name in tqdm(response.json()['result'].keys()):
        instrument = response.json()['result'][name]
        res.append(Symbol(
            instrument['wsname'],
            instrument['base'],
            instrument['quote']
        ))

    return res

def get_kucoin_symbols() -> List[Symbol]:
    response = requests.get("https://api.kucoin.com/api/v2/symbols")
    res = []

    for instrument in tqdm(response.json()['data']):
        res.append(Symbol(
            instrument['symbol'],
            instrument['baseCurrency'],
            instrument['quoteCurrency']
        ))

    return res

def get_bitstamp_symbols() -> List[Symbol]:
    response = requests.get("https://www.bitstamp.net/api/v2/ticker/")
    res = []

    for instrument in tqdm(response.json()):
        base = instrument['pair'].split("/")[0]
        quote = instrument['pair'].split("/")[1]
        res.append(Symbol(
            instrument['pair'].replace("/", "").lower(),
            base,
            quote
        ))

    return res

def add_symbols(
        curr_symbols: dict, 
        new_symbols: List[Symbol], 
        name: str) -> dict:

    for s in new_symbols:
        if s.base_currency not in curr_symbols.keys():
            curr_symbols[s.base_currency] = {}

        if s.quote_currency not in curr_symbols[s.base_currency].keys():
            curr_symbols[s.base_currency][s.quote_currency] = {}

        curr_symbols[s.base_currency][s.quote_currency][name] = s.symbol

    return curr_symbols


if __name__ == "__main__":
    
    symbols = {}
    
    add_symbols(symbols, get_bitstamp_symbols(), "Bitstamp")
    add_symbols(symbols, get_gemini_symbols(), "Gemini")
    add_symbols(symbols, get_crypto_com_symbols(), "Crypto.com")
    add_symbols(symbols, get_coinbase_symbols(), "Coinbase")
    add_symbols(symbols, get_kucoin_symbols(), "Kucoin")

    with open("../symbol/symbol_database.json", 'w') as f:
        json.dump(symbols, f)
