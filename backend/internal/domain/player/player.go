package player

import (
	"time"

	"github.com/google/uuid"
)

type Player struct {
	ID        uuid.UUID
	GameID    uuid.UUID
	UserID    uuid.UUID
	SeatOrder int
	JoinedAt  time.Time
	LeftAt    *time.Time
}

func New(gameID, userID uuid.UUID, seatOrder int) *Player {
	return &Player{
		ID:        uuid.New(),
		GameID:    gameID,
		UserID:    userID,
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
