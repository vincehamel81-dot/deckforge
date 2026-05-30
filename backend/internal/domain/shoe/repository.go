package shoe

import "github.com/google/uuid"

type SuitCount struct {
	Suit  Suit `json:"suit"`
	Count int  `json:"count"`
}

type CardCount struct {
	Suit         Suit `json:"suit"`
	Face         Face `json:"face"`
	NumericValue int  `json:"numericValue"`
	Count        int  `json:"count"`
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
