package exchange

import (
	"fmt"
	"log"

	"github.com/johnwashburne/Crypto-Price-Aggregator/symbol"
	"github.com/johnwashburne/Crypto-Price-Aggregator/ws"
)

type Gemini struct {
	updates chan MarketUpdate
	url     string
	name    string
	symbol  string
	valid   bool
}

// Create new Gemini struct
func NewGemini(pair symbol.CurrencyPair) *Gemini {
	c := make(chan MarketUpdate, updateBufSize)

	return &Gemini{
		updates: c,
		url:     fmt.Sprintf("wss://api.gemini.com/v1/marketdata/%s?top_of_book=true", pair.Gemini),
		name:    fmt.Sprintf("Gemini: %s", pair.Gemini),
		symbol:  pair.Gemini,
		valid:   pair.Gemini != "",
	}
}

// Receive book data from Gemini, send any top of book updates
// over the updates channel as a MarketUpdate struct
func (e *Gemini) Recv() {
	log.Printf("%s - Connecting to %s\n", e.name, e.url)
	conn := ws.New(e.url)
	err := conn.Connect()
	if err != nil {
		log.Println("Could not connect to", e.name)
		return
	}

	var ask string
	var askSize string
	var bid string
	var bidSize string
	for {
		var message geminiMessage
		err = conn.ReadJSON(&message)
		if err != nil {
			log.Println(e.name, err)
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
