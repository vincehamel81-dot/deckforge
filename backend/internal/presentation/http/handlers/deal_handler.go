package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/vincehamel81-dot/deckforge/internal/application/commands"
	"github.com/vincehamel81-dot/deckforge/internal/application/queries"
	"github.com/vincehamel81-dot/deckforge/internal/domain/game"
	"github.com/vincehamel81-dot/deckforge/internal/domain/player"
	"github.com/vincehamel81-dot/deckforge/internal/domain/shoe"
	"github.com/vincehamel81-dot/deckforge/internal/domain/user"
	"github.com/vincehamel81-dot/deckforge/internal/presentation/http/middleware"
)

type DealHandler struct {
	games   game.Repository
	players player.Repository
	shoes   shoe.Repository
	users   user.Repository
}

func NewDealHandler(games game.Repository, players player.Repository, shoes shoe.Repository, users user.Repository) *DealHandler {
	return &DealHandler{games: games, players: players, shoes: shoes, users: users}
}

type dealRequest struct {
	Count int `json:"count" binding:"required,min=1"`
}

func (h *DealHandler) DealCards(c *gin.Context) {
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
	var req dealRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "count is required and must be ≥ 1"})
		return
	}
	dealerID, _ := uuid.Parse(claims.UserID)

	result, err := commands.DealCards(commands.DealCardsCommand{
		GameID:       gameID,
		DealerUserID: dealerID,
		PlayerID:     playerID,
		Count:        req.Count,
	}, h.games, h.shoes, h.players)
	if err != nil {
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
	c.JSON(http.StatusOK, gin.H{
		"dealtCount": result.DealtCount,
		"gameEnded":  result.GameEnded,
	})
}

func (h *DealHandler) GetPlayerHand(c *gin.Context) {
	claims := middleware.ClaimsFromContext(c)
	playerID, err := uuid.Parse(c.Param("pid"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid player id"})
		return
	}

	// Only the player themselves or an admin can see the hand.
	p, err := h.players.FindByID(playerID)
	if err != nil || p == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "player not found"})
		return
	}
	isOwner := p.UserID.String() == claims.UserID
	isAdmin := claims.Role == "admin"
	if !isOwner && !isAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "you can only view your own hand"})
		return
	}

	hand, err := queries.GetPlayerHand(playerID, h.shoes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"cards": hand})
}

func (h *DealHandler) GetLeaderboard(c *gin.Context) {
	gameID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid game id"})
		return
	}
	board, err := queries.GetLeaderboard(gameID, h.players, h.shoes, h.users)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"leaderboard": board})
}
