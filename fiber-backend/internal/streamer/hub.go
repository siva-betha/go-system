package streamer

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gofiber/contrib/v3/websocket"
	"github.com/gofiber/fiber/v3"
)

// Client represents a single connected browser client
type Client struct {
	hub  *StreamHub
	conn *websocket.Conn
	send chan []byte // Outgoing message queue

	// Subscriptions: MachineID -> []ChamberID
	subs map[string][]string
	mu   sync.RWMutex
}

func NewClient(hub *StreamHub, conn *websocket.Conn) *Client {
	return &Client{
		hub:  hub,
		conn: conn,
		send: make(chan []byte, 256),
		subs: make(map[string][]string),
	}
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	for {
		var msg ClientMessage
		err := c.conn.ReadJSON(&msg)
		if err != nil {
			break
		}

		switch msg.Type {
		case MsgTypeSubscribe:
			c.mu.Lock()
			if c.subs == nil {
				c.subs = make(map[string][]string)
			}
			c.subs[msg.MachineID] = append(c.subs[msg.MachineID], msg.ChamberID)
			c.mu.Unlock()
		case MsgTypeUnsub:
			// Implementation for unsubscribe
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// StreamHub handles broadcasting PLC data to subscribed clients
type StreamHub struct {
	clients    map[*Client]bool
	broadcast  chan BroadcastMsg
	register   chan *Client
	unregister chan *Client

	mu sync.RWMutex
}

func NewHub() *StreamHub {
	return &StreamHub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan BroadcastMsg, 1000),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *StreamHub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			log.Printf("Client registered. Total: %d", len(h.clients))

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()
			log.Printf("Client unregistered. Total: %d", len(h.clients))

		case msg := <-h.broadcast:
			h.fanOut(msg)
		}
	}
}

func (h *StreamHub) fanOut(msg BroadcastMsg) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// Optimized: JSON encode once for all clients
	// In a real high-load system, we'd use a more efficient binary format
	// or buffer/throttle for slow clients.
	for client := range h.clients {
		client.mu.RLock()
		// Check if client is subscribed to this machine/chamber
		isSubscribed := false
		if chambers, ok := client.subs[msg.MachineID]; ok {
			for _, cid := range chambers {
				if cid == "" || cid == msg.ChamberID {
					isSubscribed = true
					break
				}
			}
		}
		client.mu.RUnlock()

		if isSubscribed {
			// Non-blocking send to avoid stalling hub
			select {
			case client.send <- h.serialize(msg):
			default:
				// Drop message if client buffer is full (backpressure)
			}
		}
	}
}

func (h *StreamHub) serialize(msg BroadcastMsg) []byte {
	b, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Failed to marshal broadcast msg: %v", err)
		return nil
	}
	return b
}

// Broadcast sends data to the h.broadcast channel
func (h *StreamHub) Broadcast(msg BroadcastMsg) {
	select {
	case h.broadcast <- msg:
	default:
		// Drop if hub is overloaded
	}
}

// NewHandler returns a Fiber handler for WebSocket upgrades
func (h *StreamHub) NewHandler() fiber.Handler {
	return websocket.New(func(c *websocket.Conn) {
		client := NewClient(h, c)
		h.register <- client

		go client.writePump()
		client.readPump()
	})
}
