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

	// Deal initial hand to each active player if requested.
	if cmd.InitialDealCount > 0 {
		activePlayers, err := players.FindActiveByGame(cmd.GameID)
		if err != nil {
			return nil, err
		}
		for _, p := range activePlayers {
			dealCmd := DealCardsCommand{
				GameID:     cmd.GameID,
				DealerUserID: cmd.DealerUserID,
				PlayerID:   p.ID,
				Count:      cmd.InitialDealCount,
			}
			if _, err := DealCards(dealCmd, games, shoes, players); err != nil {
				return nil, err
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
}

func DeleteGame(cmd DeleteGameCommand, games game.Repository) error {
	g, err := games.FindByID(cmd.GameID)
	if err != nil || g == nil {
		return ErrGameNotFound
	}
	if g.DealerUserID != cmd.DealerUserID {
		return ErrForbidden
	}
	return games.Delete(cmd.GameID)
}
