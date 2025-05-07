package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/TakuroBreath/wordle/internal/api/handlers"
	"github.com/TakuroBreath/wordle/internal/api/routes"
	"github.com/TakuroBreath/wordle/internal/repository/postgresql"
	"github.com/TakuroBreath/wordle/internal/repository/redis"
	"github.com/TakuroBreath/wordle/internal/service"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	r "github.com/redis/go-redis/v9"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Загрузка конфигурации
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "postgres")
	dbName := getEnv("DB_NAME", "wordle")
	redisAddr := getEnv("REDIS_ADDR", "localhost:6379")
	redisPassword := getEnv("REDIS_PASSWORD", "")
	redisDB := 0
	serverPort := getEnv("SERVER_PORT", "8080")

	// Подключение к PostgreSQL
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Проверка подключения к базе данных с контекстом
	dbCtx, dbCancel := context.WithTimeout(ctx, 5*time.Second)
	defer dbCancel()

	if err := db.PingContext(dbCtx); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// Подключение к Redis
	rdb := r.NewClient(&r.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       redisDB,
	})

	// Проверка подключения к Redis с контекстом
	redisCtx, redisCancel := context.WithTimeout(ctx, 5*time.Second)
	defer redisCancel()

	if err := rdb.Ping(redisCtx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer rdb.Close()

	// Инициализация репозиториев
	postgresRepo := postgresql.NewRepository(db)
	redisRepo := redis.NewRepository(rdb)

	// Инициализация сервисов
	services := service.NewService(postgresRepo, redisRepo)

	// Инициализация обработчиков
	handler := handlers.NewHandler(services)

	// Инициализация маршрутизатора
	router := routes.SetupRouter(handler, services.Auth())

	// Запуск сервера
	serverAddr := fmt.Sprintf(":%s", serverPort)
	server := startServer(router, serverAddr)

	// Ожидание сигнала для graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Graceful shutdown
	log.Println("Shutting down server...")
	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}

// getEnv возвращает значение переменной окружения или значение по умолчанию
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// startServer запускает HTTP сервер
func startServer(router *gin.Engine, addr string) *http.Server {
	server := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	go func() {
		log.Printf("Server is running on %s", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	return server
}
