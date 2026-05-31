package player

import (
	"time"

	"github.com/google/uuid"
)

type Player struct {
	ID        uuid.UUID  `json:"id"`
	GameID    uuid.UUID  `json:"gameId"`
	UserID    uuid.UUID  `json:"userId"`
	Username  string     `json:"username"`
	SeatOrder int        `json:"seatOrder"`
	JoinedAt  time.Time  `json:"joinedAt"`
	LeftAt    *time.Time `json:"leftAt,omitempty"`
}

func New(gameID, userID uuid.UUID, username string, seatOrder int) *Player {
	return &Player{
		ID:        uuid.New(),
		GameID:    gameID,
		UserID:    userID,
		Username:  username,
		SeatOrder: seatOrder,
		JoinedAt:  time.Now().UTC(),
	}
}

func (p *Player) IsActive() bool {
	return p.LeftAt == nil
}

func (p *Player) Leave() {
	now := time.Now().UTC()
	p.LeftAt = &now
}
