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
	"github.com/vincehamel81-dot/deckforge/internal/presentation/http/middleware"
)

type GameHandler struct {
	games   game.Repository
	players player.Repository
	shoes   shoe.Repository
}

func NewGameHandler(games game.Repository, players player.Repository, shoes shoe.Repository) *GameHandler {
	return &GameHandler{games: games, players: players, shoes: shoes}
}

type createGameRequest struct {
	DeckCount  int `json:"deckCount" binding:"required,min=1,max=8"`
	MinPlayers int `json:"minPlayers"`
	MaxPlayers int `json:"maxPlayers"`
}

func (h *GameHandler) CreateGame(c *gin.Context) {
	claims := middleware.ClaimsFromContext(c)
	var req createGameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.MinPlayers == 0 {
		req.MinPlayers = 2
	}
	if req.MaxPlayers == 0 {
		req.MaxPlayers = 8
	}

	dealerID, _ := uuid.Parse(claims.UserID)
	result, err := commands.CreateGame(commands.CreateGameCommand{
		DealerUserID: dealerID,
		DeckCount:    req.DeckCount,
		MinPlayers:   req.MinPlayers,
		MaxPlayers:   req.MaxPlayers,
	}, h.games, h.players)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"game": result.Game, "player": result.Player})
}

func (h *GameHandler) ListGames(c *gin.Context) {
	statusStr := c.Query("status")
	var filter *game.Status
	if statusStr != "" {
		s := game.Status(statusStr)
		filter = &s
	}
	gamesList, err := queries.ListGames(filter, h.games)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"games": gamesList})
}

func (h *GameHandler) GetGame(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid game id"})
		return
	}
	detail, err := queries.GetGame(id, h.games, h.shoes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	if detail == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "game not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"game":           detail.Game,
		"totalCards":     detail.TotalCards,
		"remainingCards": detail.RemainingCards,
	})
}

type startGameRequest struct {
	InitialDealCount int `json:"initialDealCount"`
}

func (h *GameHandler) StartGame(c *gin.Context) {
	claims := middleware.ClaimsFromContext(c)
	id, _ := uuid.Parse(c.Param("id"))
	var req startGameRequest
	_ = c.ShouldBindJSON(&req)

	dealerID, _ := uuid.Parse(claims.UserID)
	g, err := commands.StartGame(commands.StartGameCommand{
		GameID:           id,
		DealerUserID:     dealerID,
		InitialDealCount: req.InitialDealCount,
	}, h.games, h.players, h.shoes)
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
	c.JSON(http.StatusOK, gin.H{"game": g})
}

func (h *GameHandler) EndGame(c *gin.Context) {
	claims := middleware.ClaimsFromContext(c)
	id, _ := uuid.Parse(c.Param("id"))
	dealerID, _ := uuid.Parse(claims.UserID)

	g, err := commands.EndGame(commands.EndGameCommand{GameID: id, DealerUserID: dealerID}, h.games)
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
	c.JSON(http.StatusOK, gin.H{"game": g})
}

func (h *GameHandler) DeleteGame(c *gin.Context) {
	claims := middleware.ClaimsFromContext(c)
	id, _ := uuid.Parse(c.Param("id"))
	dealerID, _ := uuid.Parse(claims.UserID)

	if err := commands.DeleteGame(commands.DeleteGameCommand{GameID: id, DealerUserID: dealerID}, h.games); err != nil {
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
