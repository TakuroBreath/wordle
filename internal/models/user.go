package models

import (
	"time"
)

// User представляет собой модель пользователя в системе
type User struct {
	TelegramID  uint64    `json:"telegram_id" db:"telegram_id"`
	Username    string    `json:"username" db:"username"`
	FirstName   string    `json:"first_name" db:"first_name"`
	LastName    string    `json:"last_name" db:"last_name"`
	Wallet      string    `json:"wallet" db:"wallet"`
	BalanceTon  float64   `json:"balance_ton" db:"balance_ton"`
	BalanceUsdt float64   `json:"balance_usdt" db:"balance_usdt"`
	Wins        int       `json:"wins" db:"wins"`
	Losses      int       `json:"losses" db:"losses"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}
