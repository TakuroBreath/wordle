package service

import (
	"context"
	"fmt"

	otel "github.com/TakuroBreath/wordle/pkg/tracing"
	"go.opentelemetry.io/otel/attribute"
)

// WithTracing добавляет трейсинг к методам сервисов
func WithTracing(ctx context.Context, serviceName, methodName string, f func(context.Context) (interface{}, error)) (interface{}, error) {
	// Создаем новый span
	spanName := fmt.Sprintf("%s.%s", serviceName, methodName)
	ctx, span := otel.StartSpan(ctx, spanName)
	defer span.End()

	// Добавляем атрибуты
	otel.AddAttributesToSpan(span,
		attribute.String("service", serviceName),
		attribute.String("method", methodName),
	)

	// Выполняем функцию
	result, err := f(ctx)

	// В случае ошибки записываем её в span
	if err != nil {
		otel.RecordError(ctx, err)
		span.SetAttributes(attribute.String("error", err.Error()))
	}

	return result, err
}

// WithTracingVoid добавляет трейсинг к методам сервисов, которые не возвращают результат
func WithTracingVoid(ctx context.Context, serviceName, methodName string, f func(context.Context) error) error {
	// Создаем новый span
	spanName := fmt.Sprintf("%s.%s", serviceName, methodName)
	ctx, span := otel.StartSpan(ctx, spanName)
	defer span.End()

	// Добавляем атрибуты
	otel.AddAttributesToSpan(span,
		attribute.String("service", serviceName),
		attribute.String("method", methodName),
	)

	// Выполняем функцию
	err := f(ctx)

	// В случае ошибки записываем её в span
	if err != nil {
		otel.RecordError(ctx, err)
		span.SetAttributes(attribute.String("error", err.Error()))
	}

	return err
}
