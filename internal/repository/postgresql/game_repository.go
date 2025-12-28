package postgresql

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/TakuroBreath/wordle/internal/logger"
	"github.com/TakuroBreath/wordle/internal/models"
	"github.com/TakuroBreath/wordle/internal/repository"
	otel "github.com/TakuroBreath/wordle/pkg/tracing"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
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

	return repository.WithTracingVoid(ctx, "GameRepository", "Create", func(ctx context.Context) error {
		sqlCtx, sqlSpan := otel.StartSpan(ctx, "sql.query.CreateGame")
		defer sqlSpan.End()

		otel.AddAttributesToSpan(sqlSpan,
			attribute.String("db.system", "postgresql"),
			attribute.String("db.operation", "INSERT"),
			attribute.String("db.table", "games"),
			attribute.String("game_id", game.ID.String()),
			attribute.Int64("creator_id", int64(game.CreatorID)),
		)

		query := `
			INSERT INTO games (id, short_id, creator_id, word, length, difficulty, max_tries, title, description, 
			min_bet, max_bet, reward_multiplier, deposit_amount, commission_rate, time_limit_minutes, 
			currency, reward_pool_ton, reward_pool_usdt, status, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21)
		`

		if game.ID == uuid.Nil {
			game.ID = uuid.New()
			log.Debug("Generated new UUID for game", zap.String("game_id", game.ID.String()))
			otel.AddAttributesToSpan(sqlSpan, attribute.String("game_id", game.ID.String()))
		}

		now := time.Now()
		if game.CreatedAt.IsZero() {
			game.CreatedAt = now
		}
		game.UpdatedAt = now

		log.Debug("Executing SQL query",
			zap.String("query", query),
			zap.String("game_id", game.ID.String()),
			zap.String("status", game.Status))

		_, err := r.db.ExecContext(
			sqlCtx,
			query,
			game.ID,
			game.ShortID,
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
			game.DepositAmount,
			game.CommissionRate,
			game.TimeLimitMinutes,
			game.Currency,
			game.RewardPoolTon,
			game.RewardPoolUsdt,
			game.Status,
			game.CreatedAt,
			game.UpdatedAt,
		)

		if err != nil {
			log.Error("Failed to create game", zap.Error(err), zap.String("game_id", game.ID.String()))
			otel.RecordError(sqlCtx, err)
			return fmt.Errorf("failed to create game: %w", err)
		}

		log.Info("Game created successfully", zap.String("game_id", game.ID.String()))
		return nil
	})
}

// GetByID получает игру по ID
func (r *GameRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Game, error) {
	log := r.logger.With(zap.String("method", "GetByID"), zap.String("game_id", id.String()))
	log.Info("Getting game by ID")

	result, err := repository.WithTracing(ctx, "GameRepository", "GetByID", func(ctx context.Context) (interface{}, error) {
		sqlCtx, sqlSpan := otel.StartSpan(ctx, "sql.query.GetGameByID")
		defer sqlSpan.End()

		query := `
			SELECT id, short_id, creator_id, word, length, difficulty, max_tries, title, description,
				min_bet, max_bet, reward_multiplier, deposit_amount, commission_rate, time_limit_minutes,
				currency, reward_pool_ton, reward_pool_usdt, status, created_at, updated_at
			FROM games
			WHERE id = $1
		`

		var game models.Game
		err := r.db.QueryRowContext(sqlCtx, query, id).Scan(
			&game.ID,
			&game.ShortID,
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
			&game.DepositAmount,
			&game.CommissionRate,
			&game.TimeLimitMinutes,
			&game.Currency,
			&game.RewardPoolTon,
			&game.RewardPoolUsdt,
			&game.Status,
			&game.CreatedAt,
			&game.UpdatedAt,
		)

		if err != nil {
			if err == sql.ErrNoRows {
				return nil, models.ErrGameNotFound
			}
			return nil, fmt.Errorf("failed to get game: %w", err)
		}

		return &game, nil
	})

	if err != nil {
		return nil, err
	}

	return result.(*models.Game), nil
}

// GetByShortID получает игру по короткому ID
func (r *GameRepository) GetByShortID(ctx context.Context, shortID string) (*models.Game, error) {
	log := r.logger.With(zap.String("method", "GetByShortID"), zap.String("short_id", shortID))
	log.Info("Getting game by ShortID")

	result, err := repository.WithTracing(ctx, "GameRepository", "GetByShortID", func(ctx context.Context) (interface{}, error) {
		sqlCtx, sqlSpan := otel.StartSpan(ctx, "sql.query.GetGameByShortID")
		defer sqlSpan.End()

		query := `
			SELECT id, short_id, creator_id, word, length, difficulty, max_tries, title, description,
				min_bet, max_bet, reward_multiplier, deposit_amount, commission_rate, time_limit_minutes,
				currency, reward_pool_ton, reward_pool_usdt, status, created_at, updated_at
			FROM games
			WHERE short_id = $1
		`

		var game models.Game
		err := r.db.QueryRowContext(sqlCtx, query, shortID).Scan(
			&game.ID,
			&game.ShortID,
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
			&game.DepositAmount,
			&game.CommissionRate,
			&game.TimeLimitMinutes,
			&game.Currency,
			&game.RewardPoolTon,
			&game.RewardPoolUsdt,
			&game.Status,
			&game.CreatedAt,
			&game.UpdatedAt,
		)

		if err != nil {
			if err == sql.ErrNoRows {
				return nil, models.ErrGameNotFound
			}
			return nil, fmt.Errorf("failed to get game by short_id: %w", err)
		}

		return &game, nil
	})

	if err != nil {
		return nil, err
	}

	return result.(*models.Game), nil
}

