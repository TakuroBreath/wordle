package routes

import (
	"os"

	"github.com/TakuroBreath/wordle/internal/api/handlers"
	"github.com/TakuroBreath/wordle/internal/api/middleware"
	"github.com/TakuroBreath/wordle/internal/logger"
	"github.com/TakuroBreath/wordle/internal/models"
	"github.com/TakuroBreath/wordle/pkg/metrics"
	"github.com/gin-gonic/gin"
	otelgin "go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.uber.org/zap"
)

var (
	serviceName = os.Getenv("SERVICE_NAME")
)

// SetupRouter настраивает маршруты API и middleware
func SetupRouter(
	authService models.AuthService,
	userService models.UserService,
	gameService models.GameService,
	lobbyService models.LobbyService,
	transactionService models.TransactionService,
	botToken string,
) *gin.Engine {
	logger.Log.Info("Setting up router")

	router := gin.New()

	// Глобальные middleware
	router.Use(middleware.Recovery())
	router.Use(middleware.Logger())
	router.Use(middleware.CORS())
	router.Use(otelgin.Middleware(serviceName))
	router.Use(middleware.TracingMiddleware())
	router.Use(metrics.MiddlewareMetrics()) // Middleware для Prometheus метрик

	// Инициализация обработчиков
	authHandler := handlers.NewAuthHandler(authService, botToken)
	userHandler := handlers.NewUserHandler(userService, transactionService)
	gameHandler := handlers.NewGameHandler(gameService, userService, transactionService)
	lobbyHandler := handlers.NewLobbyHandler(lobbyService, gameService, userService)
	transactionHandler := handlers.NewTransactionHandler(transactionService, userService)

	logger.Log.Info("Handlers initialized")

	// Middleware для аутентификации
	authMiddleware := middleware.NewAuthMiddleware(authService, botToken)

	// Публичные маршруты
	public := router.Group("/api/v1")
	{
		// Аутентификация
		public.POST("/auth/telegram", authHandler.TelegramAuth)
		public.POST("/auth/verify", authHandler.VerifyAuth)
		public.POST("/auth/logout", authHandler.Logout)

		// Получение списка активных игр (не требует авторизации)
		public.GET("/games", gameHandler.GetActiveGames)
		public.GET("/games/:id", gameHandler.GetGame)
		public.GET("/games/search", gameHandler.SearchGames)
	}

	logger.Log.Info("Public routes configured", zap.String("route_group", "/api/v1"))

	// Защищенные маршруты (требуют аутентификации)
	private := router.Group("/api/v1")
	private.Use(authMiddleware.RequireAuth())
	{
		// Пользователи
		private.GET("/users/me", userHandler.GetCurrentUser)
		private.GET("/users/balance", userHandler.GetUserBalance)
		private.GET("/users/:id", userHandler.GetUserByID)
		private.PUT("/users/me", userHandler.UpdateUser)
		private.POST("/users/withdraw", userHandler.RequestWithdraw)
		private.GET("/users/withdrawals", userHandler.GetWithdrawHistory)
		private.POST("/users/wallet", userHandler.GenerateWalletAddress)

		// Игры
		private.POST("/games", gameHandler.CreateGame)
		private.GET("/games/my", gameHandler.GetUserGames)
		private.GET("/games/created", gameHandler.GetCreatedGames)
		private.DELETE("/games/:id", gameHandler.DeleteGame)
		private.POST("/games/:id/reward", gameHandler.AddToRewardPool)
		private.POST("/games/:id/activate", gameHandler.ActivateGame)
		private.POST("/games/:id/deactivate", gameHandler.DeactivateGame)

		// Лобби
		private.POST("/lobbies", lobbyHandler.JoinGame)
		private.GET("/lobbies/:id", lobbyHandler.GetLobby)
		private.GET("/lobbies/active", lobbyHandler.GetActiveLobby)
		private.GET("/lobbies", lobbyHandler.GetUserLobbies)
		private.POST("/lobbies/:id/attempt", lobbyHandler.MakeAttempt)
		private.GET("/lobbies/:id/attempts", lobbyHandler.GetAttempts)
		private.POST("/lobbies/:id/extend", lobbyHandler.ExtendLobbyTime)

		// Транзакции
		private.GET("/transactions", transactionHandler.GetUserTransactions)
		private.GET("/transactions/:id", transactionHandler.GetTransaction)
		private.GET("/transactions/by-type", transactionHandler.GetTransactionsByType)
		private.GET("/transactions/stats", transactionHandler.GetTransactionStats)
		private.POST("/transactions/deposit", transactionHandler.CreateDepositTransaction)
		private.POST("/transactions/verify", transactionHandler.VerifyDeposit)
	}

	return router
}
