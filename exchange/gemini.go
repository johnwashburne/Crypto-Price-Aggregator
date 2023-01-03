package exchange

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/gorilla/websocket"
	"github.com/johnwashburne/Crypto-Price-Aggregator/symbol"
)

type Gemini struct {
	Updates chan MarketUpdate
	url     string
	name    string
	symbol  string
}

// Create new Gemini struct
func NewGemini(pair symbol.CurrencyPair) *Gemini {
	c := make(chan MarketUpdate, updateBufSize)

	return &Gemini{
		Updates: c,
		url:     fmt.Sprintf("wss://api.gemini.com/v1/marketdata/%s?top_of_book=true", pair.Gemini),
		name:    "Gemini",
		symbol:  pair.Gemini,
	}
}

// Receive book data from Gemini, send any top of book updates
// over the Updates channel as a MarketUpdate struct
func (g *Gemini) Recv() error {
	log.Printf("Connecting to %s\n", g.url)
	conn, _, err := websocket.DefaultDialer.Dial(g.url, nil)
	if err != nil {
		return err
	}
	defer conn.Close()

	var ask string
	var askVol string
	var bid string
	var bidVol string
	for {
		_, raw_msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("Gemini:", err)
			continue
		}

		msg := geminiMessage{}
		err = json.Unmarshal([]byte(raw_msg), &msg)

		for _, e := range msg.Events {
			if e.Side == "bid" {
				bid = e.Price
				bidVol = e.Remaining
			}
			if e.Side == "ask" {
				ask = e.Price
				askVol = e.Remaining
			}
		}

		update := MarketUpdate{bid, ask, bidVol, askVol}
		g.Updates <- update
	}
}

// Name of data source
func (g *Gemini) Name() string {
	return g.name
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
