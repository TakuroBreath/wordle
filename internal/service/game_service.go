package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/TakuroBreath/wordle/internal/logger"
	"github.com/TakuroBreath/wordle/internal/models"
	"github.com/TakuroBreath/wordle/internal/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// GameServiceImpl представляет собой реализацию GameService
type GameServiceImpl struct {
	gameRepo    models.GameRepository
	redisRepo   repository.RedisRepository
	userService models.UserService
	logger      *zap.Logger
}

// NewGameService создает новый экземпляр GameService
func NewGameService(gameRepo models.GameRepository, redisRepo repository.RedisRepository, userService models.UserService) models.GameService {
	return &GameServiceImpl{
		gameRepo:    gameRepo,
		redisRepo:   redisRepo,
		userService: userService,
		logger:      logger.GetLogger(zap.String("service", "game")),
	}
}

// CreateGame создает новую игру
func (s *GameServiceImpl) CreateGame(ctx context.Context, game *models.Game) error {
	log := s.logger.With(zap.String("method", "CreateGame"))

	if game == nil {
		log.Error("Game is nil")
		return errors.New("game is nil")
	}

	log.Info("Creating new game",
		zap.String("creator_id", fmt.Sprintf("%d", game.CreatorID)),
		zap.String("word", game.Word),
		zap.String("title", game.Title),
		zap.Float64("min_bet", game.MinBet),
		zap.Float64("max_bet", game.MaxBet),
		zap.Float64("reward_multiplier", game.RewardMultiplier))

	// Валидация основных полей
	if game.CreatorID == 0 {
		log.Error("Invalid creator ID")
		return errors.New("invalid creator ID")
	}
	if game.Word == "" {
		log.Error("Word cannot be empty")
		return errors.New("word cannot be empty")
	}
	if game.Length != len([]rune(game.Word)) { // Проверяем длину слова по рунам
		log.Error("Word length mismatch",
			zap.Int("specified_length", game.Length),
			zap.Int("actual_length", len([]rune(game.Word))))
		return fmt.Errorf("word length mismatch: specified length %d, actual length %d", game.Length, len([]rune(game.Word)))
	}
	if game.Length <= 0 { // Длина тоже должна быть > 0
		log.Error("Invalid word length", zap.Int("length", game.Length))
		return errors.New("invalid word length")
	}
	if game.Title == "" {
		log.Error("Title cannot be empty")
		return errors.New("title cannot be empty")
	}

	// Валидация сложности (пример)
	allowedDifficulties := map[string]bool{"easy": true, "medium": true, "hard": true}
	if !allowedDifficulties[game.Difficulty] {
		log.Error("Invalid difficulty", zap.String("difficulty", game.Difficulty))
		return fmt.Errorf("invalid difficulty: %s", game.Difficulty)
	}

	// Валидация валюты
	if game.Currency != models.CurrencyTON && game.Currency != models.CurrencyUSDT {
		log.Error("Invalid currency", zap.String("currency", game.Currency))
		return fmt.Errorf("invalid currency: %s", game.Currency)
	}

	// Валидация параметров игры
	if game.MinBet <= 0 {
		log.Error("Min bet must be positive", zap.Float64("min_bet", game.MinBet))
		return errors.New("min bet must be positive")
	}
	if game.MaxBet < game.MinBet {
		log.Error("Max bet cannot be less than min bet",
			zap.Float64("min_bet", game.MinBet),
			zap.Float64("max_bet", game.MaxBet))
		return errors.New("max bet cannot be less than min bet")
	}
	if game.RewardMultiplier < 1.0 { // Множитель должен быть хотя бы 1
		log.Error("Invalid reward multiplier", zap.Float64("reward_multiplier", game.RewardMultiplier))
		return errors.New("invalid reward multiplier, must be >= 1.0")
	}
	if game.MaxTries <= 0 {
		log.Error("Max tries must be positive", zap.Int("max_tries", game.MaxTries))
		return errors.New("max tries must be positive")
	}

	// Установка начальных значений
	game.ID = uuid.New()
	game.Status = models.GameStatusInactive // Используем константу
	game.RewardPoolTon = 0.0                // Явно инициализируем пулы
	game.RewardPoolUsdt = 0.0
	now := time.Now()
	game.CreatedAt = now
	game.UpdatedAt = now

	log.Debug("Game parameters validated, creating game",
		zap.String("game_id", game.ID.String()),
		zap.String("status", game.Status))

	err := s.gameRepo.Create(ctx, game)
	if err != nil {
		log.Error("Failed to create game", zap.Error(err))
		return err
	}

	log.Info("Game created successfully", zap.String("game_id", game.ID.String()))
	return nil
}

