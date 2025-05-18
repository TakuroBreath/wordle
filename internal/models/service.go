package models

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// GameService определяет методы для работы с играми
type GameService interface {
	CreateGame(ctx context.Context, game *Game) error
	GetGame(ctx context.Context, id uuid.UUID) (*Game, error)
	GetUserGames(ctx context.Context, userID uint64, limit, offset int) ([]*Game, error)
	GetCreatedGames(ctx context.Context, userID uint64, limit, offset int) ([]*Game, error)
	UpdateGame(ctx context.Context, game *Game) error
	DeleteGame(ctx context.Context, id uuid.UUID) error
	GetActiveGames(ctx context.Context, limit, offset int) ([]*Game, error)
	SearchGames(ctx context.Context, minBet, maxBet float64, difficulty string, limit, offset int) ([]*Game, error)
	GetGameStats(ctx context.Context, gameID uuid.UUID) (map[string]interface{}, error)
	AddToRewardPool(ctx context.Context, gameID uuid.UUID, amount float64) error
	ActivateGame(ctx context.Context, gameID uuid.UUID) error
	DeactivateGame(ctx context.Context, gameID uuid.UUID) error
}

// UserService определяет методы для работы с пользователями
type UserService interface {
	CreateUser(ctx context.Context, user *User) error
	GetUser(ctx context.Context, telegramID uint64) (*User, error)
	UpdateUser(ctx context.Context, user *User) error
	DeleteUser(ctx context.Context, telegramID uint64) error
	GetUserByUsername(ctx context.Context, username string) (*User, error)
	UpdateTonBalance(ctx context.Context, telegramID uint64, amount float64) error
	UpdateUsdtBalance(ctx context.Context, telegramID uint64, amount float64) error
	GetTopUsers(ctx context.Context, limit int) ([]*User, error)
	GetUserStats(ctx context.Context, telegramID uint64) (map[string]interface{}, error)
	ValidateBalance(ctx context.Context, telegramID uint64, requiredAmount float64, currency string) (bool, error)
	RequestWithdraw(ctx context.Context, telegramID uint64, amount float64, currency string) error
	GetWithdrawHistory(ctx context.Context, telegramID uint64, limit, offset int) ([]*Transaction, error)
}

// LobbyService определяет методы для работы с лобби
type LobbyService interface {
	CreateLobby(ctx context.Context, lobby *Lobby) error
	GetLobby(ctx context.Context, id uuid.UUID) (*Lobby, error)
	GetGameLobbies(ctx context.Context, gameID uuid.UUID, limit, offset int) ([]*Lobby, error)
	GetUserLobbies(ctx context.Context, userID uint64, limit, offset int) ([]*Lobby, error)
	UpdateLobby(ctx context.Context, lobby *Lobby) error
	DeleteLobby(ctx context.Context, id uuid.UUID) error
	GetActiveLobbies(ctx context.Context, limit, offset int) ([]*Lobby, error)
	UpdateLobbyStatus(ctx context.Context, id uuid.UUID, status string) error
	UpdateTriesUsed(ctx context.Context, id uuid.UUID, triesUsed int) error
	GetExpiredLobbies(ctx context.Context) ([]*Lobby, error)
	GetActiveLobbyByGameAndUser(ctx context.Context, gameID uuid.UUID, userID uint64) (*Lobby, error)
	ExtendLobbyTime(ctx context.Context, id uuid.UUID, duration int) error
	GetLobbyStats(ctx context.Context, lobbyID uuid.UUID) (map[string]interface{}, error)
	ProcessAttempt(ctx context.Context, lobbyID uuid.UUID, word string) ([]int, error)
	FinishLobby(ctx context.Context, lobbyID uuid.UUID, success bool) error
}

// HistoryService определяет методы для работы с историей
type HistoryService interface {
	CreateHistory(ctx context.Context, history *History) error
	GetHistory(ctx context.Context, id uuid.UUID) (*History, error)
	GetGameHistory(ctx context.Context, gameID uuid.UUID, limit, offset int) ([]*History, error)
	GetUserHistory(ctx context.Context, userID uint64, limit, offset int) ([]*History, error)
	UpdateHistory(ctx context.Context, history *History) error
	DeleteHistory(ctx context.Context, id uuid.UUID) error
	GetLobbyHistory(ctx context.Context, lobbyID uuid.UUID) (*History, error)
	GetHistoryByStatus(ctx context.Context, status string, limit, offset int) ([]*History, error)
	GetUserStats(ctx context.Context, userID uint64) (map[string]interface{}, error)
	GetGameHistoryStats(ctx context.Context, gameID uuid.UUID) (map[string]interface{}, error)
}

// TransactionService определяет методы для работы с транзакциями
type TransactionService interface {
	CreateTransaction(ctx context.Context, transaction *Transaction) error
	GetTransaction(ctx context.Context, id uuid.UUID) (*Transaction, error)
	GetUserTransactions(ctx context.Context, userID uint64, limit, offset int) ([]*Transaction, error)
	UpdateTransaction(ctx context.Context, transaction *Transaction) error
	DeleteTransaction(ctx context.Context, id uuid.UUID) error
	GetTransactionsByType(ctx context.Context, transactionType string, limit, offset int) ([]*Transaction, error)
	GetUserBalance(ctx context.Context, userID uint64 /*, currency string опционально, если хотим баланс конкретной валюты */) (float64, error)
	GetTransactionStats(ctx context.Context, userID uint64) (map[string]interface{}, error)
	ProcessWithdraw(ctx context.Context, userID uint64, amount float64, currency string) error
	ProcessDeposit(ctx context.Context, userID uint64, amount float64, currency string, txHash string, network string) error
	ProcessReward(ctx context.Context, userID uint64, amount float64, currency string, gameID *uuid.UUID, lobbyID *uuid.UUID) error
	ConfirmDeposit(ctx context.Context, transactionID uuid.UUID) error
	ConfirmWithdrawal(ctx context.Context, transactionID uuid.UUID) error
	FailTransaction(ctx context.Context, transactionID uuid.UUID, reason string) error

	// Новые методы для интеграции с блокчейном TON
	VerifyTonTransaction(ctx context.Context, txHash string) (bool, error)
	GenerateTonWithdrawTransaction(ctx context.Context, userID uint64, amount float64, walletAddress string) (map[string]interface{}, error)
	ProcessTonDeposit(ctx context.Context, userID uint64, amount float64, txHash string) error
	GenerateTonWalletAddress(ctx context.Context, userID uint64) (string, error)
	MonitorPendingWithdrawals(ctx context.Context) error
	IsTonTransactionProcessed(ctx context.Context, hash string) bool
}

// AuthService определяет методы для аутентификации и авторизации пользователей
type AuthService interface {
	InitAuth(ctx context.Context, initData string) (string, error)
	VerifyAuth(ctx context.Context, tokenString string) (*User, error)
	GenerateToken(ctx context.Context, user User) (string, error)
	ValidateToken(ctx context.Context, tokenString string) (*User, error)
	Logout(ctx context.Context, tokenString string) error
}

// JobService определяет методы для обработки фоновых задач
type JobService interface {
	ProcessExpiredLobbies(ctx context.Context) error
	ProcessPendingTransactions(ctx context.Context) error
	StartJobScheduler(ctx context.Context, lobbyCheckInterval, transactionCheckInterval time.Duration)
	RunOnce(ctx context.Context) error
}
