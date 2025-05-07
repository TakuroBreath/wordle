package postgresql

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/TakuroBreath/wordle/internal/models"
	"github.com/google/uuid"
)

// AttemptRepository представляет собой реализацию репозитория для работы с попытками
type AttemptRepository struct {
	db *sql.DB
}

// NewAttemptRepository создает новый экземпляр AttemptRepository
func NewAttemptRepository(db *sql.DB) *AttemptRepository {
	return &AttemptRepository{
		db: db,
	}
}

// Create создает новую попытку в базе данных
func (r *AttemptRepository) Create(ctx context.Context, attempt *models.Attempt) error {
	query := `
		INSERT INTO attempts (id, game_id, user_id, word, result, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	// Генерация UUID, если он не был установлен
	if attempt.ID == uuid.Nil {
		attempt.ID = uuid.New()
	}

	// Установка текущего времени
	now := time.Now()
	if attempt.CreatedAt.IsZero() {
		attempt.CreatedAt = now
	}
	attempt.UpdatedAt = now

	// Преобразование результата в JSON
	resultJSON, err := json.Marshal(attempt.Result)
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}

	_, err = r.db.ExecContext(
		ctx,
		query,
		attempt.ID,
		attempt.GameID,
		attempt.UserID,
		attempt.Word,
		resultJSON,
		attempt.CreatedAt,
		attempt.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create attempt: %w", err)
	}

	return nil
}

// GetByID получает попытку по ID
func (r *AttemptRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Attempt, error) {
	query := `
		SELECT id, game_id, user_id, word, result, created_at, updated_at
		FROM attempts
		WHERE id = $1
	`

	var attempt models.Attempt
	var resultJSON []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&attempt.ID,
		&attempt.GameID,
		&attempt.UserID,
		&attempt.Word,
		&resultJSON,
		&attempt.CreatedAt,
		&attempt.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("attempt not found")
		}
		return nil, fmt.Errorf("failed to get attempt: %w", err)
	}

	// Преобразование JSON обратно в массив
	if err := json.Unmarshal(resultJSON, &attempt.Result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal result: %w", err)
	}

	return &attempt, nil
}

// GetByGameID получает все попытки для конкретной игры
func (r *AttemptRepository) GetByGameID(ctx context.Context, gameID uuid.UUID) ([]*models.Attempt, error) {
	query := `
		SELECT id, game_id, user_id, word, result, created_at, updated_at
		FROM attempts
		WHERE game_id = $1
		ORDER BY created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, gameID)
	if err != nil {
		return nil, fmt.Errorf("failed to get attempts by game: %w", err)
	}
	defer rows.Close()

	var attempts []*models.Attempt
	for rows.Next() {
		var attempt models.Attempt
		var resultJSON []byte

		err := rows.Scan(
			&attempt.ID,
			&attempt.GameID,
			&attempt.UserID,
			&attempt.Word,
			&resultJSON,
			&attempt.CreatedAt,
			&attempt.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan attempt: %w", err)
		}

		if err := json.Unmarshal(resultJSON, &attempt.Result); err != nil {
			return nil, fmt.Errorf("failed to unmarshal result: %w", err)
		}

		attempts = append(attempts, &attempt)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating attempts: %w", err)
	}

	return attempts, nil
}

// GetByUserID получает все попытки пользователя
func (r *AttemptRepository) GetByUserID(ctx context.Context, userID uint64) ([]*models.Attempt, error) {
	query := `
		SELECT id, game_id, user_id, word, result, created_at, updated_at
		FROM attempts
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get attempts by user: %w", err)
	}
	defer rows.Close()

	var attempts []*models.Attempt
	for rows.Next() {
		var attempt models.Attempt
		var resultJSON []byte

		err := rows.Scan(
			&attempt.ID,
			&attempt.GameID,
			&attempt.UserID,
			&attempt.Word,
			&resultJSON,
			&attempt.CreatedAt,
			&attempt.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan attempt: %w", err)
		}

		if err := json.Unmarshal(resultJSON, &attempt.Result); err != nil {
			return nil, fmt.Errorf("failed to unmarshal result: %w", err)
		}

		attempts = append(attempts, &attempt)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating attempts: %w", err)
	}

	return attempts, nil
}

// GetByGameAndUser получает попытки конкретного пользователя в конкретной игре
func (r *AttemptRepository) GetByGameAndUser(ctx context.Context, gameID uuid.UUID, userID uint64) ([]*models.Attempt, error) {
	query := `
		SELECT id, game_id, user_id, word, result, created_at, updated_at
		FROM attempts
		WHERE game_id = $1 AND user_id = $2
		ORDER BY created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, gameID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get attempts by game and user: %w", err)
	}
	defer rows.Close()

	var attempts []*models.Attempt
	for rows.Next() {
		var attempt models.Attempt
		var resultJSON []byte

		err := rows.Scan(
			&attempt.ID,
			&attempt.GameID,
			&attempt.UserID,
			&attempt.Word,
			&resultJSON,
			&attempt.CreatedAt,
			&attempt.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan attempt: %w", err)
		}

		if err := json.Unmarshal(resultJSON, &attempt.Result); err != nil {
			return nil, fmt.Errorf("failed to unmarshal result: %w", err)
		}

		attempts = append(attempts, &attempt)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating attempts: %w", err)
	}

	return attempts, nil
}

