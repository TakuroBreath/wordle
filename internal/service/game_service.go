package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/TakuroBreath/wordle/internal/logger"
	"github.com/TakuroBreath/wordle/internal/models"
	"github.com/TakuroBreath/wordle/internal/repository"
	"github.com/TakuroBreath/wordle/pkg/metrics"
	otel "github.com/TakuroBreath/wordle/pkg/tracing"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
)

// GameServiceImpl представляет собой реализацию GameService
type GameServiceImpl struct {
	gameRepo       models.GameRepository
	redisRepo      repository.RedisRepository
	userService    models.UserService
	tonService     models.TONService
	commissionRate float64
	logger         *zap.Logger
}

// NewGameService создает новый экземпляр GameService
func NewGameService(
	gameRepo models.GameRepository,
	redisRepo repository.RedisRepository,
	userService models.UserService,
	tonService models.TONService,
	commissionRate float64,
) models.GameService {
	return &GameServiceImpl{
		gameRepo:       gameRepo,
		redisRepo:      redisRepo,
		userService:    userService,
		tonService:     tonService,
		commissionRate: commissionRate,
		logger:         logger.GetLogger(zap.String("service", "game")),
	}
}

// generateShortID генерирует короткий уникальный ID
func generateShortID() string {
	const chars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	result := make([]byte, 8)
	for i := 0; i < 8; i++ {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		result[i] = chars[n.Int64()]
	}
	return string(result)
}

// CreateGame создает новую игру
func (s *GameServiceImpl) CreateGame(ctx context.Context, game *models.Game) error {
	log := s.logger.With(zap.String("method", "CreateGame"))

	return WithTracingVoid(ctx, "GameService", "CreateGame", func(ctx context.Context) error {
		if game == nil {
			log.Error("Game is nil")
			return errors.New("game is nil")
		}

		// Добавляем информацию в span
		span := otel.SpanFromContext(ctx)
		otel.AddAttributesToSpan(span,
			attribute.Int64("creator_id", int64(game.CreatorID)),
			attribute.String("word", game.Word),
			attribute.Int("length", game.Length),
			attribute.String("difficulty", game.Difficulty),
			attribute.Int("max_tries", game.MaxTries),
			attribute.Int("time_limit", game.TimeLimit),
			attribute.String("title", game.Title),
			attribute.Float64("min_bet", game.MinBet),
			attribute.Float64("max_bet", game.MaxBet),
			attribute.Float64("reward_multiplier", game.RewardMultiplier),
			attribute.Float64("deposit_amount", game.DepositAmount),
			attribute.String("currency", game.Currency),
		)

		log.Info("Creating new game",
			zap.Uint64("creator_id", game.CreatorID),
			zap.String("title", game.Title),
			zap.Float64("min_bet", game.MinBet),
			zap.Float64("max_bet", game.MaxBet),
			zap.Float64("reward_multiplier", game.RewardMultiplier))

		// Валидация основных полей
		if game.CreatorID == 0 {
			return errors.New("invalid creator ID")
		}
		if game.Word == "" {
			return errors.New("word cannot be empty")
		}

		// Нормализуем слово (lowercase)
		game.Word = strings.ToLower(game.Word)
		game.Length = len([]rune(game.Word))

		if game.Length <= 0 || game.Length > 15 {
			return errors.New("invalid word length (must be 1-15)")
		}
		if game.Title == "" {
			return errors.New("title cannot be empty")
		}
		if game.MaxTries <= 0 || game.MaxTries > 20 {
			return errors.New("max_tries must be between 1 and 20")
		}
		if game.TimeLimit <= 0 {
			game.TimeLimit = 5 // 5 минут по умолчанию
		}
		if game.TimeLimit > 60 {
			return errors.New("time_limit cannot exceed 60 minutes")
		}

		// Валидация сложности
		allowedDifficulties := map[string]bool{"easy": true, "medium": true, "hard": true}
		if !allowedDifficulties[game.Difficulty] {
			return fmt.Errorf("invalid difficulty: %s", game.Difficulty)
		}

		// Валидация валюты
		if game.Currency != models.CurrencyTON && game.Currency != models.CurrencyUSDT {
			return fmt.Errorf("invalid currency: %s", game.Currency)
		}

		// Валидация параметров ставок
		if game.MinBet <= 0 {
			return errors.New("min bet must be positive")
		}
		if game.MaxBet < game.MinBet {
			return errors.New("max bet cannot be less than min bet")
		}
		if game.RewardMultiplier < 1.0 {
			return errors.New("reward multiplier must be >= 1.0")
		}

		// Вычисляем минимальный депозит
		requiredDeposit := game.MaxBet * game.RewardMultiplier
		if game.DepositAmount < requiredDeposit {
			game.DepositAmount = requiredDeposit
		}

		// Генерируем короткий ID
		shortID := generateShortID()
		// Проверяем уникальность (в реальности нужно повторять генерацию при коллизии)
		for attempts := 0; attempts < 10; attempts++ {
			existing, err := s.gameRepo.GetByShortID(ctx, shortID)
			if err != nil || existing == nil {
				break
			}
			shortID = generateShortID()
		}

		// Установка начальных значений
		game.ID = uuid.New()
		game.ShortID = shortID
		game.Status = models.GameStatusPending // Ожидает депозита
		game.RewardPoolTon = 0.0
		game.RewardPoolUsdt = 0.0
		game.ReservedAmount = 0.0
		now := time.Now()
		game.CreatedAt = now
		game.UpdatedAt = now

		otel.AddAttributesToSpan(span,
			attribute.String("game_id", game.ID.String()),
			attribute.String("short_id", game.ShortID))

		log.Debug("Game parameters validated, creating game",
			zap.String("game_id", game.ID.String()),
			zap.String("short_id", game.ShortID),
			zap.String("status", game.Status))

		err := s.gameRepo.Create(ctx, game)
		if err != nil {
			log.Error("Failed to create game", zap.Error(err))
			return err
		}

		log.Info("Game created successfully",
			zap.String("game_id", game.ID.String()),
			zap.String("short_id", game.ShortID))

		return nil
	})
}

