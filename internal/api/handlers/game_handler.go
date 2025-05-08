package handlers

import (
	"net/http"
	"strconv"

	"github.com/TakuroBreath/wordle/internal/api/middleware"
	"github.com/TakuroBreath/wordle/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GameHandler представляет обработчики для игр
type GameHandler struct {
	gameService        models.GameService
	userService        models.UserService
	transactionService models.TransactionService
}

// NewGameHandler создает новый экземпляр GameHandler
func NewGameHandler(
	gameService models.GameService,
	userService models.UserService,
	transactionService models.TransactionService,
) *GameHandler {
	return &GameHandler{
		gameService:        gameService,
		userService:        userService,
		transactionService: transactionService,
	}
}

// CreateGame создает новую игру
func (h *GameHandler) CreateGame(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}

	var input struct {
		Word             string  `json:"word" binding:"required"`
		Length           int     `json:"length" binding:"required,min=3,max=10"`
		Difficulty       string  `json:"difficulty" binding:"required"`
		MaxTries         int     `json:"max_tries" binding:"required,min=1,max=10"`
		Title            string  `json:"title" binding:"required"`
		Description      string  `json:"description"`
		MinBet           float64 `json:"min_bet" binding:"required,gt=0"`
		MaxBet           float64 `json:"max_bet" binding:"required,gt=0"`
		RewardMultiplier float64 `json:"reward_multiplier" binding:"required,gte=1"`
		Currency         string  `json:"currency" binding:"required"`
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

	// Проверка валидности сложности
	if input.Difficulty != "easy" && input.Difficulty != "medium" && input.Difficulty != "hard" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid difficulty, must be easy, medium or hard"})
		return
	}

	// Проверка длины слова
	if len([]rune(input.Word)) != input.Length {
		c.JSON(http.StatusBadRequest, gin.H{"error": "word length does not match the specified length"})
		return
	}

	// Проверка соотношения ставок
	if input.MinBet > input.MaxBet {
		c.JSON(http.StatusBadRequest, gin.H{"error": "min_bet cannot be greater than max_bet"})
		return
	}

	game := &models.Game{
		CreatorID:        userID,
		Word:             input.Word,
		Length:           input.Length,
		Difficulty:       input.Difficulty,
		MaxTries:         input.MaxTries,
		Title:            input.Title,
		Description:      input.Description,
		MinBet:           input.MinBet,
		MaxBet:           input.MaxBet,
		RewardMultiplier: input.RewardMultiplier,
		Currency:         input.Currency,
		Status:           models.GameStatusInactive,
	}

	if err := h.gameService.CreateGame(c, game); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":                game.ID,
		"creator_id":        game.CreatorID,
		"word_length":       game.Length,
		"difficulty":        game.Difficulty,
		"max_tries":         game.MaxTries,
		"title":             game.Title,
		"description":       game.Description,
		"min_bet":           game.MinBet,
		"max_bet":           game.MaxBet,
		"reward_multiplier": game.RewardMultiplier,
		"currency":          game.Currency,
		"status":            game.Status,
	})
}

// GetGame получает информацию об игре
func (h *GameHandler) GetGame(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid game ID"})
		return
	}

	game, err := h.gameService.GetGame(c, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "game not found"})
		return
	}

	// Скрываем слово для всех, кроме создателя
	userID, exists := middleware.GetCurrentUserID(c)
	isCreator := exists && userID == game.CreatorID
	wordInfo := gin.H{"length": game.Length}
	if isCreator {
		wordInfo["word"] = game.Word
	}

	c.JSON(http.StatusOK, gin.H{
		"id":                game.ID,
		"creator_id":        game.CreatorID,
		"word":              wordInfo,
		"difficulty":        game.Difficulty,
		"max_tries":         game.MaxTries,
		"title":             game.Title,
		"description":       game.Description,
		"min_bet":           game.MinBet,
		"max_bet":           game.MaxBet,
		"reward_multiplier": game.RewardMultiplier,
		"currency":          game.Currency,
		"reward_pool_ton":   game.RewardPoolTon,
		"reward_pool_usdt":  game.RewardPoolUsdt,
		"status":            game.Status,
		"created_at":        game.CreatedAt,
	})
}

// GetActiveGames получает список активных игр
func (h *GameHandler) GetActiveGames(c *gin.Context) {
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

	games, err := h.gameService.GetActiveGames(c, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get active games"})
		return
	}

	result := make([]gin.H, 0, len(games))
	for _, game := range games {
		result = append(result, gin.H{
			"id":                game.ID,
			"creator_id":        game.CreatorID,
			"word_length":       game.Length,
			"difficulty":        game.Difficulty,
			"max_tries":         game.MaxTries,
			"title":             game.Title,
			"description":       game.Description,
			"min_bet":           game.MinBet,
			"max_bet":           game.MaxBet,
			"reward_multiplier": game.RewardMultiplier,
			"currency":          game.Currency,
			"reward_pool_ton":   game.RewardPoolTon,
			"reward_pool_usdt":  game.RewardPoolUsdt,
			"created_at":        game.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, result)
}

// GetUserGames получает список игр, созданных пользователем
func (h *GameHandler) GetUserGames(c *gin.Context) {
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

	games, err := h.gameService.GetUserGames(c, userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user games"})
		return
	}

	result := make([]gin.H, 0, len(games))
	for _, game := range games {
		result = append(result, gin.H{
			"id":                game.ID,
			"word":              game.Word, // Создатель может видеть слово
			"word_length":       game.Length,
			"difficulty":        game.Difficulty,
			"max_tries":         game.MaxTries,
			"title":             game.Title,
			"description":       game.Description,
			"min_bet":           game.MinBet,
			"max_bet":           game.MaxBet,
			"reward_multiplier": game.RewardMultiplier,
			"currency":          game.Currency,
			"reward_pool_ton":   game.RewardPoolTon,
			"reward_pool_usdt":  game.RewardPoolUsdt,
			"status":            game.Status,
			"created_at":        game.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, result)
}

// DeleteGame удаляет игру
func (h *GameHandler) DeleteGame(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid game ID"})
		return
	}

	game, err := h.gameService.GetGame(c, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "game not found"})
		return
	}

	// Проверка, что пользователь является создателем игры
	if game.CreatorID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "only the creator can delete the game"})
		return
	}

	if err := h.gameService.DeleteGame(c, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Game deleted successfully"})
}

