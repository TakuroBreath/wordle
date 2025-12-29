package worker

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/TakuroBreath/wordle/internal/logger"
	"github.com/TakuroBreath/wordle/internal/models"
	"github.com/TakuroBreath/wordle/pkg/metrics"
	"go.uber.org/zap"
)

// BlockchainWorker обрабатывает транзакции из блокчейна
type BlockchainWorker struct {
	tonService         models.TONService
	gameRepo           models.GameRepository
	lobbyRepo          models.LobbyRepository
	userRepo           models.UserRepository
	transactionRepo    models.TransactionRepository
	
	masterWalletAddress string
	pollInterval        time.Duration
	lastProcessedLt     int64
	
	mu       sync.RWMutex
	running  bool
	stopChan chan struct{}
	logger   *zap.Logger
}

// WorkerConfig конфигурация воркера
type WorkerConfig struct {
	PollInterval        time.Duration
	MasterWalletAddress string
}

// NewBlockchainWorker создаёт новый воркер
func NewBlockchainWorker(
	tonService models.TONService,
	gameRepo models.GameRepository,
	lobbyRepo models.LobbyRepository,
	userRepo models.UserRepository,
	transactionRepo models.TransactionRepository,
	config WorkerConfig,
) *BlockchainWorker {
	return &BlockchainWorker{
		tonService:          tonService,
		gameRepo:            gameRepo,
		lobbyRepo:           lobbyRepo,
		userRepo:            userRepo,
		transactionRepo:     transactionRepo,
		masterWalletAddress: config.MasterWalletAddress,
		pollInterval:        config.PollInterval,
		stopChan:            make(chan struct{}),
		logger:              logger.GetLogger(zap.String("worker", "blockchain")),
	}
}

// Start запускает воркер
func (w *BlockchainWorker) Start(ctx context.Context) error {
	w.mu.Lock()
	if w.running {
		w.mu.Unlock()
		return fmt.Errorf("worker already running")
	}
	w.running = true
	w.mu.Unlock()

	w.logger.Info("Starting blockchain worker",
		zap.Duration("poll_interval", w.pollInterval),
		zap.String("master_wallet", w.masterWalletAddress))

	// Загружаем последний обработанный lt из БД
	lastLt, err := w.transactionRepo.GetLastProcessedLt(ctx)
	if err != nil {
		w.logger.Warn("Failed to get last processed lt, starting from 0", zap.Error(err))
	} else {
		w.lastProcessedLt = lastLt
	}

	go w.run(ctx)
	return nil
}

// Stop останавливает воркер
func (w *BlockchainWorker) Stop() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.running {
		return nil
	}

	w.logger.Info("Stopping blockchain worker")
	close(w.stopChan)
	w.running = false
	return nil
}

// run основной цикл воркера
func (w *BlockchainWorker) run(ctx context.Context) {
	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	// Первая проверка сразу
	w.processTransactions(ctx)

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("Context cancelled, stopping worker")
			return
		case <-w.stopChan:
			w.logger.Info("Stop signal received")
			return
		case <-ticker.C:
			w.processTransactions(ctx)
		}
	}
}

// processTransactions обрабатывает новые транзакции
func (w *BlockchainWorker) processTransactions(ctx context.Context) {
	w.logger.Debug("Processing incoming transactions", zap.Int64("after_lt", w.lastProcessedLt))

	// Получаем новые транзакции
	txs, err := w.tonService.GetNewTransactions(ctx, w.lastProcessedLt)
	if err != nil {
		w.logger.Error("Failed to get new transactions", zap.Error(err))
		return
	}

	if len(txs) == 0 {
		w.logger.Debug("No new transactions")
		return
	}

	w.logger.Info("Found new transactions", zap.Int("count", len(txs)))

	// Обрабатываем каждую транзакцию
	for _, tx := range txs {
		// Пропускаем уже обработанные
		if tx.Lt <= w.lastProcessedLt {
			continue
		}

		// Пропускаем исходящие транзакции
		if !tx.IsIncoming {
			w.logger.Debug("Skipping outgoing transaction", zap.String("hash", tx.Hash))
			w.updateLastProcessedLt(ctx, tx.Lt)
			continue
		}

		// Проверяем, не обработана ли уже эта транзакция
		exists, err := w.transactionRepo.ExistsByTxHash(ctx, tx.Hash)
		if err != nil {
			w.logger.Error("Failed to check transaction existence", zap.Error(err), zap.String("hash", tx.Hash))
			continue
		}
		if exists {
			w.logger.Debug("Transaction already processed", zap.String("hash", tx.Hash))
			w.updateLastProcessedLt(ctx, tx.Lt)
			continue
		}

		// Обрабатываем транзакцию
		if err := w.processTransaction(ctx, tx); err != nil {
			w.logger.Error("Failed to process transaction",
				zap.Error(err),
				zap.String("hash", tx.Hash),
				zap.String("comment", tx.Comment))
		}

		w.updateLastProcessedLt(ctx, tx.Lt)
	}
}

