package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/TakuroBreath/wordle/internal/models"
	"github.com/google/uuid"
)

// TransactionServiceImpl представляет собой реализацию TransactionService
type TransactionServiceImpl struct {
	transactionRepo models.TransactionRepository
	userRepo        models.UserRepository
	// Если потребуется оповещать другие сервисы или выполнять доп. действия,
	// можно добавить сюда другие репозитории или сервисы
}

// NewTransactionServiceImpl создает новый экземпляр TransactionServiceImpl
func NewTransactionServiceImpl(
	transactionRepo models.TransactionRepository,
	userRepo models.UserRepository,
) models.TransactionService { // Убедитесь, что models.TransactionService определен
	return &TransactionServiceImpl{
		transactionRepo: transactionRepo,
		userRepo:        userRepo,
	}
}

// CreateTransaction создает новую транзакцию
func (s *TransactionServiceImpl) CreateTransaction(ctx context.Context, tx *models.Transaction) error {
	if tx == nil {
		return errors.New("transaction cannot be nil")
	}
	if tx.UserID == 0 { // В модели Transaction UserID это uint64 (TelegramID)
		return errors.New("user ID (TelegramID) in transaction cannot be zero")
	}
	// Проверка существования пользователя
	_, err := s.userRepo.GetByTelegramID(ctx, tx.UserID)
	if err != nil {
		if errors.Is(err, models.ErrUserNotFound) {
			return fmt.Errorf("user with TelegramID %d not found for transaction: %w", tx.UserID, err)
		}
		return fmt.Errorf("error fetching user for transaction: %w", err)
	}

	if tx.Amount <= 0 {
		return errors.New("transaction amount must be greater than 0")
	}

	// Валидация Currency не нужна, если она определяется на уровне Game или бизнес-логики платежа
	// if tx.Currency != "TON" && tx.Currency != "USDT" { // Поле Currency есть в модели Transaction
	// 	return errors.New("currency must be TON or USDT")
	// }

	allowedTypes := map[string]bool{
		models.TransactionTypeDeposit:    true,
		models.TransactionTypeWithdraw:   true,
		models.TransactionTypeBet:        true,
		models.TransactionTypeReward:     true,
		models.TransactionTypeCommission: true, // Если есть комиссия
		models.TransactionTypeRefund:     true, // Если есть возвраты
	}
	if !allowedTypes[tx.Type] {
		return fmt.Errorf("invalid transaction type: %s", tx.Type)
	}

	allowedStatuses := map[string]bool{
		models.TransactionStatusPending:   true,
		models.TransactionStatusCompleted: true,
		models.TransactionStatusFailed:    true,
		models.TransactionStatusCanceled:  true, // Если есть отмены
	}
	if !allowedStatuses[tx.Status] {
		return fmt.Errorf("invalid transaction status: %s", tx.Status)
	}

	if tx.ID == uuid.Nil {
		tx.ID = uuid.New()
	}
	now := time.Now()
	if tx.CreatedAt.IsZero() {
		tx.CreatedAt = now
	}
	tx.UpdatedAt = now

	return s.transactionRepo.Create(ctx, tx)
}

// GetTransaction получает транзакцию по ID
func (s *TransactionServiceImpl) GetTransaction(ctx context.Context, id uuid.UUID) (*models.Transaction, error) {
	if id == uuid.Nil {
		return nil, errors.New("transaction ID cannot be nil")
	}
	return s.transactionRepo.GetByID(ctx, id)
}

// GetUserTransactions получает транзакции пользователя
func (s *TransactionServiceImpl) GetUserTransactions(ctx context.Context, userID uint64, limit, offset int) ([]*models.Transaction, error) {
	if userID == 0 {
		return nil, errors.New("user ID (TelegramID) cannot be zero")
	}
	return s.transactionRepo.GetByUserID(ctx, userID, limit, offset)
}

