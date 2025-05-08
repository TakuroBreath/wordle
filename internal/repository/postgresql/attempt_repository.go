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

// GetByLobbyID получает попытки по ID лобби с пагинацией
func (r *AttemptRepository) GetByLobbyID(ctx context.Context, lobbyID uuid.UUID, limit, offset int) ([]*models.Attempt, error) {
	query := `
		SELECT a.id, a.game_id, a.user_id, a.word, a.result, a.created_at, a.updated_at
		FROM attempts a
		JOIN lobbies l ON a.game_id = l.game_id AND a.user_id = l.user_id
		WHERE l.id = $1
		ORDER BY a.created_at ASC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, lobbyID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get attempts by lobby: %w", err)
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

// GetByWord получает попытку по слову в лобби
func (r *AttemptRepository) GetByWord(ctx context.Context, lobbyID uuid.UUID, word string) (*models.Attempt, error) {
	query := `
		SELECT a.id, a.game_id, a.user_id, a.word, a.result, a.created_at, a.updated_at
		FROM attempts a
		JOIN lobbies l ON a.game_id = l.game_id AND a.user_id = l.user_id
		WHERE l.id = $1 AND a.word = $2
	`

	var attempt models.Attempt
	var resultJSON []byte

	err := r.db.QueryRowContext(ctx, query, lobbyID, word).Scan(
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

	if err := json.Unmarshal(resultJSON, &attempt.Result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal result: %w", err)
	}

	return &attempt, nil
}

// CountByLobbyID возвращает количество попыток в лобби
func (r *AttemptRepository) CountByLobbyID(ctx context.Context, lobbyID uuid.UUID) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM attempts a
		JOIN lobbies l ON a.game_id = l.game_id AND a.user_id = l.user_id
		WHERE l.id = $1
	`

	var count int
	err := r.db.QueryRowContext(ctx, query, lobbyID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count attempts: %w", err)
	}

	return count, nil
}

// Update обновляет попытку в базе данных
func (r *AttemptRepository) Update(ctx context.Context, attempt *models.Attempt) error {
	query := `
		UPDATE attempts
		SET game_id = $1, user_id = $2, word = $3, result = $4, updated_at = $5
		WHERE id = $6
	`

	attempt.UpdatedAt = time.Now()

	resultJSON, err := json.Marshal(attempt.Result)
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}

	_, err = r.db.ExecContext(
		ctx,
		query,
		attempt.GameID,
		attempt.UserID,
		attempt.Word,
		resultJSON,
		attempt.UpdatedAt,
		attempt.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update attempt: %w", err)
	}

	return nil
}

// GetAttemptStats получает статистику по попыткам
func (r *AttemptRepository) GetAttemptStats(ctx context.Context, lobbyID uuid.UUID) (map[string]interface{}, error) {
	query := `
		SELECT 
			COUNT(*) as total_attempts,
			COUNT(DISTINCT user_id) as unique_players,
			AVG(EXTRACT(EPOCH FROM (created_at - LAG(created_at) OVER (ORDER BY created_at)))) as avg_time_between_attempts,
			MAX(created_at) as last_attempt_time
		FROM attempts
		WHERE lobby_id = $1
	`

	var stats struct {
		TotalAttempts          int       `db:"total_attempts"`
		UniquePlayers          int       `db:"unique_players"`
		AvgTimeBetweenAttempts float64   `db:"avg_time_between_attempts"`
		LastAttemptTime        time.Time `db:"last_attempt_time"`
	}

	err := r.db.QueryRowContext(ctx, query, lobbyID).Scan(
		&stats.TotalAttempts,
		&stats.UniquePlayers,
		&stats.AvgTimeBetweenAttempts,
		&stats.LastAttemptTime,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get attempt stats: %w", err)
	}

	return map[string]interface{}{
		"total_attempts":            stats.TotalAttempts,
		"unique_players":            stats.UniquePlayers,
		"avg_time_between_attempts": stats.AvgTimeBetweenAttempts,
		"last_attempt_time":         stats.LastAttemptTime,
	}, nil
}

// ValidateWord проверяет валидность слова
func (r *AttemptRepository) ValidateWord(ctx context.Context, word string) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1
			FROM attempts
			WHERE word = $1
			AND result->>'is_valid' = 'true'
			LIMIT 1
		)
	`

	var isValid bool
	err := r.db.QueryRowContext(ctx, query, word).Scan(&isValid)
	if err != nil {
		return false, fmt.Errorf("failed to validate word: %w", err)
	}

	return isValid, nil
}
