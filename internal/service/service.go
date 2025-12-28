package service

import (
	"github.com/TakuroBreath/wordle/internal/blockchain"
	"github.com/TakuroBreath/wordle/internal/blockchain/ethereum"
	"github.com/TakuroBreath/wordle/internal/blockchain/mock"
	"github.com/TakuroBreath/wordle/internal/blockchain/ton"
	"github.com/TakuroBreath/wordle/internal/config"
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
	BlockchainProvider() blockchain.BlockchainProvider
	TONService() models.TONService
}

// ServiceImpl представляет собой реализацию сервисного слоя
type ServiceImpl struct {
	repo               repository.Repository
	redisRepo          repository.RedisRepository
	gameService        models.GameService
	userService        models.UserService
	lobbyService       models.LobbyService
	historyService     models.HistoryService
	txService          models.TransactionService
	authService        models.AuthService
	jobService         models.JobService
	blockchainProvider blockchain.BlockchainProvider
	tonService         models.TONService
	jwtSecret          string
	botToken           string
	commissionRate     float64
}

// ServiceConfig конфигурация для создания сервисов
type ServiceConfig struct {
	JWTSecret       string
	BotToken        string
	Network         string // "ton" или "evm"
	UseMockProvider bool
	CommissionRate  float64 // Ставка комиссии (по умолчанию 0.05 = 5%)
	Blockchain      config.BlockchainConfig
}

// NewService создает новый экземпляр Service
func NewService(repo repository.Repository, redisRepo repository.RedisRepository, jwtSecret, botToken string) Service {
	return NewServiceWithConfig(repo, redisRepo, ServiceConfig{
		JWTSecret:       jwtSecret,
		BotToken:        botToken,
		Network:         "ton",
		UseMockProvider: true,
	})
}

// NewServiceWithConfig создает новый экземпляр Service с полной конфигурацией
func NewServiceWithConfig(repo repository.Repository, redisRepo repository.RedisRepository, cfg ServiceConfig) Service {
	// Установка ставки комиссии по умолчанию
	commissionRate := cfg.CommissionRate
	if commissionRate <= 0 {
		commissionRate = 0.05 // 5% по умолчанию
	}

	service := &ServiceImpl{
		repo:           repo,
		redisRepo:      redisRepo,
		jwtSecret:      cfg.JWTSecret,
		botToken:       cfg.BotToken,
		commissionRate: commissionRate,
	}

	// Инициализация блокчейн провайдера
	service.blockchainProvider = createBlockchainProviderWithNetwork(cfg.Network, cfg.UseMockProvider, cfg.Blockchain)

	// Инициализация TON сервиса (может быть nil при использовании mock)
	// TON сервис создается отдельно и может быть передан извне при необходимости
	var tonService models.TONService = nil
	service.tonService = tonService

	// Инициализация сервисов - сначала создаем базовые сервисы
	txService := NewTransactionServiceImpl(repo.Transaction(), repo.User(), service.blockchainProvider)
	service.txService = txService

	service.userService = NewUserServiceImpl(repo.User(), txService)
	service.gameService = NewGameService(repo.Game(), redisRepo, service.userService, tonService, commissionRate)
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
		tonService,
		commissionRate,
	)

	service.authService = NewAuthService(repo.User(), redisRepo, cfg.JWTSecret, cfg.BotToken)

	// Создаем сервис фоновых задач
	service.jobService = NewJobService(
		service.lobbyService,
		service.txService,
		service.gameService,
		service.userService,
	)

	return service
}

// createBlockchainProviderWithNetwork создает провайдер блокчейна на основе сети и конфигурации
func createBlockchainProviderWithNetwork(network string, useMock bool, cfg config.BlockchainConfig) blockchain.BlockchainProvider {
	// Если включен mock режим, возвращаем mock провайдер
	if useMock {
		switch network {
		case "evm", "ethereum", "eth":
			return mock.NewMockProvider(blockchain.NetworkEthereum, []string{"ETH", "USDT", "USDC"})
		default: // ton
			return mock.NewMockProvider(blockchain.NetworkTON, []string{"TON", "USDT"})
		}
	}

	// Реальные провайдеры
	switch network {
	case "evm", "ethereum", "eth":
		return ethereum.NewEthereumProvider(ethereum.Config{
			RPCURL:                cfg.Ethereum.RPCURL,
			ChainID:               cfg.Ethereum.ChainID,
			MasterWallet:          cfg.Ethereum.MasterWallet,
			PrivateKey:            cfg.Ethereum.PrivateKey,
			MinWithdrawETH:        cfg.Ethereum.MinWithdrawETH,
			WithdrawFeeETH:        cfg.Ethereum.WithdrawFeeETH,
			RequiredConfirmations: cfg.Ethereum.RequiredConfirmations,
			USDTContractAddress:   cfg.Ethereum.USDTContractAddress,
		})
	default: // ton
		return ton.NewTonProvider(ton.Config{
			APIEndpoint:           cfg.TON.APIEndpoint,
			APIKey:                cfg.TON.APIKey,
			MasterWallet:          cfg.TON.MasterWallet,
			MasterWalletSecret:    cfg.TON.MasterWalletSecret,
			MinWithdrawTON:        cfg.TON.MinWithdrawTON,
			WithdrawFeeTON:        cfg.TON.WithdrawFeeTON,
			RequiredConfirmations: cfg.TON.RequiredConfirmations,
			Testnet:               cfg.TON.Testnet,
		})
	}
}

// createBlockchainProvider создает провайдер блокчейна (для обратной совместимости)
// Deprecated: используйте createBlockchainProviderWithNetwork
func createBlockchainProvider(cfg config.BlockchainConfig) blockchain.BlockchainProvider {
	return createBlockchainProviderWithNetwork("ton", false, cfg)
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

// BlockchainProvider возвращает провайдер блокчейна
func (s *ServiceImpl) BlockchainProvider() blockchain.BlockchainProvider {
	return s.blockchainProvider
}

// TONService возвращает сервис для работы с TON
func (s *ServiceImpl) TONService() models.TONService {
	return s.tonService
}

// Фабричные методы для создания сервисов определены в соответствующих файлах
