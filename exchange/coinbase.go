package exchange

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/gorilla/websocket"
	"github.com/johnwashburne/Crypto-Price-Aggregator/symbol"
)

type Coinbase struct {
	updates chan MarketUpdate
	symbol  string
	name    string
	url     string
}

func NewCoinbase(pair symbol.CurrencyPair) *Coinbase {
	c := make(chan MarketUpdate, updateBufSize)

	return &Coinbase{
		updates: c,
		symbol:  pair.Coinbase,
		name:    fmt.Sprintf("Coinbase: %s", pair.Coinbase),
		url:     "wss://ws-feed.exchange.coinbase.com",
	}
}

// Receive book data from Coinbase, send any top of book updates
// over the updates channel as a MarketUpdate struct
func (c *Coinbase) Recv() {
	// connect to websocket
	log.Printf("%s - Connecting to %s\n", c.name, c.url)
	conn, _, err := websocket.DefaultDialer.Dial(c.url, nil)
	if err != nil {
		log.Println("Could not connect to", c.name)
		return
	}
	defer conn.Close()

	// subscribe to ticker channel
	conn.WriteJSON(coinbaseRequest{
		Type:       "subscribe",
		ProductIds: []string{c.symbol},
		Channels:   []string{"ticker"},
	})

	// confirm accurate subscription
	_, raw_msg, err := conn.ReadMessage()
	if err != nil {
		log.Println(c.name, err)
		return
	}
	var resp coinbaseSubscriptionResponse
	json.Unmarshal(raw_msg, &resp)
	// TODO: subscription verification

	for {
		_, raw_msg, err := conn.ReadMessage()
		if err != nil {
			log.Println(c.name, err)
			continue
		}

		var msg coinbaseMessage
		json.Unmarshal(raw_msg, &msg)

		c.updates <- MarketUpdate{
			Ask:     msg.BestAsk,
			AskSize: msg.BestAskSize,
			Bid:     msg.BestBid,
			BidSize: msg.BestBidSize,
			Name:    c.name,
		}
	}
}

// Name of data source
func (c *Coinbase) Name() string {
	return c.name
}

// Access to update channel
func (c *Coinbase) Updates() chan MarketUpdate {
	return c.updates
}

type coinbaseRequest struct {
	Type       string   `json:"type"`
	ProductIds []string `json:"product_ids"`
	Channels   []string `json:"channels"`
}

type coinbaseSubscriptionResponse struct {
	Type     string                   `json:"type"`
	Channels []map[string]interface{} `json:"channels"`
}

type coinbaseMessage struct {
	Type        string `json:"type"`
	Sequence    int    `json:"sequence"`
	ProductId   string `json:"ETH-USD"`
	Price       string `json:"price"`
	Open24h     string `json:"open_24h"`
	Volume24h   string `json:"volume_24h"`
	Low24h      string `json:"low_24h"`
	High24h     string `json:"high_24h"`
	Volume30d   string `json:"volume_30d"`
	BestBid     string `json:"best_bid"`
	BestBidSize string `json:"best_bid_size"`
	BestAsk     string `json:"best_ask"`
	BestAskSize string `json:"best_ask_size"`
	Side        string `json:"side"`
	Time        string `json:"time"`
	TradeId     int    `json:"trade_id"`
	LastSize    string `json:"last_size"`
}
