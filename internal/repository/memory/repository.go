package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/TakuroBreath/wordle/internal/repository"
	"github.com/google/uuid"
)

// Repository представляет собой in-memory реализацию Redis репозитория
type Repository struct {
	mu         sync.RWMutex
	sessions   map[string]sessionData
	gameStates map[uuid.UUID]string
}

type sessionData struct {
	value      string
	expiresAt  time.Time
}

// NewRepository создает новый экземпляр in-memory репозитория
func NewRepository() *Repository {
	repo := &Repository{
		sessions:   make(map[string]sessionData),
		gameStates: make(map[uuid.UUID]string),
	}
	
	// Запускаем горутину для очистки истекших сессий
	go repo.cleanupExpiredSessions()
	
	return repo
}

// cleanupExpiredSessions периодически очищает истекшие сессии
func (r *Repository) cleanupExpiredSessions() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	
	for range ticker.C {
		r.mu.Lock()
		now := time.Now()
		for key, data := range r.sessions {
			if !data.expiresAt.IsZero() && now.After(data.expiresAt) {
				delete(r.sessions, key)
			}
		}
		r.mu.Unlock()
	}
}

// SetSession сохраняет сессию в памяти
func (r *Repository) SetSession(ctx context.Context, key string, value any, expiration int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	var strValue string
	switch v := value.(type) {
	case string:
		strValue = v
	default:
		jsonData, err := json.Marshal(value)
		if err != nil {
			return fmt.Errorf("failed to marshal session data: %w", err)
		}
		strValue = string(jsonData)
	}
	
	var expiresAt time.Time
	if expiration > 0 {
		expiresAt = time.Now().Add(time.Duration(expiration) * time.Second)
	}
	
	r.sessions[key] = sessionData{
		value:     strValue,
		expiresAt: expiresAt,
	}
	
	return nil
}

// GetSession получает сессию из памяти
func (r *Repository) GetSession(ctx context.Context, key string) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	data, ok := r.sessions[key]
	if !ok {
		return "", repository.ErrRedisNil
	}
	
	// Проверяем срок действия
	if !data.expiresAt.IsZero() && time.Now().After(data.expiresAt) {
		return "", repository.ErrRedisNil
	}
	
	// Если значение обернуто в кавычки JSON, удаляем их
	val := data.value
	if len(val) > 1 && val[0] == '"' && val[len(val)-1] == '"' {
		var unquoted string
		if err := json.Unmarshal([]byte(val), &unquoted); err == nil {
			return unquoted, nil
		}
	}
	
	return val, nil
}

// DeleteSession удаляет сессию из памяти
func (r *Repository) DeleteSession(ctx context.Context, key string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	delete(r.sessions, key)
	return nil
}

// SetGameState сохраняет состояние игры в памяти
func (r *Repository) SetGameState(ctx context.Context, lobbyID uuid.UUID, state any) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	jsonValue, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("failed to marshal game state: %w", err)
	}
	
	r.gameStates[lobbyID] = string(jsonValue)
	return nil
}

// GetGameState получает состояние игры из памяти
func (r *Repository) GetGameState(ctx context.Context, lobbyID uuid.UUID) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	val, ok := r.gameStates[lobbyID]
	if !ok {
		return "", repository.ErrRedisNil
	}
	
	return val, nil
}

// DeleteGameState удаляет состояние игры из памяти
func (r *Repository) DeleteGameState(ctx context.Context, lobbyID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	delete(r.gameStates, lobbyID)
	return nil
}

// Close закрывает репозиторий (для совместимости с интерфейсом)
func (r *Repository) Close() error {
	return nil
}