// updateLastProcessedLt обновляет последний обработанный lt
func (w *BlockchainWorker) updateLastProcessedLt(ctx context.Context, lt int64) {
	w.mu.Lock()
	if lt > w.lastProcessedLt {
		w.lastProcessedLt = lt
	}
	w.mu.Unlock()

	// Сохраняем в БД
	if err := w.transactionRepo.UpdateLastProcessedLt(ctx, lt); err != nil {
		w.logger.Warn("Failed to update last processed lt in DB", zap.Error(err))
	}
}

// processTransaction обрабатывает одну транзакцию
func (w *BlockchainWorker) processTransaction(ctx context.Context, tx *models.BlockchainTransaction) error {
	w.logger.Info("Processing transaction",
		zap.String("hash", tx.Hash),
		zap.Float64("amount", tx.Amount),
		zap.String("from", tx.FromAddress),
		zap.String("comment", tx.Comment))

	// Парсим комментарий для определения типа платежа
	paymentType, shortID, ok := models.ParsePaymentComment(tx.Comment)
	if !ok {
		// Это может быть простой депозит без привязки к игре
		w.logger.Debug("Transaction has no valid payment comment, checking for user deposit",
			zap.String("comment", tx.Comment))
		return w.processUserDeposit(ctx, tx)
	}

	// Сохраняем транзакцию в БД
	dbTx := &models.Transaction{
		UserID:       0, // Будет заполнен позже
		Type:         models.TransactionTypeDeposit,
		Amount:       tx.Amount,
		Currency:     tx.Currency,
		Status:       models.TransactionStatusCompleted,
		TxHash:       tx.Hash,
		BlockchainLt: tx.Lt,
		FromAddress:  tx.FromAddress,
		ToAddress:    tx.ToAddress,
		Comment:      tx.Comment,
		Network:      "ton",
		GameShortID:  shortID,
		CreatedAt:    tx.Timestamp,
		UpdatedAt:    time.Now(),
	}

	switch paymentType {
	case models.PaymentTypeGameDeposit:
		return w.processGameDeposit(ctx, shortID, tx, dbTx)
	case models.PaymentTypeLobbyBet:
		return w.processLobbyBet(ctx, shortID, tx, dbTx)
	default:
		return w.processUserDeposit(ctx, tx)
	}
}

