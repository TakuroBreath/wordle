package middleware

import (
	"github.com/TakuroBreath/wordle/internal/logger"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// TracingMiddleware добавляет дополнительные данные о запросе в существующий span
func TracingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Получаем существующий span из контекста
		span := trace.SpanFromContext(c.Request.Context())

		// Добавляем полезные атрибуты
		span.SetAttributes(
			attribute.String("http.method", c.Request.Method),
			attribute.String("http.path", c.Request.URL.Path),
			attribute.String("http.remote_addr", c.ClientIP()),
			attribute.String("http.user_agent", c.Request.UserAgent()),
		)

		// Добавляем заголовки запроса как атрибуты
		for k, v := range c.Request.Header {
			if len(v) > 0 {
				span.SetAttributes(attribute.String("http.header."+k, v[0]))
			}
		}

		// Добавляем query-параметры
		for k, v := range c.Request.URL.Query() {
			if len(v) > 0 {
				span.SetAttributes(attribute.String("http.query."+k, v[0]))
			}
		}

		// Продолжаем обработку запроса
		c.Next()

		// После обработки запроса добавляем код ответа
		span.SetAttributes(attribute.Int("http.status_code", c.Writer.Status()))

		// Логируем информацию о запросе
		logger.Log.Info("Request processed",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
			zap.String("trace_id", span.SpanContext().TraceID().String()),
			zap.String("span_id", span.SpanContext().SpanID().String()),
		)
	}
}
