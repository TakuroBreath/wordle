package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Global logger instance
var Log *zap.Logger

// SugaredLogger is a sugared version of the global logger
var SugaredLog *zap.SugaredLogger

// Init initializes the global logger
func Init(isProduction bool) {
	var config zap.Config

	if isProduction {
		config = zap.NewProductionConfig()
	} else {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	config.OutputPaths = []string{"stdout"}
	config.ErrorOutputPaths = []string{"stderr"}

	var err error
	Log, err = config.Build()
	if err != nil {
		// Fallback to a basic stdout log if zap initialization fails
		consoleEncoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
		core := zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), zapcore.DebugLevel)
		Log = zap.New(core)
		Log.Error("Failed to initialize zap logger", zap.Error(err))
	}

	SugaredLog = Log.Sugar()

	Log.Info("Logger initialized", zap.Bool("production", isProduction))
}

// GetLogger returns a logger with added fields
func GetLogger(fields ...zap.Field) *zap.Logger {
	if Log == nil {
		Init(false) // Initialize with development settings if not already initialized
	}
	return Log.With(fields...)
}

// GetSugaredLogger returns a sugared logger with added fields
func GetSugaredLogger(keysAndValues ...interface{}) *zap.SugaredLogger {
	if SugaredLog == nil {
		Init(false) // Initialize with development settings if not already initialized
	}
	return SugaredLog.With(keysAndValues...)
}

// Sync flushes any buffered log entries
func Sync() {
	if Log != nil {
		_ = Log.Sync()
	}
}
