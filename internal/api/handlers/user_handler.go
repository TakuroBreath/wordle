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
}

// NewUserHandler создает новый экземпляр UserHandler
func NewUserHandler(userService models.UserService, transactionService models.TransactionService) *UserHandler {
	return &UserHandler{
		userService:        userService,
		transactionService: transactionService,
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
		"telegram_id":  user.TelegramID,
		"username":     user.Username,
		"first_name":   user.FirstName,
		"last_name":    user.LastName,
		"wallet":       user.Wallet,
		"balance_ton":  user.BalanceTon,
		"balance_usdt": user.BalanceUsdt,
		"wins":         user.Wins,
		"losses":       user.Losses,
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
		"balance_ton":  user.BalanceTon,
		"balance_usdt": user.BalanceUsdt,
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
		Amount   float64 `json:"amount" binding:"required,gt=0"`
		Currency string  `json:"currency" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.userService.RequestWithdraw(c, userID, input.Amount, input.Currency)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Withdrawal request successful"})
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

	c.JSON(http.StatusOK, transactions)
}

// GenerateWalletAddress генерирует адрес кошелька для пользователя
func (h *UserHandler) GenerateWalletAddress(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}

	// Получаем валюту из параметров (по умолчанию TON)
	currency := c.DefaultQuery("currency", "TON")

	walletAddress, err := h.transactionService.GenerateDepositAddress(c, userID, currency)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate wallet address"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"wallet_address": walletAddress, "currency": currency})
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
		Wallet    string `json:"wallet"`
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
	if input.Wallet != "" {
		user.Wallet = input.Wallet
	}

	if err := h.userService.UpdateUser(c, user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User updated successfully"})
}
