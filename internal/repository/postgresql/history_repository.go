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
		INSERT INTO history (id, game_id, user_id, lobby_id, status, reward, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
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
		history.LobbyID,
		history.Status,
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
		SELECT id, game_id, user_id, lobby_id, status, reward, created_at, updated_at
		FROM history
		WHERE id = $1
	`

	var history models.History

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&history.ID,
		&history.GameID,
		&history.UserID,
		&history.LobbyID,
		&history.Status,
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
		SELECT id, game_id, user_id, lobby_id, status, reward, created_at, updated_at
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
			&history.LobbyID,
			&history.Status,
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

// GetByUserID получает все записи истории пользователя с пагинацией
func (r *HistoryRepository) GetByUserID(ctx context.Context, userID uint64, limit, offset int) ([]*models.History, error) {
	query := `
		SELECT id, game_id, user_id, lobby_id, status, reward, created_at, updated_at
		FROM history
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
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
			&history.LobbyID,
			&history.Status,
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
		SET game_id = $1, user_id = $2, lobby_id = $3, status = $4, reward = $5, updated_at = $6
		WHERE id = $7
	`

	history.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(
		ctx,
		query,
		history.GameID,
		history.UserID,
		history.LobbyID,
		history.Status,
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

// GetByLobbyID получает историю по ID лобби
func (r *HistoryRepository) GetByLobbyID(ctx context.Context, lobbyID uuid.UUID) (*models.History, error) {
	query := `
		SELECT id, game_id, user_id, lobby_id, status, reward, created_at, updated_at
		FROM history
		WHERE lobby_id = $1
	`

	var history models.History

	err := r.db.QueryRowContext(ctx, query, lobbyID).Scan(
		&history.ID,
		&history.GameID,
		&history.UserID,
		&history.LobbyID,
		&history.Status,
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

// GetByStatus получает записи истории по статусу с пагинацией
func (r *HistoryRepository) GetByStatus(ctx context.Context, status string, limit, offset int) ([]*models.History, error) {
	query := `
		SELECT id, game_id, user_id, lobby_id, status, reward, created_at, updated_at
		FROM history
		WHERE status = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, status, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get history by status: %w", err)
	}
	defer rows.Close()

	var histories []*models.History
	for rows.Next() {
		var history models.History

		err := rows.Scan(
			&history.ID,
			&history.GameID,
			&history.UserID,
			&history.LobbyID,
			&history.Status,
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
		return nil, fmt.Errorf("error iterating history by status: %w", err)
	}

	return histories, nil
}

// CountByUser возвращает количество записей истории пользователя
func (r *HistoryRepository) CountByUser(ctx context.Context, userID uint64) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM history
		WHERE user_id = $1
	`

	var count int
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count user history: %w", err)
	}

	return count, nil
}

// GetUserStats получает статистику пользователя
func (r *HistoryRepository) GetUserStats(ctx context.Context, userID uint64) (map[string]interface{}, error) {
	query := `
		SELECT 
			COUNT(*) as total_games,
			COUNT(CASE WHEN status = 'win' THEN 1 END) as wins,
			COUNT(CASE WHEN status = 'lose' THEN 1 END) as losses,
			SUM(reward) as total_rewards,
			AVG(CASE WHEN status = 'win' THEN reward ELSE 0 END) as avg_win_reward
		FROM history
		WHERE user_id = $1
	`

	var stats struct {
		TotalGames   int
		Wins         int
		Losses       int
		TotalRewards float64
		AvgWinReward float64
	}

	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&stats.TotalGames,
		&stats.Wins,
		&stats.Losses,
		&stats.TotalRewards,
		&stats.AvgWinReward,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get user stats: %w", err)
	}

	return map[string]interface{}{
		"total_games":    stats.TotalGames,
		"wins":           stats.Wins,
		"losses":         stats.Losses,
		"win_rate":       float64(stats.Wins) / float64(stats.TotalGames),
		"total_rewards":  stats.TotalRewards,
		"avg_win_reward": stats.AvgWinReward,
	}, nil
}

// GetTopPlayers получает топ игроков по наградам
func (r *HistoryRepository) GetTopPlayers(ctx context.Context, limit int) ([]map[string]interface{}, error) {
	query := `
		SELECT 
			user_id,
			COUNT(*) as total_games,
			COUNT(CASE WHEN status = 'win' THEN 1 END) as wins,
			SUM(reward) as total_rewards
		FROM history
		GROUP BY user_id
		ORDER BY total_rewards DESC
		LIMIT $1
	`

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get top players: %w", err)
	}
	defer rows.Close()

	var players []map[string]interface{}
	for rows.Next() {
		var player struct {
			UserID       uint64
			TotalGames   int
			Wins         int
			TotalRewards float64
		}

		err := rows.Scan(
			&player.UserID,
			&player.TotalGames,
			&player.Wins,
			&player.TotalRewards,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan player stats: %w", err)
		}

		players = append(players, map[string]interface{}{
			"user_id":       player.UserID,
			"total_games":   player.TotalGames,
			"wins":          player.Wins,
			"win_rate":      float64(player.Wins) / float64(player.TotalGames),
			"total_rewards": player.TotalRewards,
		})
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating top players: %w", err)
	}

	return players, nil
}

// GetRecentGames получает последние игры с пагинацией
func (r *HistoryRepository) GetRecentGames(ctx context.Context, limit, offset int) ([]*models.History, error) {
	query := `
		SELECT id, game_id, user_id, lobby_id, status, reward, created_at, updated_at
		FROM history
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent games: %w", err)
	}
	defer rows.Close()

	var histories []*models.History
	for rows.Next() {
		var history models.History

		err := rows.Scan(
			&history.ID,
			&history.GameID,
			&history.UserID,
			&history.LobbyID,
			&history.Status,
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
		return nil, fmt.Errorf("error iterating recent games: %w", err)
	}

	return histories, nil
}
