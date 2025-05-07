package models

import (
	"time"

	"github.com/google/uuid"
)

// Attempt представляет собой модель попытки угадать слово
type Attempt struct {
	ID        uuid.UUID `json:"id" db:"id"`
	GameID    uuid.UUID `json:"game_id" db:"game_id"`
	UserID    uint64    `json:"user_id" db:"user_id"` // Telegram ID игрока
	Word      string    `json:"word" db:"word"`
	Result    []int     `json:"result" db:"result"` // 0 - нет, 1 - есть в слове, 2 - на месте
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}