// GetAll получает все игры с пагинацией
func (r *GameRepository) GetAll(ctx context.Context, limit, offset int) ([]*models.Game, error) {
	query := `
		SELECT id, short_id, creator_id, word, length, difficulty, max_tries, title, description,
			min_bet, max_bet, reward_multiplier, deposit_amount, commission_rate, time_limit_minutes,
			currency, reward_pool_ton, reward_pool_usdt, status, created_at, updated_at
		FROM games
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`
	return r.fetchGames(ctx, query, limit, offset)
}

// GetActive получает все активные игры с пагинацией
func (r *GameRepository) GetActive(ctx context.Context, limit, offset int) ([]*models.Game, error) {
	query := `
		SELECT id, short_id, creator_id, word, length, difficulty, max_tries, title, description,
			min_bet, max_bet, reward_multiplier, deposit_amount, commission_rate, time_limit_minutes,
			currency, reward_pool_ton, reward_pool_usdt, status, created_at, updated_at
		FROM games
		WHERE status = 'active'
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`
	return r.fetchGames(ctx, query, limit, offset)
}

// GetByCreator получает игры по ID создателя с пагинацией
func (r *GameRepository) GetByCreator(ctx context.Context, creatorID uint64, limit, offset int) ([]*models.Game, error) {
	query := `
		SELECT id, short_id, creator_id, word, length, difficulty, max_tries, title, description,
			min_bet, max_bet, reward_multiplier, deposit_amount, commission_rate, time_limit_minutes,
			currency, reward_pool_ton, reward_pool_usdt, status, created_at, updated_at
		FROM games
		WHERE creator_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	return r.fetchGames(ctx, query, creatorID, limit, offset)
}

// Helper to fetch games
func (r *GameRepository) fetchGames(ctx context.Context, query string, args ...interface{}) ([]*models.Game, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var games []*models.Game
	for rows.Next() {
		var game models.Game
		err := rows.Scan(
			&game.ID,
			&game.ShortID,
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
			&game.DepositAmount,
			&game.CommissionRate,
			&game.TimeLimitMinutes,
			&game.Currency,
			&game.RewardPoolTon,
			&game.RewardPoolUsdt,
			&game.Status,
			&game.CreatedAt,
			&game.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		games = append(games, &game)
	}
	return games, nil
}


// Update обновляет игру в базе данных
func (r *GameRepository) Update(ctx context.Context, game *models.Game) error {
	query := `
		UPDATE games
		SET short_id = $1, creator_id = $2, word = $3, length = $4, difficulty = $5, max_tries = $6,
			title = $7, description = $8, min_bet = $9, max_bet = $10, reward_multiplier = $11,
			deposit_amount = $12, commission_rate = $13, time_limit_minutes = $14,
			currency = $15, reward_pool_ton = $16, reward_pool_usdt = $17, status = $18, updated_at = $19
		WHERE id = $20
	`
	game.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(
		ctx, query,
		game.ShortID, game.CreatorID, game.Word, game.Length, game.Difficulty, game.MaxTries,
		game.Title, game.Description, game.MinBet, game.MaxBet, game.RewardMultiplier,
		game.DepositAmount, game.CommissionRate, game.TimeLimitMinutes,
		game.Currency, game.RewardPoolTon, game.RewardPoolUsdt, game.Status, game.UpdatedAt,
		game.ID,
	)
	return err
}

// Delete удаляет игру из базы данных
func (r *GameRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM games WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// GetByStatus получает игры по статусу
func (r *GameRepository) GetByStatus(ctx context.Context, status string, limit, offset int) ([]*models.Game, error) {
	query := `
		SELECT id, short_id, creator_id, word, length, difficulty, max_tries, title, description,
			min_bet, max_bet, reward_multiplier, deposit_amount, commission_rate, time_limit_minutes,
			currency, reward_pool_ton, reward_pool_usdt, status, created_at, updated_at
		FROM games
		WHERE status = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	return r.fetchGames(ctx, query, status, limit, offset)
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
	return count, err
}

