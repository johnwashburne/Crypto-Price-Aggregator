package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/johnwashburne/crypto-arbitrage-go/exchange"
)

func main() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	geminiUrl := "wss://api.gemini.com/v1/marketdata/btcusd?top_of_book=true"
	gemini := exchange.NewGemini(geminiUrl)

	cryptoCom := exchange.NewCryptoCom("wss://uat-stream.3ona.co/v2/market")

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
