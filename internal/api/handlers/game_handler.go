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
	lobbyService       models.LobbyService
}

// NewGameHandler создает новый экземпляр GameHandler
func NewGameHandler(
	gameService models.GameService,
	userService models.UserService,
	transactionService models.TransactionService,
	lobbyService models.LobbyService,
) *GameHandler {
	return &GameHandler{
		gameService:        gameService,
		userService:        userService,
		transactionService: transactionService,
		lobbyService:       lobbyService,
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
		Difficulty       string  `json:"difficulty" binding:"required"`
		MaxTries         int     `json:"max_tries" binding:"required,min=1,max=20"`
		TimeLimit        int     `json:"time_limit" binding:"min=1,max=60"` // минуты
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

	// Дефолтное время
	if input.TimeLimit == 0 {
		input.TimeLimit = 5
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

	// Проверка соотношения ставок
	if input.MinBet > input.MaxBet {
		c.JSON(http.StatusBadRequest, gin.H{"error": "min_bet cannot be greater than max_bet"})
		return
	}

	// Вычисляем необходимый депозит
	depositAmount := input.MaxBet * input.RewardMultiplier

	game := &models.Game{
		CreatorID:        userID,
		Word:             input.Word,
		Length:           len([]rune(input.Word)),
		Difficulty:       input.Difficulty,
		MaxTries:         input.MaxTries,
		TimeLimit:        input.TimeLimit,
		Title:            input.Title,
		Description:      input.Description,
		MinBet:           input.MinBet,
		MaxBet:           input.MaxBet,
		RewardMultiplier: input.RewardMultiplier,
		DepositAmount:    depositAmount,
		Currency:         input.Currency,
		Status:           models.GameStatusPending,
	}

	if err := h.gameService.CreateGame(c, game); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Генерируем информацию для оплаты
	paymentInfo, err := h.gameService.GetPaymentInfo(c, game.ID)
	if err != nil {
		// Игра создана, но платёжная информация недоступна
		c.JSON(http.StatusCreated, gin.H{
			"id":                game.ID,
			"short_id":          game.ShortID,
			"creator_id":        game.CreatorID,
			"word_length":       game.Length,
			"difficulty":        game.Difficulty,
			"max_tries":         game.MaxTries,
			"time_limit":        game.TimeLimit,
			"title":             game.Title,
			"description":       game.Description,
			"min_bet":           game.MinBet,
			"max_bet":           game.MaxBet,
			"reward_multiplier": game.RewardMultiplier,
			"deposit_amount":    game.DepositAmount,
			"currency":          game.Currency,
			"status":            game.Status,
			"message":           "Game created. Please deposit to activate.",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":                game.ID,
		"short_id":          game.ShortID,
		"creator_id":        game.CreatorID,
		"word_length":       game.Length,
		"difficulty":        game.Difficulty,
		"max_tries":         game.MaxTries,
		"time_limit":        game.TimeLimit,
		"title":             game.Title,
		"description":       game.Description,
		"min_bet":           game.MinBet,
		"max_bet":           game.MaxBet,
		"reward_multiplier": game.RewardMultiplier,
		"deposit_amount":    game.DepositAmount,
		"currency":          game.Currency,
		"status":            game.Status,
		"payment":           paymentInfo,
	})
}

// GetGame получает информацию об игре
func (h *GameHandler) GetGame(c *gin.Context) {
	idStr := c.Param("id")

	// Пробуем сначала как UUID
	id, err := uuid.Parse(idStr)
	var game *models.Game
	
	if err != nil {
		// Пробуем как short_id
		game, err = h.gameService.GetGameByShortID(c, idStr)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "game not found"})
			return
		}
	} else {
		game, err = h.gameService.GetGame(c, id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "game not found"})
			return
		}
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
		"short_id":          game.ShortID,
		"creator_id":        game.CreatorID,
		"word":              wordInfo,
		"difficulty":        game.Difficulty,
		"max_tries":         game.MaxTries,
		"time_limit":        game.TimeLimit,
		"title":             game.Title,
		"description":       game.Description,
		"min_bet":           game.MinBet,
		"max_bet":           game.MaxBet,
		"reward_multiplier": game.RewardMultiplier,
		"deposit_amount":    game.DepositAmount,
		"currency":          game.Currency,
		"reward_pool_ton":   game.RewardPoolTon,
		"reward_pool_usdt":  game.RewardPoolUsdt,
		"reserved_amount":   game.ReservedAmount,
		"available_pool":    game.GetAvailableRewardPool(),
		"status":            game.Status,
		"created_at":        game.CreatedAt,
	})
}

