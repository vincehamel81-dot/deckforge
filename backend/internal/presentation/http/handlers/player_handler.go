package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/vincehamel81-dot/deckforge/internal/application/commands"
	"github.com/vincehamel81-dot/deckforge/internal/domain/game"
	"github.com/vincehamel81-dot/deckforge/internal/domain/player"
	"github.com/vincehamel81-dot/deckforge/internal/domain/shoe"
	ws "github.com/vincehamel81-dot/deckforge/internal/infrastructure/ws"
	"github.com/vincehamel81-dot/deckforge/internal/presentation/http/middleware"
)

type PlayerHandler struct {
	games   game.Repository
	players player.Repository
	shoes   shoe.Repository
	hub     *ws.Hub
}

func NewPlayerHandler(games game.Repository, players player.Repository, shoes shoe.Repository, hub *ws.Hub) *PlayerHandler {
	return &PlayerHandler{games: games, players: players, shoes: shoes, hub: hub}
}

// JoinGame godoc
// @Summary Join a game
// @Description Adds the authenticated user as a player. If the game is IN_PROGRESS a catch-up deal is applied automatically.
// @Tags players
// @Produce json
// @Security BearerAuth
// @Param id path string true "game UUID"
// @Success 201 {object} map[string]interface{} "player"
// @Failure 400 {object} map[string]string "already in a game, game full, etc."
// @Failure 404 {object} map[string]string "game not found"
// @Router /games/{id}/players [post]
func (h *PlayerHandler) JoinGame(c *gin.Context) {
	claims := middleware.ClaimsFromContext(c)
	gameID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid game id"})
		return
	}
	userID, _ := uuid.Parse(claims.UserID)

	p, err := commands.AddPlayer(commands.AddPlayerCommand{
		GameID: gameID,
		UserID: userID,
	}, h.games, h.players, h.shoes)
	if err != nil {
		status := http.StatusBadRequest
		if err == commands.ErrGameNotFound {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}
	h.hub.Broadcast(gameID.String(), ws.Message{Event: ws.EventPlayerJoined})
	c.JSON(http.StatusCreated, gin.H{"player": p})
}

// LeaveGame godoc
// @Summary Remove a player from a game
// @Description The player themselves, the dealer, or an admin may remove a player. All held cards are returned to the shoe.
// @Tags players
// @Security BearerAuth
// @Param id  path string true "game UUID"
// @Param pid path string true "player UUID"
// @Success 204 "removed"
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /games/{id}/players/{pid} [delete]
func (h *PlayerHandler) LeaveGame(c *gin.Context) {
	claims := middleware.ClaimsFromContext(c)
	gameID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid game id"})
		return
	}
	playerID, err := uuid.Parse(c.Param("pid"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid player id"})
		return
	}
	requesterID, _ := uuid.Parse(claims.UserID)

	if err := commands.RemovePlayer(commands.RemovePlayerCommand{
		GameID:          gameID,
		PlayerID:        playerID,
		RequesterUserID: requesterID,
		IsAdmin:         claims.Role == "admin",
	}, h.games, h.players, h.shoes); err != nil {
		status := http.StatusBadRequest
		if err == commands.ErrGameNotFound || err == commands.ErrPlayerNotFound {
			status = http.StatusNotFound
		}
		if err == commands.ErrForbidden {
			status = http.StatusForbidden
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}
	h.hub.Broadcast(gameID.String(), ws.Message{Event: ws.EventPlayerLeft})
	c.Status(http.StatusNoContent)
}
