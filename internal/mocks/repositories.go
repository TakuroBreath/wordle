package mocks

import (
	"context"
	"sync"
	"time"

	"github.com/TakuroBreath/wordle/internal/models"
	"github.com/google/uuid"
)

// MockUserRepository мок для UserRepository
type MockUserRepository struct {
	mu    sync.RWMutex
	users map[uint64]*models.User
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		users: make(map[uint64]*models.User),
	}
}

func (m *MockUserRepository) Create(ctx context.Context, user *models.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now()
	}
	user.UpdatedAt = time.Now()
	m.users[user.TelegramID] = user
	return nil
}

func (m *MockUserRepository) GetByTelegramID(ctx context.Context, telegramID uint64) (*models.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if user, ok := m.users[telegramID]; ok {
		return user, nil
	}
	return nil, models.ErrUserNotFound
}

func (m *MockUserRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, user := range m.users {
		if user.Username == username {
			return user, nil
		}
	}
	return nil, models.ErrUserNotFound
}

func (m *MockUserRepository) GetByWallet(ctx context.Context, wallet string) (*models.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, user := range m.users {
		if user.Wallet == wallet {
			return user, nil
		}
	}
	return nil, models.ErrUserNotFound
}

func (m *MockUserRepository) Update(ctx context.Context, user *models.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	user.UpdatedAt = time.Now()
	m.users[user.TelegramID] = user
	return nil
}

func (m *MockUserRepository) Delete(ctx context.Context, telegramID uint64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.users, telegramID)
	return nil
}

func (m *MockUserRepository) UpdateTonBalance(ctx context.Context, telegramID uint64, amount float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if user, ok := m.users[telegramID]; ok {
		user.BalanceTon += amount
		user.UpdatedAt = time.Now()
		return nil
	}
	return models.ErrUserNotFound
}

func (m *MockUserRepository) UpdateUsdtBalance(ctx context.Context, telegramID uint64, amount float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if user, ok := m.users[telegramID]; ok {
		user.BalanceUsdt += amount
		user.UpdatedAt = time.Now()
		return nil
	}
	return models.ErrUserNotFound
}

func (m *MockUserRepository) GetTopUsers(ctx context.Context, limit int) ([]*models.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var users []*models.User
	for _, user := range m.users {
		users = append(users, user)
		if len(users) >= limit {
			break
		}
	}
	return users, nil
}

func (m *MockUserRepository) UpdateWallet(ctx context.Context, telegramID uint64, wallet string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if user, ok := m.users[telegramID]; ok {
		user.Wallet = wallet
		user.UpdatedAt = time.Now()
		return nil
	}
	return models.ErrUserNotFound
}

func (m *MockUserRepository) UpdatePendingWithdrawal(ctx context.Context, telegramID uint64, amount float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if user, ok := m.users[telegramID]; ok {
		user.PendingWithdrawal = amount
		user.UpdatedAt = time.Now()
		return nil
	}
	return models.ErrUserNotFound
}

func (m *MockUserRepository) SetWithdrawalLock(ctx context.Context, telegramID uint64, until time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if user, ok := m.users[telegramID]; ok {
		user.WithdrawalLockUntil = &until
		user.UpdatedAt = time.Now()
		return nil
	}
	return models.ErrUserNotFound
}

func (m *MockUserRepository) IncrementWins(ctx context.Context, telegramID uint64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if user, ok := m.users[telegramID]; ok {
		user.Wins++
		user.UpdatedAt = time.Now()
		return nil
	}
	return models.ErrUserNotFound
}

func (m *MockUserRepository) IncrementLosses(ctx context.Context, telegramID uint64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if user, ok := m.users[telegramID]; ok {
		user.Losses++
		user.UpdatedAt = time.Now()
		return nil
	}
	return models.ErrUserNotFound
}

