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
	LobbyStatusActive         = "active"          // Лобби активно
	LobbyStatusSuccess        = "success"         // Игрок угадал слово
	LobbyStatusFailedTries    = "failed_tries"    // Попытки исчерпаны
	LobbyStatusFailedExpired  = "failed_expired"  // Время истекло
	LobbyStatusFailedInternal = "failed_internal" // Внутренняя ошибка (напр., игра не найдена)
	LobbyStatusCanceled       = "canceled"        // Отменено (например, игроком или админом)
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
	Status          string    `json:"status" db:"status"`         // Статус лобби
	Attempts        []Attempt `json:"attempts,omitempty"`         // Список попыток, опускаем если пуст
}

// UnmarshalJSON реализует интерфейс json.Unmarshaler для корректной обработки поля max_tries
// которое может быть как числом, так и строкой
func (l *Lobby) UnmarshalJSON(data []byte) error {
	// Создаем временную структуру с теми же полями, но max_tries как json.RawMessage
	type Alias struct {
		ID              uuid.UUID       `json:"id"`
		GameID          uuid.UUID       `json:"game_id"`
		UserID          uint64          `json:"user_id"`
		MaxTries        json.RawMessage `json:"max_tries"`
		TriesUsed       int             `json:"tries_used"`
		BetAmount       float64         `json:"bet_amount"`
		PotentialReward float64         `json:"potential_reward"`
		CreatedAt       time.Time       `json:"created_at"`
		UpdatedAt       time.Time       `json:"updated_at"`
		ExpiresAt       time.Time       `json:"expires_at"`
		Status          string          `json:"status"`
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
	l.UserID = aux.UserID
	l.TriesUsed = aux.TriesUsed
	l.BetAmount = aux.BetAmount
	l.PotentialReward = aux.PotentialReward
	l.CreatedAt = aux.CreatedAt
	l.UpdatedAt = aux.UpdatedAt
	l.ExpiresAt = aux.ExpiresAt
	l.Status = aux.Status
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