// GetGame получает игру по ID
func (s *GameServiceImpl) GetGame(ctx context.Context, id uuid.UUID) (*models.Game, error) {
	log := s.logger.With(zap.String("method", "GetGame"), zap.String("game_id", id.String()))
	log.Info("Getting game by ID")

	result, err := WithTracing(ctx, "GameService", "GetGame", func(ctx context.Context) (any, error) {
		span := otel.SpanFromContext(ctx)
		otel.AddAttributesToSpan(span, attribute.String("game_id", id.String()))
		return s.gameRepo.GetByID(ctx, id)
	})

	if err != nil {
		log.Error("Failed to get game", zap.Error(err))
		return nil, err
	}

	game := result.(*models.Game)
	log.Info("Game retrieved successfully",
		zap.String("title", game.Title),
		zap.String("status", game.Status))
	return game, nil
}

// GetGameByShortID получает игру по короткому ID
func (s *GameServiceImpl) GetGameByShortID(ctx context.Context, shortID string) (*models.Game, error) {
	log := s.logger.With(zap.String("method", "GetGameByShortID"), zap.String("short_id", shortID))
	log.Info("Getting game by short ID")

	game, err := s.gameRepo.GetByShortID(ctx, shortID)
	if err != nil {
		log.Error("Failed to get game by short ID", zap.Error(err))
		return nil, err
	}

	return game, nil
}

