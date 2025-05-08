package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/TakuroBreath/wordle/internal/models"
	"github.com/TakuroBreath/wordle/internal/repository"
	"github.com/google/uuid"
)

// GameServiceImpl представляет собой реализацию GameService
type GameServiceImpl struct {
	gameRepo    models.GameRepository
	redisRepo   repository.RedisRepository
	userService models.UserService
}

// NewGameService создает новый экземпляр GameService
func NewGameService(gameRepo models.GameRepository, redisRepo repository.RedisRepository, userService models.UserService) models.GameService {
	return &GameServiceImpl{
		gameRepo:    gameRepo,
		redisRepo:   redisRepo,
		userService: userService,
	}
}

// CreateGame создает новую игру
func (s *GameServiceImpl) CreateGame(ctx context.Context, game *models.Game) error {
	if game == nil {
		return errors.New("game is nil")
	}

	// Валидация основных полей
	if game.CreatorID == 0 {
		return errors.New("invalid creator ID")
	}
	if game.Word == "" {
		return errors.New("word cannot be empty")
	}
	if game.Length != len([]rune(game.Word)) { // Проверяем длину слова по рунам
		return fmt.Errorf("word length mismatch: specified length %d, actual length %d", game.Length, len([]rune(game.Word)))
	}
	if game.Length <= 0 { // Длина тоже должна быть > 0
		return errors.New("invalid word length")
	}
	if game.Title == "" {
		return errors.New("title cannot be empty")
	}

	// Валидация сложности (пример)
	allowedDifficulties := map[string]bool{"easy": true, "medium": true, "hard": true}
	if !allowedDifficulties[game.Difficulty] {
		return fmt.Errorf("invalid difficulty: %s", game.Difficulty)
	}

	// Валидация валюты
	if game.Currency != models.CurrencyTON && game.Currency != models.CurrencyUSDT {
		return fmt.Errorf("invalid currency: %s", game.Currency)
	}

	// Валидация параметров игры
	if game.MinBet <= 0 {
		return errors.New("min bet must be positive")
	}
	if game.MaxBet < game.MinBet {
		return errors.New("max bet cannot be less than min bet")
	}
	if game.RewardMultiplier < 1.0 { // Множитель должен быть хотя бы 1
		return errors.New("invalid reward multiplier, must be >= 1.0")
	}
	if game.MaxTries <= 0 {
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

	return s.gameRepo.Create(ctx, game)
}

// GetGame получает игру по ID
func (s *GameServiceImpl) GetGame(ctx context.Context, id uuid.UUID) (*models.Game, error) {
	return s.gameRepo.GetByID(ctx, id)
}

// GetUserGames получает список игр пользователя
func (s *GameServiceImpl) GetUserGames(ctx context.Context, userID uint64, limit, offset int) ([]*models.Game, error) {
	return s.gameRepo.GetByUserID(ctx, userID, limit, offset)
}

// UpdateGame обновляет игру
func (s *GameServiceImpl) UpdateGame(ctx context.Context, game *models.Game) error {
	if game == nil {
		return errors.New("game is nil")
	}
	// TODO: Добавить валидацию полей перед обновлением?
	// Получаем существующую игру, чтобы обновить только разрешенные поля
	existingGame, err := s.gameRepo.GetByID(ctx, game.ID)
	if err != nil {
		return fmt.Errorf("game not found for update: %w", err)
	}
	// Обновляем только изменяемые поля (например, Title, Description)
	existingGame.Title = game.Title
	existingGame.Description = game.Description
	// Другие поля, такие как ставки, слово, множитель, обычно не должны меняться после создания
	existingGame.UpdatedAt = time.Now()
	return s.gameRepo.Update(ctx, existingGame)
}

// DeleteGame удаляет игру и возвращает средства из reward pool создателю.
// Важно: Проверка, что вызывающий пользователь является создателем игры, должна быть на уровне handler.
func (s *GameServiceImpl) DeleteGame(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return errors.New("game ID cannot be nil for delete")
	}
	game, err := s.gameRepo.GetByID(ctx, id)
	if err != nil {
		// Обрабатываем случай, когда игра не найдена
		if errors.Is(err, models.ErrGameNotFound) { // Предполагаем наличие ErrGameNotFound
			return errors.New("game not found")
		}
		return fmt.Errorf("failed to get game for deletion: %w", err)
	}

	// Проверяем, что игра не активна (хотя по ТЗ вроде можно удалять активную, но это кажется нелогичным, если кто-то играет)
	// Уточним: ТЗ говорит "Создатель игры может удалить игру, ее reward pool переходит к нему на баланс."
	// Не сказано про статус. Но удалять активную игру с лобби - плохая идея.
	// Добавим проверку, что нет активных лобби для этой игры (потребует LobbyRepository/LobbyService)
	// Пока что оставим проверку на статус Inactive для простоты.
	if game.Status == models.GameStatusActive {
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
		var errUpdateBalance error
		if returnCurrency == models.CurrencyTON {
			errUpdateBalance = s.userService.UpdateTonBalance(ctx, game.CreatorID, returnAmount)
		} else {
			errUpdateBalance = s.userService.UpdateUsdtBalance(ctx, game.CreatorID, returnAmount)
		}
		if errUpdateBalance != nil {
			// Логгируем ошибку, но все равно пытаемся удалить игру?
			// Или возвращаем ошибку, не удаляя игру?
			// Пока вернем ошибку, не удаляя игру, т.к. не смогли вернуть средства.
			return fmt.Errorf("failed to return reward pool to creator %d (%s): %w", game.CreatorID, returnCurrency, errUpdateBalance)
		}
		// Опционально: можно создать транзакцию для записи возврата средств
	}

	// Удаляем игру
	if err := s.gameRepo.Delete(ctx, id); err != nil {
		// Если не смогли удалить игру после возврата средств - это проблема
		// Нужна компенсационная логика или логирование
		return fmt.Errorf("reward pool returned, but failed to delete game %s: %w", id.String(), err)
	}

	return nil
}

