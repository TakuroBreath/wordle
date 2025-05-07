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
