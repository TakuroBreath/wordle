package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/TakuroBreath/wordle/internal/logger"
	"github.com/TakuroBreath/wordle/internal/models"
	"github.com/gin-gonic/gin"
	initdata "github.com/telegram-mini-apps/init-data-golang"
	"go.uber.org/zap"
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
	logger      *zap.Logger
}

// NewAuthMiddleware создает новый middleware для аутентификации
func NewAuthMiddleware(authService models.AuthService, botToken string) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
		botToken:    botToken,
		logger:      logger.GetLogger(zap.String("middleware", "auth")),
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
		log := logger.GetLogger(
			zap.String("middleware", "auth"),
			zap.String("handler", "RequireAuth"),
			zap.String("path", c.Request.URL.Path),
			zap.String("method", c.Request.Method),
		)
		log.Info("Checking authentication")

		// Получаем токен из заголовка
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			log.Warn("No Authorization header provided")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization header is required"})
			c.Abort()
			return
		}

		// Проверяем формат токена
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 {
			log.Warn("Invalid Authorization header format", zap.String("header", authHeader))
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			c.Abort()
			return
		}

		authType := parts[0]
		tokenString := parts[1]

		log.Debug("Authorization header parsed",
			zap.String("type", authType),
			zap.Int("token_length", len(tokenString)))

		var user *models.User
		var err error

		// В зависимости от типа токена (JWT или TMA)
		switch authType {
		case "Bearer":
			// JWT токен
			log.Debug("Processing Bearer token")
			user, err = m.authService.VerifyAuth(c, tokenString)
			if err != nil {
				log.Error("Invalid Bearer token", zap.Error(err))
				c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
				c.Abort()
				return
			}
		case "tma":
			// Telegram Mini App token
			log.Debug("Processing TMA token", zap.Int("init_data_length", len(tokenString)))

			// Валидируем данные Telegram Mini App
			if err := initdata.Validate(tokenString, m.botToken, time.Hour); err != nil {
				log.Error("Invalid Telegram Mini App data", zap.Error(err))
				c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid telegram mini app data"})
				c.Abort()
				return
			}

			// Генерируем JWT токен для пользователя
			log.Debug("Initializing auth from Telegram Mini App data")
			jwtToken, err := m.authService.InitAuth(c, tokenString)
			if err != nil {
				log.Error("Failed to initialize auth from Telegram Mini App", zap.Error(err))
				c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
				c.Abort()
				return
			}

			// Верифицируем только что созданный JWT токен
			log.Debug("Verifying generated JWT token")
			user, err = m.authService.VerifyAuth(c, jwtToken)
			if err != nil {
				log.Error("Failed to verify auth token", zap.Error(err))
				c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
				c.Abort()
				return
			}
		default:
			log.Warn("Unsupported authentication type", zap.String("type", authType))
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unsupported authentication type"})
			c.Abort()
			return
		}

		// Устанавливаем пользователя в контекст
		log.Info("Authentication successful",
			zap.Uint64("telegram_id", user.TelegramID),
			zap.String("username", user.Username))
		c.Set("user", user)
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
