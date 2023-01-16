package exchange

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/johnwashburne/Crypto-Price-Aggregator/symbol"
	"github.com/johnwashburne/Crypto-Price-Aggregator/ws"
)

type Kucoin struct {
	updates chan MarketUpdate
	symbol  string
	name    string
	url     string
	valid   bool
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
		updates: c,
		symbol:  s,
		name:    fmt.Sprintf("Kucoin: %s", pair.Kucoin),
		url:     fmt.Sprintf("%s?token=%s", base, token),
		valid:   s != "",
	}
}

func (e *Kucoin) Recv() {
	// connect to websocket
	log.Printf("%s - Connecting to %s\n", e.name, e.url)
	conn := ws.New(e.url)
	err := conn.Connect()
	if err != nil {
		log.Println("Could not connect to", e.name)
		return
	}

	// welcome message
	var resp kucoinSubscriptionResponse
	conn.ReadJSON(&resp)

	conn.WriteJSON(kucoinSubscribe{
		Id:             1,
		Type:           "subscribe",
		Topic:          fmt.Sprintf("/spotMarket/level2Depth5:%s", e.symbol),
		PrivateChannel: false,
		Response:       true,
	})

	// ack message
	conn.ReadJSON(&resp)

	for {
		var message kucoinLevel2
		err := conn.ReadJSON(&message)
		if err != nil {
			log.Println(e.name, err)
			continue
		}

		e.updates <- MarketUpdate{
			Bid:     message.Data.Bids[0][0],
			BidSize: message.Data.Bids[0][1],
			Ask:     message.Data.Asks[0][0],
			AskSize: message.Data.Asks[0][1],
			Name:    e.name,
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
	Id             int    `json:"id"`
	Type           string `json:"type"`
	Topic          string `json:"topic"`
	PrivateChannel bool   `json:"privateChannel"`
	Response       bool   `json:"response"`
}

type kucoinSubscriptionResponse struct {
	Id   int    `json:"id"`
	Type string `json:"type"`
}

type kucoinLevel2 struct {
	Type    string           `json:"type"`
	Topic   string           `json:"topic"`
	Subject string           `json:"subject"`
	Data    kucoinLevel2Data `json:"data"`
}

type kucoinLevel2Data struct {
	Asks      [][]string `json:"asks"`
	Bids      [][]string `json:"bids"`
	Timestamp int        `json:"timestamp"`
}
