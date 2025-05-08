package service

import (
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
	Auth() models.AuthService
	Job() models.JobService
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
	authService    models.AuthService
	jobService     models.JobService
	jwtSecret      string
	botToken       string
}

// NewService создает новый экземпляр Service
func NewService(repo repository.Repository, redisRepo repository.RedisRepository, jwtSecret, botToken string) Service {
	service := &ServiceImpl{
		repo:      repo,
		redisRepo: redisRepo,
		jwtSecret: jwtSecret,
		botToken:  botToken,
	}

	// Инициализация сервисов - сначала создаем базовые сервисы
	txService := NewTransactionServiceImpl(repo.Transaction(), repo.User())
	service.txService = txService

	service.userService = NewUserServiceImpl(repo.User(), txService)
	service.gameService = NewGameService(repo.Game(), redisRepo, service.userService)
	service.historyService = NewHistoryService(repo.History(), repo.Game(), repo.User(), repo.Lobby())

	// Создаем лобби-сервис с зависимостями
	service.lobbyService = NewLobbyService(
		repo.Lobby(),
		repo.Game(),
		repo.Attempt(),
		redisRepo,
		service.userService,
		txService,
		service.historyService,
	)

	service.authService = NewAuthService(repo.User(), redisRepo, jwtSecret, botToken)

	// Создаем сервис фоновых задач
	service.jobService = NewJobService(
		service.lobbyService,
		service.txService,
		service.gameService,
		service.userService,
	)

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
func (s *ServiceImpl) Auth() models.AuthService {
	return s.authService
}

// Job возвращает сервис для фоновых задач
func (s *ServiceImpl) Job() models.JobService {
	return s.jobService
}

// Фабричные методы для создания сервисов определены в соответствующих файлах
