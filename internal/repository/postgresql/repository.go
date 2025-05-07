package postgresql

import (
	"database/sql"

	"github.com/TakuroBreath/wordle/internal/models"
)

type Repository struct {
	db              *sql.DB
	gameRepo        models.GameRepository
	userRepo        models.UserRepository
	lobbyRepo       models.LobbyRepository
	attemptRepo     models.AttemptRepository
	historyRepo     models.HistoryRepository
	transactionRepo models.TransactionRepository
}

func NewRepository(db *sql.DB) *Repository {
	repo := &Repository{
		db: db,
	}

	// Инициализация репозиториев
	repo.gameRepo = NewGameRepository(db)
	repo.userRepo = NewUserRepository(db)
	// TODO: Инициализировать остальные репозитории
	// repo.lobbyRepo = NewLobbyRepository(db)
	// repo.attemptRepo = NewAttemptRepository(db)
	// repo.historyRepo = NewHistoryRepository(db)
	// repo.transactionRepo = NewTransactionRepository(db)

	return repo
}

// Game возвращает репозиторий для работы с играми
func (r *Repository) Game() models.GameRepository {
	return r.gameRepo
}

// User возвращает репозиторий для работы с пользователями
func (r *Repository) User() models.UserRepository {
	return r.userRepo
}

// Lobby возвращает репозиторий для работы с лобби
func (r *Repository) Lobby() models.LobbyRepository {
	return r.lobbyRepo
}

// Attempt возвращает репозиторий для работы с попытками
func (r *Repository) Attempt() models.AttemptRepository {
	return r.attemptRepo
}

// History возвращает репозиторий для работы с историей
func (r *Repository) History() models.HistoryRepository {
	return r.historyRepo
}

// Transaction возвращает репозиторий для работы с транзакциями
func (r *Repository) Transaction() models.TransactionRepository {
	return r.transactionRepo
}

// Close закрывает соединение с базой данных
func (r *Repository) Close() error {
	return r.db.Close()
}
