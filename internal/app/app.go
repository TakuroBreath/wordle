package app

import (
	"context"
	"log"
	"time"

	"github.com/TakuroBreath/wordle/internal/api/server"
	"github.com/TakuroBreath/wordle/internal/config"
	"github.com/TakuroBreath/wordle/internal/repository"
	"github.com/TakuroBreath/wordle/internal/repository/postgresql"
	"github.com/TakuroBreath/wordle/internal/repository/redis"
	"github.com/TakuroBreath/wordle/internal/service"
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

	// Инициализация сервисов
	services := service.NewService(repos, redisRepos, cfg.Auth.JWTSecret, cfg.Auth.BotToken)

	// Инициализация сервера
	serverConfig := server.Config{
		Port:         cfg.HTTP.Port,
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
		IdleTimeout:  cfg.HTTP.IdleTimeout,
		BotToken:     cfg.Auth.BotToken,
	}
	httpServer := server.NewServer(serverConfig, services)

	return &App{
		cfg:      cfg,
		server:   httpServer,
		services: services,
		repos:    repos,
		redis:    redisRepos,
	}, nil
}

// Run запускает приложение
func (a *App) Run() error {
	// Запуск джобов для фоновой обработки
	ctx := context.Background()
	go func() {
		if err := a.services.Job().StartJobScheduler(
			ctx,
			time.Minute,   // Проверка истекших лобби каждую минуту
			time.Minute*5, // Проверка ожидающих транзакций каждые 5 минут
		); err != nil {
			log.Printf("Failed to start job scheduler: %v", err)
		}
	}()

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
			log.Printf("Failed to close Redis connection: %v", err)
		}
	}
}
