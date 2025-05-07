package postgresql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/TakuroBreath/wordle/internal/models"
	"github.com/google/uuid"
)

// HistoryRepository представляет собой реализацию репозитория для работы с историей
type HistoryRepository struct {
	db *sql.DB
}

// NewHistoryRepository создает новый экземпляр HistoryRepository
func NewHistoryRepository(db *sql.DB) *HistoryRepository {
	return &HistoryRepository{
		db: db,
	}
}

// Create создает новую запись в истории
func (r *HistoryRepository) Create(ctx context.Context, history *models.History) error {
	query := `
		INSERT INTO history (id, game_id, user_id, result, reward, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	// Генерация UUID, если он не был установлен
	if history.ID == uuid.Nil {
		history.ID = uuid.New()
	}

	// Установка текущего времени
	now := time.Now()
	history.CreatedAt = now
	history.UpdatedAt = now

	_, err := r.db.ExecContext(
		ctx,
		query,
		history.ID,
		history.GameID,
		history.UserID,
		history.Result,
		history.Reward,
		history.CreatedAt,
		history.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create history: %w", err)
	}

	return nil
}

// GetByID получает запись истории по ID
func (r *HistoryRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.History, error) {
	query := `
		SELECT id, game_id, user_id, result, reward, created_at, updated_at
		FROM history
		WHERE id = $1
	`

	var history models.History

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&history.ID,
		&history.GameID,
		&history.UserID,
		&history.Result,
		&history.Reward,
		&history.CreatedAt,
		&history.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("history not found")
		}
		return nil, fmt.Errorf("failed to get history: %w", err)
	}

	return &history, nil
}

// GetByGameID получает все записи истории для конкретной игры с пагинацией
func (r *HistoryRepository) GetByGameID(ctx context.Context, gameID uuid.UUID, limit, offset int) ([]*models.History, error) {
	query := `
		SELECT id, game_id, user_id, result, reward, created_at, updated_at
		FROM history
		WHERE game_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, gameID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get game history: %w", err)
	}
	defer rows.Close()

	var histories []*models.History
	for rows.Next() {
		var history models.History

		err := rows.Scan(
			&history.ID,
			&history.GameID,
			&history.UserID,
			&history.Result,
			&history.Reward,
			&history.CreatedAt,
			&history.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan history: %w", err)
		}

		histories = append(histories, &history)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating game history: %w", err)
	}

	return histories, nil
}

// GetByUserID получает все записи истории пользователя
func (r *HistoryRepository) GetByUserID(ctx context.Context, userID uint64) ([]*models.History, error) {
	query := `
		SELECT id, game_id, user_id, result, reward, created_at, updated_at
		FROM history
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user history: %w", err)
	}
	defer rows.Close()

	var histories []*models.History
	for rows.Next() {
		var history models.History

		err := rows.Scan(
			&history.ID,
			&history.GameID,
			&history.UserID,
			&history.Result,
			&history.Reward,
			&history.CreatedAt,
			&history.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan history: %w", err)
		}

		histories = append(histories, &history)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating user history: %w", err)
	}

	return histories, nil
}

// Update обновляет запись в истории
func (r *HistoryRepository) Update(ctx context.Context, history *models.History) error {
	query := `
		UPDATE history
		SET game_id = $1, user_id = $2, result = $3, reward = $4, updated_at = $5
		WHERE id = $6
	`

	history.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(
		ctx,
		query,
		history.GameID,
		history.UserID,
		history.Result,
		history.Reward,
		history.UpdatedAt,
		history.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update history: %w", err)
	}

	return nil
}

// Delete удаляет запись из истории
func (r *HistoryRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM history WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete history: %w", err)
	}

	return nil
}
