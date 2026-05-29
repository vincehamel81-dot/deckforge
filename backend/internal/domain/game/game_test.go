package game_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/vincehamel81-dot/deckforge/internal/domain/game"
)

func TestGame_StateMachine(t *testing.T) {
	g := game.New(uuid.New(), 1, 2, 8)

	if g.Status != game.StatusWaiting {
		t.Errorf("new game should be WAITING, got %s", g.Status)
	}

	// Cannot end before starting
	if err := g.End(); err != nil {
		t.Errorf("End() on WAITING should succeed (allowed), got %v", err)
	}

	// Already finished
	if err := g.End(); err == nil {
		t.Error("End() on FINISHED should return error")
	}
}

func TestGame_StateMachine_Start(t *testing.T) {
	g := game.New(uuid.New(), 1, 2, 8)

	if err := g.Start(); err != nil {
		t.Fatalf("Start() on WAITING failed: %v", err)
	}
	if g.Status != game.StatusInProgress {
		t.Errorf("after Start(), expected IN_PROGRESS, got %s", g.Status)
	}
	if g.StartedAt == nil {
		t.Error("StartedAt should be set after Start()")
	}

	// Cannot start again
	if err := g.Start(); err == nil {
		t.Error("Start() on IN_PROGRESS should return error")
	}
}

func TestGame_CanAddDeck(t *testing.T) {
	g := game.New(uuid.New(), 1, 2, 8)

	if !g.CanAddDeck() {
		t.Error("CanAddDeck() should be true in WAITING")
	}

	_ = g.Start()
	if g.CanAddDeck() {
		t.Error("CanAddDeck() should be false in IN_PROGRESS")
	}
}

func TestGame_CanJoin(t *testing.T) {
	g := game.New(uuid.New(), 1, 2, 8)
	if !g.CanJoin() {
		t.Error("CanJoin() should be true in WAITING")
	}
	_ = g.Start()
	if !g.CanJoin() {
		t.Error("CanJoin() should be true in IN_PROGRESS")
	}
	_ = g.End()
	if g.CanJoin() {
		t.Error("CanJoin() should be false in FINISHED")
	}
}

func TestGame_TotalCards(t *testing.T) {
	g := game.New(uuid.New(), 3, 2, 8)
	if g.TotalCards() != 156 {
		t.Errorf("TotalCards() = %d, want 156 (3×52)", g.TotalCards())
	}
}