// processGameDeposit обрабатывает депозит для активации игры
func (w *BlockchainWorker) processGameDeposit(ctx context.Context, gameShortID string, tx *models.BlockchainTransaction, dbTx *models.Transaction) error {
	w.logger.Info("Processing game deposit",
		zap.String("game_short_id", gameShortID),
		zap.Float64("amount", tx.Amount))

	// Находим игру по short ID
	game, err := w.gameRepo.GetByShortID(ctx, gameShortID)
	if err != nil {
		return fmt.Errorf("game not found: %w", err)
	}

	// Проверяем статус игры
	if game.Status != models.GameStatusPending {
		w.logger.Warn("Game is not pending, skipping deposit",
			zap.String("game_id", game.ID.String()),
			zap.String("status", game.Status))
		return nil
	}

	// Находим пользователя по адресу кошелька
	user, err := w.userRepo.GetByWallet(ctx, tx.FromAddress)
	if err != nil {
		// Пользователь не найден - возможно кошелёк ещё не привязан
		w.logger.Warn("User not found by wallet, using game creator",
			zap.String("wallet", tx.FromAddress))
		user, err = w.userRepo.GetByTelegramID(ctx, game.CreatorID)
		if err != nil {
			return fmt.Errorf("creator not found: %w", err)
		}
	}

	// Проверяем, что отправитель - создатель игры или его кошелёк
	if user.TelegramID != game.CreatorID && user.Wallet != tx.FromAddress {
		w.logger.Warn("Deposit from non-creator, crediting to creator anyway",
			zap.Uint64("sender_user_id", user.TelegramID),
			zap.Uint64("creator_id", game.CreatorID))
	}

	// Обновляем транзакцию
	dbTx.UserID = game.CreatorID
	dbTx.Type = models.TransactionTypeGameDeposit
	dbTx.GameID = &game.ID
	dbTx.GameShortID = gameShortID
	dbTx.Description = fmt.Sprintf("Game deposit for %s", game.Title)

	// Сохраняем транзакцию
	if err := w.transactionRepo.Create(ctx, dbTx); err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}

	// Добавляем в reward pool
	if game.Currency == models.CurrencyTON {
		game.RewardPoolTon += tx.Amount
	} else {
		game.RewardPoolUsdt += tx.Amount
	}
	game.DepositTxHash = tx.Hash
	game.DepositAmount = tx.Amount
	game.UpdatedAt = time.Now()

	// Проверяем, достаточно ли средств для активации
	requiredDeposit := game.GetRequiredDeposit()
	var currentPool float64
	if game.Currency == models.CurrencyTON {
		currentPool = game.RewardPoolTon
	} else {
		currentPool = game.RewardPoolUsdt
	}

	if currentPool >= requiredDeposit {
		w.logger.Info("Sufficient funds for game activation",
			zap.String("game_id", game.ID.String()),
			zap.Float64("required", requiredDeposit),
			zap.Float64("current", currentPool))

		game.Status = models.GameStatusActive
	} else {
		w.logger.Info("Insufficient funds for game activation",
			zap.String("game_id", game.ID.String()),
			zap.Float64("required", requiredDeposit),
			zap.Float64("current", currentPool))
	}

	// Обновляем игру
	if err := w.gameRepo.Update(ctx, game); err != nil {
		return fmt.Errorf("failed to update game: %w", err)
	}

	metrics.AddDeposit(tx.Amount, tx.Currency, "game_deposit")

	w.logger.Info("Game deposit processed successfully",
		zap.String("game_id", game.ID.String()),
		zap.String("status", game.Status),
		zap.Float64("amount", tx.Amount))

	return nil
}

// processLobbyBet обрабатывает ставку для вступления в игру
func (w *BlockchainWorker) processLobbyBet(ctx context.Context, gameShortID string, tx *models.BlockchainTransaction, dbTx *models.Transaction) error {
	receivedAmount := tx.Amount

	w.logger.Info("Processing lobby bet",
		zap.String("game_short_id", gameShortID),
		zap.Float64("amount", tx.Amount))

	// Находим игру по short ID
	game, err := w.gameRepo.GetByShortID(ctx, gameShortID)
	if err != nil {
		return fmt.Errorf("game not found: %w", err)
	}

	// Проверяем статус игры
	if game.Status != models.GameStatusActive {
		w.logger.Warn("Game is not active",
			zap.String("game_id", game.ID.String()),
			zap.String("status", game.Status))
		// TODO: Вернуть деньги отправителю
		return nil
	}

	// Находим пользователя по адресу кошелька
	user, err := w.userRepo.GetByWallet(ctx, tx.FromAddress)
	if err != nil {
		w.logger.Warn("User not found by wallet address",
			zap.String("wallet", tx.FromAddress))
		// TODO: Вернуть деньги или создать пользователя
		return nil
	}

	// Проверяем сумму ставки
	if tx.Amount < game.MinBet {
		w.logger.Warn("Bet amount too low",
			zap.Float64("bet", tx.Amount),
			zap.Float64("min_bet", game.MinBet))
		// TODO: Вернуть деньги
		return nil
	}

	if tx.Amount > game.MaxBet {
		w.logger.Warn("Bet amount too high, capping to max_bet",
			zap.Float64("bet", tx.Amount),
			zap.Float64("max_bet", game.MaxBet))
		tx.Amount = game.MaxBet
	}

	// Проверяем, достаточно ли средств в пуле игры
	if !game.CanAcceptBet(tx.Amount) {
		w.logger.Warn("Game cannot accept bet, insufficient pool",
			zap.Float64("bet", tx.Amount),
			zap.Float64("available", game.GetAvailableRewardPool()))
		// TODO: Вернуть деньги
		return nil
	}

	// Проверяем, нет ли уже активного лобби
	existingLobby, err := w.lobbyRepo.GetActiveByGameAndUser(ctx, game.ID, user.TelegramID)
	if err == nil && existingLobby != nil {
		w.logger.Warn("User already has active lobby",
			zap.Uint64("user_id", user.TelegramID),
			zap.String("lobby_id", existingLobby.ID.String()))
		// Добавляем средства на баланс пользователя
		if game.Currency == models.CurrencyTON {
			if err := w.userRepo.UpdateTonBalance(ctx, user.TelegramID, tx.Amount); err != nil {
				return fmt.Errorf("failed to credit user balance: %w", err)
			}
		}
		return nil
	}

	// Резервируем средства в игре
	potentialReward := tx.Amount * game.RewardMultiplier
	if err := w.gameRepo.IncrementReservedAmount(ctx, game.ID, potentialReward); err != nil {
		return fmt.Errorf("failed to reserve funds: %w", err)
	}

	// Создаём лобби
	now := time.Now()
	expiresAt := now.Add(time.Duration(game.TimeLimit) * time.Minute)
	lobby := &models.Lobby{
		GameID:          game.ID,
		GameShortID:     game.ShortID,
		UserID:          user.TelegramID,
		MaxTries:        game.MaxTries,
		TriesUsed:       0,
		BetAmount:       tx.Amount,
		PotentialReward: potentialReward,
		PaymentTxHash:   tx.Hash,
		Currency:        game.Currency,
		Status:          models.LobbyStatusActive,
		StartedAt:       &now,
		ExpiresAt:       expiresAt,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if err := w.lobbyRepo.Create(ctx, lobby); err != nil {
		// Откатываем резервирование
		_ = w.gameRepo.DecrementReservedAmount(ctx, game.ID, potentialReward)
		return fmt.Errorf("failed to create lobby: %w", err)
	}

	// Обновляем транзакцию
	dbTx.UserID = user.TelegramID
	dbTx.Type = models.TransactionTypeBet
	dbTx.GameID = &game.ID
	dbTx.GameShortID = gameShortID
	dbTx.LobbyID = &lobby.ID
	dbTx.Description = fmt.Sprintf("Bet for game %s", game.Title)

	// Сохраняем транзакцию
	if err := w.transactionRepo.Create(ctx, dbTx); err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}

	metrics.AddDeposit(receivedAmount, tx.Currency, "lobby_bet")

	w.logger.Info("Lobby created successfully",
		zap.String("lobby_id", lobby.ID.String()),
		zap.Uint64("user_id", user.TelegramID),
		zap.Float64("bet", tx.Amount),
		zap.Float64("potential_reward", potentialReward))

	return nil
}

