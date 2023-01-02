package exchange

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/gorilla/websocket"
)

// message method for heartbeat response
const heartbeatRequestMethod = "public/respond-heartbeat"

type CryptoCom struct {
	Updates chan MarketUpdate
	url     string
}

// create new Crypto.com struct
func NewCryptoCom(url string) *CryptoCom {
	c := make(chan MarketUpdate, 100)

	return &CryptoCom{
		Updates: c,
		url:     url,
	}
}

// Receive book data from Crypto.com, send any top of book updates
// over the Updates channel as a MarketUpdate struct
func (c *CryptoCom) Recv() error {
	log.Printf("Connecting to %s\n", c.url)
	conn, _, err := websocket.DefaultDialer.Dial(c.url, nil)
	if err != nil {
		return err
	}
	defer conn.Close()

	subscriptionMessage, err := buildSubscription("BTCUSD-PERP")
	conn.WriteMessage(websocket.TextMessage, subscriptionMessage)
	lastUpdate := MarketUpdate{}

	for {
		_, raw_msg, err := conn.ReadMessage()
		if err != nil {
			log.Fatal(err)
			return err
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
				return err
			}
			conn.WriteMessage(websocket.TextMessage, b)
		} else if jsonMap["method"] == "subscribe" {
			var bookMsg cryptoComBookMsg
			json.Unmarshal(raw_msg, &bookMsg)

			update, err := parseBookData(&bookMsg)
			if err != nil {
				log.Println(err)
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
	return "Gemini"
}

func parseBookData(c *cryptoComBookMsg) (MarketUpdate, error) {
	if len(c.Result.Data) == 0 {
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

func buildCryptoComHeartbeat(id int) ([]byte, error) {
	response := cryptoComHeartbeat{id, heartbeatRequestMethod}
	b, err := json.Marshal(response)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func buildSubscription(symbol string) ([]byte, error) {
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
