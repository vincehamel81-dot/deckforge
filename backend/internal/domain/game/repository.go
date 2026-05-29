package game

import "github.com/google/uuid"

type Repository interface {
	Create(g *Game) error
	FindByID(id uuid.UUID) (*Game, error)
	FindAll(statusFilter *Status) ([]*Game, error)
	Update(g *Game) error
	Delete(id uuid.UUID) error
}
