package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/TakuroBreath/wordle/internal/blockchain"
	"github.com/TakuroBreath/wordle/internal/models"
)

// JobServiceImpl представляет собой реализацию models.JobService
type JobServiceImpl struct {
	lobbyService       models.LobbyService
	transactionService models.TransactionService
	gameService        models.GameService
	userService        models.UserService
	blockchainProvider blockchain.BlockchainProvider
}

// NewJobService создает новый экземпляр models.JobService
func NewJobService(
	lobbyService models.LobbyService,
	transactionService models.TransactionService,
	gameService models.GameService,
	userService models.UserService,
) models.JobService {
	return &JobServiceImpl{
		lobbyService:       lobbyService,
		transactionService: transactionService,
		gameService:        gameService,
		userService:        userService,
	}
}

// NewJobServiceWithBlockchain создает JobService с поддержкой блокчейн провайдера
func NewJobServiceWithBlockchain(
	lobbyService models.LobbyService,
	transactionService models.TransactionService,
	gameService models.GameService,
	userService models.UserService,
	blockchainProvider blockchain.BlockchainProvider,
) models.JobService {
	service := NewJobService(lobbyService, transactionService, gameService, userService).(*JobServiceImpl)
	service.blockchainProvider = blockchainProvider
	return service
}

// ProcessExpiredLobbies обрабатывает истекшие лобби
func (s *JobServiceImpl) ProcessExpiredLobbies(ctx context.Context) error {
	expiredLobbies, err := s.lobbyService.GetExpiredLobbies(ctx)
	if err != nil {
		return fmt.Errorf("failed to get expired lobbies: %w", err)
	}

	for _, lobby := range expiredLobbies {
		// Пропускаем лобби, которые уже не активны
		if lobby.Status != models.LobbyStatusActive {
			continue
		}

		fmt.Printf("Processing expired lobby: %s\n", lobby.ID)
		err := s.lobbyService.FinishLobby(ctx, lobby.ID, false)
		if err != nil {
			fmt.Printf("ERROR: Failed to finish expired lobby %s: %v\n", lobby.ID, err)
		}
	}

	return nil
}

// ProcessPendingTransactions обрабатывает отложенные транзакции
func (s *JobServiceImpl) ProcessPendingTransactions(ctx context.Context) error {
	// Обрабатываем отложенные выводы
	err := s.transactionService.MonitorPendingWithdrawals(ctx)
	if err != nil {
		fmt.Printf("ERROR: Failed to process pending withdrawals: %v\n", err)
	}

	// Здесь можно добавить обработку других типов отложенных транзакций
	return nil
}

