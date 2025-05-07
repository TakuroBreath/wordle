package service

import (
	"context"
	"errors"
	"time"

	"github.com/TakuroBreath/wordle/internal/models"
	"github.com/google/uuid"
)

// UserService представляет собой реализацию сервиса для работы с пользователями
type UserService struct {
	repo models.UserRepository
}

// NewUserService создает новый экземпляр UserService
func NewUserService(repo models.UserRepository) *UserService {
	return &UserService{
		repo: repo,
	}
}

// Create создает нового пользователя
func (s *UserService) Create(ctx context.Context, user *models.User) error {
	// Проверка валидности данных
	if user.TelegramID <= 0 {
		return errors.New("telegram ID must be greater than 0")
	}

	// Проверка, существует ли пользователь с таким Telegram ID
	existingUser, err := s.repo.GetByTelegramID(ctx, user.TelegramID)
	if err == nil && existingUser != nil {
		return errors.New("user with this Telegram ID already exists")
	}

	// Инициализация счетчиков побед и поражений
	user.Wins = 0
	user.Losses = 0

	// Инициализация баланса
	user.Balance = 0

	// Сохранение пользователя в базе данных
	return s.repo.Create(ctx, user)
}

// GetByID получает пользователя по ID
func (s *UserService) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	return s.repo.GetByID(ctx, id)
}

// GetByTelegramID получает пользователя по Telegram ID
func (s *UserService) GetByTelegramID(ctx context.Context, telegramID int64) (*models.User, error) {
	return s.repo.GetByTelegramID(ctx, telegramID)
}

// Update обновляет данные пользователя
func (s *UserService) Update(ctx context.Context, user *models.User) error {
	// Проверка существования пользователя
	existingUser, err := s.repo.GetByID(ctx, user.TelegramID)
	if err != nil {
		return errors.New("user not found")
	}

	// Проверка валидности данных
	if user.TelegramID <= 0 {
		return errors.New("telegram ID must be greater than 0")
	}

	// Проверка, не пытается ли пользователь изменить свой Telegram ID на уже существующий
	if user.TelegramID != existingUser.TelegramID {
		existingByTelegram, err := s.repo.GetByTelegramID(ctx, user.TelegramID)
		if err == nil && existingByTelegram != nil && existingByTelegram.ID != user.ID {
			return errors.New("user with this Telegram ID already exists")
		}
	}

	// Сохранение пользователя в базе данных
	return s.repo.Update(ctx, user)
}

// Delete удаляет пользователя
func (s *UserService) Delete(ctx context.Context, id uuid.UUID) error {
	// Проверка существования пользователя
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return errors.New("user not found")
	}

	// Удаление пользователя
	return s.repo.Delete(ctx, id)
}

// TransactionService представляет собой реализацию сервиса для работы с транзакциями
type TransactionService struct {
	repo     models.TransactionRepository
	userRepo models.UserRepository
}

// NewTransactionService создает новый экземпляр TransactionService
func NewTransactionService(repo models.TransactionRepository, userRepo models.UserRepository) *TransactionService {
	return &TransactionService{
		repo:     repo,
		userRepo: userRepo,
	}
}

// Create создает новую транзакцию
func (s *TransactionService) Create(ctx context.Context, tx *models.Transaction) error {
	// Проверка валидности данных
	if tx.UserID == uuid.Nil {
		return errors.New("user ID cannot be empty")
	}

	if tx.Amount <= 0 {
		return errors.New("amount must be greater than 0")
	}

	if tx.Currency != "TON" && tx.Currency != "USDT" {
		return errors.New("currency must be TON or USDT")
	}

	if tx.Type != "deposit" && tx.Type != "withdraw" && tx.Type != "win" && tx.Type != "loss" {
		return errors.New("invalid transaction type")
	}

	if tx.Status != "pending" && tx.Status != "completed" && tx.Status != "failed" {
		return errors.New("invalid transaction status")
	}

	// Генерация ID, если он не был установлен
	if tx.ID == uuid.Nil {
		tx.ID = uuid.New()
	}

	// Установка текущего времени, если оно не было установлено
	if tx.CreatedAt.IsZero() {
		tx.CreatedAt = time.Now()
	}

	// Сохранение транзакции в базе данных
	return s.repo.Create(ctx, tx)
}

