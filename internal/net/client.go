package net

import (
	"encoding/json"
	"net"
	"sync"
)

type Client struct {
	conn     net.Conn
	decoder  *json.Decoder
	encoder  *json.Encoder
	mu       sync.Mutex
	players  []PlayerState
	stopChan chan struct{}
}

func DialClient(addr string) (*Client, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	c := &Client{
		conn:     conn,
		decoder:  json.NewDecoder(conn),
		encoder:  json.NewEncoder(conn),
		stopChan: make(chan struct{}),
	}
	go c.receiveLoop()
	return c, nil
}

func (c *Client) receiveLoop() {
	for {
		select {
		case <-c.stopChan:
			return
		default:
		}
		var msg Message
		if err := c.decoder.Decode(&msg); err != nil {
			return
		}
		if msg.Type == "state" {
			c.mu.Lock()
			c.players = msg.Players
			c.mu.Unlock()
		}
	}
}

func (c *Client) SendState(p PlayerState) {
	c.mu.Lock()
	defer c.mu.Unlock()
	_ = c.encoder.Encode(Message{Type: "state", Player: &p})
}

func (c *Client) GetRemotePlayers() []PlayerState {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]PlayerState, len(c.players))
	copy(out, c.players)
	return out
}

func (c *Client) Stop() {
	select {
	case <-c.stopChan:
	default:
		close(c.stopChan)
	}
	c.conn.Close()
}
