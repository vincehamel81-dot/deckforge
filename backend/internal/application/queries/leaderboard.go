package queries

import (
	"sort"

	"github.com/google/uuid"
	"github.com/vincehamel81-dot/deckforge/internal/domain/player"
	"github.com/vincehamel81-dot/deckforge/internal/domain/shoe"
)

type LeaderboardEntry struct {
	PlayerID  uuid.UUID `json:"playerId"`
	UserID    uuid.UUID `json:"userId"`
	Username  string    `json:"username"`
	SeatOrder int       `json:"seatOrder"`
	HandValue int       `json:"handValue"`
	CardCount int       `json:"cardCount"`
}

func GetLeaderboard(gameID uuid.UUID, players player.Repository, shoes shoe.Repository) ([]LeaderboardEntry, error) {
	activePlayers, err := players.FindActiveByGame(gameID)
	if err != nil {
		return nil, err
	}

	entries := make([]LeaderboardEntry, 0, len(activePlayers))
	for _, p := range activePlayers {
		hand, err := shoes.FindByPlayer(p.ID)
		if err != nil {
			return nil, err
		}
		value := 0
		for _, c := range hand {
			value += c.NumericValue
		}
		entries = append(entries, LeaderboardEntry{
			PlayerID:  p.ID,
			UserID:    p.UserID,
			Username:  p.Username,
			SeatOrder: p.SeatOrder,
			HandValue: value,
			CardCount: len(hand),
		})
	}

	// Sort descending by hand value; tie-break by seat_order ascending.
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].HandValue != entries[j].HandValue {
			return entries[i].HandValue > entries[j].HandValue
		}
		return entries[i].SeatOrder < entries[j].SeatOrder
	})

	return entries, nil
}

func GetPlayerHand(playerID uuid.UUID, shoes shoe.Repository) ([]*shoe.ShoeCard, error) {
	return shoes.FindByPlayer(playerID)
}
