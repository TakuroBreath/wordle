package service

import (
	"context"
	"fmt"
	"time"

	"github.com/TakuroBreath/wordle/internal/models"
)

// JobServiceImpl представляет собой реализацию models.JobService
type JobServiceImpl struct {
	lobbyService       models.LobbyService
	transactionService models.TransactionService
	gameService        models.GameService
	userService        models.UserService
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
	// например, депозитов, наград и т.д.

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

	// Ожидаем сигнал остановки
	<-ctx.Done()
	close(done)
}

// RunOnce выполняет все задачи один раз (полезно для тестирования)
func (s *JobServiceImpl) RunOnce(ctx context.Context) error {
	// Обрабатываем истекшие лобби
	if err := s.ProcessExpiredLobbies(ctx); err != nil {
		return fmt.Errorf("failed to process expired lobbies: %w", err)
	}

	// Обрабатываем отложенные транзакции
	if err := s.ProcessPendingTransactions(ctx); err != nil {
		return fmt.Errorf("failed to process pending transactions: %w", err)
	}

	return nil
}
