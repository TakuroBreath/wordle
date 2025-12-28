package handlers

import (
	"net/http"
	"strconv"

	"github.com/TakuroBreath/wordle/internal/api/middleware"
	"github.com/TakuroBreath/wordle/internal/models"
	otel "github.com/TakuroBreath/wordle/pkg/tracing"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
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

// JoinGame присоединяет пользователя к игре (оплата с баланса)
func (h *LobbyHandler) JoinGame(c *gin.Context) {
	ctx, span := otel.StartSpan(c.Request.Context(), "handler.JoinGame")
	defer span.End()
	c.Request = c.Request.WithContext(ctx)

	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}

	span.SetAttributes(attribute.Int64("user_id", int64(userID)))

	var input struct {
		GameID    string  `json:"game_id" binding:"required"` // UUID или short_id
		BetAmount float64 `json:"bet_amount" binding:"required,gt=0"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Находим игру по ID или short_id
	var game *models.Game
	var err error

	gameUUID, parseErr := uuid.Parse(input.GameID)
	if parseErr != nil {
		game, err = h.gameService.GetGameByShortID(c, input.GameID)
	} else {
		game, err = h.gameService.GetGame(c, gameUUID)
	}

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "game not found"})
		return
	}

	span.SetAttributes(attribute.String("game_id", game.ID.String()))

	if game.Status != models.GameStatusActive {
		c.JSON(http.StatusBadRequest, gin.H{"error": "game is not active"})
		return
	}

	// Проверяем ставку
	if input.BetAmount < game.MinBet || input.BetAmount > game.MaxBet {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "bet amount is out of range",
			"min_bet": game.MinBet,
			"max_bet": game.MaxBet,
		})
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
			"lobby":    formatLobbyResponse(existingLobby),
		})
		return
	}

	// Создаём лобби
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
		"game_id":          lobby.GameID,
		"game_short_id":    game.ShortID,
		"user_id":          lobby.UserID,
		"max_tries":        lobby.MaxTries,
		"tries_used":       lobby.TriesUsed,
		"bet_amount":       lobby.BetAmount,
		"potential_reward": lobby.PotentialReward,
		"currency":         lobby.Currency,
		"status":           lobby.Status,
		"expires_at":       lobby.ExpiresAt,
		"remaining_time":   lobby.GetRemainingTime(),
	})
}

// GetLobby получает информацию о лобби
func (h *LobbyHandler) GetLobby(c *gin.Context) {
	ctx, span := otel.StartSpan(c.Request.Context(), "handler.GetLobby")
	defer span.End()
	c.Request = c.Request.WithContext(ctx)

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

	// Проверяем доступ
	if lobby.UserID != userID {
		game, err := h.gameService.GetGame(c, lobby.GameID)
		if err != nil || game.CreatorID != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}
	}

	c.JSON(http.StatusOK, formatLobbyResponse(lobby))
}

// GetActiveLobby получает активное лобби пользователя для игры
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

	// Пробуем как UUID или short_id
	var game *models.Game
	var err error

	gameUUID, parseErr := uuid.Parse(gameIDStr)
	if parseErr != nil {
		game, err = h.gameService.GetGameByShortID(c, gameIDStr)
	} else {
		game, err = h.gameService.GetGame(c, gameUUID)
	}

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "game not found"})
		return
	}

	lobby, err := h.lobbyService.GetActiveLobbyByGameAndUser(c, game.ID, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "active lobby not found"})
		return
	}

	c.JSON(http.StatusOK, formatLobbyResponse(lobby))
}

// MakeAttempt отправляет попытку угадать слово
func (h *LobbyHandler) MakeAttempt(c *gin.Context) {
	ctx, span := otel.StartSpan(c.Request.Context(), "handler.MakeAttempt")
	defer span.End()
	c.Request = c.Request.WithContext(ctx)

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

	span.SetAttributes(
		attribute.String("lobby_id", id.String()),
		attribute.String("word", input.Word),
	)

	// Получаем лобби
	lobby, err := h.lobbyService.GetLobby(c, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "lobby not found"})
		return
	}

	// Проверяем доступ
	if lobby.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	// Проверяем статус
	if lobby.Status != models.LobbyStatusActive {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  "lobby is not active",
			"status": lobby.Status,
		})
		return
	}

	// Проверяем время
	if lobby.IsExpired() {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  "lobby time expired",
			"status": models.LobbyStatusFailedExpired,
		})
		return
	}

	// Обрабатываем попытку
	result, err := h.lobbyService.ProcessAttempt(c, id, input.Word)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Получаем обновлённое лобби
	updatedLobby, err := h.lobbyService.GetLobby(c, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get updated lobby"})
		return
	}

	isCorrect := isAllCorrect(result)

	response := gin.H{
		"result":         result,
		"word":           input.Word,
		"tries_used":     updatedLobby.TriesUsed,
		"tries_left":     updatedLobby.MaxTries - updatedLobby.TriesUsed,
		"status":         updatedLobby.Status,
		"is_correct":     isCorrect,
		"remaining_time": updatedLobby.GetRemainingTime(),
	}

	// Если игра завершена, добавляем информацию о результате
	if updatedLobby.Status != models.LobbyStatusActive {
		if updatedLobby.Status == models.LobbyStatusSuccess {
			response["reward"] = updatedLobby.PotentialReward
			response["message"] = "Congratulations! You won!"
		} else {
			response["message"] = "Game over"
		}

		// Получаем игру для показа правильного слова
		game, err := h.gameService.GetGame(c, lobby.GameID)
		if err == nil {
			response["correct_word"] = game.Word
		}
	}

	c.JSON(http.StatusOK, response)
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

	lobby, err := h.lobbyService.GetLobby(c, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "lobby not found"})
		return
	}

	// Проверяем доступ
	if lobby.UserID != userID {
		game, err := h.gameService.GetGame(c, lobby.GameID)
		if err != nil || game.CreatorID != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}
	}

	attempts := make([]gin.H, 0, len(lobby.Attempts))
	for _, a := range lobby.Attempts {
		attempts = append(attempts, gin.H{
			"id":         a.ID,
			"word":       a.Word,
			"result":     a.Result,
			"created_at": a.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"attempts":       attempts,
		"tries_used":     lobby.TriesUsed,
		"tries_left":     lobby.MaxTries - lobby.TriesUsed,
		"max_tries":      lobby.MaxTries,
		"status":         lobby.Status,
		"remaining_time": lobby.GetRemainingTime(),
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
			"game_short_id":    lobby.GameShortID,
			"max_tries":        lobby.MaxTries,
			"tries_used":       lobby.TriesUsed,
			"bet_amount":       lobby.BetAmount,
			"potential_reward": lobby.PotentialReward,
			"currency":         lobby.Currency,
			"status":           lobby.Status,
			"created_at":       lobby.CreatedAt,
			"expires_at":       lobby.ExpiresAt,
		})
	}

	c.JSON(http.StatusOK, result)
}

// CancelLobby отменяет активное лобби
func (h *LobbyHandler) CancelLobby(c *gin.Context) {
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

	if lobby.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	if lobby.Status != models.LobbyStatusActive {
		c.JSON(http.StatusBadRequest, gin.H{"error": "lobby is not active"})
		return
	}

	// Завершаем лобби как неудачное (игрок сдался)
	if err := h.lobbyService.FinishLobby(c, id, false); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Lobby cancelled successfully"})
}

// ExtendLobbyTime продлевает время жизни лобби (если это разрешено)
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
		Duration int `json:"duration" binding:"required,min=1,max=30"` // Минуты
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	lobby, err := h.lobbyService.GetLobby(c, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "lobby not found"})
		return
	}

	if lobby.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	if lobby.Status != models.LobbyStatusActive {
		c.JSON(http.StatusBadRequest, gin.H{"error": "lobby is not active"})
		return
	}

	// TODO: Возможно взимать плату за продление времени
	if err := h.lobbyService.ExtendLobbyTime(c, id, input.Duration); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	updatedLobby, err := h.lobbyService.GetLobby(c, id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "Lobby time extended"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":        "Lobby time extended",
		"expires_at":     updatedLobby.ExpiresAt,
		"remaining_time": updatedLobby.GetRemainingTime(),
	})
}

// formatLobbyResponse форматирует ответ для лобби
func formatLobbyResponse(lobby *models.Lobby) gin.H {
	attempts := make([]gin.H, 0, len(lobby.Attempts))
	for _, a := range lobby.Attempts {
		attempts = append(attempts, gin.H{
			"id":         a.ID,
			"word":       a.Word,
			"result":     a.Result,
			"created_at": a.CreatedAt,
		})
	}

	return gin.H{
		"id":               lobby.ID,
		"game_id":          lobby.GameID,
		"game_short_id":    lobby.GameShortID,
		"user_id":          lobby.UserID,
		"max_tries":        lobby.MaxTries,
		"tries_used":       lobby.TriesUsed,
		"tries_left":       lobby.MaxTries - lobby.TriesUsed,
		"bet_amount":       lobby.BetAmount,
		"potential_reward": lobby.PotentialReward,
		"currency":         lobby.Currency,
		"status":           lobby.Status,
		"expires_at":       lobby.ExpiresAt,
		"remaining_time":   lobby.GetRemainingTime(),
		"created_at":       lobby.CreatedAt,
		"updated_at":       lobby.UpdatedAt,
		"attempts":         attempts,
	}
}

// isAllCorrect проверяет, все ли буквы на правильных местах
func isAllCorrect(result []int) bool {
	for _, r := range result {
		if r != 2 {
			return false
		}
	}
	return true
}
