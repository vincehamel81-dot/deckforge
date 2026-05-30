package persistence

import (
	"errors"

	"github.com/google/uuid"
	"github.com/vincehamel81-dot/deckforge/internal/domain/shoe"
	"gorm.io/gorm"
)

type shoeRepo struct{ db *gorm.DB }

func NewShoeRepo(db *gorm.DB) shoe.Repository {
	return &shoeRepo{db: db}
}

func (r *shoeRepo) AddCards(cards []*shoe.ShoeCard) error {
	models := make([]ShoeCardModel, len(cards))
	for i, c := range cards {
		models[i] = toShoeCardModel(c)
	}
	return r.db.Create(&models).Error
}

func (r *shoeRepo) FindUndealtByGame(gameID uuid.UUID) ([]*shoe.ShoeCard, error) {
	var models []ShoeCardModel
	err := r.db.Where("game_id = ? AND held_by_player_id IS NULL", gameID.String()).
		Order("position asc").Find(&models).Error
	if err != nil {
		return nil, err
	}
	return toShoeCardDomainSlice(models), nil
}

func (r *shoeRepo) FindByPlayer(playerID uuid.UUID) ([]*shoe.ShoeCard, error) {
	var models []ShoeCardModel
	err := r.db.Where("held_by_player_id = ?", playerID.String()).
		Order("position asc").Find(&models).Error
	if err != nil {
		return nil, err
	}
	return toShoeCardDomainSlice(models), nil
}

func (r *shoeRepo) UpdatePositions(cards []*shoe.ShoeCard) error {
	// Wrap in a single transaction — one commit for all N position updates
	// instead of N individual auto-committed writes (critical for large shoes).
	return r.db.Transaction(func(tx *gorm.DB) error {
		for _, c := range cards {
			if err := tx.Model(&ShoeCardModel{}).
				Where("id = ?", c.ID.String()).
				Update("position", c.Position).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *shoeRepo) DealCard(cardID uuid.UUID, playerID uuid.UUID) error {
	pid := playerID.String()
	return r.db.Model(&ShoeCardModel{}).
		Where("id = ? AND held_by_player_id IS NULL", cardID.String()).
		Update("held_by_player_id", pid).Error
}

func (r *shoeRepo) ReturnCardsByPlayer(playerID uuid.UUID) error {
	return r.db.Model(&ShoeCardModel{}).
		Where("held_by_player_id = ?", playerID.String()).
		Update("held_by_player_id", nil).Error
}

func (r *shoeRepo) UndealtCount(gameID uuid.UUID) (int, error) {
	var count int64
	err := r.db.Model(&ShoeCardModel{}).
		Where("game_id = ? AND held_by_player_id IS NULL", gameID.String()).
		Count(&count).Error
	return int(count), err
}

func (r *shoeRepo) CountBySuit(gameID uuid.UUID) ([]shoe.SuitCount, error) {
	type row struct {
		Suit  string
		Count int
	}
	var rows []row
	err := r.db.Model(&ShoeCardModel{}).
		Select("suit, count(*) as count").
		Where("game_id = ? AND held_by_player_id IS NULL", gameID.String()).
		Group("suit").Find(&rows).Error
	if err != nil {
		return nil, err
	}

	// Build a map then return in canonical suit order.
	countMap := make(map[shoe.Suit]int)
	for _, r := range rows {
		countMap[shoe.Suit(r.Suit)] = r.Count
	}
	result := make([]shoe.SuitCount, 0, len(shoe.SuitOrder))
	for _, s := range shoe.SuitOrder {
		if c, ok := countMap[s]; ok {
			result = append(result, shoe.SuitCount{Suit: s, Count: c})
		}
	}
	return result, nil
}

func (r *shoeRepo) CountByCard(gameID uuid.UUID) ([]shoe.CardCount, error) {
	type row struct {
		Suit         string
		Face         string
		NumericValue int
		Count        int
	}
	var rows []row
	err := r.db.Model(&ShoeCardModel{}).
		Select("suit, face, numeric_value, count(*) as count").
		Where("game_id = ? AND held_by_player_id IS NULL", gameID.String()).
		Group("suit, face, numeric_value").Find(&rows).Error
	if err != nil {
		return nil, err
	}

	// Build map[suit][face]=count then emit in canonical order.
	type key struct{ suit, face string }
	countMap := make(map[key]int)
	for _, r := range rows {
		countMap[key{r.Suit, r.Face}] = r.Count
	}

	var result []shoe.CardCount
	for _, s := range shoe.SuitOrder {
		for _, f := range shoe.AllFaces {
			k := key{string(s), string(f)}
			if c, ok := countMap[k]; ok {
				result = append(result, shoe.CardCount{
					Suit:         s,
					Face:         f,
					NumericValue: shoe.NumericValue[f],
					Count:        c,
				})
			}
		}
	}
	return result, nil
}

func toShoeCardModel(c *shoe.ShoeCard) ShoeCardModel {
	m := ShoeCardModel{
		ID:           c.ID.String(),
		GameID:       c.GameID.String(),
		Suit:         string(c.Suit),
		Face:         string(c.Face),
		NumericValue: c.NumericValue,
		Position:     c.Position,
	}
	if c.HeldByPlayerID != nil {
		s := c.HeldByPlayerID.String()
		m.HeldByPlayerID = &s
	}
	return m
}

func toShoeCardDomain(m *ShoeCardModel) *shoe.ShoeCard {
	id, _ := uuid.Parse(m.ID)
	gameID, _ := uuid.Parse(m.GameID)
	c := &shoe.ShoeCard{
		ID:           id,
		GameID:       gameID,
		Suit:         shoe.Suit(m.Suit),
		Face:         shoe.Face(m.Face),
		NumericValue: m.NumericValue,
		Position:     m.Position,
	}
	if m.HeldByPlayerID != nil {
		pid, err := uuid.Parse(*m.HeldByPlayerID)
		if err == nil {
			c.HeldByPlayerID = &pid
		}
	}
	return c
}

func toShoeCardDomainSlice(models []ShoeCardModel) []*shoe.ShoeCard {
	cards := make([]*shoe.ShoeCard, len(models))
	for i := range models {
		cards[i] = toShoeCardDomain(&models[i])
	}
	return cards
}

// suppress unused import
var _ = errors.Is

var _ shoe.Repository = (*shoeRepo)(nil)
