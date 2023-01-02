package exchange

import (
	"encoding/json"
	"log"

	"github.com/gorilla/websocket"
)

type Gemini struct {
	Updates chan MarketUpdate
	url     string
}

// Create new Gemini struct
func NewGemini(url string) *Gemini {
	c := make(chan MarketUpdate, 100)

	return &Gemini{
		Updates: c,
		url:     url,
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
			log.Println(err)
			return err
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
	return "Gemini"
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
