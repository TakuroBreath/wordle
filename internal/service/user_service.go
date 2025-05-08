package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/TakuroBreath/wordle/internal/models"
	// "github.com/google/uuid" // uuid не используется для User ID
)

// UserServiceImpl представляет собой реализацию сервиса для работы с пользователями
type UserServiceImpl struct {
	repo               models.UserRepository
	transactionService models.TransactionService
}

// NewUserServiceImpl создает новый экземпляр UserServiceImpl
func NewUserServiceImpl(repo models.UserRepository, ts models.TransactionService) models.UserService {
	return &UserServiceImpl{
		repo:               repo,
		transactionService: ts,
	}
}

// CreateUser создает нового пользователя
func (s *UserServiceImpl) CreateUser(ctx context.Context, user *models.User) error {
	if user == nil {
		return errors.New("user cannot be nil")
	}
	if user.TelegramID == 0 {
		return errors.New("telegram ID must be valid")
	}

	existingUser, err := s.repo.GetByTelegramID(ctx, user.TelegramID)
	if err != nil && !errors.Is(err, models.ErrUserNotFound) {
		return fmt.Errorf("error checking for existing user: %w", err)
	}
	if existingUser != nil {
		return errors.New("user with this Telegram ID already exists")
	}

	user.Wins = 0
	user.Losses = 0
	user.BalanceTon = 0.0
	user.BalanceUsdt = 0.0
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	return s.repo.Create(ctx, user)
}

// GetUser получает пользователя по Telegram ID
func (s *UserServiceImpl) GetUser(ctx context.Context, telegramID uint64) (*models.User, error) {
	if telegramID == 0 {
		return nil, errors.New("telegram ID must be valid")
	}
	return s.repo.GetByTelegramID(ctx, telegramID)
}

// UpdateUser обновляет данные пользователя
func (s *UserServiceImpl) UpdateUser(ctx context.Context, user *models.User) error {
	if user == nil {
		return errors.New("user cannot be nil")
	}
	if user.TelegramID == 0 {
		return errors.New("telegram ID must be valid")
	}

	// Получаем существующего пользователя, чтобы убедиться, что он есть
	// и чтобы не обновить случайно несуществующего или не изменить критичные поля
	existingUser, err := s.repo.GetByTelegramID(ctx, user.TelegramID)
	if err != nil {
		if errors.Is(err, models.ErrUserNotFound) {
			return errors.New("user not found")
		}
		return fmt.Errorf("error fetching user for update: %w", err)
	}

	// Обновляем только разрешенные поля. Не позволяем менять TelegramID, балансы напрямую этим методом.
	existingUser.Username = user.Username
	existingUser.FirstName = user.FirstName
	existingUser.LastName = user.LastName
	existingUser.Wallet = user.Wallet // Разрешаем обновление кошелька
	existingUser.UpdatedAt = time.Now()

	return s.repo.Update(ctx, existingUser)
}

// DeleteUser удаляет пользователя по Telegram ID
func (s *UserServiceImpl) DeleteUser(ctx context.Context, telegramID uint64) error {
	if telegramID == 0 {
		return errors.New("telegram ID must be valid")
	}
	// Сначала проверим, существует ли пользователь
	_, err := s.repo.GetByTelegramID(ctx, telegramID)
	if err != nil {
		if errors.Is(err, models.ErrUserNotFound) {
			return errors.New("user not found")
		}
		return fmt.Errorf("error fetching user for delete: %w", err)
	}
	return s.repo.Delete(ctx, telegramID)
}

// GetUserByUsername получает пользователя по Username
func (s *UserServiceImpl) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	if username == "" {
		return nil, errors.New("username cannot be empty")
	}
	return s.repo.GetByUsername(ctx, username)
}

// UpdateTonBalance обновляет баланс пользователя в TON.
// amount может быть положительным (пополнение) или отрицательным (списание).
func (s *UserServiceImpl) UpdateTonBalance(ctx context.Context, telegramID uint64, amount float64) error {
	if telegramID == 0 {
		return errors.New("telegram ID must be valid")
	}
	// Можно добавить проверку, что amount не приведет к отрицательному балансу, если это бизнес-требование
	// или делегировать это репозиторию/базе данных
	return s.repo.UpdateTonBalance(ctx, telegramID, amount)
}

// UpdateUsdtBalance обновляет баланс пользователя в USDT.
// amount может быть положительным (пополнение) или отрицательным (списание).
func (s *UserServiceImpl) UpdateUsdtBalance(ctx context.Context, telegramID uint64, amount float64) error {
	if telegramID == 0 {
		return errors.New("telegram ID must be valid")
	}
	return s.repo.UpdateUsdtBalance(ctx, telegramID, amount)
}

// GetTopUsers получает список топ-пользователей (например, по количеству побед)
func (s *UserServiceImpl) GetTopUsers(ctx context.Context, limit int) ([]*models.User, error) {
	if limit <= 0 {
		limit = 10 // Значение по умолчанию
	}
	return s.repo.GetTopUsers(ctx, limit)
}