// CountByGameAndUser возвращает количество попыток пользователя в игре
func (r *AttemptRepository) CountByGameAndUser(ctx context.Context, gameID uuid.UUID, userID uint64) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM attempts
		WHERE game_id = $1 AND user_id = $2
	`

	var count int
	err := r.db.QueryRowContext(ctx, query, gameID, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count attempts: %w", err)
	}

	return count, nil
}

// Delete удаляет попытку
func (r *AttemptRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM attempts WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete attempt: %w", err)
	}

	return nil
}

// GetLastAttempt получает последнюю попытку пользователя в игре
func (r *AttemptRepository) GetLastAttempt(ctx context.Context, gameID uuid.UUID, userID uint64) (*models.Attempt, error) {
	query := `
		SELECT id, game_id, user_id, word, result, created_at, updated_at
		FROM attempts
		WHERE game_id = $1 AND user_id = $2
		ORDER BY created_at DESC
		LIMIT 1
	`

	var attempt models.Attempt
	var resultJSON []byte

	err := r.db.QueryRowContext(ctx, query, gameID, userID).Scan(
		&attempt.ID,
		&attempt.GameID,
		&attempt.UserID,
		&attempt.Word,
		&resultJSON,
		&attempt.CreatedAt,
		&attempt.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no attempts found")
		}
		return nil, fmt.Errorf("failed to get last attempt: %w", err)
	}

	if err := json.Unmarshal(resultJSON, &attempt.Result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal result: %w", err)
	}

	return &attempt, nil
}

// GetSuccessfulAttempts получает успешные попытки (где все буквы угаданы)
func (r *AttemptRepository) GetSuccessfulAttempts(ctx context.Context, gameID uuid.UUID) ([]*models.Attempt, error) {
	query := `
		SELECT id, game_id, user_id, word, result, created_at, updated_at
		FROM attempts
		WHERE game_id = $1 AND result @> '[2,2,2,2,2]'::jsonb
		ORDER BY created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, gameID)
	if err != nil {
		return nil, fmt.Errorf("failed to get successful attempts: %w", err)
	}
	defer rows.Close()

	var attempts []*models.Attempt
	for rows.Next() {
		var attempt models.Attempt
		var resultJSON []byte

		err := rows.Scan(
			&attempt.ID,
			&attempt.GameID,
			&attempt.UserID,
			&attempt.Word,
			&resultJSON,
			&attempt.CreatedAt,
			&attempt.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan attempt: %w", err)
		}

		if err := json.Unmarshal(resultJSON, &attempt.Result); err != nil {
			return nil, fmt.Errorf("failed to unmarshal result: %w", err)
		}

		attempts = append(attempts, &attempt)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating successful attempts: %w", err)
	}

	return attempts, nil
}

// GetAttemptsByDateRange получает попытки за определенный период
func (r *AttemptRepository) GetAttemptsByDateRange(ctx context.Context, startDate, endDate time.Time) ([]*models.Attempt, error) {
	query := `
		SELECT id, game_id, user_id, word, result, created_at, updated_at
		FROM attempts
		WHERE created_at BETWEEN $1 AND $2
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get attempts by date range: %w", err)
	}
	defer rows.Close()

	var attempts []*models.Attempt
	for rows.Next() {
		var attempt models.Attempt
		var resultJSON []byte

		err := rows.Scan(
			&attempt.ID,
			&attempt.GameID,
			&attempt.UserID,
			&attempt.Word,
			&resultJSON,
			&attempt.CreatedAt,
			&attempt.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan attempt: %w", err)
		}

		if err := json.Unmarshal(resultJSON, &attempt.Result); err != nil {
			return nil, fmt.Errorf("failed to unmarshal result: %w", err)
		}

		attempts = append(attempts, &attempt)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating attempts by date range: %w", err)
	}

	return attempts, nil
}

// GetUserStats получает статистику попыток пользователя
func (r *AttemptRepository) GetUserStats(ctx context.Context, userID uint64) (totalAttempts, successfulAttempts int, err error) {
	query := `
		SELECT 
			COUNT(*) as total_attempts,
			COUNT(*) FILTER (WHERE result @> '[2,2,2,2,2]'::jsonb) as successful_attempts
		FROM attempts
		WHERE user_id = $1
	`

	err = r.db.QueryRowContext(ctx, query, userID).Scan(&totalAttempts, &successfulAttempts)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get user stats: %w", err)
	}

	return totalAttempts, successfulAttempts, nil
}

// GetGameStats получает статистику попыток для игры
func (r *AttemptRepository) GetGameStats(ctx context.Context, gameID uuid.UUID) (totalAttempts, uniquePlayers int, err error) {
	query := `
		SELECT 
			COUNT(*) as total_attempts,
			COUNT(DISTINCT user_id) as unique_players
		FROM attempts
		WHERE game_id = $1
	`

	err = r.db.QueryRowContext(ctx, query, gameID).Scan(&totalAttempts, &uniquePlayers)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get game stats: %w", err)
	}

	return totalAttempts, uniquePlayers, nil
}
