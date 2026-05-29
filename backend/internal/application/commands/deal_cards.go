package commands

import (
	"github.com/google/uuid"
	"github.com/vincehamel81-dot/deckforge/internal/domain/game"
	"github.com/vincehamel81-dot/deckforge/internal/domain/player"
	"github.com/vincehamel81-dot/deckforge/internal/domain/shoe"
)

type DealCardsCommand struct {
	GameID       uuid.UUID
	DealerUserID uuid.UUID
	PlayerID     uuid.UUID
	Count        int
}

type DealCardsResult struct {
	DealtCount  int
	GameEnded   bool
}

func DealCards(cmd DealCardsCommand, games game.Repository, shoes shoe.Repository, players player.Repository) (*DealCardsResult, error) {
	g, err := games.FindByID(cmd.GameID)
	if err != nil || g == nil {
		return nil, ErrGameNotFound
	}
	if g.DealerUserID != cmd.DealerUserID {
		return nil, ErrForbidden
	}
	if g.Status != game.StatusInProgress {
		return nil, game.ErrNotInProgress
	}

	// Verify target player is active in this game.
	p, err := players.FindByID(cmd.PlayerID)
	if err != nil || p == nil || p.GameID != cmd.GameID || !p.IsActive() {
		return nil, ErrPlayerNotFound
	}

	undealt, err := shoes.FindUndealtByGame(cmd.GameID)
	if err != nil {
		return nil, err
	}

	toDeal := cmd.Count
	if toDeal > len(undealt) {
		toDeal = len(undealt)
	}

	for i := 0; i < toDeal; i++ {
		if err := shoes.DealCard(undealt[i].ID, cmd.PlayerID); err != nil {
			return nil, err
		}
	}

	result := &DealCardsResult{DealtCount: toDeal}

	// Auto-end check: if remaining cards < active players, game ends.
	ended, err := checkAutoEnd(cmd.GameID, games, shoes, players)
	if err != nil {
		return nil, err
	}
	result.GameEnded = ended

	return result, nil
}

// checkAutoEnd ends the game if the shoe cannot deal a full round.
// Returns true if the game was ended.
func checkAutoEnd(gameID uuid.UUID, games game.Repository, shoes shoe.Repository, players player.Repository) (bool, error) {
	remaining, err := shoes.UndealtCount(gameID)
	if err != nil {
		return false, err
	}
	activeCount, err := players.CountActive(gameID)
	if err != nil {
		return false, err
	}
	if activeCount == 0 || remaining >= activeCount {
		return false, nil
	}
	g, err := games.FindByID(gameID)
	if err != nil || g == nil {
		return false, err
	}
	if g.Status == game.StatusFinished {
		return false, nil
	}
	if err := g.End(); err != nil {
		return false, err
	}
	return true, games.Update(g)
}
