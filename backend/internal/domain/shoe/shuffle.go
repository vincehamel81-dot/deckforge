package shoe

import (
	"crypto/rand"
	"math/big"
)

// Shuffle implements Fisher-Yates on a slice of ShoeCards.
// Uses crypto/rand for randomness — library shuffle functions are not used
// per assignment requirement. Only undealt cards should be passed in.
func Shuffle(cards []*ShoeCard) {
	n := len(cards)
	for i := n - 1; i > 0; i-- {
		j := cryptoRandInt(i + 1)
		cards[i], cards[j] = cards[j], cards[i]
	}
	// Re-assign Position values so DB order reflects the shuffle result.
	for i, c := range cards {
		c.Position = i
	}
}

func cryptoRandInt(max int) int {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		// crypto/rand failure is unrecoverable — panic is appropriate here.
		panic("crypto/rand unavailable: " + err.Error())
	}
	return int(n.Int64())
}
