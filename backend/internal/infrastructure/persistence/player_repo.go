package persistence

import (
	"errors"

	"github.com/google/uuid"
	"github.com/vincehamel81-dot/deckforge/internal/domain/player"
	"gorm.io/gorm"
)

type playerRepo struct{ db *gorm.DB }

func NewPlayerRepo(db *gorm.DB) player.Repository {
	return &playerRepo{db: db}
}

func (r *playerRepo) Create(p *player.Player) error {
	return r.db.Create(toPlayerModel(p)).Error
}

func (r *playerRepo) FindByGame(gameID uuid.UUID) ([]*player.Player, error) {
	var models []PlayerModel
	if err := r.db.Where("game_id = ?", gameID.String()).Order("seat_order asc").Find(&models).Error; err != nil {
		return nil, err
	}
	return toPlayerDomainSlice(models), nil
}

func (r *playerRepo) FindActiveByGame(gameID uuid.UUID) ([]*player.Player, error) {
	var models []PlayerModel
	if err := r.db.Where("game_id = ? AND left_at IS NULL", gameID.String()).Order("seat_order asc").Find(&models).Error; err != nil {
		return nil, err
	}
	return toPlayerDomainSlice(models), nil
}

func (r *playerRepo) FindByUserAndGame(userID, gameID uuid.UUID) (*player.Player, error) {
	var m PlayerModel
	err := r.db.First(&m, "user_id = ? AND game_id = ?", userID.String(), gameID.String()).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toPlayerDomain(&m), nil
}

func (r *playerRepo) FindActiveByUser(userID uuid.UUID) (*player.Player, error) {
	var m PlayerModel
	err := r.db.First(&m, "user_id = ? AND left_at IS NULL", userID.String()).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toPlayerDomain(&m), nil
}

func (r *playerRepo) FindByID(id uuid.UUID) (*player.Player, error) {
	var m PlayerModel
	if err := r.db.First(&m, "id = ?", id.String()).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toPlayerDomain(&m), nil
}

func (r *playerRepo) Update(p *player.Player) error {
	return r.db.Save(toPlayerModel(p)).Error
}

func (r *playerRepo) CountActive(gameID uuid.UUID) (int, error) {
	var count int64
	err := r.db.Model(&PlayerModel{}).Where("game_id = ? AND left_at IS NULL", gameID.String()).Count(&count).Error
	return int(count), err
}

func toPlayerModel(p *player.Player) *PlayerModel {
	return &PlayerModel{
		ID:        p.ID.String(),
		GameID:    p.GameID.String(),
		UserID:    p.UserID.String(),
		SeatOrder: p.SeatOrder,
		JoinedAt:  p.JoinedAt,
		LeftAt:    p.LeftAt,
	}
}

func toPlayerDomain(m *PlayerModel) *player.Player {
	id, _ := uuid.Parse(m.ID)
	gameID, _ := uuid.Parse(m.GameID)
	userID, _ := uuid.Parse(m.UserID)
	return &player.Player{
		ID:        id,
		GameID:    gameID,
		UserID:    userID,
		SeatOrder: m.SeatOrder,
		JoinedAt:  m.JoinedAt,
		LeftAt:    m.LeftAt,
	}
}

func toPlayerDomainSlice(models []PlayerModel) []*player.Player {
	players := make([]*player.Player, len(models))
	for i := range models {
		players[i] = toPlayerDomain(&models[i])
	}
	return players
}

var _ player.Repository = (*playerRepo)(nil)
