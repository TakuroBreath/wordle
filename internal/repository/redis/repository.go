package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// Repository представляет собой реализацию Redis репозитория
type Repository struct {
	client *redis.Client
}

// NewRepository создает новый экземпляр Repository
func NewRepository(client *redis.Client) *Repository {
	return &Repository{
		client: client,
	}
}

// SetSession сохраняет сессию в Redis
func (r *Repository) SetSession(ctx context.Context, key string, value interface{}, expiration int) error {
	jsonValue, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal session data: %w", err)
	}

	err = r.client.Set(ctx, key, jsonValue, time.Duration(expiration)*time.Second).Err()
	if err != nil {
		return fmt.Errorf("failed to set session in Redis: %w", err)
	}

	return nil
}

// GetSession получает сессию из Redis
func (r *Repository) GetSession(ctx context.Context, key string) (string, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", fmt.Errorf("session not found")
		}
		return "", fmt.Errorf("failed to get session from Redis: %w", err)
	}

	return val, nil
}

// DeleteSession удаляет сессию из Redis
func (r *Repository) DeleteSession(ctx context.Context, key string) error {
	err := r.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete session from Redis: %w", err)
	}

	return nil
}

// SetGameState сохраняет состояние игры в Redis
func (r *Repository) SetGameState(ctx context.Context, lobbyID uuid.UUID, state interface{}) error {
	key := fmt.Sprintf("game:%s", lobbyID.String())
	jsonValue, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("failed to marshal game state: %w", err)
	}

	err = r.client.Set(ctx, key, jsonValue, 24*time.Hour).Err()
	if err != nil {
		return fmt.Errorf("failed to set game state in Redis: %w", err)
	}

	return nil
}

// GetGameState получает состояние игры из Redis
func (r *Repository) GetGameState(ctx context.Context, lobbyID uuid.UUID) (string, error) {
	key := fmt.Sprintf("game:%s", lobbyID.String())
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", fmt.Errorf("game state not found")
		}
		return "", fmt.Errorf("failed to get game state from Redis: %w", err)
	}

	return val, nil
}

// DeleteGameState удаляет состояние игры из Redis
func (r *Repository) DeleteGameState(ctx context.Context, lobbyID uuid.UUID) error {
	key := fmt.Sprintf("game:%s", lobbyID.String())
	err := r.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete game state from Redis: %w", err)
	}

	return nil
}

// Close закрывает соединение с Redis
func (r *Repository) Close() error {
	return r.client.Close()
}