func (m *MockUserRepository) GetUserStats(ctx context.Context, userID uint64) (map[string]any, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if user, ok := m.users[userID]; ok {
		return map[string]any{
			"wins":   user.Wins,
			"losses": user.Losses,
			"total":  user.Wins + user.Losses,
		}, nil
	}
	return nil, models.ErrUserNotFound
}

func (m *MockUserRepository) ValidateBalance(ctx context.Context, userID uint64, requiredAmount float64) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if user, ok := m.users[userID]; ok {
		return user.BalanceTon >= requiredAmount, nil
	}
	return false, models.ErrUserNotFound
}

// MockGameRepository мок для GameRepository
type MockGameRepository struct {
	mu    sync.RWMutex
	games map[uuid.UUID]*models.Game
}

func NewMockGameRepository() *MockGameRepository {
	return &MockGameRepository{
		games: make(map[uuid.UUID]*models.Game),
	}
}

func (m *MockGameRepository) Create(ctx context.Context, game *models.Game) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if game.ID == uuid.Nil {
		game.ID = uuid.New()
	}
	if game.CreatedAt.IsZero() {
		game.CreatedAt = time.Now()
	}
	game.UpdatedAt = time.Now()
	m.games[game.ID] = game
	return nil
}

func (m *MockGameRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Game, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if game, ok := m.games[id]; ok {
		return game, nil
	}
	return nil, models.ErrGameNotFound
}

func (m *MockGameRepository) GetByShortID(ctx context.Context, shortID string) (*models.Game, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, game := range m.games {
		if game.ShortID == shortID {
			return game, nil
		}
	}
	return nil, models.ErrGameNotFound
}

func (m *MockGameRepository) GetByUserID(ctx context.Context, userID uint64, limit, offset int) ([]*models.Game, error) {
	return m.GetByCreator(ctx, userID, limit, offset)
}

func (m *MockGameRepository) GetByCreator(ctx context.Context, creatorID uint64, limit, offset int) ([]*models.Game, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var games []*models.Game
	for _, game := range m.games {
		if game.CreatorID == creatorID {
			games = append(games, game)
		}
	}
	return games, nil
}

func (m *MockGameRepository) GetActive(ctx context.Context, limit, offset int) ([]*models.Game, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var games []*models.Game
	for _, game := range m.games {
		if game.Status == models.GameStatusActive {
			games = append(games, game)
		}
	}
	return games, nil
}

func (m *MockGameRepository) GetByStatus(ctx context.Context, status string, limit, offset int) ([]*models.Game, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var games []*models.Game
	for _, game := range m.games {
		if game.Status == status {
			games = append(games, game)
		}
	}
	return games, nil
}

func (m *MockGameRepository) GetPending(ctx context.Context, limit, offset int) ([]*models.Game, error) {
	return m.GetByStatus(ctx, models.GameStatusPending, limit, offset)
}

func (m *MockGameRepository) Update(ctx context.Context, game *models.Game) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	game.UpdatedAt = time.Now()
	m.games[game.ID] = game
	return nil
}

func (m *MockGameRepository) Delete(ctx context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.games, id)
	return nil
}

func (m *MockGameRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if game, ok := m.games[id]; ok {
		game.Status = status
		game.UpdatedAt = time.Now()
		return nil
	}
	return models.ErrGameNotFound
}

func (m *MockGameRepository) UpdateRewardPool(ctx context.Context, id uuid.UUID, rewardPoolTon, rewardPoolUsdt float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if game, ok := m.games[id]; ok {
		game.RewardPoolTon = rewardPoolTon
		game.RewardPoolUsdt = rewardPoolUsdt
		game.UpdatedAt = time.Now()
		return nil
	}
	return models.ErrGameNotFound
}

