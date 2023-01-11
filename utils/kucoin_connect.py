import requests

if __name__ == "__main__":
    response = requests.post("https://api.kucoin.com/api/v1/bullet-public")
    print(response.json())