package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/TakuroBreath/wordle/internal/api/middleware"
	"github.com/TakuroBreath/wordle/internal/logger"
	"github.com/TakuroBreath/wordle/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// LobbyHandler представляет обработчики для лобби
type LobbyHandler struct {
	lobbyService models.LobbyService
	gameService  models.GameService
	userService  models.UserService
}

// NewLobbyHandler создает новый экземпляр LobbyHandler
func NewLobbyHandler(
	lobbyService models.LobbyService,
	gameService models.GameService,
	userService models.UserService,
) *LobbyHandler {
	return &LobbyHandler{
		lobbyService: lobbyService,
		gameService:  gameService,
		userService:  userService,
	}
}

// JoinGame присоединяет пользователя к игре
func (h *LobbyHandler) JoinGame(c *gin.Context) {
	log := logger.GetLogger().With(zap.String("handler", "JoinGame"))
	log.Info("JoinGame handler called")

	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		log.Error("User not found in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}
	log.Info("User ID from context", zap.Uint64("user_id", userID))

	var input struct {
		GameID    uuid.UUID `json:"game_id" binding:"required"`
		BetAmount float64   `json:"bet_amount" binding:"required,gt=0"`
		Currency  string    `json:"currency" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		log.Error("Invalid request data", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	log.Info("Request data",
		zap.String("game_id", input.GameID.String()),
		zap.Float64("bet_amount", input.BetAmount))

	// Проверяем, что игра существует и активна
	log.Info("Getting game", zap.String("game_id", input.GameID.String()))
	game, err := h.gameService.GetGame(c, input.GameID)
	if err != nil {
		log.Error("Game not found", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "game not found"})
		return
	}
	log.Info("Game found",
		zap.String("id", game.ID.String()),
		zap.String("title", game.Title),
		zap.String("status", game.Status))

	if game.Status != models.GameStatusActive {
		log.Error("Game is not active",
			zap.String("game_id", game.ID.String()),
			zap.String("status", game.Status))
		c.JSON(http.StatusBadRequest, gin.H{"error": "game is not active"})
		return
	}

	// Проверяем, что ставка в допустимых пределах
	if input.BetAmount < game.MinBet || input.BetAmount > game.MaxBet {
		log.Error("Bet amount out of range",
			zap.Float64("bet_amount", input.BetAmount),
			zap.Float64("min_bet", game.MinBet),
			zap.Float64("max_bet", game.MaxBet))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "bet amount is out of range",
			"min_bet": game.MinBet,
			"max_bet": game.MaxBet,
		})
		return
	}

	// Создаем лобби
	log.Info("Creating lobby",
		zap.Uint64("user_id", userID),
		zap.String("game_id", input.GameID.String()),
		zap.Float64("bet_amount", input.BetAmount))
	lobby := &models.Lobby{
		GameID:    input.GameID,
		UserID:    userID,
		BetAmount: input.BetAmount,
	}

	// Пробуем создать лобби напрямую через репозиторий для отладки
	log.Info("Trying direct repository access for debugging")

	// Проверяем, есть ли активное лобби
	activeLobby, err := h.lobbyService.GetActiveLobbyByGameAndUser(c, input.GameID, userID)
	if err != nil {
		if errors.Is(err, models.ErrLobbyNotFound) {
			log.Info("No active lobby found, can proceed with creation")
		} else {
			log.Error("Error checking for active lobby",
				zap.Error(err),
				zap.String("error_type", fmt.Sprintf("%T", err)))
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to check for existing active lobby: %v", err)})
			return
		}
	} else {
		log.Error("User already has active lobby", zap.String("lobby_id", activeLobby.ID.String()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "user already has an active lobby for this game"})
		return
	}

	if err := h.lobbyService.CreateLobby(c, lobby); err != nil {
		log.Error("Failed to create lobby", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Info("Lobby created successfully",
		zap.String("id", lobby.ID.String()),
		zap.String("game_id", lobby.GameID.String()),
		zap.Uint64("user_id", lobby.UserID))
	c.JSON(http.StatusCreated, gin.H{
		"id":               lobby.ID,
		"game_id":          lobby.GameID,
		"user_id":          lobby.UserID,
		"max_tries":        lobby.MaxTries,
		"tries_used":       lobby.TriesUsed,
		"bet_amount":       lobby.BetAmount,
		"potential_reward": lobby.PotentialReward,
		"status":           lobby.Status,
		"expires_at":       lobby.ExpiresAt,
	})
}

// GetLobby получает информацию о лобби
func (h *LobbyHandler) GetLobby(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid lobby ID"})
		return
	}

	lobby, err := h.lobbyService.GetLobby(c, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "lobby not found"})
		return
	}

	// Проверяем, что пользователь имеет доступ к этому лобби
	if lobby.UserID != userID {
		// Получаем игру, чтобы проверить, является ли пользователь создателем игры
		game, err := h.gameService.GetGame(c, lobby.GameID)
		if err != nil || game.CreatorID != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied to this lobby"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"id":               lobby.ID,
		"game_id":          lobby.GameID,
		"user_id":          lobby.UserID,
		"max_tries":        lobby.MaxTries,
		"tries_used":       lobby.TriesUsed,
		"bet_amount":       lobby.BetAmount,
		"potential_reward": lobby.PotentialReward,
		"status":           lobby.Status,
		"expires_at":       lobby.ExpiresAt,
		"created_at":       lobby.CreatedAt,
		"updated_at":       lobby.UpdatedAt,
		"attempts":         lobby.Attempts,
	})
}

// GetActiveLobby получает активное лобби пользователя
func (h *LobbyHandler) GetActiveLobby(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}

	gameIDStr := c.Query("game_id")
	if gameIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "game_id is required"})
		return
	}

	gameID, err := uuid.Parse(gameIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid game ID"})
		return
	}

	lobby, err := h.lobbyService.GetActiveLobbyByGameAndUser(c, gameID, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "active lobby not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":               lobby.ID,
		"game_id":          lobby.GameID,
		"user_id":          lobby.UserID,
		"max_tries":        lobby.MaxTries,
		"tries_used":       lobby.TriesUsed,
		"bet_amount":       lobby.BetAmount,
		"potential_reward": lobby.PotentialReward,
		"status":           lobby.Status,
		"expires_at":       lobby.ExpiresAt,
		"created_at":       lobby.CreatedAt,
		"updated_at":       lobby.UpdatedAt,
		"attempts":         lobby.Attempts,
	})
}

// MakeAttempt отправляет попытку угадать слово
func (h *LobbyHandler) MakeAttempt(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid lobby ID"})
		return
	}

	var input struct {
		Word string `json:"word" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Проверяем, что лобби существует
	lobby, err := h.lobbyService.GetLobby(c, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "lobby not found"})
		return
	}

	// Проверяем, что пользователь имеет доступ к этому лобби
	if lobby.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied to this lobby"})
		return
	}

	// Проверяем, что лобби активно
	if lobby.Status != models.LobbyStatusActive {
		c.JSON(http.StatusBadRequest, gin.H{"error": "lobby is not active"})
		return
	}

	// Обрабатываем попытку
	result, err := h.lobbyService.ProcessAttempt(c, id, input.Word)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Обновляем данные лобби после попытки
	updatedLobby, err := h.lobbyService.GetLobby(c, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error getting updated lobby"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"result":     result,
		"tries_used": updatedLobby.TriesUsed,
		"status":     updatedLobby.Status,
		"is_correct": isAllCorrect(result),
	})
}

