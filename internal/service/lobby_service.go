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

// LobbyService представляет собой реализацию сервиса для работы с лобби
type LobbyService struct {
	repo        models.LobbyRepository
	gameRepo    models.GameRepository
	userRepo    models.UserRepository
	attemptRepo models.AttemptRepository
	redisRepo   repository.RedisRepository
}

// NewLobbyService создает новый экземпляр LobbyService
func NewLobbyService(
	repo models.LobbyRepository,
	gameRepo models.GameRepository,
	userRepo models.UserRepository,
	attemptRepo models.AttemptRepository,
	redisRepo repository.RedisRepository,
) *LobbyService {
	return &LobbyService{
		repo:        repo,
		gameRepo:    gameRepo,
		userRepo:    userRepo,
		attemptRepo: attemptRepo,
		redisRepo:   redisRepo,
	}
}

// Create создает новое лобби
func (s *LobbyService) Create(ctx context.Context, lobby *models.Lobby) error {
	// Проверка валидности данных
	if lobby.GameID == uuid.Nil {
		return errors.New("game ID cannot be empty")
	}

	if lobby.UserID == uuid.Nil {
		return errors.New("user ID cannot be empty")
	}

	if lobby.MaxTries <= 0 {
		return errors.New("max tries must be greater than 0")
	}

	if lobby.Bet <= 0 {
		return errors.New("bet must be greater than 0")
	}

	// Проверка существования игры
	game, err := s.gameRepo.GetByID(ctx, lobby.GameID)
	if err != nil {
		return fmt.Errorf("failed to get game: %w", err)
	}

	// Проверка, что игра активна
	if game.Status != "active" {
		return errors.New("game is not active")
	}

	// Проверка, что ставка находится в допустимом диапазоне
	if lobby.Bet < game.MinBet || lobby.Bet > game.MaxBet {
		return fmt.Errorf("bet must be between %f and %f", game.MinBet, game.MaxBet)
	}

	// Проверка существования пользователя
	user, err := s.userRepo.GetByID(ctx, lobby.UserID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Проверка, что у пользователя достаточно средств
	if user.Balance < lobby.Bet {
		return errors.New("insufficient funds")
	}

	// Генерация ID, если он не был установлен
	if lobby.ID == uuid.Nil {
		lobby.ID = uuid.New()
	}

	// Установка текущего времени, если оно не было установлено
	if lobby.CreatedAt.IsZero() {
		lobby.CreatedAt = time.Now()
	}

	// Установка времени истечения (например, 24 часа)
	if lobby.ExpiresAt.IsZero() {
		lobby.ExpiresAt = lobby.CreatedAt.Add(24 * time.Hour)
	}

	// Установка статуса
	lobby.Status = "active"

	// Установка количества использованных попыток
	lobby.TriesUsed = 0

	// Расчет потенциального выигрыша
	lobby.PotentialReward = s.calculatePotentialReward(game.RewardMultiplier, lobby.Bet)

	// Списание средств с баланса пользователя
	user.Balance -= lobby.Bet
	err = s.userRepo.Update(ctx, user)
	if err != nil {
		return fmt.Errorf("failed to update user balance: %w", err)
	}

	// Увеличение призового фонда игры
	game.PrizePool += lobby.Bet
	err = s.gameRepo.Update(ctx, game)
	if err != nil {
		return fmt.Errorf("failed to update game prize pool: %w", err)
	}

	// Сохранение лобби в базе данных
	return s.repo.Create(ctx, lobby)
}

// GetByID получает лобби по ID
func (s *LobbyService) GetByID(ctx context.Context, id uuid.UUID) (*models.Lobby, error) {
	lobby, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Получение попыток для лобби
	attempts, err := s.attemptRepo.GetByLobbyID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get attempts: %w", err)
	}

	// Преобразование типа []*models.Attempt в []models.Attempt
	lobbyAttempts := make([]models.Attempt, len(attempts))
	for i, attempt := range attempts {
		lobbyAttempts[i] = *attempt
	}

	lobby.Attempts = lobbyAttempts

	return lobby, nil
}

