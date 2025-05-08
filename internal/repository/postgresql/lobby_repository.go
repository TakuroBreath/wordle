package postgresql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/TakuroBreath/wordle/internal/models"
	"github.com/google/uuid"
)

// Определяем ошибку для случая, когда лобби не найдено
var ErrLobbyNotFound = errors.New("lobby not found")

// LobbyRepository представляет собой реализацию репозитория для работы с лобби
type LobbyRepository struct {
	db *sql.DB
}

// NewLobbyRepository создает новый экземпляр LobbyRepository
func NewLobbyRepository(db *sql.DB) *LobbyRepository {
	return &LobbyRepository{
		db: db,
	}
}

// Create создает новое лобби в базе данных
func (r *LobbyRepository) Create(ctx context.Context, lobby *models.Lobby) error {
	query := `
		INSERT INTO lobbies (
			id, game_id, user_id, max_tries, tries_used, bet_amount, 
			potential_reward, status, created_at, updated_at, expires_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	// Генерация UUID, если он не был установлен
	if lobby.ID == uuid.Nil {
		lobby.ID = uuid.New()
	}

	// Установка текущего времени
	now := time.Now()
	if lobby.CreatedAt.IsZero() {
		lobby.CreatedAt = now
	}
	lobby.UpdatedAt = now

	// Установка времени истечения (5 минут после создания)
	if lobby.ExpiresAt.IsZero() {
		lobby.ExpiresAt = now.Add(5 * time.Minute)
	}

	_, err := r.db.ExecContext(
		ctx,
		query,
		lobby.ID,
		lobby.GameID,
		lobby.UserID,
		lobby.MaxTries,
		lobby.TriesUsed,
		lobby.BetAmount,
		lobby.PotentialReward,
		lobby.Status,
		lobby.CreatedAt,
		lobby.UpdatedAt,
		lobby.ExpiresAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create lobby: %w", err)
	}

	return nil
}

// GetByID получает лобби по ID
func (r *LobbyRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Lobby, error) {
	query := `
		SELECT id, game_id, user_id, max_tries, tries_used, bet_amount,
			potential_reward, status, created_at, updated_at, expires_at
		FROM lobbies
		WHERE id = $1
	`

	var lobby models.Lobby

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&lobby.ID,
		&lobby.GameID,
		&lobby.UserID,
		&lobby.MaxTries,
		&lobby.TriesUsed,
		&lobby.BetAmount,
		&lobby.PotentialReward,
		&lobby.Status,
		&lobby.CreatedAt,
		&lobby.UpdatedAt,
		&lobby.ExpiresAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("lobby not found")
		}
		return nil, fmt.Errorf("failed to get lobby: %w", err)
	}

	return &lobby, nil
}

// GetByGameID получает все лобби для конкретной игры
func (r *LobbyRepository) GetByGameID(ctx context.Context, gameID uuid.UUID, limit, offset int) ([]*models.Lobby, error) {
	query := `
		SELECT id, game_id, user_id, max_tries, tries_used, bet_amount,
			potential_reward, status, created_at, updated_at, expires_at
		FROM lobbies
		WHERE game_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, gameID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get lobbies: %w", err)
	}
	defer rows.Close()

	var lobbies []*models.Lobby
	for rows.Next() {
		var lobby models.Lobby

		err := rows.Scan(
			&lobby.ID,
			&lobby.GameID,
			&lobby.UserID,
			&lobby.MaxTries,
			&lobby.TriesUsed,
			&lobby.BetAmount,
			&lobby.PotentialReward,
			&lobby.Status,
			&lobby.CreatedAt,
			&lobby.UpdatedAt,
			&lobby.ExpiresAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan lobby: %w", err)
		}

		lobbies = append(lobbies, &lobby)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating lobbies: %w", err)
	}

	return lobbies, nil
}

