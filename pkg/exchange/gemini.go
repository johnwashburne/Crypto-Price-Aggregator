// Aggreagate top of book updates from a currency pair listed on Gemini

package exchange

import (
	"fmt"

	"github.com/johnwashburne/Crypto-Price-Aggregator/pkg/logger"
	"github.com/johnwashburne/Crypto-Price-Aggregator/pkg/symbol"
	"github.com/johnwashburne/Crypto-Price-Aggregator/pkg/ws"
)

type Gemini struct {
	updates chan MarketUpdate
	url     string
	name    string
	symbol  string
	valid   bool
	logger  *logger.Logger
}

// Create new Gemini struct
func NewGemini(pair symbol.CurrencyPair) *Gemini {
	c := make(chan MarketUpdate, updateBufSize)
	name := fmt.Sprintf("Gemini: %s", pair.Gemini)

	return &Gemini{
		updates: c,
		url:     fmt.Sprintf("wss://api.gemini.com/v1/marketdata/%s?top_of_book=true", pair.Gemini),
		name:    name,
		symbol:  pair.Gemini,
		valid:   pair.Gemini != "",
		logger:  logger.Named(name),
	}
}

// Receive book data from Gemini, send any top of book updates
// over the updates channel as a MarketUpdate struct
func (e *Gemini) Recv() {
	e.logger.Debug("connecting to socket")
	conn := ws.New(e.url)
	if err := conn.Connect(); err != nil {
		e.logger.Info("could not connect to socket")
		return
	}
	e.logger.Debug("connected to socket")

	var ask string
	var askSize string
	var bid string
	var bidSize string
	for {
		var message geminiMessage
		if err := conn.ReadJSON(&message); err != nil {
			e.logger.Warn(err)
			continue
		}

		for _, event := range message.Events {
			if event.Side == "bid" {
				bid = event.Price
				bidSize = event.Remaining
			}
			if event.Side == "ask" {
				ask = event.Price
				askSize = event.Remaining
			}
		}

		e.updates <- MarketUpdate{
			Ask:     ask,
			AskSize: askSize,
			Bid:     bid,
			BidSize: bidSize,
			Name:    e.name,
		}
	}
}

// Name of data source
func (e *Gemini) Name() string {
	return e.name
}

func (e *Gemini) Updates() chan MarketUpdate {
	return e.updates
}

func (e *Gemini) Valid() bool {
	return e.valid
}

// Struct to represent Gemini json message
type geminiMessage struct {
	Type           string        `json:"type"`
	EventId        int           `json:"eventId"`
	Timestamp      int           `json:"timestamp"`
	TimestampMS    int           `json:"timestampms"`
	SocketSequence int           `json:"socket_sequence"`
	Events         []geminiEvent `json:"events"`
}

// Struct to represent Gemini json message
type geminiEvent struct {
	Type      string `json:"type"`
	Side      string `json:"side"`
	Price     string `json:"price"`
	Remaining string `json:"remaining"`
	Delta     string `json:"delta"`
	Reason    string `json:"reason"`
}
