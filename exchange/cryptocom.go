package exchange

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/johnwashburne/Crypto-Price-Aggregator/symbol"
	"github.com/johnwashburne/Crypto-Price-Aggregator/ws"
)

// message method for heartbeat response
const heartbeatRequestMethod = "public/respond-heartbeat"

type CryptoCom struct {
	updates chan MarketUpdate
	url     string
	name    string
	symbol  string
	valid   bool
}

// create new Crypto.com struct
func NewCryptoCom(pair symbol.CurrencyPair) *CryptoCom {
	c := make(chan MarketUpdate, updateBufSize)

	return &CryptoCom{
		updates: c,
		url:     "wss://stream.crypto.com/exchange/v1/market",
		name:    fmt.Sprintf("Crypto.com: %s", pair.CryptoCom),
		symbol:  pair.CryptoCom,
		valid:   pair.CryptoCom != "",
	}
}

// Receive book data from Crypto.com, send any top of book updates
// over the updates channel as a MarketUpdate struct
func (e *CryptoCom) Recv() {
	log.Printf("%s - Connecting to %s\n", e.name, e.url)
	conn := ws.New(e.url)

	conn.SetOnConnect(func(c *ws.Client) error {
		c.WriteJSON(buildCryptoComSubscription(e.symbol))

		var resp cryptoComSubscriptionResponse
		conn.ReadJSON(&resp)

		if resp.Method != "subscribe" || resp.Code != 0 {
			return errors.New(fmt.Sprint(e.name, "could not subscribe"))
		}

		return nil
	})

	if err := conn.Connect(); err != nil {
		log.Println("Could not connect to", e.name)
		return
	}

	lastUpdate := MarketUpdate{}
	for {
		_, raw_msg, err := conn.ReadMessage()
		if err != nil {
			log.Println(e.name, err)
			return
		}

		var message cryptoComMessage
		json.Unmarshal(raw_msg, &message)

		if message.Method == "public/heartbeat" {
			// Crypto.com requires a heartbeat message response every
			// 	 30 seconds to keep websocket connection alive
			var h cryptoComHeartbeat
			json.Unmarshal(raw_msg, &h)
			conn.WriteJSON(cryptoComMessage{
				Id:     h.Id,
				Method: heartbeatRequestMethod,
			})
		} else if message.Method == "subscribe" {
			var bookMsg cryptoComBookMsg
			json.Unmarshal(raw_msg, &bookMsg)

			update := parseCryptoComBookData(&bookMsg)
			update.Name = e.name
			if update != lastUpdate {
				e.updates <- update
			}

			lastUpdate = update
		} else {
			log.Println(e.name, "unidentified message:", message.Method)
		}
	}
}

// Name of data source
func (e *CryptoCom) Name() string {
	return e.name
}

// Access to update channel
func (e *CryptoCom) Updates() chan MarketUpdate {
	return e.updates
}

func (e *CryptoCom) Valid() bool {
	return e.valid
}

// parse a Crypto.com book websocket message into our market update object
// best bid and ask, as well as volume for both
func parseCryptoComBookData(c *cryptoComBookMsg) MarketUpdate {
	var ask string
	var askSize string
	if len(c.Result.Data[0].Asks) != 0 {
		askSlice := c.Result.Data[0].Asks
		ask = askSlice[0][0]
		askSize = stringMultiply(askSlice[0][1], askSlice[0][2])
	}

	var bid string
	var bidSize string
	if len(c.Result.Data[0].Bids) != 0 {
		bidSlice := c.Result.Data[0].Bids
		bid = bidSlice[0][0]
		bidSize = stringMultiply(bidSlice[0][1], bidSlice[0][2])
	}

	return MarketUpdate{
		Ask:     ask,
		AskSize: askSize,
		Bid:     bid,
		BidSize: bidSize,
	}
}

// Build the byte message payload for subscribing to a certain symbol's book data
func buildCryptoComSubscription(symbol string) subscription {
	c := []string{fmt.Sprintf("book.%s", symbol)}
	params := map[string][]string{
		"channels": c,
	}

	return subscription{
		Params: params,
		cryptoComMessage: cryptoComMessage{
			Id:     1,
			Method: "subscribe",
		},
	}
}

// Models

type cryptoComHeartbeat struct {
	cryptoComMessage
	Code int `json:"code"`
}

type subscription struct {
	cryptoComMessage
	Params map[string][]string `json:"params"`
}

type cryptoComMessage struct {
	Id     int    `json:"id"`
	Method string `json:"method"`
}

type cryptoComSubscriptionResponse struct {
	cryptoComMessage
	Code    int    `json:"code"`
	Channel string `json:"channel"`
}

type cryptoComBookMsg struct {
	cryptoComMessage
	Code   int             `json:"code"`
	Result cryptoComResult `json:"result"`
}

type cryptoComResult struct {
	Channel        string              `json:"book"`
	Subscription   string              `json:"subscription"`
	InstrumentName string              `json:"instrument_name"`
	Data           []cryptoComBookData `json:"data"`
}

type cryptoComBookData struct {
	Asks        [][]string `json:"asks"`
	Bids        [][]string `json:"bids"`
	LastUpdate  int        `json:"t"`
	MessageTime int        `json:"tt"`
}