// GetByUserID получает все лобби пользователя
func (r *LobbyRepository) GetByUserID(ctx context.Context, userID uint64, limit, offset int) ([]*models.Lobby, error) {
	query := `
		SELECT id, game_id, user_id, max_tries, tries_used, bet_amount,
			potential_reward, status, created_at, updated_at, expires_at
		FROM lobbies
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get user lobbies: %w", err)
	}
	defer rows.Close()

	var lobbies []*models.Lobby
	for rows.Next() {
		var lobby models.Lobby

		err := rows.Scan(
			&lobby.ID,
			&lobby.GameID,
			&lobby.UserID,
			&lobby.MaxTries,
			&lobby.TriesUsed,
			&lobby.BetAmount,
			&lobby.PotentialReward,
			&lobby.Status,
			&lobby.CreatedAt,
			&lobby.UpdatedAt,
			&lobby.ExpiresAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan lobby: %w", err)
		}

		lobbies = append(lobbies, &lobby)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating user lobbies: %w", err)
	}

	return lobbies, nil
}

// GetActive получает все активные лобби
func (r *LobbyRepository) GetActive(ctx context.Context, limit, offset int) ([]*models.Lobby, error) {
	query := `
		SELECT id, game_id, user_id, max_tries, tries_used, bet_amount,
			potential_reward, status, created_at, updated_at, expires_at
		FROM lobbies
		WHERE status = 'active' AND expires_at > NOW()
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get active lobbies: %w", err)
	}
	defer rows.Close()

	var lobbies []*models.Lobby
	for rows.Next() {
		var lobby models.Lobby

		err := rows.Scan(
			&lobby.ID,
			&lobby.GameID,
			&lobby.UserID,
			&lobby.MaxTries,
			&lobby.TriesUsed,
			&lobby.BetAmount,
			&lobby.PotentialReward,
			&lobby.Status,
			&lobby.CreatedAt,
			&lobby.UpdatedAt,
			&lobby.ExpiresAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan lobby: %w", err)
		}

		lobbies = append(lobbies, &lobby)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating active lobbies: %w", err)
	}

	return lobbies, nil
}

// Update обновляет лобби в базе данных
func (r *LobbyRepository) Update(ctx context.Context, lobby *models.Lobby) error {
	query := `
		UPDATE lobbies
		SET game_id = $1, user_id = $2, max_tries = $3, tries_used = $4,
			bet_amount = $5, potential_reward = $6, status = $7,
			updated_at = $8, expires_at = $9
		WHERE id = $10
	`

	lobby.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(
		ctx,
		query,
		lobby.GameID,
		lobby.UserID,
		lobby.MaxTries,
		lobby.TriesUsed,
		lobby.BetAmount,
		lobby.PotentialReward,
		lobby.Status,
		lobby.UpdatedAt,
		lobby.ExpiresAt,
		lobby.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update lobby: %w", err)
	}

	return nil
}

// UpdateStatus обновляет только статус лобби
func (r *LobbyRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	query := `
		UPDATE lobbies
		SET status = $1, updated_at = $2
		WHERE id = $3
	`

	_, err := r.db.ExecContext(ctx, query, status, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update lobby status: %w", err)
	}

	return nil
}

// UpdateTriesUsed обновляет количество использованных попыток
func (r *LobbyRepository) UpdateTriesUsed(ctx context.Context, id uuid.UUID, triesUsed int) error {
	query := `
		UPDATE lobbies
		SET tries_used = $1, updated_at = $2
		WHERE id = $3
	`

	_, err := r.db.ExecContext(ctx, query, triesUsed, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update tries used: %w", err)
	}

	return nil
}

