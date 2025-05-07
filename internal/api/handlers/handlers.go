package handlers

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/TakuroBreath/wordle/internal/models"
	"github.com/TakuroBreath/wordle/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler представляет собой структуру для обработчиков API
type Handler struct {
	services service.Service
}

// NewHandler создает новый экземпляр Handler
func NewHandler(services service.Service) *Handler {
	return &Handler{
		services: services,
	}
}

// RegisterRoutes регистрирует все маршруты API
func (h *Handler) RegisterRoutes(router *gin.Engine) {
	api := router.Group("/api")

	// Аутентификация - публичные маршруты
	auth := api.Group("/auth")
	{
		auth.POST("/init", h.initAuth)
		auth.POST("/verify", h.verifyAuth)
	}

	// Публичные маршруты для игр
	games := api.Group("/games")
	{
		games.GET("", h.getAllGames)
		games.GET("/:id", h.getGameByID)
	}

	// Защищенные маршруты - требуют аутентификации
	authMiddleware := h.authMiddleware()

	// Защищенные маршруты для игр
	protectedGames := api.Group("/games")
	// protectedGames.Use(authMiddleware)
	{
		protectedGames.POST("", h.createGame)
		protectedGames.PUT("/:id", h.updateGame)
		protectedGames.DELETE("/:id", h.deleteGame)
		protectedGames.POST("/:id/join", h.joinGame)
	}

	// Лобби - все маршруты защищены
	lobbies := api.Group("/lobbies")
	lobbies.Use(authMiddleware)
	{
		lobbies.GET("/:id", h.getLobbyByID)
		lobbies.POST("/:id/attempt", h.makeAttempt)
	}

	// Пользователи - все маршруты защищены
	user := api.Group("/user")
	user.Use(authMiddleware)
	{
		user.GET("", h.getCurrentUser)
		user.GET("/games", h.getUserGames)
		user.GET("/history", h.getUserHistory)
	}

	// Транзакции - все маршруты защищены
	transactions := api.Group("/transactions")
	transactions.Use(authMiddleware)
	{
		transactions.POST("/deposit", h.createDeposit)
		transactions.GET("", h.getTransactions)
	}
}

// Обработчики аутентификации
func (h *Handler) initAuth(c *gin.Context) {
	var request struct {
		InitData string `json:"init_data" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request data"})
		return
	}

	token, err := h.services.Auth().InitAuth(c, request.InitData)
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to initialize authentication: %v", err)})
		return
	}

	c.JSON(200, gin.H{"token": token})
}

func (h *Handler) verifyAuth(c *gin.Context) {
	var request struct {
		Token string `json:"token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request data"})
		return
	}

	user, err := h.services.Auth().VerifyAuth(c, request.Token)
	if err != nil {
		c.JSON(401, gin.H{"error": "Invalid token"})
		return
	}

	c.JSON(200, gin.H{"user": user})
}

