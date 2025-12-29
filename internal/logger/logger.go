package logger

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

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
	// FilePath: путь к JSON-логу, который будет считываться Promtail и отправляться в Loki.
	// Человекочитаемые логи всегда пишутся в stdout.
	FilePath     string `yaml:"filePath" env:"LOG_FILE_PATH" env-default:"logs/app.json"`
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

	// 1) Человекочитаемые логи в stdout
	consoleEncoderCfg := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.MillisDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
	consoleCore := zapcore.NewCore(
		zapcore.NewConsoleEncoder(consoleEncoderCfg),
		zapcore.AddSync(os.Stdout),
		level,
	)

	// 2) Структурированные JSON логи в файл (для Loki через Promtail)
	jsonEncoderCfg := zapcore.EncoderConfig{
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

	var cores []zapcore.Core
	cores = append(cores, consoleCore)

	jsonEnabled := cfg.FilePath != ""
	if jsonEnabled {
		// lumberjack не создаёт директории сам
		if err := os.MkdirAll(filepath.Dir(cfg.FilePath), 0o755); err != nil {
			fmt.Fprintf(os.Stderr, "failed to create log directory: %v\n", err)
			jsonEnabled = false
		}
	}

	if jsonEnabled {
		rotator := &lumberjack.Logger{
			Filename:   cfg.FilePath,
			MaxSize:    cfg.MaxSize,    // мегабайты
			MaxBackups: cfg.MaxBackups, // количество бэкапов
			MaxAge:     cfg.MaxAge,     // дни
			Compress:   cfg.Compress,   // сжатие
		}
		jsonCore := zapcore.NewCore(
			zapcore.NewJSONEncoder(jsonEncoderCfg),
			zapcore.AddSync(rotator),
			level,
		)
		cores = append(cores, jsonCore)
	}

	core := zapcore.NewTee(cores...)

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

	// Базовые поля (попадают и в console, и в JSON)
	if serviceName := os.Getenv("SERVICE_NAME"); serviceName != "" {
		Log = Log.With(zap.String("service", serviceName))
	}
	SugaredLog = Log.Sugar()

	Log.Info("Logger initialized",
		zap.String("level", cfg.Level),
		zap.String("format", cfg.Format),
		zap.String("output", cfg.Output),
		zap.Bool("production", cfg.IsProduction),
		zap.Bool("json_file_enabled", jsonEnabled),
		zap.String("json_file_path", cfg.FilePath))
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
func GetSugaredLogger(keysAndValues ...any) *zap.SugaredLogger {
	if SugaredLog == nil {
		InitDefault()
	}
	return SugaredLog.With(keysAndValues...)
}

// WithContext обогащает лог контекстом (трейсинг удалён).
func WithContext(ctx context.Context) *zap.Logger {
	if Log == nil {
		InitDefault()
	}
	_ = ctx
	return Log
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
