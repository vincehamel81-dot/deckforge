package commands

import (
	"errors"

	"github.com/google/uuid"
	"github.com/vincehamel81-dot/deckforge/internal/domain/game"
	"github.com/vincehamel81-dot/deckforge/internal/domain/player"
)

var (
	ErrInvalidDeckCount      = errors.New("deck count must be between 1 and 8")
	ErrInvalidPlayerRange    = errors.New("min players must be ≥ 2 and ≤ max players")
	ErrMaxPlayersExceeded    = errors.New("game is at maximum player capacity")
	ErrGameNotFound          = errors.New("game not found")
	ErrForbidden             = errors.New("you do not have permission to perform this action")
	ErrPlayerNotFound        = errors.New("player not found")
	ErrAlreadyInGame         = errors.New("ALREADY_IN_GAME")
	ErrNotEnoughPlayers      = errors.New("not enough players to start the game")
	ErrShoeEmpty             = errors.New("shoe has no cards — add decks before starting")
	// ErrNotEnoughCardsToJoin is returned when joining an IN_PROGRESS game whose shoe
	// cannot deal a full catch-up hand to the new player (remaining < min hand size of peers).
	ErrNotEnoughCardsToJoin  = errors.New("SHOE_TOO_LOW_FOR_JOIN")
)

type CreateGameCommand struct {
	DealerUserID uuid.UUID
	DeckCount    int
	MinPlayers   int
	MaxPlayers   int
}

type CreateGameResult struct {
	Game   *game.Game
	Player *player.Player
}

func CreateGame(cmd CreateGameCommand, games game.Repository, players player.Repository) (*CreateGameResult, error) {
	if cmd.DeckCount < 1 || cmd.DeckCount > 8 {
		return nil, ErrInvalidDeckCount
	}
	if cmd.MinPlayers < 2 || cmd.MinPlayers > cmd.MaxPlayers {
		return nil, ErrInvalidPlayerRange
	}

	existing, err := players.FindActiveByUser(cmd.DealerUserID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrAlreadyInGame
	}

	// DeckCount starts at 0 — physical decks are added via AddDeckToShoe after creation.
	g := game.New(cmd.DealerUserID, 0, cmd.MinPlayers, cmd.MaxPlayers)
	if err := games.Create(g); err != nil {
		return nil, err
	}

	// Dealer auto-joins as player 0 (seat 0).
	p := player.New(g.ID, cmd.DealerUserID, 0)
	if err := players.Create(p); err != nil {
		return nil, err
	}

	return &CreateGameResult{Game: g, Player: p}, nil
}