// MonitorWalletTransactions отслеживает транзакции кошелька и активирует неактивные игры
func (s *JobServiceImpl) MonitorWalletTransactions(ctx context.Context) error {
	// Если клиент не инициализирован, пропускаем
	if s.blockchainProvider == nil {
		return nil
	}

	// Получаем последние транзакции (входящие)
	transactions, err := s.blockchainProvider.GetRecentTransactions(ctx, 50)
	if err != nil {
		return fmt.Errorf("failed to get transactions: %w", err)
	}

	for _, tx := range transactions {
		// Проверяем, была ли транзакция уже обработана
		// Мы проверяем как в БД (через Service), так и в провайдере (кеш)
		if s.transactionService.IsTransactionProcessed(ctx, tx.Hash, string(tx.Network)) {
			continue
		}

		comment := tx.Comment
		fmt.Printf("Processing transaction %s with comment: %s, amount: %.4f\n", tx.Hash, comment, tx.Amount)

		// Вариант 1: Депозит пользователя (формат: user_<id>)
		if strings.HasPrefix(comment, "user_") {
			uidStr := strings.TrimPrefix(comment, "user_")
			uid, err := strconv.ParseUint(uidStr, 10, 64)
			if err == nil {
				fmt.Printf("Processing deposit for user %d\n", uid)
				err := s.transactionService.ProcessBlockchainDeposit(ctx, uid, tx.Amount, tx.Currency, tx.Hash, string(tx.Network))
				if err != nil {
					fmt.Printf("ERROR: Failed to process deposit for user %d: %v\n", uid, err)
				}
				continue
			}
		}

		// Вариант 2: Активация игры (ShortID)
		// Пробуем найти игру по short_id (комментарию)
		// Проверяем, что комментарий похож на short_id (например, 8 символов)
		if len(comment) >= 4 && len(comment) <= 20 {
			game, err := s.gameService.GetByShortID(ctx, comment)
			if err == nil && game != nil {
				// Игра найдена
				fmt.Printf("Found game %s for activation via tx %s\n", game.ID, tx.Hash)

				if game.Status == models.GameStatusPendingActivation {
					// Проверяем сумму
					if tx.Amount >= game.DepositAmount {
						// Добавляем в reward pool
						err = s.gameService.AddToRewardPool(ctx, game.ID, tx.Amount)
						if err != nil {
							fmt.Printf("ERROR: Failed to add to reward pool for game %s: %v\n", game.ID, err)
						} else {
							// Активируем игру
							err = s.gameService.ActivateGame(ctx, game.ID)
							if err != nil {
								fmt.Printf("ERROR: Failed to activate game %s: %v\n", game.ID, err)
							} else {
								// Записываем транзакцию депозита игры (фактически это депозит создателя)
								// Мы используем ProcessBlockchainDeposit, но помечаем описанием игры
								// Или создаем специфичную транзакцию?
								// Для простоты используем депозит создателя, так как деньги пришли от него
								// Но они уже ушли в пул игры.
								// Чтобы не дублировать баланс пользователя, мы НЕ вызываем ProcessBlockchainDeposit,
								// так как AddToRewardPool уже увеличил пул игры, а ActivateGame списал бы их, если бы они были на балансе.
								// Здесь деньги пришли НАПРЯМУЮ в игру.
								// Поэтому мы просто создаем запись транзакции для истории, но БЕЗ начисления на баланс юзера.
								
								// НО: TransactionService.ProcessBlockchainDeposit начисляет на баланс!
								// Нам нужно создать транзакцию, но не начислять на баланс, или начислить и сразу списать.
								// Лучше: создать транзакцию с типом "deposit" и статусом "completed", но не вызывать userRepo.UpdateBalance.
								// Но у нас нет метода в TransactionService для этого.
								// Тогда сделаем так: Начислим на баланс юзера, а потом спишем в "Bet" или "GameFunding"?
								// Логика игры: Creator вносит депозит.
								// Если мы начислим ему на баланс, а потом игра использует reward pool (отдельное поле), то деньги дублируются?
								// Нет, RewardPool - это поле в игре. Balance - у юзера.
								// Деньги на кошельке компании.
								// Они должны быть либо на балансе юзера, либо в пуле игры.
								// Раз мы сделали AddToRewardPool, значит они в пуле.
								// Значит на баланс юзера начислять НЕ НАДО.
								// Просто сохраняем транзакцию для отчетности.
								
								txRecord := &models.Transaction{
									ID:          models.NewUUID(),
									UserID:      game.CreatorID,
									Type:        models.TransactionTypeDeposit,
									Amount:      tx.Amount,
									Currency:    tx.Currency,
									Status:      models.TransactionStatusCompleted,
									TxHash:      tx.Hash,
									Network:     string(tx.Network),
									GameID:      &game.ID,
									Description: fmt.Sprintf("Game activation deposit for %s", game.ShortID),
									CreatedAt:   time.Now(),
									UpdatedAt:   time.Now(),
								}
								// Прямой доступ к репо здесь недоступен, нужен метод в TransactionService.
								// Или добавим метод RecordGameDeposit в TransactionService.
								// Пока используем CreateTransaction через сервис, если он позволяет
								s.transactionService.CreateTransaction(ctx, txRecord)
							}
						}
					} else {
						fmt.Printf("WARNING: Insufficient funds for game activation. Needed %.2f, got %.2f\n", game.DepositAmount, tx.Amount)
					}
				}
				continue
			}
		}
	}

	return nil
}

// StartJobScheduler запускает планировщик фоновых задач
func (s *JobServiceImpl) StartJobScheduler(ctx context.Context, lobbyCheckInterval, transactionCheckInterval time.Duration) {
	// Канал для остановки планировщика
	done := make(chan bool)

	// Запускаем обработку истекших лобби
	go func() {
		lobbyTicker := time.NewTicker(lobbyCheckInterval)
		defer lobbyTicker.Stop()

		for {
			select {
			case <-lobbyTicker.C:
				err := s.ProcessExpiredLobbies(ctx)
				if err != nil {
					fmt.Printf("ERROR: Failed to process expired lobbies: %v\n", err)
				}
			case <-done:
				return
			case <-ctx.Done():
				return
			}
		}
	}()

	// Запускаем обработку отложенных транзакций
	go func() {
		txTicker := time.NewTicker(transactionCheckInterval)
		defer txTicker.Stop()

		for {
			select {
			case <-txTicker.C:
				err := s.ProcessPendingTransactions(ctx)
				if err != nil {
					fmt.Printf("ERROR: Failed to process pending transactions: %v\n", err)
				}
			case <-done:
				return
			case <-ctx.Done():
				return
			}
		}
	}()

	// Запускаем мониторинг транзакций кошелька и активацию игр
	go func() {
		// Проверяем кошелек каждые 10 секунд
		walletTicker := time.NewTicker(10 * time.Second)
		defer walletTicker.Stop()

		for {
			select {
			case <-walletTicker.C:
				err := s.MonitorWalletTransactions(ctx)
				if err != nil {
					fmt.Printf("ERROR: Failed to monitor wallet transactions: %v\n", err)
				}
			case <-done:
				return
			case <-ctx.Done():
				return
			}
		}
	}()

	// Ожидаем сигнал остановки
	<-ctx.Done()
	close(done)
}

// RunOnce выполняет все задачи один раз
func (s *JobServiceImpl) RunOnce(ctx context.Context) error {
	// Обрабатываем истекшие лобби
	if err := s.ProcessExpiredLobbies(ctx); err != nil {
		return fmt.Errorf("failed to process expired lobbies: %w", err)
	}

	// Обрабатываем отложенные транзакции
	if err := s.ProcessPendingTransactions(ctx); err != nil {
		return fmt.Errorf("failed to process pending transactions: %w", err)
	}

	// Мониторим транзакции кошелька и активируем игры
	if err := s.MonitorWalletTransactions(ctx); err != nil {
		return fmt.Errorf("failed to monitor wallet transactions: %w", err)
	}

	return nil
}