// UpdateTransaction обновляет транзакцию (в основном статус)
func (s *TransactionServiceImpl) UpdateTransaction(ctx context.Context, tx *models.Transaction) error {
	if tx == nil {
		return errors.New("transaction cannot be nil")
	}
	if tx.ID == uuid.Nil {
		return errors.New("transaction ID cannot be nil for update")
	}

	existingTx, err := s.transactionRepo.GetByID(ctx, tx.ID)
	if err != nil {
		return fmt.Errorf("failed to get transaction for update: %w", err)
	}

	// Проверка, что не пытаемся изменить уже завершенную или проваленную транзакцию на что-то другое, кроме как на Canceled (если это разрешено)
	if (existingTx.Status == models.TransactionStatusCompleted || existingTx.Status == models.TransactionStatusFailed) &&
		(tx.Status != existingTx.Status && tx.Status != models.TransactionStatusCanceled) {
		return fmt.Errorf("cannot update already processed transaction status from %s to %s", existingTx.Status, tx.Status)
	}

	allowedStatuses := map[string]bool{
		models.TransactionStatusPending:   true,
		models.TransactionStatusCompleted: true,
		models.TransactionStatusFailed:    true,
		models.TransactionStatusCanceled:  true,
	}
	if !allowedStatuses[tx.Status] {
		return fmt.Errorf("invalid transaction status for update: %s", tx.Status)
	}

	tx.UpdatedAt = time.Now()
	return s.transactionRepo.Update(ctx, tx)
}

// DeleteTransaction удаляет транзакцию (использовать с осторожностью)
func (s *TransactionServiceImpl) DeleteTransaction(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return errors.New("transaction ID cannot be nil for delete")
	}
	// Можно добавить проверку, можно ли удалять транзакцию в текущем статусе
	return s.transactionRepo.Delete(ctx, id)
}

// GetTransactionsByType получает транзакции по типу
func (s *TransactionServiceImpl) GetTransactionsByType(ctx context.Context, transactionType string, limit, offset int) ([]*models.Transaction, error) {
	allowedTypes := map[string]bool{ // Перепроверить с константами в models
		models.TransactionTypeDeposit:  true,
		models.TransactionTypeWithdraw: true,
		models.TransactionTypeBet:      true,
		models.TransactionTypeReward:   true,
	}
	if !allowedTypes[transactionType] {
		return nil, fmt.Errorf("invalid transaction type for query: %s", transactionType)
	}
	return s.transactionRepo.GetByType(ctx, transactionType, limit, offset)
}

// GetUserBalance получает баланс пользователя.
// Для конкретной валюты лучше иметь отдельные методы или передавать параметр валюты.
func (s *TransactionServiceImpl) GetUserBalance(ctx context.Context, userID uint64 /*, currency string*/) (float64, error) {
	if userID == 0 {
		return 0, errors.New("user ID (TelegramID) cannot be zero")
	}
	user, err := s.userRepo.GetByTelegramID(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to get user for balance: %w", err)
	}
	// По умолчанию вернем баланс в TON, как было в прошлой версии.
	// В идеале, метод должен либо принимать валюту, либо возвращать оба баланса.
	return user.BalanceTon, nil
}

