package main

import (
	"github.com/TakuroBreath/wordle/internal/app"
	"github.com/TakuroBreath/wordle/internal/config"
	"github.com/TakuroBreath/wordle/internal/logger"
	"go.uber.org/zap"
)

func main() {
	// Загрузка конфигурации до инициализации логгера
	cfg, err := config.New()
	if err != nil {
		// Пока используем стандартный логгер, так как наш еще не инициализирован
		panic("Failed to load config: " + err.Error())
	}

	// Инициализируем логгер с загруженной конфигурацией
	logger.Init(cfg.Logging)
	defer logger.Sync()

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
