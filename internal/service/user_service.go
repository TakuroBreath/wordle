package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/TakuroBreath/wordle/internal/logger"
	"github.com/TakuroBreath/wordle/internal/models"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// UserServiceImpl представляет собой реализацию сервиса для работы с пользователями
type UserServiceImpl struct {
	repo               models.UserRepository
	transactionRepo    models.TransactionRepository
	tonService         models.TONService
	minWithdrawTON     float64
	minWithdrawUSDT    float64
	withdrawFeeTON     float64
	withdrawFeeUSDT    float64
	withdrawLockMinutes int
	logger             *zap.Logger
}

// UserServiceConfig конфигурация сервиса пользователей
type UserServiceConfig struct {
	MinWithdrawTON      float64
	MinWithdrawUSDT     float64
	WithdrawFeeTON      float64
	WithdrawFeeUSDT     float64
	WithdrawLockMinutes int
}

// NewUserService создает новый экземпляр UserService
func NewUserService(
	repo models.UserRepository,
	transactionRepo models.TransactionRepository,
	tonService models.TONService,
	config UserServiceConfig,
) models.UserService {
	return &UserServiceImpl{
		repo:                repo,
		transactionRepo:     transactionRepo,
		tonService:          tonService,
		minWithdrawTON:      config.MinWithdrawTON,
		minWithdrawUSDT:     config.MinWithdrawUSDT,
		withdrawFeeTON:      config.WithdrawFeeTON,
		withdrawFeeUSDT:     config.WithdrawFeeUSDT,
		withdrawLockMinutes: config.WithdrawLockMinutes,
		logger:              logger.GetLogger(zap.String("service", "user")),
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
	user.PendingWithdrawal = 0.0
	user.TotalDeposited = 0.0
	user.TotalWithdrawn = 0.0
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

	existingUser, err := s.repo.GetByTelegramID(ctx, user.TelegramID)
	if err != nil {
		if errors.Is(err, models.ErrUserNotFound) {
			return errors.New("user not found")
		}
		return fmt.Errorf("error fetching user for update: %w", err)
	}

	existingUser.Username = user.Username
	existingUser.FirstName = user.FirstName
	existingUser.LastName = user.LastName
	existingUser.UpdatedAt = time.Now()

	return s.repo.Update(ctx, existingUser)
}

// DeleteUser удаляет пользователя по Telegram ID
func (s *UserServiceImpl) DeleteUser(ctx context.Context, telegramID uint64) error {
	if telegramID == 0 {
		return errors.New("telegram ID must be valid")
	}

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

// UpdateWallet обновляет адрес кошелька пользователя
func (s *UserServiceImpl) UpdateWallet(ctx context.Context, telegramID uint64, wallet string) error {
	log := s.logger.With(zap.String("method", "UpdateWallet"), zap.Uint64("user_id", telegramID))

	if telegramID == 0 {
		return errors.New("telegram ID must be valid")
	}
	if wallet == "" {
		return errors.New("wallet address cannot be empty")
	}

	// Валидируем адрес
	if s.tonService != nil && !s.tonService.ValidateAddress(wallet) {
		return errors.New("invalid wallet address")
	}

	// Проверяем, не занят ли кошелёк
	existingUser, err := s.repo.GetByWallet(ctx, wallet)
	if err == nil && existingUser != nil && existingUser.TelegramID != telegramID {
		return errors.New("wallet already linked to another user")
	}

	log.Info("Updating wallet", zap.String("wallet", wallet))
	return s.repo.UpdateWallet(ctx, telegramID, wallet)
}

// GetUserByWallet получает пользователя по адресу кошелька
func (s *UserServiceImpl) GetUserByWallet(ctx context.Context, wallet string) (*models.User, error) {
	if wallet == "" {
		return nil, errors.New("wallet cannot be empty")
	}
	return s.repo.GetByWallet(ctx, wallet)
}

// UpdateTonBalance обновляет баланс пользователя в TON
func (s *UserServiceImpl) UpdateTonBalance(ctx context.Context, telegramID uint64, amount float64) error {
	if telegramID == 0 {
		return errors.New("telegram ID must be valid")
	}
	return s.repo.UpdateTonBalance(ctx, telegramID, amount)
}

// UpdateUsdtBalance обновляет баланс пользователя в USDT
func (s *UserServiceImpl) UpdateUsdtBalance(ctx context.Context, telegramID uint64, amount float64) error {
	if telegramID == 0 {
		return errors.New("telegram ID must be valid")
	}
	return s.repo.UpdateUsdtBalance(ctx, telegramID, amount)
}

// GetTopUsers получает список топ-пользователей
func (s *UserServiceImpl) GetTopUsers(ctx context.Context, limit int) ([]*models.User, error) {
	if limit <= 0 {
		limit = 10
	}
	return s.repo.GetTopUsers(ctx, limit)
}

// GetUserStats получает статистику пользователя
func (s *UserServiceImpl) GetUserStats(ctx context.Context, telegramID uint64) (map[string]any, error) {
	if telegramID == 0 {
		return nil, errors.New("telegram ID must be valid")
	}

	user, err := s.repo.GetByTelegramID(ctx, telegramID)
	if err != nil {
		return nil, err
	}

	stats := map[string]any{
		"telegram_id":        user.TelegramID,
		"username":           user.Username,
		"first_name":         user.FirstName,
		"last_name":          user.LastName,
		"wallet":             user.Wallet,
		"wins":               user.Wins,
		"losses":             user.Losses,
		"balance_ton":        user.BalanceTon,
		"balance_usdt":       user.BalanceUsdt,
		"pending_withdrawal": user.PendingWithdrawal,
		"total_deposited":    user.TotalDeposited,
		"total_withdrawn":    user.TotalWithdrawn,
	}

	if user.Wins+user.Losses > 0 {
		stats["win_rate"] = float64(user.Wins) / float64(user.Wins+user.Losses)
	} else {
		stats["win_rate"] = 0.0
	}

	return stats, nil
}

// IncrementWins увеличивает количество побед
func (s *UserServiceImpl) IncrementWins(ctx context.Context, telegramID uint64) error {
	return s.repo.IncrementWins(ctx, telegramID)
}

// IncrementLosses увеличивает количество поражений
func (s *UserServiceImpl) IncrementLosses(ctx context.Context, telegramID uint64) error {
	return s.repo.IncrementLosses(ctx, telegramID)
}

// ValidateBalance проверяет, достаточен ли баланс пользователя
func (s *UserServiceImpl) ValidateBalance(ctx context.Context, telegramID uint64, requiredAmount float64, currency string) (bool, error) {
	if telegramID == 0 {
		return false, errors.New("telegram ID must be valid")
	}
	if requiredAmount < 0 {
		return false, errors.New("required amount cannot be negative")
	}
	if requiredAmount == 0 {
		return true, nil
	}

	user, err := s.repo.GetByTelegramID(ctx, telegramID)
	if err != nil {
		return false, err
	}

	switch currency {
	case models.CurrencyTON:
		return user.GetAvailableBalance(models.CurrencyTON) >= requiredAmount, nil
	case models.CurrencyUSDT:
		return user.GetAvailableBalance(models.CurrencyUSDT) >= requiredAmount, nil
	default:
		return false, fmt.Errorf("unknown currency: %s", currency)
	}
}

// CanWithdraw проверяет, может ли пользователь сделать вывод
func (s *UserServiceImpl) CanWithdraw(ctx context.Context, telegramID uint64) (bool, error) {
	user, err := s.repo.GetByTelegramID(ctx, telegramID)
	if err != nil {
		return false, err
	}
	return user.CanWithdraw(), nil
}

// RequestWithdraw запрашивает вывод средств
func (s *UserServiceImpl) RequestWithdraw(ctx context.Context, telegramID uint64, amount float64, currency string, toAddress string) (*models.WithdrawResult, error) {
	log := s.logger.With(zap.String("method", "RequestWithdraw"),
		zap.Uint64("user_id", telegramID),
		zap.Float64("amount", amount),
		zap.String("currency", currency))

	if telegramID == 0 {
		return nil, errors.New("telegram ID must be valid")
	}
	if amount <= 0 {
		return nil, errors.New("withdraw amount must be positive")
	}
	if currency != models.CurrencyTON && currency != models.CurrencyUSDT {
		return nil, fmt.Errorf("unsupported currency: %s", currency)
	}

	// Определяем минимальную сумму и комиссию
	var minWithdraw, fee float64
	if currency == models.CurrencyTON {
		minWithdraw = s.minWithdrawTON
		fee = s.withdrawFeeTON
	} else {
		minWithdraw = s.minWithdrawUSDT
		fee = s.withdrawFeeUSDT
	}

	if amount < minWithdraw {
		return nil, fmt.Errorf("minimum withdrawal is %.4f %s", minWithdraw, currency)
	}

	// Получаем пользователя
	user, err := s.repo.GetByTelegramID(ctx, telegramID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Проверяем возможность вывода
	if !user.CanWithdraw() {
		return nil, errors.New("withdrawal is temporarily locked, please wait")
	}

	// Если адрес не передан, используем привязанный кошелёк
	if toAddress == "" {
		toAddress = user.Wallet
	}
	if toAddress == "" {
		return nil, errors.New("wallet address not set")
	}

	// Валидируем адрес
	if s.tonService != nil && !s.tonService.ValidateAddress(toAddress) {
		return nil, errors.New("invalid wallet address")
	}

	// Проверяем баланс (с учётом комиссии)
	totalRequired := amount
	hasBalance, err := s.ValidateBalance(ctx, telegramID, totalRequired, currency)
	if err != nil {
		return nil, fmt.Errorf("failed to validate balance: %w", err)
	}
	if !hasBalance {
		return nil, fmt.Errorf("insufficient %s balance", currency)
	}

	// Проверяем, нет ли уже pending вывода
	if user.PendingWithdrawal > 0 {
		return nil, errors.New("there is already a pending withdrawal")
	}

	// Начинаем транзакцию (упрощённо без реальной БД транзакции)
	log.Info("Processing withdrawal request",
		zap.Float64("amount", amount),
		zap.Float64("fee", fee),
		zap.String("to_address", toAddress))

	// Списываем с баланса
	if currency == models.CurrencyTON {
		if err := s.repo.UpdateTonBalance(ctx, telegramID, -amount); err != nil {
			return nil, fmt.Errorf("failed to deduct balance: %w", err)
		}
	} else {
		if err := s.repo.UpdateUsdtBalance(ctx, telegramID, -amount); err != nil {
			return nil, fmt.Errorf("failed to deduct balance: %w", err)
		}
	}

	// Устанавливаем pending withdrawal и lock
	if err := s.repo.UpdatePendingWithdrawal(ctx, telegramID, amount); err != nil {
		// Откатываем
		if currency == models.CurrencyTON {
			_ = s.repo.UpdateTonBalance(ctx, telegramID, amount)
		} else {
			_ = s.repo.UpdateUsdtBalance(ctx, telegramID, amount)
		}
		return nil, fmt.Errorf("failed to set pending withdrawal: %w", err)
	}

	// Устанавливаем блокировку на вывод
	lockUntil := time.Now().Add(time.Duration(s.withdrawLockMinutes) * time.Minute)
	if err := s.repo.SetWithdrawalLock(ctx, telegramID, lockUntil); err != nil {
		log.Warn("Failed to set withdrawal lock", zap.Error(err))
	}

	// Создаём транзакцию
	tx := &models.Transaction{
		ID:          uuid.New(),
		UserID:      telegramID,
		Type:        models.TransactionTypeWithdraw,
		Amount:      amount,
		Fee:         fee,
		Currency:    currency,
		Status:      models.TransactionStatusPending,
		ToAddress:   toAddress,
		Description: fmt.Sprintf("Withdraw %.4f %s to %s", amount-fee, currency, toAddress),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.transactionRepo.Create(ctx, tx); err != nil {
		// Откатываем
		_ = s.repo.UpdatePendingWithdrawal(ctx, telegramID, -amount)
		if currency == models.CurrencyTON {
			_ = s.repo.UpdateTonBalance(ctx, telegramID, amount)
		} else {
			_ = s.repo.UpdateUsdtBalance(ctx, telegramID, amount)
		}
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	log.Info("Withdrawal request created",
		zap.String("transaction_id", tx.ID.String()))

	return &models.WithdrawResult{
		TransactionID: tx.ID,
		Status:        models.TransactionStatusPending,
		Fee:           fee,
	}, nil
}

// GetWithdrawHistory получает историю выводов средств пользователя
func (s *UserServiceImpl) GetWithdrawHistory(ctx context.Context, telegramID uint64, limit, offset int) ([]*models.Transaction, error) {
	if telegramID == 0 {
		return nil, errors.New("telegram ID must be valid")
	}

	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	// Получаем транзакции типа withdraw
	transactions, err := s.transactionRepo.GetByUserID(ctx, telegramID, 1000, 0)
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

// Deprecated: для обратной совместимости
func NewUserServiceImpl(repo models.UserRepository, ts models.TransactionService) models.UserService {
	return &UserServiceImpl{
		repo:                repo,
		minWithdrawTON:      0.1,
		minWithdrawUSDT:     1.0,
		withdrawFeeTON:      0.05,
		withdrawFeeUSDT:     0.5,
		withdrawLockMinutes: 5,
		logger:              logger.GetLogger(zap.String("service", "user")),
	}
}
