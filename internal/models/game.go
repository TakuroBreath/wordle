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
	GameStatusPending  = "pending"  // Ожидает оплаты депозита
	GameStatusActive   = "active"   // Активна и доступна для игры
	GameStatusInactive = "inactive" // Неактивна (временно выключена создателем)
	GameStatusClosed   = "closed"   // Закрыта (завершена)
)

// Game представляет собой модель игры
type Game struct {
	ID               uuid.UUID `json:"id" db:"id"`
	ShortID          string    `json:"short_id" db:"short_id"`                       // Короткий уникальный ID (6-8 символов)
	CreatorID        uint64    `json:"creator_id" db:"creator_id"`                   // Telegram ID создателя
	Word             string    `json:"word" db:"word"`                               // Загаданное слово
	Length           int       `json:"length" db:"length"`                           // Длина слова
	Difficulty       string    `json:"difficulty" db:"difficulty"`                   // Сложность
	MaxTries         int       `json:"max_tries" db:"max_tries"`                     // Максимум попыток
	TimeLimit        int       `json:"time_limit" db:"time_limit"`                   // Лимит времени в минутах
	Title            string    `json:"title" db:"title"`                             // Название игры
	Description      string    `json:"description" db:"description"`                 // Описание
	MinBet           float64   `json:"min_bet" db:"min_bet"`                         // Минимальная ставка
	MaxBet           float64   `json:"max_bet" db:"max_bet"`                         // Максимальная ставка
	RewardMultiplier float64   `json:"reward_multiplier" db:"reward_multiplier"`     // Мультипликатор награды
	DepositAmount    float64   `json:"deposit_amount" db:"deposit_amount"`           // Размер депозита (должен быть >= max_bet * multiplier)
	Currency         string    `json:"currency" db:"currency"`                       // Валюта (TON или USDT)
	RewardPoolTon    float64   `json:"reward_pool_ton" db:"reward_pool_ton"`         // Пул наград в TON
	RewardPoolUsdt   float64   `json:"reward_pool_usdt" db:"reward_pool_usdt"`       // Пул наград в USDT
	ReservedAmount   float64   `json:"reserved_amount" db:"reserved_amount"`         // Зарезервированная сумма для активных игр
	DepositTxHash    string    `json:"deposit_tx_hash,omitempty" db:"deposit_tx_hash"` // Хеш транзакции депозита
	Status           string    `json:"status" db:"status"`                           // Статус игры
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
}

// GetRequiredDeposit возвращает минимально необходимый депозит для игры
func (g *Game) GetRequiredDeposit() float64 {
	return g.MaxBet * g.RewardMultiplier
}

// GetAvailableRewardPool возвращает доступный пул наград (за вычетом зарезервированного)
func (g *Game) GetAvailableRewardPool() float64 {
	if g.Currency == CurrencyTON {
		return g.RewardPoolTon - g.ReservedAmount
	}
	return g.RewardPoolUsdt - g.ReservedAmount
}

// CanAcceptBet проверяет, может ли игра принять ставку указанного размера
func (g *Game) CanAcceptBet(betAmount float64) bool {
	potentialReward := betAmount * g.RewardMultiplier
	return g.GetAvailableRewardPool() >= potentialReward
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
		TimeLimit        int             `json:"time_limit"`
		Title            string          `json:"title"`
		Description      string          `json:"description"`
		MinBet           float64         `json:"min_bet"`
		MaxBet           float64         `json:"max_bet"`
		RewardMultiplier float64         `json:"reward_multiplier"`
		DepositAmount    float64         `json:"deposit_amount"`
		Currency         string          `json:"currency"`
		RewardPoolTon    float64         `json:"reward_pool_ton"`
		RewardPoolUsdt   float64         `json:"reward_pool_usdt"`
		ReservedAmount   float64         `json:"reserved_amount"`
		DepositTxHash    string          `json:"deposit_tx_hash"`
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
	g.TimeLimit = aux.TimeLimit
	g.Title = aux.Title
	g.Description = aux.Description
	g.MinBet = aux.MinBet
	g.MaxBet = aux.MaxBet
	g.RewardMultiplier = aux.RewardMultiplier
	g.DepositAmount = aux.DepositAmount
	g.Currency = aux.Currency
	g.RewardPoolTon = aux.RewardPoolTon
	g.RewardPoolUsdt = aux.RewardPoolUsdt
	g.ReservedAmount = aux.ReservedAmount
	g.DepositTxHash = aux.DepositTxHash
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
