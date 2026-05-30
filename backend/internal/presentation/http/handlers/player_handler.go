package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/vincehamel81-dot/deckforge/internal/application/commands"
	"github.com/vincehamel81-dot/deckforge/internal/domain/game"
	"github.com/vincehamel81-dot/deckforge/internal/domain/player"
	"github.com/vincehamel81-dot/deckforge/internal/domain/shoe"
	"github.com/vincehamel81-dot/deckforge/internal/presentation/http/middleware"
)

type PlayerHandler struct {
	games   game.Repository
	players player.Repository
	shoes   shoe.Repository
}

func NewPlayerHandler(games game.Repository, players player.Repository, shoes shoe.Repository) *PlayerHandler {
	return &PlayerHandler{games: games, players: players, shoes: shoes}
}

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
	c.JSON(http.StatusCreated, gin.H{"player": p})
}

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
	c.Status(http.StatusNoContent)
}
