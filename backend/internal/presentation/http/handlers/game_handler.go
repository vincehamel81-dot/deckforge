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
	ws "github.com/vincehamel81-dot/deckforge/internal/infrastructure/ws"
	"github.com/vincehamel81-dot/deckforge/internal/presentation/http/middleware"
)

type GameHandler struct {
	games   game.Repository
	players player.Repository
	shoes   shoe.Repository
	users   user.Repository
	hub     *ws.Hub
	autoEnd bool
}

func NewGameHandler(games game.Repository, players player.Repository, shoes shoe.Repository, users user.Repository, hub *ws.Hub, autoEnd bool) *GameHandler {
	return &GameHandler{games: games, players: players, shoes: shoes, users: users, hub: hub, autoEnd: autoEnd}
}

type createGameRequest struct {
	DeckCount  int `json:"deckCount"  binding:"required,min=1,max=8" example:"2"`
	MinPlayers int `json:"minPlayers" example:"2"`
	MaxPlayers int `json:"maxPlayers" example:"8"`
}

// CreateGame godoc
// @Summary Create a game
// @Description Creates a new game. The caller becomes the dealer and auto-joins as player seat 0. Initial shoe cards are added automatically for the requested deck count.
// @Tags games
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body createGameRequest true "game config"
// @Success 201 {object} map[string]interface{} "game + player"
// @Failure 400 {object} map[string]string
// @Router /games [post]
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

// ListGames godoc
// @Summary List games
// @Description Returns all non-finished games, optionally filtered by status.
// @Tags games
// @Produce json
// @Security BearerAuth
// @Param status query string false "WAITING or IN_PROGRESS"
// @Success 200 {object} map[string]interface{} "games array with playerCount and dealerUsername"
// @Router /games [get]
func (h *GameHandler) ListGames(c *gin.Context) {
	statusStr := c.Query("status")
	var filter *game.Status
	if statusStr != "" {
		s := game.Status(statusStr)
		filter = &s
	}
	summaries, err := queries.ListGames(filter, h.games, h.players, h.users)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"games": summaries})
}

// GetGame godoc
// @Summary Get game detail
// @Description Returns a game's current state, shoe status, and dealer username.
// @Tags games
// @Produce json
// @Security BearerAuth
// @Param id path string true "game UUID"
// @Success 200 {object} map[string]interface{} "game + totalCards + remainingCards + dealerUsername"
// @Failure 404 {object} map[string]string
// @Router /games/{id} [get]
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
	dealerUsername := ""
	if u, _ := h.users.FindByID(detail.Game.DealerUserID); u != nil {
		dealerUsername = u.Username
	}
	c.JSON(http.StatusOK, gin.H{
		"game":           detail.Game,
		"totalCards":     detail.TotalCards,
		"remainingCards": detail.RemainingCards,
		"dealerUsername": dealerUsername,
	})
}

type startGameRequest struct {
	InitialDealCount int `json:"initialDealCount" example:"2"`
}

// StartGame godoc
// @Summary Start a game
// @Description Transitions the game from WAITING to IN_PROGRESS and optionally deals an initial hand to each player.
// @Tags games
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "game UUID"
// @Param body body startGameRequest false "optional initial deal count per player"
// @Success 200 {object} map[string]interface{} "updated game"
// @Failure 400 {object} map[string]string
// @Router /games/{id}/start [post]
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
		AutoEnd:          h.autoEnd,
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
	h.hub.Broadcast(id.String(), ws.Message{Event: ws.EventGameStarted})
	c.JSON(http.StatusOK, gin.H{"game": g})
}

// EndGame godoc
// @Summary End a game
// @Description Transitions an IN_PROGRESS game to FINISHED and returns the final state.
// @Tags games
// @Produce json
// @Security BearerAuth
// @Param id path string true "game UUID"
// @Success 200 {object} map[string]interface{} "updated game"
// @Failure 400 {object} map[string]string
// @Router /games/{id}/end [post]
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
	h.hub.Broadcast(id.String(), ws.Message{Event: ws.EventGameEnded})
	c.JSON(http.StatusOK, gin.H{"game": g})
}

// DeleteGame godoc
// @Summary Delete a game
// @Description Permanently removes a game. Allowed for the dealer of that game or any admin.
// @Tags games
// @Produce json
// @Security BearerAuth
// @Param id path string true "game UUID"
// @Success 204 "no content"
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /games/{id} [delete]
func (h *GameHandler) DeleteGame(c *gin.Context) {
	claims := middleware.ClaimsFromContext(c)
	id, _ := uuid.Parse(c.Param("id"))
	dealerID, _ := uuid.Parse(claims.UserID)

	if err := commands.DeleteGame(commands.DeleteGameCommand{GameID: id, DealerUserID: dealerID, IsAdmin: claims.Role == "admin"}, h.games); err != nil {
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
	// Push all connected players off the table immediately.
	h.hub.Broadcast(id.String(), ws.Message{
		Event:   ws.EventGameEnded,
		Payload: map[string]interface{}{"reason": "deleted"},
	})
	c.Status(http.StatusNoContent)
}
