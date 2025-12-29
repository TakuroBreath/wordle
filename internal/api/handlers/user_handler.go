package handlers

import (
	"net/http"
	"strconv"

	"github.com/TakuroBreath/wordle/internal/api/middleware"
	"github.com/TakuroBreath/wordle/internal/models"
	"github.com/gin-gonic/gin"
)

// UserHandler представляет обработчики для пользователей
type UserHandler struct {
	userService        models.UserService
	transactionService models.TransactionService
	tonService         models.TONService
}

// NewUserHandler создает новый экземпляр UserHandler
func NewUserHandler(
	userService models.UserService,
	transactionService models.TransactionService,
	tonService models.TONService,
) *UserHandler {
	return &UserHandler{
		userService:        userService,
		transactionService: transactionService,
		tonService:         tonService,
	}
}

// GetCurrentUser возвращает данные текущего пользователя
func (h *UserHandler) GetCurrentUser(c *gin.Context) {
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"telegram_id":        user.TelegramID,
		"username":           user.Username,
		"first_name":         user.FirstName,
		"last_name":          user.LastName,
		"wallet":             user.Wallet,
		"balance_ton":        user.BalanceTon,
		"balance_usdt":       user.BalanceUsdt,
		"pending_withdrawal": user.PendingWithdrawal,
		"wins":               user.Wins,
		"losses":             user.Losses,
		"total_deposited":    user.TotalDeposited,
		"total_withdrawn":    user.TotalWithdrawn,
		"can_withdraw":       user.CanWithdraw(),
	})
}

// GetUserByID возвращает данные пользователя по ID
func (h *UserHandler) GetUserByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	user, err := h.userService.GetUser(c, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	// Для чужих пользователей показываем ограниченную информацию
	c.JSON(http.StatusOK, gin.H{
		"telegram_id": user.TelegramID,
		"username":    user.Username,
		"first_name":  user.FirstName,
		"last_name":   user.LastName,
		"wins":        user.Wins,
		"losses":      user.Losses,
	})
}

// GetUserBalance возвращает баланс пользователя
func (h *UserHandler) GetUserBalance(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}

	user, err := h.userService.GetUser(c, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"balance_ton":          user.BalanceTon,
		"balance_usdt":         user.BalanceUsdt,
		"available_ton":        user.GetAvailableBalance(models.CurrencyTON),
		"available_usdt":       user.GetAvailableBalance(models.CurrencyUSDT),
		"pending_withdrawal":   user.PendingWithdrawal,
		"can_withdraw":         user.CanWithdraw(),
	})
}