func (m *MockGameRepository) SearchGames(ctx context.Context, minBet, maxBet float64, difficulty string, limit, offset int) ([]*models.Game, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var games []*models.Game
	for _, game := range m.games {
		if game.Status == models.GameStatusActive {
			if minBet > 0 && game.MinBet < minBet {
				continue
			}
			if maxBet > 0 && game.MaxBet > maxBet {
				continue
			}
			if difficulty != "" && game.Difficulty != difficulty {
				continue
			}
			games = append(games, game)
		}
	}
	return games, nil
}

func (m *MockGameRepository) UpdateReservedAmount(ctx context.Context, id uuid.UUID, reservedAmount float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if game, ok := m.games[id]; ok {
		game.ReservedAmount = reservedAmount
		game.UpdatedAt = time.Now()
		return nil
	}
	return models.ErrGameNotFound
}

func (m *MockGameRepository) IncrementReservedAmount(ctx context.Context, id uuid.UUID, amount float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if game, ok := m.games[id]; ok {
		game.ReservedAmount += amount
		game.UpdatedAt = time.Now()
		return nil
	}
	return models.ErrGameNotFound
}

func (m *MockGameRepository) DecrementReservedAmount(ctx context.Context, id uuid.UUID, amount float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if game, ok := m.games[id]; ok {
		game.ReservedAmount -= amount
		if game.ReservedAmount < 0 {
			game.ReservedAmount = 0
		}
		game.UpdatedAt = time.Now()
		return nil
	}
	return models.ErrGameNotFound
}

func (m *MockGameRepository) CountByUser(ctx context.Context, userID uint64) (int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	count := 0
	for _, game := range m.games {
		if game.CreatorID == userID {
			count++
		}
	}
	return count, nil
}

func (m *MockGameRepository) GetGameStats(ctx context.Context, gameID uuid.UUID) (map[string]any, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if game, ok := m.games[gameID]; ok {
		return map[string]any{
			"id":          game.ID,
			"status":      game.Status,
			"reward_pool": game.RewardPoolTon + game.RewardPoolUsdt,
		}, nil
	}
	return nil, models.ErrGameNotFound
}

// MockTransactionRepository мок для TransactionRepository
type MockTransactionRepository struct {
	mu           sync.RWMutex
	transactions map[uuid.UUID]*models.Transaction
	lastLt       int64
}

func NewMockTransactionRepository() *MockTransactionRepository {
	return &MockTransactionRepository{
		transactions: make(map[uuid.UUID]*models.Transaction),
	}
}

func (m *MockTransactionRepository) Create(ctx context.Context, tx *models.Transaction) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if tx.ID == uuid.Nil {
		tx.ID = uuid.New()
	}
	if tx.CreatedAt.IsZero() {
		tx.CreatedAt = time.Now()
	}
	tx.UpdatedAt = time.Now()
	m.transactions[tx.ID] = tx
	return nil
}

func (m *MockTransactionRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Transaction, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if tx, ok := m.transactions[id]; ok {
		return tx, nil
	}
	return nil, models.ErrTransactionNotFound
}

func (m *MockTransactionRepository) GetByUserID(ctx context.Context, userID uint64, limit, offset int) ([]*models.Transaction, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var txs []*models.Transaction
	for _, tx := range m.transactions {
		if tx.UserID == userID {
			txs = append(txs, tx)
		}
	}
	return txs, nil
}

func (m *MockTransactionRepository) GetByTxHash(ctx context.Context, txHash string) (*models.Transaction, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, tx := range m.transactions {
		if tx.TxHash == txHash {
			return tx, nil
		}
	}
	return nil, models.ErrTransactionNotFound
}

func (m *MockTransactionRepository) ExistsByTxHash(ctx context.Context, txHash string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, tx := range m.transactions {
		if tx.TxHash == txHash {
			return true, nil
		}
	}
	return false, nil
}

func (m *MockTransactionRepository) Update(ctx context.Context, tx *models.Transaction) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	tx.UpdatedAt = time.Now()
	m.transactions[tx.ID] = tx
	return nil
}

