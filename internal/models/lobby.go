package models

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
)

// Статусы Лобби
const (
	LobbyStatusPending        = "pending"          // Ожидает оплаты
	LobbyStatusActive         = "active"           // Лобби активно, игрок играет
	LobbyStatusSuccess        = "success"          // Игрок угадал слово
	LobbyStatusFailedTries    = "failed_tries"     // Попытки исчерпаны
	LobbyStatusFailedExpired  = "failed_expired"   // Время истекло
	LobbyStatusFailedInternal = "failed_internal"  // Внутренняя ошибка
	LobbyStatusCanceled       = "canceled"         // Отменено
)

// Lobby представляет собой модель игрового лобби (игровая сессия)
type Lobby struct {
	ID              uuid.UUID `json:"id" db:"id"`
	GameID          uuid.UUID `json:"game_id" db:"game_id"`
	GameShortID     string    `json:"game_short_id" db:"game_short_id"`           // Короткий ID игры
	UserID          uint64    `json:"user_id" db:"user_id"`                       // Telegram ID игрока
	MaxTries        int       `json:"max_tries" db:"max_tries"`                   // Максимум попыток
	TriesUsed       int       `json:"tries_used" db:"tries_used"`                 // Использовано попыток
	BetAmount       float64   `json:"bet_amount" db:"bet_amount"`                 // Размер ставки
	PotentialReward float64   `json:"potential_reward" db:"potential_reward"`     // Потенциальная награда
	PaymentTxHash   string    `json:"payment_tx_hash,omitempty" db:"payment_tx_hash"` // Хеш транзакции оплаты
	Currency        string    `json:"currency" db:"currency"`                     // Валюта
	Status          string    `json:"status" db:"status"`                         // Статус лобби
	StartedAt       *time.Time `json:"started_at,omitempty" db:"started_at"`      // Время начала игры
	ExpiresAt       time.Time `json:"expires_at" db:"expires_at"`                 // Время истечения
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
	Attempts        []Attempt `json:"attempts,omitempty"`                         // Список попыток
}

// IsExpired проверяет, истекло ли время лобби
func (l *Lobby) IsExpired() bool {
	return time.Now().After(l.ExpiresAt)
}

// CanMakeAttempt проверяет, может ли игрок сделать попытку
func (l *Lobby) CanMakeAttempt() bool {
	if l.Status != LobbyStatusActive {
		return false
	}
	if l.IsExpired() {
		return false
	}
	if l.TriesUsed >= l.MaxTries {
		return false
	}
	return true
}

// GetRemainingTime возвращает оставшееся время в секундах
func (l *Lobby) GetRemainingTime() int64 {
	remaining := l.ExpiresAt.Sub(time.Now())
	if remaining < 0 {
		return 0
	}
	return int64(remaining.Seconds())
}

// UnmarshalJSON реализует интерфейс json.Unmarshaler для корректной обработки поля max_tries
// которое может быть как числом, так и строкой
func (l *Lobby) UnmarshalJSON(data []byte) error {
	// Создаем временную структуру с теми же полями, но max_tries как json.RawMessage
	type Alias struct {
		ID              uuid.UUID       `json:"id"`
		GameID          uuid.UUID       `json:"game_id"`
		GameShortID     string          `json:"game_short_id"`
		UserID          uint64          `json:"user_id"`
		MaxTries        json.RawMessage `json:"max_tries"`
		TriesUsed       int             `json:"tries_used"`
		BetAmount       float64         `json:"bet_amount"`
		PotentialReward float64         `json:"potential_reward"`
		PaymentTxHash   string          `json:"payment_tx_hash"`
		Currency        string          `json:"currency"`
		Status          string          `json:"status"`
		StartedAt       *time.Time      `json:"started_at"`
		ExpiresAt       time.Time       `json:"expires_at"`
		CreatedAt       time.Time       `json:"created_at"`
		UpdatedAt       time.Time       `json:"updated_at"`
		Attempts        []Attempt       `json:"attempts,omitempty"`
	}

	// Анмаршалим данные во временную структуру
	aux := &Alias{}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	// Копируем все поля, кроме max_tries
	l.ID = aux.ID
	l.GameID = aux.GameID
	l.GameShortID = aux.GameShortID
	l.UserID = aux.UserID
	l.TriesUsed = aux.TriesUsed
	l.BetAmount = aux.BetAmount
	l.PotentialReward = aux.PotentialReward
	l.PaymentTxHash = aux.PaymentTxHash
	l.Currency = aux.Currency
	l.Status = aux.Status
	l.StartedAt = aux.StartedAt
	l.ExpiresAt = aux.ExpiresAt
	l.CreatedAt = aux.CreatedAt
	l.UpdatedAt = aux.UpdatedAt
	l.Attempts = aux.Attempts

	// Если max_tries отсутствует или null, оставляем значение по умолчанию
	if len(aux.MaxTries) == 0 || string(aux.MaxTries) == "null" {
		return nil
	}

	// Пробуем сначала как число
	var maxTries int
	if err := json.Unmarshal(aux.MaxTries, &maxTries); err == nil {
		l.MaxTries = maxTries
		return nil
	}

	// Если не получилось как число, пробуем как строку
	var maxTriesStr string
	if err := json.Unmarshal(aux.MaxTries, &maxTriesStr); err != nil {
		return fmt.Errorf("max_tries field cannot be parsed: %w", err)
	}

	// Преобразуем строку в число
	maxTriesVal, err := strconv.Atoi(maxTriesStr)
	if err != nil {
		return fmt.Errorf("max_tries is not a valid integer: %w", err)
	}

	l.MaxTries = maxTriesVal
	return nil
}
