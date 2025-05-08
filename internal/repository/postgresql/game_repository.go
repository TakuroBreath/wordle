package postgresql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/TakuroBreath/wordle/internal/logger"
	"github.com/TakuroBreath/wordle/internal/models"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// GameRepository представляет собой реализацию репозитория для работы с играми
type GameRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewGameRepository создает новый экземпляр GameRepository
func NewGameRepository(db *sql.DB) *GameRepository {
	return &GameRepository{
		db:     db,
		logger: logger.GetLogger(zap.String("repository", "game")),
	}
}

// Create создает новую игру в базе данных
func (r *GameRepository) Create(ctx context.Context, game *models.Game) error {
	log := r.logger.With(zap.String("method", "Create"))
	log.Info("Creating new game",
		zap.String("creator_id", fmt.Sprintf("%d", game.CreatorID)),
		zap.String("word", game.Word),
		zap.String("title", game.Title))

	query := `
		INSERT INTO games (id, creator_id, word, length, difficulty, max_tries, title, description, min_bet, max_bet, reward_multiplier, currency, reward_pool_ton, reward_pool_usdt, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
	`

	// Генерация UUID, если он не был установлен
	if game.ID == uuid.Nil {
		game.ID = uuid.New()
		log.Debug("Generated new UUID for game", zap.String("game_id", game.ID.String()))
	}

	// Установка текущего времени, если оно не было установлено
	now := time.Now()
	if game.CreatedAt.IsZero() {
		game.CreatedAt = now
	}
	game.UpdatedAt = now

	log.Debug("Executing SQL query",
		zap.String("query", query),
		zap.String("game_id", game.ID.String()),
		zap.String("word", game.Word),
		zap.Int("length", game.Length),
		zap.String("difficulty", game.Difficulty),
		zap.String("status", game.Status))

	_, err := r.db.ExecContext(
		ctx,
		query,
		game.ID,
		game.CreatorID,
		game.Word,
		game.Length,
		game.Difficulty,
		game.MaxTries,
		game.Title,
		game.Description,
		game.MinBet,
		game.MaxBet,
		game.RewardMultiplier,
		game.Currency,
		game.RewardPoolTon,
		game.RewardPoolUsdt,
		game.Status,
		game.CreatedAt,
		game.UpdatedAt,
	)

	if err != nil {
		log.Error("Failed to create game", zap.Error(err), zap.String("game_id", game.ID.String()))
		return fmt.Errorf("failed to create game: %w", err)
	}

	log.Info("Game created successfully", zap.String("game_id", game.ID.String()))
	return nil
}

// GetByID получает игру по ID
func (r *GameRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Game, error) {
	log := r.logger.With(zap.String("method", "GetByID"), zap.String("game_id", id.String()))
	log.Info("Getting game by ID")

	query := `
		SELECT id, creator_id, word, length, difficulty, max_tries, title, description,
			min_bet, max_bet, reward_multiplier, currency, reward_pool_ton, reward_pool_usdt,
			status, created_at, updated_at
		FROM games
		WHERE id = $1
	`

	log.Debug("Executing SQL query", zap.String("query", query))

	var game models.Game

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&game.ID,
		&game.CreatorID,
		&game.Word,
		&game.Length,
		&game.Difficulty,
		&game.MaxTries,
		&game.Title,
		&game.Description,
		&game.MinBet,
		&game.MaxBet,
		&game.RewardMultiplier,
		&game.Currency,
		&game.RewardPoolTon,
		&game.RewardPoolUsdt,
		&game.Status,
		&game.CreatedAt,
		&game.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			log.Warn("Game not found", zap.String("game_id", id.String()))
			return nil, models.ErrGameNotFound
		}
		log.Error("Failed to get game", zap.Error(err), zap.String("game_id", id.String()))
		return nil, fmt.Errorf("failed to get game: %w", err)
	}

	log.Info("Game found",
		zap.String("game_id", game.ID.String()),
		zap.String("title", game.Title),
		zap.String("status", game.Status))
	return &game, nil
}