// GetGame получает игру по ID
func (s *GameServiceImpl) GetGame(ctx context.Context, id uuid.UUID) (*models.Game, error) {
	log := s.logger.With(zap.String("method", "GetGame"), zap.String("game_id", id.String()))
	log.Info("Getting game by ID")

	game, err := s.gameRepo.GetByID(ctx, id)
	if err != nil {
		log.Error("Failed to get game", zap.Error(err))
		return nil, err
	}

	log.Info("Game retrieved successfully",
		zap.String("title", game.Title),
		zap.String("status", game.Status))
	return game, nil
}

// GetUserGames получает список игр пользователя
func (s *GameServiceImpl) GetUserGames(ctx context.Context, userID uint64, limit, offset int) ([]*models.Game, error) {
	log := s.logger.With(zap.String("method", "GetUserGames"),
		zap.Uint64("user_id", userID),
		zap.Int("limit", limit),
		zap.Int("offset", offset))
	log.Info("Getting user games")

	games, err := s.gameRepo.GetByUserID(ctx, userID, limit, offset)
	if err != nil {
		log.Error("Failed to get user games", zap.Error(err))
		return nil, err
	}

	log.Info("User games retrieved successfully", zap.Int("count", len(games)))
	return games, nil
}

// UpdateGame обновляет игру
func (s *GameServiceImpl) UpdateGame(ctx context.Context, game *models.Game) error {
	log := s.logger.With(zap.String("method", "UpdateGame"))

	if game == nil {
		log.Error("Game is nil")
		return errors.New("game is nil")
	}

	log.Info("Updating game",
		zap.String("game_id", game.ID.String()),
		zap.String("title", game.Title))

	// Получаем существующую игру, чтобы обновить только разрешенные поля
	existingGame, err := s.gameRepo.GetByID(ctx, game.ID)
	if err != nil {
		log.Error("Game not found for update", zap.Error(err), zap.String("game_id", game.ID.String()))
		return fmt.Errorf("game not found for update: %w", err)
	}

	log.Debug("Game found, updating fields",
		zap.String("game_id", existingGame.ID.String()),
		zap.String("old_title", existingGame.Title),
		zap.String("new_title", game.Title))

	// Обновляем только изменяемые поля (например, Title, Description)
	existingGame.Title = game.Title
	existingGame.Description = game.Description
	// Другие поля, такие как ставки, слово, множитель, обычно не должны меняться после создания
	existingGame.UpdatedAt = time.Now()

	err = s.gameRepo.Update(ctx, existingGame)
	if err != nil {
		log.Error("Failed to update game", zap.Error(err), zap.String("game_id", game.ID.String()))
		return err
	}

	log.Info("Game updated successfully", zap.String("game_id", game.ID.String()))
	return nil
}

