package queries

import (
	"github.com/google/uuid"
	"github.com/vincehamel81-dot/deckforge/internal/domain/game"
	"github.com/vincehamel81-dot/deckforge/internal/domain/player"
	"github.com/vincehamel81-dot/deckforge/internal/domain/shoe"
	"github.com/vincehamel81-dot/deckforge/internal/domain/user"
)


type GameDetail struct {
	Game           *game.Game
	TotalCards     int
	RemainingCards int
}

// GameSummary enriches a Game with player count, dealer username, and remaining shoe
// cards for the lobby list. RemainingCards allows the frontend to gate the join button
// without an extra API call.
type GameSummary struct {
	*game.Game
	PlayerCount    int    `json:"playerCount"`
	DealerUsername string `json:"dealerUsername"`
	RemainingCards int    `json:"remainingCards"`
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

func ListGames(statusFilter *game.Status, games game.Repository, players player.Repository, shoes shoe.Repository, users user.Repository) ([]*GameSummary, error) {
	gamesList, err := games.FindAll(statusFilter)
	if err != nil {
		return nil, err
	}
	summaries := make([]*GameSummary, 0, len(gamesList))
	for _, g := range gamesList {
		count, err := players.CountActive(g.ID)
		if err != nil {
			return nil, err
		}
		remaining, err := shoes.UndealtCount(g.ID)
		if err != nil {
			return nil, err
		}
		u, err := users.FindByID(g.DealerUserID)
		if err != nil {
			return nil, err
		}
		dealerUsername := ""
		if u != nil {
			dealerUsername = u.Username
		}
		summaries = append(summaries, &GameSummary{
			Game:           g,
			PlayerCount:    count,
			DealerUsername: dealerUsername,
			RemainingCards: remaining,
		})
	}
	return summaries, nil
}