// GetPaymentInfo получает информацию для оплаты депозита игры
func (h *GameHandler) GetPaymentInfo(c *gin.Context) {
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

	// Только создатель может получить платёжную информацию для депозита
	if game.CreatorID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "only the creator can get payment info"})
		return
	}

	if game.Status != models.GameStatusPending {
		c.JSON(http.StatusBadRequest, gin.H{"error": "game is not pending payment"})
		return
	}

	paymentInfo, err := h.gameService.GetPaymentInfo(c, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, paymentInfo)
}

// JoinGame получает информацию для вступления в игру
func (h *GameHandler) JoinGame(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}

	var input struct {
		GameID    string  `json:"game_id" binding:"required"` // Может быть UUID или short_id
		BetAmount float64 `json:"bet_amount" binding:"required,gt=0"`
		UseBalance bool   `json:"use_balance"` // Использовать баланс вместо блокчейн платежа
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Находим игру
	var game *models.Game
	var err error

	id, parseErr := uuid.Parse(input.GameID)
	if parseErr != nil {
		game, err = h.gameService.GetGameByShortID(c, input.GameID)
	} else {
		game, err = h.gameService.GetGame(c, id)
	}

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "game not found"})
		return
	}

	if game.Status != models.GameStatusActive {
		c.JSON(http.StatusBadRequest, gin.H{"error": "game is not active"})
		return
	}

	// Проверяем ставку
	if input.BetAmount < game.MinBet || input.BetAmount > game.MaxBet {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "bet amount out of range",
			"min_bet": game.MinBet,
			"max_bet": game.MaxBet,
		})
		return
	}

	// Проверяем, может ли игра принять ставку
	if !game.CanAcceptBet(input.BetAmount) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "game cannot accept bet: insufficient reward pool"})
		return
	}

	// Нельзя играть в свою игру
	if game.CreatorID == userID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot play your own game"})
		return
	}

	// Проверяем, нет ли уже активного лобби
	existingLobby, err := h.lobbyService.GetActiveLobbyByGameAndUser(c, game.ID, userID)
	if err == nil && existingLobby != nil {
		c.JSON(http.StatusConflict, gin.H{
			"error":    "you already have an active game session",
			"lobby_id": existingLobby.ID,
		})
		return
	}

	if input.UseBalance {
		// Создаём лобби с оплатой с баланса
		lobby := &models.Lobby{
			GameID:    game.ID,
			UserID:    userID,
			BetAmount: input.BetAmount,
		}

		if err := h.lobbyService.CreateLobby(c, lobby); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"lobby_id":         lobby.ID,
			"game_id":          game.ID,
			"game_short_id":    game.ShortID,
			"bet_amount":       lobby.BetAmount,
			"potential_reward": lobby.PotentialReward,
			"max_tries":        lobby.MaxTries,
			"time_limit":       game.TimeLimit,
			"expires_at":       lobby.ExpiresAt,
			"status":           lobby.Status,
		})
		return
	}

	// Генерируем платёжную информацию для блокчейн оплаты
	paymentInfo, err := h.lobbyService.GetJoinPaymentInfo(c, game.ShortID, userID, input.BetAmount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"payment":          paymentInfo,
		"game_id":          game.ID,
		"game_short_id":    game.ShortID,
		"bet_amount":       input.BetAmount,
		"potential_reward": input.BetAmount * game.RewardMultiplier,
		"max_tries":        game.MaxTries,
		"time_limit":       game.TimeLimit,
		"message":          "Complete the payment to start the game",
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

	// Получаем ID текущего пользователя для фильтрации своих игр
	userID, _ := middleware.GetCurrentUserID(c)

	result := make([]gin.H, 0, len(games))
	for _, game := range games {
		// Не показываем свои игры в списке доступных
		if game.CreatorID == userID {
			continue
		}

		result = append(result, gin.H{
			"id":                game.ID,
			"short_id":          game.ShortID,
			"creator_id":        game.CreatorID,
			"word_length":       game.Length,
			"difficulty":        game.Difficulty,
			"max_tries":         game.MaxTries,
			"time_limit":        game.TimeLimit,
			"title":             game.Title,
			"description":       game.Description,
			"min_bet":           game.MinBet,
			"max_bet":           game.MaxBet,
			"reward_multiplier": game.RewardMultiplier,
			"currency":          game.Currency,
			"available_pool":    game.GetAvailableRewardPool(),
			"created_at":        game.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, result)
}

// GetUserGames получает список игр пользователя
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
			"short_id":          game.ShortID,
			"word":              game.Word, // Создатель видит слово
			"word_length":       game.Length,
			"difficulty":        game.Difficulty,
			"max_tries":         game.MaxTries,
			"time_limit":        game.TimeLimit,
			"title":             game.Title,
			"description":       game.Description,
			"min_bet":           game.MinBet,
			"max_bet":           game.MaxBet,
			"reward_multiplier": game.RewardMultiplier,
			"currency":          game.Currency,
			"reward_pool_ton":   game.RewardPoolTon,
			"reward_pool_usdt":  game.RewardPoolUsdt,
			"reserved_amount":   game.ReservedAmount,
			"status":            game.Status,
			"created_at":        game.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, result)
}

