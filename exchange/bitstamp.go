package exchange

import (
	"fmt"

	"github.com/johnwashburne/Crypto-Price-Aggregator/symbol"
	"github.com/johnwashburne/Crypto-Price-Aggregator/ws"
	"go.uber.org/zap"
)

type Bitstamp struct {
	updates chan MarketUpdate
	url     string
	name    string
	symbol  string
	valid   bool
	logger  *zap.SugaredLogger
}

func NewBitstamp(pair symbol.CurrencyPair) *Bitstamp {
	c := make(chan MarketUpdate, updateBufSize)
	name := fmt.Sprintf("Bitstamp: %s", pair.Bitstamp)

	return &Bitstamp{
		updates: c,
		url:     "wss://ws.bitstamp.net",
		symbol:  pair.Bitstamp,
		name:    name,
		valid:   pair.Bitstamp != "",
		logger:  zap.S().Named(name),
	}
}

func (e *Bitstamp) Recv() {
	e.logger.Debug("connecting to socket")
	conn := ws.New(e.url)

	conn.SetOnConnect(func(c *ws.Client) error {
		err := c.WriteJSON(bitstampSubscription{
			bitstampMessage: bitstampMessage{
				Event: "bts:subscribe",
			},
			Data: map[string]string{
				"channel": fmt.Sprintf("order_book_%s", e.symbol),
			},
		})

		if err != nil {
			return err
		}

		var subcribtionAck bitstampMessage
		return c.ReadJSON(&subcribtionAck)
	})

	if err := conn.Connect(); err != nil {
		e.logger.Warn("Could not connect to socket")
		return
	}
	e.logger.Debug("connected to socket")

	lastUpdate := MarketUpdate{}
	for {
		var message bitstampOrderBook
		if err := conn.ReadJSON(&message); err != nil {
			e.logger.Warn(err)
			return
		}

		update := MarketUpdate{
			Bid:     message.Data.Bids[0][0],
			BidSize: message.Data.Bids[0][1],
			Ask:     message.Data.Asks[0][0],
			AskSize: message.Data.Asks[0][1],
			Name:    e.name,
		}

		if update != lastUpdate {
			e.updates <- update
		}
		lastUpdate = update
	}
}

func (e *Bitstamp) Name() string {
	return e.name
}

func (e *Bitstamp) Updates() chan MarketUpdate {
	return e.updates
}

func (e *Bitstamp) Valid() bool {
	return e.valid
}

type bitstampMessage struct {
	Event   string `json:"event"`
	Channel string `json:"channel,omitempty"`
}

type bitstampSubscription struct {
	bitstampMessage
	Data map[string]string `json:"data"`
}

type bitstampOrderBook struct {
	bitstampMessage
	Data bitstampOrderBookData `json:"data"`
}

type bitstampOrderBookData struct {
	Timestamp      string     `json:"timestamp"`
	Microtimestamp string     `json:"microtimestamp"`
	Bids           [][]string `json:"bids"`
	Asks           [][]string `json:"asks"`
}