// GetAll получает все игры с пагинацией
func (r *GameRepository) GetAll(ctx context.Context, limit, offset int) ([]*models.Game, error) {
	log := r.logger.With(zap.String("method", "GetAll"))
	log.Info("Getting all games", zap.Int("limit", limit), zap.Int("offset", offset))

	query := `
		SELECT id, creator_id, word, length, difficulty, max_tries, title, description,
			min_bet, max_bet, reward_multiplier, currency, reward_pool_ton, reward_pool_usdt,
			status, created_at, updated_at
		FROM games
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	log.Debug("Executing SQL query", zap.String("query", query))

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		log.Error("Failed to get games", zap.Error(err))
		return nil, fmt.Errorf("failed to get games: %w", err)
	}
	defer rows.Close()

	var games []*models.Game
	for rows.Next() {
		var game models.Game

		err := rows.Scan(
			&game.ID,
			&game.CreatorID,
			&game.Word,
			&game.Length,
			&game.Difficulty,
			&game.MaxTries,
			&game.Title,
			&game.Description,
			&game.MinBet,
			&game.MaxBet,
			&game.RewardMultiplier,
			&game.Currency,
			&game.RewardPoolTon,
			&game.RewardPoolUsdt,
			&game.Status,
			&game.CreatedAt,
			&game.UpdatedAt,
		)

		if err != nil {
			log.Error("Failed to scan game", zap.Error(err))
			return nil, fmt.Errorf("failed to scan game: %w", err)
		}

		games = append(games, &game)
	}

	if err = rows.Err(); err != nil {
		log.Error("Error iterating games", zap.Error(err))
		return nil, fmt.Errorf("error iterating games: %w", err)
	}

	log.Info("Retrieved games", zap.Int("count", len(games)))
	return games, nil
}

// GetActive получает все активные игры с пагинацией
func (r *GameRepository) GetActive(ctx context.Context, limit, offset int) ([]*models.Game, error) {
	log := r.logger.With(zap.String("method", "GetActive"))
	log.Info("Getting active games", zap.Int("limit", limit), zap.Int("offset", offset))

	query := `
		SELECT id, creator_id, word, length, difficulty, max_tries, title, description,
			min_bet, max_bet, reward_multiplier, currency, reward_pool_ton, reward_pool_usdt,
			status, created_at, updated_at
		FROM games
		WHERE status = 'active'
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	log.Debug("Executing SQL query", zap.String("query", query))

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		log.Error("Failed to get active games", zap.Error(err))
		return nil, fmt.Errorf("failed to get active games: %w", err)
	}
	defer rows.Close()

	var games []*models.Game
	for rows.Next() {
		var game models.Game

		err := rows.Scan(
			&game.ID,
			&game.CreatorID,
			&game.Word,
			&game.Length,
			&game.Difficulty,
			&game.MaxTries,
			&game.Title,
			&game.Description,
			&game.MinBet,
			&game.MaxBet,
			&game.RewardMultiplier,
			&game.Currency,
			&game.RewardPoolTon,
			&game.RewardPoolUsdt,
			&game.Status,
			&game.CreatedAt,
			&game.UpdatedAt,
		)

		if err != nil {
			log.Error("Failed to scan game", zap.Error(err))
			return nil, fmt.Errorf("failed to scan game: %w", err)
		}

		games = append(games, &game)
	}

	if err = rows.Err(); err != nil {
		log.Error("Error iterating active games", zap.Error(err))
		return nil, fmt.Errorf("error iterating active games: %w", err)
	}

	log.Info("Retrieved active games", zap.Int("count", len(games)))
	return games, nil
}

