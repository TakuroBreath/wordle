package postgresql

import (
	"context"
	"database/sql"
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
	attempt.CreatedAt = now
	attempt.UpdatedAt = now

	_, err := r.db.ExecContext(
		ctx,
		query,
		attempt.ID,
		attempt.GameID,
		attempt.UserID,
		attempt.Word,
		attempt.Result,
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

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&attempt.ID,
		&attempt.GameID,
		&attempt.UserID,
		&attempt.Word,
		&attempt.Result,
		&attempt.CreatedAt,
		&attempt.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("attempt not found")
		}
		return nil, fmt.Errorf("failed to get attempt: %w", err)
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
		return nil, fmt.Errorf("failed to get attempts: %w", err)
	}
	defer rows.Close()

	var attempts []*models.Attempt
	for rows.Next() {
		var attempt models.Attempt

		err := rows.Scan(
			&attempt.ID,
			&attempt.GameID,
			&attempt.UserID,
			&attempt.Word,
			&attempt.Result,
			&attempt.CreatedAt,
			&attempt.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan attempt: %w", err)
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
		return nil, fmt.Errorf("failed to get user attempts: %w", err)
	}
	defer rows.Close()

	var attempts []*models.Attempt
	for rows.Next() {
		var attempt models.Attempt

		err := rows.Scan(
			&attempt.ID,
			&attempt.GameID,
			&attempt.UserID,
			&attempt.Word,
			&attempt.Result,
			&attempt.CreatedAt,
			&attempt.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan attempt: %w", err)
		}

		attempts = append(attempts, &attempt)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating user attempts: %w", err)
	}

	return attempts, nil
}

// GetByLobbyID получает все попытки для конкретного лобби
func (r *AttemptRepository) GetByLobbyID(ctx context.Context, lobbyID uuid.UUID) ([]*models.Attempt, error) {
	query := `
		SELECT a.id, a.game_id, a.user_id, a.word, a.result, a.created_at, a.updated_at
		FROM attempts a
		JOIN lobbies l ON a.game_id = l.game_id AND a.user_id = l.user_id
		WHERE l.id = $1
		ORDER BY a.created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, lobbyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get lobby attempts: %w", err)
	}
	defer rows.Close()

	var attempts []*models.Attempt
	for rows.Next() {
		var attempt models.Attempt

		err := rows.Scan(
			&attempt.ID,
			&attempt.GameID,
			&attempt.UserID,
			&attempt.Word,
			&attempt.Result,
			&attempt.CreatedAt,
			&attempt.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan attempt: %w", err)
		}

		attempts = append(attempts, &attempt)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating lobby attempts: %w", err)
	}

	return attempts, nil
}

// Update обновляет попытку в базе данных
func (r *AttemptRepository) Update(ctx context.Context, attempt *models.Attempt) error {
	query := `
		UPDATE attempts
		SET game_id = $1, user_id = $2, word = $3, result = $4, updated_at = $5
		WHERE id = $6
	`

	attempt.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(
		ctx,
		query,
		attempt.GameID,
		attempt.UserID,
		attempt.Word,
		attempt.Result,
		attempt.UpdatedAt,
		attempt.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update attempt: %w", err)
	}

	return nil
}

// Delete удаляет попытку из базы данных
func (r *AttemptRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM attempts WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete attempt: %w", err)
	}

	return nil
}