// GetExpired получает истекшие лобби
func (r *LobbyRepository) GetExpired(ctx context.Context) ([]*models.Lobby, error) {
	query := `
		SELECT id, game_id, user_id, max_tries, tries_used, bet_amount,
			potential_reward, status, created_at, updated_at, expires_at
		FROM lobbies
		WHERE status = 'active' AND expires_at <= NOW()
		ORDER BY expires_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get expired lobbies: %w", err)
	}
	defer rows.Close()

	var lobbies []*models.Lobby
	for rows.Next() {
		var lobby models.Lobby

		err := rows.Scan(
			&lobby.ID,
			&lobby.GameID,
			&lobby.UserID,
			&lobby.MaxTries,
			&lobby.TriesUsed,
			&lobby.BetAmount,
			&lobby.PotentialReward,
			&lobby.Status,
			&lobby.CreatedAt,
			&lobby.UpdatedAt,
			&lobby.ExpiresAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan lobby: %w", err)
		}

		lobbies = append(lobbies, &lobby)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating expired lobbies: %w", err)
	}

	return lobbies, nil
}

// Delete удаляет лобби из базы данных
func (r *LobbyRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM lobbies WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete lobby: %w", err)
	}

	return nil
}

// GetActiveByGameAndUser получает активное лобби пользователя в игре
func (r *LobbyRepository) GetActiveByGameAndUser(ctx context.Context, gameID uuid.UUID, userID uint64) (*models.Lobby, error) {
	query := `
		SELECT id, game_id, user_id, max_tries, tries_used, bet_amount,
			potential_reward, status, created_at, updated_at, expires_at
		FROM lobbies
		WHERE game_id = $1 AND user_id = $2 AND status = 'active' AND expires_at > NOW()
		ORDER BY created_at DESC
		LIMIT 1
	`

	var lobby models.Lobby

	err := r.db.QueryRowContext(ctx, query, gameID, userID).Scan(
		&lobby.ID,
		&lobby.GameID,
		&lobby.UserID,
		&lobby.MaxTries,
		&lobby.TriesUsed,
		&lobby.BetAmount,
		&lobby.PotentialReward,
		&lobby.Status,
		&lobby.CreatedAt,
		&lobby.UpdatedAt,
		&lobby.ExpiresAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrLobbyNotFound
		}
		return nil, fmt.Errorf("failed to get active lobby: %w", err)
	}

	return &lobby, nil
}

// CountActiveByGame возвращает количество активных лобби в игре
func (r *LobbyRepository) CountActiveByGame(ctx context.Context, gameID uuid.UUID) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM lobbies
		WHERE game_id = $1 AND status = 'active' AND expires_at > NOW()
	`

	var count int
	err := r.db.QueryRowContext(ctx, query, gameID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count active lobbies: %w", err)
	}

	return count, nil
}

// CountByUser возвращает количество лобби пользователя
func (r *LobbyRepository) CountByUser(ctx context.Context, userID uint64) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM lobbies
		WHERE user_id = $1
	`

	var count int
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count user lobbies: %w", err)
	}

	return count, nil
}

// GetActiveWithAttempts получает активные лобби с попытками
func (r *LobbyRepository) GetActiveWithAttempts(ctx context.Context, limit, offset int) ([]*models.Lobby, error) {
	query := `
		SELECT l.id, l.game_id, l.user_id, l.max_tries, l.tries_used, l.bet_amount,
			l.potential_reward, l.status, l.created_at, l.updated_at, l.expires_at,
			COUNT(a.id) as attempts_count
		FROM lobbies l
		LEFT JOIN attempts a ON l.id = a.lobby_id
		WHERE l.status = 'active' AND l.expires_at > NOW()
		GROUP BY l.id
		ORDER BY l.created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get active lobbies with attempts: %w", err)
	}
	defer rows.Close()

	var lobbies []*models.Lobby
	for rows.Next() {
		var lobby models.Lobby
		var attemptsCount int

		err := rows.Scan(
			&lobby.ID,
			&lobby.GameID,
			&lobby.UserID,
			&lobby.MaxTries,
			&lobby.TriesUsed,
			&lobby.BetAmount,
			&lobby.PotentialReward,
			&lobby.Status,
			&lobby.CreatedAt,
			&lobby.UpdatedAt,
			&lobby.ExpiresAt,
			&attemptsCount,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan lobby: %w", err)
		}

		lobbies = append(lobbies, &lobby)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating active lobbies: %w", err)
	}

	return lobbies, nil
}

// ExtendExpirationTime продлевает время жизни лобби
func (r *LobbyRepository) ExtendExpirationTime(ctx context.Context, id uuid.UUID, duration int) error {
	query := `
		UPDATE lobbies
		SET expires_at = expires_at + ($1 * interval '1 minute'), updated_at = $2
		WHERE id = $3 AND status = 'active'
	`

	_, err := r.db.ExecContext(ctx, query, duration, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to extend lobby expiration time: %w", err)
	}

	return nil
}

