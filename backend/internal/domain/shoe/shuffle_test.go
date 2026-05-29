package shoe_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/vincehamel81-dot/deckforge/internal/domain/shoe"
)

func TestNewDeck_52UniqueCards(t *testing.T) {
	gameID := uuid.New()
	cards := shoe.NewDeck(gameID, 0)

	if len(cards) != 52 {
		t.Fatalf("expected 52 cards, got %d", len(cards))
	}

	seen := make(map[string]bool)
	for _, c := range cards {
		key := string(c.Suit) + ":" + string(c.Face)
		if seen[key] {
			t.Errorf("duplicate card: %s", key)
		}
		seen[key] = true
	}
}

func TestNumericValues(t *testing.T) {
	cases := []struct {
		face     shoe.Face
		expected int
	}{
		{shoe.FaceAce, 1},
		{shoe.FaceTwo, 2},
		{shoe.FaceTen, 10},
		{shoe.FaceJack, 11},
		{shoe.FaceQueen, 12},
		{shoe.FaceKing, 13},
	}
	for _, tc := range cases {
		if v := shoe.NumericValue[tc.face]; v != tc.expected {
			t.Errorf("NumericValue[%s] = %d, want %d", tc.face, v, tc.expected)
		}
	}
}

func TestShuffle_AllCardsPresent(t *testing.T) {
	gameID := uuid.New()
	cards := shoe.NewDeck(gameID, 0)

	// Record original card IDs
	original := make(map[string]bool)
	for _, c := range cards {
		original[c.ID.String()] = true
	}

	shoe.Shuffle(cards)

	// All 52 cards must still be present after shuffle
	if len(cards) != 52 {
		t.Fatalf("shuffle changed card count: %d", len(cards))
	}
	for _, c := range cards {
		if !original[c.ID.String()] {
			t.Errorf("shuffle introduced unknown card %s", c.ID)
		}
	}
}

func TestShuffle_PositionsReassigned(t *testing.T) {
	gameID := uuid.New()
	cards := shoe.NewDeck(gameID, 0)
	shoe.Shuffle(cards)

	for i, c := range cards {
		if c.Position != i {
			t.Errorf("card at index %d has Position=%d, expected %d", i, c.Position, i)
		}
	}
}

func TestShuffle_IsDifferentFromOriginal(t *testing.T) {
	// Probabilistic test: 52! orderings, odds of same order ~1/52! ≈ 0.
	gameID := uuid.New()
	cards := shoe.NewDeck(gameID, 0)

	originalOrder := make([]string, len(cards))
	for i, c := range cards {
		originalOrder[i] = c.ID.String()
	}

	shoe.Shuffle(cards)

	different := false
	for i, c := range cards {
		if c.ID.String() != originalOrder[i] {
			different = true
			break
		}
	}
	if !different {
		t.Error("shuffle produced identical order — extremely unlikely unless broken")
	}
}
