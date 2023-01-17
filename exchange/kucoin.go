package exchange

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/johnwashburne/Crypto-Price-Aggregator/symbol"
	"github.com/johnwashburne/Crypto-Price-Aggregator/ws"
)

type Kucoin struct {
	updates      chan MarketUpdate
	symbol       string
	name         string
	url          string
	valid        bool
	pingInterval int
}

func NewKucoin(pair symbol.CurrencyPair) *Kucoin {
	c := make(chan MarketUpdate, updateBufSize)
	s := pair.Kucoin

	resp, err := http.Post("https://api.kucoin.com/api/v1/bullet-public", "", nil)
	if err != nil {
		log.Panic("Could not generate Kucoin Websocket URL", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Panic("Could not generate Kucoin Websocket URL", err)
	}

	var httpResponse kucoinHttpResponse
	json.Unmarshal(body, &httpResponse)

	base := httpResponse.Data.InstanceServers[0].Endpoint
	token := httpResponse.Data.Token

	return &Kucoin{
		updates:      c,
		symbol:       s,
		name:         fmt.Sprintf("Kucoin: %s", pair.Kucoin),
		url:          fmt.Sprintf("%s?token=%s", base, token),
		valid:        s != "",
		pingInterval: httpResponse.Data.InstanceServers[0].PingInterval,
	}
}

func (e *Kucoin) Recv() {
	// connect to websocket
	log.Printf("%s - Connecting to %s\n", e.name, e.url)
	conn := ws.New(e.url)

	conn.SetOnConnect(func(c *ws.Client) error {
		// welcome message
		var welcomeMessage kucoinMessage
		if err := c.ReadJSON(&welcomeMessage); err != nil {
			return err
		}

		err := c.WriteJSON(kucoinSubscribe{
			kucoinMessage: kucoinMessage{
				Type: "subscribe",
				Id:   "1",
			},
			Topic:          fmt.Sprintf("/market/ticker:%s", e.symbol),
			PrivateChannel: false,
			Response:       true,
		})
		if err != nil {
			return err
		}

		var ackMessage kucoinMessage
		if err := c.ReadJSON(&ackMessage); err != nil {
			return err
		}

		return nil
	})

	err := conn.Connect()
	if err != nil {
		log.Println("Could not connect to", e.name)
		return
	}

	ticker := time.NewTicker(time.Duration(e.pingInterval) * time.Millisecond)
	lastUpdate := MarketUpdate{}
	for {
		select {
		case <-ticker.C:
			conn.WriteJSON(kucoinMessage{
				Id:   "1",
				Type: "ping",
			})
		default:
			_, rawMessage, err := conn.ReadMessage()
			if err != nil {
				log.Println(e.name, "Could not read")
			}

			var message kucoinMessage
			json.Unmarshal(rawMessage, &message)
			if message.Type == "pong" {
				continue
			} else if message.Type == "message" {
				var tickerMessage kucoinTickerMessage
				json.Unmarshal(rawMessage, &tickerMessage)
				update := MarketUpdate{
					Ask:     tickerMessage.Data.BestAsk,
					AskSize: tickerMessage.Data.BestAskSize,
					Bid:     tickerMessage.Data.BestBid,
					BidSize: tickerMessage.Data.BestBidSize,
					Name:    e.name,
				}

				if update != lastUpdate {
					e.updates <- update
				}
				lastUpdate = update
			} else {
				log.Println(e.name, "unknown message", string(rawMessage))
			}
		}
	}
}

func (e *Kucoin) Updates() chan MarketUpdate {
	return e.updates
}

func (e *Kucoin) Valid() bool {
	return e.valid
}

func (e *Kucoin) Name() string {
	return e.name
}

type kucoinHttpResponse struct {
	Code string                      `json:"code"`
	Data kucoinWebsocketHttpResponse `json:"data"`
}

type kucoinWebsocketHttpResponse struct {
	InstanceServers []kucoinInstanceServer `json:"instanceServers"`
	Token           string                 `json:"token"`
}

type kucoinInstanceServer struct {
	Endpoint     string `json:"endpoint"`
	Protocol     string `json:"protocol"`
	Encrypt      bool   `json:"encrypt"`
	PingInterval int    `json:"pingInterval"`
	PingTimeout  int    `json:"pingTimeout"`
}

type kucoinSubscribe struct {
	kucoinMessage
	Topic          string `json:"topic"`
	PrivateChannel bool   `json:"privateChannel"`
	Response       bool   `json:"response"`
}

type kucoinMessage struct {
	Type string `json:"type"`
	Id   string `json:"id,omitempty"`
}

type kucoinTickerMessage struct {
	kucoinMessage
	Topic   string           `json:"topic"`
	Subject string           `json:"subject"`
	Data    kucoinTickerData `json:"data"`
}

type kucoinTickerData struct {
	Sequence    string `json:"sequence"`
	Price       string `json:"price"`
	Size        string `json:"size"`
	BestAsk     string `json:"bestAsk"`
	BestAskSize string `json:"bestAskSize"`
	BestBid     string `json:"bestBid"`
	BestBidSize string `json:"bestBidSize"`
}
