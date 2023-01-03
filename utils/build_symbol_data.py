import requests

def get_gemini_symbols() -> dict:
    response = requests.get("https://api.gemini.com/v1/symbols")
    symbols = response.json()
    
    for symbol in symbols:
        response = requests.get(f"https://api.gemini.com/v1/symbols/details/{symbol}")
        print(response.json())

def get_crypto_com_symbols() -> dict:
    response = requests.get("https://api.crypto.com/v2/public/get-instruments")
    for instrument in response.json()['result']['instruments']:
        print(instrument)

if __name__ == "__main__":
    get_gemini_symbols()