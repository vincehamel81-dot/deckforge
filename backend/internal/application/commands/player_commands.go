package commands

import (
	"github.com/google/uuid"
	"github.com/vincehamel81-dot/deckforge/internal/domain/game"
	"github.com/vincehamel81-dot/deckforge/internal/domain/player"
	"github.com/vincehamel81-dot/deckforge/internal/domain/shoe"
)

type AddPlayerCommand struct {
	GameID uuid.UUID
	UserID uuid.UUID
}

func AddPlayer(cmd AddPlayerCommand, games game.Repository, players player.Repository, shoes shoe.Repository) (*player.Player, error) {
	g, err := games.FindByID(cmd.GameID)
	if err != nil || g == nil {
		return nil, ErrGameNotFound
	}
	if !g.CanJoin() {
		return nil, ErrGameNotJoinable
	}

	// A user can only be in one active game at a time.
	// Idempotent: re-joining the same game returns the existing player record.
	existing, err := players.FindActiveByUser(cmd.UserID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		if existing.GameID == cmd.GameID {
			return existing, nil
		}
		return nil, ErrAlreadyInGame
	}

	activeCount, err := players.CountActive(cmd.GameID)
	if err != nil {
		return nil, err
	}
	if activeCount >= g.MaxPlayers {
		return nil, ErrMaxPlayersExceeded
	}

	// Pre-validate shoe capacity before committing the join.
	// If the shoe cannot deal a full catch-up hand the join is rejected — a player
	// who joins with fewer cards than their opponents has no fair path to win.
	if g.Status == game.StatusInProgress {
		existingActive, err := players.FindActiveByGame(cmd.GameID)
		if err != nil {
			return nil, err
		}
		needed := minHandSize(existingActive, shoes)
		if needed > 0 {
			remaining, err := shoes.UndealtCount(cmd.GameID)
			if err != nil {
				return nil, err
			}
			if remaining < needed {
				return nil, ErrNotEnoughCardsToJoin
			}
		}
	}

	p := player.New(cmd.GameID, cmd.UserID, activeCount)
	if err := players.Create(p); err != nil {
		return nil, err
	}

	// Catch-up deal: deal the minimum hand size to the new player so all active
	// players hold the same number of cards entering their first turn.
	if g.Status == game.StatusInProgress {
		activePlayers, err := players.FindActiveByGame(cmd.GameID)
		if err != nil {
			return nil, err
		}
		handSize := catchUpHandSize(activePlayers, p.ID, shoes)
		if handSize > 0 {
			undealt, err := shoes.FindUndealtByGame(cmd.GameID)
			if err != nil {
				return nil, err
			}
			for i := 0; i < handSize && i < len(undealt); i++ {
				if err := shoes.DealCard(undealt[i].ID, p.ID); err != nil {
					return nil, err
				}
			}
			if _, err := checkAutoEnd(cmd.GameID, games, shoes, players); err != nil {
				return nil, err
			}
		}
	}

	return p, nil
}

// minHandSize returns the minimum hand size among all provided players.
// Used to pre-validate shoe capacity before a player joins.
func minHandSize(activePlayers []*player.Player, shoes shoe.Repository) int {
	min := -1
	for _, p := range activePlayers {
		hand, err := shoes.FindByPlayer(p.ID)
		if err != nil {
			continue
		}
		if min == -1 || len(hand) < min {
			min = len(hand)
		}
	}
	if min == -1 {
		return 0
	}
	return min
}

// catchUpHandSize returns the minimum hand size among all active players except the newcomer.
func catchUpHandSize(activePlayers []*player.Player, newPlayerID uuid.UUID, shoes shoe.Repository) int {
	min := -1
	for _, p := range activePlayers {
		if p.ID == newPlayerID {
			continue
		}
		hand, err := shoes.FindByPlayer(p.ID)
		if err != nil {
			continue
		}
		if min == -1 || len(hand) < min {
			min = len(hand)
		}
	}
	if min == -1 {
		return 0
	}
	return min
}

type RemovePlayerCommand struct {
	GameID          uuid.UUID
	PlayerID        uuid.UUID
	RequesterUserID uuid.UUID
	IsAdmin         bool
}

// RemovePlayer removes a player from a game and returns cards to the shoe.
// Returns (gameAutoEnded, error) — gameAutoEnded is true when the player
// count drop triggered an automatic FINISHED transition.
func RemovePlayer(cmd RemovePlayerCommand, games game.Repository, players player.Repository, shoes shoe.Repository) (bool, error) {
	g, err := games.FindByID(cmd.GameID)
	if err != nil || g == nil {
		return false, ErrGameNotFound
	}

	p, err := players.FindByID(cmd.PlayerID)
	if err != nil || p == nil || p.GameID != cmd.GameID {
		return false, ErrPlayerNotFound
	}
	if !p.IsActive() {
		return false, ErrPlayerNotFound
	}

	// Admins can remove anyone; otherwise only the dealer or the player themselves.
	isDealerRemoving := g.DealerUserID.String() == cmd.RequesterUserID.String()
	isPlayerLeaving := p.UserID.String() == cmd.RequesterUserID.String()
	if !cmd.IsAdmin && !isDealerRemoving && !isPlayerLeaving {
		return false, ErrForbidden
	}

	// Return player's cards to the shoe.
	if err := shoes.ReturnCardsByPlayer(cmd.PlayerID); err != nil {
		return false, err
	}

	p.Leave()
	if err := players.Update(p); err != nil {
		return false, err
	}

	return checkAutoEnd(cmd.GameID, games, shoes, players)
}

var ErrGameNotJoinable = ErrAlreadyInGame // re-use for joinable check (overridden below)

func init() {
	// Ensure ErrGameNotJoinable has its own distinct value.
	ErrGameNotJoinable = newSentinelError("game is not joinable (finished or at capacity)")
}

func newSentinelError(msg string) error {
	return &sentinelError{msg}
}

type sentinelError struct{ msg string }

func (e *sentinelError) Error() string { return e.msg }
