package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/TakuroBreath/wordle/internal/api/routes"
	"github.com/TakuroBreath/wordle/internal/service"
	"github.com/gin-gonic/gin"
)

// Config представляет конфигурацию сервера
type Config struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
	BotToken     string
	AuthEnabled  bool
}

// Server представляет HTTP-сервер приложения
type Server struct {
	httpServer *http.Server
	router     *gin.Engine
}

// NewServer создает новый экземпляр сервера
func NewServer(cfg Config, services *service.ServiceImpl) *Server {
	// Настройка маршрутов с конфигурацией
	router := routes.SetupRouterWithConfig(
		services.Auth(),
		services.User(),
		services.Game(),
		services.Lobby(),
		services.Transaction(),
		routes.RouterConfig{
			AuthEnabled: cfg.AuthEnabled,
			BotToken:    cfg.BotToken,
		},
	)

	// Создание HTTP-сервера
	httpServer := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	return &Server{
		httpServer: httpServer,
		router:     router,
	}
}

// Run запускает сервер
func (s *Server) Run() error {
	// Создаем канал для получения сигналов остановки
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	// Запускаем сервер в отдельной горутине
	go func() {
		log.Printf("Server is running on port %s", s.httpServer.Addr)
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Ожидаем сигнал остановки
	<-quit
	log.Println("Shutting down server...")

	// Создаем контекст с таймаутом для graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Останавливаем сервер
	if err := s.httpServer.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
		return fmt.Errorf("server forced to shutdown: %w", err)
	}

	log.Println("Server exited properly")
	return nil
}

// Router возвращает роутер сервера
func (s *Server) Router() *gin.Engine {
	return s.router
}
