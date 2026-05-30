package commands

import (
	"github.com/google/uuid"
	"github.com/vincehamel81-dot/deckforge/internal/domain/game"
	"github.com/vincehamel81-dot/deckforge/internal/domain/player"
	"github.com/vincehamel81-dot/deckforge/internal/domain/shoe"
)

type StartGameCommand struct {
	GameID           uuid.UUID
	DealerUserID     uuid.UUID
	InitialDealCount int
	AutoEnd          bool
}

func StartGame(cmd StartGameCommand, games game.Repository, players player.Repository, shoes shoe.Repository) (*game.Game, error) {
	g, err := games.FindByID(cmd.GameID)
	if err != nil || g == nil {
		return nil, ErrGameNotFound
	}
	if g.DealerUserID != cmd.DealerUserID {
		return nil, ErrForbidden
	}

	activeCount, err := players.CountActive(cmd.GameID)
	if err != nil {
		return nil, err
	}
	if activeCount < g.MinPlayers {
		return nil, ErrNotEnoughPlayers
	}

	undealt, err := shoes.UndealtCount(cmd.GameID)
	if err != nil {
		return nil, err
	}
	if undealt == 0 {
		return nil, ErrShoeEmpty
	}

	if err := g.Start(); err != nil {
		return nil, err
	}
	if err := games.Update(g); err != nil {
		return nil, err
	}

	// Auto-shuffle before dealing — assumption A11: shoe is always shuffled at game start.
	undealtCards, err := shoes.FindUndealtByGame(cmd.GameID)
	if err != nil {
		return nil, err
	}
	shoe.Shuffle(undealtCards)
	if err := shoes.UpdatePositions(undealtCards); err != nil {
		return nil, err
	}

	// Deal initial hand to each active player if requested.
	// Fetch undealt once and advance an index across players (same optimisation
	// as DealRound) to avoid one full table scan per player.
	if cmd.InitialDealCount > 0 {
		activePlayers, err := players.FindActiveByGame(cmd.GameID)
		if err != nil {
			return nil, err
		}
		undealtForDeal, err := shoes.FindUndealtByGame(cmd.GameID)
		if err != nil {
			return nil, err
		}
		cardIdx := 0
		for _, p := range activePlayers {
			available := len(undealtForDeal) - cardIdx
			toDeal := cmd.InitialDealCount
			if toDeal > available {
				toDeal = available
			}
			for i := 0; i < toDeal; i++ {
				if err := shoes.DealCard(undealtForDeal[cardIdx].ID, p.ID); err != nil {
					return nil, err
				}
				cardIdx++
			}
		}
	}

	return g, nil
}

type EndGameCommand struct {
	GameID       uuid.UUID
	DealerUserID uuid.UUID
}

func EndGame(cmd EndGameCommand, games game.Repository) (*game.Game, error) {
	g, err := games.FindByID(cmd.GameID)
	if err != nil || g == nil {
		return nil, ErrGameNotFound
	}
	if g.DealerUserID != cmd.DealerUserID {
		return nil, ErrForbidden
	}
	if err := g.End(); err != nil {
		return nil, err
	}
	if err := games.Update(g); err != nil {
		return nil, err
	}
	return g, nil
}

type DeleteGameCommand struct {
	GameID       uuid.UUID
	DealerUserID uuid.UUID
	IsAdmin      bool
}

func DeleteGame(cmd DeleteGameCommand, games game.Repository) error {
	g, err := games.FindByID(cmd.GameID)
	if err != nil || g == nil {
		return ErrGameNotFound
	}
	if !cmd.IsAdmin && g.DealerUserID != cmd.DealerUserID {
		return ErrForbidden
	}
	return games.Delete(cmd.GameID)
}