// GetUserGames получает список игр пользователя
func (s *GameServiceImpl) GetUserGames(ctx context.Context, userID uint64, limit, offset int) ([]*models.Game, error) {
	log := s.logger.With(zap.String("method", "GetUserGames"),
		zap.Uint64("user_id", userID),
		zap.Int("limit", limit),
		zap.Int("offset", offset))
	log.Info("Getting user games")

	result, err := WithTracing(ctx, "GameService", "GetUserGames", func(ctx context.Context) (any, error) {
		span := otel.SpanFromContext(ctx)
		otel.AddAttributesToSpan(span,
			attribute.Int64("user_id", int64(userID)),
			attribute.Int("limit", limit),
			attribute.Int("offset", offset))
		return s.gameRepo.GetByUserID(ctx, userID, limit, offset)
	})

	if err != nil {
		log.Error("Failed to get user games", zap.Error(err))
		return nil, err
	}

	games := result.([]*models.Game)
	log.Info("User games retrieved successfully", zap.Int("count", len(games)))
	return games, nil
}

// GetCreatedGames получает список игр, созданных пользователем
func (s *GameServiceImpl) GetCreatedGames(ctx context.Context, userID uint64, limit, offset int) ([]*models.Game, error) {
	log := s.logger.With(zap.String("method", "GetCreatedGames"),
		zap.Uint64("user_id", userID),
		zap.Int("limit", limit),
		zap.Int("offset", offset))
	log.Info("Getting games created by user")

	result, err := WithTracing(ctx, "GameService", "GetCreatedGames", func(ctx context.Context) (any, error) {
		span := otel.SpanFromContext(ctx)
		otel.AddAttributesToSpan(span,
			attribute.Int64("user_id", int64(userID)),
			attribute.Int("limit", limit),
			attribute.Int("offset", offset))
		return s.gameRepo.GetByCreator(ctx, userID, limit, offset)
	})

	if err != nil {
		log.Error("Failed to get created games", zap.Error(err))
		return nil, err
	}

	games := result.([]*models.Game)
	log.Info("Created games retrieved successfully", zap.Int("count", len(games)))
	return games, nil
}

// UpdateGame обновляет игру
func (s *GameServiceImpl) UpdateGame(ctx context.Context, game *models.Game) error {
	log := s.logger.With(zap.String("method", "UpdateGame"))

	if game == nil {
		return errors.New("game is nil")
	}

	log.Info("Updating game", zap.String("game_id", game.ID.String()))

	existingGame, err := s.gameRepo.GetByID(ctx, game.ID)
	if err != nil {
		return fmt.Errorf("game not found: %w", err)
	}

	// Обновляем только разрешённые поля
	existingGame.Title = game.Title
	existingGame.Description = game.Description
	existingGame.UpdatedAt = time.Now()

	if err := s.gameRepo.Update(ctx, existingGame); err != nil {
		return err
	}

	log.Info("Game updated successfully", zap.String("game_id", game.ID.String()))
	return nil
}

// DeleteGame удаляет игру
func (s *GameServiceImpl) DeleteGame(ctx context.Context, id uuid.UUID) error {
	log := s.logger.With(zap.String("method", "DeleteGame"), zap.String("game_id", id.String()))
	log.Info("Deleting game")

	game, err := s.gameRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("game not found: %w", err)
	}

	// Нельзя удалить активную игру
	if game.Status == models.GameStatusActive {
		return errors.New("cannot delete an active game")
	}

	// Возвращаем средства создателю
	var returnAmount float64
	if game.Currency == models.CurrencyTON {
		returnAmount = game.RewardPoolTon
	} else {
		returnAmount = game.RewardPoolUsdt
	}

	if returnAmount > 0 {
		log.Info("Returning reward pool to creator",
			zap.Float64("amount", returnAmount),
			zap.String("currency", game.Currency))

		if game.Currency == models.CurrencyTON {
			err = s.userService.UpdateTonBalance(ctx, game.CreatorID, returnAmount)
		} else {
			err = s.userService.UpdateUsdtBalance(ctx, game.CreatorID, returnAmount)
		}

		if err != nil {
			return fmt.Errorf("failed to return funds to creator: %w", err)
		}
	}

	if err := s.gameRepo.Delete(ctx, id); err != nil {
		return err
	}

	log.Info("Game deleted successfully")
	return nil
}

