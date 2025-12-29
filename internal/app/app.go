package app

import (
	"context"
	"time"

	"github.com/TakuroBreath/wordle/internal/api/server"
	"github.com/TakuroBreath/wordle/internal/config"
	"github.com/TakuroBreath/wordle/internal/logger"
	"github.com/TakuroBreath/wordle/internal/repository"
	"github.com/TakuroBreath/wordle/internal/repository/postgresql"
	"github.com/TakuroBreath/wordle/internal/repository/redis"
	"github.com/TakuroBreath/wordle/internal/service"
	"github.com/TakuroBreath/wordle/pkg/metrics"
	"go.uber.org/zap"
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

	redisDB, err := redis.NewRedisClient(cfg.Redis)
	if err != nil {
		return nil, err
	}

	repos := postgresql.NewRepository(postgresDB)
	redisRepos := redis.NewRedisRepository(redisDB)

	// Инициализация сервисов с полной конфигурацией
	serviceCfg := service.ServiceConfig{
		JWTSecret:       cfg.Auth.JWTSecret,
		BotToken:        cfg.Auth.BotToken,
		Network:         string(cfg.Network),
		UseMockProvider: cfg.UseMockProvider,
		Blockchain:      cfg.Blockchain,
	}
	services := service.NewServiceWithConfig(repos, redisRepos, serviceCfg)
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
		redis:    redisRepos,
	}, nil
}

// Run запускает приложение
func (a *App) Run() error {
	// Логируем конфигурацию
	logger.Log.Info("Starting application",
		zap.String("config", a.cfg.String()),
		zap.String("environment", string(a.cfg.Environment)),
		zap.String("network", string(a.cfg.Network)),
		zap.Bool("auth_enabled", a.cfg.IsAuthEnabled()),
		zap.Bool("mock_provider", a.cfg.UseMockProvider),
	)

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

	// Запуск HTTP-сервера
	return a.server.Run()
}

// Shutdown выполняет корректное завершение работы приложения
func (a *App) Shutdown() {
	// Закрытие соединений с базами данных
	if closer, ok := a.repos.(interface{ Close() error }); ok {
		if err := closer.Close(); err != nil {
			logger.Log.Warn("Failed to close PostgreSQL connection", zap.Error(err))
		}
	}

	if closer, ok := a.redis.(interface{ Close() error }); ok {
		if err := closer.Close(); err != nil {
			logger.Log.Warn("Failed to close Redis connection", zap.Error(err))
		}
	}
}
