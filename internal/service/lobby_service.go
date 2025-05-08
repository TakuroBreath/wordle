package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/TakuroBreath/wordle/internal/models"
	"github.com/TakuroBreath/wordle/internal/repository"
	"github.com/google/uuid"
)

// LobbyServiceImpl представляет собой реализацию LobbyService
type LobbyServiceImpl struct {
	lobbyRepo          models.LobbyRepository
	gameRepo           models.GameRepository
	attemptRepo        models.AttemptRepository
	redisRepo          repository.RedisRepository // Пока не используется
	userService        models.UserService
	transactionService models.TransactionService
	historyService     models.HistoryService
	// gameService       models.GameService // Может понадобиться для AddToRewardPool
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
	// gameService models.GameService,
) models.LobbyService {
	return &LobbyServiceImpl{
		lobbyRepo:          lobbyRepo,
		gameRepo:           gameRepo,
		attemptRepo:        attemptRepo,
		redisRepo:          redisRepo,
		userService:        userService,
		transactionService: transactionService,
		historyService:     historyService,
		// gameService:        gameService,
	}
}

// CreateLobby создает новое лобби
func (s *LobbyServiceImpl) CreateLobby(ctx context.Context, lobby *models.Lobby) error {
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

	// Проверяем существование игры и ее статус
	game, err := s.gameRepo.GetByID(ctx, lobby.GameID)
	if err != nil {
		if errors.Is(err, models.ErrGameNotFound) {
			return errors.New("game not found")
		}
		return fmt.Errorf("failed to get game: %w", err)
	}
	if game.Status != models.GameStatusActive {
		return errors.New("game is not active")
	}

	// Проверяем ставку пользователя
	if lobby.BetAmount < game.MinBet || lobby.BetAmount > game.MaxBet {
		return fmt.Errorf("bet amount %.2f is out of range [%.2f, %.2f]", lobby.BetAmount, game.MinBet, game.MaxBet)
	}

	// Проверяем баланс пользователя через userService
	hasBalance, err := s.userService.ValidateBalance(ctx, lobby.UserID, lobby.BetAmount, game.Currency)
	if err != nil {
		return fmt.Errorf("failed to validate user balance: %w", err)
	}
	if !hasBalance {
		return fmt.Errorf("insufficient %s balance", game.Currency)
	}

	// Проверяем, нет ли уже активного лобби у пользователя для этой игры
	activeLobby, err := s.lobbyRepo.GetActiveByGameAndUser(ctx, lobby.GameID, lobby.UserID)
	if err != nil {
		if errors.Is(err, models.ErrLobbyNotFound) {
			// Если лобби не найдено, это нормально - продолжаем создание нового
			activeLobby = nil
		} else {
			return fmt.Errorf("failed to check for existing active lobby: %w", err)
		}
	}
	if activeLobby != nil {
		return errors.New("user already has an active lobby for this game")
	}

	// Списываем ставку с баланса пользователя через userService
	if game.Currency == models.CurrencyTON {
		err = s.userService.UpdateTonBalance(ctx, lobby.UserID, -lobby.BetAmount)
	} else if game.Currency == models.CurrencyUSDT {
		err = s.userService.UpdateUsdtBalance(ctx, lobby.UserID, -lobby.BetAmount)
	} else {
		return fmt.Errorf("unknown game currency: %s", game.Currency)
	}
	if err != nil {
		return fmt.Errorf("failed to deduct bet from user balance: %w", err)
	}

	// Создаем транзакцию ставки
	betTx := &models.Transaction{
		UserID:   lobby.UserID,
		Type:     models.TransactionTypeBet,
		Amount:   lobby.BetAmount,
		Currency: game.Currency,
		Status:   models.TransactionStatusCompleted,
		GameID:   &game.ID,
	}
	if err := s.transactionService.CreateTransaction(ctx, betTx); err != nil {
		var refundErr error
		if game.Currency == models.CurrencyTON {
			refundErr = s.userService.UpdateTonBalance(ctx, lobby.UserID, lobby.BetAmount)
		} else {
			refundErr = s.userService.UpdateUsdtBalance(ctx, lobby.UserID, lobby.BetAmount)
		}
		if refundErr != nil {
			return fmt.Errorf("CRITICAL: failed to create bet transaction and failed to refund user %d: %w; refund error: %v", lobby.UserID, err, refundErr)
		}
		return fmt.Errorf("failed to create bet transaction (user balance was refunded): %w", err)
	}

	// Устанавливаем начальные значения для лобби
	lobby.ID = uuid.New()
	lobby.Status = models.LobbyStatusActive // Используем константу (нужно определить)
	lobby.CreatedAt = time.Now()
	lobby.UpdatedAt = lobby.CreatedAt
	lobby.ExpiresAt = lobby.CreatedAt.Add(5 * time.Minute)
	lobby.MaxTries = game.MaxTries
	lobby.TriesUsed = 0
	lobby.PotentialReward = lobby.BetAmount * game.RewardMultiplier
	lobby.Attempts = nil

	// Сохраняем лобби
	if err := s.lobbyRepo.Create(ctx, lobby); err != nil {
		return fmt.Errorf("failed to create lobby after processing bet: %w", err)
	}

	return nil
}