// GetActiveGames получает список активных игр
func (s *GameServiceImpl) GetActiveGames(ctx context.Context, limit, offset int) ([]*models.Game, error) {
	log := s.logger.With(zap.String("method", "GetActiveGames"))
	log.Info("Getting active games")

	result, err := WithTracing(ctx, "GameService", "GetActiveGames", func(ctx context.Context) (any, error) {
		return s.gameRepo.GetActive(ctx, limit, offset)
	})

	if err != nil {
		return nil, err
	}

	games := result.([]*models.Game)
	log.Info("Active games retrieved", zap.Int("count", len(games)))
	return games, nil
}

// GetPendingGames получает список игр, ожидающих депозита
func (s *GameServiceImpl) GetPendingGames(ctx context.Context, limit, offset int) ([]*models.Game, error) {
	return s.gameRepo.GetPending(ctx, limit, offset)
}

// SearchGames ищет игры по параметрам
func (s *GameServiceImpl) SearchGames(ctx context.Context, minBet, maxBet float64, difficulty string, limit, offset int) ([]*models.Game, error) {
	return s.gameRepo.SearchGames(ctx, minBet, maxBet, difficulty, limit, offset)
}

// GetGameStats получает статистику игры
func (s *GameServiceImpl) GetGameStats(ctx context.Context, gameID uuid.UUID) (map[string]any, error) {
	return s.gameRepo.GetGameStats(ctx, gameID)
}

// AddToRewardPool добавляет средства в reward pool игры
func (s *GameServiceImpl) AddToRewardPool(ctx context.Context, gameID uuid.UUID, amount float64) error {
	log := s.logger.With(zap.String("method", "AddToRewardPool"),
		zap.String("game_id", gameID.String()),
		zap.Float64("amount", amount))
	log.Info("Adding to reward pool")

	game, err := s.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		return err
	}

	if amount <= 0 {
		return errors.New("amount must be positive")
	}

	// Добавляем в пул
	if game.Currency == models.CurrencyTON {
		game.RewardPoolTon += amount
	} else {
		game.RewardPoolUsdt += amount
	}
	game.UpdatedAt = time.Now()

	if err := s.gameRepo.Update(ctx, game); err != nil {
		return err
	}

	log.Info("Reward pool updated successfully")
	return nil
}

// ActivateGame активирует игру
func (s *GameServiceImpl) ActivateGame(ctx context.Context, gameID uuid.UUID) error {
	log := s.logger.With(zap.String("method", "ActivateGame"), zap.String("game_id", gameID.String()))
	log.Info("Activating game")

	game, err := s.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		return err
	}

	if game.Status == models.GameStatusActive {
		return errors.New("game is already active")
	}

	// Проверяем достаточность средств
	requiredDeposit := game.GetRequiredDeposit()
	var currentPool float64
	if game.Currency == models.CurrencyTON {
		currentPool = game.RewardPoolTon
	} else {
		currentPool = game.RewardPoolUsdt
	}

	if currentPool < requiredDeposit {
		return fmt.Errorf("insufficient funds: need %.4f %s, have %.4f",
			requiredDeposit, game.Currency, currentPool)
	}

	err = s.gameRepo.UpdateStatus(ctx, gameID, models.GameStatusActive)
	if err != nil {
		return err
	}

	log.Info("Game activated successfully")
	metrics.IncrementGameStart()
	return nil
}

// DeactivateGame деактивирует игру
func (s *GameServiceImpl) DeactivateGame(ctx context.Context, gameID uuid.UUID) error {
	log := s.logger.With(zap.String("method", "DeactivateGame"), zap.String("game_id", gameID.String()))
	log.Info("Deactivating game")

	game, err := s.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		return err
	}

	if game.Status != models.GameStatusActive {
		return errors.New("game is not active")
	}

	// Проверяем, нет ли зарезервированных средств (активных игроков)
	if game.ReservedAmount > 0 {
		return errors.New("cannot deactivate game with active players")
	}

	err = s.gameRepo.UpdateStatus(ctx, gameID, models.GameStatusInactive)
	if err != nil {
		return err
	}

	log.Info("Game deactivated successfully")
	return nil
}

