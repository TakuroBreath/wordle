package models

import (
	"time"

	"github.com/google/uuid"
)

// Lobby представляет собой модель игрового лобби
type Lobby struct {
	ID              uuid.UUID `json:"id" db:"id"`
	GameID          uuid.UUID `json:"game_id" db:"game_id"`
	UserID          uint64    `json:"user_id" db:"user_id"` // Telegram ID игрока
	MaxTries        int       `json:"max_tries" db:"max_tries"`
	TriesUsed       int       `json:"tries_used" db:"tries_used"`
	BetAmount       float64   `json:"bet_amount" db:"bet_amount"`
	PotentialReward float64   `json:"potential_reward" db:"potential_reward"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
	ExpiresAt       time.Time `json:"expires_at" db:"expires_at"` // 5 минут после создания
	Status          string    `json:"status" db:"status"` // "active" или "inactive"
	Attempts        []Attempt `json:"attempts"`
}