// GetByID получает транзакцию по ID
func (s *TransactionService) GetByID(ctx context.Context, id uuid.UUID) (*models.Transaction, error) {
	return s.repo.GetByID(ctx, id)
}

// GetByUserID получает транзакции пользователя
func (s *TransactionService) GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.Transaction, error) {
	return s.repo.GetByUserID(ctx, userID, limit, offset)
}

// GetByStatus получает транзакции по статусу
func (s *TransactionService) GetByStatus(ctx context.Context, status string, limit, offset int) ([]*models.Transaction, error) {
	return s.repo.GetByStatus(ctx, status, limit, offset)
}

// Update обновляет данные транзакции
func (s *TransactionService) Update(ctx context.Context, tx *models.Transaction) error {
	// Проверка существования транзакции
	existingTx, err := s.repo.GetByID(ctx, tx.ID)
	if err != nil {
		return errors.New("transaction not found")
	}

	// Проверка, что транзакция не завершена
	if existingTx.Status == "completed" || existingTx.Status == "failed" {
		return errors.New("cannot update completed or failed transaction")
	}

	// Проверка валидности данных
	if tx.Status != "pending" && tx.Status != "completed" && tx.Status != "failed" {
		return errors.New("invalid transaction status")
	}

	// Если транзакция завершена, устанавливаем время завершения
	if tx.Status == "completed" || tx.Status == "failed" {
		now := time.Now()
		tx.CompletedAt = &now
	}

	// Сохранение транзакции в базе данных
	return s.repo.Update(ctx, tx)
}

// ProcessDeposit обрабатывает депозит средств
func (s *TransactionService) ProcessDeposit(ctx context.Context, userID uuid.UUID, amount float64, currency string, txHash string) error {
	// Проверка валидности данных
	if userID == uuid.Nil {
		return errors.New("user ID cannot be empty")
	}

	if amount <= 0 {
		return errors.New("amount must be greater than 0")
	}

	if currency != "TON" && currency != "USDT" {
		return errors.New("currency must be TON or USDT")
	}

	if txHash == "" {
		return errors.New("transaction hash cannot be empty")
	}

	// Проверка существования пользователя
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return errors.New("user not found")
	}

	// Создание транзакции
	tx := &models.Transaction{
		ID:        uuid.New(),
		UserID:    userID,
		Amount:    amount,
		Currency:  currency,
		Type:      "deposit",
		Status:    "pending",
		TxHash:    txHash,
		CreatedAt: time.Now(),
	}

	// Сохранение транзакции
	err = s.repo.Create(ctx, tx)
	if err != nil {
		return err
	}

	// Обновление баланса пользователя
	user.Balance += amount
	err = s.userRepo.Update(ctx, user)
	if err != nil {
		return err
	}

	// Обновление статуса транзакции
	tx.Status = "completed"
	now := time.Now()
	tx.CompletedAt = &now
	return s.repo.Update(ctx, tx)
}

// ProcessWithdraw обрабатывает вывод средств
func (s *TransactionService) ProcessWithdraw(ctx context.Context, userID uuid.UUID, amount float64, currency string) (*models.Transaction, error) {
	// Проверка валидности данных
	if userID == uuid.Nil {
		return nil, errors.New("user ID cannot be empty")
	}

	if amount <= 0 {
		return nil, errors.New("amount must be greater than 0")
	}

	if currency != "TON" && currency != "USDT" {
		return nil, errors.New("currency must be TON or USDT")
	}

	// Проверка существования пользователя
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// Проверка достаточности средств
	if user.Balance < amount {
		return nil, errors.New("insufficient funds")
	}

	// Создание транзакции
	tx := &models.Transaction{
		ID:        uuid.New(),
		UserID:    userID,
		Amount:    amount,
		Currency:  currency,
		Type:      "withdraw",
		Status:    "pending",
		CreatedAt: time.Now(),
	}

	// Сохранение транзакции
	err = s.repo.Create(ctx, tx)
	if err != nil {
		return nil, err
	}

	// Обновление баланса пользователя
	user.Balance -= amount
	err = s.userRepo.Update(ctx, user)
	if err != nil {
		return nil, err
	}

	return tx, nil
}
