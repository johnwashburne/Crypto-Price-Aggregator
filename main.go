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

	pair := symbolManager.GetCurrencyPair("BTC", "USDT")
	cc := exchange.NewCryptoCom(pair)

	go cc.Recv()
	for {
		select {
		case msg := <-cc.Updates():
			log.Println(msg)
		case <-interrupt:
			return
		}
	}

	/*
		agg := aggregator.New(
			exchange.NewGemini(pair),
			exchange.NewCryptoCom(pair),
			exchange.NewCoinbase(pair),
			//exchange.NewKucoin(pair),
		)

		go agg.Recv()
		for {
			select {
			case msg, ok := <-agg.Updates():
				if !ok {
					return
				}
				log.Println(msg.Bid, msg.BidPlatform, msg.Ask, msg.AskPlatform)
			case <-interrupt:
				return
			}
		}
	*/

}
