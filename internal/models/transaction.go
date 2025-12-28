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
	TransactionTypeDeposit       = "deposit"        // Депозит на баланс
	TransactionTypeWithdraw      = "withdraw"       // Вывод средств
	TransactionTypeBet           = "bet"            // Ставка в игре
	TransactionTypeReward        = "reward"         // Награда за победу
	TransactionTypeCommission    = "commission"     // Комиссия сервиса
	TransactionTypeRefund        = "refund"         // Возврат средств
	TransactionTypeGameDeposit   = "game_deposit"   // Депозит для создания игры
	TransactionTypeGameRefund    = "game_refund"    // Возврат депозита игры при закрытии
	TransactionTypeReserve       = "reserve"        // Резервирование средств
	TransactionTypeReleaseReserve = "release_reserve" // Освобождение резерва
)

// Статусы транзакций
const (
	TransactionStatusPending    = "pending"    // Ожидает подтверждения
	TransactionStatusConfirming = "confirming" // Подтверждается в блокчейне
	TransactionStatusCompleted  = "completed"  // Завершена
	TransactionStatusFailed     = "failed"     // Ошибка
	TransactionStatusCanceled   = "canceled"   // Отменена
)

// Transaction представляет собой модель транзакции
type Transaction struct {
	ID            uuid.UUID  `json:"id" db:"id"`
	UserID        uint64     `json:"user_id" db:"user_id"`                                 // Telegram ID пользователя
	Type          string     `json:"type" db:"type"`                                       // Тип транзакции
	Amount        float64    `json:"amount" db:"amount"`                                   // Сумма транзакции
	Fee           float64    `json:"fee,omitempty" db:"fee"`                               // Комиссия
	Currency      string     `json:"currency" db:"currency"`                               // Валюта
	Description   string     `json:"description,omitempty" db:"description"`               // Описание
	Status        string     `json:"status" db:"status"`                                   // Статус
	TxHash        string     `json:"tx_hash,omitempty" db:"tx_hash"`                       // Хеш транзакции в блокчейне
	BlockchainLt  int64      `json:"blockchain_lt,omitempty" db:"blockchain_lt"`           // Logical time транзакции в TON
	FromAddress   string     `json:"from_address,omitempty" db:"from_address"`             // Адрес отправителя
	ToAddress     string     `json:"to_address,omitempty" db:"to_address"`                 // Адрес получателя
	Comment       string     `json:"comment,omitempty" db:"comment"`                       // Комментарий к транзакции
	Network       string     `json:"network,omitempty" db:"network"`                       // Сеть
	GameID        *uuid.UUID `json:"game_id,omitempty" db:"game_id"`                       // ID игры
	GameShortID   string     `json:"game_short_id,omitempty" db:"game_short_id"`           // Короткий ID игры
	LobbyID       *uuid.UUID `json:"lobby_id,omitempty" db:"lobby_id"`                     // ID лобби
	ErrorMessage  string     `json:"error_message,omitempty" db:"error_message"`           // Сообщение об ошибке
	Confirmations int        `json:"confirmations,omitempty" db:"confirmations"`           // Количество подтверждений
	ProcessedAt   *time.Time `json:"processed_at,omitempty" db:"processed_at"`             // Время обработки
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
}

// BlockchainTransaction представляет транзакцию из блокчейна
type BlockchainTransaction struct {
	Hash        string    `json:"hash"`
	Lt          int64     `json:"lt"`          // Logical time
	FromAddress string    `json:"from_address"`
	ToAddress   string    `json:"to_address"`
	Amount      float64   `json:"amount"`      // В основных единицах (TON, не nanoTON)
	Fee         float64   `json:"fee"`
	Comment     string    `json:"comment"`     // Комментарий (memo)
	Currency    string    `json:"currency"`
	Timestamp   time.Time `json:"timestamp"`
	IsIncoming  bool      `json:"is_incoming"` // Входящая или исходящая
}

// PaymentInfo содержит информацию для совершения платежа
type PaymentInfo struct {
	Address    string  `json:"address"`     // Адрес для оплаты
	Amount     float64 `json:"amount"`      // Сумма
	Currency   string  `json:"currency"`    // Валюта
	Comment    string  `json:"comment"`     // Комментарий для идентификации
	GameID     string  `json:"game_id"`     // ID игры (короткий)
	ExpireAt   int64   `json:"expire_at"`   // Время истечения платежа
	DeepLink   string  `json:"deep_link"`   // Deep link для TON кошелька
}

// WithdrawRequest запрос на вывод средств
type WithdrawRequest struct {
	UserID    uint64  `json:"user_id"`
	Amount    float64 `json:"amount"`
	Currency  string  `json:"currency"`
	ToAddress string  `json:"to_address"`
}

// WithdrawResult результат вывода средств
type WithdrawResult struct {
	TransactionID uuid.UUID `json:"transaction_id"`
	TxHash        string    `json:"tx_hash"`
	Status        string    `json:"status"`
	Fee           float64   `json:"fee"`
	Error         string    `json:"error,omitempty"`
}
