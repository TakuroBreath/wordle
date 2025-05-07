package models

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Game представляет собой модель игры
type Game struct {
	ID          uuid.UUID `json:"id"`
	CreatorID   uint64    `json:"creator_id"` // Telegram ID создателя
	Word        string    `json:"word"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	MinBet      float64   `json:"min_bet"`
	MaxBet      float64   `json:"max_bet"`
	RewardRate  float64   `json:"reward_rate"`
	Status      string    `json:"status"` // "active" или "inactive"
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	ExpiresAt   time.Time `json:"expires_at"`
}

// GameRepository определяет интерфейс для работы с играми в базе данных
type GameRepository interface {
	Create(ctx context.Context, game *Game) error
	GetByID(ctx context.Context, id uuid.UUID) (*Game, error)
	GetAll(ctx context.Context, limit, offset int) ([]*Game, error)
	GetActive(ctx context.Context, limit, offset int) ([]*Game, error)
	GetByCreator(ctx context.Context, creatorID uint64, limit, offset int) ([]*Game, error)
	Update(ctx context.Context, game *Game) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// GameService определяет интерфейс для бизнес-логики работы с играми
type GameService interface {
	Create(ctx context.Context, game *Game) error
	GetByID(ctx context.Context, id uuid.UUID) (*Game, error)
	GetAll(ctx context.Context, limit, offset int) ([]*Game, error)
	GetActive(ctx context.Context, limit, offset int) ([]*Game, error)
	GetByCreator(ctx context.Context, creatorID uint64, limit, offset int) ([]*Game, error)
	Update(ctx context.Context, game *Game) error
	Delete(ctx context.Context, id uuid.UUID) error
	CheckWord(ctx context.Context, gameID uuid.UUID, word string) ([]int, error)
}
