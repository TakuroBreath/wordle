package postgresql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/TakuroBreath/wordle/internal/models"
	"github.com/google/uuid"
)

// TransactionRepository представляет собой реализацию репозитория для работы с транзакциями
type TransactionRepository struct {
	db *sql.DB
}

// NewTransactionRepository создает новый экземпляр TransactionRepository
func NewTransactionRepository(db *sql.DB) *TransactionRepository {
	return &TransactionRepository{
		db: db,
	}
}

// Create создает новую транзакцию в базе данных
func (r *TransactionRepository) Create(ctx context.Context, transaction *models.Transaction) error {
	query := `
		INSERT INTO transactions (id, user_id, amount, type, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	// Генерация UUID, если он не был установлен
	if transaction.ID == uuid.Nil {
		transaction.ID = uuid.New()
	}

	// Установка текущего времени
	now := time.Now()
	transaction.CreatedAt = now
	transaction.UpdatedAt = now

	_, err := r.db.ExecContext(
		ctx,
		query,
		transaction.ID,
		transaction.UserID,
		transaction.Amount,
		transaction.Type,
		transaction.Status,
		transaction.CreatedAt,
		transaction.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}

	return nil
}

// GetByID получает транзакцию по ID
func (r *TransactionRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Transaction, error) {
	query := `
		SELECT id, user_id, amount, type, status, created_at, updated_at
		FROM transactions
		WHERE id = $1
	`

	var transaction models.Transaction

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&transaction.ID,
		&transaction.UserID,
		&transaction.Amount,
		&transaction.Type,
		&transaction.Status,
		&transaction.CreatedAt,
		&transaction.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("transaction not found")
		}
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	return &transaction, nil
}

// GetByUserID получает все транзакции пользователя с пагинацией
func (r *TransactionRepository) GetByUserID(ctx context.Context, userID uint64, limit, offset int) ([]*models.Transaction, error) {
	query := `
		SELECT id, user_id, amount, type, status, created_at, updated_at
		FROM transactions
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get user transactions: %w", err)
	}
	defer rows.Close()

	var transactions []*models.Transaction
	for rows.Next() {
		var transaction models.Transaction

		err := rows.Scan(
			&transaction.ID,
			&transaction.UserID,
			&transaction.Amount,
			&transaction.Type,
			&transaction.Status,
			&transaction.CreatedAt,
			&transaction.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan transaction: %w", err)
		}

		transactions = append(transactions, &transaction)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating user transactions: %w", err)
	}

	return transactions, nil
}

// GetByStatus получает все транзакции с определенным статусом
func (r *TransactionRepository) GetByStatus(ctx context.Context, status string) ([]*models.Transaction, error) {
	query := `
		SELECT id, user_id, amount, type, status, created_at, updated_at
		FROM transactions
		WHERE status = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, status)
	if err != nil {
		return nil, fmt.Errorf("failed to get transactions by status: %w", err)
	}
	defer rows.Close()

	var transactions []*models.Transaction
	for rows.Next() {
		var transaction models.Transaction

		err := rows.Scan(
			&transaction.ID,
			&transaction.UserID,
			&transaction.Amount,
			&transaction.Type,
			&transaction.Status,
			&transaction.CreatedAt,
			&transaction.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan transaction: %w", err)
		}

		transactions = append(transactions, &transaction)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating transactions by status: %w", err)
	}

	return transactions, nil
}

// GetByType получает транзакции по типу с пагинацией
func (r *TransactionRepository) GetByType(ctx context.Context, transactionType string, limit, offset int) ([]*models.Transaction, error) {
	query := `
		SELECT id, user_id, amount, type, status, created_at, updated_at
		FROM transactions
		WHERE type = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, transactionType, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get transactions by type: %w", err)
	}
	defer rows.Close()

	var transactions []*models.Transaction
	for rows.Next() {
		var transaction models.Transaction

		err := rows.Scan(
			&transaction.ID,
			&transaction.UserID,
			&transaction.Amount,
			&transaction.Type,
			&transaction.Status,
			&transaction.CreatedAt,
			&transaction.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan transaction: %w", err)
		}

		transactions = append(transactions, &transaction)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating transactions by type: %w", err)
	}

	return transactions, nil
}

// Update обновляет транзакцию в базе данных
func (r *TransactionRepository) Update(ctx context.Context, transaction *models.Transaction) error {
	query := `
		UPDATE transactions
		SET user_id = $1, amount = $2, type = $3, status = $4, updated_at = $5
		WHERE id = $6
	`

	transaction.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(
		ctx,
		query,
		transaction.UserID,
		transaction.Amount,
		transaction.Type,
		transaction.Status,
		transaction.UpdatedAt,
		transaction.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update transaction: %w", err)
	}

	return nil
}

// Delete удаляет транзакцию из базы данных
func (r *TransactionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM transactions WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete transaction: %w", err)
	}

	return nil
}

// CountByUser возвращает количество транзакций пользователя
func (r *TransactionRepository) CountByUser(ctx context.Context, userID uint64) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM transactions
		WHERE user_id = $1
	`

	var count int
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count user transactions: %w", err)
	}

	return count, nil
}

// GetUserBalance получает баланс пользователя
func (r *TransactionRepository) GetUserBalance(ctx context.Context, userID uint64) (float64, error) {
	query := `
		SELECT COALESCE(SUM(CASE WHEN type = 'deposit' THEN amount ELSE -amount END), 0)
		FROM transactions
		WHERE user_id = $1 AND status = 'completed'
	`

	var balance float64
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&balance)
	if err != nil {
		return 0, fmt.Errorf("failed to get user balance: %w", err)
	}

	return balance, nil
}
