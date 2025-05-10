package repository

import (
	"context"
	"fmt"

	otel "github.com/TakuroBreath/wordle/pkg/tracing"
	"go.opentelemetry.io/otel/attribute"
)

// WithTracing добавляет трейсинг к методам репозитория
func WithTracing(ctx context.Context, repoName, methodName string, f func(context.Context) (interface{}, error)) (interface{}, error) {
	// Создаем новый span
	spanName := fmt.Sprintf("repository.%s.%s", repoName, methodName)
	ctx, span := otel.StartSpan(ctx, spanName)
	defer span.End()

	// Добавляем атрибуты
	otel.AddAttributesToSpan(span,
		attribute.String("repository", repoName),
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

// WithTracingVoid добавляет трейсинг к методам репозитория, которые не возвращают результат
func WithTracingVoid(ctx context.Context, repoName, methodName string, f func(context.Context) error) error {
	// Создаем новый span
	spanName := fmt.Sprintf("repository.%s.%s", repoName, methodName)
	ctx, span := otel.StartSpan(ctx, spanName)
	defer span.End()

	// Добавляем атрибуты
	otel.AddAttributesToSpan(span,
		attribute.String("repository", repoName),
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
