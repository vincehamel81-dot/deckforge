package player

import "github.com/google/uuid"

type Repository interface {
	Create(p *Player) error
	FindByGame(gameID uuid.UUID) ([]*Player, error)
	FindActiveByGame(gameID uuid.UUID) ([]*Player, error)
	FindByUserAndGame(userID, gameID uuid.UUID) (*Player, error)
	FindActiveByUser(userID uuid.UUID) (*Player, error)
	FindByID(id uuid.UUID) (*Player, error)
	Update(p *Player) error
	CountActive(gameID uuid.UUID) (int, error)
	MarkAllLeft(gameID uuid.UUID) error
}
