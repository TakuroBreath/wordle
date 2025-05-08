package service

import (
	"context"
	"errors"
	"time"

	"github.com/TakuroBreath/wordle/internal/models"
	"github.com/google/uuid"
)

// HistoryServiceImpl представляет собой реализацию HistoryService
type HistoryServiceImpl struct {
	historyRepo models.HistoryRepository
	gameRepo    models.GameRepository
	userRepo    models.UserRepository
	lobbyRepo   models.LobbyRepository
}

// NewHistoryService создает новый экземпляр HistoryService
func NewHistoryService(
	historyRepo models.HistoryRepository,
	gameRepo models.GameRepository,
	userRepo models.UserRepository,
	lobbyRepo models.LobbyRepository,
) models.HistoryService {
	return &HistoryServiceImpl{
		historyRepo: historyRepo,
		gameRepo:    gameRepo,
		userRepo:    userRepo,
		lobbyRepo:   lobbyRepo,
	}
}

// CreateHistory создает новую запись в истории
func (s *HistoryServiceImpl) CreateHistory(ctx context.Context, history *models.History) error {
	if history == nil {
		return errors.New("history is nil")
	}

	// Проверяем существование игры
	_, err := s.gameRepo.GetByID(ctx, history.GameID)
	if err != nil {
		return err
	}

	// Проверяем существование пользователя
	_, err = s.userRepo.GetByTelegramID(ctx, history.UserID)
	if err != nil {
		return err
	}

	// Устанавливаем начальные значения
	history.ID = uuid.New()
	history.CreatedAt = time.Now()
	history.UpdatedAt = time.Now()

	return s.historyRepo.Create(ctx, history)
}

// GetHistory получает запись истории по ID
func (s *HistoryServiceImpl) GetHistory(ctx context.Context, id uuid.UUID) (*models.History, error) {
	return s.historyRepo.GetByID(ctx, id)
}

// GetGameHistory получает историю игры
func (s *HistoryServiceImpl) GetGameHistory(ctx context.Context, gameID uuid.UUID, limit, offset int) ([]*models.History, error) {
	return s.historyRepo.GetByGameID(ctx, gameID, limit, offset)
}

// GetUserHistory получает историю пользователя
func (s *HistoryServiceImpl) GetUserHistory(ctx context.Context, userID uint64, limit, offset int) ([]*models.History, error) {
	return s.historyRepo.GetByUserID(ctx, userID, limit, offset)
}

// UpdateHistory обновляет запись истории
func (s *HistoryServiceImpl) UpdateHistory(ctx context.Context, history *models.History) error {
	if history == nil {
		return errors.New("history is nil")
	}

	history.UpdatedAt = time.Now()
	return s.historyRepo.Update(ctx, history)
}

// DeleteHistory удаляет запись истории
func (s *HistoryServiceImpl) DeleteHistory(ctx context.Context, id uuid.UUID) error {
	return s.historyRepo.Delete(ctx, id)
}

// GetLobbyHistory получает историю лобби
func (s *HistoryServiceImpl) GetLobbyHistory(ctx context.Context, lobbyID uuid.UUID) (*models.History, error) {
	return s.historyRepo.GetByLobbyID(ctx, lobbyID)
}

// GetHistoryByStatus получает историю по статусу
func (s *HistoryServiceImpl) GetHistoryByStatus(ctx context.Context, status string, limit, offset int) ([]*models.History, error) {
	return s.historyRepo.GetByStatus(ctx, status, limit, offset)
}

// GetUserStats получает статистику пользователя
func (s *HistoryServiceImpl) GetUserStats(ctx context.Context, userID uint64) (map[string]interface{}, error) {
	// Получаем историю пользователя
	history, err := s.historyRepo.GetByUserID(ctx, userID, 0, 0)
	if err != nil {
		return nil, err
	}

	// Получаем данные пользователя
	user, err := s.userRepo.GetByTelegramID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Считаем статистику
	stats := make(map[string]interface{})
	stats["total_games"] = len(history)
	stats["wins"] = user.Wins
	stats["losses"] = user.Losses
	stats["win_rate"] = float64(user.Wins) / float64(user.Wins+user.Losses)
	stats["balance_ton"] = user.BalanceTon
	stats["balance_usdt"] = user.BalanceUsdt

	// Считаем среднюю награду
	var totalReward float64
	var gamesWithReward int
	for _, h := range history {
		if h.Reward > 0 {
			totalReward += h.Reward
			gamesWithReward++
		}
	}
	if gamesWithReward > 0 {
		stats["average_reward"] = totalReward / float64(gamesWithReward)
	} else {
		stats["average_reward"] = 0
	}

	return stats, nil
}

// GetGameHistoryStats получает статистику истории игры
func (s *HistoryServiceImpl) GetGameHistoryStats(ctx context.Context, gameID uuid.UUID) (map[string]interface{}, error) {
	// Получаем историю игры
	history, err := s.historyRepo.GetByGameID(ctx, gameID, 0, 0)
	if err != nil {
		return nil, err
	}

	// Получаем данные игры
	game, err := s.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		return nil, err
	}

	// Считаем статистику
	stats := make(map[string]interface{})
	stats["total_attempts"] = len(history)
	stats["word"] = game.Word
	stats["difficulty"] = game.Difficulty
	stats["reward_pool_ton"] = game.RewardPoolTon
	stats["reward_pool_usdt"] = game.RewardPoolUsdt

	// Считаем количество побед и поражений
	var wins, losses int
	for _, h := range history {
		if h.Status == "success" {
			wins++
		} else if h.Status == "failed" {
			losses++
		}
	}
	stats["wins"] = wins
	stats["losses"] = losses
	stats["win_rate"] = float64(wins) / float64(wins+losses)

	// Считаем среднее количество попыток
	var totalTries int
	for _, h := range history {
		lobby, err := s.lobbyRepo.GetByID(ctx, h.LobbyID)
		if err == nil && lobby != nil {
			totalTries += lobby.TriesUsed
		}
	}
	if len(history) > 0 {
		stats["average_tries"] = float64(totalTries) / float64(len(history))
	} else {
		stats["average_tries"] = 0
	}

	return stats, nil
}
