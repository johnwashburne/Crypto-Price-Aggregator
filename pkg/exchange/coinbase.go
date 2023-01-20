package exchange

import (
	"fmt"

	"github.com/johnwashburne/Crypto-Price-Aggregator/pkg/logger"
	"github.com/johnwashburne/Crypto-Price-Aggregator/pkg/symbol"
	"github.com/johnwashburne/Crypto-Price-Aggregator/pkg/ws"
)

type Coinbase struct {
	updates chan MarketUpdate
	symbol  string
	name    string
	url     string
	valid   bool
	logger  *logger.Logger
}

func NewCoinbase(pair symbol.CurrencyPair) *Coinbase {
	c := make(chan MarketUpdate, updateBufSize)
	name := fmt.Sprintf("Coinbase: %s", pair.Coinbase)

	return &Coinbase{
		updates: c,
		symbol:  pair.Coinbase,
		name:    name,
		url:     "wss://ws-feed.exchange.coinbase.com",
		valid:   pair.Coinbase != "",
		logger:  logger.Named(name),
	}
}

// Receive book data from Coinbase, send any top of book updates
// over the updates channel as a MarketUpdate struct
func (e *Coinbase) Recv() {
	// connect to websocket
	e.logger.Debug("connecting to socket")
	conn := ws.New(e.url)

	conn.SetOnConnect(func(c *ws.Client) error {
		// subscribe to ticker channel
		err := conn.WriteJSON(coinbaseRequest{
			Type:       "subscribe",
			ProductIds: []string{e.symbol},
			Channels:   []string{"ticker"},
		})

		if err != nil {
			return err
		}

		// confirm accurate subscription
		var resp coinbaseSubscriptionResponse
		if err := conn.ReadJSON(&resp); err != nil {
			return err
		}

		return nil
	})

	if err := conn.Connect(); err != nil {
		e.logger.Warn("could not connect to socket")
		return
	}
	e.logger.Debug("connected to socket")

	for {
		var message coinbaseMessage
		if err := conn.ReadJSON(&message); err != nil {
			e.logger.Info(err)
			continue
		}

		e.updates <- MarketUpdate{
			Ask:     message.BestAsk,
			AskSize: message.BestAskSize,
			Bid:     message.BestBid,
			BidSize: message.BestBidSize,
			Name:    e.name,
		}
	}
}

// Name of data source
func (e *Coinbase) Name() string {
	return e.name
}

// Access to update channel
func (e *Coinbase) Updates() chan MarketUpdate {
	return e.updates
}

func (e *Coinbase) Valid() bool {
	return e.valid
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
