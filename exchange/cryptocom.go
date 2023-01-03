package exchange

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/gorilla/websocket"
	"github.com/johnwashburne/Crypto-Price-Aggregator/symbol"
)

// message method for heartbeat response
const heartbeatRequestMethod = "public/respond-heartbeat"

type CryptoCom struct {
	Updates chan MarketUpdate
	url     string
	name    string
	symbol  string
}

// create new Crypto.com struct
func NewCryptoCom(pair symbol.CurrencyPair) *CryptoCom {
	c := make(chan MarketUpdate, updateBufSize)

	return &CryptoCom{
		Updates: c,
		url:     "wss://uat-stream.3ona.co/v2/market",
		name:    "Crypto.com",
		symbol:  pair.CryptoCom,
	}
}

// Receive book data from Crypto.com, send any top of book updates
// over the Updates channel as a MarketUpdate struct
func (c *CryptoCom) Recv() {
	log.Printf("Connecting to %s\n", c.url)
	conn, _, err := websocket.DefaultDialer.Dial(c.url, nil)
	if err != nil {
		log.Println("Could not connect to Crypto.com")
		return
	}
	defer conn.Close()

	subscriptionMessage, err := buildCryptoComSubscription(c.symbol)
	conn.WriteMessage(websocket.TextMessage, subscriptionMessage)
	lastUpdate := MarketUpdate{}

	for {
		_, raw_msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("Crypto.com:", err)
			continue
		}

		var jsonMap map[string]any
		json.Unmarshal(raw_msg, &jsonMap)

		// Crypto.com requires a heartbeat message response every
		// 	 30 seconds to keep websocket connection alive
		if jsonMap["method"] == "public/heartbeat" {
			var h cryptoComHeartbeat
			json.Unmarshal(raw_msg, &h)
			b, err := buildCryptoComHeartbeat(h.Id)
			if err != nil {
				log.Println("Crypto.com heartbeat:", err)
				continue
			}
			conn.WriteMessage(websocket.TextMessage, b)
		} else if jsonMap["method"] == "subscribe" {
			var bookMsg cryptoComBookMsg
			json.Unmarshal(raw_msg, &bookMsg)

			update, err := parseCryptoComBookData(&bookMsg)
			if err != nil {
				// length of bid/ask is zero
				// so continue without update
				continue
			}

			if update != lastUpdate {
				c.Updates <- update
			}
			lastUpdate = update
		} else {
			log.Panic("unidentified message", jsonMap["method"])
		}
	}
}

// Name of data source
func (c *CryptoCom) Name() string {
	return c.name
}

// parse a Crypto.com book websocket message into our market update object
// best bid and ask, as well as volume for both
func parseCryptoComBookData(c *cryptoComBookMsg) (MarketUpdate, error) {
	if len(c.Result.Data) == 0 || len(c.Result.Data[0].Asks) == 0 || len(c.Result.Data[0].Asks) == 0 {
		return MarketUpdate{}, errors.New("no data")
	}
	askSlice := c.Result.Data[0].Asks
	bidSlice := c.Result.Data[0].Bids

	var ask string
	var askSize string
	ask = askSlice[0][0]
	askSize = stringMultiply(askSlice[0][1], askSlice[0][2])

	var bid string
	var bidSize string
	bid = bidSlice[0][0]
	bidSize = stringMultiply(bidSlice[0][1], bidSlice[0][2])

	return MarketUpdate{
		Ask:       ask,
		AskVolume: askSize,
		Bid:       bid,
		BidVolume: bidSize,
	}, nil
}

// Build the heatbeat response required by Crypto.com
func buildCryptoComHeartbeat(id int) ([]byte, error) {
	response := cryptoComHeartbeat{id, heartbeatRequestMethod}
	b, err := json.Marshal(response)
	if err != nil {
		return nil, err
	}

	return b, nil
}

// Build the byte message payload for subscribing to a certain symbol's book data
func buildCryptoComSubscription(symbol string) ([]byte, error) {
	params := make(map[string][]string)
	c := make([]string, 1)
	c[0] = fmt.Sprintf("book.%s", symbol)
	params["channels"] = c
	response := subscription{
		Id:     1,
		Method: "subscribe",
		Params: params,
	}
	b, err := json.Marshal(response)
	if err != nil {
		return nil, err
	}

	return b, nil
}

type cryptoComHeartbeat struct {
	Id     int    `json:"id"`
	Method string `json:"method"`
}

type subscription struct {
	Id     int                 `json:"id"`
	Method string              `json:"method"`
	Params map[string][]string `json:"params"`
}

type cryptoComBookMsg struct {
	Id     int             `json:"id"`
	Method string          `json:"method"`
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