func (m *MockTransactionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.transactions, id)
	return nil
}

func (m *MockTransactionRepository) GetByType(ctx context.Context, txType string, limit, offset int) ([]*models.Transaction, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var txs []*models.Transaction
	for _, tx := range m.transactions {
		if tx.Type == txType {
			txs = append(txs, tx)
		}
	}
	return txs, nil
}

func (m *MockTransactionRepository) GetByStatus(ctx context.Context, status string, limit, offset int) ([]*models.Transaction, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var txs []*models.Transaction
	for _, tx := range m.transactions {
		if tx.Status == status {
			txs = append(txs, tx)
		}
	}
	return txs, nil
}

func (m *MockTransactionRepository) GetPendingByGameShortID(ctx context.Context, gameShortID string) ([]*models.Transaction, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var txs []*models.Transaction
	for _, tx := range m.transactions {
		if tx.GameShortID == gameShortID && tx.Status == models.TransactionStatusPending {
			txs = append(txs, tx)
		}
	}
	return txs, nil
}

func (m *MockTransactionRepository) GetPendingWithdrawals(ctx context.Context, limit int) ([]*models.Transaction, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var txs []*models.Transaction
	for _, tx := range m.transactions {
		if tx.Type == models.TransactionTypeWithdraw && tx.Status == models.TransactionStatusPending {
			txs = append(txs, tx)
		}
	}
	return txs, nil
}

func (m *MockTransactionRepository) GetLastProcessedLt(ctx context.Context) (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.lastLt, nil
}

func (m *MockTransactionRepository) UpdateLastProcessedLt(ctx context.Context, lt int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.lastLt = lt
	return nil
}

func (m *MockTransactionRepository) CountByUser(ctx context.Context, userID uint64) (int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	count := 0
	for _, tx := range m.transactions {
		if tx.UserID == userID {
			count++
		}
	}
	return count, nil
}

func (m *MockTransactionRepository) GetUserBalance(ctx context.Context, userID uint64) (float64, error) {
	return 0, nil // Simplified mock
}

func (m *MockTransactionRepository) GetTransactionStats(ctx context.Context, userID uint64) (map[string]any, error) {
	return map[string]any{"total": 0}, nil
}

func (m *MockTransactionRepository) CheckSufficientFunds(ctx context.Context, userID uint64, amount float64) (bool, error) {
	return true, nil // Simplified mock
}

// MockLobbyRepository мок для LobbyRepository
type MockLobbyRepository struct {
	mu      sync.RWMutex
	lobbies map[uuid.UUID]*models.Lobby
}

func NewMockLobbyRepository() *MockLobbyRepository {
	return &MockLobbyRepository{
		lobbies: make(map[uuid.UUID]*models.Lobby),
	}
}

func (m *MockLobbyRepository) Create(ctx context.Context, lobby *models.Lobby) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if lobby.ID == uuid.Nil {
		lobby.ID = uuid.New()
	}
	if lobby.CreatedAt.IsZero() {
		lobby.CreatedAt = time.Now()
	}
	lobby.UpdatedAt = time.Now()
	m.lobbies[lobby.ID] = lobby
	return nil
}

func (m *MockLobbyRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Lobby, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if lobby, ok := m.lobbies[id]; ok {
		return lobby, nil
	}
	return nil, models.ErrLobbyNotFound
}

func (m *MockLobbyRepository) GetByGameID(ctx context.Context, gameID uuid.UUID, limit, offset int) ([]*models.Lobby, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var lobbies []*models.Lobby
	for _, lobby := range m.lobbies {
		if lobby.GameID == gameID {
			lobbies = append(lobbies, lobby)
		}
	}
	return lobbies, nil
}

func (m *MockLobbyRepository) GetByUserID(ctx context.Context, userID uint64, limit, offset int) ([]*models.Lobby, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var lobbies []*models.Lobby
	for _, lobby := range m.lobbies {
		if lobby.UserID == userID {
			lobbies = append(lobbies, lobby)
		}
	}
	return lobbies, nil
}

