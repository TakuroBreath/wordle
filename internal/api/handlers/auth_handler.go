package handlers

import (
	"net/http"

	"github.com/TakuroBreath/wordle/internal/models"
	"github.com/gin-gonic/gin"
)

// AuthHandler представляет обработчики для аутентификации
type AuthHandler struct {
	authService models.AuthService
}

// NewAuthHandler создает новый экземпляр AuthHandler
func NewAuthHandler(authService models.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// TelegramAuth обрабатывает аутентификацию через Telegram Mini App
func (h *AuthHandler) TelegramAuth(c *gin.Context) {
	var input struct {
		InitData string `json:"init_data" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := h.authService.InitAuth(c, input.InitData)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

// VerifyAuth проверяет валидность токена
func (h *AuthHandler) VerifyAuth(c *gin.Context) {
	var input struct {
		Token string `json:"token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.authService.VerifyAuth(c, input.Token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"authenticated": true,
		"user": gin.H{
			"telegram_id": user.TelegramID,
			"username":    user.Username,
			"first_name":  user.FirstName,
			"last_name":   user.LastName,
			"wallet":      user.Wallet,
		},
	})
}

// Logout обрабатывает выход пользователя
func (h *AuthHandler) Logout(c *gin.Context) {
	var input struct {
		Token string `json:"token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.authService.Logout(c, input.Token); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully logged out"})
}