// ConnectWallet подключает TON кошелёк пользователя
func (h *UserHandler) ConnectWallet(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}

	var input struct {
		Wallet string `json:"wallet" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Валидируем адрес
	if h.tonService != nil && !h.tonService.ValidateAddress(input.Wallet) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid wallet address"})
		return
	}

	if err := h.userService.UpdateWallet(c, userID, input.Wallet); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Wallet connected successfully",
		"wallet":  input.Wallet,
	})
}

// GetDepositInfo возвращает информацию для пополнения баланса
func (h *UserHandler) GetDepositInfo(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}

	currency := c.DefaultQuery("currency", models.CurrencyTON)
	if currency != models.CurrencyTON && currency != models.CurrencyUSDT {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid currency"})
		return
	}

	// Получаем адрес для депозита
	masterWallet := h.tonService.GetMasterWalletAddress()

	// Генерируем уникальный комментарий для идентификации
	comment := "deposit" // Простой депозит без комментария - будет идентифицирован по адресу кошелька

	// Получаем пользователя для проверки кошелька
	user, err := h.userService.GetUser(c, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user"})
		return
	}

	var deepLink string
	if h.tonService != nil {
		deepLink = h.tonService.GeneratePaymentDeepLink(masterWallet, 0, comment)
	}

	c.JSON(http.StatusOK, gin.H{
		"address":        masterWallet,
		"currency":       currency,
		"deep_link":      deepLink,
		"user_wallet":    user.Wallet,
		"instructions": "Send TON to the address above. Make sure your wallet is connected so we can identify your deposit.",
	})
}

// RequestWithdraw обрабатывает запрос на вывод средств
func (h *UserHandler) RequestWithdraw(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}

	var input struct {
		Amount    float64 `json:"amount" binding:"required,gt=0"`
		Currency  string  `json:"currency" binding:"required"`
		ToAddress string  `json:"to_address"` // Опционально, по умолчанию - привязанный кошелёк
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.Currency != models.CurrencyTON && input.Currency != models.CurrencyUSDT {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid currency"})
		return
	}

	// Проверяем возможность вывода
	canWithdraw, err := h.userService.CanWithdraw(c, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check withdrawal status"})
		return
	}
	if !canWithdraw {
		c.JSON(http.StatusBadRequest, gin.H{"error": "withdrawal is temporarily locked, please wait"})
		return
	}

	result, err := h.userService.RequestWithdraw(c, userID, input.Amount, input.Currency, input.ToAddress)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":        "Withdrawal request created",
		"transaction_id": result.TransactionID,
		"status":         result.Status,
		"amount":         input.Amount,
		"fee":            result.Fee,
		"net_amount":     input.Amount - result.Fee,
		"currency":       input.Currency,
	})
}

// GetWithdrawHistory возвращает историю выводов средств
func (h *UserHandler) GetWithdrawHistory(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}

	limit := 10
	offset := 0

	if limitStr := c.Query("limit"); limitStr != "" {
		if val, err := strconv.Atoi(limitStr); err == nil && val > 0 {
			limit = val
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if val, err := strconv.Atoi(offsetStr); err == nil && val >= 0 {
			offset = val
		}
	}

	transactions, err := h.userService.GetWithdrawHistory(c, userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get withdraw history"})
		return
	}

	result := make([]gin.H, 0, len(transactions))
	for _, tx := range transactions {
		result = append(result, gin.H{
			"id":         tx.ID,
			"amount":     tx.Amount,
			"fee":        tx.Fee,
			"net_amount": tx.Amount - tx.Fee,
			"currency":   tx.Currency,
			"status":     tx.Status,
			"to_address": tx.ToAddress,
			"tx_hash":    tx.TxHash,
			"created_at": tx.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, result)
}

// GetTransactionHistory возвращает историю всех транзакций
func (h *UserHandler) GetTransactionHistory(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}

	limit := 20
	offset := 0
	txType := c.Query("type") // deposit, withdraw, bet, reward, all

	if limitStr := c.Query("limit"); limitStr != "" {
		if val, err := strconv.Atoi(limitStr); err == nil && val > 0 {
			limit = val
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if val, err := strconv.Atoi(offsetStr); err == nil && val >= 0 {
			offset = val
		}
	}

	transactions, err := h.transactionService.GetUserTransactions(c, userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get transactions"})
		return
	}

	result := make([]gin.H, 0, len(transactions))
	for _, tx := range transactions {
		// Фильтрация по типу
		if txType != "" && txType != "all" && tx.Type != txType {
			continue
		}

		result = append(result, gin.H{
			"id":          tx.ID,
			"type":        tx.Type,
			"amount":      tx.Amount,
			"fee":         tx.Fee,
			"currency":    tx.Currency,
			"status":      tx.Status,
			"description": tx.Description,
			"tx_hash":     tx.TxHash,
			"game_id":     tx.GameID,
			"lobby_id":    tx.LobbyID,
			"created_at":  tx.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, result)
}

// GetUserStats возвращает статистику пользователя
func (h *UserHandler) GetUserStats(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}

	stats, err := h.userService.GetUserStats(c, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user stats"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// UpdateUser обновляет данные пользователя
func (h *UserHandler) UpdateUser(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}

	var input struct {
		Username  string `json:"username"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.userService.GetUser(c, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user"})
		return
	}

	if input.Username != "" {
		user.Username = input.Username
	}
	if input.FirstName != "" {
		user.FirstName = input.FirstName
	}
	if input.LastName != "" {
		user.LastName = input.LastName
	}

	if err := h.userService.UpdateUser(c, user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User updated successfully"})
}

// GetTopUsers возвращает топ пользователей
func (h *UserHandler) GetTopUsers(c *gin.Context) {
	limit := 10

	if limitStr := c.Query("limit"); limitStr != "" {
		if val, err := strconv.Atoi(limitStr); err == nil && val > 0 && val <= 100 {
			limit = val
		}
	}

	users, err := h.userService.GetTopUsers(c, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get top users"})
		return
	}

	result := make([]gin.H, 0, len(users))
	for i, user := range users {
		result = append(result, gin.H{
			"rank":       i + 1,
			"telegram_id": user.TelegramID,
			"username":   user.Username,
			"first_name": user.FirstName,
			"wins":       user.Wins,
			"losses":     user.Losses,
		})
	}

	c.JSON(http.StatusOK, result)
}

// GenerateWalletAddress генерирует адрес кошелька для пользователя (deprecated)
func (h *UserHandler) GenerateWalletAddress(c *gin.Context) {
	h.GetDepositInfo(c)
}
