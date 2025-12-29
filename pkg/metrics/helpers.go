package metrics

import (
	"math"
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

// AddDeposit добавляет сумму пополнения в денежные метрики.
// amount должен быть в основных единицах валюты (TON, USDT).
func AddDeposit(amount float64, currency string, source string) {
	if registry == nil {
		return
	}
	if amount <= 0 || math.IsNaN(amount) || math.IsInf(amount, 0) {
		return
	}
	if currency == "" {
		currency = "unknown"
	}
	if source == "" {
		source = "unknown"
	}
	registry.MoneyDepositedTotal.WithLabelValues(currency, source).Add(amount)
}

// AddWithdraw добавляет сумму вывода (нетто) в денежные метрики.
func AddWithdraw(amount float64, currency string) {
	if registry == nil {
		return
	}
	if amount <= 0 || math.IsNaN(amount) || math.IsInf(amount, 0) {
		return
	}
	if currency == "" {
		currency = "unknown"
	}
	registry.MoneyWithdrawnTotal.WithLabelValues(currency).Add(amount)
}

// AddRevenue добавляет сумму ревеню (комиссии/fees) в денежные метрики.
func AddRevenue(amount float64, currency string, revenueType string) {
	if registry == nil {
		return
	}
	if amount <= 0 || math.IsNaN(amount) || math.IsInf(amount, 0) {
		return
	}
	if currency == "" {
		currency = "unknown"
	}
	if revenueType == "" {
		revenueType = "unknown"
	}
	registry.RevenueTotal.WithLabelValues(currency, revenueType).Add(amount)
}
