package postgresql

import (
	"database/sql"
	"fmt"

	"github.com/TakuroBreath/wordle/internal/config"
	_ "github.com/lib/pq"
)

// NewConnection создает новое соединение с PostgreSQL
func NewConnection(cfg config.PostgresConfig) (*sql.DB, error) {
	// Используем DSN из конфигурации
	dsn := cfg.DSN()

	// Открываем соединение
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Проверяем соединение
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}
