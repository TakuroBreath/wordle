package main

import (
	"github.com/TakuroBreath/wordle/internal/app"
	"github.com/TakuroBreath/wordle/internal/config"
	"github.com/TakuroBreath/wordle/internal/logger"
	"go.uber.org/zap"
)

func main() {
	// Инициализируем логгер
	isProduction := true // Можно сделать настраиваемым через конфиг
	logger.Init(isProduction)
	defer logger.Sync()

	// Загрузка конфигурации
	cfg, err := config.New()
	if err != nil {
		logger.Log.Fatal("Failed to load config", zap.Error(err))
	}

	logger.Log.Info("Configuration loaded successfully")

	// Создание и запуск приложения
	application, err := app.New(cfg)
	if err != nil {
		logger.Log.Fatal("Failed to initialize application", zap.Error(err))
	}

	logger.Log.Info("Application initialized successfully")

	// Запуск приложения
	if err := application.Run(); err != nil {
		logger.Log.Fatal("Error running application", zap.Error(err))
	}
}
