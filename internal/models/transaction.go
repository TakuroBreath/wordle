package models

import (
	"time"

	"github.com/google/uuid"
)

// Transaction представляет собой модель транзакции
type Transaction struct {
	ID          uuid.UUID `json:"id"`
	UserID      uint64    `json:"user_id"`     // Telegram ID пользователя
	Type        string    `json:"type"`        // Тип транзакции (deposit, withdraw, bet, reward)
	Amount      float64   `json:"amount"`      // Сумма транзакции
	Description string    `json:"description"` // Описание транзакции
	Status      string    `json:"status"`      // Статус транзакции (pending, completed, failed)
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
