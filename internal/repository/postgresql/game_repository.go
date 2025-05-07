package postgresql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/TakuroBreath/wordle/internal/models"
	"github.com/google/uuid"
)

// GameRepository представляет собой реализацию репозитория для работы с играми
type GameRepository struct {
	db *sql.DB
}

// NewGameRepository создает новый экземпляр GameRepository
func NewGameRepository(db *sql.DB) *GameRepository {
	return &GameRepository{
		db: db,
	}
}

// Create создает новую игру в базе данных
func (r *GameRepository) Create(ctx context.Context, game *models.Game) error {
	query := `
		INSERT INTO games (id, title, word, length, creator_id, difficulty, max_tries, 
			reward_multiplier, currency, prize_pool, min_bet, max_bet, created_at, updated_at, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`

	// Генерация UUID, если он не был установлен
	if game.ID == uuid.Nil {
		game.ID = uuid.New()
	}

	// Установка текущего времени, если оно не было установлено
	now := time.Now()
	if game.CreatedAt.IsZero() {
		game.CreatedAt = now
	}
	game.UpdatedAt = now

	_, err := r.db.ExecContext(
		ctx,
		query,
		game.ID,
		game.Title,
		game.Word,
		game.Length,
		game.CreatorID,
		game.Difficulty,
		game.MaxTries,
		game.RewardMultiplier,
		game.Currency,
		game.PrizePool,
		game.MinBet,
		game.MaxBet,
		game.CreatedAt,
		game.UpdatedAt,
		game.Status,
	)

	if err != nil {
		return fmt.Errorf("failed to create game: %w", err)
	}

	return nil
}

// GetByID получает игру по ID
func (r *GameRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Game, error) {
	query := `
		SELECT id, title, word, length, creator_id, difficulty, max_tries, 
			reward_multiplier, currency, prize_pool, min_bet, max_bet, created_at, updated_at, status
		FROM games
		WHERE id = $1
	`

	var game models.Game

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&game.ID,
		&game.Title,
		&game.Word,
		&game.Length,
		&game.CreatorID,
		&game.Difficulty,
		&game.MaxTries,
		&game.RewardMultiplier,
		&game.Currency,
		&game.PrizePool,
		&game.MinBet,
		&game.MaxBet,
		&game.CreatedAt,
		&game.UpdatedAt,
		&game.Status,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("game not found")
		}
		return nil, fmt.Errorf("failed to get game: %w", err)
	}

	return &game, nil
}

// GetAll получает все игры с пагинацией
func (r *GameRepository) GetAll(ctx context.Context, limit, offset int) ([]*models.Game, error) {
	query := `
		SELECT id, title, word, length, creator_id, difficulty, max_tries, 
			reward_multiplier, currency, prize_pool, min_bet, max_bet, created_at, updated_at, status
		FROM games
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get games: %w", err)
	}
	defer rows.Close()

	var games []*models.Game
	for rows.Next() {
		var game models.Game

		err := rows.Scan(
			&game.ID,
			&game.Title,
			&game.Word,
			&game.Length,
			&game.CreatorID,
			&game.Difficulty,
			&game.MaxTries,
			&game.RewardMultiplier,
			&game.Currency,
			&game.PrizePool,
			&game.MinBet,
			&game.MaxBet,
			&game.CreatedAt,
			&game.UpdatedAt,
			&game.Status,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan game: %w", err)
		}

		games = append(games, &game)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating games: %w", err)
	}

	return games, nil
}

// GetActive получает все активные игры с пагинацией
func (r *GameRepository) GetActive(ctx context.Context, limit, offset int) ([]*models.Game, error) {
	query := `
		SELECT id, title, word, length, creator_id, difficulty, max_tries, 
			reward_multiplier, currency, prize_pool, min_bet, max_bet, created_at, updated_at, status
		FROM games
		WHERE status = 'active'
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get active games: %w", err)
	}
	defer rows.Close()

	var games []*models.Game
	for rows.Next() {
		var game models.Game

		err := rows.Scan(
			&game.ID,
			&game.Title,
			&game.Word,
			&game.Length,
			&game.CreatorID,
			&game.Difficulty,
			&game.MaxTries,
			&game.RewardMultiplier,
			&game.Currency,
			&game.PrizePool,
			&game.MinBet,
			&game.MaxBet,
			&game.CreatedAt,
			&game.UpdatedAt,
			&game.Status,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan game: %w", err)
		}

		games = append(games, &game)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating active games: %w", err)
	}

	return games, nil
}

// GetByCreator получает игры по ID создателя с пагинацией
func (r *GameRepository) GetByCreator(ctx context.Context, creatorID uint64, limit, offset int) ([]*models.Game, error) {
	query := `
		SELECT id, title, word, length, creator_id, difficulty, max_tries, 
			reward_multiplier, currency, prize_pool, min_bet, max_bet, created_at, updated_at, status
		FROM games
		WHERE creator_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, creatorID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get games by creator: %w", err)
	}
	defer rows.Close()

	var games []*models.Game
	for rows.Next() {
		var game models.Game

		err := rows.Scan(
			&game.ID,
			&game.Title,
			&game.Word,
			&game.Length,
			&game.CreatorID,
			&game.Difficulty,
			&game.MaxTries,
			&game.RewardMultiplier,
			&game.Currency,
			&game.PrizePool,
			&game.MinBet,
			&game.MaxBet,
			&game.CreatedAt,
			&game.UpdatedAt,
			&game.Status,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan game: %w", err)
		}

		games = append(games, &game)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating games by creator: %w", err)
	}

	return games, nil
}

// Update обновляет игру в базе данных
func (r *GameRepository) Update(ctx context.Context, game *models.Game) error {
	query := `
		UPDATE games
		SET title = $1, word = $2, length = $3, creator_id = $4, difficulty = $5,
			max_tries = $6, reward_multiplier = $7, currency = $8, prize_pool = $9,
			min_bet = $10, max_bet = $11, updated_at = $12, status = $13
		WHERE id = $14
	`

	game.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(
		ctx,
		query,
		game.Title,
		game.Word,
		game.Length,
		game.CreatorID,
		game.Difficulty,
		game.MaxTries,
		game.RewardMultiplier,
		game.Currency,
		game.PrizePool,
		game.MinBet,
		game.MaxBet,
		game.UpdatedAt,
		game.Status,
		game.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update game: %w", err)
	}

	return nil
}

// Delete удаляет игру из базы данных
func (r *GameRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM games WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete game: %w", err)
	}

	return nil
}
