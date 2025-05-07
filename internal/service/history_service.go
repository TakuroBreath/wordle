package service

import (
	"context"
	"errors"
	"time"

	"github.com/TakuroBreath/wordle/internal/models"
	"github.com/google/uuid"
)

// HistoryService представляет собой реализацию сервиса для работы с историей игр
type HistoryService struct {
	repo models.HistoryRepository
}

// NewHistoryService создает новый экземпляр HistoryService
func NewHistoryService(repo models.HistoryRepository) *HistoryService {
	return &HistoryService{
		repo: repo,
	}
}

// Create создает новую запись в истории
func (s *HistoryService) Create(ctx context.Context, history *models.History) error {
	// Проверка валидности данных
	if history.UserID == uuid.Nil {
		return errors.New("user ID cannot be empty")
	}

	if history.GameID == uuid.Nil {
		return errors.New("game ID cannot be empty")
	}

	if history.LobbyID == uuid.Nil {
		return errors.New("lobby ID cannot be empty")
	}

	if history.Status != "creator_win" && history.Status != "player_win" {
		return errors.New("invalid status")
	}

	if history.Reward < 0 {
		return errors.New("reward cannot be negative")
	}

	// Генерация ID, если он не был установлен
	if history.ID == uuid.Nil {
		history.ID = uuid.New()
	}

	// Установка текущего времени, если оно не было установлено
	if history.CreatedAt.IsZero() {
		history.CreatedAt = time.Now()
	}

	// Сохранение записи в базе данных
	return s.repo.Create(ctx, history)
}

// GetByID получает запись истории по ID
func (s *HistoryService) GetByID(ctx context.Context, id uuid.UUID) (*models.History, error) {
	return s.repo.GetByID(ctx, id)
}

// GetByUserID получает историю игр пользователя
func (s *HistoryService) GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.History, error) {
	return s.repo.GetByUserID(ctx, userID, limit, offset)
}

// GetByGameID получает историю для конкретной игры
func (s *HistoryService) GetByGameID(ctx context.Context, gameID uuid.UUID, limit, offset int) ([]*models.History, error) {
	return s.repo.GetByGameID(ctx, gameID, limit, offset)
}
