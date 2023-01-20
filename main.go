package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/johnwashburne/Crypto-Price-Aggregator/pkg/aggregator"
	"github.com/johnwashburne/Crypto-Price-Aggregator/pkg/exchange"
	"github.com/johnwashburne/Crypto-Price-Aggregator/pkg/logger"
	"github.com/johnwashburne/Crypto-Price-Aggregator/pkg/symbol"
)

func main() {

	logger.CreateLogger()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	symbolManager, err := symbol.LoadJsonSymbolData()
	if err != nil {
		return
	}

	pair := symbolManager.GetCurrencyPair("BTC", "USDT")
	fmt.Println(pair.BinanceUS)

	agg := aggregator.New(
		exchange.NewBinanceUS(pair),
		exchange.NewBitstamp(pair),
		exchange.NewCoinbase(pair),
		exchange.NewCryptoCom(pair),
		exchange.NewGemini(pair),
		exchange.NewKucoin(pair),
	)

	go agg.Recv()
	for {
		select {
		case msg, ok := <-agg.Updates():
			if !ok {
				return
			}
			fmt.Println(msg.Bid, msg.BidPlatform, msg.Ask, msg.AskPlatform)
		case <-interrupt:
			return
		}
	}

}
