package queries

import (
	"github.com/google/uuid"
	"github.com/vincehamel81-dot/deckforge/internal/domain/game"
	"github.com/vincehamel81-dot/deckforge/internal/domain/shoe"
)

type GameDetail struct {
	Game           *game.Game
	TotalCards     int
	RemainingCards int
}

func GetGame(id uuid.UUID, games game.Repository, shoes shoe.Repository) (*GameDetail, error) {
	g, err := games.FindByID(id)
	if err != nil {
		return nil, err
	}
	if g == nil {
		return nil, nil
	}
	remaining, err := shoes.UndealtCount(id)
	if err != nil {
		return nil, err
	}
	return &GameDetail{
		Game:           g,
		TotalCards:     g.TotalCards(),
		RemainingCards: remaining,
	}, nil
}

func ListGames(statusFilter *game.Status, games game.Repository) ([]*game.Game, error) {
	return games.FindAll(statusFilter)
}