func (m *MockLobbyRepository) GetPendingByGameShortID(ctx context.Context, gameShortID string, userID uint64) (*models.Lobby, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, lobby := range m.lobbies {
		if lobby.GameShortID == gameShortID && lobby.UserID == userID && lobby.Status == models.LobbyStatusPending {
			return lobby, nil
		}
	}
	return nil, models.ErrLobbyNotFound
}

func (m *MockLobbyRepository) GetActive(ctx context.Context, limit, offset int) ([]*models.Lobby, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var lobbies []*models.Lobby
	for _, lobby := range m.lobbies {
		if lobby.Status == models.LobbyStatusActive {
			lobbies = append(lobbies, lobby)
		}
	}
	return lobbies, nil
}

func (m *MockLobbyRepository) Update(ctx context.Context, lobby *models.Lobby) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	lobby.UpdatedAt = time.Now()
	m.lobbies[lobby.ID] = lobby
	return nil
}

func (m *MockLobbyRepository) Delete(ctx context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.lobbies, id)
	return nil
}

func (m *MockLobbyRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if lobby, ok := m.lobbies[id]; ok {
		lobby.Status = status
		lobby.UpdatedAt = time.Now()
		return nil
	}
	return models.ErrLobbyNotFound
}

func (m *MockLobbyRepository) UpdateTriesUsed(ctx context.Context, id uuid.UUID, triesUsed int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if lobby, ok := m.lobbies[id]; ok {
		lobby.TriesUsed = triesUsed
		lobby.UpdatedAt = time.Now()
		return nil
	}
	return models.ErrLobbyNotFound
}

func (m *MockLobbyRepository) GetExpired(ctx context.Context) ([]*models.Lobby, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var lobbies []*models.Lobby
	now := time.Now()
	for _, lobby := range m.lobbies {
		if lobby.Status == models.LobbyStatusActive && lobby.ExpiresAt.Before(now) {
			lobbies = append(lobbies, lobby)
		}
	}
	return lobbies, nil
}

func (m *MockLobbyRepository) GetActiveByGameAndUser(ctx context.Context, gameID uuid.UUID, userID uint64) (*models.Lobby, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, lobby := range m.lobbies {
		if lobby.GameID == gameID && lobby.UserID == userID && lobby.Status == models.LobbyStatusActive {
			return lobby, nil
		}
	}
	return nil, models.ErrLobbyNotFound
}

func (m *MockLobbyRepository) CountActiveByGame(ctx context.Context, gameID uuid.UUID) (int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	count := 0
	for _, lobby := range m.lobbies {
		if lobby.GameID == gameID && lobby.Status == models.LobbyStatusActive {
			count++
		}
	}
	return count, nil
}

func (m *MockLobbyRepository) CountByUser(ctx context.Context, userID uint64) (int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	count := 0
	for _, lobby := range m.lobbies {
		if lobby.UserID == userID {
			count++
		}
	}
	return count, nil
}

func (m *MockLobbyRepository) GetActiveWithAttempts(ctx context.Context, limit, offset int) ([]*models.Lobby, error) {
	return m.GetActive(ctx, limit, offset)
}

func (m *MockLobbyRepository) ExtendExpirationTime(ctx context.Context, id uuid.UUID, duration int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if lobby, ok := m.lobbies[id]; ok {
		lobby.ExpiresAt = lobby.ExpiresAt.Add(time.Duration(duration) * time.Minute)
		lobby.UpdatedAt = time.Now()
		return nil
	}
	return models.ErrLobbyNotFound
}

func (m *MockLobbyRepository) GetLobbiesByStatus(ctx context.Context, status string, limit, offset int) ([]*models.Lobby, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var lobbies []*models.Lobby
	for _, lobby := range m.lobbies {
		if lobby.Status == status {
			lobbies = append(lobbies, lobby)
		}
	}
	return lobbies, nil
}