// GetByCreator получает игры по ID создателя с пагинацией
func (r *GameRepository) GetByCreator(ctx context.Context, creatorID uint64, limit, offset int) ([]*models.Game, error) {
	query := `
		SELECT id, creator_id, word, length, difficulty, max_tries, title, description,
			min_bet, max_bet, reward_multiplier, currency, reward_pool_ton, reward_pool_usdt,
			status, created_at, updated_at
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
			&game.CreatorID,
			&game.Word,
			&game.Length,
			&game.Difficulty,
			&game.MaxTries,
			&game.Title,
			&game.Description,
			&game.MinBet,
			&game.MaxBet,
			&game.RewardMultiplier,
			&game.Currency,
			&game.RewardPoolTon,
			&game.RewardPoolUsdt,
			&game.Status,
			&game.CreatedAt,
			&game.UpdatedAt,
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
	log := r.logger.With(zap.String("method", "Update"), zap.String("game_id", game.ID.String()))
	log.Info("Updating game",
		zap.String("title", game.Title),
		zap.String("status", game.Status))

	query := `
		UPDATE games
		SET creator_id = $1, word = $2, length = $3, difficulty = $4, max_tries = $5,
			title = $6, description = $7, min_bet = $8, max_bet = $9, reward_multiplier = $10,
			currency = $11, reward_pool_ton = $12, reward_pool_usdt = $13, status = $14, updated_at = $15
		WHERE id = $16
	`

	game.UpdatedAt = time.Now()

	log.Debug("Executing SQL query", zap.String("query", query))

	_, err := r.db.ExecContext(
		ctx,
		query,
		game.CreatorID,
		game.Word,
		game.Length,
		game.Difficulty,
		game.MaxTries,
		game.Title,
		game.Description,
		game.MinBet,
		game.MaxBet,
		game.RewardMultiplier,
		game.Currency,
		game.RewardPoolTon,
		game.RewardPoolUsdt,
		game.Status,
		game.UpdatedAt,
		game.ID,
	)

	if err != nil {
		log.Error("Failed to update game", zap.Error(err))
		return fmt.Errorf("failed to update game: %w", err)
	}

	log.Info("Game updated successfully", zap.String("game_id", game.ID.String()))
	return nil
}

// Delete удаляет игру из базы данных
func (r *GameRepository) Delete(ctx context.Context, id uuid.UUID) error {
	log := r.logger.With(zap.String("method", "Delete"), zap.String("game_id", id.String()))
	log.Info("Deleting game")

	query := `DELETE FROM games WHERE id = $1`

	log.Debug("Executing SQL query", zap.String("query", query))

	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		log.Error("Failed to delete game", zap.Error(err))
		return fmt.Errorf("failed to delete game: %w", err)
	}

	log.Info("Game deleted successfully", zap.String("game_id", id.String()))
	return nil
}

// GetActiveByCreator получает активные игры создателя
func (r *GameRepository) GetActiveByCreator(ctx context.Context, creatorID uint64) ([]*models.Game, error) {
	query := `
		SELECT id, creator_id, word, length, difficulty, max_tries, title, description,
			min_bet, max_bet, reward_multiplier, currency, reward_pool_ton, reward_pool_usdt,
			status, created_at, updated_at
		FROM games
		WHERE creator_id = $1 AND status = 'active'
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, creatorID)
	if err != nil {
		return nil, fmt.Errorf("failed to get active games by creator: %w", err)
	}
	defer rows.Close()

	var games []*models.Game
	for rows.Next() {
		var game models.Game
		err := rows.Scan(
			&game.ID,
			&game.CreatorID,
			&game.Word,
			&game.Length,
			&game.Difficulty,
			&game.MaxTries,
			&game.Title,
			&game.Description,
			&game.MinBet,
			&game.MaxBet,
			&game.RewardMultiplier,
			&game.Currency,
			&game.RewardPoolTon,
			&game.RewardPoolUsdt,
			&game.Status,
			&game.CreatedAt,
			&game.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan game: %w", err)
		}
		games = append(games, &game)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating active games by creator: %w", err)
	}

	return games, nil
}

// UpdateStatus обновляет статус игры
func (r *GameRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	log := r.logger.With(
		zap.String("method", "UpdateStatus"),
		zap.String("game_id", id.String()),
		zap.String("status", status))
	log.Info("Updating game status")

	query := `UPDATE games SET status = $1, updated_at = $2 WHERE id = $3`

	log.Debug("Executing SQL query", zap.String("query", query))

	_, err := r.db.ExecContext(ctx, query, status, time.Now(), id)
	if err != nil {
		log.Error("Failed to update game status", zap.Error(err))
		return fmt.Errorf("failed to update game status: %w", err)
	}

	log.Info("Game status updated successfully")
	return nil
}