// Обработчики игр
func (h *Handler) getAllGames(c *gin.Context) {
	limit := 100
	offset := 0

	// Получение параметров пагинации из запроса
	limitParam := c.Query("limit")
	offsetParam := c.Query("offset")

	if limitParam != "" {
		parsedLimit, err := strconv.Atoi(limitParam)
		if err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	if offsetParam != "" {
		parsedOffset, err := strconv.Atoi(offsetParam)
		if err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	games, err := h.services.Game().GetAll(c, limit, offset)
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to get games: %v", err)})
		return
	}

	c.JSON(200, gin.H{"games": games})
}

func (h *Handler) createGame(c *gin.Context) {
	var request struct {
		WordLength       int     `json:"word_length" binding:"required"`
		MaxAttempts      int     `json:"max_attempts" binding:"required"`
		Language         string  `json:"language" binding:"required"`
		Word             string  `json:"word" binding:"required"`
		MinBet           float64 `json:"min_bet" binding:"required"`
		MaxBet           float64 `json:"max_bet" binding:"required"`
		RewardMultiplier float64 `json:"reward_multiplier" binding:"required"`
		Currency         string  `json:"currency" binding:"required"`
		Title            *string `json:"title"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request data"})
		return
	}

	// user, exists := c.Get("user")
	// if !exists {
	// 	c.JSON(401, gin.H{"error": "Unauthorized"})
	// 	return
	// }

	game := &models.Game{
		ID:       uuid.New(),
		Title:    request.Title,
		Word:     request.Word,
		Length:   request.WordLength,
		MaxTries: request.MaxAttempts,
		Language: request.Language,
		// CreatorID:        user.(*models.User).ID,
		CreatorID:        "123",
		MinBet:           request.MinBet,
		MaxBet:           request.MaxBet,
		RewardMultiplier: request.RewardMultiplier,
		Currency:         request.Currency,
		Status:           "active",
		CreatedAt:        time.Now(),
	}

	if err := h.services.Game().Create(c, game); err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to create game: %v", err)})
		return
	}

	c.JSON(201, gin.H{"game": game})
}

func (h *Handler) getGameByID(c *gin.Context) {
	gameID := c.Param("id")
	if gameID == "" {
		c.JSON(400, gin.H{"error": "Game ID is required"})
		return
	}

	game, err := h.services.Game().GetByID(c, uuid.MustParse(gameID))
	if err != nil {
		c.JSON(404, gin.H{"error": "Game not found"})
		return
	}

	c.JSON(200, gin.H{"game": game})
}

func (h *Handler) updateGame(c *gin.Context) {
	gameID := c.Param("id")
	if gameID == "" {
		c.JSON(400, gin.H{"error": "Game ID is required"})
		return
	}

	var request struct {
		WordLength  int    `json:"word_length"`
		MaxAttempts int    `json:"max_attempts"`
		Language    string `json:"language"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request data"})
		return
	}

	user, exists := c.Get("user")
	if !exists {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	// Получаем существующую игру
	gameUUID, err := uuid.Parse(gameID)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid game ID format"})
		return
	}

	game, err := h.services.Game().GetByID(c, gameUUID)
	if err != nil {
		c.JSON(404, gin.H{"error": "Game not found"})
		return
	}

	// Проверяем, что пользователь является создателем игры
	if game.CreatorID != user.(*models.User).TelegramID {
		c.JSON(403, gin.H{"error": "You are not authorized to update this game"})
		return
	}

	// Обновляем поля игры
	if request.WordLength > 0 {
		game.Length = request.WordLength
	}
	if request.MaxAttempts > 0 {
		game.MaxTries = request.MaxAttempts
	}
	if request.Language != "" {
		game.Language = request.Language
	}

	// Сохраняем изменения
	err = h.services.Game().Update(c, game)
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to update game: %v", err)})
		return
	}

	c.JSON(200, gin.H{"message": "Game updated successfully", "game": game})
}

func (h *Handler) deleteGame(c *gin.Context) {
	gameID := c.Param("id")
	if gameID == "" {
		c.JSON(400, gin.H{"error": "Game ID is required"})
		return
	}

	user, exists := c.Get("user")
	if !exists {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	// Получаем существующую игру
	gameUUID, err := uuid.Parse(gameID)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid game ID format"})
		return
	}

	game, err := h.services.Game().GetByID(c, gameUUID)
	if err != nil {
		c.JSON(404, gin.H{"error": "Game not found"})
		return
	}

	// Проверяем, что пользователь является создателем игры
	if game.CreatorID != user.(*models.User).ID {
		c.JSON(403, gin.H{"error": "You are not authorized to delete this game"})
		return
	}

	// Удаляем игру
	err = h.services.Game().Delete(c, gameUUID)
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to delete game: %v", err)})
		return
	}

	c.JSON(200, gin.H{"message": "Game deleted successfully"})
}

func (h *Handler) joinGame(c *gin.Context) {
	gameID := c.Param("id")
	if gameID == "" {
		c.JSON(400, gin.H{"error": "Game ID is required"})
		return
	}

	var request struct {
		Bet float64 `json:"bet" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request data"})
		return
	}

	user, exists := c.Get("user")
	if !exists {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	gameUUID, err := uuid.Parse(gameID)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid game ID format"})
		return
	}

	lobby, err := h.services.Lobby().JoinGame(c, gameUUID, user.(*models.User).ID, request.Bet)
	if err != nil {
		c.JSON(403, gin.H{"error": fmt.Sprintf("Failed to join game: %v", err)})
		return
	}

	c.JSON(201, gin.H{"lobby": lobby})
}

// Обработчики лобби
func (h *Handler) getLobbyByID(c *gin.Context) {
	lobbyID := c.Param("id")
	if lobbyID == "" {
		c.JSON(400, gin.H{"error": "Lobby ID is required"})
		return
	}

	lobbyUUID, err := uuid.Parse(lobbyID)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid lobby ID format"})
		return
	}

	lobby, err := h.services.Lobby().GetByID(c, lobbyUUID)
	if err != nil {
		c.JSON(404, gin.H{"error": "Lobby not found"})
		return
	}

	c.JSON(200, gin.H{"lobby": lobby})
}

func (h *Handler) makeAttempt(c *gin.Context) {
	lobbyID := c.Param("id")
	if lobbyID == "" {
		c.JSON(400, gin.H{"error": "Lobby ID is required"})
		return
	}

	var request struct {
		Word string `json:"word" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request data"})
		return
	}

	user, exists := c.Get("user")
	if !exists {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	lobbyUUID, err := uuid.Parse(lobbyID)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid lobby ID format"})
		return
	}

	// Проверяем, что пользователь участвует в этом лобби
	lobby, err := h.services.Lobby().GetByID(c, lobbyUUID)
	if err != nil {
		c.JSON(404, gin.H{"error": "Lobby not found"})
		return
	}

	if lobby.UserID != user.(*models.User).ID {
		c.JSON(403, gin.H{"error": "You are not authorized to make attempts in this lobby"})
		return
	}

	result, err := h.services.Lobby().MakeAttempt(c, lobbyUUID, request.Word)
	if err != nil {
		c.JSON(403, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"result": result})
}

