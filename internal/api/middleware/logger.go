package middleware

import (
	"time"

	"github.com/TakuroBreath/wordle/internal/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Logger возвращает middleware для логирования запросов
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Начало времени запроса
		startTime := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		method := c.Request.Method
		ip := c.ClientIP()

		// Создаем логгер для этого запроса
		requestLogger := logger.GetLogger(
			zap.String("path", path),
			zap.String("method", method),
			zap.String("ip", ip),
			zap.String("query", query),
		)

		requestLogger.Info("Request started")

		// Добавляем логгер в контекст
		c.Set("logger", requestLogger)

		// Обработка запроса
		c.Next()

		// Конец времени запроса
		endTime := time.Now()
		latency := endTime.Sub(startTime)

		// Получение статуса ответа
		statusCode := c.Writer.Status()
		errorMessage := c.Errors.ByType(gin.ErrorTypePrivate).String()

		// Логирование результата запроса
		requestLogger.Info(
			"Request completed",
			zap.Int("status", statusCode),
			zap.Duration("latency", latency),
			zap.String("error", errorMessage),
		)
	}
}

// Recovery возвращает middleware для обработки паники
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Получаем логгер из контекста или создаем новый
				var log *zap.Logger
				if l, exists := c.Get("logger"); exists {
					log = l.(*zap.Logger)
				} else {
					log = logger.GetLogger(
						zap.String("path", c.Request.URL.Path),
						zap.String("method", c.Request.Method),
					)
				}

				log.Error("Panic recovered",
					zap.Any("error", err),
					zap.String("request", c.Request.URL.String()),
				)

				c.AbortWithStatus(500)
			}
		}()
		c.Next()
	}
}
