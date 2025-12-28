package postgresql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/TakuroBreath/wordle/internal/logger"
	"github.com/TakuroBreath/wordle/internal/models"
	"go.uber.org/zap"
)

// UserRepository представляет собой реализацию репозитория для работы с пользователями
type UserRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewUserRepository создает новый экземпляр UserRepository
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{
		db:     db,
		logger: logger.GetLogger(zap.String("repository", "user")),
	}
}

// Create создает нового пользователя в базе данных
func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	log := r.logger.With(zap.String("method", "Create"),
		zap.Uint64("telegram_id", user.TelegramID))
	log.Info("Creating new user",
		zap.String("username", user.Username),
		zap.String("first_name", user.FirstName),
		zap.String("last_name", user.LastName))

	query := `
		INSERT INTO users (telegram_id, username, first_name, last_name, wallet, balance_ton, balance_usdt, wins, losses, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	log.Debug("Executing SQL query",
		zap.String("query", query),
		zap.Uint64("telegram_id", user.TelegramID),
		zap.String("username", user.Username),
		zap.String("first_name", user.FirstName),
		zap.String("last_name", user.LastName),
		zap.String("wallet", user.Wallet),
		zap.Float64("balance_ton", user.BalanceTon),
		zap.Float64("balance_usdt", user.BalanceUsdt))

	_, err := r.db.ExecContext(
		ctx,
		query,
		user.TelegramID,
		user.Username,
		user.FirstName,
		user.LastName,
		user.Wallet,
		user.BalanceTon,
		user.BalanceUsdt,
		user.Wins,
		user.Losses,
		user.CreatedAt,
		user.UpdatedAt,
	)

	if err != nil {
		log.Error("Failed to create user", zap.Error(err))
		return fmt.Errorf("failed to create user: %w", err)
	}

	log.Info("User created successfully", zap.Uint64("telegram_id", user.TelegramID))
	return nil
}

// GetByTelegramID получает пользователя по Telegram ID
func (r *UserRepository) GetByTelegramID(ctx context.Context, telegramID uint64) (*models.User, error) {
	log := r.logger.With(zap.String("method", "GetByTelegramID"),
		zap.Uint64("telegram_id", telegramID))
	log.Info("Getting user by Telegram ID")

	query := `
		SELECT telegram_id, username, first_name, last_name, wallet, balance_ton, balance_usdt, wins, losses, created_at, updated_at
		FROM users
		WHERE telegram_id = $1
	`

	log.Debug("Executing SQL query", zap.String("query", query))

	var user models.User

	err := r.db.QueryRowContext(ctx, query, telegramID).Scan(
		&user.TelegramID,
		&user.Username,
		&user.FirstName,
		&user.LastName,
		&user.Wallet,
		&user.BalanceTon,
		&user.BalanceUsdt,
		&user.Wins,
		&user.Losses,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			log.Warn("User not found", zap.Uint64("telegram_id", telegramID))
			return nil, models.ErrUserNotFound
		}
		log.Error("Failed to get user", zap.Error(err))
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	log.Info("User found",
		zap.Uint64("telegram_id", user.TelegramID),
		zap.String("username", user.Username))
	return &user, nil
}

// Update обновляет пользователя в базе данных
func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	log := r.logger.With(zap.String("method", "Update"),
		zap.Uint64("telegram_id", user.TelegramID))
	log.Info("Updating user", zap.String("username", user.Username))

	query := `
		UPDATE users
		SET username = $1, first_name = $2, last_name = $3, wallet = $4, balance_ton = $5, balance_usdt = $6, wins = $7, losses = $8, updated_at = $9
		WHERE telegram_id = $10
	`

	user.UpdatedAt = time.Now()

	log.Debug("Executing SQL query",
		zap.String("query", query),
		zap.Uint64("telegram_id", user.TelegramID),
		zap.String("username", user.Username),
		zap.String("first_name", user.FirstName),
		zap.String("last_name", user.LastName))

	_, err := r.db.ExecContext(
		ctx,
		query,
		user.Username,
		user.FirstName,
		user.LastName,
		user.Wallet,
		user.BalanceTon,
		user.BalanceUsdt,
		user.Wins,
		user.Losses,
		user.UpdatedAt,
		user.TelegramID,
	)

	if err != nil {
		log.Error("Failed to update user", zap.Error(err))
		return fmt.Errorf("failed to update user: %w", err)
	}

	log.Info("User updated successfully", zap.Uint64("telegram_id", user.TelegramID))
	return nil
}

// Delete удаляет пользователя из базы данных
func (r *UserRepository) Delete(ctx context.Context, telegramID uint64) error {
	log := r.logger.With(zap.String("method", "Delete"),
		zap.Uint64("telegram_id", telegramID))
	log.Info("Deleting user")

	query := `DELETE FROM users WHERE telegram_id = $1`

	log.Debug("Executing SQL query", zap.String("query", query))

	_, err := r.db.ExecContext(ctx, query, telegramID)
	if err != nil {
		log.Error("Failed to delete user", zap.Error(err))
		return fmt.Errorf("failed to delete user: %w", err)
	}

	log.Info("User deleted successfully", zap.Uint64("telegram_id", telegramID))
	return nil
}

// GetByID получает пользователя по ID
func (r *UserRepository) GetByID(ctx context.Context, id uint64) (*models.User, error) {
	log := r.logger.With(zap.String("method", "GetByID"),
		zap.Uint64("id", id))
	log.Info("Getting user by ID (alias to GetByTelegramID)")
	return r.GetByTelegramID(ctx, id)
}

// GetByUsername получает пользователя по имени пользователя
func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	log := r.logger.With(zap.String("method", "GetByUsername"),
		zap.String("username", username))
	log.Info("Getting user by username")

	query := `
		SELECT telegram_id, username, first_name, last_name, wallet, balance_ton, balance_usdt, wins, losses, created_at, updated_at
		FROM users
		WHERE username = $1
	`

	log.Debug("Executing SQL query", zap.String("query", query))

	var user models.User

	err := r.db.QueryRowContext(ctx, query, username).Scan(
		&user.TelegramID,
		&user.Username,
		&user.FirstName,
		&user.LastName,
		&user.Wallet,
		&user.BalanceTon,
		&user.BalanceUsdt,
		&user.Wins,
		&user.Losses,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			log.Warn("User not found", zap.String("username", username))
			return nil, models.ErrUserNotFound
		}
		log.Error("Failed to get user", zap.Error(err))
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	log.Info("User found",
		zap.Uint64("telegram_id", user.TelegramID),
		zap.String("username", user.Username))
	return &user, nil
}

// UpdateBalance обновляет баланс пользователя
func (r *UserRepository) UpdateBalance(ctx context.Context, id uint64, amount float64) error {
	log := r.logger.With(zap.String("method", "UpdateBalance"),
		zap.Uint64("user_id", id))
	log.Info("Updating user balance", zap.Float64("amount", amount))

	query := `
		UPDATE users
		SET balance_ton = balance_ton + $1, updated_at = $2
		WHERE telegram_id = $3
	`

	log.Debug("Executing SQL query", zap.String("query", query))

	_, err := r.db.ExecContext(ctx, query, amount, time.Now(), id)
	if err != nil {
		log.Error("Failed to update user balance", zap.Error(err))
		return fmt.Errorf("failed to update user balance: %w", err)
	}

	log.Info("User balance updated successfully",
		zap.Uint64("user_id", id),
		zap.Float64("amount", amount))
	return nil
}

// GetTopUsers получает топ пользователей
func (r *UserRepository) GetTopUsers(ctx context.Context, limit int) ([]*models.User, error) {
	log := r.logger.With(zap.String("method", "GetTopUsers"))
	log.Info("Getting top users", zap.Int("limit", limit))

	query := `
		SELECT telegram_id, username, first_name, last_name, wallet, balance_ton, balance_usdt, wins, losses, created_at, updated_at
		FROM users
		ORDER BY wins DESC, balance_ton DESC
		LIMIT $1
	`

	log.Debug("Executing SQL query", zap.String("query", query))

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		log.Error("Failed to get top users", zap.Error(err))
		return nil, fmt.Errorf("failed to get top users: %w", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		var user models.User

		err := rows.Scan(
			&user.TelegramID,
			&user.Username,
			&user.FirstName,
			&user.LastName,
			&user.Wallet,
			&user.BalanceTon,
			&user.BalanceUsdt,
			&user.Wins,
			&user.Losses,
			&user.CreatedAt,
			&user.UpdatedAt,
		)

		if err != nil {
			log.Error("Failed to scan user", zap.Error(err))
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}

		users = append(users, &user)
	}

	if err = rows.Err(); err != nil {
		log.Error("Error iterating top users", zap.Error(err))
		return nil, fmt.Errorf("error iterating top users: %w", err)
	}

	log.Info("Retrieved top users", zap.Int("count", len(users)))
	return users, nil
}

// GetUserStats получает статистику пользователя
func (r *UserRepository) GetUserStats(ctx context.Context, userID uint64) (map[string]interface{}, error) {
	log := r.logger.With(zap.String("method", "GetUserStats"),
		zap.Uint64("user_id", userID))
	log.Info("Getting user statistics")

	query := `
		SELECT 
			COUNT(DISTINCT g.id) as total_games,
			COUNT(DISTINCT CASE WHEN h.status = 'win' THEN h.id END) as wins,
			COUNT(DISTINCT CASE WHEN h.status = 'lose' THEN h.id END) as losses,
			SUM(CASE WHEN h.status = 'win' THEN h.reward ELSE 0 END) as total_rewards,
			AVG(CASE WHEN h.status = 'win' THEN h.reward ELSE NULL END) as avg_win_reward
		FROM games g
		LEFT JOIN history h ON g.id = h.game_id AND h.user_id = $1
		WHERE g.creator_id = $1 OR h.user_id = $1
	`

	log.Debug("Executing SQL query", zap.String("query", query))

	var stats struct {
		TotalGames   int     `db:"total_games"`
		Wins         int     `db:"wins"`
		Losses       int     `db:"losses"`
		TotalRewards float64 `db:"total_rewards"`
		AvgWinReward float64 `db:"avg_win_reward"`
	}

	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&stats.TotalGames,
		&stats.Wins,
		&stats.Losses,
		&stats.TotalRewards,
		&stats.AvgWinReward,
	)

	if err != nil {
		log.Error("Failed to get user stats", zap.Error(err))
		return nil, fmt.Errorf("failed to get user stats: %w", err)
	}

	log.Info("User statistics retrieved",
		zap.Int("total_games", stats.TotalGames),
		zap.Int("wins", stats.Wins),
		zap.Int("losses", stats.Losses),
		zap.Float64("total_rewards", stats.TotalRewards))

	return map[string]interface{}{
		"total_games":    stats.TotalGames,
		"wins":           stats.Wins,
		"losses":         stats.Losses,
		"total_rewards":  stats.TotalRewards,
		"avg_win_reward": stats.AvgWinReward,
	}, nil
}

// ValidateBalance проверяет достаточность средств
func (r *UserRepository) ValidateBalance(ctx context.Context, userID uint64, requiredAmount float64) (bool, error) {
	log := r.logger.With(zap.String("method", "ValidateBalance"),
		zap.Uint64("user_id", userID))
	log.Info("Validating user balance", zap.Float64("required_amount", requiredAmount))

	query := `
		SELECT balance_ton >= $2
		FROM users
		WHERE telegram_id = $1
	`

	log.Debug("Executing SQL query", zap.String("query", query))

	var hasEnough bool
	err := r.db.QueryRowContext(ctx, query, userID, requiredAmount).Scan(&hasEnough)
	if err != nil {
		log.Error("Failed to validate balance", zap.Error(err))
		return false, fmt.Errorf("failed to validate balance: %w", err)
	}

	log.Info("Balance validation completed",
		zap.Bool("has_enough", hasEnough),
		zap.Float64("required_amount", requiredAmount))
	return hasEnough, nil
}

// UpdateTonBalance обновляет баланс TON пользователя
func (r *UserRepository) UpdateTonBalance(ctx context.Context, telegramID uint64, amount float64) error {
	log := r.logger.With(zap.String("method", "UpdateTonBalance"),
		zap.Uint64("telegram_id", telegramID))
	log.Info("Updating TON balance", zap.Float64("amount", amount))

	query := `UPDATE users SET balance_ton = balance_ton + $1, updated_at = NOW() WHERE telegram_id = $2`

	log.Debug("Executing SQL query", zap.String("query", query))

	_, err := r.db.ExecContext(ctx, query, amount, telegramID)
	if err != nil {
		log.Error("Failed to update TON balance", zap.Error(err))
		return fmt.Errorf("failed to update TON balance: %w", err)
	}

	log.Info("TON balance updated successfully",
		zap.Uint64("telegram_id", telegramID),
		zap.Float64("amount", amount))
	return nil
}

// UpdateUsdtBalance обновляет баланс USDT пользователя
func (r *UserRepository) UpdateUsdtBalance(ctx context.Context, telegramID uint64, amount float64) error {
	log := r.logger.With(zap.String("method", "UpdateUsdtBalance"),
		zap.Uint64("telegram_id", telegramID))
	log.Info("Updating USDT balance", zap.Float64("amount", amount))

	query := `UPDATE users SET balance_usdt = balance_usdt + $1, updated_at = NOW() WHERE telegram_id = $2`

	log.Debug("Executing SQL query", zap.String("query", query))

	_, err := r.db.ExecContext(ctx, query, amount, telegramID)
	if err != nil {
		log.Error("Failed to update USDT balance", zap.Error(err))
		return fmt.Errorf("failed to update USDT balance: %w", err)
	}

	log.Info("USDT balance updated successfully",
		zap.Uint64("telegram_id", telegramID),
		zap.Float64("amount", amount))
	return nil
}

// GetByWallet получает пользователя по адресу кошелька
func (r *UserRepository) GetByWallet(ctx context.Context, wallet string) (*models.User, error) {
	query := `
		SELECT telegram_id, username, first_name, last_name, wallet, balance_ton, balance_usdt,
			COALESCE(pending_withdrawal, 0), withdrawal_lock_until, wins, losses,
			COALESCE(total_deposited, 0), COALESCE(total_withdrawn, 0), created_at, updated_at
		FROM users
		WHERE wallet = $1
	`

	var user models.User
	var withdrawalLockUntil sql.NullTime
	err := r.db.QueryRowContext(ctx, query, wallet).Scan(
		&user.TelegramID,
		&user.Username,
		&user.FirstName,
		&user.LastName,
		&user.Wallet,
		&user.BalanceTon,
		&user.BalanceUsdt,
		&user.PendingWithdrawal,
		&withdrawalLockUntil,
		&user.Wins,
		&user.Losses,
		&user.TotalDeposited,
		&user.TotalWithdrawn,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, models.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by wallet: %w", err)
	}

	if withdrawalLockUntil.Valid {
		user.WithdrawalLockUntil = &withdrawalLockUntil.Time
	}

	return &user, nil
}

// UpdateWallet обновляет адрес кошелька пользователя
func (r *UserRepository) UpdateWallet(ctx context.Context, telegramID uint64, wallet string) error {
	query := `UPDATE users SET wallet = $1, updated_at = $2 WHERE telegram_id = $3`
	_, err := r.db.ExecContext(ctx, query, wallet, time.Now(), telegramID)
	if err != nil {
		return fmt.Errorf("failed to update wallet: %w", err)
	}
	return nil
}

// UpdatePendingWithdrawal обновляет сумму в процессе вывода
func (r *UserRepository) UpdatePendingWithdrawal(ctx context.Context, telegramID uint64, amount float64) error {
	query := `UPDATE users SET pending_withdrawal = COALESCE(pending_withdrawal, 0) + $1, updated_at = $2 WHERE telegram_id = $3`
	_, err := r.db.ExecContext(ctx, query, amount, time.Now(), telegramID)
	if err != nil {
		return fmt.Errorf("failed to update pending withdrawal: %w", err)
	}
	return nil
}

// SetWithdrawalLock устанавливает блокировку вывода до указанного времени
func (r *UserRepository) SetWithdrawalLock(ctx context.Context, telegramID uint64, lockUntil time.Time) error {
	query := `UPDATE users SET withdrawal_lock_until = $1, updated_at = $2 WHERE telegram_id = $3`
	_, err := r.db.ExecContext(ctx, query, lockUntil, time.Now(), telegramID)
	if err != nil {
		return fmt.Errorf("failed to set withdrawal lock: %w", err)
	}
	return nil
}

// IncrementWins увеличивает количество побед
func (r *UserRepository) IncrementWins(ctx context.Context, telegramID uint64) error {
	query := `UPDATE users SET wins = wins + 1, updated_at = $1 WHERE telegram_id = $2`
	_, err := r.db.ExecContext(ctx, query, time.Now(), telegramID)
	if err != nil {
		return fmt.Errorf("failed to increment wins: %w", err)
	}
	return nil
}

// IncrementLosses увеличивает количество поражений
func (r *UserRepository) IncrementLosses(ctx context.Context, telegramID uint64) error {
	query := `UPDATE users SET losses = losses + 1, updated_at = $1 WHERE telegram_id = $2`
	_, err := r.db.ExecContext(ctx, query, time.Now(), telegramID)
	if err != nil {
		return fmt.Errorf("failed to increment losses: %w", err)
	}
	return nil
}
