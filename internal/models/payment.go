package models

import (
	"time"

	"github.com/google/uuid"
)

// Типы платежей
const (
	PaymentTypeGameDeposit = "game_deposit" // Депозит для создания игры
	PaymentTypeLobbyBet    = "lobby_bet"    // Ставка для вступления в игру
)

// Статусы платежей
const (
	PaymentStatusPending   = "pending"
	PaymentStatusCompleted = "completed"
	PaymentStatusExpired   = "expired"
	PaymentStatusCanceled  = "canceled"
)

// PendingPayment представляет ожидающий платёж
type PendingPayment struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	UserID      uint64     `json:"user_id" db:"user_id"`
	GameID      *uuid.UUID `json:"game_id,omitempty" db:"game_id"`
	GameShortID string     `json:"game_short_id,omitempty" db:"game_short_id"`
	LobbyID     *uuid.UUID `json:"lobby_id,omitempty" db:"lobby_id"`
	PaymentType string     `json:"payment_type" db:"payment_type"`
	Amount      float64    `json:"amount" db:"amount"`
	Currency    string     `json:"currency" db:"currency"`
	Comment     string     `json:"comment" db:"comment"` // Уникальный комментарий для идентификации
	ExpiresAt   time.Time  `json:"expires_at" db:"expires_at"`
	Status      string     `json:"status" db:"status"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
}

// IsExpired проверяет, истёк ли срок платежа
func (p *PendingPayment) IsExpired() bool {
	return time.Now().After(p.ExpiresAt)
}

// BlockchainState хранит состояние синхронизации с блокчейном
type BlockchainState struct {
	ID                 int       `json:"id" db:"id"`
	LastProcessedLt    int64     `json:"last_processed_lt" db:"last_processed_lt"`
	LastProcessedHash  string    `json:"last_processed_hash" db:"last_processed_hash"`
	UpdatedAt          time.Time `json:"updated_at" db:"updated_at"`
}

// GeneratePaymentComment генерирует уникальный комментарий для платежа
// Формат: TYPE_SHORTID_TIMESTAMP
// Например: GD_ABC123_1704067200 (Game Deposit)
// Или: LB_XYZ789_1704067200 (Lobby Bet)
func GeneratePaymentComment(paymentType string, shortID string) string {
	prefix := "GD" // Game Deposit
	if paymentType == PaymentTypeLobbyBet {
		prefix = "LB" // Lobby Bet
	}
	timestamp := time.Now().Unix()
	return prefix + "_" + shortID + "_" + string(rune(timestamp))
}

// ParsePaymentComment парсит комментарий платежа
// Возвращает тип платежа и short ID
func ParsePaymentComment(comment string) (paymentType string, shortID string, ok bool) {
	if len(comment) < 4 {
		return "", "", false
	}
	
	prefix := comment[:2]
	switch prefix {
	case "GD":
		paymentType = PaymentTypeGameDeposit
	case "LB":
		paymentType = PaymentTypeLobbyBet
	default:
		return "", "", false
	}
	
	// Находим short ID между первым и вторым подчёркиванием
	if comment[2] != '_' {
		return "", "", false
	}
	
	rest := comment[3:]
	for i := 0; i < len(rest); i++ {
		if rest[i] == '_' {
			shortID = rest[:i]
			return paymentType, shortID, true
		}
	}
	
	// Если второе подчёркивание не найдено, весь остаток - это short ID
	return paymentType, rest, true
}