// GetCreatedGames получает список игр, созданных пользователем
func (h *GameHandler) GetCreatedGames(c *gin.Context) {
	h.GetUserGames(c) // Алиас
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

	if game.CreatorID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "only the creator can activate the game"})
		return
	}

	if err := h.gameService.ActivateGame(c, id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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

	if game.CreatorID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "only the creator can deactivate the game"})
		return
	}

	if err := h.gameService.DeactivateGame(c, id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Game deactivated successfully"})
}

// SearchGames осуществляет поиск игр по параметрам
func (h *GameHandler) SearchGames(c *gin.Context) {
	minBet := 0.0
	maxBet := 1000000.0
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
			"short_id":          game.ShortID,
			"creator_id":        game.CreatorID,
			"word_length":       game.Length,
			"difficulty":        game.Difficulty,
			"max_tries":         game.MaxTries,
			"time_limit":        game.TimeLimit,
			"title":             game.Title,
			"description":       game.Description,
			"min_bet":           game.MinBet,
			"max_bet":           game.MaxBet,
			"reward_multiplier": game.RewardMultiplier,
			"currency":          game.Currency,
			"available_pool":    game.GetAvailableRewardPool(),
			"created_at":        game.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, result)
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

	if game.CreatorID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "only the creator can add to reward pool"})
		return
	}

	// Проверяем баланс
	hasBalance, err := h.userService.ValidateBalance(c, userID, input.Amount, game.Currency)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to validate balance"})
		return
	}
	if !hasBalance {
		c.JSON(http.StatusBadRequest, gin.H{"error": "insufficient balance"})
		return
	}

	// Списываем с баланса
	if game.Currency == models.CurrencyTON {
		err = h.userService.UpdateTonBalance(c, userID, -input.Amount)
	} else {
		err = h.userService.UpdateUsdtBalance(c, userID, -input.Amount)
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update balance"})
		return
	}

	// Добавляем в пул
	if err := h.gameService.AddToRewardPool(c, id, input.Amount); err != nil {
		// Возвращаем деньги
		if game.Currency == models.CurrencyTON {
			_ = h.userService.UpdateTonBalance(c, userID, input.Amount)
		} else {
			_ = h.userService.UpdateUsdtBalance(c, userID, input.Amount)
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Added to reward pool successfully"})
}
