package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/vincehamel81-dot/deckforge/internal/domain/game"
	jwtpkg "github.com/vincehamel81-dot/deckforge/internal/infrastructure/auth"
	ws "github.com/vincehamel81-dot/deckforge/internal/infrastructure/ws"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Origin validation is handled by the CORS middleware on the HTTP layer.
	CheckOrigin: func(r *http.Request) bool { return true },
}

type WSHandler struct {
	hub       *ws.Hub
	games     game.Repository
	jwtSecret string
}

func NewWSHandler(hub *ws.Hub, games game.Repository, jwtSecret string) *WSHandler {
	return &WSHandler{hub: hub, games: games, jwtSecret: jwtSecret}
}

// ServeWS godoc
// @Summary Open a WebSocket connection for live game events
// @Description Upgrades the connection to WebSocket. Pass the JWT as ?token=<jwt>
// @Description because browser WebSocket APIs cannot send custom headers.
// @Description Events pushed: game_started, game_ended, cards_dealt, player_joined, player_left, shoe_shuffled.
// @Tags realtime
// @Param id    path  string true "game UUID"
// @Param token query string true "JWT"
// @Router /games/{id}/ws [get]
func (h *WSHandler) ServeWS(c *gin.Context) {
	// JWT from query param — WebSocket API cannot set headers.
	token := c.Query("token")
	if _, err := jwtpkg.Validate(token, h.jwtSecret); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	gameID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid game id"})
		return
	}

	g, err := h.games.FindByID(gameID)
	if err != nil || g == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "game not found"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	client := ws.NewClient(h.hub, gameID.String(), conn)
	go client.Serve()
}