// DeleteGame удаляет игру и возвращает средства из reward pool создателю.
// Важно: Проверка, что вызывающий пользователь является создателем игры, должна быть на уровне handler.
func (s *GameServiceImpl) DeleteGame(ctx context.Context, id uuid.UUID) error {
	log := s.logger.With(zap.String("method", "DeleteGame"), zap.String("game_id", id.String()))
	log.Info("Deleting game")

	if id == uuid.Nil {
		log.Error("Game ID cannot be nil for delete")
		return errors.New("game ID cannot be nil for delete")
	}

	game, err := s.gameRepo.GetByID(ctx, id)
	if err != nil {
		// Обрабатываем случай, когда игра не найдена
		if errors.Is(err, models.ErrGameNotFound) { // Предполагаем наличие ErrGameNotFound
			log.Error("Game not found", zap.String("game_id", id.String()))
			return errors.New("game not found")
		}
		log.Error("Failed to get game for deletion", zap.Error(err))
		return fmt.Errorf("failed to get game for deletion: %w", err)
	}

	log.Debug("Game found",
		zap.String("game_id", game.ID.String()),
		zap.String("status", game.Status),
		zap.Float64("reward_pool_ton", game.RewardPoolTon),
		zap.Float64("reward_pool_usdt", game.RewardPoolUsdt))

	// Проверяем, что игра не активна
	if game.Status == models.GameStatusActive {
		log.Error("Cannot delete an active game", zap.String("game_id", id.String()))
		return errors.New("cannot delete an active game (consider deactivating first or checking for active lobbies)")
	}

	// Определяем сумму и валюту возврата
	var returnAmount float64
	var returnCurrency string
	if game.Currency == models.CurrencyTON && game.RewardPoolTon > 0 {
		returnAmount = game.RewardPoolTon
		returnCurrency = models.CurrencyTON
	} else if game.Currency == models.CurrencyUSDT && game.RewardPoolUsdt > 0 {
		returnAmount = game.RewardPoolUsdt
		returnCurrency = models.CurrencyUSDT
	}

	// Если есть что возвращать, обновляем баланс создателя
	if returnAmount > 0 {
		log.Info("Returning reward pool to creator",
			zap.Float64("amount", returnAmount),
			zap.String("currency", returnCurrency),
			zap.Uint64("creator_id", game.CreatorID))

		var errUpdateBalance error
		if returnCurrency == models.CurrencyTON {
			errUpdateBalance = s.userService.UpdateTonBalance(ctx, game.CreatorID, returnAmount)
		} else {
			errUpdateBalance = s.userService.UpdateUsdtBalance(ctx, game.CreatorID, returnAmount)
		}

		if errUpdateBalance != nil {
			log.Error("Failed to return reward pool to creator",
				zap.Error(errUpdateBalance),
				zap.Uint64("creator_id", game.CreatorID),
				zap.Float64("amount", returnAmount),
				zap.String("currency", returnCurrency))
			return fmt.Errorf("failed to return reward pool to creator %d (%s): %w", game.CreatorID, returnCurrency, errUpdateBalance)
		}

		log.Info("Reward pool returned successfully to creator",
			zap.Uint64("creator_id", game.CreatorID),
			zap.Float64("amount", returnAmount),
			zap.String("currency", returnCurrency))
	} else {
		log.Info("No reward pool to return")
	}

	// Удаляем игру
	log.Debug("Deleting game from repository", zap.String("game_id", id.String()))
	if err := s.gameRepo.Delete(ctx, id); err != nil {
		log.Error("Failed to delete game after returning reward pool",
			zap.Error(err),
			zap.String("game_id", id.String()))
		return fmt.Errorf("reward pool returned, but failed to delete game %s: %w", id.String(), err)
	}

	log.Info("Game deleted successfully", zap.String("game_id", id.String()))
	return nil
}

// GetActiveGames получает список активных игр
func (s *GameServiceImpl) GetActiveGames(ctx context.Context, limit, offset int) ([]*models.Game, error) {
	log := s.logger.With(zap.String("method", "GetActiveGames"),
		zap.Int("limit", limit),
		zap.Int("offset", offset))
	log.Info("Getting active games")

	games, err := s.gameRepo.GetActive(ctx, limit, offset)
	if err != nil {
		log.Error("Failed to get active games", zap.Error(err))
		return nil, err
	}

	log.Info("Active games retrieved successfully", zap.Int("count", len(games)))
	return games, nil
}

// SearchGames ищет игры по параметрам
func (s *GameServiceImpl) SearchGames(ctx context.Context, minBet, maxBet float64, difficulty string, limit, offset int) ([]*models.Game, error) {
	log := s.logger.With(zap.String("method", "SearchGames"),
		zap.Float64("min_bet", minBet),
		zap.Float64("max_bet", maxBet),
		zap.String("difficulty", difficulty),
		zap.Int("limit", limit),
		zap.Int("offset", offset))
	log.Info("Searching games")

	games, err := s.gameRepo.SearchGames(ctx, minBet, maxBet, difficulty, limit, offset)
	if err != nil {
		log.Error("Failed to search games", zap.Error(err))
		return nil, err
	}

	log.Info("Games search completed successfully", zap.Int("count", len(games)))
	return games, nil
}

// GetGameStats получает статистику игры
func (s *GameServiceImpl) GetGameStats(ctx context.Context, gameID uuid.UUID) (map[string]interface{}, error) {
	log := s.logger.With(zap.String("method", "GetGameStats"), zap.String("game_id", gameID.String()))
	log.Info("Getting game statistics")

	stats, err := s.gameRepo.GetGameStats(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game statistics", zap.Error(err))
		return nil, err
	}

	log.Info("Game statistics retrieved successfully")
	return stats, nil
}

