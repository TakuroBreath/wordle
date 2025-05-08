package models

import (
	"time"

	"github.com/google/uuid"
)

// Типы валют
const (
	CurrencyTON  = "TON"
	CurrencyUSDT = "USDT"
)

// Типы транзакций
const (
	TransactionTypeDeposit    = "deposit"
	TransactionTypeWithdraw   = "withdraw"
	TransactionTypeBet        = "bet"
	TransactionTypeReward     = "reward"
	TransactionTypeCommission = "commission"
	TransactionTypeRefund     = "refund"
)

// Статусы транзакций
const (
	TransactionStatusPending   = "pending"
	TransactionStatusCompleted = "completed"
	TransactionStatusFailed    = "failed"
	TransactionStatusCanceled  = "canceled"
)

// Transaction представляет собой модель транзакции
type Transaction struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	UserID      uint64     `json:"user_id" db:"user_id"`                   // Telegram ID пользователя
	Type        string     `json:"type" db:"type"`                         // Тип транзакции (deposit, withdraw, bet, reward)
	Amount      float64    `json:"amount" db:"amount"`                     // Сумма транзакции
	Currency    string     `json:"currency" db:"currency"`                 // Валюта транзакции (TON, USDT)
	Description string     `json:"description,omitempty" db:"description"` // Описание транзакции
	Status      string     `json:"status" db:"status"`                     // Статус транзакции (pending, completed, failed)
	TxHash      string     `json:"tx_hash,omitempty" db:"tx_hash"`         // Хеш транзакции в блокчейне (для deposit/withdraw)
	Network     string     `json:"network,omitempty" db:"network"`         // Сеть (для deposit/withdraw)
	GameID      *uuid.UUID `json:"game_id,omitempty" db:"game_id"`         // ID игры, если транзакция связана с игрой
	LobbyID     *uuid.UUID `json:"lobby_id,omitempty" db:"lobby_id"`       // ID лобби, если транзакция связана с лобби
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}
