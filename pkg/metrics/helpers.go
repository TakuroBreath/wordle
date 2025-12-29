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

// RecordDeposit записывает успешный депозит
func RecordDeposit(currency string, amount float64) {
	if registry != nil {
		registry.DepositsTotal.WithLabelValues(currency).Add(amount)
		registry.DepositsCount.WithLabelValues(currency).Inc()
	}
}

// RecordWithdrawal записывает успешный вывод
func RecordWithdrawal(currency string, amount float64) {
	if registry != nil {
		registry.WithdrawalsTotal.WithLabelValues(currency).Add(amount)
		registry.WithdrawalsCount.WithLabelValues(currency).Inc()
	}
}

// RecordCommission записывает комиссию (revenue)
func RecordCommission(currency string, amount float64) {
	if registry != nil {
		registry.CommissionsTotal.WithLabelValues(currency).Add(amount)
	}
}

// RecordBet записывает ставку
func RecordBet(currency string, amount float64) {
	if registry != nil {
		registry.BetsTotal.WithLabelValues(currency).Add(amount)
	}
}

// RecordReward записывает выплату игроку
func RecordReward(currency string, amount float64) {
	if registry != nil {
		registry.RewardsTotal.WithLabelValues(currency).Add(amount)
	}
}

// IncrementPendingWithdrawals увеличивает счётчик ожидающих выводов
func IncrementPendingWithdrawals() {
	if registry != nil {
		registry.PendingWithdrawals.Inc()
	}
}

// DecrementPendingWithdrawals уменьшает счётчик ожидающих выводов
func DecrementPendingWithdrawals() {
	if registry != nil {
		registry.PendingWithdrawals.Dec()
	}
}

// SetTotalUsersBalance устанавливает общий баланс пользователей
func SetTotalUsersBalance(currency string, balance float64) {
	if registry != nil {
		registry.TotalUsersBalance.WithLabelValues(currency).Set(balance)
	}
}