// Обработчики пользователей
func (h *Handler) getCurrentUser(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	c.JSON(200, gin.H{"user": user})
}

func (h *Handler) getUserGames(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	limit := 100
	offset := 0

	// Получение параметров пагинации из запроса
	limitParam := c.Query("limit")
	offsetParam := c.Query("offset")

	if limitParam != "" {
		parsedLimit, err := strconv.Atoi(limitParam)
		if err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	if offsetParam != "" {
		parsedOffset, err := strconv.Atoi(offsetParam)
		if err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	games, err := h.services.Game().GetByCreator(c, user.(*models.User).ID, limit, offset)
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to get user games: %v", err)})
		return
	}

	c.JSON(200, gin.H{"games": games})
}

func (h *Handler) getUserHistory(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	limit := 100
	offset := 0

	// Получение параметров пагинации из запроса
	limitParam := c.Query("limit")
	offsetParam := c.Query("offset")

	if limitParam != "" {
		parsedLimit, err := strconv.Atoi(limitParam)
		if err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	if offsetParam != "" {
		parsedOffset, err := strconv.Atoi(offsetParam)
		if err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	history, err := h.services.History().GetByUserID(c, user.(*models.User).ID, limit, offset)
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to get user history: %v", err)})
		return
	}

	c.JSON(200, gin.H{"history": history})
}

// Обработчики транзакций
func (h *Handler) createDeposit(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	var request struct {
		Amount   float64 `json:"amount" binding:"required"`
		Currency string  `json:"currency" binding:"required"`
		TxHash   string  `json:"tx_hash" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request data"})
		return
	}

	// Проверка валидности данных
	if request.Amount <= 0 {
		c.JSON(400, gin.H{"error": "Amount must be greater than 0"})
		return
	}

	if request.Currency != "TON" && request.Currency != "USDT" {
		c.JSON(400, gin.H{"error": "Currency must be TON or USDT"})
		return
	}

	err := h.services.Transaction().ProcessDeposit(c, user.(*models.User).ID, request.Amount, request.Currency, request.TxHash)
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to process deposit: %v", err)})
		return
	}

	c.JSON(200, gin.H{"message": "Deposit processed successfully"})
}

func (h *Handler) getTransactions(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	limit := 100
	offset := 0

	// Получение параметров пагинации из запроса
	limitParam := c.Query("limit")
	offsetParam := c.Query("offset")

	if limitParam != "" {
		parsedLimit, err := strconv.Atoi(limitParam)
		if err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	if offsetParam != "" {
		parsedOffset, err := strconv.Atoi(offsetParam)
		if err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	transactions, err := h.services.Transaction().GetByUserID(c, user.(*models.User).ID, limit, offset)
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to get transactions: %v", err)})
		return
	}

	c.JSON(200, gin.H{"transactions": transactions})
}

// authMiddleware создает middleware для проверки аутентификации
func (h *Handler) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Получение токена из заголовка Authorization
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(401, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return
		}

		// Проверка формата токена (Bearer token)
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(401, gin.H{"error": "Invalid authorization format, expected 'Bearer token'"})
			c.Abort()
			return
		}

		token := parts[1]

		// Проверка токена
		user, err := h.services.Auth().ValidateToken(c, token)
		if err != nil {
			c.JSON(401, gin.H{"error": fmt.Sprintf("Invalid token: %v", err)})
			c.Abort()
			return
		}

		// Сохранение пользователя в контексте
		c.Set("user", &user)
		c.Next()
	}
}
