package models

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Game представляет собой модель игры Wordle с механиками ставок
type Game struct {
	ID               uuid.UUID `json:"id"`
	Title            string    `json:"title"`
	Word             string    `json:"word"`
	Length           int       `json:"length"`
	Language         string    `json:"language"`   // "en" или "ru"
	CreatorID        uint64    `json:"creator_id"` // Telegram ID создателя
	Difficulty       int       `json:"difficulty"`
	MaxTries         int       `json:"max_tries"`
	RewardMultiplier float64   `json:"reward_multiplier"`
	Currency         string    `json:"currency"` // "TON" или "USDT"
	PrizePool        float64   `json:"prize_pool"`
	MinBet           float64   `json:"min_bet"`
	MaxBet           float64   `json:"max_bet"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	Status           string    `json:"status"` // "active" или "inactive"
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
	GetAllGames(ctx context.Context) ([]*Game, error)
	GetActive(ctx context.Context, limit, offset int) ([]*Game, error)
	GetByCreator(ctx context.Context, creatorID uint64, limit, offset int) ([]*Game, error)
	Update(ctx context.Context, game *Game) error
	Delete(ctx context.Context, id uuid.UUID) error
}
