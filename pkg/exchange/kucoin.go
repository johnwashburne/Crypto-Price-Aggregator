// Aggreagate top of book updates from a currency pair listed on Kucoin

package exchange

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/johnwashburne/Crypto-Price-Aggregator/pkg/logger"
	"github.com/johnwashburne/Crypto-Price-Aggregator/pkg/symbol"
	"github.com/johnwashburne/Crypto-Price-Aggregator/pkg/ws"
)

type Kucoin struct {
	updates      chan MarketUpdate
	symbol       string
	name         string
	url          string
	valid        bool
	pingInterval int
	logger       *logger.Logger
}

func NewKucoin(pair symbol.CurrencyPair) *Kucoin {
	c := make(chan MarketUpdate, updateBufSize)
	s := pair.Kucoin
	name := fmt.Sprintf("Kucoin: %s", pair.Kucoin)
	logger := logger.Named(name)

	k := &Kucoin{
		updates: c,
		symbol:  s,
		name:    name,
		valid:   s != "",
		logger:  logger,
	}

	if err := k.applyForInstanceServer(); err != nil {
		k.valid = false
		logger.Warn(err)
	}

	return k
}

func (e *Kucoin) Recv() {
	// connect to websocket
	e.logger.Debug("connecting to socket")
	conn := ws.New(e.url)

	conn.SetOnConnect(func(c *ws.Client) error {
		// welcome message
		var welcomeMessage kucoinMessage
		if err := c.ReadJSON(&welcomeMessage); err != nil {
			// if error in connection, apply for new token
			if err := e.applyForInstanceServer(); err != nil {
				e.logger.Warn(err)
				return err
			}

			e.logger.Warn(err)
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
			// if error in connection, apply for new token
			e.applyForInstanceServer()
			return err
		}

		var ackMessage kucoinMessage
		if err := c.ReadJSON(&ackMessage); err != nil {
			// if error in connection, apply for new token
			e.applyForInstanceServer()
			return err
		}

		return nil
	})

	if err := conn.Connect(); err != nil {
		e.logger.Warn("could not connect to socket, RETURNING")
		return
	}
	e.logger.Debug("connected to socket")

	ticker := time.NewTicker(time.Duration(e.pingInterval) * time.Millisecond)
	lastUpdate := MarketUpdate{}
	for {
		select {
		case <-ticker.C:
			e.logger.Debug("sending ping")
			conn.WriteJSON(kucoinMessage{
				Id:   "1",
				Type: "ping",
			})
		default:
			_, rawMessage, err := conn.ReadMessage()
			if err != nil {
				e.logger.Warn("Could not read message", err)
				continue
			}

			var message kucoinMessage
			json.Unmarshal(rawMessage, &message)
			if message.Type == "pong" {
				e.logger.Debug("pong received")
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
				e.logger.Warn("unknown message", string(rawMessage))
			}
		}
	}
}

func (e *Kucoin) applyForInstanceServer() error {
	resp, err := http.Post("https://api.kucoin.com/api/v1/bullet-public", "", nil)
	e.logger.Info("applying for instance server token")
	if err != nil {
		e.logger.Warn("Could not generate Kucoin Websocket URL", err)
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		e.logger.Warn("Could not generate Kucoin Websocket URL", err)
		return err
	}

	var httpResponse kucoinHttpResponse
	json.Unmarshal(body, &httpResponse)

	base := httpResponse.Data.InstanceServers[0].Endpoint
	token := httpResponse.Data.Token

	e.url = fmt.Sprintf("%s?token=%s", base, token)
	e.pingInterval = httpResponse.Data.InstanceServers[0].PingInterval
	e.logger.Info("granted instance server token ", token)
	return nil
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
