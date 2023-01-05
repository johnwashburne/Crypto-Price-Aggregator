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

	go gemini.Recv()
	go cryptoCom.Recv()

	for {
		select {
		case c := <-cryptoCom.Updates:
			log.Printf("CC Bid: %s @ %s, Ask: %s @ %s", c.BidVolume, c.Bid, c.AskVolume, c.Ask)
		case g := <-gemini.Updates:
			log.Printf("G Bid: %s @ %s, Ask: %s @ %s", g.BidVolume, g.Bid, g.AskVolume, g.Ask)
		case <-interrupt:
			return
		}
	}
}
