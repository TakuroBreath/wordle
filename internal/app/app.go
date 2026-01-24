package app

import (
	"context"
	"log"
	"time"

	"github.com/TakuroBreath/wordle/internal/api/server"
	"github.com/TakuroBreath/wordle/internal/config"
	"github.com/TakuroBreath/wordle/internal/repository"
	"github.com/TakuroBreath/wordle/internal/repository/memory"
	"github.com/TakuroBreath/wordle/internal/repository/postgresql"
	"github.com/TakuroBreath/wordle/internal/service"
	"github.com/TakuroBreath/wordle/internal/worker"
	"github.com/TakuroBreath/wordle/pkg/metrics"
)

// App представляет структуру приложения
type App struct {
	cfg      *config.Config
	server   *server.Server
	services *service.ServiceImpl
	repos    repository.Repository
	redis    repository.RedisRepository
}

// New создает новое приложение
func New(cfg *config.Config) (*App, error) {
	// Инициализация репозиториев
	postgresDB, err := postgresql.NewConnection(cfg.Postgres)
	if err != nil {
		return nil, err
	}

	repos := postgresql.NewRepository(postgresDB)
	memoryRepos := memory.NewRepository()

	// Инициализация сервисов с полной конфигурацией
	serviceCfg := service.ServiceConfig{
		JWTSecret:       cfg.Auth.JWTSecret,
		BotToken:        cfg.Auth.BotToken,
		Network:         string(cfg.Network),
		UseMockProvider: cfg.UseMockProvider,
		Blockchain:      cfg.Blockchain,
	}
	services := service.NewServiceWithConfig(repos, memoryRepos, serviceCfg)
	servicesImpl := services.(*service.ServiceImpl)

	// Инициализация сервера
	serverConfig := server.Config{
		Port:         cfg.HTTP.Port,
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
		IdleTimeout:  cfg.HTTP.IdleTimeout,
		BotToken:     cfg.Auth.BotToken,
		AuthEnabled:  cfg.IsAuthEnabled(),
	}
	httpServer := server.NewServer(serverConfig, servicesImpl)

	return &App{
		cfg:      cfg,
		server:   httpServer,
		services: servicesImpl,
		repos:    repos,
		redis:    memoryRepos,
	}, nil
}

// Run запускает приложение
func (a *App) Run() error {
	// Логируем конфигурацию
	log.Printf("Starting application with config: %s", a.cfg.String())
	log.Printf("Environment: %s, Network: %s, Auth enabled: %v, Mock provider: %v",
		a.cfg.Environment, a.cfg.Network, a.cfg.IsAuthEnabled(), a.cfg.UseMockProvider)

	// Инициализация Prometheus метрик
	metrics.InitMetrics(a.cfg.Metrics.Enabled, a.cfg.Metrics.Port)

	// Запуск джобов для фоновой обработки
	ctx := context.Background()
	go func() {
		a.services.Job().StartJobScheduler(
			ctx,
			time.Minute,   // Проверка истекших лобби каждую минуту
			time.Minute*5, // Проверка ожидающих транзакций каждые 5 минут
		)
	}()

	// Запуск блокчейн воркера (только если используется TON и не в режиме mock)
	if a.cfg.Network == "ton" && !a.cfg.UseMockProvider && a.services.TONService() != nil {
		pollInterval := time.Duration(a.cfg.Blockchain.TON.WorkerPollInterval) * time.Second
		if pollInterval <= 0 {
			pollInterval = 10 * time.Second
		}

		bcWorker := worker.NewBlockchainWorker(
			a.services.TONService(),
			a.repos.Game(),
			a.repos.Lobby(),
			a.repos.User(),
			a.repos.Transaction(),
			worker.WorkerConfig{
				PollInterval:        pollInterval,
				MasterWalletAddress: a.services.TONService().GetMasterWalletAddress(),
			},
		)

		go func() {
			if err := bcWorker.Start(ctx); err != nil {
				log.Printf("ERROR: Failed to start blockchain worker: %v", err)
			}
		}()
	}

	// Запуск HTTP-сервера
	return a.server.Run()
}

// Shutdown выполняет корректное завершение работы приложения
func (a *App) Shutdown() {
	// Закрытие соединений с базами данных
	if closer, ok := a.repos.(interface{ Close() error }); ok {
		if err := closer.Close(); err != nil {
			log.Printf("Failed to close PostgreSQL connection: %v", err)
		}
	}

	if closer, ok := a.redis.(interface{ Close() error }); ok {
		if err := closer.Close(); err != nil {
			log.Printf("Failed to close memory repository: %v", err)
		}
	}
}
