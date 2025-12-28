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
		INSERT INTO transactions (
			id, user_id, amount, type, status, currency, description, tx_hash, 
			network, game_id, lobby_id, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`

	// Генерация UUID, если он не был установлен
	if transaction.ID == uuid.Nil {
		transaction.ID = uuid.New()
	}

	// Установка текущего времени
	now := time.Now()
	if transaction.CreatedAt.IsZero() {
		transaction.CreatedAt = now
	}
	transaction.UpdatedAt = now

	_, err := r.db.ExecContext(
		ctx,
		query,
		transaction.ID,
		transaction.UserID,
		transaction.Amount,
		transaction.Type,
		transaction.Status,
		transaction.Currency,
		transaction.Description,
		transaction.TxHash,
		transaction.Network,
		transaction.GameID,
		transaction.LobbyID,
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
		SELECT id, user_id, amount, type, status, currency, description, tx_hash, 
		       network, game_id, lobby_id, created_at, updated_at
		FROM transactions
		WHERE id = $1
	`

	var transaction models.Transaction
	var gameID, lobbyID sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&transaction.ID,
		&transaction.UserID,
		&transaction.Amount,
		&transaction.Type,
		&transaction.Status,
		&transaction.Currency,
		&transaction.Description,
		&transaction.TxHash,
		&transaction.Network,
		&gameID,
		&lobbyID,
		&transaction.CreatedAt,
		&transaction.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("transaction not found")
		}
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	// Преобразование nullable полей
	if gameID.Valid {
		gameUUID, err := uuid.Parse(gameID.String)
		if err == nil {
			transaction.GameID = &gameUUID
		}
	}

	if lobbyID.Valid {
		lobbyUUID, err := uuid.Parse(lobbyID.String)
		if err == nil {
			transaction.LobbyID = &lobbyUUID
		}
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

// GetTransactionStats получает статистику по транзакциям
func (r *TransactionRepository) GetTransactionStats(ctx context.Context, userID uint64) (map[string]interface{}, error) {
	query := `
		SELECT 
			COUNT(*) as total_transactions,
			COUNT(DISTINCT type) as unique_types,
			SUM(CASE WHEN type = 'deposit' THEN amount ELSE 0 END) as total_deposits,
			SUM(CASE WHEN type = 'withdraw' THEN amount ELSE 0 END) as total_withdrawals,
			MAX(created_at) as last_transaction_time,
			COUNT(CASE WHEN status = 'completed' THEN 1 END) as completed_transactions,
			COUNT(CASE WHEN status = 'pending' THEN 1 END) as pending_transactions
		FROM transactions
		WHERE user_id = $1
	`

	var stats struct {
		TotalTransactions     int       `db:"total_transactions"`
		UniqueTypes           int       `db:"unique_types"`
		TotalDeposits         float64   `db:"total_deposits"`
		TotalWithdrawals      float64   `db:"total_withdrawals"`
		LastTransactionTime   time.Time `db:"last_transaction_time"`
		CompletedTransactions int       `db:"completed_transactions"`
		PendingTransactions   int       `db:"pending_transactions"`
	}

	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&stats.TotalTransactions,
		&stats.UniqueTypes,
		&stats.TotalDeposits,
		&stats.TotalWithdrawals,
		&stats.LastTransactionTime,
		&stats.CompletedTransactions,
		&stats.PendingTransactions,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get transaction stats: %w", err)
	}

	return map[string]interface{}{
		"total_transactions":     stats.TotalTransactions,
		"unique_types":           stats.UniqueTypes,
		"total_deposits":         stats.TotalDeposits,
		"total_withdrawals":      stats.TotalWithdrawals,
		"last_transaction_time":  stats.LastTransactionTime,
		"completed_transactions": stats.CompletedTransactions,
		"pending_transactions":   stats.PendingTransactions,
		"net_balance":            stats.TotalDeposits - stats.TotalWithdrawals,
	}, nil
}

// CheckSufficientFunds проверяет достаточность средств
func (r *TransactionRepository) CheckSufficientFunds(ctx context.Context, userID uint64, amount float64) (bool, error) {
	query := `
		SELECT COALESCE(SUM(CASE WHEN type = 'deposit' THEN amount ELSE -amount END), 0) >= $2
		FROM transactions
		WHERE user_id = $1 AND status = 'completed'
	`

	var hasEnough bool
	err := r.db.QueryRowContext(ctx, query, userID, amount).Scan(&hasEnough)
	if err != nil {
		return false, fmt.Errorf("failed to check sufficient funds: %w", err)
	}

	return hasEnough, nil
}

// GetByTxHash получает транзакцию по хешу блокчейна
func (r *TransactionRepository) GetByTxHash(ctx context.Context, txHash string) (*models.Transaction, error) {
	query := `
		SELECT id, user_id, type, amount, currency, status, tx_hash, network, game_id, lobby_id,
			COALESCE(description, ''), COALESCE(fee, 0), COALESCE(blockchain_lt, 0),
			COALESCE(from_address, ''), COALESCE(to_address, ''), COALESCE(comment, ''),
			COALESCE(game_short_id, ''), COALESCE(error_message, ''), COALESCE(confirmations, 0),
			processed_at, created_at, updated_at
		FROM transactions
		WHERE tx_hash = $1
	`

	var tx models.Transaction
	var processedAt sql.NullTime
	var gameID, lobbyID sql.NullString
	var network sql.NullString

	err := r.db.QueryRowContext(ctx, query, txHash).Scan(
		&tx.ID,
		&tx.UserID,
		&tx.Type,
		&tx.Amount,
		&tx.Currency,
		&tx.Status,
		&tx.TxHash,
		&network,
		&gameID,
		&lobbyID,
		&tx.Description,
		&tx.Fee,
		&tx.BlockchainLt,
		&tx.FromAddress,
		&tx.ToAddress,
		&tx.Comment,
		&tx.GameShortID,
		&tx.ErrorMessage,
		&tx.Confirmations,
		&processedAt,
		&tx.CreatedAt,
		&tx.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, models.ErrTransactionNotFound
		}
		return nil, fmt.Errorf("failed to get transaction by tx_hash: %w", err)
	}

	if network.Valid {
		tx.Network = network.String
	}
	if gameID.Valid {
		id, _ := uuid.Parse(gameID.String)
		tx.GameID = &id
	}
	if lobbyID.Valid {
		id, _ := uuid.Parse(lobbyID.String)
		tx.LobbyID = &id
	}
	if processedAt.Valid {
		tx.ProcessedAt = &processedAt.Time
	}

	return &tx, nil
}

