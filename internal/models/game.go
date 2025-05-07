package models

import (
	"time"

	"github.com/google/uuid"
)

// Game представляет собой модель игры
type Game struct {
	ID               uuid.UUID `json:"id" db:"id"`
	CreatorID        uint64    `json:"creator_id" db:"creator_id"` // Telegram ID создателя
	Word             string    `json:"word" db:"word"`
	Length           int       `json:"length" db:"length"`
	Difficulty       string    `json:"difficulty" db:"difficulty"`
	MaxTries         int       `json:"max_tries" db:"max_tries"`
	Title            string    `json:"title" db:"title"`
	Description      string    `json:"description" db:"description"`
	MinBet           float64   `json:"min_bet" db:"min_bet"`
	MaxBet           float64   `json:"max_bet" db:"max_bet"`
	RewardMultiplier float64   `json:"reward_multiplier" db:"reward_multiplier"`
	Currency         string    `json:"currency" db:"currency"` // "TON" или "USDT"
	RewardPoolTon    float64   `json:"reward_pool_ton" db:"reward_pool_ton"`
	RewardPoolUsdt   float64   `json:"reward_pool_usdt" db:"reward_pool_usdt"`
	Status           string    `json:"status" db:"status"` // "active" или "inactive"
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
}
