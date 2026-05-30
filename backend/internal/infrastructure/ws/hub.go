package ws

import (
	"encoding/json"
	"sync"

	"github.com/gorilla/websocket"
)

// EventType identifies the kind of game event being pushed to clients.
type EventType string

const (
	EventGameStarted  EventType = "game_started"
	EventGameEnded    EventType = "game_ended"
	EventCardsDealt   EventType = "cards_dealt"   // payload: {totalDealt} — never card values
	EventPlayerJoined EventType = "player_joined"
	EventPlayerLeft   EventType = "player_left"
	EventShoeShuffled EventType = "shoe_shuffled"
)

// Message is the JSON envelope sent to every connected client.
type Message struct {
	Event   EventType   `json:"event"`
	Payload interface{} `json:"payload,omitempty"`
}

// Hub manages WebSocket clients grouped by game ID.
// One Hub instance is shared across the entire server lifetime.
type Hub struct {
	mu      sync.RWMutex
	clients map[string]map[*Client]struct{} // gameID → connected clients
}

func NewHub() *Hub {
	return &Hub{clients: make(map[string]map[*Client]struct{})}
}

func (h *Hub) register(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.clients[c.gameID] == nil {
		h.clients[c.gameID] = make(map[*Client]struct{})
	}
	h.clients[c.gameID][c] = struct{}{}
}

func (h *Hub) deregister(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if group, ok := h.clients[c.gameID]; ok {
		delete(group, c)
		if len(group) == 0 {
			delete(h.clients, c.gameID)
		}
	}
	close(c.send)
}

// Broadcast sends a message to every client connected to the given game.
// Safe to call from any goroutine.
func (h *Hub) Broadcast(gameID string, msg Message) {
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	for c := range h.clients[gameID] {
		select {
		case c.send <- data:
		default:
			// Client is too slow — drop the message; its read pump will close it shortly.
		}
	}
}

// Client represents one connected WebSocket player.
type Client struct {
	hub    *Hub
	gameID string
	conn   *websocket.Conn
	send   chan []byte
}

func NewClient(hub *Hub, gameID string, conn *websocket.Conn) *Client {
	return &Client{hub: hub, gameID: gameID, conn: conn, send: make(chan []byte, 32)}
}

// Serve registers the client with the hub and starts its read/write pumps.
// Blocks until the connection closes.
func (c *Client) Serve() {
	c.hub.register(c)
	go c.writePump()
	c.readPump() // blocks
}

// writePump forwards messages from the send channel to the WebSocket.
func (c *Client) writePump() {
	defer c.conn.Close()
	for msg := range c.send {
		if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			return
		}
	}
}

// readPump reads from the WebSocket solely to detect disconnection.
func (c *Client) readPump() {
	defer func() {
		c.hub.deregister(c)
		c.conn.Close()
	}()
	c.conn.SetReadLimit(512)
	for {
		if _, _, err := c.conn.ReadMessage(); err != nil {
			break
		}
	}
}
