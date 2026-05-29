package persistence

import "time"

// GORM models are separate from domain entities.
// Mappers (toModel/toDomain) live in each repo file.

type UserModel struct {
	ID        string `gorm:"primaryKey"`
	Username  string `gorm:"uniqueIndex;not null"`
	Role      string `gorm:"not null;default:'user'"`
	CreatedAt time.Time
}

func (UserModel) TableName() string { return "users" }

type GameModel struct {
	ID                  string  `gorm:"primaryKey"`
	DealerUserID        string  `gorm:"not null;index"`
	Status              string  `gorm:"not null;default:'WAITING'"`
	DeckCount           int     `gorm:"not null;default:0"`
	MinPlayers          int     `gorm:"not null;default:2"`
	MaxPlayers          int     `gorm:"not null;default:8"`
	CurrentTurnPlayerID *string `gorm:"index"`
	CreatedAt           time.Time
	StartedAt           *time.Time
	FinishedAt          *time.Time
}

func (GameModel) TableName() string { return "games" }

// ShoeCardModel stores one row per card instance in the shoe.
// HeldByPlayerID IS NULL means the card is undealt (in the shoe).
// HeldByPlayerID IS NOT NULL means the card has been dealt to that player.
// There is no separate state field — the null check is the state.
type ShoeCardModel struct {
	ID             string  `gorm:"primaryKey"`
	GameID         string  `gorm:"not null;index"`
	Suit           string  `gorm:"not null"`
	Face           string  `gorm:"not null"`
	NumericValue   int     `gorm:"not null"`
	Position       int     `gorm:"not null;index"`
	HeldByPlayerID *string `gorm:"index"`
}

func (ShoeCardModel) TableName() string { return "shoe_cards" }

type PlayerModel struct {
	ID        string `gorm:"primaryKey"`
	GameID    string `gorm:"not null;index"`
	UserID    string `gorm:"not null;index"`
	SeatOrder int    `gorm:"not null"`
	JoinedAt  time.Time
	LeftAt    *time.Time
}

func (PlayerModel) TableName() string { return "players" }
