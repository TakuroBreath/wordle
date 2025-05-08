package postgresql

import (
	"database/sql"

	"github.com/TakuroBreath/wordle/internal/models"
)

// Repository представляет собой реализацию всех репозиториев
type Repository struct {
	db *sql.DB

	game        models.GameRepository
	user        models.UserRepository
	lobby       models.LobbyRepository
	attempt     models.AttemptRepository
	history     models.HistoryRepository
	transaction models.TransactionRepository
}

// NewRepository создает новый экземпляр Repository
func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		db: db,
	}
}

// Game возвращает репозиторий для работы с играми
func (r *Repository) Game() models.GameRepository {
	if r.game == nil {
		r.game = NewGameRepository(r.db)
	}
	return r.game
}

// User возвращает репозиторий для работы с пользователями
func (r *Repository) User() models.UserRepository {
	if r.user == nil {
		r.user = NewUserRepository(r.db)
	}
	return r.user
}

// Lobby возвращает репозиторий для работы с лобби
func (r *Repository) Lobby() models.LobbyRepository {
	if r.lobby == nil {
		r.lobby = NewLobbyRepository(r.db)
	}
	return r.lobby
}

// Attempt возвращает репозиторий для работы с попытками
func (r *Repository) Attempt() models.AttemptRepository {
	if r.attempt == nil {
		r.attempt = NewAttemptRepository(r.db)
	}
	return r.attempt
}

// History возвращает репозиторий для работы с историей
func (r *Repository) History() models.HistoryRepository {
	if r.history == nil {
		r.history = NewHistoryRepository(r.db)
	}
	return r.history
}

// Transaction возвращает репозиторий для работы с транзакциями
func (r *Repository) Transaction() models.TransactionRepository {
	if r.transaction == nil {
		r.transaction = NewTransactionRepository(r.db)
	}
	return r.transaction
}

// Close закрывает соединение с базой данных
func (r *Repository) Close() error {
	return r.db.Close()
}