// AddToRewardPool добавляет средства в reward pool игры
func (s *GameServiceImpl) AddToRewardPool(ctx context.Context, gameID uuid.UUID, amount float64) error {
	log := s.logger.With(zap.String("method", "AddToRewardPool"),
		zap.String("game_id", gameID.String()),
		zap.Float64("amount", amount))
	log.Info("Adding to game reward pool")

	game, err := s.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game", zap.Error(err))
		return err
	}

	log.Debug("Game found",
		zap.String("game_id", game.ID.String()),
		zap.String("currency", game.Currency),
		zap.String("status", game.Status),
		zap.Float64("reward_pool_ton", game.RewardPoolTon),
		zap.Float64("reward_pool_usdt", game.RewardPoolUsdt))

	if amount <= 0 {
		log.Error("Deposit amount must be positive", zap.Float64("amount", amount))
		return errors.New("deposit amount must be positive")
	}

	// Проверяем, что игра активна для депозита
	if game.Status != models.GameStatusActive {
		log.Error("Cannot add to reward pool of inactive game",
			zap.String("status", game.Status))
		return errors.New("cannot add to reward pool of inactive game")
	}

	// В зависимости от валюты игры добавляем средства в соответствующий пул
	if game.Currency == models.CurrencyTON {
		log.Info("Adding to TON reward pool",
			zap.Float64("current_pool", game.RewardPoolTon),
			zap.Float64("amount_to_add", amount))
		game.RewardPoolTon += amount
	} else if game.Currency == models.CurrencyUSDT {
		log.Info("Adding to USDT reward pool",
			zap.Float64("current_pool", game.RewardPoolUsdt),
			zap.Float64("amount_to_add", amount))
		game.RewardPoolUsdt += amount
	} else {
		log.Error("Unknown currency", zap.String("currency", game.Currency))
		return fmt.Errorf("unknown currency: %s", game.Currency)
	}

	game.UpdatedAt = time.Now()

	// Обновляем игру в репозитории
	log.Debug("Updating game with new reward pool",
		zap.Float64("reward_pool_ton", game.RewardPoolTon),
		zap.Float64("reward_pool_usdt", game.RewardPoolUsdt))

	if err := s.gameRepo.Update(ctx, game); err != nil {
		log.Error("Failed to update reward pool", zap.Error(err))
		return fmt.Errorf("failed to update reward pool: %w", err)
	}

	log.Info("Reward pool updated successfully")
	return nil
}

// ActivateGame активирует игру после подтверждения депозита
func (s *GameServiceImpl) ActivateGame(ctx context.Context, gameID uuid.UUID) error {
	log := s.logger.With(zap.String("method", "ActivateGame"), zap.String("game_id", gameID.String()))
	log.Info("Activating game")

	game, err := s.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game", zap.Error(err))
		return err
	}

	log.Debug("Game found",
		zap.String("game_id", game.ID.String()),
		zap.String("status", game.Status),
		zap.String("currency", game.Currency),
		zap.Float64("reward_pool_ton", game.RewardPoolTon),
		zap.Float64("reward_pool_usdt", game.RewardPoolUsdt))

	// Проверяем, что игра неактивна
	if game.Status != models.GameStatusInactive {
		log.Error("Game is already active", zap.String("status", game.Status))
		return errors.New("game is already active")
	}

	// Проверяем, что reward pool содержит необходимую сумму
	var rewardPoolBalance float64
	var requiredBalance float64 = game.MaxBet * game.RewardMultiplier

	if game.Currency == models.CurrencyTON {
		rewardPoolBalance = game.RewardPoolTon
	} else if game.Currency == models.CurrencyUSDT {
		rewardPoolBalance = game.RewardPoolUsdt
	} else {
		log.Error("Unknown currency", zap.String("currency", game.Currency))
		return fmt.Errorf("unknown currency: %s", game.Currency)
	}

	log.Debug("Checking reward pool balance",
		zap.Float64("reward_pool_balance", rewardPoolBalance),
		zap.Float64("required_balance", requiredBalance))

	// Проверяем достаточность средств для активации
	if rewardPoolBalance < requiredBalance {
		log.Error("Insufficient funds in reward pool",
			zap.Float64("reward_pool_balance", rewardPoolBalance),
			zap.Float64("required_balance", requiredBalance))
		return fmt.Errorf("insufficient funds in reward pool: need %.2f %s, but have %.2f",
			requiredBalance, game.Currency, rewardPoolBalance)
	}

	// Обновляем статус на активный
	log.Info("Updating game status to active")
	err = s.gameRepo.UpdateStatus(ctx, gameID, models.GameStatusActive)
	if err != nil {
		log.Error("Failed to update game status", zap.Error(err))
		return fmt.Errorf("failed to activate game: %w", err)
	}

	log.Info("Game activated successfully")
	return nil
}

