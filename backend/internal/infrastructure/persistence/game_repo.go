package persistence

import (
	"errors"

	"github.com/google/uuid"
	"github.com/vincehamel81-dot/deckforge/internal/domain/game"
	"gorm.io/gorm"
)

type gameRepo struct{ db *gorm.DB }

func NewGameRepo(db *gorm.DB) game.Repository {
	return &gameRepo{db: db}
}

func (r *gameRepo) Create(g *game.Game) error {
	return r.db.Create(toGameModel(g)).Error
}

func (r *gameRepo) FindByID(id uuid.UUID) (*game.Game, error) {
	var m GameModel
	if err := r.db.First(&m, "id = ?", id.String()).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toGameDomain(&m), nil
}

func (r *gameRepo) FindAll(statusFilter *game.Status) ([]*game.Game, error) {
	var models []GameModel
	q := r.db.Model(&GameModel{})
	if statusFilter != nil {
		q = q.Where("status = ?", string(*statusFilter))
	}
	if err := q.Order("created_at desc").Find(&models).Error; err != nil {
		return nil, err
	}
	games := make([]*game.Game, len(models))
	for i := range models {
		games[i] = toGameDomain(&models[i])
	}
	return games, nil
}

func (r *gameRepo) Update(g *game.Game) error {
	return r.db.Save(toGameModel(g)).Error
}

func (r *gameRepo) Delete(id uuid.UUID) error {
	return r.db.Delete(&GameModel{}, "id = ?", id.String()).Error
}

func toGameModel(g *game.Game) *GameModel {
	m := &GameModel{
		ID:           g.ID.String(),
		DealerUserID: g.DealerUserID.String(),
		Status:       string(g.Status),
		DeckCount:    g.DeckCount,
		MinPlayers:   g.MinPlayers,
		MaxPlayers:   g.MaxPlayers,
		CreatedAt:    g.CreatedAt,
		StartedAt:    g.StartedAt,
		FinishedAt:   g.FinishedAt,
	}
	if g.CurrentTurnPlayerID != nil {
		s := g.CurrentTurnPlayerID.String()
		m.CurrentTurnPlayerID = &s
	}
	return m
}

func toGameDomain(m *GameModel) *game.Game {
	id, _ := uuid.Parse(m.ID)
	dealerID, _ := uuid.Parse(m.DealerUserID)
	g := &game.Game{
		ID:           id,
		DealerUserID: dealerID,
		Status:       game.Status(m.Status),
		DeckCount:    m.DeckCount,
		MinPlayers:   m.MinPlayers,
		MaxPlayers:   m.MaxPlayers,
		CreatedAt:    m.CreatedAt,
		StartedAt:    m.StartedAt,
		FinishedAt:   m.FinishedAt,
	}
	if m.CurrentTurnPlayerID != nil {
		pid, _ := uuid.Parse(*m.CurrentTurnPlayerID)
		g.CurrentTurnPlayerID = &pid
	}
	return g
}

var _ game.Repository = (*gameRepo)(nil)
