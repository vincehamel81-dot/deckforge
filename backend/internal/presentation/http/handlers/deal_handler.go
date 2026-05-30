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
	Count int `json:"count" binding:"required,min=1" example:"1"`
}

// DealCards godoc
// @Summary Deal cards to one player
// @Description Deals N cards from the shoe to a specific player. If the shoe has fewer than N cards remaining, all available cards are dealt (no error). Auto-ends the game if the shoe can no longer serve a full round.
// @Tags dealing
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id  path string true "game UUID"
// @Param pid path string true "player UUID"
// @Param body body dealRequest true "number of cards"
// @Success 200 {object} map[string]interface{} "dealtCount + gameEnded"
// @Failure 400 {object} map[string]string
// @Router /games/{id}/players/{pid}/deal [post]
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

// DealRound godoc
// @Summary Deal cards to all active players
// @Description Server-side atomic deal: deals N cards to every active player in seat order. Eliminates N sequential HTTP round-trips and race conditions. Auto-ends if shoe is exhausted.
// @Tags dealing
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id   path string true "game UUID"
// @Param body body dealRequest true "cards per player"
// @Success 200 {object} map[string]interface{} "totalDealt + gameEnded"
// @Failure 400 {object} map[string]string
// @Router /games/{id}/deal-round [post]
func (h *DealHandler) DealRound(c *gin.Context) {
	claims := middleware.ClaimsFromContext(c)
	gameID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid game id"})
		return
	}
	var req dealRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "count is required and must be ≥ 1"})
		return
	}
	dealerID, _ := uuid.Parse(claims.UserID)

	result, err := commands.DealRound(commands.DealRoundCommand{
		GameID:       gameID,
		DealerUserID: dealerID,
		Count:        req.Count,
	}, h.games, h.shoes, h.players)
	if err != nil {
		status := http.StatusBadRequest
		if err == commands.ErrGameNotFound {
			status = http.StatusNotFound
		}
		if err == commands.ErrForbidden {
			status = http.StatusForbidden
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"totalDealt": result.DealtCount, "gameEnded": result.GameEnded})
}

// GetPlayerHand godoc
// @Summary Get a player's hand
// @Description Returns the cards held by a player. Only the player themselves or an admin may view a hand.
// @Tags dealing
// @Produce json
// @Security BearerAuth
// @Param id  path string true "game UUID"
// @Param pid path string true "player UUID"
// @Success 200 {object} map[string]interface{} "cards array"
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /games/{id}/players/{pid}/hand [get]
func (h *DealHandler) GetPlayerHand(c *gin.Context) {
	claims := middleware.ClaimsFromContext(c)
	playerID, err := uuid.Parse(c.Param("pid"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid player id"})
		return
	}

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

// GetLeaderboard godoc
// @Summary Get the leaderboard
// @Description Returns all active players sorted by total hand value descending. Ties broken by seat order ascending. Scores are intentionally public during play — cards remain private.
// @Tags players
// @Produce json
// @Security BearerAuth
// @Param id path string true "game UUID"
// @Success 200 {object} map[string]interface{} "leaderboard array [{playerId, userId, username, seatOrder, handValue, cardCount}]"
// @Router /games/{id}/players [get]
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