// processUserDeposit обрабатывает обычный депозит пользователя
func (w *BlockchainWorker) processUserDeposit(ctx context.Context, tx *models.BlockchainTransaction) error {
	w.logger.Info("Processing user deposit",
		zap.Float64("amount", tx.Amount),
		zap.String("from", tx.FromAddress))

	// Находим пользователя по адресу кошелька
	user, err := w.userRepo.GetByWallet(ctx, tx.FromAddress)
	if err != nil {
		w.logger.Warn("User not found by wallet, deposit will be credited when wallet is linked",
			zap.String("wallet", tx.FromAddress))
		
		// Сохраняем транзакцию как pending
		dbTx := &models.Transaction{
			UserID:       0,
			Type:         models.TransactionTypeDeposit,
			Amount:       tx.Amount,
			Currency:     tx.Currency,
			Status:       models.TransactionStatusPending,
			TxHash:       tx.Hash,
			BlockchainLt: tx.Lt,
			FromAddress:  tx.FromAddress,
			ToAddress:    tx.ToAddress,
			Comment:      tx.Comment,
			Network:      "ton",
			Description:  "Deposit from unknown wallet",
			CreatedAt:    tx.Timestamp,
			UpdatedAt:    time.Now(),
		}
		return w.transactionRepo.Create(ctx, dbTx)
	}

	// Обновляем баланс пользователя
	if tx.Currency == models.CurrencyTON {
		if err := w.userRepo.UpdateTonBalance(ctx, user.TelegramID, tx.Amount); err != nil {
			return fmt.Errorf("failed to update user balance: %w", err)
		}
	} else {
		if err := w.userRepo.UpdateUsdtBalance(ctx, user.TelegramID, tx.Amount); err != nil {
			return fmt.Errorf("failed to update user balance: %w", err)
		}
	}

	// Сохраняем транзакцию
	dbTx := &models.Transaction{
		UserID:       user.TelegramID,
		Type:         models.TransactionTypeDeposit,
		Amount:       tx.Amount,
		Currency:     tx.Currency,
		Status:       models.TransactionStatusCompleted,
		TxHash:       tx.Hash,
		BlockchainLt: tx.Lt,
		FromAddress:  tx.FromAddress,
		ToAddress:    tx.ToAddress,
		Comment:      tx.Comment,
		Network:      "ton",
		Description:  "Deposit",
		CreatedAt:    tx.Timestamp,
		UpdatedAt:    time.Now(),
	}

	if err := w.transactionRepo.Create(ctx, dbTx); err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}

	metrics.AddDeposit(tx.Amount, tx.Currency, "user_deposit")

	w.logger.Info("User deposit processed successfully",
		zap.Uint64("user_id", user.TelegramID),
		zap.Float64("amount", tx.Amount))

	return nil
}

