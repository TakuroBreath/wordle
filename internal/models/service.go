package models

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// GameService определяет методы для работы с играми
type GameService interface {
	// CRUD операции
	CreateGame(ctx context.Context, game *Game) error
	GetGame(ctx context.Context, id uuid.UUID) (*Game, error)
	GetGameByShortID(ctx context.Context, shortID string) (*Game, error)
	GetUserGames(ctx context.Context, userID uint64, limit, offset int) ([]*Game, error)
	GetCreatedGames(ctx context.Context, userID uint64, limit, offset int) ([]*Game, error)
	UpdateGame(ctx context.Context, game *Game) error
	DeleteGame(ctx context.Context, id uuid.UUID) error
	
	// Получение списков
	GetActiveGames(ctx context.Context, limit, offset int) ([]*Game, error)
	GetPendingGames(ctx context.Context, limit, offset int) ([]*Game, error)
	SearchGames(ctx context.Context, minBet, maxBet float64, difficulty string, limit, offset int) ([]*Game, error)
	GetGameStats(ctx context.Context, gameID uuid.UUID) (map[string]any, error)
	
	// Управление статусом и балансом
	AddToRewardPool(ctx context.Context, gameID uuid.UUID, amount float64) error
	ActivateGame(ctx context.Context, gameID uuid.UUID) error
	DeactivateGame(ctx context.Context, gameID uuid.UUID) error
	
	// Резервирование средств
	ReserveForBet(ctx context.Context, gameID uuid.UUID, betAmount float64, multiplier float64) error
	ReleaseReservation(ctx context.Context, gameID uuid.UUID, amount float64) error
	
	// Генерация платежной информации
	GetPaymentInfo(ctx context.Context, gameID uuid.UUID) (*PaymentInfo, error)
	
	// Проверка слова (утилиты)
	CheckWord(word, target string) string
	IsWordCorrect(word, target string) bool
	CalculateReward(bet float64, multiplier float64, triesUsed, maxTries int) float64
}

// UserService определяет методы для работы с пользователями
type UserService interface {
	CreateUser(ctx context.Context, user *User) error
	GetUser(ctx context.Context, telegramID uint64) (*User, error)
	UpdateUser(ctx context.Context, user *User) error
	DeleteUser(ctx context.Context, telegramID uint64) error
	GetUserByUsername(ctx context.Context, username string) (*User, error)
	
	// Управление кошельком
	UpdateWallet(ctx context.Context, telegramID uint64, wallet string) error
	GetUserByWallet(ctx context.Context, wallet string) (*User, error)
	
	// Управление балансом
	UpdateTonBalance(ctx context.Context, telegramID uint64, amount float64) error
	UpdateUsdtBalance(ctx context.Context, telegramID uint64, amount float64) error
	ValidateBalance(ctx context.Context, telegramID uint64, requiredAmount float64, currency string) (bool, error)
	
	// Статистика
	GetTopUsers(ctx context.Context, limit int) ([]*User, error)
	GetUserStats(ctx context.Context, telegramID uint64) (map[string]any, error)
	IncrementWins(ctx context.Context, telegramID uint64) error
	IncrementLosses(ctx context.Context, telegramID uint64) error
	
	// Вывод средств
	RequestWithdraw(ctx context.Context, telegramID uint64, amount float64, currency string, toAddress string) (*WithdrawResult, error)
	GetWithdrawHistory(ctx context.Context, telegramID uint64, limit, offset int) ([]*Transaction, error)
	CanWithdraw(ctx context.Context, telegramID uint64) (bool, error)
}

// LobbyService определяет методы для работы с лобби
type LobbyService interface {
	// CRUD операции
	CreateLobby(ctx context.Context, lobby *Lobby) error
	GetLobby(ctx context.Context, id uuid.UUID) (*Lobby, error)
	GetGameLobbies(ctx context.Context, gameID uuid.UUID, limit, offset int) ([]*Lobby, error)
	GetUserLobbies(ctx context.Context, userID uint64, limit, offset int) ([]*Lobby, error)
	UpdateLobby(ctx context.Context, lobby *Lobby) error
	DeleteLobby(ctx context.Context, id uuid.UUID) error
	
	// Управление статусом
	GetActiveLobbies(ctx context.Context, limit, offset int) ([]*Lobby, error)
	UpdateLobbyStatus(ctx context.Context, id uuid.UUID, status string) error
	UpdateTriesUsed(ctx context.Context, id uuid.UUID, triesUsed int) error
	GetExpiredLobbies(ctx context.Context) ([]*Lobby, error)
	GetActiveLobbyByGameAndUser(ctx context.Context, gameID uuid.UUID, userID uint64) (*Lobby, error)
	ExtendLobbyTime(ctx context.Context, id uuid.UUID, duration int) error
	GetLobbyStats(ctx context.Context, lobbyID uuid.UUID) (map[string]any, error)
	
	// Игровой процесс
	ProcessAttempt(ctx context.Context, lobbyID uuid.UUID, word string) ([]int, error)
	FinishLobby(ctx context.Context, lobbyID uuid.UUID, success bool) error
	StartLobby(ctx context.Context, lobbyID uuid.UUID) error
	
	// Генерация платежной информации для вступления в игру
	GetJoinPaymentInfo(ctx context.Context, gameShortID string, userID uint64, betAmount float64) (*PaymentInfo, error)
	
	// Утилиты
	CheckWord(word, target string) []int
	IsWordCorrect(word, target string) bool
	CalculateReward(bet float64, multiplier float64, triesUsed, maxTries int) float64
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
	GetUserStats(ctx context.Context, userID uint64) (map[string]any, error)
	GetGameHistoryStats(ctx context.Context, gameID uuid.UUID) (map[string]any, error)
}

