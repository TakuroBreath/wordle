package routes

import (
	"sync"
	"time"

	"github.com/TakuroBreath/wordle/internal/api/handlers"
	"github.com/TakuroBreath/wordle/internal/service"
	"github.com/gin-gonic/gin"
)

// SetupRouter настраивает маршрутизацию для API
func SetupRouter(handler *handlers.Handler, authService service.AuthService) *gin.Engine {
	router := gin.Default()

	// Настройка CORS
	router.Use(corsMiddleware())

	// Настройка middleware для ограничения запросов
	router.Use(rateLimitMiddleware())

	// Регистрация обычных маршрутов API
	handler.RegisterRoutes(router)

	// Регистрация защищенных маршрутов API
	RegisterProtectedRoutes(router, handler, authService)

	return router
}

// RegisterProtectedRoutes регистрирует защищенные маршруты, требующие аутентификации
func RegisterProtectedRoutes(router *gin.Engine, handler *handlers.Handler, authService service.AuthService) {
	// Группа защищенных маршрутов
	protected := router.Group("/api/protected")
	protected.Use(authMiddleware(authService))

	// Здесь можно добавить защищенные маршруты, например:
	// protected.GET("/user/profile", handler.GetUserProfile)
	// protected.POST("/user/settings", handler.UpdateUserSettings)
}

// corsMiddleware настраивает CORS для API
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// rateLimitMiddleware настраивает ограничение запросов для API
func rateLimitMiddleware() gin.HandlerFunc {
	// Создаем простую карту для отслеживания запросов по IP
	visits := make(map[string][]time.Time)
	var mu sync.Mutex

	return func(c *gin.Context) {
		ip := c.ClientIP()
		mu.Lock()

		// Очищаем устаревшие записи (старше 1 минуты)
		now := time.Now()
		var recent []time.Time
		for _, t := range visits[ip] {
			if now.Sub(t) < time.Minute {
				recent = append(recent, t)
			}
		}
		visits[ip] = recent

		// Проверяем количество запросов (не более 60 запросов в минуту)
		if len(visits[ip]) >= 60 {
			mu.Unlock()
			c.JSON(429, gin.H{"error": "Too many requests"})
			c.Abort()
			return
		}

		// Добавляем текущее время в список запросов
		visits[ip] = append(visits[ip], now)
		mu.Unlock()

		c.Next()
	}
}

// authMiddleware настраивает проверку аутентификации для API
func authMiddleware(authService service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(401, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		// Проверка токена
		user, err := authService.ValidateToken(c, token)
		if err != nil {
			c.JSON(401, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Сохранение пользователя в контексте
		c.Set("user", user)
		c.Next()
	}
}