// GetGameStats получает статистику по игре
func (r *GameRepository) GetGameStats(ctx context.Context, gameID uuid.UUID) (map[string]interface{}, error) {
	// Old implementation logic remains valid as it queries other tables
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
		TotalBets     sql.NullFloat64 `db:"total_bets"`
		TotalRewards  sql.NullFloat64 `db:"total_rewards"`
		AverageReward sql.NullFloat64 `db:"avg_reward"`
	}

	err := r.db.QueryRowContext(ctx, query, gameID).Scan(
		&stats.TotalLobbies,
		&stats.UniquePlayers,
		&stats.TotalBets,
		&stats.TotalRewards,
		&stats.AverageReward,
	)

	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"total_lobbies":  stats.TotalLobbies,
		"unique_players": stats.UniquePlayers,
		"total_bets":     stats.TotalBets.Float64,
		"total_rewards":  stats.TotalRewards.Float64,
		"average_reward": stats.AverageReward.Float64,
	}, nil
}

// SearchGames ищет игры по параметрам
func (r *GameRepository) SearchGames(ctx context.Context, minBet, maxBet float64, difficulty string, limit, offset int) ([]*models.Game, error) {
	query := `
		SELECT id, short_id, creator_id, word, length, difficulty, max_tries, title, description,
			min_bet, max_bet, reward_multiplier, deposit_amount, commission_rate, time_limit_minutes,
			currency, reward_pool_ton, reward_pool_usdt, status, created_at, updated_at
		FROM games
		WHERE status = 'active'
		AND ($1 = 0 OR min_bet >= $1)
		AND ($2 = 0 OR max_bet <= $2)
		AND ($3 = '' OR difficulty = $3)
		ORDER BY created_at DESC
		LIMIT $4 OFFSET $5
	`
	return r.fetchGames(ctx, query, minBet, maxBet, difficulty, limit, offset)
}

// UpdateStatus обновляет статус игры
func (r *GameRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	query := `UPDATE games SET status = $1, updated_at = $2 WHERE id = $3`
	_, err := r.db.ExecContext(ctx, query, status, time.Now(), id)
	return err
}

// UpdateRewardPool обновляет пул наград игры
func (r *GameRepository) UpdateRewardPool(ctx context.Context, id uuid.UUID, rewardPoolTon, rewardPoolUsdt float64) error {
	query := `UPDATE games SET reward_pool_ton = $1, reward_pool_usdt = $2, updated_at = $3 WHERE id = $4`
	_, err := r.db.ExecContext(ctx, query, rewardPoolTon, rewardPoolUsdt, time.Now(), id)
	return err
}

// CountActive возвращает количество активных игр
func (r *GameRepository) CountActive(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM games WHERE status = 'active'`
	var count int
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	return count, err
}

// CountByCreator возвращает количество игр создателя
func (r *GameRepository) CountByCreator(ctx context.Context, creatorID uint64) (int, error) {
	query := `SELECT COUNT(*) FROM games WHERE creator_id = $1`
	var count int
	err := r.db.QueryRowContext(ctx, query, creatorID).Scan(&count)
	return count, err
}

// GetByDifficulty получает игры по сложности
func (r *GameRepository) GetByDifficulty(ctx context.Context, difficulty string, limit, offset int) ([]*models.Game, error) {
	query := `
		SELECT id, short_id, creator_id, word, length, difficulty, max_tries, title, description,
			min_bet, max_bet, reward_multiplier, deposit_amount, commission_rate, time_limit_minutes,
			currency, reward_pool_ton, reward_pool_usdt, status, created_at, updated_at
		FROM games
		WHERE difficulty = $1 AND status = 'active'
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	return r.fetchGames(ctx, query, difficulty, limit, offset)
}

// GetByUserID получает игры по ID пользователя с пагинацией
func (r *GameRepository) GetByUserID(ctx context.Context, userID uint64, limit, offset int) ([]*models.Game, error) {
	query := `
		SELECT g.id, g.short_id, g.creator_id, g.word, g.length, g.difficulty, g.max_tries, g.title, g.description,
			g.min_bet, g.max_bet, g.reward_multiplier, g.deposit_amount, g.commission_rate, g.time_limit_minutes,
			g.currency, g.reward_pool_ton, g.reward_pool_usdt, g.status, g.created_at, g.updated_at
		FROM games g
		JOIN lobbies l ON g.id = l.game_id
		WHERE l.user_id = $1
		ORDER BY g.created_at DESC
		LIMIT $2 OFFSET $3
	`
	return r.fetchGames(ctx, query, userID, limit, offset)
}

// GetActiveByCreator получает активные игры создателя
func (r *GameRepository) GetActiveByCreator(ctx context.Context, creatorID uint64) ([]*models.Game, error) {
	query := `
		SELECT id, short_id, creator_id, word, length, difficulty, max_tries, title, description,
			min_bet, max_bet, reward_multiplier, deposit_amount, commission_rate, time_limit_minutes,
			currency, reward_pool_ton, reward_pool_usdt, status, created_at, updated_at
		FROM games
		WHERE creator_id = $1 AND status = 'active'
		ORDER BY created_at DESC
	`
	return r.fetchGames(ctx, query, creatorID)
}