// GetByUserID получает лобби пользователя
func (s *LobbyService) GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.Lobby, error) {
	return s.repo.GetByUserID(ctx, userID, limit, offset)
}

// GetByGameID получает лобби для игры
func (s *LobbyService) GetByGameID(ctx context.Context, gameID uuid.UUID, limit, offset int) ([]*models.Lobby, error) {
	return s.repo.GetByGameID(ctx, gameID, limit, offset)
}

// GetActive получает активные лобби
func (s *LobbyService) GetActive(ctx context.Context, limit, offset int) ([]*models.Lobby, error) {
	return s.repo.GetActive(ctx, limit, offset)
}

// Update обновляет данные лобби
func (s *LobbyService) Update(ctx context.Context, lobby *models.Lobby) error {
	// Проверка существования лобби
	existingLobby, err := s.repo.GetByID(ctx, lobby.ID)
	if err != nil {
		return fmt.Errorf("failed to get lobby: %w", err)
	}

	// Проверка, что лобби активно
	if existingLobby.Status != "active" {
		return errors.New("lobby is not active")
	}

	// Сохранение лобби в базе данных
	return s.repo.Update(ctx, lobby)
}

// Delete удаляет лобби
func (s *LobbyService) Delete(ctx context.Context, id uuid.UUID) error {
	// Проверка существования лобби
	lobby, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get lobby: %w", err)
	}

	// Проверка, что лобби активно
	if lobby.Status != "active" {
		return errors.New("lobby is not active")
	}

	// Получение игры
	game, err := s.gameRepo.GetByID(ctx, lobby.GameID)
	if err != nil {
		return fmt.Errorf("failed to get game: %w", err)
	}

	// Получение пользователя
	user, err := s.userRepo.GetByID(ctx, lobby.UserID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Возврат ставки пользователю
	user.Balance += lobby.Bet
	err = s.userRepo.Update(ctx, user)
	if err != nil {
		return fmt.Errorf("failed to update user balance: %w", err)
	}

	// Уменьшение призового фонда игры
	game.PrizePool -= lobby.Bet
	err = s.gameRepo.Update(ctx, game)
	if err != nil {
		return fmt.Errorf("failed to update game prize pool: %w", err)
	}

	// Удаление лобби (установка статуса inactive)
	return s.repo.Delete(ctx, id)
}

// JoinGame создает новое лобби для присоединения к игре
func (s *LobbyService) JoinGame(ctx context.Context, gameID, userID uuid.UUID, bet float64) (*models.Lobby, error) {
	// Проверка существования игры
	game, err := s.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		return nil, fmt.Errorf("failed to get game: %w", err)
	}

	// Проверка, что игра активна
	if game.Status != "active" {
		return nil, errors.New("game is not active")
	}

	// Проверка, что ставка находится в допустимом диапазоне
	if bet < game.MinBet || bet > game.MaxBet {
		return nil, fmt.Errorf("bet must be between %f and %f", game.MinBet, game.MaxBet)
	}

	// Проверка существования пользователя
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Проверка, что у пользователя достаточно средств
	if user.Balance < bet {
		return nil, errors.New("insufficient funds")
	}

	// Создание нового лобби
	lobby := &models.Lobby{
		ID:              uuid.New(),
		GameID:          gameID,
		UserID:          userID,
		MaxTries:        game.MaxTries,
		TriesUsed:       0,
		Bet:             bet,
		PotentialReward: s.calculatePotentialReward(game.RewardMultiplier, bet),
		CreatedAt:       time.Now(),
		ExpiresAt:       time.Now().Add(24 * time.Hour),
		Status:          "active",
		Attempts:        []models.Attempt{},
	}

	// Списание средств с баланса пользователя
	user.Balance -= bet
	err = s.userRepo.Update(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to update user balance: %w", err)
	}

	// Увеличение призового фонда игры
	game.PrizePool += bet
	err = s.gameRepo.Update(ctx, game)
	if err != nil {
		return nil, fmt.Errorf("failed to update game prize pool: %w", err)
	}

	// Сохранение лобби в базе данных
	err = s.repo.Create(ctx, lobby)
	if err != nil {
		return nil, err
	}

	return lobby, nil
}

