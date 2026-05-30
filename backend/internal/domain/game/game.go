package game

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type Status string

const (
	StatusWaiting    Status = "WAITING"
	StatusInProgress Status = "IN_PROGRESS"
	StatusFinished   Status = "FINISHED"
)

var (
	ErrAlreadyStarted  = errors.New("game already started")
	ErrNotInProgress   = errors.New("game is not in progress")
	ErrAlreadyFinished = errors.New("game already finished")
	ErrShoeSealed      = errors.New("shoe is sealed once game is in progress")
	ErrGameNotJoinable = errors.New("game is not joinable")
)

type Game struct {
	ID                  uuid.UUID  `json:"id"`
	DealerUserID        uuid.UUID  `json:"dealerUserId"`
	Status              Status     `json:"status"`
	DeckCount           int        `json:"deckCount"`
	MinPlayers          int        `json:"minPlayers"`
	MaxPlayers          int        `json:"maxPlayers"`
	CurrentTurnPlayerID *uuid.UUID `json:"currentTurnPlayerId,omitempty"`
	CreatedAt           time.Time  `json:"createdAt"`
	StartedAt           *time.Time `json:"startedAt,omitempty"`
	FinishedAt          *time.Time `json:"finishedAt,omitempty"`
}

func New(dealerUserID uuid.UUID, deckCount, minPlayers, maxPlayers int) *Game {
	now := time.Now().UTC()
	return &Game{
		ID:           uuid.New(),
		DealerUserID: dealerUserID,
		Status:       StatusWaiting,
		DeckCount:    deckCount,
		MinPlayers:   minPlayers,
		MaxPlayers:   maxPlayers,
		CreatedAt:    now,
	}
}

func (g *Game) Start() error {
	if g.Status != StatusWaiting {
		return ErrAlreadyStarted
	}
	now := time.Now().UTC()
	g.Status = StatusInProgress
	g.StartedAt = &now
	return nil
}

func (g *Game) End() error {
	if g.Status == StatusFinished {
		return ErrAlreadyFinished
	}
	now := time.Now().UTC()
	g.Status = StatusFinished
	g.FinishedAt = &now
	return nil
}

func (g *Game) CanAddDeck() bool {
	return g.Status == StatusWaiting
}

func (g *Game) CanJoin() bool {
	return g.Status == StatusWaiting || g.Status == StatusInProgress
}

func (g *Game) TotalCards() int {
	return g.DeckCount * 52
}
