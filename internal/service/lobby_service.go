package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/TakuroBreath/wordle/internal/logger"
	"github.com/TakuroBreath/wordle/internal/models"
	"github.com/TakuroBreath/wordle/internal/repository"
	"github.com/TakuroBreath/wordle/pkg/metrics"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// LobbyServiceImpl представляет собой реализацию LobbyService
type LobbyServiceImpl struct {
	lobbyRepo          models.LobbyRepository
	gameRepo           models.GameRepository
	attemptRepo        models.AttemptRepository
	redisRepo          repository.RedisRepository
	userService        models.UserService
	transactionService models.TransactionService
	historyService     models.HistoryService
	tonService         models.TONService
	commissionRate     float64
	logger             *zap.Logger
}

// NewLobbyService создает новый экземпляр LobbyService
func NewLobbyService(
	lobbyRepo models.LobbyRepository,
	gameRepo models.GameRepository,
	attemptRepo models.AttemptRepository,
	redisRepo repository.RedisRepository,
	userService models.UserService,
	transactionService models.TransactionService,
	historyService models.HistoryService,
	tonService models.TONService,
	commissionRate float64,
) models.LobbyService {
	return &LobbyServiceImpl{
		lobbyRepo:          lobbyRepo,
		gameRepo:           gameRepo,
		attemptRepo:        attemptRepo,
		redisRepo:          redisRepo,
		userService:        userService,
		transactionService: transactionService,
		historyService:     historyService,
		tonService:         tonService,
		commissionRate:     commissionRate,
		logger:             logger.GetLogger(zap.String("service", "lobby")),
	}
}