func (m *MockLobbyRepository) GetActiveLobbiesByTimeRange(ctx context.Context, startTime, endTime time.Time, limit, offset int) ([]*models.Lobby, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var lobbies []*models.Lobby
	for _, lobby := range m.lobbies {
		if lobby.Status == models.LobbyStatusActive {
			if lobby.CreatedAt.After(startTime) && lobby.CreatedAt.Before(endTime) {
				lobbies = append(lobbies, lobby)
			}
		}
	}
	return lobbies, nil
}

func (m *MockLobbyRepository) GetLobbyStats(ctx context.Context, lobbyID uuid.UUID) (map[string]any, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if lobby, ok := m.lobbies[lobbyID]; ok {
		return map[string]any{
			"id":         lobby.ID,
			"status":     lobby.Status,
			"tries_used": lobby.TriesUsed,
		}, nil
	}
	return nil, models.ErrLobbyNotFound
}

func (m *MockLobbyRepository) StartLobby(ctx context.Context, id uuid.UUID, expiresAt time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if lobby, ok := m.lobbies[id]; ok {
		lobby.Status = models.LobbyStatusActive
		now := time.Now()
		lobby.StartedAt = &now
		lobby.ExpiresAt = expiresAt
		lobby.UpdatedAt = now
		return nil
	}
	return models.ErrLobbyNotFound
}

// MockAttemptRepository мок для AttemptRepository
type MockAttemptRepository struct {
	mu       sync.RWMutex
	attempts map[uuid.UUID]*models.Attempt
}

func NewMockAttemptRepository() *MockAttemptRepository {
	return &MockAttemptRepository{
		attempts: make(map[uuid.UUID]*models.Attempt),
	}
}

func (m *MockAttemptRepository) Create(ctx context.Context, attempt *models.Attempt) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if attempt.ID == uuid.Nil {
		attempt.ID = uuid.New()
	}
	if attempt.CreatedAt.IsZero() {
		attempt.CreatedAt = time.Now()
	}
	m.attempts[attempt.ID] = attempt
	return nil
}

func (m *MockAttemptRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Attempt, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if attempt, ok := m.attempts[id]; ok {
		return attempt, nil
	}
	return nil, models.ErrGameNotFound
}

func (m *MockAttemptRepository) GetByLobbyID(ctx context.Context, lobbyID uuid.UUID, limit, offset int) ([]*models.Attempt, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var attempts []*models.Attempt
	for _, attempt := range m.attempts {
		if attempt.LobbyID != nil && *attempt.LobbyID == lobbyID {
			attempts = append(attempts, attempt)
		}
	}
	return attempts, nil
}

func (m *MockAttemptRepository) Update(ctx context.Context, attempt *models.Attempt) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.attempts[attempt.ID] = attempt
	return nil
}

func (m *MockAttemptRepository) Delete(ctx context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.attempts, id)
	return nil
}

func (m *MockAttemptRepository) GetLastAttempt(ctx context.Context, lobbyID uuid.UUID, userID uint64) (*models.Attempt, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var lastAttempt *models.Attempt
	for _, attempt := range m.attempts {
		if attempt.LobbyID != nil && *attempt.LobbyID == lobbyID && attempt.UserID == userID {
			if lastAttempt == nil || attempt.CreatedAt.After(lastAttempt.CreatedAt) {
				lastAttempt = attempt
			}
		}
	}
	if lastAttempt == nil {
		return nil, models.ErrGameNotFound
	}
	return lastAttempt, nil
}

func (m *MockAttemptRepository) CountByLobbyID(ctx context.Context, lobbyID uuid.UUID) (int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	count := 0
	for _, attempt := range m.attempts {
		if attempt.LobbyID != nil && *attempt.LobbyID == lobbyID {
			count++
		}
	}
	return count, nil
}

