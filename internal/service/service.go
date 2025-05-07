package service

import (
	"context"

	"github.com/TakuroBreath/wordle/internal/models"
	"github.com/TakuroBreath/wordle/internal/repository"
)

// Service представляет собой интерфейс для всех сервисов в приложении
type Service interface {
	Game() models.GameService
	User() models.UserService
	Lobby() models.LobbyService
	History() models.HistoryService
	Transaction() models.TransactionService
	Auth() AuthService
}

// AuthService определяет интерфейс для аутентификации пользователей
type AuthService interface {
	InitAuth(ctx context.Context, initData string) (string, error)
	VerifyAuth(ctx context.Context, token string) (models.User, error)
	GenerateToken(ctx context.Context, user models.User) (string, error)
	ValidateToken(ctx context.Context, token string) (models.User, error)
}

// ServiceImpl представляет собой реализацию сервисного слоя
type ServiceImpl struct {
	repo           repository.Repository
	redisRepo      repository.RedisRepository
	gameService    models.GameService
	userService    models.UserService
	lobbyService   models.LobbyService
	historyService models.HistoryService
	txService      models.TransactionService
	authService    AuthService
}

// NewService создает новый экземпляр Service
func NewService(repo repository.Repository, redisRepo repository.RedisRepository) Service {
	service := &ServiceImpl{
		repo:      repo,
		redisRepo: redisRepo,
	}

	// Инициализация сервисов
	service.gameService = NewGameService(repo.Game(), redisRepo)
	service.userService = NewUserService(repo.User())
	service.lobbyService = NewLobbyService(repo.Lobby(), repo.Game(), repo.User(), repo.Attempt(), redisRepo)
	service.historyService = NewHistoryService(repo.History())
	service.txService = NewTransactionService(repo.Transaction(), repo.User())
	service.authService = NewAuthService(repo.User(), redisRepo)

	return service
}

// Game возвращает сервис для работы с играми
func (s *ServiceImpl) Game() models.GameService {
	return s.gameService
}

// User возвращает сервис для работы с пользователями
func (s *ServiceImpl) User() models.UserService {
	return s.userService
}

// Lobby возвращает сервис для работы с лобби
func (s *ServiceImpl) Lobby() models.LobbyService {
	return s.lobbyService
}

// History возвращает сервис для работы с историей
func (s *ServiceImpl) History() models.HistoryService {
	return s.historyService
}

// Transaction возвращает сервис для работы с транзакциями
func (s *ServiceImpl) Transaction() models.TransactionService {
	return s.txService
}

// Auth возвращает сервис для работы с аутентификацией
func (s *ServiceImpl) Auth() AuthService {
	return s.authService
}

// Фабричные методы для создания сервисов

// NewUserService определен в user_service.go
// NewLobbyService определен в lobby_service.go
// NewHistoryService определен в history_service.go
// NewTransactionService определен в user_service.go

// NewAuthService создает новый экземпляр AuthService
