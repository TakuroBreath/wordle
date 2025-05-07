package postgresql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/TakuroBreath/wordle/internal/models"
	"github.com/google/uuid"
)

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
		INSERT INTO lobbies (id, game_id, user_id, bet_amount, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	// Генерация UUID, если он не был установлен
	if lobby.ID == uuid.Nil {
		lobby.ID = uuid.New()
	}

	// Установка текущего времени
	now := time.Now()
	lobby.CreatedAt = now
	lobby.UpdatedAt = now

	_, err := r.db.ExecContext(
		ctx,
		query,
		lobby.ID,
		lobby.GameID,
		lobby.UserID,
		lobby.BetAmount,
		lobby.Status,
		lobby.CreatedAt,
		lobby.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create lobby: %w", err)
	}

	return nil
}

// GetByID получает лобби по ID
func (r *LobbyRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Lobby, error) {
	query := `
		SELECT id, game_id, user_id, bet_amount, status, created_at, updated_at
		FROM lobbies
		WHERE id = $1
	`

	var lobby models.Lobby

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&lobby.ID,
		&lobby.GameID,
		&lobby.UserID,
		&lobby.BetAmount,
		&lobby.Status,
		&lobby.CreatedAt,
		&lobby.UpdatedAt,
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
func (r *LobbyRepository) GetByGameID(ctx context.Context, gameID uuid.UUID) ([]*models.Lobby, error) {
	query := `
		SELECT id, game_id, user_id, bet_amount, status, created_at, updated_at
		FROM lobbies
		WHERE game_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, gameID)
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
			&lobby.BetAmount,
			&lobby.Status,
			&lobby.CreatedAt,
			&lobby.UpdatedAt,
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
func (r *LobbyRepository) GetByUserID(ctx context.Context, userID uint64) ([]*models.Lobby, error) {
	query := `
		SELECT id, game_id, user_id, bet_amount, status, created_at, updated_at
		FROM lobbies
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
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
			&lobby.BetAmount,
			&lobby.Status,
			&lobby.CreatedAt,
			&lobby.UpdatedAt,
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
func (r *LobbyRepository) GetActive(ctx context.Context) ([]*models.Lobby, error) {
	query := `
		SELECT id, game_id, user_id, bet_amount, status, created_at, updated_at
		FROM lobbies
		WHERE status = 'active'
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
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
			&lobby.BetAmount,
			&lobby.Status,
			&lobby.CreatedAt,
			&lobby.UpdatedAt,
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
		SET game_id = $1, user_id = $2, bet_amount = $3, status = $4, updated_at = $5
		WHERE id = $6
	`

	lobby.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(
		ctx,
		query,
		lobby.GameID,
		lobby.UserID,
		lobby.BetAmount,
		lobby.Status,
		lobby.UpdatedAt,
		lobby.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update lobby: %w", err)
	}

	return nil
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