// GetLobbiesByStatus получает лобби по статусу
func (r *LobbyRepository) GetLobbiesByStatus(ctx context.Context, status string, limit, offset int) ([]*models.Lobby, error) {
	query := `
		SELECT id, game_id, user_id, max_tries, tries_used, bet_amount,
			potential_reward, status, created_at, updated_at, expires_at
		FROM lobbies
		WHERE status = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, status, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get lobbies by status: %w", err)
	}
	defer rows.Close()

	var lobbies []*models.Lobby
	for rows.Next() {
		var lobby models.Lobby

		err := rows.Scan(
			&lobby.ID,
			&lobby.GameID,
			&lobby.UserID,
			&lobby.MaxTries,
			&lobby.TriesUsed,
			&lobby.BetAmount,
			&lobby.PotentialReward,
			&lobby.Status,
			&lobby.CreatedAt,
			&lobby.UpdatedAt,
			&lobby.ExpiresAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan lobby: %w", err)
		}

		lobbies = append(lobbies, &lobby)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating lobbies by status: %w", err)
	}

	return lobbies, nil
}

// GetActiveLobbiesByTimeRange получает активные лобби в указанном временном диапазоне
func (r *LobbyRepository) GetActiveLobbiesByTimeRange(ctx context.Context, startTime, endTime time.Time, limit, offset int) ([]*models.Lobby, error) {
	query := `
		SELECT l.id, l.game_id, l.user_id, l.max_tries, l.tries_used, l.bet_amount,
			   l.potential_reward, l.created_at, l.updated_at, l.expires_at, l.status
		FROM lobbies l
		WHERE l.status = 'active'
		AND l.created_at BETWEEN $1 AND $2
		ORDER BY l.created_at DESC
		LIMIT $3 OFFSET $4
	`

	rows, err := r.db.QueryContext(ctx, query, startTime, endTime, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get active lobbies by time range: %w", err)
	}
	defer rows.Close()

	var lobbies []*models.Lobby
	for rows.Next() {
		var lobby models.Lobby
		err := rows.Scan(
			&lobby.ID,
			&lobby.GameID,
			&lobby.UserID,
			&lobby.MaxTries,
			&lobby.TriesUsed,
			&lobby.BetAmount,
			&lobby.PotentialReward,
			&lobby.CreatedAt,
			&lobby.UpdatedAt,
			&lobby.ExpiresAt,
			&lobby.Status,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan lobby: %w", err)
		}
		lobbies = append(lobbies, &lobby)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating lobbies: %w", err)
	}

	return lobbies, nil
}

// GetLobbyStats получает статистику по лобби
func (r *LobbyRepository) GetLobbyStats(ctx context.Context, lobbyID uuid.UUID) (map[string]interface{}, error) {
	query := `
		SELECT 
			l.bet_amount,
			l.potential_reward,
			l.tries_used,
			l.max_tries,
			COUNT(a.id) as total_attempts,
			MAX(a.created_at) as last_attempt_time,
			EXTRACT(EPOCH FROM (l.expires_at - NOW())) as time_remaining
		FROM lobbies l
		LEFT JOIN attempts a ON l.id = a.lobby_id
		WHERE l.id = $1
		GROUP BY l.id, l.bet_amount, l.potential_reward, l.tries_used, l.max_tries
	`

	var stats struct {
		BetAmount       float64   `db:"bet_amount"`
		PotentialReward float64   `db:"potential_reward"`
		TriesUsed       int       `db:"tries_used"`
		MaxTries        int       `db:"max_tries"`
		TotalAttempts   int       `db:"total_attempts"`
		LastAttemptTime time.Time `db:"last_attempt_time"`
		TimeRemaining   float64   `db:"time_remaining"`
	}

	err := r.db.QueryRowContext(ctx, query, lobbyID).Scan(
		&stats.BetAmount,
		&stats.PotentialReward,
		&stats.TriesUsed,
		&stats.MaxTries,
		&stats.TotalAttempts,
		&stats.LastAttemptTime,
		&stats.TimeRemaining,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get lobby stats: %w", err)
	}

	return map[string]interface{}{
		"bet_amount":        stats.BetAmount,
		"potential_reward":  stats.PotentialReward,
		"tries_used":        stats.TriesUsed,
		"max_tries":         stats.MaxTries,
		"total_attempts":    stats.TotalAttempts,
		"last_attempt_time": stats.LastAttemptTime,
		"time_remaining":    stats.TimeRemaining,
	}, nil
}