// GetAttempts получает историю попыток для лобби
func (h *LobbyHandler) GetAttempts(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid lobby ID"})
		return
	}

	// Проверяем, что лобби существует
	lobby, err := h.lobbyService.GetLobby(c, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "lobby not found"})
		return
	}

	// Проверяем, что пользователь имеет доступ к этому лобби
	if lobby.UserID != userID {
		// Получаем игру, чтобы проверить, является ли пользователь создателем игры
		game, err := h.gameService.GetGame(c, lobby.GameID)
		if err != nil || game.CreatorID != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied to this lobby"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"attempts":   lobby.Attempts,
		"tries_used": lobby.TriesUsed,
		"max_tries":  lobby.MaxTries,
		"status":     lobby.Status,
	})
}

// GetUserLobbies получает список лобби пользователя
func (h *LobbyHandler) GetUserLobbies(c *gin.Context) {
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

	lobbies, err := h.lobbyService.GetUserLobbies(c, userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user lobbies"})
		return
	}

	result := make([]gin.H, 0, len(lobbies))
	for _, lobby := range lobbies {
		result = append(result, gin.H{
			"id":               lobby.ID,
			"game_id":          lobby.GameID,
			"max_tries":        lobby.MaxTries,
			"tries_used":       lobby.TriesUsed,
			"bet_amount":       lobby.BetAmount,
			"potential_reward": lobby.PotentialReward,
			"status":           lobby.Status,
			"created_at":       lobby.CreatedAt,
			"expires_at":       lobby.ExpiresAt,
		})
	}

	c.JSON(http.StatusOK, result)
}

// ExtendLobbyTime продлевает время жизни лобби
func (h *LobbyHandler) ExtendLobbyTime(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid lobby ID"})
		return
	}

	var input struct {
		Duration int `json:"duration" binding:"required,min=1,max=30"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Проверяем, что лобби существует
	lobby, err := h.lobbyService.GetLobby(c, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "lobby not found"})
		return
	}

	// Проверяем, что пользователь имеет доступ к этому лобби
	if lobby.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied to this lobby"})
		return
	}

	// Проверяем, что лобби активно
	if lobby.Status != models.LobbyStatusActive {
		c.JSON(http.StatusBadRequest, gin.H{"error": "lobby is not active"})
		return
	}

	if err := h.lobbyService.ExtendLobbyTime(c, id, input.Duration); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	updatedLobby, err := h.lobbyService.GetLobby(c, id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "Lobby time extended successfully"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Lobby time extended successfully",
		"expires_at": updatedLobby.ExpiresAt,
	})
}

// Вспомогательная функция для проверки, все ли буквы на правильных местах
func isAllCorrect(result []int) bool {
	for _, r := range result {
		if r != 2 {
			return false
		}
	}
	return true
}