// UpdateRewardPool обновляет пул наград игры
func (r *GameRepository) UpdateRewardPool(ctx context.Context, id uuid.UUID, rewardPoolTon, rewardPoolUsdt float64) error {
	log := r.logger.With(
		zap.String("method", "UpdateRewardPool"),
		zap.String("game_id", id.String()),
		zap.Float64("reward_pool_ton", rewardPoolTon),
		zap.Float64("reward_pool_usdt", rewardPoolUsdt))
	log.Info("Updating game reward pool")

	query := `UPDATE games SET reward_pool_ton = $1, reward_pool_usdt = $2, updated_at = $3 WHERE id = $4`

	log.Debug("Executing SQL query", zap.String("query", query))

	_, err := r.db.ExecContext(ctx, query, rewardPoolTon, rewardPoolUsdt, time.Now(), id)
	if err != nil {
		log.Error("Failed to update game reward pool", zap.Error(err))
		return fmt.Errorf("failed to update game reward pool: %w", err)
	}

	log.Info("Game reward pool updated successfully")
	return nil
}

// CountActive возвращает количество активных игр
func (r *GameRepository) CountActive(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM games WHERE status = 'active'`

	var count int
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count active games: %w", err)
	}

	return count, nil
}

// CountByCreator возвращает количество игр создателя
func (r *GameRepository) CountByCreator(ctx context.Context, creatorID uint64) (int, error) {
	query := `SELECT COUNT(*) FROM games WHERE creator_id = $1`

	var count int
	err := r.db.QueryRowContext(ctx, query, creatorID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count games by creator: %w", err)
	}

	return count, nil
}

// GetByDifficulty получает игры по сложности
func (r *GameRepository) GetByDifficulty(ctx context.Context, difficulty string, limit, offset int) ([]*models.Game, error) {
	query := `
		SELECT id, creator_id, word, length, difficulty, max_tries, title, description,
			min_bet, max_bet, reward_multiplier, currency, reward_pool_ton, reward_pool_usdt,
			status, created_at, updated_at
		FROM games
		WHERE difficulty = $1 AND status = 'active'
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, difficulty, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get games by difficulty: %w", err)
	}
	defer rows.Close()

	var games []*models.Game
	for rows.Next() {
		var game models.Game
		err := rows.Scan(
			&game.ID,
			&game.CreatorID,
			&game.Word,
			&game.Length,
			&game.Difficulty,
			&game.MaxTries,
			&game.Title,
			&game.Description,
			&game.MinBet,
			&game.MaxBet,
			&game.RewardMultiplier,
			&game.Currency,
			&game.RewardPoolTon,
			&game.RewardPoolUsdt,
			&game.Status,
			&game.CreatedAt,
			&game.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan game: %w", err)
		}
		games = append(games, &game)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating games by difficulty: %w", err)
	}

	return games, nil
}

// GetByUserID получает игры по ID пользователя с пагинацией
func (r *GameRepository) GetByUserID(ctx context.Context, userID uint64, limit, offset int) ([]*models.Game, error) {
	query := `
		SELECT g.id, g.creator_id, g.word, g.length, g.difficulty, g.max_tries, g.title, g.description,
			g.min_bet, g.max_bet, g.reward_multiplier, g.currency, g.reward_pool_ton, g.reward_pool_usdt,
			g.status, g.created_at, g.updated_at
		FROM games g
		JOIN lobbies l ON g.id = l.game_id
		WHERE l.user_id = $1
		ORDER BY g.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get games by user: %w", err)
	}
	defer rows.Close()

	var games []*models.Game
	for rows.Next() {
		var game models.Game

		err := rows.Scan(
			&game.ID,
			&game.CreatorID,
			&game.Word,
			&game.Length,
			&game.Difficulty,
			&game.MaxTries,
			&game.Title,
			&game.Description,
			&game.MinBet,
			&game.MaxBet,
			&game.RewardMultiplier,
			&game.Currency,
			&game.RewardPoolTon,
			&game.RewardPoolUsdt,
			&game.Status,
			&game.CreatedAt,
			&game.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan game: %w", err)
		}

		games = append(games, &game)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating games by user: %w", err)
	}

	return games, nil
}

// GetByStatus получает игры по статусу с пагинацией
func (r *GameRepository) GetByStatus(ctx context.Context, status string, limit, offset int) ([]*models.Game, error) {
	query := `
		SELECT id, creator_id, word, length, difficulty, max_tries, title, description,
			min_bet, max_bet, reward_multiplier, currency, reward_pool_ton, reward_pool_usdt,
			status, created_at, updated_at
		FROM games
		WHERE status = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, status, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get games by status: %w", err)
	}
	defer rows.Close()

	var games []*models.Game
	for rows.Next() {
		var game models.Game

		err := rows.Scan(
			&game.ID,
			&game.CreatorID,
			&game.Word,
			&game.Length,
			&game.Difficulty,
			&game.MaxTries,
			&game.Title,
			&game.Description,
			&game.MinBet,
			&game.MaxBet,
			&game.RewardMultiplier,
			&game.Currency,
			&game.RewardPoolTon,
			&game.RewardPoolUsdt,
			&game.Status,
			&game.CreatedAt,
			&game.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan game: %w", err)
		}

		games = append(games, &game)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating games by status: %w", err)
	}

	return games, nil
}

