package models

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// User представляет собой модель пользователя в системе
type User struct {
	TelegramID uint64    `json:"telegram_id"`
	Username   string    `json:"username"`
	Wallet     string    `json:"wallet"`
	Balance    float64   `json:"balance"`
	Wins       int       `json:"wins"`
	Losses     int       `json:"losses"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// UserRepository определяет интерфейс для работы с пользователями в базе данных
type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByTelegramID(ctx context.Context, telegramID uint64) (*User, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, telegramID uint64) error
}

// UserService определяет интерфейс для бизнес-логики работы с пользователями
type UserService interface {
	Create(ctx context.Context, user *User) error
	GetByTelegramID(ctx context.Context, telegramID uint64) (*User, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, telegramID uint64) error
}

// Transaction представляет собой модель транзакции в системе
type Transaction struct {
	ID          uuid.UUID  `json:"id"`
	UserID      uint64     `json:"user_id"` // Telegram ID
	Amount      float64    `json:"amount"`
	Currency    string     `json:"currency"`
	Type        string     `json:"type"`   // "deposit", "withdraw", "win", "loss"
	Status      string     `json:"status"` // "pending", "completed", "failed"
	TxHash      string     `json:"tx_hash"`
	CreatedAt   time.Time  `json:"created_at"`
	CompletedAt *time.Time `json:"completed_at"`
}

// TransactionRepository определяет интерфейс для работы с транзакциями в базе данных
type TransactionRepository interface {
	Create(ctx context.Context, tx *Transaction) error
	GetByID(ctx context.Context, id uuid.UUID) (*Transaction, error)
	GetByUserID(ctx context.Context, userID uint64, limit, offset int) ([]*Transaction, error)
	GetByStatus(ctx context.Context, status string, limit, offset int) ([]*Transaction, error)
	Update(ctx context.Context, tx *Transaction) error
}

// TransactionService определяет интерфейс для бизнес-логики работы с транзакциями
type TransactionService interface {
	Create(ctx context.Context, tx *Transaction) error
	GetByID(ctx context.Context, id uuid.UUID) (*Transaction, error)
	GetByUserID(ctx context.Context, userID uint64, limit, offset int) ([]*Transaction, error)
	GetByStatus(ctx context.Context, status string, limit, offset int) ([]*Transaction, error)
	Update(ctx context.Context, tx *Transaction) error
	ProcessDeposit(ctx context.Context, userID uint64, amount float64, currency string, txHash string) error
	ProcessWithdraw(ctx context.Context, userID uint64, amount float64, currency string) (*Transaction, error)
}
