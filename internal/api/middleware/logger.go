package middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

// Logger возвращает middleware для логирования запросов
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Время начала запроса
		startTime := time.Now()

		// Обработка запроса
		c.Next()

		// Получение статуса ответа и времени обработки
		statusCode := c.Writer.Status()
		duration := time.Since(startTime)

		// Логирование запроса
		fmt.Printf("[%s] %s %s %d %s\n",
			startTime.Format("2006-01-02 15:04:05"),
			c.Request.Method,
			c.Request.URL.Path,
			statusCode,
			duration,
		)
	}
}

// Recovery возвращает middleware для обработки паники
func Recovery() gin.HandlerFunc {
	return gin.Recovery()
}
