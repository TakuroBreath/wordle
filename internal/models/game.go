package models

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
)

// Статусы игры
const (
	GameStatusInactive          = "inactive"
	GameStatusActive            = "active"
	GameStatusPendingActivation = "pending_activation"
	GameStatusFinished          = "finished"
)

// Game представляет собой модель игры
type Game struct {
	ID               uuid.UUID `json:"id" db:"id"`
	ShortID          string    `json:"short_id" db:"short_id"`
	CreatorID        uint64    `json:"creator_id" db:"creator_id"` // Telegram ID создателя
	Word             string    `json:"word" db:"word"`
	Length           int       `json:"length" db:"length"`
	Difficulty       string    `json:"difficulty" db:"difficulty"`
	MaxTries         int       `json:"max_tries" db:"max_tries"`
	Title            string    `json:"title" db:"title"`
	Description      string    `json:"description" db:"description"`
	MinBet           float64   `json:"min_bet" db:"min_bet"`
	MaxBet           float64   `json:"max_bet" db:"max_bet"`
	RewardMultiplier float64   `json:"reward_multiplier" db:"reward_multiplier"`
	DepositAmount    float64   `json:"deposit_amount" db:"deposit_amount"`
	CommissionRate   float64   `json:"commission_rate" db:"commission_rate"`
	TimeLimitMinutes int       `json:"time_limit_minutes" db:"time_limit_minutes"`
	Currency         string    `json:"currency" db:"currency"` // "TON" или "USDT"
	RewardPoolTon    float64   `json:"reward_pool_ton" db:"reward_pool_ton"`
	RewardPoolUsdt   float64   `json:"reward_pool_usdt" db:"reward_pool_usdt"`
	Status           string    `json:"status" db:"status"` // "active", "inactive", "pending_activation"
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
}

// UnmarshalJSON реализует интерфейс json.Unmarshaler для корректной обработки поля max_tries
// которое может быть как числом, так и строкой
func (g *Game) UnmarshalJSON(data []byte) error {
	// Создаем временную структуру с теми же полями, но max_tries как json.RawMessage
	type Alias struct {
		ID               uuid.UUID       `json:"id"`
		ShortID          string          `json:"short_id"`
		CreatorID        uint64          `json:"creator_id"`
		Word             string          `json:"word"`
		Length           int             `json:"length"`
		Difficulty       string          `json:"difficulty"`
		MaxTries         json.RawMessage `json:"max_tries"`
		Title            string          `json:"title"`
		Description      string          `json:"description"`
		MinBet           float64         `json:"min_bet"`
		MaxBet           float64         `json:"max_bet"`
		RewardMultiplier float64         `json:"reward_multiplier"`
		DepositAmount    float64         `json:"deposit_amount"`
		CommissionRate   float64         `json:"commission_rate"`
		TimeLimitMinutes int             `json:"time_limit_minutes"`
		Currency         string          `json:"currency"`
		RewardPoolTon    float64         `json:"reward_pool_ton"`
		RewardPoolUsdt   float64         `json:"reward_pool_usdt"`
		Status           string          `json:"status"`
		CreatedAt        time.Time       `json:"created_at"`
		UpdatedAt        time.Time       `json:"updated_at"`
	}

	// Анмаршалим данные во временную структуру
	aux := &Alias{}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	// Копируем все поля, кроме max_tries
	g.ID = aux.ID
	g.ShortID = aux.ShortID
	g.CreatorID = aux.CreatorID
	g.Word = aux.Word
	g.Length = aux.Length
	g.Difficulty = aux.Difficulty
	g.Title = aux.Title
	g.Description = aux.Description
	g.MinBet = aux.MinBet
	g.MaxBet = aux.MaxBet
	g.RewardMultiplier = aux.RewardMultiplier
	g.DepositAmount = aux.DepositAmount
	g.CommissionRate = aux.CommissionRate
	g.TimeLimitMinutes = aux.TimeLimitMinutes
	g.Currency = aux.Currency
	g.RewardPoolTon = aux.RewardPoolTon
	g.RewardPoolUsdt = aux.RewardPoolUsdt
	g.Status = aux.Status
	g.CreatedAt = aux.CreatedAt
	g.UpdatedAt = aux.UpdatedAt

	// Если max_tries отсутствует или null, оставляем значение по умолчанию
	if len(aux.MaxTries) == 0 || string(aux.MaxTries) == "null" {
		return nil
	}

	// Пробуем сначала как число
	var maxTries int
	if err := json.Unmarshal(aux.MaxTries, &maxTries); err == nil {
		g.MaxTries = maxTries
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

	g.MaxTries = maxTriesVal
	return nil
}