// GetTransactionStats получает статистику транзакций пользователя
func (s *TransactionServiceImpl) GetTransactionStats(ctx context.Context, userID uint64) (map[string]interface{}, error) {
	if userID == 0 {
		return nil, errors.New("user ID (TelegramID) cannot be zero")
	}
	user, err := s.userRepo.GetByTelegramID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user for stats: %w", err)
	}

	transactions, err := s.transactionRepo.GetByUserID(ctx, userID, 0, 0) // 0, 0 для получения всех транзакций
	if err != nil {
		return nil, fmt.Errorf("failed to get user transactions for stats: %w", err)
	}

	stats := make(map[string]interface{})
	stats["balance_ton"] = user.BalanceTon
	stats["balance_usdt"] = user.BalanceUsdt
	stats["total_transactions"] = len(transactions)

	totalDepositsTon := 0.0
	totalDepositsUsdt := 0.0
	countDeposits := 0

	totalWithdrawalsTon := 0.0
	totalWithdrawalsUsdt := 0.0
	countWithdrawals := 0

	totalBetsTon := 0.0
	totalBetsUsdt := 0.0
	countBets := 0

	totalRewardsTon := 0.0
	totalRewardsUsdt := 0.0
	countRewards := 0

	for _, tx := range transactions {
		if tx.Status == models.TransactionStatusCompleted { // Учитываем только завершенные транзакции для сумм
			switch tx.Type {
			case models.TransactionTypeDeposit:
				countDeposits++
				if tx.Currency == models.CurrencyTON {
					totalDepositsTon += tx.Amount
				} else if tx.Currency == models.CurrencyUSDT {
					totalDepositsUsdt += tx.Amount
				}
			case models.TransactionTypeWithdraw:
				countWithdrawals++
				if tx.Currency == models.CurrencyTON {
					totalWithdrawalsTon += tx.Amount
				} else if tx.Currency == models.CurrencyUSDT {
					totalWithdrawalsUsdt += tx.Amount
				}
			case models.TransactionTypeBet:
				countBets++
				if tx.Currency == models.CurrencyTON {
					totalBetsTon += tx.Amount
				} else if tx.Currency == models.CurrencyUSDT {
					totalBetsUsdt += tx.Amount
				}
			case models.TransactionTypeReward:
				countRewards++
				if tx.Currency == models.CurrencyTON {
					totalRewardsTon += tx.Amount
				} else if tx.Currency == models.CurrencyUSDT {
					totalRewardsUsdt += tx.Amount
				}
			}
		}
	}

	stats["deposits"] = map[string]interface{}{
		"count":      countDeposits,
		"total_ton":  totalDepositsTon,
		"total_usdt": totalDepositsUsdt,
	}
	stats["withdrawals"] = map[string]interface{}{
		"count":      countWithdrawals,
		"total_ton":  totalWithdrawalsTon,
		"total_usdt": totalWithdrawalsUsdt,
	}
	stats["bets"] = map[string]interface{}{
		"count":      countBets,
		"total_ton":  totalBetsTon,
		"total_usdt": totalBetsUsdt,
	}
	stats["rewards"] = map[string]interface{}{
		"count":      countRewards,
		"total_ton":  totalRewardsTon,
		"total_usdt": totalRewardsUsdt,
	}
	stats["net_profit_ton"] = totalRewardsTon - totalBetsTon
	stats["net_profit_usdt"] = totalRewardsUsdt - totalBetsUsdt

	return stats, nil
}

// ProcessWithdraw обрабатывает вывод средств
func (s *TransactionServiceImpl) ProcessWithdraw(ctx context.Context, userID uint64, amount float64, currency string) error {
	if userID == 0 {
		return errors.New("user ID (TelegramID) cannot be zero")
	}
	if amount <= 0 {
		return errors.New("withdraw amount must be positive")
	}

	user, err := s.userRepo.GetByTelegramID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found for withdrawal: %w", err)
	}

	// Проверка баланса в зависимости от валюты
	sufficientBalance := false
	if currency == models.CurrencyTON && user.BalanceTon >= amount {
		sufficientBalance = true
	} else if currency == models.CurrencyUSDT && user.BalanceUsdt >= amount {
		sufficientBalance = true
	}

	if !sufficientBalance {
		return fmt.Errorf("insufficient %s balance for withdrawal. Has: %.2f, Wants: %.2f", currency, func() float64 {
			if currency == models.CurrencyTON {
				return user.BalanceTon
			} else {
				return user.BalanceUsdt
			}
		}(), amount)
	}

	tx := &models.Transaction{
		ID:        uuid.New(),
		UserID:    userID,
		Type:      models.TransactionTypeWithdraw,
		Amount:    amount,
		Currency:  currency,
		Status:    models.TransactionStatusPending, // Статус Pending, пока не подтвердится реальный вывод
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.transactionRepo.Create(ctx, tx); err != nil {
		return fmt.Errorf("failed to create withdrawal transaction: %w", err)
	}

	// Списание с баланса происходит ПОСЛЕ успешного подтверждения вывода
	// Здесь мы только создаем транзакцию. Фактическое обновление баланса - отдельный шаг.
	// Или, если это внутренний перевод, можно обновлять баланс сразу, но статус транзакции ставить Completed.
	// Пока оставляем так, подразумевая, что есть внешний процесс для подтверждения вывода.
	return nil
}