// DeactivateGame деактивирует игру
func (s *GameServiceImpl) DeactivateGame(ctx context.Context, gameID uuid.UUID) error {
	log := s.logger.With(zap.String("method", "DeactivateGame"), zap.String("game_id", gameID.String()))
	log.Info("Deactivating game")

	game, err := s.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game", zap.Error(err))
		return err
	}

	log.Debug("Game found",
		zap.String("game_id", game.ID.String()),
		zap.String("status", game.Status))

	// Проверяем, что игра активна
	if game.Status != models.GameStatusActive {
		log.Error("Game is already inactive", zap.String("status", game.Status))
		return errors.New("game is already inactive")
	}

	// Можно было бы добавить здесь проверку на отсутствие активных лобби
	// Но это требует дополнительного слоя логики с LobbyService

	// Обновляем статус на неактивный
	log.Info("Updating game status to inactive")
	err = s.gameRepo.UpdateStatus(ctx, gameID, models.GameStatusInactive)
	if err != nil {
		log.Error("Failed to update game status", zap.Error(err))
		return fmt.Errorf("failed to deactivate game: %w", err)
	}

	log.Info("Game deactivated successfully")
	return nil
}

// CheckWord проверяет слово и возвращает результат
func (s *GameServiceImpl) CheckWord(word, target string) string {
	log := s.logger.With(zap.String("method", "CheckWord"))
	log.Debug("Checking word",
		zap.String("word", word),
		zap.String("target", target))

	// Проверяем длину слов
	wordRunes := []rune(word)
	targetRunes := []rune(target)

	if len(wordRunes) != len(targetRunes) {
		log.Error("Words have different lengths",
			zap.Int("word_length", len(wordRunes)),
			zap.Int("target_length", len(targetRunes)))
		return ""
	}

	// Массив для результатов проверки
	result := make([]int, len(wordRunes))
	targetUsed := make([]bool, len(targetRunes))

	// Сначала проверяем точные совпадения (буква на своем месте)
	for i := 0; i < len(wordRunes); i++ {
		if wordRunes[i] == targetRunes[i] {
			result[i] = 2 // 2 означает точное совпадение
			targetUsed[i] = true
		}
	}

	// Затем проверяем наличие буквы в слове, но не на своем месте
	for i := 0; i < len(wordRunes); i++ {
		if result[i] == 2 {
			continue // Уже нашли точное совпадение
		}

		for j := 0; j < len(targetRunes); j++ {
			if targetUsed[j] {
				continue // Эта буква цели уже использована
			}

			if wordRunes[i] == targetRunes[j] {
				result[i] = 1 // 1 означает, что буква есть в слове, но не на своем месте
				targetUsed[j] = true
				break
			}
		}
	}

	// Преобразуем массив в строку
	resultStr := ""
	for _, r := range result {
		resultStr += fmt.Sprintf("%d", r)
	}

	log.Debug("Word check result", zap.String("result", resultStr))
	return resultStr
}

// IsWordCorrect проверяет, совпадает ли слово с целевым
func (s *GameServiceImpl) IsWordCorrect(word, target string) bool {
	return word == target
}

// CalculateReward вычисляет награду на основе ставки, множителя и числа использованных попыток
func (s *GameServiceImpl) CalculateReward(bet float64, multiplier float64, triesUsed, maxTries int) float64 {
	log := s.logger.With(zap.String("method", "CalculateReward"),
		zap.Float64("bet", bet),
		zap.Float64("multiplier", multiplier),
		zap.Int("tries_used", triesUsed),
		zap.Int("max_tries", maxTries))
	log.Debug("Calculating reward")

	if triesUsed <= 0 || maxTries <= 0 || triesUsed > maxTries {
		log.Error("Invalid tries parameters",
			zap.Int("tries_used", triesUsed),
			zap.Int("max_tries", maxTries))
		return 0
	}

	// Базовая награда
	baseReward := bet * multiplier

	// Чем меньше попыток использовано, тем выше награда
	triesMultiplier := 1.0 - (float64(triesUsed-1) * 0.1) // -10% за каждую дополнительную попытку
	if triesMultiplier < 0.5 {
		triesMultiplier = 0.5 // Минимальный множитель - 50% от базовой награды
	}

	reward := baseReward * triesMultiplier

	log.Debug("Reward calculated",
		zap.Float64("base_reward", baseReward),
		zap.Float64("tries_multiplier", triesMultiplier),
		zap.Float64("final_reward", reward))

	return reward
}
