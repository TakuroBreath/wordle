package redis

import (
	"github.com/TakuroBreath/wordle/internal/config"
	"github.com/redis/go-redis/v9"
)

// NewRedisClient создает и возвращает новый клиент Redis
func NewRedisClient(cfg config.RedisConfig) (*redis.Client, error) {
	// Создаем новый клиент Redis с параметрами из конфигурации
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr(),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	return client, nil
}

// NewRedisRepository создает и возвращает новый репозиторий Redis
func NewRedisRepository(client *redis.Client) *Repository {
	return NewRepository(client)
}