// ProcessDeposit обрабатывает пополнение средств
func (s *TransactionServiceImpl) ProcessDeposit(ctx context.Context, userID uint64, amount float64, currency string, txHash string, network string) error {
	if userID == 0 {
		return errors.New("user ID (TelegramID) cannot be zero")
	}
	if amount <= 0 {
		return errors.New("deposit amount must be positive")
	}

	// Проверка существования пользователя (опционально, если UserID в транзакции достаточно)
	// _, err := s.userRepo.GetByTelegramID(ctx, userID)
	// if err != nil {
	// 	return fmt.Errorf("user not found for deposit: %w", err)
	// }

	tx := &models.Transaction{
		ID:        uuid.New(),
		UserID:    userID,
		Type:      models.TransactionTypeDeposit,
		Amount:    amount,
		Currency:  currency,
		Status:    models.TransactionStatusPending, // Статус Pending, пока транзакция не подтвердится в сети
		TxHash:    txHash,                          // Хеш транзакции в блокчейне
		Network:   network,                         // Сеть, например, "TON", "BSC"
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.transactionRepo.Create(ctx, tx); err != nil {
		return fmt.Errorf("failed to create deposit transaction: %w", err)
	}

	// Фактическое обновление баланса должно происходить после подтверждения транзакции
	// внешним обработчиком, который следит за блокчейном.
	return nil
}

// ProcessReward обрабатывает начисление награды
func (s *TransactionServiceImpl) ProcessReward(ctx context.Context, userID uint64, amount float64, currency string, gameID *uuid.UUID, lobbyID *uuid.UUID) error {
	if userID == 0 {
		return errors.New("user ID (TelegramID) cannot be zero for reward")
	}
	if amount <= 0 {
		return errors.New("reward amount must be positive")
	}

	// Проверяем существование пользователя перед созданием транзакции
	// Это опционально, если мы доверяем, что userID всегда валиден на этом этапе
	/* user, err := s.userRepo.GetByTelegramID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found for reward: %w", err)
	} */

	tx := &models.Transaction{
		ID:        uuid.New(),
		UserID:    userID,
		Type:      models.TransactionTypeReward,
		Amount:    amount,
		Currency:  currency,
		Status:    models.TransactionStatusCompleted, // Награды обычно начисляются сразу
		GameID:    gameID,
		LobbyID:   lobbyID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.transactionRepo.Create(ctx, tx); err != nil {
		return fmt.Errorf("failed to create reward transaction: %w", err)
	}

	// Обновляем баланс пользователя, используя специфичные методы UserRepository
	var errUpdateBalance error
	if currency == models.CurrencyTON {
		errUpdateBalance = s.userRepo.UpdateTonBalance(ctx, userID, amount)
	} else if currency == models.CurrencyUSDT {
		errUpdateBalance = s.userRepo.UpdateUsdtBalance(ctx, userID, amount)
	} else {
		// Откатывать созданную транзакцию или нет - вопрос бизнес-логики.
		// Можно пометить ее как failed или удалить.
		// Пока просто возвращаем ошибку, что баланс не обновлен.
		return fmt.Errorf("unknown currency '%s' for reward, balance not updated, transaction %s created", currency, tx.ID.String())
	}

	if errUpdateBalance != nil {
		// Тут сложный момент: транзакция создана, но баланс не обновлен.
		// Нужна логика компенсации: пометить транзакцию как failed, или поставить в очередь на повтор.
		// tx.Status = models.TransactionStatusFailed
		// s.transactionRepo.Update(ctx, tx) // Попытка обновить статус транзакции
		return fmt.Errorf("reward transaction %s created, but failed to update user %s balance: %w", tx.ID.String(), currency, errUpdateBalance)
	}

	return nil
}

// ConfirmDeposit подтверждает транзакцию пополнения и обновляет баланс пользователя.
func (s *TransactionServiceImpl) ConfirmDeposit(ctx context.Context, transactionID uuid.UUID) error {
	if transactionID == uuid.Nil {
		return errors.New("transaction ID cannot be nil for deposit confirmation")
	}

	tx, err := s.transactionRepo.GetByID(ctx, transactionID)
	if err != nil {
		return fmt.Errorf("failed to get transaction for deposit confirmation: %w", err)
	}

	if tx.Type != models.TransactionTypeDeposit {
		return fmt.Errorf("transaction %s is not a deposit type, got %s", transactionID, tx.Type)
	}

	if tx.Status != models.TransactionStatusPending {
		return fmt.Errorf("deposit transaction %s is not in pending status, got %s", transactionID, tx.Status)
	}

	// Зачисляем средства на баланс
	var errUpdateBalance error
	if tx.Currency == models.CurrencyTON {
		errUpdateBalance = s.userRepo.UpdateTonBalance(ctx, tx.UserID, tx.Amount)
	} else if tx.Currency == models.CurrencyUSDT {
		errUpdateBalance = s.userRepo.UpdateUsdtBalance(ctx, tx.UserID, tx.Amount)
	} else {
		tx.Status = models.TransactionStatusFailed
		tx.Description = fmt.Sprintf("Failed due to unknown currency: %s", tx.Currency)
		_ = s.transactionRepo.Update(ctx, tx) // Попытка обновить статус, ошибку не обрабатываем критично здесь
		return fmt.Errorf("unknown currency '%s' for deposit %s, balance not updated", tx.Currency, transactionID)
	}

	if errUpdateBalance != nil {
		tx.Status = models.TransactionStatusFailed
		tx.Description = fmt.Sprintf("Failed to update user balance: %s", errUpdateBalance.Error())
		_ = s.transactionRepo.Update(ctx, tx) // Попытка обновить статус
		return fmt.Errorf("failed to update user balance for deposit %s: %w", transactionID, errUpdateBalance)
	}

	tx.Status = models.TransactionStatusCompleted
	tx.UpdatedAt = time.Now()
	if err := s.transactionRepo.Update(ctx, tx); err != nil {
		// Баланс обновлен, но статус транзакции не удалось обновить. Это серьезная проблема.
		// Требуется логика для обработки таких несоответствий, возможно, фоновый процесс.
		return fmt.Errorf("user balance updated for deposit %s, but failed to update transaction status to completed: %w", transactionID, err)
	}

	return nil
}

// ConfirmWithdrawal подтверждает транзакцию вывода и обновляет баланс пользователя.
func (s *TransactionServiceImpl) ConfirmWithdrawal(ctx context.Context, transactionID uuid.UUID) error {
	if transactionID == uuid.Nil {
		return errors.New("transaction ID cannot be nil for withdrawal confirmation")
	}

	tx, err := s.transactionRepo.GetByID(ctx, transactionID)
	if err != nil {
		return fmt.Errorf("failed to get transaction for withdrawal confirmation: %w", err)
	}

	if tx.Type != models.TransactionTypeWithdraw {
		return fmt.Errorf("transaction %s is not a withdrawal type, got %s", transactionID, tx.Type)
	}

	if tx.Status != models.TransactionStatusPending {
		return fmt.Errorf("withdrawal transaction %s is not in pending status, got %s", transactionID, tx.Status)
	}

	// Проверяем актуальный баланс перед списанием (на случай, если он изменился)
	user, err := s.userRepo.GetByTelegramID(ctx, tx.UserID)
	if err != nil {
		tx.Status = models.TransactionStatusFailed
		tx.Description = fmt.Sprintf("Failed to get user for balance check: %s", err.Error())
		_ = s.transactionRepo.Update(ctx, tx)
		return fmt.Errorf("failed to get user for withdrawal %s: %w", transactionID, err)
	}

	sufficientBalance := false
	if tx.Currency == models.CurrencyTON && user.BalanceTon >= tx.Amount {
		sufficientBalance = true
	} else if tx.Currency == models.CurrencyUSDT && user.BalanceUsdt >= tx.Amount {
		sufficientBalance = true
	}

	if !sufficientBalance {
		tx.Status = models.TransactionStatusFailed
		tx.Description = "Failed due to insufficient funds at the time of confirmation."
		_ = s.transactionRepo.Update(ctx, tx)
		return fmt.Errorf("insufficient %s balance for withdrawal %s at confirmation. Has: %.2f, Wants: %.2f",
			tx.Currency, transactionID,
			func() float64 {
				if tx.Currency == models.CurrencyTON {
					return user.BalanceTon
				} else {
					return user.BalanceUsdt
				}
			}(),
			tx.Amount)
	}

	// Списываем средства с баланса
	var errUpdateBalance error
	if tx.Currency == models.CurrencyTON {
		errUpdateBalance = s.userRepo.UpdateTonBalance(ctx, tx.UserID, -tx.Amount) // Отрицательная сумма для списания
	} else if tx.Currency == models.CurrencyUSDT {
		errUpdateBalance = s.userRepo.UpdateUsdtBalance(ctx, tx.UserID, -tx.Amount) // Отрицательная сумма для списания
	} else {
		tx.Status = models.TransactionStatusFailed
		tx.Description = fmt.Sprintf("Failed due to unknown currency: %s", tx.Currency)
		_ = s.transactionRepo.Update(ctx, tx)
		return fmt.Errorf("unknown currency '%s' for withdrawal %s, balance not updated", tx.Currency, transactionID)
	}

	if errUpdateBalance != nil {
		tx.Status = models.TransactionStatusFailed
		tx.Description = fmt.Sprintf("Failed to update user balance: %s", errUpdateBalance.Error())
		_ = s.transactionRepo.Update(ctx, tx)
		return fmt.Errorf("failed to update user balance for withdrawal %s: %w", transactionID, errUpdateBalance)
	}

	tx.Status = models.TransactionStatusCompleted
	tx.UpdatedAt = time.Now()
	if err := s.transactionRepo.Update(ctx, tx); err != nil {
		// Баланс обновлен, но статус транзакции не удалось обновить. Это серьезная проблема.
		return fmt.Errorf("user balance updated for withdrawal %s, but failed to update transaction status to completed: %w", transactionID, err)
	}

	return nil
}

// FailTransaction отклоняет ожидающую транзакцию.
func (s *TransactionServiceImpl) FailTransaction(ctx context.Context, transactionID uuid.UUID, reason string) error {
	if transactionID == uuid.Nil {
		return errors.New("transaction ID cannot be nil for failure")
	}

	tx, err := s.transactionRepo.GetByID(ctx, transactionID)
	if err != nil {
		return fmt.Errorf("failed to get transaction for failure: %w", err)
	}

	if tx.Status != models.TransactionStatusPending {
		return fmt.Errorf("transaction %s is not in pending status, cannot fail. Current status: %s", transactionID, tx.Status)
	}

	tx.Status = models.TransactionStatusFailed
	tx.UpdatedAt = time.Now()
	if reason != "" {
		if tx.Description != "" {
			tx.Description += "; Failed: " + reason
		} else {
			tx.Description = "Failed: " + reason
		}
	}

	if err := s.transactionRepo.Update(ctx, tx); err != nil {
		return fmt.Errorf("failed to update transaction %s status to failed: %w", transactionID, err)
	}

	return nil
}

// VerifyTonTransaction проверяет транзакцию TON в блокчейне
func (s *TransactionServiceImpl) VerifyTonTransaction(ctx context.Context, txHash string) (bool, error) {
	if txHash == "" {
		return false, errors.New("transaction hash cannot be empty")
	}

	// TODO: Реализовать проверку транзакции в блокчейне TON через API или SDK
	// Псевдокод:
	// 1. Подключиться к ноде TON
	// 2. Получить информацию о транзакции по хэшу
	// 3. Проверить статус и другие параметры транзакции
	// 4. Вернуть true, если транзакция подтверждена

	// Пример заглушки для разработки
	fmt.Printf("Verifying TON transaction: %s\n", txHash)
	return true, nil
}

// GenerateTonWithdrawTransaction генерирует данные для транзакции вывода TON
func (s *TransactionServiceImpl) GenerateTonWithdrawTransaction(ctx context.Context, userID uint64, amount float64, walletAddress string) (map[string]interface{}, error) {
	if userID == 0 {
		return nil, errors.New("user ID cannot be zero")
	}
	if amount <= 0 {
		return nil, errors.New("withdraw amount must be positive")
	}
	if walletAddress == "" {
		return nil, errors.New("wallet address cannot be empty")
	}

	// Проверяем существование пользователя
	user, err := s.userRepo.GetByTelegramID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// TODO: Генерация данных для транзакции TON
	// Псевдокод:
	// 1. Сформировать данные транзакции (получатель, сумма, комментарий)
	// 2. Подготовить данные для отправки на клиент или для подписи

	// Пример ответа с данными транзакции
	transactionData := map[string]interface{}{
		"user_id":        userID,
		"wallet":         walletAddress,
		"amount":         amount,
		"currency":       models.CurrencyTON,
		"network":        "TON",
		"user_balance":   user.BalanceTon,
		"transaction_id": uuid.New().String(),
		// Другие данные, необходимые для фронтенда
	}

	return transactionData, nil
}

// ProcessTonDeposit обрабатывает депозит TON
func (s *TransactionServiceImpl) ProcessTonDeposit(ctx context.Context, userID uint64, amount float64, txHash string) error {
	if userID == 0 {
		return errors.New("user ID cannot be zero")
	}
	if amount <= 0 {
		return errors.New("deposit amount must be positive")
	}
	if txHash == "" {
		return errors.New("transaction hash cannot be empty")
	}

	// Проверяем транзакцию в блокчейне
	verified, err := s.VerifyTonTransaction(ctx, txHash)
	if err != nil {
		return fmt.Errorf("failed to verify TON transaction: %w", err)
	}
	if !verified {
		return errors.New("TON transaction verification failed")
	}

	// Создаем новую транзакцию депозита
	tx := &models.Transaction{
		ID:          uuid.New(),
		UserID:      userID,
		Type:        models.TransactionTypeDeposit,
		Amount:      amount,
		Currency:    models.CurrencyTON,
		Status:      models.TransactionStatusCompleted,
		TxHash:      txHash,
		Network:     "TON",
		Description: fmt.Sprintf("TON deposit of %.6f", amount),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Сохраняем транзакцию
	if err := s.transactionRepo.Create(ctx, tx); err != nil {
		return fmt.Errorf("failed to create deposit transaction: %w", err)
	}

	// Обновляем баланс пользователя
	if err := s.userRepo.UpdateTonBalance(ctx, userID, amount); err != nil {
		// Если не удалось обновить баланс, отмечаем транзакцию как неуспешную
		errorMsg := fmt.Sprintf("Failed to update user balance: %v", err)
		failErr := s.FailTransaction(ctx, tx.ID, errorMsg)
		if failErr != nil {
			fmt.Printf("ERROR: Failed to mark transaction %s as failed: %v\n", tx.ID, failErr)
		}
		return fmt.Errorf("failed to update TON balance: %w", err)
	}

	return nil
}

// GenerateTonWalletAddress генерирует адрес TON кошелька для депозита
func (s *TransactionServiceImpl) GenerateTonWalletAddress(ctx context.Context, userID uint64) (string, error) {
	if userID == 0 {
		return "", errors.New("user ID cannot be zero")
	}

	// Проверяем существование пользователя
	user, err := s.userRepo.GetByTelegramID(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("failed to get user: %w", err)
	}

	// Если у пользователя уже есть адрес кошелька, возвращаем его
	if user.Wallet != "" {
		return user.Wallet, nil
	}

	// TODO: Интеграция с сервисом кошельков TON для генерации адреса
	// Здесь должна быть реальная логика генерации или получения адреса

	// Пример для разработки
	generatedWallet := fmt.Sprintf("EQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQ%d", userID)

	// Обновляем адрес кошелька пользователя
	user.Wallet = generatedWallet
	if err := s.userRepo.Update(ctx, user); err != nil {
		return "", fmt.Errorf("failed to update user wallet address: %w", err)
	}

	return generatedWallet, nil
}

// MonitorPendingWithdrawals отслеживает и обрабатывает отложенные выводы средств
func (s *TransactionServiceImpl) MonitorPendingWithdrawals(ctx context.Context) error {
	// Получаем отложенные транзакции вывода
	pendingWithdrawals, err := s.transactionRepo.GetByType(ctx, models.TransactionTypeWithdraw, 1000, 0)
	if err != nil {
		return fmt.Errorf("failed to get pending withdrawals: %w", err)
	}

	for _, tx := range pendingWithdrawals {
		if tx.Status != models.TransactionStatusPending {
			continue
		}

		// Здесь должна быть логика обработки вывода средств:
		// 1. Проверка статуса в блокчейне или внешней системе
		// 2. Обновление статуса транзакции
		// 3. Обновление баланса пользователя, если необходимо

		// Для примера просто логируем
		fmt.Printf("Monitoring pending withdrawal: %s, amount: %.6f %s\n",
			tx.ID, tx.Amount, tx.Currency)
	}

	return nil
}
