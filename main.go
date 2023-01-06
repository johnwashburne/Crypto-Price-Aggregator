package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/johnwashburne/Crypto-Price-Aggregator/aggregator"
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
	agg := aggregator.New(
		exchange.NewGemini(pair),
		exchange.NewCryptoCom(pair),
		exchange.NewCoinbase(pair),
	)

	go agg.Recv()

	for {
		select {
		case msg := <-agg.Updates:
			log.Println(msg.Bid, msg.BidPlatform, msg.Ask, msg.AskPlatform)
		case <-interrupt:
			return
		}
	}
}
