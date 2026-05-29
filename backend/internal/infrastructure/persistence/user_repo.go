package persistence

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/vincehamel81-dot/deckforge/internal/domain/user"
	"gorm.io/gorm"
)

type userRepo struct{ db *gorm.DB }

func NewUserRepo(db *gorm.DB) user.Repository {
	return &userRepo{db: db}
}

func (r *userRepo) Create(u *user.User) error {
	return r.db.Create(toUserModel(u)).Error
}

func (r *userRepo) FindByID(id uuid.UUID) (*user.User, error) {
	var m UserModel
	if err := r.db.First(&m, "id = ?", id.String()).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toUserDomain(&m), nil
}

func (r *userRepo) FindByUsername(username string) (*user.User, error) {
	var m UserModel
	if err := r.db.First(&m, "username = ?", username).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toUserDomain(&m), nil
}

func (r *userRepo) ExistsByUsername(username string) (bool, error) {
	var count int64
	err := r.db.Model(&UserModel{}).Where("username = ?", username).Count(&count).Error
	return count > 0, err
}

func (r *userRepo) Update(u *user.User) error {
	return r.db.Save(toUserModel(u)).Error
}

func toUserModel(u *user.User) *UserModel {
	return &UserModel{
		ID:        u.ID.String(),
		Username:  u.Username,
		Role:      string(u.Role),
		CreatedAt: u.CreatedAt,
	}
}

func toUserDomain(m *UserModel) *user.User {
	id, _ := uuid.Parse(m.ID)
	return &user.User{
		ID:        id,
		Username:  m.Username,
		Role:      user.Role(m.Role),
		CreatedAt: m.CreatedAt,
	}
}

// ensure interface is satisfied at compile time
var _ user.Repository = (*userRepo)(nil)

// suppress unused import warning
var _ = time.Now
