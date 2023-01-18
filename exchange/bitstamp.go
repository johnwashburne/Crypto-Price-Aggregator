package exchange

import (
	"fmt"
	"log"

	"github.com/johnwashburne/Crypto-Price-Aggregator/symbol"
	"github.com/johnwashburne/Crypto-Price-Aggregator/ws"
)

type Bitstamp struct {
	updates chan MarketUpdate
	url     string
	name    string
	symbol  string
	valid   bool
}

func NewBitstamp(pair symbol.CurrencyPair) *Bitstamp {
	c := make(chan MarketUpdate, updateBufSize)

	return &Bitstamp{
		updates: c,
		url:     "wss://ws.bitstamp.net",
		symbol:  pair.Bitstamp,
		name:    fmt.Sprintf("Bitstamp: %s", pair.Bitstamp),
		valid:   pair.Bitstamp != "",
	}
}

func (e *Bitstamp) Recv() {
	log.Printf("%s - Connecting to %s\n", e.name, e.url)
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
		log.Println("Could not connect to", e.name)
		return
	}

	lastUpdate := MarketUpdate{}
	for {
		var message bitstampOrderBook
		if err := conn.ReadJSON(&message); err != nil {
			log.Println(e.name, err)
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
