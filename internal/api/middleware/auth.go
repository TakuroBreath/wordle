package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/TakuroBreath/wordle/internal/models"
	"github.com/gin-gonic/gin"
	initdata "github.com/telegram-mini-apps/init-data-golang"
)

type contextKey string

const (
	_initDataKey contextKey = "init-data"
	_userKey     contextKey = "user"
)

// AuthMiddleware представляет middleware для аутентификации
type AuthMiddleware struct {
	authService models.AuthService
	botToken    string
}

// NewAuthMiddleware создает новый middleware для аутентификации
func NewAuthMiddleware(authService models.AuthService, botToken string) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
		botToken:    botToken,
	}
}

// WithInitData добавляет данные инициализации TMA в контекст
func WithInitData(ctx context.Context, initData initdata.InitData) context.Context {
	return context.WithValue(ctx, _initDataKey, initData)
}

// GetInitDataFromContext извлекает данные инициализации TMA из контекста
func GetInitDataFromContext(ctx context.Context) (initdata.InitData, bool) {
	initData, ok := ctx.Value(_initDataKey).(initdata.InitData)
	return initData, ok
}

// RequireAuth проверяет авторизацию пользователя
func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Получаем заголовок Authorization
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authorization header is required"})
			return
		}

		// Извлекаем тип авторизации и данные
		authParts := strings.Split(authHeader, " ")
		if len(authParts) != 2 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization format"})
			return
		}

		authType := authParts[0]
		authData := authParts[1]

		switch authType {
		case "tma":
			// Проверяем инициализационные данные по официальному алгоритму TMA
			if err := initdata.Validate(authData, m.botToken, time.Hour); err != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
				return
			}

			// Парсим данные
			tmaData, err := initdata.Parse(authData)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			// Добавляем данные в контекст
			c.Request = c.Request.WithContext(WithInitData(c.Request.Context(), tmaData))

			// Получаем или создаем пользователя из данных Telegram
			user, err := m.authService.InitAuth(c, authData)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
				return
			}

			// Сохраняем пользователя в контексте
			c.Set("user", user)
			c.Set("user_id", user.TelegramID)

		case "Bearer":
			// Прежний метод JWT-авторизации
			user, err := m.authService.ValidateToken(c, authData)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
				return
			}

			// Сохраняем пользователя в контексте
			c.Set("user", user)
			c.Set("user_id", user.TelegramID)

		default:
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unsupported authorization type"})
			return
		}

		c.Next()
	}
}

// GetCurrentUser возвращает текущего пользователя из контекста
func GetCurrentUser(c *gin.Context) (*models.User, bool) {
	user, exists := c.Get("user")
	if !exists {
		return nil, false
	}

	u, ok := user.(*models.User)
	return u, ok
}

// GetCurrentUserID возвращает ID текущего пользователя из контекста
func GetCurrentUserID(c *gin.Context) (uint64, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}

	id, ok := userID.(uint64)
	return id, ok
}
