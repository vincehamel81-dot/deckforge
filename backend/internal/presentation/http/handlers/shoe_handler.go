package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/vincehamel81-dot/deckforge/internal/application/commands"
	"github.com/vincehamel81-dot/deckforge/internal/domain/game"
	"github.com/vincehamel81-dot/deckforge/internal/domain/shoe"
	ws "github.com/vincehamel81-dot/deckforge/internal/infrastructure/ws"
	"github.com/vincehamel81-dot/deckforge/internal/presentation/http/middleware"
)

type ShoeHandler struct {
	games game.Repository
	shoes shoe.Repository
	hub   *ws.Hub
}

func NewShoeHandler(games game.Repository, shoes shoe.Repository, hub *ws.Hub) *ShoeHandler {
	return &ShoeHandler{games: games, shoes: shoes, hub: hub}
}

// CreateDeck godoc
// @Summary Generate a deck UUID
// @Description Generates a fresh deck UUID without writing to the database. Pass this ID to POST /games/{id}/shoe/decks.
// @Tags shoe
// @Produce json
// @Success 201 {object} map[string]string "deck id"
// @Router /decks [post]
func (h *ShoeHandler) CreateDeck(c *gin.Context) {
	c.JSON(http.StatusCreated, gin.H{
		"id":   uuid.New().String(),
		"note": "single-use — add to a game shoe via POST /games/:id/shoe/decks",
	})
}

type addDeckRequest struct {
	DeckID string `json:"deckId" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// AddDeckToShoe godoc
// @Summary Add a deck to the shoe
// @Description Inserts 52 ShoeCard rows into the game's shoe and increments Game.deckCount. Only allowed while the game is WAITING.
// @Tags shoe
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "game UUID"
// @Param body body addDeckRequest true "deck id from POST /decks"
// @Success 200 {object} map[string]interface{} "deckCount + totalCards + remainingCards"
// @Failure 400 {object} map[string]string
// @Router /games/{id}/shoe/decks [post]
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

// ShuffleShoe godoc
// @Summary Shuffle the shoe
// @Description Randomly permutes all undealt cards using Fisher-Yates (no library shuffle). Can be called in WAITING or IN_PROGRESS. Returns 204 with no body.
// @Tags shoe
// @Security BearerAuth
// @Param id path string true "game UUID"
// @Success 204 "shuffled"
// @Failure 400 {object} map[string]string
// @Router /games/{id}/shoe/shuffle [post]
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
	h.hub.Broadcast(gameID.String(), ws.Message{Event: ws.EventShoeShuffled})
	c.Status(http.StatusNoContent)
}

// GetSuitCounts godoc
// @Summary Undealt cards per suit
// @Description Returns the count of remaining (undealt) cards for each suit in canonical order: hearts, spades, clubs, diamonds.
// @Tags shoe
// @Produce json
// @Security BearerAuth
// @Param id path string true "game UUID"
// @Success 200 {object} map[string]interface{} "suits array [{suit, count}]"
// @Router /games/{id}/shoe/suits [get]
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

// GetCardCounts godoc
// @Summary Undealt cards per suit and face value
// @Description Returns remaining card counts sorted by suit (hearts→spades→clubs→diamonds) then face value descending (King→Ace).
// @Tags shoe
// @Produce json
// @Security BearerAuth
// @Param id path string true "game UUID"
// @Success 200 {object} map[string]interface{} "cards array [{suit, value, count}]"
// @Router /games/{id}/shoe/cards [get]
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
