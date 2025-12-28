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

// AuthConfig конфигурация авторизации
type AuthConfig struct {
	Enabled      bool
	BotToken     string
	DefaultUser  *models.User // Пользователь по умолчанию для dev режима
}

// AuthMiddleware представляет middleware для аутентификации
type AuthMiddleware struct {
	authService models.AuthService
	config      AuthConfig
	logger      *zap.Logger
}

// NewAuthMiddleware создает новый middleware для аутентификации
func NewAuthMiddleware(authService models.AuthService, botToken string) *AuthMiddleware {
	return NewAuthMiddlewareWithConfig(authService, AuthConfig{
		Enabled:  true,
		BotToken: botToken,
	})
}

// NewAuthMiddlewareWithConfig создает новый middleware с полной конфигурацией
func NewAuthMiddlewareWithConfig(authService models.AuthService, config AuthConfig) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
		config:      config,
		logger:      logger.GetLogger(zap.String("middleware", "auth")),
	}
}

// DefaultDevUser возвращает пользователя по умолчанию для dev режима
func DefaultDevUser() *models.User {
	return &models.User{
		TelegramID:  12345678,
		Username:    "dev_user",
		FirstName:   "Dev",
		LastName:    "User",
		BalanceTon:  100.0,
		BalanceUsdt: 100.0,
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

		// Если авторизация отключена, используем dev пользователя
		if !m.config.Enabled {
			log.Debug("Auth disabled, using default user")
			user := m.config.DefaultUser
			if user == nil {
				user = DefaultDevUser()
			}
			c.Set("user", user)
			c.Next()
			return
		}

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
			if err := initdata.Validate(tokenString, m.config.BotToken, time.Hour); err != nil {
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

// OptionalAuth проверяет авторизацию, но не требует её
// Если пользователь авторизован - добавляет его в контекст
// Если нет - продолжает без пользователя
func (m *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		log := logger.GetLogger(
			zap.String("middleware", "auth"),
			zap.String("handler", "OptionalAuth"),
		)

		// Если авторизация отключена, используем dev пользователя
		if !m.config.Enabled {
			user := m.config.DefaultUser
			if user == nil {
				user = DefaultDevUser()
			}
			c.Set("user", user)
			c.Next()
			return
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// Нет авторизации - продолжаем без пользователя
			c.Next()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 {
			c.Next()
			return
		}

		authType := parts[0]
		tokenString := parts[1]

		var user *models.User
		var err error

		switch authType {
		case "Bearer":
			user, err = m.authService.VerifyAuth(c, tokenString)
		case "tma":
			if verr := initdata.Validate(tokenString, m.config.BotToken, time.Hour); verr == nil {
				jwtToken, jerr := m.authService.InitAuth(c, tokenString)
				if jerr == nil {
					user, err = m.authService.VerifyAuth(c, jwtToken)
				}
			}
		}

		if err == nil && user != nil {
			log.Debug("Optional auth: user authenticated",
				zap.Uint64("telegram_id", user.TelegramID))
			c.Set("user", user)
		}

		c.Next()
	}
}

// IsAuthEnabled возвращает, включена ли авторизация
func (m *AuthMiddleware) IsAuthEnabled() bool {
	return m.config.Enabled
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
	user, ok := GetCurrentUser(c)
	if !ok {
		return 0, false
	}
	return user.TelegramID, true
}