// ExistsByTxHash проверяет существование транзакции по хешу
func (r *TransactionRepository) ExistsByTxHash(ctx context.Context, txHash string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM transactions WHERE tx_hash = $1)`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, txHash).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check transaction existence: %w", err)
	}
	return exists, nil
}

// GetByStatus получает транзакции по статусу
func (r *TransactionRepository) GetByStatus(ctx context.Context, status string, limit, offset int) ([]*models.Transaction, error) {
	query := `
		SELECT id, user_id, type, amount, currency, status, tx_hash, network, game_id, lobby_id,
			COALESCE(description, ''), created_at, updated_at
		FROM transactions
		WHERE status = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, status, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get transactions by status: %w", err)
	}
	defer rows.Close()

	var transactions []*models.Transaction
	for rows.Next() {
		var tx models.Transaction
		var gameID, lobbyID, network sql.NullString

		err := rows.Scan(
			&tx.ID,
			&tx.UserID,
			&tx.Type,
			&tx.Amount,
			&tx.Currency,
			&tx.Status,
			&tx.TxHash,
			&network,
			&gameID,
			&lobbyID,
			&tx.Description,
			&tx.CreatedAt,
			&tx.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan transaction: %w", err)
		}

		if network.Valid {
			tx.Network = network.String
		}
		if gameID.Valid {
			id, _ := uuid.Parse(gameID.String)
			tx.GameID = &id
		}
		if lobbyID.Valid {
			id, _ := uuid.Parse(lobbyID.String)
			tx.LobbyID = &id
		}

		transactions = append(transactions, &tx)
	}

	return transactions, rows.Err()
}

// GetPendingByGameShortID получает pending транзакции для игры
func (r *TransactionRepository) GetPendingByGameShortID(ctx context.Context, gameShortID string) ([]*models.Transaction, error) {
	query := `
		SELECT id, user_id, type, amount, currency, status, tx_hash, created_at, updated_at
		FROM transactions
		WHERE game_short_id = $1 AND status = 'pending'
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, gameShortID)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending transactions: %w", err)
	}
	defer rows.Close()

	var transactions []*models.Transaction
	for rows.Next() {
		var tx models.Transaction
		err := rows.Scan(
			&tx.ID,
			&tx.UserID,
			&tx.Type,
			&tx.Amount,
			&tx.Currency,
			&tx.Status,
			&tx.TxHash,
			&tx.CreatedAt,
			&tx.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan transaction: %w", err)
		}
		transactions = append(transactions, &tx)
	}

	return transactions, rows.Err()
}

// GetPendingWithdrawals получает pending выводы
func (r *TransactionRepository) GetPendingWithdrawals(ctx context.Context, limit int) ([]*models.Transaction, error) {
	query := `
		SELECT id, user_id, type, amount, currency, status, 
			COALESCE(to_address, ''), COALESCE(fee, 0), COALESCE(comment, ''),
			created_at, updated_at
		FROM transactions
		WHERE type = 'withdraw' AND status = 'pending'
		ORDER BY created_at ASC
		LIMIT $1
	`

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending withdrawals: %w", err)
	}
	defer rows.Close()

	var transactions []*models.Transaction
	for rows.Next() {
		var tx models.Transaction
		err := rows.Scan(
			&tx.ID,
			&tx.UserID,
			&tx.Type,
			&tx.Amount,
			&tx.Currency,
			&tx.Status,
			&tx.ToAddress,
			&tx.Fee,
			&tx.Comment,
			&tx.CreatedAt,
			&tx.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan transaction: %w", err)
		}
		transactions = append(transactions, &tx)
	}

	return transactions, rows.Err()
}

// GetLastProcessedLt получает последний обработанный lt
func (r *TransactionRepository) GetLastProcessedLt(ctx context.Context) (int64, error) {
	query := `SELECT COALESCE(last_processed_lt, 0) FROM blockchain_state WHERE id = 1`
	var lt int64
	err := r.db.QueryRowContext(ctx, query).Scan(&lt)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to get last processed lt: %w", err)
	}
	return lt, nil
}

// UpdateLastProcessedLt обновляет последний обработанный lt
func (r *TransactionRepository) UpdateLastProcessedLt(ctx context.Context, lt int64) error {
	query := `
		INSERT INTO blockchain_state (id, last_processed_lt, updated_at)
		VALUES (1, $1, $2)
		ON CONFLICT (id) DO UPDATE SET last_processed_lt = $1, updated_at = $2
	`
	_, err := r.db.ExecContext(ctx, query, lt, time.Now())
	if err != nil {
		return fmt.Errorf("failed to update last processed lt: %w", err)
	}
	return nil
}