// CountByUser возвращает количество игр пользователя
func (r *GameRepository) CountByUser(ctx context.Context, userID uint64) (int, error) {
	query := `
		SELECT COUNT(DISTINCT g.id)
		FROM games g
		JOIN lobbies l ON g.id = l.game_id
		WHERE l.user_id = $1
	`

	var count int
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count user games: %w", err)
	}

	return count, nil
}

// GetGameStats получает статистику по игре
func (r *GameRepository) GetGameStats(ctx context.Context, gameID uuid.UUID) (map[string]interface{}, error) {
	query := `
		SELECT 
			COUNT(DISTINCT l.id) as total_lobbies,
			COUNT(DISTINCT l.user_id) as unique_players,
			SUM(l.bet_amount) as total_bets,
			SUM(CASE WHEN h.status = 'win' THEN h.reward ELSE 0 END) as total_rewards,
			AVG(CASE WHEN h.status = 'win' THEN h.reward ELSE NULL END) as avg_reward
		FROM lobbies l
		LEFT JOIN history h ON l.id = h.lobby_id
		WHERE l.game_id = $1
	`

	var stats struct {
		TotalLobbies  int     `db:"total_lobbies"`
		UniquePlayers int     `db:"unique_players"`
		TotalBets     float64 `db:"total_bets"`
		TotalRewards  float64 `db:"total_rewards"`
		AverageReward float64 `db:"avg_reward"`
	}

	err := r.db.QueryRowContext(ctx, query, gameID).Scan(
		&stats.TotalLobbies,
		&stats.UniquePlayers,
		&stats.TotalBets,
		&stats.TotalRewards,
		&stats.AverageReward,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get game stats: %w", err)
	}

	return map[string]interface{}{
		"total_lobbies":  stats.TotalLobbies,
		"unique_players": stats.UniquePlayers,
		"total_bets":     stats.TotalBets,
		"total_rewards":  stats.TotalRewards,
		"average_reward": stats.AverageReward,
	}, nil
}

// SearchGames ищет игры по параметрам
func (r *GameRepository) SearchGames(ctx context.Context, minBet, maxBet float64, difficulty string, limit, offset int) ([]*models.Game, error) {
	query := `
		SELECT id, creator_id, word, length, difficulty, max_tries, title, description,
			   min_bet, max_bet, reward_multiplier, currency, reward_pool_ton, reward_pool_usdt,
			   status, created_at, updated_at
		FROM games
		WHERE status = 'active'
		AND ($1 = 0 OR min_bet >= $1)
		AND ($2 = 0 OR max_bet <= $2)
		AND ($3 = '' OR difficulty = $3)
		ORDER BY created_at DESC
		LIMIT $4 OFFSET $5
	`

	rows, err := r.db.QueryContext(ctx, query, minBet, maxBet, difficulty, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search games: %w", err)
	}
	defer rows.Close()

	var games []*models.Game
	for rows.Next() {
		var game models.Game
		err := rows.Scan(
			&game.ID,
			&game.CreatorID,
			&game.Word,
			&game.Length,
			&game.Difficulty,
			&game.MaxTries,
			&game.Title,
			&game.Description,
			&game.MinBet,
			&game.MaxBet,
			&game.RewardMultiplier,
			&game.Currency,
			&game.RewardPoolTon,
			&game.RewardPoolUsdt,
			&game.Status,
			&game.CreatedAt,
			&game.UpdatedAt,
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
