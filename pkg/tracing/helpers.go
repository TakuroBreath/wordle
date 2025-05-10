package otel

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var (
	tracer = otel.Tracer("wordle")
)

// StartSpan создает новый span и возвращает его вместе с новым контекстом
func StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return tracer.Start(ctx, name, opts...)
}

// AddAttributesToSpan добавляет атрибуты к существующему span
func AddAttributesToSpan(span trace.Span, attributes ...attribute.KeyValue) {
	if span != nil {
		span.SetAttributes(attributes...)
	}
}

// SpanFromContext получает текущий span из контекста
func SpanFromContext(ctx context.Context) trace.Span {
	return trace.SpanFromContext(ctx)
}

// EndSpan завершает span
func EndSpan(span trace.Span) {
	if span != nil {
		span.End()
	}
}

// RecordError записывает ошибку в текущий span
func RecordError(ctx context.Context, err error) {
	if err != nil {
		span := SpanFromContext(ctx)
		span.RecordError(err)
	}
}
