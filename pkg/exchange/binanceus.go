package exchange

import (
	"fmt"

	"github.com/johnwashburne/Crypto-Price-Aggregator/pkg/logger"
	"github.com/johnwashburne/Crypto-Price-Aggregator/pkg/symbol"
	"github.com/johnwashburne/Crypto-Price-Aggregator/pkg/ws"
)

type BinanceUS struct {
	updates chan MarketUpdate
	url     string
	name    string
	symbol  string
	valid   bool
	logger  *logger.Logger
}

func NewBinanceUS(pair symbol.CurrencyPair) *BinanceUS {
	c := make(chan MarketUpdate, updateBufSize)
	name := fmt.Sprintf("Binance.US: %s", pair.BinanceUS)

	return &BinanceUS{
		updates: c,
		url:     fmt.Sprintf("wss://stream.binance.us:9443/ws/%s@bookTicker", pair.BinanceUS),
		name:    name,
		symbol:  pair.BinanceUS,
		valid:   pair.BinanceUS != "",
		logger:  logger.Named(name),
	}
}

func (e *BinanceUS) Recv() {
	e.logger.Debug("connecting to socket")
	conn := ws.New(e.url)
	if err := conn.Connect(); err != nil {
		e.logger.Warn("could not connect to socket, RETURNING")
		return
	}
	e.logger.Debug("connected to socket")

	for {
		var message binanceUSMessage
		if err := conn.ReadJSON(&message); err != nil {
			e.logger.Warn(err, " RETURNING")
			return
		}

		e.updates <- MarketUpdate{
			Ask:     message.Ask,
			AskSize: message.AskSize,
			Bid:     message.Bid,
			BidSize: message.BidSize,
			Name:    e.name,
		}
	}
}

func (e *BinanceUS) Updates() chan MarketUpdate {
	return e.updates
}

func (e *BinanceUS) Name() string {
	return e.name
}

func (e *BinanceUS) Valid() bool {
	return e.valid
}

type binanceUSMessage struct {
	UpdateID int    `json:"u"`
	Symbol   string `json:"s"`
	Bid      string `json:"b"`
	BidSize  string `json:"B"`
	Ask      string `json:"a"`
	AskSize  string `json:"A"`
}
