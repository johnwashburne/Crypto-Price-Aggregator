package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/johnwashburne/Crypto-Price-Aggregator/exchange"
	"github.com/johnwashburne/Crypto-Price-Aggregator/symbol"
)

func main() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	symbolManager, err := symbol.LoadJsonSymbolData()
	if err != nil {
		log.Println("Could not load symbol manager:", err)
		return
	}

	pair := symbolManager.GetCurrencyPair("BTC", "USD")
	gemini := exchange.NewGemini(pair)
	cryptoCom := exchange.NewCryptoCom(pair)
	coinbase := exchange.NewCoinbase(pair)

	go gemini.Recv()
	go cryptoCom.Recv()
	go coinbase.Recv()

	for {
		select {
		case cc := <-cryptoCom.Updates:
			log.Printf("CC Bid: %s @ %s, Ask: %s @ %s", cc.BidVolume, cc.Bid, cc.AskVolume, cc.Ask)
		case g := <-gemini.Updates:
			log.Printf("G Bid: %s @ %s, Ask: %s @ %s", g.BidVolume, g.Bid, g.AskVolume, g.Ask)
		case cb := <-coinbase.Updates:
			log.Printf("CB Bid: %s @ %s, Ask %s @ %s", cb.BidVolume, cb.Bid, cb.AskVolume, cb.Ask)
		case <-interrupt:
			return
		}
	}
}