// GetActiveGames получает список активных игр
func (s *GameServiceImpl) GetActiveGames(ctx context.Context, limit, offset int) ([]*models.Game, error) {
	return s.gameRepo.GetActive(ctx, limit, offset)
}

// SearchGames ищет игры по параметрам
func (s *GameServiceImpl) SearchGames(ctx context.Context, minBet, maxBet float64, difficulty string, limit, offset int) ([]*models.Game, error) {
	return s.gameRepo.SearchGames(ctx, minBet, maxBet, difficulty, limit, offset)
}

// GetGameStats получает статистику игры
func (s *GameServiceImpl) GetGameStats(ctx context.Context, gameID uuid.UUID) (map[string]interface{}, error) {
	return s.gameRepo.GetGameStats(ctx, gameID)
}

// AddToRewardPool добавляет средства в reward pool игры
func (s *GameServiceImpl) AddToRewardPool(ctx context.Context, gameID uuid.UUID, amount float64) error {
	game, err := s.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		return err
	}

	if amount <= 0 {
		return errors.New("deposit amount must be positive")
	}

	// Проверяем, что игра активна для депозита
	if game.Status != models.GameStatusActive {
		return errors.New("cannot add to reward pool of inactive game")
	}

	// В зависимости от валюты игры добавляем средства в соответствующий пул
	if game.Currency == models.CurrencyTON {
		game.RewardPoolTon += amount
	} else if game.Currency == models.CurrencyUSDT {
		game.RewardPoolUsdt += amount
	} else {
		return fmt.Errorf("unknown currency: %s", game.Currency)
	}

	game.UpdatedAt = time.Now()

	// Обновляем игру в репозитории
	if err := s.gameRepo.Update(ctx, game); err != nil {
		return fmt.Errorf("failed to update reward pool: %w", err)
	}

	return nil
}

// ActivateGame активирует игру после подтверждения депозита
func (s *GameServiceImpl) ActivateGame(ctx context.Context, gameID uuid.UUID) error {
	game, err := s.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		return err
	}

	// Проверяем, что игра неактивна
	if game.Status != models.GameStatusInactive {
		return errors.New("game is already active")
	}

	// Проверяем, что reward pool содержит необходимую сумму
	var rewardPoolBalance float64
	var requiredBalance float64 = game.MaxBet * game.RewardMultiplier

	if game.Currency == models.CurrencyTON {
		rewardPoolBalance = game.RewardPoolTon
	} else if game.Currency == models.CurrencyUSDT {
		rewardPoolBalance = game.RewardPoolUsdt
	}

	if rewardPoolBalance < requiredBalance {
		return fmt.Errorf("insufficient reward pool balance: got %.2f %s, need %.2f",
			rewardPoolBalance, game.Currency, requiredBalance)
	}

	// Активируем игру
	game.Status = models.GameStatusActive
	game.UpdatedAt = time.Now()

	// Обновляем игру в репозитории
	if err := s.gameRepo.Update(ctx, game); err != nil {
		return fmt.Errorf("failed to activate game: %w", err)
	}

	return nil
}

// DeactivateGame деактивирует игру
func (s *GameServiceImpl) DeactivateGame(ctx context.Context, gameID uuid.UUID) error {
	game, err := s.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		return err
	}

	// Проверяем, что игра активна
	if game.Status != models.GameStatusActive {
		return errors.New("game is not active")
	}

	// Деактивируем игру
	game.Status = models.GameStatusInactive
	game.UpdatedAt = time.Now()

	// Обновляем игру в репозитории
	if err := s.gameRepo.Update(ctx, game); err != nil {
		return fmt.Errorf("failed to deactivate game: %w", err)
	}

	return nil
}

// Реализация методов для проверки слова

// CheckWord проверяет слово и возвращает результат проверки
func (s *GameServiceImpl) CheckWord(word, target string) string {
	if len(word) != len(target) {
		return ""
	}

	result := make([]rune, len(word))
	used := make([]bool, len(target))

	// Сначала проверяем точные совпадения (буква на правильной позиции)
	for i, char := range word {
		if i < len(target) && char == rune(target[i]) {
			result[i] = 'G' // Green - правильная буква на правильной позиции
			used[i] = true
		} else {
			result[i] = 'X' // Пока считаем, что буква отсутствует
		}
	}

	// Затем проверяем буквы, которые есть в слове, но на неправильной позиции
	for i, char := range word {
		if result[i] == 'G' {
			continue // Уже обработано
		}

		for j, targetChar := range target {
			if !used[j] && char == targetChar {
				result[i] = 'Y' // Yellow - правильная буква на неправильной позиции
				used[j] = true
				break
			}
		}
	}

	return string(result)
}

// IsWordCorrect проверяет, правильно ли угадано слово
func (s *GameServiceImpl) IsWordCorrect(word, target string) bool {
	return word == target
}

// CalculateReward вычисляет награду за угадывание слова
func (s *GameServiceImpl) CalculateReward(bet float64, multiplier float64, triesUsed, maxTries int) float64 {
	// Базовая награда
	baseReward := bet * multiplier

	// Бонус за быстрое угадывание
	triesBonus := float64(maxTries-triesUsed) / float64(maxTries) * 0.5

	// Итоговая награда
	return baseReward * (1 + triesBonus)
}
