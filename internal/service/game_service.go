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

// GameService представляет собой реализацию сервиса для работы с играми
type GameService struct {
	repo      models.GameRepository
	redisRepo repository.RedisRepository
}

// NewGameService создает новый экземпляр GameService
func NewGameService(repo models.GameRepository, redisRepo repository.RedisRepository) *GameService {
	return &GameService{
		repo:      repo,
		redisRepo: redisRepo,
	}
}

// Реализация GameService
func (s *GameService) Create(ctx context.Context, game *models.Game) error {
	// Проверка валидности данных
	if game.Word == "" {
		return errors.New("word cannot be empty")
	}

	if game.MaxTries <= 0 {
		return errors.New("max tries must be greater than 0")
	}

	if game.MinBet <= 0 || game.MaxBet <= 0 || game.MinBet > game.MaxBet {
		return errors.New("invalid bet range")
	}

	if game.RewardMultiplier <= 0 {
		return errors.New("reward multiplier must be greater than 0")
	}

	if game.Currency != "TON" && game.Currency != "USDT" {
		return errors.New("currency must be TON or USDT")
	}

	// Установка длины слова
	game.Length = len(game.Word)

	// Установка статуса
	game.Status = "active"

	// Генерация ID, если он не был установлен
	if game.ID == uuid.Nil {
		game.ID = uuid.New()
	}

	// Установка текущего времени, если оно не было установлено
	if game.CreatedAt.IsZero() {
		game.CreatedAt = time.Now()
	}

	// Сохранение игры в базе данных
	return s.repo.Create(ctx, game)
}

func (s *GameService) GetByID(ctx context.Context, id uuid.UUID) (*models.Game, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *GameService) GetAll(ctx context.Context, limit, offset int) ([]*models.Game, error) {
	return s.repo.GetAll(ctx, limit, offset)
}

func (s *GameService) GetAllGames(ctx context.Context) ([]*models.Game, error) {
	return s.repo.GetAll(ctx, 0, 0)
}

func (s *GameService) GetActive(ctx context.Context, limit, offset int) ([]*models.Game, error) {
	return s.repo.GetActive(ctx, limit, offset)
}

func (s *GameService) GetByCreator(ctx context.Context, creatorID uuid.UUID, limit, offset int) ([]*models.Game, error) {
	return s.repo.GetByCreator(ctx, creatorID, limit, offset)
}

func (s *GameService) Update(ctx context.Context, game *models.Game) error {
	// Проверка существования игры
	existingGame, err := s.repo.GetByID(ctx, game.ID)
	if err != nil {
		return fmt.Errorf("failed to get game: %w", err)
	}

	// Проверка, что пользователь является создателем игры
	if existingGame.CreatorID != game.CreatorID {
		return errors.New("only creator can update the game")
	}

	// Проверка валидности данных
	if game.Word == "" {
		return errors.New("word cannot be empty")
	}

	if game.MaxTries <= 0 {
		return errors.New("max tries must be greater than 0")
	}

	if game.MinBet <= 0 || game.MaxBet <= 0 || game.MinBet > game.MaxBet {
		return errors.New("invalid bet range")
	}

	if game.RewardMultiplier <= 0 {
		return errors.New("reward multiplier must be greater than 0")
	}

	if game.Currency != "TON" && game.Currency != "USDT" {
		return errors.New("currency must be TON or USDT")
	}

	// Обновление длины слова
	game.Length = len(game.Word)

	// Сохранение игры в базе данных
	return s.repo.Update(ctx, game)
}

func (s *GameService) Delete(ctx context.Context, id uuid.UUID) error {
	// Проверка существования игры
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get game: %w", err)
	}

	// Удаление игры (установка статуса inactive)
	return s.repo.Delete(ctx, id)
}

// Реализация методов для проверки слова

// CheckWord проверяет слово и возвращает результат проверки
func (s *GameService) CheckWord(word, target string) string {
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
func (s *GameService) IsWordCorrect(word, target string) bool {
	return word == target
}

// CalculateReward вычисляет награду за угадывание слова
func (s *GameService) CalculateReward(bet float64, multiplier float64, triesUsed, maxTries int) float64 {
	// Базовая награда
	baseReward := bet * multiplier

	// Бонус за быстрое угадывание
	triesBonus := float64(maxTries-triesUsed) / float64(maxTries) * 0.5

	// Итоговая награда
	return baseReward * (1 + triesBonus)
}
