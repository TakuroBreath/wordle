package models

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

// ErrUserNotFound ошибка, когда пользователь не найден
var ErrUserNotFound = errors.New("user not found")

// ErrGameNotFound ошибка, когда игра не найдена
var ErrGameNotFound = errors.New("game not found")

// ErrLobbyNotFound ошибка, когда лобби не найдено
var ErrLobbyNotFound = errors.New("lobby not found")

// GameRepository определяет методы для работы с играми
type GameRepository interface {
	Create(ctx context.Context, game *Game) error
	GetByID(ctx context.Context, id uuid.UUID) (*Game, error)
	GetByUserID(ctx context.Context, userID uint64, limit, offset int) ([]*Game, error)
	Update(ctx context.Context, game *Game) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetActive(ctx context.Context, limit, offset int) ([]*Game, error)
	GetByStatus(ctx context.Context, status string, limit, offset int) ([]*Game, error)
	CountByUser(ctx context.Context, userID uint64) (int, error)
	GetGameStats(ctx context.Context, gameID uuid.UUID) (map[string]interface{}, error)
	SearchGames(ctx context.Context, minBet, maxBet float64, difficulty string, limit, offset int) ([]*Game, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
	UpdateRewardPool(ctx context.Context, id uuid.UUID, rewardPoolTon, rewardPoolUsdt float64) error
}

// UserRepository определяет методы для работы с пользователями
type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByTelegramID(ctx context.Context, telegramID uint64) (*User, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, telegramID uint64) error
	GetByUsername(ctx context.Context, username string) (*User, error)
	UpdateTonBalance(ctx context.Context, telegramID uint64, amount float64) error
	UpdateUsdtBalance(ctx context.Context, telegramID uint64, amount float64) error
	GetTopUsers(ctx context.Context, limit int) ([]*User, error)
	GetUserStats(ctx context.Context, userID uint64) (map[string]interface{}, error)
	ValidateBalance(ctx context.Context, userID uint64, requiredAmount float64) (bool, error)
}

// LobbyRepository определяет методы для работы с лобби
type LobbyRepository interface {
	Create(ctx context.Context, lobby *Lobby) error
	GetByID(ctx context.Context, id uuid.UUID) (*Lobby, error)
	GetByGameID(ctx context.Context, gameID uuid.UUID, limit, offset int) ([]*Lobby, error)
	GetByUserID(ctx context.Context, userID uint64, limit, offset int) ([]*Lobby, error)
	Update(ctx context.Context, lobby *Lobby) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetActive(ctx context.Context, limit, offset int) ([]*Lobby, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
	UpdateTriesUsed(ctx context.Context, id uuid.UUID, triesUsed int) error
	GetExpired(ctx context.Context) ([]*Lobby, error)
	GetActiveByGameAndUser(ctx context.Context, gameID uuid.UUID, userID uint64) (*Lobby, error)
	CountActiveByGame(ctx context.Context, gameID uuid.UUID) (int, error)
	CountByUser(ctx context.Context, userID uint64) (int, error)
	GetActiveWithAttempts(ctx context.Context, limit, offset int) ([]*Lobby, error)
	ExtendExpirationTime(ctx context.Context, id uuid.UUID, duration int) error
	GetLobbiesByStatus(ctx context.Context, status string, limit, offset int) ([]*Lobby, error)
	GetActiveLobbiesByTimeRange(ctx context.Context, startTime, endTime time.Time, limit, offset int) ([]*Lobby, error)
	GetLobbyStats(ctx context.Context, lobbyID uuid.UUID) (map[string]interface{}, error)
}

// AttemptRepository определяет методы для работы с попытками
type AttemptRepository interface {
	Create(ctx context.Context, attempt *Attempt) error
	GetByID(ctx context.Context, id uuid.UUID) (*Attempt, error)
	GetByLobbyID(ctx context.Context, lobbyID uuid.UUID, limit, offset int) ([]*Attempt, error)
	Update(ctx context.Context, attempt *Attempt) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetLastAttempt(ctx context.Context, lobbyID uuid.UUID, userID uint64) (*Attempt, error)
	CountByLobbyID(ctx context.Context, lobbyID uuid.UUID) (int, error)
	GetByWord(ctx context.Context, lobbyID uuid.UUID, word string) (*Attempt, error)
	GetAttemptStats(ctx context.Context, lobbyID uuid.UUID) (map[string]interface{}, error)
	ValidateWord(ctx context.Context, word string) (bool, error)
}

// HistoryRepository определяет методы для работы с историей
type HistoryRepository interface {
	Create(ctx context.Context, history *History) error
	GetByID(ctx context.Context, id uuid.UUID) (*History, error)
	GetByGameID(ctx context.Context, gameID uuid.UUID, limit, offset int) ([]*History, error)
	GetByUserID(ctx context.Context, userID uint64, limit, offset int) ([]*History, error)
	Update(ctx context.Context, history *History) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByLobbyID(ctx context.Context, lobbyID uuid.UUID) (*History, error)
	GetByStatus(ctx context.Context, status string, limit, offset int) ([]*History, error)
	CountByUser(ctx context.Context, userID uint64) (int, error)
	GetUserStats(ctx context.Context, userID uint64) (map[string]interface{}, error)
	GetGameHistoryStats(ctx context.Context, gameID uuid.UUID) (map[string]interface{}, error)
}

// TransactionRepository определяет методы для работы с транзакциями
type TransactionRepository interface {
	Create(ctx context.Context, transaction *Transaction) error
	GetByID(ctx context.Context, id uuid.UUID) (*Transaction, error)
	GetByUserID(ctx context.Context, userID uint64, limit, offset int) ([]*Transaction, error)
	Update(ctx context.Context, transaction *Transaction) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByType(ctx context.Context, transactionType string, limit, offset int) ([]*Transaction, error)
	CountByUser(ctx context.Context, userID uint64) (int, error)
	GetUserBalance(ctx context.Context, userID uint64) (float64, error)
	GetTransactionStats(ctx context.Context, userID uint64) (map[string]interface{}, error)
	CheckSufficientFunds(ctx context.Context, userID uint64, amount float64) (bool, error)
}
