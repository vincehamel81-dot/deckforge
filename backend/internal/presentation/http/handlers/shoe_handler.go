package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/vincehamel81-dot/deckforge/internal/application/commands"
	"github.com/vincehamel81-dot/deckforge/internal/domain/game"
	"github.com/vincehamel81-dot/deckforge/internal/domain/shoe"
	"github.com/vincehamel81-dot/deckforge/internal/presentation/http/middleware"
)

type ShoeHandler struct {
	games game.Repository
	shoes shoe.Repository
}

func NewShoeHandler(games game.Repository, shoes shoe.Repository) *ShoeHandler {
	return &ShoeHandler{games: games, shoes: shoes}
}

// CreateDeck generates and returns a deck UUID — no DB write.
// The UUID is passed to AddDeckToShoe.
func (h *ShoeHandler) CreateDeck(c *gin.Context) {
	c.JSON(http.StatusCreated, gin.H{
		"id":   uuid.New().String(),
		"note": "single-use — add to a game shoe via POST /games/:id/shoe/decks",
	})
}

type addDeckRequest struct {
	DeckID string `json:"deckId"`
}

func (h *ShoeHandler) AddDeckToShoe(c *gin.Context) {
	claims := middleware.ClaimsFromContext(c)
	gameID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid game id"})
		return
	}
	var req addDeckRequest
	_ = c.ShouldBindJSON(&req) // deckId is informational only

	dealerID, _ := uuid.Parse(claims.UserID)
	if err := commands.AddDeckToShoe(commands.AddDeckToShoeCommand{
		GameID:       gameID,
		DealerUserID: dealerID,
	}, h.games, h.shoes); err != nil {
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

	g, _ := h.games.FindByID(gameID)
	remaining, _ := h.shoes.UndealtCount(gameID)
	c.JSON(http.StatusOK, gin.H{
		"deckCount":      g.DeckCount,
		"totalCards":     g.TotalCards(),
		"remainingCards": remaining,
	})
}

func (h *ShoeHandler) ShuffleShoe(c *gin.Context) {
	claims := middleware.ClaimsFromContext(c)
	gameID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid game id"})
		return
	}
	dealerID, _ := uuid.Parse(claims.UserID)
	if err := commands.ShuffleShoe(commands.ShuffleShoeCommand{
		GameID:       gameID,
		DealerUserID: dealerID,
	}, h.games, h.shoes); err != nil {
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
	c.Status(http.StatusNoContent)
}

func (h *ShoeHandler) GetSuitCounts(c *gin.Context) {
	gameID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid game id"})
		return
	}
	counts, err := h.shoes.CountBySuit(gameID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"suits": counts})
}

func (h *ShoeHandler) GetCardCounts(c *gin.Context) {
	gameID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid game id"})
		return
	}
	counts, err := h.shoes.CountByCard(gameID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"cards": counts})
}
