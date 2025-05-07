package models

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Lobby представляет собой модель игрового лобби
type Lobby struct {
	ID              uuid.UUID `json:"id"`
	GameID          uuid.UUID `json:"game_id"`
	UserID          uint64    `json:"user_id"` // Telegram ID игрока
	MaxTries        int       `json:"max_tries"`
	TriesUsed       int       `json:"tries_used"`
	BetAmount       float64   `json:"bet_amount"`
	PotentialReward float64   `json:"potential_reward"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	ExpiresAt       time.Time `json:"expires_at"`
	Status          string    `json:"status"` // "active" или "inactive"
	Attempts        []Attempt `json:"attempts"`
}

// Attempt представляет собой модель попытки угадать слово
type Attempt struct {
	ID        uuid.UUID `json:"id"`
	GameID    uuid.UUID `json:"game_id"`
	UserID    uint64    `json:"user_id"` // Telegram ID игрока
	Word      string    `json:"word"`
	Result    []int     `json:"result"` // 0 - нет, 1 - есть в слове, 2 - на месте
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// History представляет собой модель истории игр
type History struct {
	ID        uuid.UUID `json:"id"`
	UserID    uint64    `json:"user_id"` // Telegram ID игрока
	GameID    uuid.UUID `json:"game_id"`
	LobbyID   uuid.UUID `json:"lobby_id"`
	Status    string    `json:"status"` // "creator_win" или "player_win"
	Reward    float64   `json:"reward"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// LobbyRepository определяет интерфейс для работы с лобби в базе данных
type LobbyRepository interface {
	Create(ctx context.Context, lobby *Lobby) error
	GetByID(ctx context.Context, id uuid.UUID) (*Lobby, error)
	GetByUserID(ctx context.Context, userID uint64, limit, offset int) ([]*Lobby, error)
	GetByGameID(ctx context.Context, gameID uuid.UUID, limit, offset int) ([]*Lobby, error)
	GetActive(ctx context.Context, limit, offset int) ([]*Lobby, error)
	Update(ctx context.Context, lobby *Lobby) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// LobbyService определяет интерфейс для бизнес-логики работы с лобби
type LobbyService interface {
	Create(ctx context.Context, lobby *Lobby) error
	GetByID(ctx context.Context, id uuid.UUID) (*Lobby, error)
	GetByUserID(ctx context.Context, userID uint64, limit, offset int) ([]*Lobby, error)
	GetByGameID(ctx context.Context, gameID uuid.UUID, limit, offset int) ([]*Lobby, error)
	GetActive(ctx context.Context, limit, offset int) ([]*Lobby, error)
	Update(ctx context.Context, lobby *Lobby) error
	Delete(ctx context.Context, id uuid.UUID) error
	JoinGame(ctx context.Context, gameID uuid.UUID, userID uint64, bet float64) (*Lobby, error)
	MakeAttempt(ctx context.Context, lobbyID uuid.UUID, word string) (*Attempt, error)
}

// AttemptRepository определяет интерфейс для работы с попытками в базе данных
type AttemptRepository interface {
	Create(ctx context.Context, attempt *Attempt) error
	GetByID(ctx context.Context, id uuid.UUID) (*Attempt, error)
	GetByGameID(ctx context.Context, gameID uuid.UUID) ([]*Attempt, error)
	GetByUserID(ctx context.Context, userID uint64) ([]*Attempt, error)
	GetByLobbyID(ctx context.Context, lobbyID uuid.UUID) ([]*Attempt, error)
	Update(ctx context.Context, attempt *Attempt) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// HistoryRepository определяет интерфейс для работы с историей в базе данных
type HistoryRepository interface {
	Create(ctx context.Context, history *History) error
	GetByID(ctx context.Context, id uuid.UUID) (*History, error)
	GetByUserID(ctx context.Context, userID uint64, limit, offset int) ([]*History, error)
	GetByGameID(ctx context.Context, gameID uuid.UUID, limit, offset int) ([]*History, error)
	Update(ctx context.Context, history *History) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// HistoryService определяет интерфейс для бизнес-логики работы с историей
type HistoryService interface {
	Create(ctx context.Context, history *History) error
	GetByID(ctx context.Context, id uuid.UUID) (*History, error)
	GetByUserID(ctx context.Context, userID uint64, limit, offset int) ([]*History, error)
	GetByGameID(ctx context.Context, gameID uuid.UUID, limit, offset int) ([]*History, error)
}
