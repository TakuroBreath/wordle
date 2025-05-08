package main

import (
	"log"

	"github.com/TakuroBreath/wordle/internal/app"
	"github.com/TakuroBreath/wordle/internal/config"
)

func main() {
	// Загрузка конфигурации
	cfg, err := config.New()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Создание и запуск приложения
	application, err := app.New(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}

	// Запуск приложения
	if err := application.Run(); err != nil {
		log.Fatalf("Error running application: %v", err)
	}
}