// ReserveForBet резервирует средства для ставки
func (s *GameServiceImpl) ReserveForBet(ctx context.Context, gameID uuid.UUID, betAmount float64, multiplier float64) error {
	potentialReward := betAmount * multiplier
	return s.gameRepo.IncrementReservedAmount(ctx, gameID, potentialReward)
}

// ReleaseReservation освобождает зарезервированные средства
func (s *GameServiceImpl) ReleaseReservation(ctx context.Context, gameID uuid.UUID, amount float64) error {
	return s.gameRepo.DecrementReservedAmount(ctx, gameID, amount)
}

// GetPaymentInfo генерирует информацию для оплаты депозита игры
func (s *GameServiceImpl) GetPaymentInfo(ctx context.Context, gameID uuid.UUID) (*models.PaymentInfo, error) {
	game, err := s.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		return nil, err
	}

	if game.Status != models.GameStatusPending {
		return nil, errors.New("game is not pending payment")
	}

	// Генерируем комментарий для идентификации платежа
	comment := fmt.Sprintf("GD_%s_%d", game.ShortID, time.Now().Unix())

	masterWallet := s.tonService.GetMasterWalletAddress()
	deepLink := s.tonService.GeneratePaymentDeepLink(masterWallet, game.DepositAmount, comment)

	return &models.PaymentInfo{
		Address:  masterWallet,
		Amount:   game.DepositAmount,
		Currency: game.Currency,
		Comment:  comment,
		GameID:   game.ShortID,
		ExpireAt: time.Now().Add(30 * time.Minute).Unix(),
		DeepLink: deepLink,
	}, nil
}

// CheckWord проверяет слово и возвращает результат в виде строки
func (s *GameServiceImpl) CheckWord(word, target string) string {
	wordRunes := []rune(strings.ToLower(word))
	targetRunes := []rune(strings.ToLower(target))

	if len(wordRunes) != len(targetRunes) {
		return ""
	}

	result := make([]int, len(wordRunes))
	targetUsed := make([]bool, len(targetRunes))

	// Сначала точные совпадения
	for i := 0; i < len(wordRunes); i++ {
		if wordRunes[i] == targetRunes[i] {
			result[i] = 2
			targetUsed[i] = true
		}
	}

	// Затем буквы на неправильных позициях
	for i := 0; i < len(wordRunes); i++ {
		if result[i] == 2 {
			continue
		}
		for j := 0; j < len(targetRunes); j++ {
			if !targetUsed[j] && wordRunes[i] == targetRunes[j] {
				result[i] = 1
				targetUsed[j] = true
				break
			}
		}
	}

	// Преобразуем в строку
	var sb strings.Builder
	for _, r := range result {
		sb.WriteString(fmt.Sprintf("%d", r))
	}
	return sb.String()
}

// IsWordCorrect проверяет, угадано ли слово
func (s *GameServiceImpl) IsWordCorrect(word, target string) bool {
	return strings.ToLower(word) == strings.ToLower(target)
}

// CalculateReward вычисляет награду с учётом комиссии
func (s *GameServiceImpl) CalculateReward(bet float64, multiplier float64, triesUsed, maxTries int) float64 {
	if triesUsed <= 0 || maxTries <= 0 || triesUsed > maxTries {
		return 0
	}

	// Базовая награда
	baseReward := bet * multiplier

	// Бонус за быстрое угадывание (чем меньше попыток - тем больше бонус)
	triesBonus := 1.0 + (float64(maxTries-triesUsed)/float64(maxTries))*0.5

	grossReward := baseReward * triesBonus

	// Вычитаем комиссию 5%
	netReward := grossReward * (1 - s.commissionRate)

	return netReward
}
