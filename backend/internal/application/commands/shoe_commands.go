package commands

import (
	"github.com/google/uuid"
	"github.com/vincehamel81-dot/deckforge/internal/domain/game"
	"github.com/vincehamel81-dot/deckforge/internal/domain/shoe"
)

type AddDeckToShoeCommand struct {
	GameID       uuid.UUID
	DealerUserID uuid.UUID
}

func AddDeckToShoe(cmd AddDeckToShoeCommand, games game.Repository, shoes shoe.Repository) error {
	g, err := games.FindByID(cmd.GameID)
	if err != nil || g == nil {
		return ErrGameNotFound
	}
	if g.DealerUserID != cmd.DealerUserID {
		return ErrForbidden
	}
	if !g.CanAddDeck() {
		return game.ErrShoeSealed
	}

	// Count existing cards to set position offset for new deck.
	existing, err := shoes.UndealtCount(cmd.GameID)
	if err != nil {
		return err
	}

	cards := shoe.NewDeck(cmd.GameID, existing)
	if err := shoes.AddCards(cards); err != nil {
		return err
	}

	g.DeckCount++
	return games.Update(g)
}

type ShuffleShoeCommand struct {
	GameID       uuid.UUID
	DealerUserID uuid.UUID
}

func ShuffleShoe(cmd ShuffleShoeCommand, games game.Repository, shoes shoe.Repository) error {
	g, err := games.FindByID(cmd.GameID)
	if err != nil || g == nil {
		return ErrGameNotFound
	}
	if g.DealerUserID != cmd.DealerUserID {
		return ErrForbidden
	}

	undealt, err := shoes.FindUndealtByGame(cmd.GameID)
	if err != nil {
		return err
	}
	if len(undealt) == 0 {
		return nil
	}

	shoe.Shuffle(undealt)
	return shoes.UpdatePositions(undealt)
}
