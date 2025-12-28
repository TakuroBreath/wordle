package service

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/TakuroBreath/wordle/internal/blockchain"
	"github.com/TakuroBreath/wordle/internal/models"
	"github.com/google/uuid"
	"github.com/tonkeeper/tonapi-go"
)

// JobServiceImpl представляет собой реализацию models.JobService
type JobServiceImpl struct {
	lobbyService       models.LobbyService
	transactionService models.TransactionService
	gameService        models.GameService
	userService        models.UserService
	blockchainProvider blockchain.BlockchainProvider
	tonapiClient       *tonapi.Client
}

// NewJobService создает новый экземпляр models.JobService
func NewJobService(
	lobbyService models.LobbyService,
	transactionService models.TransactionService,
	gameService models.GameService,
	userService models.UserService,
) models.JobService {
	token := os.Getenv("TONAPI_KEY")

	var client *tonapi.Client
	var err error
	
	// Создаём клиент только если токен указан
	if token != "" {
		client, err = tonapi.NewClient(tonapi.TonApiURL, tonapi.WithToken(token))
		if err != nil {
			log.Printf("WARNING: Failed to create TONAPI client: %v", err)
		}
	}

	return &JobServiceImpl{
		lobbyService:       lobbyService,
		transactionService: transactionService,
		gameService:        gameService,
		userService:        userService,
		tonapiClient:       client,
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
	// Обрабатываем отложенные транзакции через blockchain provider
	// Примечание: мониторинг выводов осуществляется через BlockchainWorker
	// Здесь можно добавить дополнительную логику обработки транзакций при необходимости

	return nil
}

// MonitorWalletTransactions отслеживает транзакции кошелька и активирует неактивные игры
func (s *JobServiceImpl) MonitorWalletTransactions(ctx context.Context) error {
	// Если клиент не инициализирован, пропускаем
	if s.tonapiClient == nil {
		return nil
	}

	transactions, err := s.tonapiClient.GetBlockchainAccountTransactions(context.Background(), tonapi.GetBlockchainAccountTransactionsParams{
		AccountID: "UQC9L3EkJxGMEYKCOhN3KQGCLh1i51ohTpO1NV21BbMHxTyE",
		Limit:     tonapi.NewOptInt32(10),
	})
	if err != nil {
		return fmt.Errorf("failed to get transactions: %w", err)
	}

	for _, transaction := range transactions.Transactions {
		// Проверяем, была ли транзакция уже обработана
		if s.transactionService.IsTransactionProcessed(ctx, transaction.Hash) {
			continue
		}
		
		if transaction.InMsg.Value.DecodedBody.String() != "" {
			comment := transaction.InMsg.Value.DecodedBody.String()
			
			// Пробуем распарсить comment как UUID игры
			gameID, parseErr := uuid.Parse(comment)
			if parseErr != nil {
				// Комментарий не является UUID игры, пропускаем
				continue
			}
			
			game, err := s.gameService.GetGame(ctx, gameID)
			if err != nil {
				fmt.Printf("WARNING: Failed to get game %s: %v\n", gameID, err)
				continue
			}

			if game.Status == models.GameStatusInactive {
				err = s.gameService.ActivateGame(ctx, game.ID)
				if err != nil {
					fmt.Printf("WARNING: Failed to activate game %s: %v\n", game.ID, err)
					continue
				}
			}
			
			amount := min(float64(transaction.InMsg.Value.Value)/1000000000, game.MaxBet*game.RewardMultiplier)

			err = s.gameService.AddToRewardPool(ctx, game.ID, amount)
			if err != nil {
				fmt.Printf("WARNING: Failed to add to reward pool for game %s: %v\n", game.ID, err)
				continue
			}

			// Создаем запись о депозите через ProcessDeposit
			err = s.transactionService.ProcessDeposit(ctx, game.CreatorID, amount, models.CurrencyTON, transaction.Hash, "TON")
			if err != nil {
				fmt.Printf("WARNING: Failed to process deposit for game %s: %v\n", game.ID, err)
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
		// Проверяем кошелек каждые 2 минуты
		walletTicker := time.NewTicker(5 * time.Second)
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
