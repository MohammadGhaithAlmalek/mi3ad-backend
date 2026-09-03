package ws

import (
	"github.com/gorilla/websocket"
)

type Client struct {
	userID string
	conn   *websocket.Conn
	hub    *Hub
	outbox chan any
}

func newClient(userID string, conn *websocket.Conn, hub *Hub) *Client {
	c := &Client{
		userID: userID,
		conn:   conn,
		hub:    hub,
		outbox: make(chan any, 16),
	}
	go c.writeLoop()
	return c
}

func (c *Client) send(event any) {
	select {
	case c.outbox <- event:
	default:
		// outbox full (client not draining) - drop rather than block the hub
	}
}

func (c *Client) writeLoop() {
	defer c.conn.Close()
	for event := range c.outbox {
		if err := c.conn.WriteJSON(event); err != nil {
			c.hub.unregister(c.userID)
			return
		}
	}
}

// readLoop just keeps the connection alive and cleans up on disconnect.
// This project doesn't need the client to send anything back over the socket.
func (c *Client) readLoop() {
	defer func() {
		c.hub.unregister(c.userID)
		close(c.outbox)
	}()
	for {
		if _, _, err := c.conn.ReadMessage(); err != nil {
			return
		}
	}
}
