package shoe

import "github.com/google/uuid"

type SuitCount struct {
	Suit  Suit
	Count int
}

type CardCount struct {
	Suit         Suit
	Face         Face
	NumericValue int
	Count        int
}

type Repository interface {
	AddCards(cards []*ShoeCard) error
	FindUndealtByGame(gameID uuid.UUID) ([]*ShoeCard, error)
	FindByPlayer(playerID uuid.UUID) ([]*ShoeCard, error)
	UpdatePositions(cards []*ShoeCard) error
	DealCard(cardID uuid.UUID, playerID uuid.UUID) error
	ReturnCardsByPlayer(playerID uuid.UUID) error
	CountBySuit(gameID uuid.UUID) ([]SuitCount, error)
	CountByCard(gameID uuid.UUID) ([]CardCount, error)
	UndealtCount(gameID uuid.UUID) (int, error)
}