// AddToRewardPool добавляет средства в reward pool игры
func (h *GameHandler) AddToRewardPool(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid game ID"})
		return
	}

	var input struct {
		Amount float64 `json:"amount" binding:"required,gt=0"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	game, err := h.gameService.GetGame(c, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "game not found"})
		return
	}

	// Проверка, что пользователь является создателем игры
	if game.CreatorID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "only the creator can add to reward pool"})
		return
	}

	// Проверка, что у пользователя достаточно средств
	hasBalance, err := h.userService.ValidateBalance(c, userID, input.Amount, game.Currency)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to validate balance"})
		return
	}
	if !hasBalance {
		c.JSON(http.StatusBadRequest, gin.H{"error": "insufficient balance"})
		return
	}

	// Списание средств с баланса пользователя
	if game.Currency == models.CurrencyTON {
		err = h.userService.UpdateTonBalance(c, userID, -input.Amount)
	} else {
		err = h.userService.UpdateUsdtBalance(c, userID, -input.Amount)
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update user balance"})
		return
	}

	// Добавление средств в reward pool
	if err := h.gameService.AddToRewardPool(c, id, input.Amount); err != nil {
		// Возвращаем средства пользователю при ошибке
		if game.Currency == models.CurrencyTON {
			_ = h.userService.UpdateTonBalance(c, userID, input.Amount)
		} else {
			_ = h.userService.UpdateUsdtBalance(c, userID, input.Amount)
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Создаем транзакцию
	tx := &models.Transaction{
		UserID:      userID,
		Type:        models.TransactionTypeDeposit,
		Amount:      input.Amount,
		Currency:    game.Currency,
		Status:      models.TransactionStatusCompleted,
		Description: "Deposit to game reward pool",
		GameID:      &id,
	}
	if err := h.transactionService.CreateTransaction(c, tx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "reward pool updated but failed to create transaction record"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Added to reward pool successfully"})
}

// ActivateGame активирует игру
func (h *GameHandler) ActivateGame(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid game ID"})
		return
	}

	game, err := h.gameService.GetGame(c, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "game not found"})
		return
	}

	// Проверка, что пользователь является создателем игры
	if game.CreatorID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "only the creator can activate the game"})
		return
	}

	if err := h.gameService.ActivateGame(c, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Game activated successfully"})
}

// DeactivateGame деактивирует игру
func (h *GameHandler) DeactivateGame(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid game ID"})
		return
	}

	game, err := h.gameService.GetGame(c, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "game not found"})
		return
	}

	// Проверка, что пользователь является создателем игры
	if game.CreatorID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "only the creator can deactivate the game"})
		return
	}

	if err := h.gameService.DeactivateGame(c, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Game deactivated successfully"})
}

// SearchGames осуществляет поиск игр по параметрам
func (h *GameHandler) SearchGames(c *gin.Context) {
	minBet := 0.0
	maxBet := 1000000.0 // Высокое значение по умолчанию
	difficulty := ""
	limit := 10
	offset := 0

	if minBetStr := c.Query("min_bet"); minBetStr != "" {
		if val, err := strconv.ParseFloat(minBetStr, 64); err == nil && val > 0 {
			minBet = val
		}
	}

	if maxBetStr := c.Query("max_bet"); maxBetStr != "" {
		if val, err := strconv.ParseFloat(maxBetStr, 64); err == nil && val > 0 {
			maxBet = val
		}
	}

	difficulty = c.Query("difficulty")

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

	games, err := h.gameService.SearchGames(c, minBet, maxBet, difficulty, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to search games"})
		return
	}

	result := make([]gin.H, 0, len(games))
	for _, game := range games {
		result = append(result, gin.H{
			"id":                game.ID,
			"creator_id":        game.CreatorID,
			"word_length":       game.Length,
			"difficulty":        game.Difficulty,
			"max_tries":         game.MaxTries,
			"title":             game.Title,
			"description":       game.Description,
			"min_bet":           game.MinBet,
			"max_bet":           game.MaxBet,
			"reward_multiplier": game.RewardMultiplier,
			"currency":          game.Currency,
			"reward_pool_ton":   game.RewardPoolTon,
			"reward_pool_usdt":  game.RewardPoolUsdt,
			"created_at":        game.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, result)
}