// GetUserStats получает статистику пользователя
func (s *UserServiceImpl) GetUserStats(ctx context.Context, telegramID uint64) (map[string]interface{}, error) {
	if telegramID == 0 {
		return nil, errors.New("telegram ID must be valid")
	}
	// Этот метод может быть реализован либо прямым вызовом s.repo.GetUserStats,
	// либо сбором данных из user и, возможно, других репозиториев (например, historyRepo)
	user, err := s.repo.GetByTelegramID(ctx, telegramID)
	if err != nil {
		return nil, err
	}
	stats := map[string]interface{}{
		"telegram_id":  user.TelegramID,
		"username":     user.Username,
		"first_name":   user.FirstName,
		"last_name":    user.LastName,
		"wallet":       user.Wallet,
		"wins":         user.Wins,
		"losses":       user.Losses,
		"balance_ton":  user.BalanceTon,
		"balance_usdt": user.BalanceUsdt,
		// Можно добавить win_rate, если это считается на уровне сервиса
	}
	if user.Wins+user.Losses > 0 {
		stats["win_rate"] = float64(user.Wins) / float64(user.Wins+user.Losses)
	} else {
		stats["win_rate"] = 0.0
	}
	return stats, nil
}

// ValidateBalance проверяет, достаточен ли баланс пользователя для указанной суммы
func (s *UserServiceImpl) ValidateBalance(ctx context.Context, telegramID uint64, requiredAmount float64, currency string) (bool, error) {
	if telegramID == 0 {
		return false, errors.New("telegram ID must be valid")
	}
	if requiredAmount < 0 {
		// Снятие отрицательной суммы не имеет смысла для проверки баланса
		return false, errors.New("required amount cannot be negative")
	}
	if requiredAmount == 0 {
		return true, nil // Для нулевой суммы баланс всегда достаточен
	}

	user, err := s.repo.GetByTelegramID(ctx, telegramID)
	if err != nil {
		return false, err
	}

	switch currency {
	case models.CurrencyTON:
		return user.BalanceTon >= requiredAmount, nil
	case models.CurrencyUSDT:
		return user.BalanceUsdt >= requiredAmount, nil
	default:
		return false, fmt.Errorf("unknown currency: %s", currency)
	}
}

// RequestWithdraw запрашивает вывод средств с баланса пользователя
func (s *UserServiceImpl) RequestWithdraw(ctx context.Context, telegramID uint64, amount float64, currency string) error {
	if telegramID == 0 {
		return errors.New("telegram ID must be valid")
	}
	if amount <= 0 {
		return errors.New("withdraw amount must be positive")
	}

	// Проверяем поддерживаемую валюту
	if currency != models.CurrencyTON && currency != models.CurrencyUSDT {
		return fmt.Errorf("unsupported currency: %s", currency)
	}

	// Проверяем, достаточно ли средств на балансе
	hasBalance, err := s.ValidateBalance(ctx, telegramID, amount, currency)
	if err != nil {
		return fmt.Errorf("failed to validate balance: %w", err)
	}
	if !hasBalance {
		return fmt.Errorf("insufficient %s balance", currency)
	}

	// Получаем пользователя для проверки адреса кошелька
	user, err := s.repo.GetByTelegramID(ctx, telegramID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	if user.Wallet == "" {
		return errors.New("wallet address not set")
	}

	// Создаем транзакцию вывода средств через transactionService
	tx := &models.Transaction{
		UserID:      telegramID,
		Type:        models.TransactionTypeWithdraw,
		Amount:      amount,
		Currency:    currency,
		Status:      models.TransactionStatusPending,
		Description: fmt.Sprintf("Withdraw %f %s to wallet %s", amount, currency, user.Wallet),
	}

	if err := s.transactionService.CreateTransaction(ctx, tx); err != nil {
		return fmt.Errorf("failed to create withdraw transaction: %w", err)
	}

	// Уменьшаем баланс пользователя (блокируем средства)
	if currency == models.CurrencyTON {
		if err := s.UpdateTonBalance(ctx, telegramID, -amount); err != nil {
			// Если не удалось обновить баланс, отменяем транзакцию
			if txErr := s.transactionService.FailTransaction(ctx, tx.ID, "Failed to update user balance"); txErr != nil {
				// Логируем ошибку, но возвращаем первоначальную
				fmt.Printf("ERROR: Failed to mark transaction %s as failed: %v\n", tx.ID, txErr)
			}
			return fmt.Errorf("failed to update TON balance: %w", err)
		}
	} else {
		if err := s.UpdateUsdtBalance(ctx, telegramID, -amount); err != nil {
			// Если не удалось обновить баланс, отменяем транзакцию
			if txErr := s.transactionService.FailTransaction(ctx, tx.ID, "Failed to update user balance"); txErr != nil {
				fmt.Printf("ERROR: Failed to mark transaction %s as failed: %v\n", tx.ID, txErr)
			}
			return fmt.Errorf("failed to update USDT balance: %w", err)
		}
	}

	return nil
}

// GetWithdrawHistory получает историю выводов средств пользователя
func (s *UserServiceImpl) GetWithdrawHistory(ctx context.Context, telegramID uint64, limit, offset int) ([]*models.Transaction, error) {
	if telegramID == 0 {
		return nil, errors.New("telegram ID must be valid")
	}

	if limit <= 0 {
		limit = 10 // Значение по умолчанию
	}
	if offset < 0 {
		offset = 0
	}

	// Получаем все транзакции пользователя
	transactions, err := s.transactionService.GetUserTransactions(ctx, telegramID, 1000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get user transactions: %w", err)
	}

	// Фильтруем по типу "withdraw"
	var withdrawals []*models.Transaction
	for _, tx := range transactions {
		if tx.Type == models.TransactionTypeWithdraw {
			withdrawals = append(withdrawals, tx)
		}
	}

	// Применяем пагинацию
	if offset >= len(withdrawals) {
		return []*models.Transaction{}, nil
	}
	end := offset + limit
	if end > len(withdrawals) {
		end = len(withdrawals)
	}

	return withdrawals[offset:end], nil
}