// MakeAttempt делает попытку угадать слово
func (s *LobbyService) MakeAttempt(ctx context.Context, lobbyID uuid.UUID, word string) (*models.Attempt, error) {
	// Проверка существования лобби
	lobby, err := s.repo.GetByID(ctx, lobbyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get lobby: %w", err)
	}

	// Проверка, что лобби активно
	if lobby.Status != "active" {
		return nil, errors.New("lobby is not active")
	}

	// Проверка, что не превышено максимальное количество попыток
	if lobby.TriesUsed >= lobby.MaxTries {
		return nil, errors.New("maximum number of attempts reached")
	}

	// Получение игры
	game, err := s.gameRepo.GetByID(ctx, lobby.GameID)
	if err != nil {
		return nil, fmt.Errorf("failed to get game: %w", err)
	}

	// Проверка длины слова
	if len(word) != game.Length {
		return nil, fmt.Errorf("word must be %d characters long", game.Length)
	}

	// Создание экземпляра GameService для проверки слова
	gameService := NewGameService(s.gameRepo, s.redisRepo)

	// Проверка слова
	feedback := gameService.CheckWord(word, game.Word)

	// Создание новой попытки
	attempt := &models.Attempt{
		ID:        uuid.New(),
		LobbyID:   lobbyID,
		Word:      word,
		Feedback:  feedback,
		CreatedAt: time.Now(),
	}

	// Сохранение попытки в базе данных
	err = s.attemptRepo.Create(ctx, attempt)
	if err != nil {
		return nil, err
	}

	// Увеличение счетчика использованных попыток
	lobby.TriesUsed++

	// Проверка, угадано ли слово
	isCorrect := gameService.IsWordCorrect(word, game.Word)
	if isCorrect {
		// Слово угадано, обновление статуса лобби
		lobby.Status = "inactive"

		// Получение пользователя
		user, err := s.userRepo.GetByID(ctx, lobby.UserID)
		if err != nil {
			return nil, fmt.Errorf("failed to get user: %w", err)
		}

		// Расчет награды
		reward := gameService.CalculateReward(lobby.Bet, game.RewardMultiplier, lobby.TriesUsed, lobby.MaxTries)

		// Начисление награды пользователю
		user.Balance += reward
		user.Wins++
		err = s.userRepo.Update(ctx, user)
		if err != nil {
			return nil, fmt.Errorf("failed to update user balance: %w", err)
		}

		// Уменьшение призового фонда игры
		game.PrizePool -= reward
		err = s.gameRepo.Update(ctx, game)
		if err != nil {
			return nil, fmt.Errorf("failed to update game prize pool: %w", err)
		}
	} else if lobby.TriesUsed >= lobby.MaxTries {
		// Превышено максимальное количество попыток, обновление статуса лобби
		lobby.Status = "inactive"

		// Получение пользователя
		user, err := s.userRepo.GetByID(ctx, lobby.UserID)
		if err != nil {
			return nil, fmt.Errorf("failed to get user: %w", err)
		}

		// Увеличение счетчика поражений
		user.Losses++
		err = s.userRepo.Update(ctx, user)
		if err != nil {
			return nil, fmt.Errorf("failed to update user stats: %w", err)
		}
	}

	// Обновление лобби в базе данных
	err = s.repo.Update(ctx, lobby)
	if err != nil {
		return nil, err
	}

	return attempt, nil
}

// calculatePotentialReward вычисляет потенциальную награду за угадывание слова
func (s *LobbyService) calculatePotentialReward(multiplier, bet float64) float64 {
	return bet * multiplier
}
