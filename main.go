package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/johnwashburne/Crypto-Price-Aggregator/pkg/aggregator"
	"github.com/johnwashburne/Crypto-Price-Aggregator/pkg/exchange"
	"github.com/johnwashburne/Crypto-Price-Aggregator/pkg/symbol"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {

	initLogger()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	symbolManager, err := symbol.LoadJsonSymbolData()
	if err != nil {
		return
	}

	pair := symbolManager.GetCurrencyPair("ETH", "USD")

	agg := aggregator.New(
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

func initLogger() {
	// set up logging
	logger, _ := zap.Config{
		Encoding:    "json",
		Level:       zap.NewAtomicLevelAt(zapcore.DebugLevel),
		OutputPaths: []string{"logs/logs.txt"},
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey: "message", // <--
			NameKey:    "name",

			TimeKey:    "time",
			EncodeTime: zapcore.ISO8601TimeEncoder,

			CallerKey:    "caller",
			EncodeCaller: zapcore.ShortCallerEncoder,

			LevelKey:    "level",
			EncodeLevel: zapcore.CapitalLevelEncoder,
		},
	}.Build()
	zap.ReplaceGlobals(logger)

}
