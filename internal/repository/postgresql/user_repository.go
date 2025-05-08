package postgresql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/TakuroBreath/wordle/internal/models"
)

// UserRepository представляет собой реализацию репозитория для работы с пользователями
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository создает новый экземпляр UserRepository
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

// Create создает нового пользователя в базе данных
func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (telegram_id, username, wallet, balance_ton, balance_usdt, wins, losses, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	_, err := r.db.ExecContext(
		ctx,
		query,
		user.TelegramID,
		user.Username,
		user.Wallet,
		user.BalanceTon,
		user.BalanceUsdt,
		user.Wins,
		user.Losses,
		user.CreatedAt,
		user.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetByTelegramID получает пользователя по Telegram ID
func (r *UserRepository) GetByTelegramID(ctx context.Context, telegramID uint64) (*models.User, error) {
	query := `
		SELECT telegram_id, username, wallet, balance_ton, balance_usdt, wins, losses, created_at, updated_at
		FROM users
		WHERE telegram_id = $1
	`

	var user models.User

	err := r.db.QueryRowContext(ctx, query, telegramID).Scan(
		&user.TelegramID,
		&user.Username,
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
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// Update обновляет пользователя в базе данных
func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	query := `
		UPDATE users
		SET username = $1, wallet = $2, balance_ton = $3, balance_usdt = $4, wins = $5, losses = $6, updated_at = $7
		WHERE telegram_id = $8
	`

	user.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(
		ctx,
		query,
		user.Username,
		user.Wallet,
		user.BalanceTon,
		user.BalanceUsdt,
		user.Wins,
		user.Losses,
		user.UpdatedAt,
		user.TelegramID,
	)

	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// Delete удаляет пользователя из базы данных
func (r *UserRepository) Delete(ctx context.Context, telegramID uint64) error {
	query := `DELETE FROM users WHERE telegram_id = $1`

	_, err := r.db.ExecContext(ctx, query, telegramID)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

// GetByID получает пользователя по ID
func (r *UserRepository) GetByID(ctx context.Context, id uint64) (*models.User, error) {
	return r.GetByTelegramID(ctx, id)
}

// GetByUsername получает пользователя по имени пользователя
func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	query := `
		SELECT telegram_id, username, wallet, balance_ton, balance_usdt, wins, losses, created_at, updated_at
		FROM users
		WHERE username = $1
	`

	var user models.User

	err := r.db.QueryRowContext(ctx, query, username).Scan(
		&user.TelegramID,
		&user.Username,
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
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// UpdateBalance обновляет баланс пользователя
func (r *UserRepository) UpdateBalance(ctx context.Context, id uint64, amount float64) error {
	query := `
		UPDATE users
		SET balance_ton = balance_ton + $1, updated_at = $2
		WHERE telegram_id = $3
	`

	_, err := r.db.ExecContext(ctx, query, amount, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update user balance: %w", err)
	}

	return nil
}

// GetTopUsers получает топ пользователей
func (r *UserRepository) GetTopUsers(ctx context.Context, limit int) ([]*models.User, error) {
	query := `
		SELECT telegram_id, username, wallet, balance_ton, balance_usdt, wins, losses, created_at, updated_at
		FROM users
		ORDER BY wins DESC, balance_ton DESC
		LIMIT $1
	`

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get top users: %w", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		var user models.User

		err := rows.Scan(
			&user.TelegramID,
			&user.Username,
			&user.Wallet,
			&user.BalanceTon,
			&user.BalanceUsdt,
			&user.Wins,
			&user.Losses,
			&user.CreatedAt,
			&user.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}

		users = append(users, &user)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating top users: %w", err)
	}

	return users, nil
}

// GetUserStats получает статистику пользователя
func (r *UserRepository) GetUserStats(ctx context.Context, userID uint64) (map[string]interface{}, error) {
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
		return nil, fmt.Errorf("failed to get user stats: %w", err)
	}

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
	query := `
		SELECT balance >= $2
		FROM users
		WHERE id = $1
	`

	var hasEnough bool
	err := r.db.QueryRowContext(ctx, query, userID, requiredAmount).Scan(&hasEnough)
	if err != nil {
		return false, fmt.Errorf("failed to validate balance: %w", err)
	}

	return hasEnough, nil
}
