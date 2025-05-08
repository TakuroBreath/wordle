package models

import (
	"time"

	"github.com/google/uuid"
)

// Статусы записи истории
const (
	HistoryStatusPlayerWin  = "player_win"  // Игрок выиграл
	HistoryStatusCreatorWin = "creator_win" // Создатель выиграл (игрок проиграл)
)

// History представляет собой модель истории игр
type History struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UserID    uint64    `json:"user_id" db:"user_id"` // Telegram ID игрока
	GameID    uuid.UUID `json:"game_id" db:"game_id"`
	LobbyID   uuid.UUID `json:"lobby_id" db:"lobby_id"`     // Связь с конкретным лобби
	Status    string    `json:"status" db:"status"`         // Статус завершения (player_win, creator_win)
	BetAmount float64   `json:"bet_amount" db:"bet_amount"` // Сумма ставки
	Reward    float64   `json:"reward" db:"reward"`         // Сумма выигрыша (0 при проигрыше)
	Currency  string    `json:"currency" db:"currency"`     // Валюта ставки/выигрыша
	TriesUsed int       `json:"tries_used" db:"tries_used"` // Количество использованных попыток
	// WordGuessed string    `json:"word_guessed,omitempty" db:"word_guessed"` // Опционально: последнее слово игрока
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}
