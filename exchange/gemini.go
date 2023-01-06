package exchange

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/gorilla/websocket"
	"github.com/johnwashburne/Crypto-Price-Aggregator/symbol"
)

type Gemini struct {
	updates chan MarketUpdate
	url     string
	name    string
	symbol  string
}

// Create new Gemini struct
func NewGemini(pair symbol.CurrencyPair) *Gemini {
	c := make(chan MarketUpdate, updateBufSize)

	return &Gemini{
		updates: c,
		url:     fmt.Sprintf("wss://api.gemini.com/v1/marketdata/%s?top_of_book=true", pair.Gemini),
		name:    fmt.Sprintf("Gemini: %s", pair.Gemini),
		symbol:  pair.Gemini,
	}
}

// Receive book data from Gemini, send any top of book updates
// over the updates channel as a MarketUpdate struct
func (g *Gemini) Recv() {
	log.Printf("%s - Connecting to %s\n", g.name, g.url)
	conn, _, err := websocket.DefaultDialer.Dial(g.url, nil)
	if err != nil {
		log.Println("Could not connect to ", g.name)
	}
	defer conn.Close()

	var ask string
	var askSize string
	var bid string
	var bidSize string
	for {
		_, raw_msg, err := conn.ReadMessage()
		if err != nil {
			log.Println(g.name, err)
			continue
		}

		msg := geminiMessage{}
		json.Unmarshal([]byte(raw_msg), &msg)

		for _, e := range msg.Events {
			if e.Side == "bid" {
				bid = e.Price
				bidSize = e.Remaining
			}
			if e.Side == "ask" {
				ask = e.Price
				askSize = e.Remaining
			}
		}

		g.updates <- MarketUpdate{
			Ask:     ask,
			AskSize: askSize,
			Bid:     bid,
			BidSize: bidSize,
			Name:    g.name,
		}
	}
}

// Name of data source
func (g *Gemini) Name() string {
	return g.name
}

func (g *Gemini) Updates() chan MarketUpdate {
	return g.updates
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