// ProcessPendingWithdrawals обрабатывает ожидающие выводы
func (w *BlockchainWorker) ProcessPendingWithdrawals(ctx context.Context) error {
	w.logger.Debug("Processing pending withdrawals")

	// Получаем ожидающие выводы
	txs, err := w.transactionRepo.GetPendingWithdrawals(ctx, 10)
	if err != nil {
		return fmt.Errorf("failed to get pending withdrawals: %w", err)
	}

	if len(txs) == 0 {
		return nil
	}

	w.logger.Info("Found pending withdrawals", zap.Int("count", len(txs)))

	for _, tx := range txs {
		if err := w.processWithdrawal(ctx, tx); err != nil {
			w.logger.Error("Failed to process withdrawal",
				zap.Error(err),
				zap.String("tx_id", tx.ID.String()))
		}
	}

	return nil
}

// processWithdrawal обрабатывает один вывод
func (w *BlockchainWorker) processWithdrawal(ctx context.Context, tx *models.Transaction) error {
	w.logger.Info("Processing withdrawal",
		zap.String("tx_id", tx.ID.String()),
		zap.Float64("amount", tx.Amount),
		zap.String("to", tx.ToAddress))

	// Отправляем транзакцию
	var txHash string
	var err error

	if tx.Currency == models.CurrencyTON {
		txHash, err = w.tonService.SendTON(ctx, tx.ToAddress, tx.Amount-tx.Fee, tx.Comment)
	} else {
		txHash, err = w.tonService.SendUSDT(ctx, tx.ToAddress, tx.Amount-tx.Fee, tx.Comment)
	}

	if err != nil {
		// Обновляем статус на failed
		tx.Status = models.TransactionStatusFailed
		tx.ErrorMessage = err.Error()
		tx.UpdatedAt = time.Now()
		_ = w.transactionRepo.Update(ctx, tx)
		return fmt.Errorf("failed to send transaction: %w", err)
	}

	// Обновляем транзакцию
	now := time.Now()
	tx.TxHash = txHash
	tx.Status = models.TransactionStatusCompleted
	tx.ProcessedAt = &now
	tx.UpdatedAt = now

	if err := w.transactionRepo.Update(ctx, tx); err != nil {
		return fmt.Errorf("failed to update transaction: %w", err)
	}

	netAmount := tx.Amount - tx.Fee
	if netAmount > 0 {
		metrics.AddWithdraw(netAmount, tx.Currency)
	}
	if tx.Fee > 0 {
		metrics.AddRevenue(tx.Fee, tx.Currency, "withdraw_fee")
	}

	// Обновляем pending_withdrawal у пользователя
	if err := w.userRepo.UpdatePendingWithdrawal(ctx, tx.UserID, -tx.Amount); err != nil {
		w.logger.Error("Failed to update user pending withdrawal",
			zap.Error(err),
			zap.Uint64("user_id", tx.UserID))
	}

	w.logger.Info("Withdrawal processed successfully",
		zap.String("tx_id", tx.ID.String()),
		zap.String("blockchain_tx", txHash))

	return nil
}

// IsValidPaymentComment проверяет, является ли комментарий валидным платёжным комментарием
func IsValidPaymentComment(comment string) bool {
	if len(comment) < 4 {
		return false
	}
	prefix := comment[:2]
	return (prefix == "GD" || prefix == "LB") && comment[2] == '_'
}

// ExtractGameShortIDFromComment извлекает short ID игры из комментария
func ExtractGameShortIDFromComment(comment string) string {
	if !IsValidPaymentComment(comment) {
		return ""
	}
	
	rest := comment[3:]
	idx := strings.Index(rest, "_")
	if idx == -1 {
		return rest
	}
	return rest[:idx]
}
