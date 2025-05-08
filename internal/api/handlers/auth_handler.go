package handlers

import (
	"net/http"

	"github.com/TakuroBreath/wordle/internal/logger"
	"github.com/TakuroBreath/wordle/internal/models"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AuthHandler представляет обработчики для аутентификации
type AuthHandler struct {
	authService models.AuthService
	botToken    string
}

// NewAuthHandler создает новый экземпляр AuthHandler
func NewAuthHandler(authService models.AuthService, botToken string) *AuthHandler {
	logger.Log.Info("Creating new AuthHandler")
	return &AuthHandler{
		authService: authService,
		botToken:    botToken,
	}
}

// TelegramAuth обрабатывает аутентификацию через Telegram Mini App
func (h *AuthHandler) TelegramAuth(c *gin.Context) {
	log := getLoggerFromContext(c).With(zap.String("handler", "TelegramAuth"))
	log.Info("Processing Telegram authentication request")

	var input struct {
		InitData string `json:"init_data" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		log.Error("Failed to bind JSON input", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Debug("Received init_data", zap.String("init_data", input.InitData))

	// Генерируем JWT токен для клиента на основе данных инициализации
	log.Info("Generating JWT token based on initialization data")
	token, err := h.authService.InitAuth(c, input.InitData)
	if err != nil {
		log.Error("Failed to initialize authentication", zap.Error(err))
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	log.Info("Authentication successful, token generated")
	c.JSON(http.StatusOK, gin.H{"token": token})
}

// VerifyAuth проверяет валидность токена
func (h *AuthHandler) VerifyAuth(c *gin.Context) {
	log := getLoggerFromContext(c).With(zap.String("handler", "VerifyAuth"))
	log.Info("Processing token verification request")

	var input struct {
		Token string `json:"token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		log.Error("Failed to bind JSON input", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Debug("Verifying token")
	user, err := h.authService.VerifyAuth(c, input.Token)
	if err != nil {
		log.Error("Token verification failed", zap.Error(err))
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	log.Info("Token verification successful",
		zap.Uint64("telegram_id", user.TelegramID),
		zap.String("username", user.Username))

	c.JSON(http.StatusOK, gin.H{
		"authenticated": true,
		"user": gin.H{
			"telegram_id":  user.TelegramID,
			"username":     user.Username,
			"first_name":   user.FirstName,
			"last_name":    user.LastName,
			"wallet":       user.Wallet,
			"balance_ton":  user.BalanceTon,
			"balance_usdt": user.BalanceUsdt,
		},
	})
}

// Logout обрабатывает выход пользователя
func (h *AuthHandler) Logout(c *gin.Context) {
	log := getLoggerFromContext(c).With(zap.String("handler", "Logout"))
	log.Info("Processing logout request")

	var input struct {
		Token string `json:"token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		log.Error("Failed to bind JSON input", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Debug("Attempting to logout user")
	if err := h.authService.Logout(c, input.Token); err != nil {
		log.Error("Logout failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Info("Logout successful")
	c.JSON(http.StatusOK, gin.H{"message": "Successfully logged out"})
}

// getLoggerFromContext получает логгер из контекста Gin или создает новый, если он отсутствует
func getLoggerFromContext(c *gin.Context) *zap.Logger {
	if l, exists := c.Get("logger"); exists {
		return l.(*zap.Logger)
	}
	return logger.GetLogger()
}