// TransactionService определяет методы для работы с транзакциями
type TransactionService interface {
	// CRUD операции
	CreateTransaction(ctx context.Context, transaction *Transaction) error
	GetTransaction(ctx context.Context, id uuid.UUID) (*Transaction, error)
	GetTransactionByTxHash(ctx context.Context, txHash string) (*Transaction, error)
	GetUserTransactions(ctx context.Context, userID uint64, limit, offset int) ([]*Transaction, error)
	UpdateTransaction(ctx context.Context, transaction *Transaction) error
	DeleteTransaction(ctx context.Context, id uuid.UUID) error
	GetTransactionsByType(ctx context.Context, transactionType string, limit, offset int) ([]*Transaction, error)
	GetUserBalance(ctx context.Context, userID uint64) (float64, error)
	GetTransactionStats(ctx context.Context, userID uint64) (map[string]any, error)
	
	// Обработка транзакций
	ProcessWithdraw(ctx context.Context, userID uint64, amount float64, currency string, toAddress string) (*Transaction, error)
	ProcessDeposit(ctx context.Context, userID uint64, amount float64, currency string, txHash string, network string) error
	ProcessReward(ctx context.Context, userID uint64, amount float64, currency string, gameID *uuid.UUID, lobbyID *uuid.UUID) error
	ConfirmDeposit(ctx context.Context, transactionID uuid.UUID) error
	ConfirmWithdrawal(ctx context.Context, transactionID uuid.UUID, txHash string) error
	FailTransaction(ctx context.Context, transactionID uuid.UUID, reason string) error
	
	// Проверки
	IsTransactionProcessed(ctx context.Context, txHash string) bool
	ExistsByTxHash(ctx context.Context, txHash string) (bool, error)
}

// TONService определяет методы для работы с TON блокчейном
type TONService interface {
	// Получение информации
	GetMasterWalletAddress() string
	GetBalance(ctx context.Context, address string) (float64, error)
	ValidateAddress(address string) bool
	
	// Получение транзакций
	GetTransactions(ctx context.Context, address string, limit int, lt int64, hash string) ([]*BlockchainTransaction, error)
	GetNewTransactions(ctx context.Context, afterLt int64) ([]*BlockchainTransaction, error)
	
	// Отправка транзакций
	SendTON(ctx context.Context, toAddress string, amount float64, comment string) (string, error)
	SendUSDT(ctx context.Context, toAddress string, amount float64, comment string) (string, error)
	
	// Генерация deep link
	GeneratePaymentDeepLink(address string, amount float64, comment string) string
}

// BlockchainWorkerService определяет методы для фоновой обработки блокчейн транзакций
type BlockchainWorkerService interface {
	// Запуск и остановка
	Start(ctx context.Context) error
	Stop() error
	
	// Обработка транзакций
	ProcessIncomingTransactions(ctx context.Context) error
	ProcessPendingWithdrawals(ctx context.Context) error
	
	// Активация игр по депозитам
	ActivateGameByDeposit(ctx context.Context, gameShortID string, amount float64, txHash string) error
	
	// Активация лобби по депозитам  
	ActivateLobbyByDeposit(ctx context.Context, lobbyID uuid.UUID, txHash string) error
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

// CommissionService определяет методы для работы с комиссиями
type CommissionService interface {
	// CalculateCommission рассчитывает комиссию для суммы
	CalculateCommission(amount float64) float64
	// GetCommissionRate возвращает текущую ставку комиссии (0.05 = 5%)
	GetCommissionRate() float64
	// DeductCommission вычитает комиссию и возвращает итоговую сумму
	DeductCommission(amount float64) (netAmount float64, commission float64)
}
