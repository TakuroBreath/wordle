package logger

import (
	"context"
	"os"
	"time"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Global logger instance
var Log *zap.Logger

// SugaredLogger is a sugared version of the global logger
var SugaredLog *zap.SugaredLogger

// Config содержит конфигурацию логгера
type Config struct {
	Level        string `yaml:"level" env:"LOG_LEVEL" env-default:"info"`
	Format       string `yaml:"format" env:"LOG_FORMAT" env-default:"json"`
	Output       string `yaml:"output" env:"LOG_OUTPUT" env-default:"stdout"`
	IsProduction bool   `yaml:"isProduction" env:"PRODUCTION" env-default:"false"`
	FilePath     string `yaml:"filePath" env:"LOG_FILE_PATH" env-default:"logs/app.log"`
	MaxSize      int    `yaml:"maxSize" env:"LOG_MAX_SIZE" env-default:"100"` // в МБ
	MaxBackups   int    `yaml:"maxBackups" env:"LOG_MAX_BACKUPS" env-default:"30"`
	MaxAge       int    `yaml:"maxAge" env:"LOG_MAX_AGE" env-default:"30"` // в днях
	Compress     bool   `yaml:"compress" env:"LOG_COMPRESS" env-default:"true"`
}

// Init initializes the global logger
func Init(cfg Config) {
	// Определяем уровень логирования
	level := zap.InfoLevel
	switch cfg.Level {
	case "debug":
		level = zap.DebugLevel
	case "info":
		level = zap.InfoLevel
	case "warn":
		level = zap.WarnLevel
	case "error":
		level = zap.ErrorLevel
	case "fatal":
		level = zap.FatalLevel
	}

	// Настройка вывода логов
	var sink zapcore.WriteSyncer
	if cfg.Output == "file" {
		// Настройка ротации логов
		rotator := &lumberjack.Logger{
			Filename:   cfg.FilePath,
			MaxSize:    cfg.MaxSize,    // мегабайты
			MaxBackups: cfg.MaxBackups, // количество бэкапов
			MaxAge:     cfg.MaxAge,     // дни
			Compress:   cfg.Compress,   // сжатие
		}
		sink = zapcore.AddSync(rotator)
	} else {
		sink = zapcore.AddSync(os.Stdout)
	}

	// Конфигурация энкодера
	var encoder zapcore.Encoder
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.MillisDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	if cfg.Format == "json" {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		// Для консоли используем цветное форматирование
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	// Создаем ядро логгера
	core := zapcore.NewCore(encoder, sink, level)

	// Дополнительные опции
	opts := []zap.Option{
		zap.AddCaller(),
		zap.AddCallerSkip(1),
	}

	if !cfg.IsProduction {
		opts = append(opts, zap.Development())
	}

	// Создаем логгер
	Log = zap.New(core, opts...)
	SugaredLog = Log.Sugar()

	Log.Info("Logger initialized",
		zap.String("level", cfg.Level),
		zap.String("format", cfg.Format),
		zap.String("output", cfg.Output),
		zap.Bool("production", cfg.IsProduction))
}

// InitDefault инициализирует логгер с настройками по умолчанию
func InitDefault() {
	config := Config{
		Level:        "info",
		Format:       "json",
		Output:       "stdout",
		IsProduction: false,
	}
	Init(config)
}

// GetLogger returns a logger with added fields
func GetLogger(fields ...zap.Field) *zap.Logger {
	if Log == nil {
		InitDefault()
	}
	return Log.With(fields...)
}

// GetSugaredLogger returns a sugared logger with added fields
func GetSugaredLogger(keysAndValues ...interface{}) *zap.SugaredLogger {
	if SugaredLog == nil {
		InitDefault()
	}
	return SugaredLog.With(keysAndValues...)
}

// WithContext обогащает лог контекстом и информацией о трейсинге
func WithContext(ctx context.Context) *zap.Logger {
	if Log == nil {
		InitDefault()
	}

	fields := []zap.Field{}

	// Добавляем trace_id и span_id, если они есть в контексте
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		traceID := span.SpanContext().TraceID().String()
		spanID := span.SpanContext().SpanID().String()
		fields = append(fields,
			zap.String("trace_id", traceID),
			zap.String("span_id", spanID))
	}

	// Можно добавить другие поля из контекста
	// Например, user_id, request_id и т.д., если они есть в контексте

	return Log.With(fields...)
}

// LogError логирует ошибку с контекстом
func LogError(ctx context.Context, msg string, err error, fields ...zap.Field) {
	logger := WithContext(ctx)
	if err != nil {
		fields = append(fields, zap.Error(err))
	}
	logger.Error(msg, fields...)
}

// LogInfo логирует информацию с контекстом
func LogInfo(ctx context.Context, msg string, fields ...zap.Field) {
	WithContext(ctx).Info(msg, fields...)
}

// LogDebug логирует отладочную информацию с контекстом
func LogDebug(ctx context.Context, msg string, fields ...zap.Field) {
	WithContext(ctx).Debug(msg, fields...)
}

// LogWarn логирует предупреждение с контекстом
func LogWarn(ctx context.Context, msg string, fields ...zap.Field) {
	WithContext(ctx).Warn(msg, fields...)
}

// Sync flushes any buffered log entries
func Sync() {
	if Log != nil {
		_ = Log.Sync()
	}
}

// WithFields добавляет поля к логгеру
func WithFields(fields ...zap.Field) *zap.Logger {
	if Log == nil {
		InitDefault()
	}
	return Log.With(fields...)
}

// CustomTimeEncoder форматирует время определенным образом
func CustomTimeEncoder() zapcore.TimeEncoder {
	return func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format(time.RFC3339Nano))
	}
}
