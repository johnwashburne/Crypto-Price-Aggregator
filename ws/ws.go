package ws

import (
	"log"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/gorilla/websocket"
)

type Client struct {
	url           string
	conn          *websocket.Conn
	backoff       backoff.BackOff
	onConnectFunc func(c *Client) error
}

func New(url string) *Client {
	return &Client{
		url:     url,
		backoff: backoff.WithMaxRetries(backoff.NewExponentialBackOff(), 10),
	}
}

func (c *Client) Connect() error {
	if c.conn != nil {
		return nil
	}

	return backoff.RetryNotify(c.connect(), c.backoff, nil)
}

func (c *Client) connect() func() error {
	return func() error {
		conn, _, err := websocket.DefaultDialer.Dial(c.url, nil)
		if err != nil {
			return err
		}

		conn.SetPongHandler(func(string) error {
			log.Println("Handling pong")
			return conn.SetReadDeadline(time.Now().Add(10 * time.Second))
		})

		c.conn = conn
		if c.onConnectFunc != nil {
			return c.onConnectFunc(c)
		}

		return nil
	}
}

func (c *Client) reconnect() error {
	c.conn.Close()
	c.conn = nil
	log.Println("reconnecting to", c.url)
	return backoff.RetryNotify(c.connect(), c.backoff, nil)
}

// specify a function to run on websocket connection and reconnection
func (c *Client) SetOnConnect(onConnect func(c *Client) error) {
	c.onConnectFunc = onConnect
}

func (c *Client) ReadJSON(v interface{}) error {
	if err := c.conn.ReadJSON(v); err != nil {
		if err := c.reconnect(); err != nil {
			return err
		}

		return c.conn.ReadJSON(v)
	}

	return nil
}

func (c *Client) WriteJSON(v interface{}) error {
	if err := c.conn.WriteJSON(v); err != nil {
		if err := c.reconnect(); err != nil {
			return err
		}

		return c.conn.WriteJSON(v)
	}

	return nil
}

func (c *Client) ReadMessage() (int, []byte, error) {
	messageType, p, err := c.conn.ReadMessage()
	if err != nil {
		if err := c.reconnect(); err != nil {
			return 0, nil, err
		}

		return c.conn.ReadMessage()
	}

	return messageType, p, err
}
