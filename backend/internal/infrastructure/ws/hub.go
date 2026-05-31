package ws

import (
	"encoding/json"
	"sync"
	"time"

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

// DisconnectHandler is called after a player's WebSocket connection has been
// closed for longer than the configured disconnect timeout with no reconnect.
// gameID and userID are string UUIDs.
type DisconnectHandler func(gameID, userID string)

// Hub manages WebSocket clients grouped by game ID.
// One Hub instance is shared across the entire server lifetime.
type Hub struct {
	mu                 sync.RWMutex
	clients            map[string]map[*Client]struct{} // gameID → connected clients
	pendingDisconnects map[string]*time.Timer           // gameID:userID → timer
	disconnectTimeout  time.Duration
	onDisconnect       DisconnectHandler
}

// NewHub creates a Hub. If disconnectTimeout > 0 and onDisconnect is non-nil,
// players whose WS connection closes are auto-removed after the timeout unless
// they reconnect first.
func NewHub(disconnectTimeout time.Duration, onDisconnect DisconnectHandler) *Hub {
	return &Hub{
		clients:            make(map[string]map[*Client]struct{}),
		pendingDisconnects: make(map[string]*time.Timer),
		disconnectTimeout:  disconnectTimeout,
		onDisconnect:       onDisconnect,
	}
}

func (h *Hub) register(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.clients[c.gameID] == nil {
		h.clients[c.gameID] = make(map[*Client]struct{})
	}
	h.clients[c.gameID][c] = struct{}{}

	// Cancel any pending disconnect timer — the player reconnected in time.
	if c.userID != "" {
		key := c.gameID + ":" + c.userID
		if t, ok := h.pendingDisconnects[key]; ok {
			t.Stop()
			delete(h.pendingDisconnects, key)
		}
	}
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

	if c.userID == "" || h.onDisconnect == nil || h.disconnectTimeout == 0 {
		return
	}

	// If another tab/window for the same player is still connected, skip the timer.
	for client := range h.clients[c.gameID] {
		if client.userID == c.userID {
			return
		}
	}

	// Start the disconnect grace period.
	key := c.gameID + ":" + c.userID
	if t, ok := h.pendingDisconnects[key]; ok {
		t.Stop() // shouldn't exist, but be safe
	}
	gameID, userID := c.gameID, c.userID
	h.pendingDisconnects[key] = time.AfterFunc(h.disconnectTimeout, func() {
		h.mu.Lock()
		delete(h.pendingDisconnects, key)
		h.mu.Unlock()
		h.onDisconnect(gameID, userID)
	})
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
	userID string // from JWT — used to correlate with the player record on disconnect
	conn   *websocket.Conn
	send   chan []byte
}

func NewClient(hub *Hub, gameID, userID string, conn *websocket.Conn) *Client {
	return &Client{hub: hub, gameID: gameID, userID: userID, conn: conn, send: make(chan []byte, 32)}
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
