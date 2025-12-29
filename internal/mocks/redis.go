package mocks

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/google/uuid"
)

// MockRedisRepository мок для Redis репозитория (implements repository.RedisRepository)
type MockRedisRepository struct {
	mu         sync.RWMutex
	data       map[string]string
	expiries   map[string]time.Time
	gameStates map[uuid.UUID]string
}

func NewMockRedisRepository() *MockRedisRepository {
	return &MockRedisRepository{
		data:       make(map[string]string),
		expiries:   make(map[string]time.Time),
		gameStates: make(map[uuid.UUID]string),
	}
}

func (m *MockRedisRepository) SetSession(ctx context.Context, key string, value any, expirationSeconds int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	switch v := value.(type) {
	case string:
		m.data[key] = v
	default:
		jsonData, _ := json.Marshal(value)
		m.data[key] = string(jsonData)
	}
	
	if expirationSeconds > 0 {
		m.expiries[key] = time.Now().Add(time.Duration(expirationSeconds) * time.Second)
	}
	return nil
}

func (m *MockRedisRepository) GetSession(ctx context.Context, key string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	// Проверяем срок действия
	if expiry, ok := m.expiries[key]; ok {
		if time.Now().After(expiry) {
			return "", nil
		}
	}
	
	if value, ok := m.data[key]; ok {
		return value, nil
	}
	return "", nil
}

func (m *MockRedisRepository) DeleteSession(ctx context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.data, key)
	delete(m.expiries, key)
	return nil
}

func (m *MockRedisRepository) SetGameState(ctx context.Context, lobbyID uuid.UUID, state any) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	jsonData, _ := json.Marshal(state)
	m.gameStates[lobbyID] = string(jsonData)
	return nil
}

func (m *MockRedisRepository) GetGameState(ctx context.Context, lobbyID uuid.UUID) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if state, ok := m.gameStates[lobbyID]; ok {
		return state, nil
	}
	return "", nil
}

func (m *MockRedisRepository) DeleteGameState(ctx context.Context, lobbyID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.gameStates, lobbyID)
	return nil
}

func (m *MockRedisRepository) Close() error {
	return nil
}
