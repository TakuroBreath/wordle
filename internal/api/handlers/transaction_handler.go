package handlers

import (
	"net/http"
	"strconv"

	"github.com/TakuroBreath/wordle/internal/api/middleware"
	"github.com/TakuroBreath/wordle/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// TransactionHandler представляет обработчики для транзакций
type TransactionHandler struct {
	transactionService models.TransactionService
	userService        models.UserService
}

// NewTransactionHandler создает новый экземпляр TransactionHandler
func NewTransactionHandler(
	transactionService models.TransactionService,
	userService models.UserService,
) *TransactionHandler {
	return &TransactionHandler{
		transactionService: transactionService,
		userService:        userService,
	}
}

// GetUserTransactions получает список транзакций пользователя
func (h *TransactionHandler) GetUserTransactions(c *gin.Context) {
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

	transactions, err := h.transactionService.GetUserTransactions(c, userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user transactions"})
		return
	}

	result := make([]gin.H, 0, len(transactions))
	for _, tx := range transactions {
		txData := gin.H{
			"id":          tx.ID,
			"type":        tx.Type,
			"amount":      tx.Amount,
			"currency":    tx.Currency,
			"status":      tx.Status,
			"description": tx.Description,
			"created_at":  tx.CreatedAt,
			"updated_at":  tx.UpdatedAt,
			"network":     tx.Network,
			"tx_hash":     tx.TxHash,
		}

		if tx.GameID != nil {
			txData["game_id"] = *tx.GameID
		}

		if tx.LobbyID != nil {
			txData["lobby_id"] = *tx.LobbyID
		}

		result = append(result, txData)
	}

	c.JSON(http.StatusOK, result)
}

// GetTransaction получает информацию о транзакции
func (h *TransactionHandler) GetTransaction(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid transaction ID"})
		return
	}

	tx, err := h.transactionService.GetTransaction(c, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "transaction not found"})
		return
	}

	// Проверяем, что пользователь имеет доступ к этой транзакции
	if tx.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied to this transaction"})
		return
	}

	result := gin.H{
		"id":          tx.ID,
		"user_id":     tx.UserID,
		"type":        tx.Type,
		"amount":      tx.Amount,
		"currency":    tx.Currency,
		"status":      tx.Status,
		"description": tx.Description,
		"created_at":  tx.CreatedAt,
		"updated_at":  tx.UpdatedAt,
		"network":     tx.Network,
		"tx_hash":     tx.TxHash,
	}

	if tx.GameID != nil {
		result["game_id"] = *tx.GameID
	}

	if tx.LobbyID != nil {
		result["lobby_id"] = *tx.LobbyID
	}

	c.JSON(http.StatusOK, result)
}

// GetTransactionsByType получает список транзакций по типу
func (h *TransactionHandler) GetTransactionsByType(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}

	txType := c.Query("type")
	if txType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "transaction type is required"})
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

	transactions, err := h.transactionService.GetTransactionsByType(c, txType, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get transactions by type"})
		return
	}

	// Фильтруем транзакции только для текущего пользователя
	userTransactions := make([]*models.Transaction, 0)
	for _, tx := range transactions {
		if tx.UserID == userID {
			userTransactions = append(userTransactions, tx)
		}
	}

	result := make([]gin.H, 0, len(userTransactions))
	for _, tx := range userTransactions {
		txData := gin.H{
			"id":          tx.ID,
			"type":        tx.Type,
			"amount":      tx.Amount,
			"currency":    tx.Currency,
			"status":      tx.Status,
			"description": tx.Description,
			"created_at":  tx.CreatedAt,
			"updated_at":  tx.UpdatedAt,
			"network":     tx.Network,
			"tx_hash":     tx.TxHash,
		}

		if tx.GameID != nil {
			txData["game_id"] = *tx.GameID
		}

		if tx.LobbyID != nil {
			txData["lobby_id"] = *tx.LobbyID
		}

		result = append(result, txData)
	}

	c.JSON(http.StatusOK, result)
}

// GetTransactionStats получает статистику транзакций пользователя
func (h *TransactionHandler) GetTransactionStats(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}

	stats, err := h.transactionService.GetTransactionStats(c, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get transaction stats"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// CreateDepositTransaction создает транзакцию пополнения баланса
func (h *TransactionHandler) CreateDepositTransaction(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}

	var input struct {
		Amount   float64 `json:"amount" binding:"required,gt=0"`
		Currency string  `json:"currency" binding:"required"`
		TxHash   string  `json:"tx_hash" binding:"required"`
		Network  string  `json:"network" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Проверка валидности валюты
	if input.Currency != models.CurrencyTON && input.Currency != models.CurrencyUSDT {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid currency, must be TON or USDT"})
		return
	}

	// Создаем транзакцию пополнения
	tx := &models.Transaction{
		UserID:      userID,
		Type:        models.TransactionTypeDeposit,
		Amount:      input.Amount,
		Currency:    input.Currency,
		Status:      models.TransactionStatusPending,
		Description: "Deposit via blockchain",
		TxHash:      input.TxHash,
		Network:     input.Network,
	}

	if err := h.transactionService.CreateTransaction(c, tx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":         tx.ID,
		"status":     tx.Status,
		"message":    "Deposit transaction created successfully. Will be processed shortly.",
		"amount":     tx.Amount,
		"currency":   tx.Currency,
		"tx_hash":    tx.TxHash,
		"network":    tx.Network,
		"created_at": tx.CreatedAt,
	})
}

// VerifyDeposit проверяет транзакцию пополнения
func (h *TransactionHandler) VerifyDeposit(c *gin.Context) {
	_, exists := middleware.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}

	var input struct {
		TxHash  string `json:"tx_hash" binding:"required"`
		Network string `json:"network" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Проверяем транзакцию в блокчейне
	verified, err := h.transactionService.VerifyBlockchainTransaction(c, input.TxHash, input.Network)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if !verified {
		c.JSON(http.StatusBadRequest, gin.H{"error": "transaction not verified"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"verified": true,
		"message":  "Transaction verified successfully. Your balance will be updated soon.",
	})
}

// GetDepositAddress получает адрес для депозита
func (h *TransactionHandler) GetDepositAddress(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}

	currency := c.DefaultQuery("currency", "TON")

	address, err := h.transactionService.GenerateDepositAddress(c, userID, currency)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"address":  address,
		"currency": currency,
	})
}

// PrepareWithdraw подготавливает данные для вывода
func (h *TransactionHandler) PrepareWithdraw(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}

	var input struct {
		Amount   float64 `json:"amount" binding:"required,gt=0"`
		Currency string  `json:"currency" binding:"required"`
		Wallet   string  `json:"wallet" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	txData, err := h.transactionService.GenerateWithdrawTransaction(c, userID, input.Amount, input.Currency, input.Wallet)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, txData)
}
