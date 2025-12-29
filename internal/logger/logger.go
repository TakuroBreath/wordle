package logger

import (
	"context"
	"io"
	"os"
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
	IsProduction bool   `yaml:"isProduction" env:"PRODUCTION" env-default:"false"`
	// Путь к JSON логам для Loki (если пусто - JSON не пишется)
	LokiFilePath string `yaml:"lokiFilePath" env:"LOG_LOKI_FILE_PATH" env-default:"logs/app.json"`
	// Настройки ротации для JSON файла
	MaxSize    int  `yaml:"maxSize" env:"LOG_MAX_SIZE" env-default:"100"` // в МБ
	MaxBackups int  `yaml:"maxBackups" env:"LOG_MAX_BACKUPS" env-default:"30"`
	MaxAge     int  `yaml:"maxAge" env:"LOG_MAX_AGE" env-default:"7"` // в днях
	Compress   bool `yaml:"compress" env:"LOG_COMPRESS" env-default:"true"`
	// Отключить консольный вывод (для тестов)
	DisableConsole bool `yaml:"disableConsole" env:"LOG_DISABLE_CONSOLE" env-default:"false"`
	// Отключить JSON вывод в файл
	DisableLokiFile bool `yaml:"disableLokiFile" env:"LOG_DISABLE_LOKI_FILE" env-default:"false"`
}

// Init initializes the global logger with dual output:
// - Console (human-readable) → stdout
// - JSON (structured) → file for Loki/Promtail
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

	var cores []zapcore.Core

	// Console encoder для человекочитаемого вывода в stdout
	if !cfg.DisableConsole {
		consoleEncoderConfig := zapcore.EncoderConfig{
			TimeKey:        "T",
			LevelKey:       "L",
			NameKey:        "N",
			CallerKey:      "C",
			FunctionKey:    zapcore.OmitKey,
			MessageKey:     "M",
			StacktraceKey:  "S",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.CapitalColorLevelEncoder, // Цветной вывод уровней
			EncodeTime:     zapcore.TimeEncoderOfLayout("15:04:05.000"),
			EncodeDuration: zapcore.StringDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		}
		consoleEncoder := zapcore.NewConsoleEncoder(consoleEncoderConfig)
		consoleCore := zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), level)
		cores = append(cores, consoleCore)
	}

	// JSON encoder для Loki (через файл, который читает Promtail)
	if !cfg.DisableLokiFile && cfg.LokiFilePath != "" {
		jsonEncoderConfig := zapcore.EncoderConfig{
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

		// Настройка ротации логов для JSON файла
		rotator := &lumberjack.Logger{
			Filename:   cfg.LokiFilePath,
			MaxSize:    cfg.MaxSize,    // мегабайты
			MaxBackups: cfg.MaxBackups, // количество бэкапов
			MaxAge:     cfg.MaxAge,     // дни
			Compress:   cfg.Compress,   // сжатие
		}

		jsonEncoder := zapcore.NewJSONEncoder(jsonEncoderConfig)
		jsonCore := zapcore.NewCore(jsonEncoder, zapcore.AddSync(rotator), level)
		cores = append(cores, jsonCore)
	}

	// Если нет ни одного core, создаем nop core
	if len(cores) == 0 {
		cores = append(cores, zapcore.NewNopCore())
	}

	// Объединяем все ядра
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
	SugaredLog = Log.Sugar()

	Log.Info("Logger initialized",
		zap.String("level", cfg.Level),
		zap.Bool("production", cfg.IsProduction),
		zap.Bool("console_enabled", !cfg.DisableConsole),
		zap.Bool("loki_file_enabled", !cfg.DisableLokiFile),
		zap.String("loki_file_path", cfg.LokiFilePath))
}

// InitDefault инициализирует логгер с настройками по умолчанию
func InitDefault() {
	config := Config{
		Level:          "info",
		IsProduction:   false,
		LokiFilePath:   "logs/app.json",
		DisableConsole: false,
	}
	Init(config)
}

// InitForTesting инициализирует логгер для тестов (выводит в io.Writer)
func InitForTesting(w io.Writer) {
	consoleEncoderConfig := zapcore.EncoderConfig{
		TimeKey:        "T",
		LevelKey:       "L",
		NameKey:        "N",
		CallerKey:      "C",
		MessageKey:     "M",
		StacktraceKey:  "S",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.TimeEncoderOfLayout("15:04:05.000"),
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
	consoleEncoder := zapcore.NewConsoleEncoder(consoleEncoderConfig)
	core := zapcore.NewCore(consoleEncoder, zapcore.AddSync(w), zap.DebugLevel)

	Log = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	SugaredLog = Log.Sugar()
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

// WithContext обогащает лог контекстом
func WithContext(ctx context.Context) *zap.Logger {
	if Log == nil {
		InitDefault()
	}

	fields := []zap.Field{}

	// Добавляем request_id если есть в контексте
	if reqID := ctx.Value("request_id"); reqID != nil {
		if id, ok := reqID.(string); ok && id != "" {
			fields = append(fields, zap.String("request_id", id))
		}
	}

	// Добавляем user_id если есть в контексте
	if userID := ctx.Value("user_id"); userID != nil {
		if id, ok := userID.(string); ok && id != "" {
			fields = append(fields, zap.String("user_id", id))
		}
	}

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