// CreateLobby создает новое лобби (для оплаты с баланса)
func (s *LobbyServiceImpl) CreateLobby(ctx context.Context, lobby *models.Lobby) error {
	log := s.logger.With(zap.String("method", "CreateLobby"))

	if lobby == nil {
		return errors.New("lobby is nil")
	}
	if lobby.GameID == uuid.Nil {
		return errors.New("game ID is required")
	}
	if lobby.UserID == 0 {
		return errors.New("user ID is required")
	}
	if lobby.BetAmount <= 0 {
		return errors.New("bet amount must be positive")
	}

	log.Info("Creating lobby",
		zap.Uint64("user_id", lobby.UserID),
		zap.String("game_id", lobby.GameID.String()),
		zap.Float64("bet_amount", lobby.BetAmount))

	// Получаем игру
	game, err := s.gameRepo.GetByID(ctx, lobby.GameID)
	if err != nil {
		return fmt.Errorf("game not found: %w", err)
	}

	if game.Status != models.GameStatusActive {
		return errors.New("game is not active")
	}

	// Проверяем ставку
	if lobby.BetAmount < game.MinBet || lobby.BetAmount > game.MaxBet {
		return fmt.Errorf("bet amount must be between %.4f and %.4f", game.MinBet, game.MaxBet)
	}

	// Проверяем, может ли игра принять ставку
	if !game.CanAcceptBet(lobby.BetAmount) {
		return errors.New("game cannot accept bet: insufficient reward pool")
	}

	// Проверяем баланс пользователя
	hasBalance, err := s.userService.ValidateBalance(ctx, lobby.UserID, lobby.BetAmount, game.Currency)
	if err != nil {
		return fmt.Errorf("failed to validate balance: %w", err)
	}
	if !hasBalance {
		return fmt.Errorf("insufficient %s balance", game.Currency)
	}

	// Проверяем, нет ли уже активного лобби
	existingLobby, err := s.lobbyRepo.GetActiveByGameAndUser(ctx, lobby.GameID, lobby.UserID)
	if err == nil && existingLobby != nil {
		return errors.New("user already has an active lobby for this game")
	}

	// Списываем ставку с баланса
	if game.Currency == models.CurrencyTON {
		err = s.userService.UpdateTonBalance(ctx, lobby.UserID, -lobby.BetAmount)
	} else {
		err = s.userService.UpdateUsdtBalance(ctx, lobby.UserID, -lobby.BetAmount)
	}
	if err != nil {
		return fmt.Errorf("failed to deduct bet: %w", err)
	}

	// Резервируем средства в игре
	potentialReward := lobby.BetAmount * game.RewardMultiplier
	if err := s.gameRepo.IncrementReservedAmount(ctx, game.ID, potentialReward); err != nil {
		// Возвращаем деньги
		if game.Currency == models.CurrencyTON {
			_ = s.userService.UpdateTonBalance(ctx, lobby.UserID, lobby.BetAmount)
		} else {
			_ = s.userService.UpdateUsdtBalance(ctx, lobby.UserID, lobby.BetAmount)
		}
		return fmt.Errorf("failed to reserve funds: %w", err)
	}

	// Создаём транзакцию ставки
	betTx := &models.Transaction{
		UserID:      lobby.UserID,
		Type:        models.TransactionTypeBet,
		Amount:      lobby.BetAmount,
		Currency:    game.Currency,
		Status:      models.TransactionStatusCompleted,
		GameID:      &game.ID,
		GameShortID: game.ShortID,
		Description: fmt.Sprintf("Bet for game %s", game.Title),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := s.transactionService.CreateTransaction(ctx, betTx); err != nil {
		log.Error("Failed to create bet transaction", zap.Error(err))
	}

	// Создаём лобби
	now := time.Now()
	expiresAt := now.Add(time.Duration(game.TimeLimit) * time.Minute)

	lobby.ID = uuid.New()
	lobby.GameShortID = game.ShortID
	lobby.Status = models.LobbyStatusActive
	lobby.MaxTries = game.MaxTries
	lobby.TriesUsed = 0
	lobby.PotentialReward = potentialReward
	lobby.Currency = game.Currency
	lobby.StartedAt = &now
	lobby.ExpiresAt = expiresAt
	lobby.CreatedAt = now
	lobby.UpdatedAt = now
	lobby.Attempts = nil

	if err := s.lobbyRepo.Create(ctx, lobby); err != nil {
		// Откатываем
		_ = s.gameRepo.DecrementReservedAmount(ctx, game.ID, potentialReward)
		if game.Currency == models.CurrencyTON {
			_ = s.userService.UpdateTonBalance(ctx, lobby.UserID, lobby.BetAmount)
		} else {
			_ = s.userService.UpdateUsdtBalance(ctx, lobby.UserID, lobby.BetAmount)
		}
		return fmt.Errorf("failed to create lobby: %w", err)
	}

	log.Info("Lobby created successfully",
		zap.String("lobby_id", lobby.ID.String()),
		zap.Uint64("user_id", lobby.UserID))

	return nil
}

// GetJoinPaymentInfo генерирует информацию для оплаты вступления в игру через блокчейн
func (s *LobbyServiceImpl) GetJoinPaymentInfo(ctx context.Context, gameShortID string, userID uint64, betAmount float64) (*models.PaymentInfo, error) {
	log := s.logger.With(zap.String("method", "GetJoinPaymentInfo"))

	game, err := s.gameRepo.GetByShortID(ctx, gameShortID)
	if err != nil {
		return nil, fmt.Errorf("game not found: %w", err)
	}

	if game.Status != models.GameStatusActive {
		return nil, errors.New("game is not active")
	}

	if betAmount < game.MinBet || betAmount > game.MaxBet {
		return nil, fmt.Errorf("bet amount must be between %.4f and %.4f", game.MinBet, game.MaxBet)
	}

	if !game.CanAcceptBet(betAmount) {
		return nil, errors.New("game cannot accept bet")
	}

	// Генерируем комментарий
	comment := fmt.Sprintf("LB_%s_%d", gameShortID, time.Now().Unix())

	masterWallet := s.tonService.GetMasterWalletAddress()
	deepLink := s.tonService.GeneratePaymentDeepLink(masterWallet, betAmount, comment)

	log.Info("Generated payment info",
		zap.String("game_short_id", gameShortID),
		zap.Float64("bet_amount", betAmount),
		zap.String("comment", comment))

	return &models.PaymentInfo{
		Address:  masterWallet,
		Amount:   betAmount,
		Currency: game.Currency,
		Comment:  comment,
		GameID:   gameShortID,
		ExpireAt: time.Now().Add(15 * time.Minute).Unix(),
		DeepLink: deepLink,
	}, nil
}

// GetLobby получает лобби по ID
func (s *LobbyServiceImpl) GetLobby(ctx context.Context, id uuid.UUID) (*models.Lobby, error) {
	lobby, err := s.lobbyRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Загружаем попытки
	attempts, err := s.attemptRepo.GetByLobbyID(ctx, id, 100, 0)
	if err == nil {
		lobby.Attempts = make([]models.Attempt, len(attempts))
		for i, a := range attempts {
			lobby.Attempts[i] = *a
		}
	}

	return lobby, nil
}

// GetGameLobbies получает список лобби для игры
func (s *LobbyServiceImpl) GetGameLobbies(ctx context.Context, gameID uuid.UUID, limit, offset int) ([]*models.Lobby, error) {
	return s.lobbyRepo.GetByGameID(ctx, gameID, limit, offset)
}

// GetUserLobbies получает список лобби пользователя
func (s *LobbyServiceImpl) GetUserLobbies(ctx context.Context, userID uint64, limit, offset int) ([]*models.Lobby, error) {
	return s.lobbyRepo.GetByUserID(ctx, userID, limit, offset)
}

// UpdateLobby обновляет лобби
func (s *LobbyServiceImpl) UpdateLobby(ctx context.Context, lobby *models.Lobby) error {
	if lobby == nil {
		return errors.New("lobby is nil")
	}
	lobby.UpdatedAt = time.Now()
	return s.lobbyRepo.Update(ctx, lobby)
}

// DeleteLobby удаляет лобби
func (s *LobbyServiceImpl) DeleteLobby(ctx context.Context, id uuid.UUID) error {
	return s.lobbyRepo.Delete(ctx, id)
}

// GetActiveLobbies получает список активных лобби
func (s *LobbyServiceImpl) GetActiveLobbies(ctx context.Context, limit, offset int) ([]*models.Lobby, error) {
	return s.lobbyRepo.GetActive(ctx, limit, offset)
}

// UpdateLobbyStatus обновляет статус лобби
func (s *LobbyServiceImpl) UpdateLobbyStatus(ctx context.Context, id uuid.UUID, status string) error {
	return s.lobbyRepo.UpdateStatus(ctx, id, status)
}

// UpdateTriesUsed обновляет количество использованных попыток
func (s *LobbyServiceImpl) UpdateTriesUsed(ctx context.Context, id uuid.UUID, triesUsed int) error {
	return s.lobbyRepo.UpdateTriesUsed(ctx, id, triesUsed)
}

// GetExpiredLobbies получает список истекших лобби
func (s *LobbyServiceImpl) GetExpiredLobbies(ctx context.Context) ([]*models.Lobby, error) {
	return s.lobbyRepo.GetExpired(ctx)
}

// GetActiveLobbyByGameAndUser получает активное лобби пользователя для игры
func (s *LobbyServiceImpl) GetActiveLobbyByGameAndUser(ctx context.Context, gameID uuid.UUID, userID uint64) (*models.Lobby, error) {
	return s.lobbyRepo.GetActiveByGameAndUser(ctx, gameID, userID)
}

// ExtendLobbyTime продлевает время жизни лобби
func (s *LobbyServiceImpl) ExtendLobbyTime(ctx context.Context, id uuid.UUID, duration int) error {
	return s.lobbyRepo.ExtendExpirationTime(ctx, id, duration)
}

// GetLobbyStats получает статистику лобби
func (s *LobbyServiceImpl) GetLobbyStats(ctx context.Context, lobbyID uuid.UUID) (map[string]any, error) {
	return s.lobbyRepo.GetLobbyStats(ctx, lobbyID)
}

// StartLobby запускает лобби (если оно было в статусе pending)
func (s *LobbyServiceImpl) StartLobby(ctx context.Context, lobbyID uuid.UUID) error {
	lobby, err := s.lobbyRepo.GetByID(ctx, lobbyID)
	if err != nil {
		return err
	}

	if lobby.Status != models.LobbyStatusPending {
		return errors.New("lobby is not pending")
	}

	game, err := s.gameRepo.GetByID(ctx, lobby.GameID)
	if err != nil {
		return err
	}

	now := time.Now()
	expiresAt := now.Add(time.Duration(game.TimeLimit) * time.Minute)

	return s.lobbyRepo.StartLobby(ctx, lobbyID, expiresAt)
}

// ProcessAttempt обрабатывает попытку угадать слово
func (s *LobbyServiceImpl) ProcessAttempt(ctx context.Context, lobbyID uuid.UUID, word string) ([]int, error) {
	log := s.logger.With(zap.String("method", "ProcessAttempt"),
		zap.String("lobby_id", lobbyID.String()))

	if lobbyID == uuid.Nil {
		return nil, errors.New("lobby ID cannot be nil")
	}

	lobby, err := s.lobbyRepo.GetByID(ctx, lobbyID)
	if err != nil {
		metrics.RecordError("attempt_validation")
		return nil, fmt.Errorf("failed to get lobby: %w", err)
	}

	if lobby.Status != models.LobbyStatusActive {
		return nil, fmt.Errorf("lobby is not active, status: %s", lobby.Status)
	}

	// Проверяем время
	if lobby.IsExpired() {
		log.Info("Lobby time expired", zap.String("lobby_id", lobbyID.String()))
		_ = s.handleLobbyFinish(ctx, lobby, nil, models.LobbyStatusFailedExpired)
		return nil, errors.New("lobby time expired")
	}

	// Проверяем попытки
	if lobby.TriesUsed >= lobby.MaxTries {
		log.Info("Max tries exceeded", zap.String("lobby_id", lobbyID.String()))
		_ = s.handleLobbyFinish(ctx, lobby, nil, models.LobbyStatusFailedTries)
		return nil, errors.New("max tries exceeded")
	}

	// Получаем игру
	game, err := s.gameRepo.GetByID(ctx, lobby.GameID)
	if err != nil {
		_ = s.handleLobbyFinish(ctx, lobby, nil, models.LobbyStatusFailedInternal)
		return nil, errors.New("game not found")
	}

	// Нормализуем слово
	word = strings.ToLower(strings.TrimSpace(word))

	// Проверяем длину
	if len([]rune(word)) != game.Length {
		return nil, fmt.Errorf("invalid word length: expected %d, got %d", game.Length, len([]rune(word)))
	}

	// Проверяем слово
	result := s.CheckWord(word, game.Word)

	// Создаём попытку
	attempt := &models.Attempt{
		ID:        uuid.New(),
		GameID:    game.ID,
		LobbyID:   &lobby.ID,
		UserID:    lobby.UserID,
		Word:      word,
		Result:    result,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.attemptRepo.Create(ctx, attempt); err != nil {
		return nil, fmt.Errorf("failed to save attempt: %w", err)
	}

	// Обновляем попытки
	lobby.TriesUsed++
	if err := s.lobbyRepo.UpdateTriesUsed(ctx, lobbyID, lobby.TriesUsed); err != nil {
		log.Warn("Failed to update tries used", zap.Error(err))
	}

	log.Info("Attempt processed",
		zap.String("word", word),
		zap.Int("tries_used", lobby.TriesUsed),
		zap.Int("max_tries", lobby.MaxTries))

	// Проверяем результат
	if s.isWordCorrect(result) {
		log.Info("Word guessed correctly!")
		_ = s.handleLobbyFinish(ctx, lobby, game, models.LobbyStatusSuccess)
		return result, nil
	}

	// Проверяем, остались ли попытки
	if lobby.TriesUsed >= lobby.MaxTries {
		log.Info("No more tries left")
		_ = s.handleLobbyFinish(ctx, lobby, game, models.LobbyStatusFailedTries)
	}

	return result, nil
}

// handleLobbyFinish обрабатывает завершение лобби
func (s *LobbyServiceImpl) handleLobbyFinish(ctx context.Context, lobby *models.Lobby, game *models.Game, finalStatus string) error {
	log := s.logger.With(zap.String("method", "handleLobbyFinish"),
		zap.String("lobby_id", lobby.ID.String()),
		zap.String("status", finalStatus))

	if lobby.Status != models.LobbyStatusActive {
		log.Debug("Lobby already finished", zap.String("current_status", lobby.Status))
		return nil
	}

	// Загружаем игру если не передана
	var err error
	if game == nil {
		game, err = s.gameRepo.GetByID(ctx, lobby.GameID)
		if err != nil {
			_ = s.lobbyRepo.UpdateStatus(ctx, lobby.ID, finalStatus)
			return fmt.Errorf("failed to get game: %w", err)
		}
	}

	var reward float64
	var historyStatus string

	if finalStatus == models.LobbyStatusSuccess {
		// Игрок выиграл
		historyStatus = models.HistoryStatusPlayerWin
		reward = s.CalculateReward(lobby.BetAmount, game.RewardMultiplier, lobby.TriesUsed, lobby.MaxTries)

		log.Info("Player won",
			zap.Float64("bet", lobby.BetAmount),
			zap.Float64("reward", reward))

		// Начисляем награду игроку
		if game.Currency == models.CurrencyTON {
			err = s.userService.UpdateTonBalance(ctx, lobby.UserID, reward)
		} else {
			err = s.userService.UpdateUsdtBalance(ctx, lobby.UserID, reward)
		}
		if err != nil {
			log.Error("Failed to credit reward", zap.Error(err))
		}

		// Списываем из пула игры
		if game.Currency == models.CurrencyTON {
			game.RewardPoolTon -= reward
		} else {
			game.RewardPoolUsdt -= reward
		}

		// Создаём транзакцию награды
		rewardTx := &models.Transaction{
			UserID:      lobby.UserID,
			Type:        models.TransactionTypeReward,
			Amount:      reward,
			Currency:    game.Currency,
			Status:      models.TransactionStatusCompleted,
			GameID:      &game.ID,
			GameShortID: game.ShortID,
			LobbyID:     &lobby.ID,
			Description: fmt.Sprintf("Reward for winning game %s", game.Title),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		_ = s.transactionService.CreateTransaction(ctx, rewardTx)

		// Обновляем статистику пользователя
		_ = s.userService.IncrementWins(ctx, lobby.UserID)

		metrics.IncrementGameComplete(lobby.TriesUsed)
	} else {
		// Игрок проиграл
		historyStatus = models.HistoryStatusCreatorWin

		log.Info("Player lost",
			zap.Float64("bet", lobby.BetAmount),
			zap.String("reason", finalStatus))

		// Ставка остаётся в пуле игры (уже была добавлена при создании)
		// Комиссия берётся с потенциальной награды
		commission := lobby.BetAmount * s.commissionRate

		// Начисляем комиссию сервису (в реальности - отдельный кошелёк)
		// Здесь просто логируем
		log.Info("Commission earned", zap.Float64("amount", commission))

		// Обновляем статистику пользователя
		_ = s.userService.IncrementLosses(ctx, lobby.UserID)

		metrics.IncrementGameAbandoned()
	}

	// Освобождаем резерв
	if err := s.gameRepo.DecrementReservedAmount(ctx, game.ID, lobby.PotentialReward); err != nil {
		log.Error("Failed to release reservation", zap.Error(err))
	}

	// Обновляем игру
	if err := s.gameRepo.Update(ctx, game); err != nil {
		log.Error("Failed to update game", zap.Error(err))
	}

	// Обновляем статус лобби
	if err := s.lobbyRepo.UpdateStatus(ctx, lobby.ID, finalStatus); err != nil {
		log.Error("Failed to update lobby status", zap.Error(err))
	}

	// Создаём запись в истории
	history := &models.History{
		UserID:    lobby.UserID,
		GameID:    lobby.GameID,
		LobbyID:   lobby.ID,
		Status:    historyStatus,
		Reward:    reward,
		BetAmount: lobby.BetAmount,
		Currency:  game.Currency,
		TriesUsed: lobby.TriesUsed,
	}
	if err := s.historyService.CreateHistory(ctx, history); err != nil {
		log.Error("Failed to create history", zap.Error(err))
	}

	return nil
}

// FinishLobby завершает лобби принудительно
func (s *LobbyServiceImpl) FinishLobby(ctx context.Context, lobbyID uuid.UUID, success bool) error {
	lobby, err := s.lobbyRepo.GetByID(ctx, lobbyID)
	if err != nil {
		return fmt.Errorf("lobby not found: %w", err)
	}

	if lobby.Status != models.LobbyStatusActive {
		return fmt.Errorf("lobby is not active, status: %s", lobby.Status)
	}

	var status string
	if success {
		status = models.LobbyStatusSuccess
	} else {
		status = models.LobbyStatusCanceled
	}

	return s.handleLobbyFinish(ctx, lobby, nil, status)
}

// CheckWord проверяет слово и возвращает результат []int
func (s *LobbyServiceImpl) CheckWord(word, target string) []int {
	wordRunes := []rune(strings.ToLower(word))
	targetRunes := []rune(strings.ToLower(target))

	if len(wordRunes) != len(targetRunes) {
		return nil
	}

	result := make([]int, len(targetRunes))
	targetUsed := make([]bool, len(targetRunes))
	wordUsed := make([]bool, len(wordRunes))

	// Первый проход: точные совпадения (зелёные)
	for i := 0; i < len(wordRunes); i++ {
		if wordRunes[i] == targetRunes[i] {
			result[i] = 2
			targetUsed[i] = true
			wordUsed[i] = true
		}
	}

	// Второй проход: буквы есть, но не на месте (жёлтые)
	for i := 0; i < len(wordRunes); i++ {
		if wordUsed[i] {
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

	return result
}

// isWordCorrect проверяет, является ли результат полным совпадением
func (s *LobbyServiceImpl) isWordCorrect(result []int) bool {
	for _, r := range result {
		if r != 2 {
			return false
		}
	}
	return true
}

// IsWordCorrect проверяет, правильно ли угадано слово
func (s *LobbyServiceImpl) IsWordCorrect(word, target string) bool {
	return strings.ToLower(word) == strings.ToLower(target)
}

// CalculateReward вычисляет награду с учётом комиссии
func (s *LobbyServiceImpl) CalculateReward(bet float64, multiplier float64, triesUsed, maxTries int) float64 {
	if triesUsed <= 0 || maxTries <= 0 || triesUsed > maxTries {
		return 0
	}

	// Базовая награда
	baseReward := bet * multiplier

	// Бонус за быстрое угадывание
	triesBonus := 1.0 + (float64(maxTries-triesUsed)/float64(maxTries))*0.5

	grossReward := baseReward * triesBonus

	// Вычитаем комиссию 5%
	netReward := grossReward * (1 - s.commissionRate)

	return netReward
}
