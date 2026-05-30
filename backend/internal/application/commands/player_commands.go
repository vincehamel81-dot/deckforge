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
	existing, err := players.FindActiveByUser(cmd.UserID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrAlreadyInGame
	}

	activeCount, err := players.CountActive(cmd.GameID)
	if err != nil {
		return nil, err
	}
	if activeCount >= g.MaxPlayers {
		return nil, ErrMaxPlayersExceeded
	}

	p := player.New(cmd.GameID, cmd.UserID, activeCount)
	if err := players.Create(p); err != nil {
		return nil, err
	}

	// Catch-up deal: if game is in progress, deal current hand size to new player.
	if g.Status == game.StatusInProgress {
		activePlayers, err := players.FindActiveByGame(cmd.GameID)
		if err != nil {
			return nil, err
		}
		// Current hand size = minimum cards any existing active player holds.
		handSize := catchUpHandSize(activePlayers, p.ID, shoes)
		if handSize > 0 {
			undealt, err := shoes.FindUndealtByGame(cmd.GameID)
			if err != nil {
				return nil, err
			}
			toDeal := handSize
			if toDeal > len(undealt) {
				toDeal = len(undealt)
			}
			for i := 0; i < toDeal; i++ {
				if err := shoes.DealCard(undealt[i].ID, p.ID); err != nil {
					return nil, err
				}
			}
			// Check auto-end after catch-up.
			if _, err := checkAutoEnd(cmd.GameID, games, shoes, players); err != nil {
				return nil, err
			}
		}
	}

	return p, nil
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

func RemovePlayer(cmd RemovePlayerCommand, games game.Repository, players player.Repository, shoes shoe.Repository) error {
	g, err := games.FindByID(cmd.GameID)
	if err != nil || g == nil {
		return ErrGameNotFound
	}

	p, err := players.FindByID(cmd.PlayerID)
	if err != nil || p == nil || p.GameID != cmd.GameID {
		return ErrPlayerNotFound
	}
	if !p.IsActive() {
		return ErrPlayerNotFound
	}

	// Admins can remove anyone; otherwise only the dealer or the player themselves.
	isDealerRemoving := g.DealerUserID.String() == cmd.RequesterUserID.String()
	isPlayerLeaving := p.UserID.String() == cmd.RequesterUserID.String()
	if !cmd.IsAdmin && !isDealerRemoving && !isPlayerLeaving {
		return ErrForbidden
	}

	// Return player's cards to the shoe.
	if err := shoes.ReturnCardsByPlayer(cmd.PlayerID); err != nil {
		return err
	}

	p.Leave()
	if err := players.Update(p); err != nil {
		return err
	}

	// Check auto-end after player leaves.
	_, err = checkAutoEnd(cmd.GameID, games, shoes, players)
	return err
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
