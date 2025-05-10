package metrics

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	registry *MetricsRegistry
)

// GetRegistry возвращает текущий экземпляр реестра метрик
func GetRegistry() *MetricsRegistry {
	return registry
}

// SetRegistry устанавливает глобальный экземпляр реестра метрик
func SetRegistry(r *MetricsRegistry) {
	registry = r
}

// MiddlewareMetrics возвращает middleware для сбора метрик HTTP запросов
func MiddlewareMetrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		if registry == nil {
			c.Next()
			return
		}

		start := time.Now()
		path := c.FullPath()
		if path == "" {
			path = "unknown"
		}
		method := c.Request.Method

		c.Next()

		status := strconv.Itoa(c.Writer.Status())
		duration := time.Since(start).Seconds()

		// Учитываем метрики
		registry.RequestTotal.WithLabelValues(method, path, status).Inc()
		registry.RequestDuration.WithLabelValues(method, path, status).Observe(duration)
	}
}

// IncrementGameStart увеличивает счетчик начатых игр
func IncrementGameStart() {
	if registry != nil {
		registry.GameStartTotal.Inc()
		registry.ActiveGamesGauge.Inc()
	}
}

// IncrementGameComplete увеличивает счетчик завершенных игр с определенным числом попыток
func IncrementGameComplete(attempts int) {
	if registry != nil {
		registry.GameCompleteTotal.Inc()
		registry.ActiveGamesGauge.Dec()
		registry.WordGuessedCounters.WithLabelValues(strconv.Itoa(attempts)).Inc()
	}
}

// IncrementGameAbandoned увеличивает счетчик брошенных игр
func IncrementGameAbandoned() {
	if registry != nil {
		registry.GameAbandonedTotal.Inc()
		registry.ActiveGamesGauge.Dec()
	}
}

// RecordError записывает ошибку в метрики
func RecordError(errorType string) {
	if registry != nil {
		registry.ErrorsTotal.WithLabelValues(errorType).Inc()
	}
}