func (m *MockAttemptRepository) GetByWord(ctx context.Context, lobbyID uuid.UUID, word string) (*models.Attempt, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, attempt := range m.attempts {
		if attempt.LobbyID != nil && *attempt.LobbyID == lobbyID && attempt.Word == word {
			return attempt, nil
		}
	}
	return nil, models.ErrGameNotFound
}

func (m *MockAttemptRepository) GetAttemptStats(ctx context.Context, lobbyID uuid.UUID) (map[string]any, error) {
	count, _ := m.CountByLobbyID(ctx, lobbyID)
	return map[string]any{"count": count}, nil
}

func (m *MockAttemptRepository) ValidateWord(ctx context.Context, word string) (bool, error) {
	return len(word) > 0, nil
}

// MockHistoryRepository мок для HistoryRepository
type MockHistoryRepository struct {
	mu        sync.RWMutex
	histories map[uuid.UUID]*models.History
}

func NewMockHistoryRepository() *MockHistoryRepository {
	return &MockHistoryRepository{
		histories: make(map[uuid.UUID]*models.History),
	}
}

func (m *MockHistoryRepository) Create(ctx context.Context, history *models.History) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if history.ID == uuid.Nil {
		history.ID = uuid.New()
	}
	if history.CreatedAt.IsZero() {
		history.CreatedAt = time.Now()
	}
	m.histories[history.ID] = history
	return nil
}

func (m *MockHistoryRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.History, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if history, ok := m.histories[id]; ok {
		return history, nil
	}
	return nil, models.ErrGameNotFound
}

func (m *MockHistoryRepository) GetByGameID(ctx context.Context, gameID uuid.UUID, limit, offset int) ([]*models.History, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var histories []*models.History
	for _, history := range m.histories {
		if history.GameID == gameID {
			histories = append(histories, history)
		}
	}
	return histories, nil
}

func (m *MockHistoryRepository) GetByUserID(ctx context.Context, userID uint64, limit, offset int) ([]*models.History, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var histories []*models.History
	for _, history := range m.histories {
		if history.UserID == userID {
			histories = append(histories, history)
		}
	}
	return histories, nil
}

func (m *MockHistoryRepository) Update(ctx context.Context, history *models.History) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.histories[history.ID] = history
	return nil
}

func (m *MockHistoryRepository) Delete(ctx context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.histories, id)
	return nil
}

func (m *MockHistoryRepository) GetByLobbyID(ctx context.Context, lobbyID uuid.UUID) (*models.History, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, history := range m.histories {
		if history.LobbyID == lobbyID {
			return history, nil
		}
	}
	return nil, models.ErrGameNotFound
}

func (m *MockHistoryRepository) GetByStatus(ctx context.Context, status string, limit, offset int) ([]*models.History, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var histories []*models.History
	for _, history := range m.histories {
		if history.Status == status {
			histories = append(histories, history)
		}
	}
	return histories, nil
}

func (m *MockHistoryRepository) CountByUser(ctx context.Context, userID uint64) (int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	count := 0
	for _, history := range m.histories {
		if history.UserID == userID {
			count++
		}
	}
	return count, nil
}

func (m *MockHistoryRepository) GetUserStats(ctx context.Context, userID uint64) (map[string]any, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	wins := 0
	losses := 0
	for _, history := range m.histories {
		if history.UserID == userID {
			if history.Status == models.HistoryStatusPlayerWin {
				wins++
			} else {
				losses++
			}
		}
	}
	return map[string]any{"wins": wins, "losses": losses}, nil
}

func (m *MockHistoryRepository) GetGameHistoryStats(ctx context.Context, gameID uuid.UUID) (map[string]any, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	count := 0
	for _, history := range m.histories {
		if history.GameID == gameID {
			count++
		}
	}
	return map[string]any{"count": count}, nil
}
