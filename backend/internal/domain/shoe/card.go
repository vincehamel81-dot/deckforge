package shoe

import (
	"github.com/google/uuid"
)

type Suit string

const (
	SuitHearts   Suit = "HEARTS"
	SuitSpades   Suit = "SPADES"
	SuitClubs    Suit = "CLUBS"
	SuitDiamonds Suit = "DIAMONDS"
)

type Face string

const (
	FaceAce   Face = "ACE"
	FaceTwo   Face = "TWO"
	FaceThree Face = "THREE"
	FaceFour  Face = "FOUR"
	FaceFive  Face = "FIVE"
	FaceSix   Face = "SIX"
	FaceSeven Face = "SEVEN"
	FaceEight Face = "EIGHT"
	FaceNine  Face = "NINE"
	FaceTen   Face = "TEN"
	FaceJack  Face = "JACK"
	FaceQueen Face = "QUEEN"
	FaceKing  Face = "KING"
)

// NumericValue maps each face to its scoring value per assignment spec.
// ACE=1, 2-10=face value, JACK=11, QUEEN=12, KING=13.
var NumericValue = map[Face]int{
	FaceAce:   1,
	FaceTwo:   2,
	FaceThree: 3,
	FaceFour:  4,
	FaceFive:  5,
	FaceSix:   6,
	FaceSeven: 7,
	FaceEight: 8,
	FaceNine:  9,
	FaceTen:   10,
	FaceJack:  11,
	FaceQueen: 12,
	FaceKing:  13,
}

// SuitOrder defines the canonical display order for shoe count queries.
var SuitOrder = []Suit{SuitHearts, SuitSpades, SuitClubs, SuitDiamonds}

// AllFaces lists faces in descending value order (for sorted count queries).
var AllFaces = []Face{
	FaceKing, FaceQueen, FaceJack, FaceTen, FaceNine, FaceEight,
	FaceSeven, FaceSix, FaceFive, FaceFour, FaceThree, FaceTwo, FaceAce,
}

type ShoeCard struct {
	ID             uuid.UUID
	GameID         uuid.UUID
	Suit           Suit
	Face           Face
	NumericValue   int
	Position       int
	HeldByPlayerID *uuid.UUID
}

func (c *ShoeCard) IsDealt() bool {
	return c.HeldByPlayerID != nil
}

// NewDeck generates 52 ShoeCards for a given game with sequential positions
// starting at the given offset (used when multiple decks are added).
func NewDeck(gameID uuid.UUID, positionOffset int) []*ShoeCard {
	cards := make([]*ShoeCard, 0, 52)
	pos := positionOffset
	for _, suit := range SuitOrder {
		for _, face := range []Face{
			FaceAce, FaceTwo, FaceThree, FaceFour, FaceFive, FaceSix,
			FaceSeven, FaceEight, FaceNine, FaceTen, FaceJack, FaceQueen, FaceKing,
		} {
			cards = append(cards, &ShoeCard{
				ID:           uuid.New(),
				GameID:       gameID,
				Suit:         suit,
				Face:         face,
				NumericValue: NumericValue[face],
				Position:     pos,
			})
			pos++
		}
	}
	return cards
}
