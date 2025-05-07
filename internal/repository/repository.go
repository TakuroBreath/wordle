package repository

import (
	"context"

	"github.com/TakuroBreath/wordle/internal/models"
	"github.com/google/uuid"
)

// Repository представляет собой интерфейс для всех репозиториев в приложении
type Repository interface {
	Game() models.GameRepository
	User() models.UserRepository
	Lobby() models.LobbyRepository
	Attempt() models.AttemptRepository
	History() models.HistoryRepository
	Transaction() models.TransactionRepository
}

// PostgresRepository представляет собой реализацию репозитория для PostgreSQL
type PostgresRepository interface {
	Repository
	Close() error
}

// RedisRepository представляет собой интерфейс для работы с Redis
type RedisRepository interface {
	SetSession(ctx context.Context, key string, value interface{}, expiration int) error
	GetSession(ctx context.Context, key string) (string, error)
	DeleteSession(ctx context.Context, key string) error
	SetGameState(ctx context.Context, lobbyID uuid.UUID, state interface{}) error
	GetGameState(ctx context.Context, lobbyID uuid.UUID) (string, error)
	DeleteGameState(ctx context.Context, lobbyID uuid.UUID) error
	Close() error
}
