package models

import (
	"time"

	"github.com/google/uuid"
)

// History представляет собой модель истории игр
type History struct {
	ID        uuid.UUID `json:"id"`
	UserID    uint64    `json:"user_id"` // Telegram ID игрока
	GameID    uuid.UUID `json:"game_id"`
	LobbyID   uuid.UUID `json:"lobby_id"`
	Status    string    `json:"status"` // "creator_win" или "player_win"
	Reward    float64   `json:"reward"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