// GetLobby получает лобби по ID
func (s *LobbyServiceImpl) GetLobby(ctx context.Context, id uuid.UUID) (*models.Lobby, error) {
	return s.lobbyRepo.GetByID(ctx, id)
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
func (s *LobbyServiceImpl) GetLobbyStats(ctx context.Context, lobbyID uuid.UUID) (map[string]interface{}, error) {
	return s.lobbyRepo.GetLobbyStats(ctx, lobbyID)
}

// ProcessAttempt обрабатывает попытку угадать слово
func (s *LobbyServiceImpl) ProcessAttempt(ctx context.Context, lobbyID uuid.UUID, word string) ([]int, error) {
	if lobbyID == uuid.Nil {
		return nil, errors.New("lobby ID cannot be nil")
	}
	lobby, err := s.lobbyRepo.GetByID(ctx, lobbyID)
	if err != nil {
		if errors.Is(err, models.ErrLobbyNotFound) {
			return nil, errors.New("lobby not found")
		}
		return nil, fmt.Errorf("failed to get lobby: %w", err)
	}

	if lobby.Status != models.LobbyStatusActive {
		return nil, fmt.Errorf("lobby is not active, current status: %s", lobby.Status)
	}

	if time.Now().After(lobby.ExpiresAt) {
		err := s.handleLobbyFinish(ctx, lobby, nil, models.LobbyStatusFailedExpired)
		if err != nil {
			fmt.Printf("ERROR: Failed to finish expired lobby %s: %v\n", lobbyID, err)
		}
		return nil, errors.New("lobby time expired")
	}

	if lobby.TriesUsed >= lobby.MaxTries {
		err := s.handleLobbyFinish(ctx, lobby, nil, models.LobbyStatusFailedTries)
		if err != nil {
			fmt.Printf("ERROR: Failed to finish lobby %s with no tries left: %v\n", lobbyID, err)
		}
		return nil, errors.New("max tries exceeded")
	}

	game, err := s.gameRepo.GetByID(ctx, lobby.GameID)
	if err != nil {
		if errors.Is(err, models.ErrGameNotFound) {
			_ = s.handleLobbyFinish(ctx, lobby, nil, models.LobbyStatusFailedInternal)
			return nil, errors.New("associated game not found for lobby")
		}
		return nil, fmt.Errorf("failed to get game for lobby: %w", err)
	}

	lowerWord := strings.ToLower(word)
	if len([]rune(lowerWord)) != game.Length {
		return nil, fmt.Errorf("invalid word length: expected %d, got %d", game.Length, len([]rune(lowerWord)))
	}

	attempt := &models.Attempt{
		ID:        uuid.New(),
		GameID:    game.ID,
		LobbyID:   &lobby.ID,
		UserID:    lobby.UserID,
		Word:      lowerWord,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	result := checkWord(lowerWord, strings.ToLower(game.Word)) // Вызов локальной функции
	attempt.Result = result

	if err := s.attemptRepo.Create(ctx, attempt); err != nil {
		return nil, fmt.Errorf("failed to save attempt: %w", err)
	}

	lobby.TriesUsed++
	if err := s.lobbyRepo.UpdateTriesUsed(ctx, lobbyID, lobby.TriesUsed); err != nil {
		fmt.Printf("WARN: Failed to update tries used for lobby %s after attempt %s: %v\n", lobbyID, attempt.ID, err)
	}

	if isWordCorrect(result) { // Вызов локальной функции
		errFinish := s.handleLobbyFinish(ctx, lobby, game, models.LobbyStatusSuccess)
		if errFinish != nil {
			fmt.Printf("ERROR: Failed to finish successful lobby %s: %v\n", lobbyID, errFinish)
		}
		return result, nil
	} else if lobby.TriesUsed >= lobby.MaxTries {
		errFinish := s.handleLobbyFinish(ctx, lobby, game, models.LobbyStatusFailedTries)
		if errFinish != nil {
			fmt.Printf("ERROR: Failed to finish failed lobby %s (no tries left): %v\n", lobbyID, errFinish)
		}
		return result, nil
	}

	return result, nil
}

// handleLobbyFinish обрабатывает завершение лобби (приватный метод)
func (s *LobbyServiceImpl) handleLobbyFinish(ctx context.Context, lobby *models.Lobby, game *models.Game, finalStatus string) error {
	if lobby.Status != models.LobbyStatusActive {
		fmt.Printf("INFO: Attempted to finish lobby %s which is already in status %s\n", lobby.ID, lobby.Status)
		return nil
	}

	lobby.Status = finalStatus
	lobby.UpdatedAt = time.Now()

	var err error
	if game == nil {
		game, err = s.gameRepo.GetByID(ctx, lobby.GameID)
		if err != nil {
			_ = s.lobbyRepo.UpdateStatus(ctx, lobby.ID, finalStatus)
			return fmt.Errorf("failed to get game %s details to finish lobby %s: %w", lobby.GameID, lobby.ID, err)
		}
	}

	var finalReward float64 = 0
	var historyStatus string

	if finalStatus == models.LobbyStatusSuccess {
		historyStatus = models.HistoryStatusPlayerWin
		finalReward = calculateReward(lobby.BetAmount, game.RewardMultiplier, lobby.TriesUsed, lobby.MaxTries) // Вызов локальной функции
		var errUpdateBalance error
		if game.Currency == models.CurrencyTON {
			errUpdateBalance = s.userService.UpdateTonBalance(ctx, lobby.UserID, finalReward)
		} else if game.Currency == models.CurrencyUSDT {
			errUpdateBalance = s.userService.UpdateUsdtBalance(ctx, lobby.UserID, finalReward)
		} else {
			finalReward = 0
			fmt.Printf("ERROR: Unknown currency %s for successful lobby %s, reward not processed\n", game.Currency, lobby.ID)
		}

		if errUpdateBalance != nil {
			fmt.Printf("ERROR: Failed to update balance for successful lobby %s (user %d, amount %.2f %s): %v\n",
				lobby.ID, lobby.UserID, finalReward, game.Currency, errUpdateBalance)
			finalReward = 0
		} else if finalReward > 0 {
			rewardTx := &models.Transaction{
				UserID:   lobby.UserID,
				Type:     models.TransactionTypeReward,
				Amount:   finalReward,
				Currency: game.Currency,
				Status:   models.TransactionStatusCompleted,
				GameID:   &game.ID,
				LobbyID:  &lobby.ID,
			}
			if errTx := s.transactionService.CreateTransaction(ctx, rewardTx); errTx != nil {
				fmt.Printf("ERROR: Failed to create reward transaction for lobby %s: %v\n", lobby.ID, errTx)
			}
		}
	} else { // Лобби завершено неуспешно
		historyStatus = models.HistoryStatusCreatorWin
		// TODO: Перечисление ставки в reward pool
		fmt.Printf("INFO: Lobby %s failed. Status: %s. Bet amount %.2f %s should be transferred to reward pool.\n",
			lobby.ID, finalStatus, lobby.BetAmount, game.Currency)
	}

	// Обновляем статус лобби в базе
	if err := s.lobbyRepo.UpdateStatus(ctx, lobby.ID, finalStatus); err != nil {
		fmt.Printf("ERROR: Failed to update final status '%s' for lobby %s: %v\n", finalStatus, lobby.ID, err)
	}

	// Создаем запись в истории
	history := &models.History{
		UserID:    lobby.UserID,
		GameID:    lobby.GameID,
		LobbyID:   lobby.ID,
		Status:    historyStatus,
		Reward:    finalReward,
		BetAmount: lobby.BetAmount,
		Currency:  game.Currency,
		TriesUsed: lobby.TriesUsed,
	}
	if err := s.historyService.CreateHistory(ctx, history); err != nil {
		fmt.Printf("ERROR: Failed to create history record for lobby %s: %v\n", lobby.ID, err)
	}

	return nil
}

// FinishLobby завершает лобби (публичный метод интерфейса)
func (s *LobbyServiceImpl) FinishLobby(ctx context.Context, lobbyID uuid.UUID, success bool) error {
	lobby, err := s.lobbyRepo.GetByID(ctx, lobbyID)
	if err != nil {
		return fmt.Errorf("cannot finish lobby, failed to get lobby %s: %w", lobbyID, err)
	}
	if lobby.Status != models.LobbyStatusActive {
		return fmt.Errorf("cannot finish lobby %s, it is not active (status: %s)", lobbyID, lobby.Status)
	}

	var finalStatus string
	if success {
		finalStatus = models.LobbyStatusSuccess
	} else {
		finalStatus = models.LobbyStatusCanceled // Используем Canceled для принудительного завершения
	}
	return s.handleLobbyFinish(ctx, lobby, nil, finalStatus)
}

// CheckWord проверяет слово и возвращает результат проверки
func (s *LobbyServiceImpl) CheckWord(word, target string) []int {
	if len(word) != len(target) {
		return nil
	}

	result := make([]int, len(word))
	used := make([]bool, len(target))

	// Сначала проверяем точные совпадения
	for i := 0; i < len(word); i++ {
		if word[i] == target[i] {
			result[i] = 2
			used[i] = true
		}
	}

	// Затем проверяем буквы, которые есть в слове, но не на своем месте
	for i := 0; i < len(word); i++ {
		if result[i] == 2 {
			continue
		}
		for j := 0; j < len(target); j++ {
			if !used[j] && word[i] == target[j] {
				result[i] = 1
				used[j] = true
				break
			}
		}
	}

	return result
}

// IsWordCorrect проверяет, правильно ли угадано слово
func (s *LobbyServiceImpl) IsWordCorrect(word, target string) bool {
	return word == target
}

// CalculateReward вычисляет награду за угадывание слова
func (s *LobbyServiceImpl) CalculateReward(bet float64, multiplier float64, triesUsed, maxTries int) float64 {
	// Базовая награда
	baseReward := bet * multiplier

	// Уменьшаем награду в зависимости от количества попыток
	triesFactor := float64(maxTries-triesUsed) / float64(maxTries)
	reward := baseReward * triesFactor

	// Вычитаем комиссию сервиса (5%)
	reward = reward * 0.95

	return reward
}

// checkWord проверяет слово и возвращает результат []int
func checkWord(word, target string) []int {
	result := make([]int, len(target))
	targetRunes := []rune(target)
	wordRunes := []rune(word)
	usedTarget := make([]bool, len(target))
	usedWord := make([]bool, len(word))

	// 1. Проход на точные совпадения (Green)
	for i := 0; i < len(targetRunes); i++ {
		if i < len(wordRunes) && wordRunes[i] == targetRunes[i] {
			result[i] = 2 // На месте
			usedTarget[i] = true
			usedWord[i] = true
		}
	}

	// 2. Проход на наличие буквы не на своем месте (Yellow)
	for i := 0; i < len(wordRunes); i++ {
		if usedWord[i] { // Пропускаем уже угаданные буквы слова
			continue
		}
		for j := 0; j < len(targetRunes); j++ {
			if !usedTarget[j] && wordRunes[i] == targetRunes[j] {
				result[i] = 1 // Есть в слове
				usedTarget[j] = true
				break // Переходим к следующей букве слова
			}
		}
	}
	return result
}

// isWordCorrect проверяет, является ли результат полным совпадением
func isWordCorrect(result []int) bool {
	for _, r := range result {
		if r != 2 {
			return false
		}
	}
	return true
}

// calculateReward вычисляет награду
func calculateReward(bet float64, multiplier float64, triesUsed, maxTries int) float64 {
	baseReward := bet * multiplier
	if maxTries <= 0 {
		return baseReward
	}
	triesBonus := float64(maxTries-triesUsed) / float64(maxTries) * 0.5
	return baseReward * (1 + triesBonus)
}
