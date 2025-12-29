package routes

import (
	"github.com/TakuroBreath/wordle/internal/api/handlers"
	"github.com/TakuroBreath/wordle/internal/api/middleware"
	"github.com/TakuroBreath/wordle/internal/logger"
	"github.com/TakuroBreath/wordle/internal/models"
	"github.com/TakuroBreath/wordle/pkg/metrics"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RouterConfig конфигурация роутера
type RouterConfig struct {
	AuthEnabled bool
	BotToken    string
}

// Services содержит все сервисы для роутера
type Services struct {
	AuthService        models.AuthService
	UserService        models.UserService
	GameService        models.GameService
	LobbyService       models.LobbyService
	TransactionService models.TransactionService
	TONService         models.TONService
}

// SetupRouter настраивает маршруты API и middleware
func SetupRouter(
	authService models.AuthService,
	userService models.UserService,
	gameService models.GameService,
	lobbyService models.LobbyService,
	transactionService models.TransactionService,
	botToken string,
) *gin.Engine {
	// Обратная совместимость
	return SetupRouterWithConfig(
		authService,
		userService,
		gameService,
		lobbyService,
		transactionService,
		nil, // No TON service in legacy mode
		RouterConfig{
			AuthEnabled: true,
			BotToken:    botToken,
		},
	)
}

// SetupRouterWithConfig настраивает маршруты API с полной конфигурацией
func SetupRouterWithConfig(
	authService models.AuthService,
	userService models.UserService,
	gameService models.GameService,
	lobbyService models.LobbyService,
	transactionService models.TransactionService,
	tonService models.TONService,
	config RouterConfig,
) *gin.Engine {
	logger.Log.Info("Setting up router",
		zap.Bool("auth_enabled", config.AuthEnabled))

	router := gin.New()

	// Глобальные middleware
	router.Use(middleware.Recovery())
	router.Use(middleware.Logger())
	router.Use(middleware.CORS())
	router.Use(metrics.MiddlewareMetrics())

	// Инициализация обработчиков
	authHandler := handlers.NewAuthHandler(authService, config.BotToken)
	userHandler := handlers.NewUserHandler(userService, transactionService, tonService)
	gameHandler := handlers.NewGameHandler(gameService, userService, transactionService, lobbyService)
	lobbyHandler := handlers.NewLobbyHandler(lobbyService, gameService, userService)
	transactionHandler := handlers.NewTransactionHandler(transactionService, userService)

	logger.Log.Info("Handlers initialized")

	// Middleware для аутентификации с конфигурацией
	authMiddleware := middleware.NewAuthMiddlewareWithConfig(authService, middleware.AuthConfig{
		Enabled:     config.AuthEnabled,
		BotToken:    config.BotToken,
		DefaultUser: middleware.DefaultDevUser(),
	})

	// Публичные маршруты
	public := router.Group("/api/v1")
	{
		// Аутентификация
		public.POST("/auth/telegram", authHandler.TelegramAuth)
		public.POST("/auth/verify", authHandler.VerifyAuth)
		public.POST("/auth/logout", authHandler.Logout)

		// Получение списка активных игр (не требует авторизации)
		public.GET("/games", gameHandler.GetActiveGames)
		public.GET("/games/search", gameHandler.SearchGames)
		public.GET("/games/:id", gameHandler.GetGame)
	}

	logger.Log.Info("Public routes configured", zap.String("route_group", "/api/v1"))

	// Защищенные маршруты (требуют аутентификации)
	private := router.Group("/api/v1")
	private.Use(authMiddleware.RequireAuth())
	{
		// Пользователи
		private.GET("/users/me", userHandler.GetCurrentUser)
		private.GET("/users/balance", userHandler.GetUserBalance)
		private.GET("/users/stats", userHandler.GetUserStats)
		private.GET("/users/top", userHandler.GetTopUsers)
		private.GET("/users/:id", userHandler.GetUserByID)
		private.PUT("/users/me", userHandler.UpdateUser)

		// Кошелёк и депозиты
		private.POST("/users/wallet", userHandler.ConnectWallet)
		private.GET("/users/deposit", userHandler.GetDepositInfo)
		private.POST("/users/withdraw", userHandler.RequestWithdraw)
		private.GET("/users/withdrawals", userHandler.GetWithdrawHistory)
		private.GET("/users/transactions", userHandler.GetTransactionHistory)

		// Игры
		private.POST("/games", gameHandler.CreateGame)
		private.GET("/games/my", gameHandler.GetUserGames)
		private.GET("/games/created", gameHandler.GetCreatedGames)
		private.DELETE("/games/:id", gameHandler.DeleteGame)
		private.GET("/games/:id/payment", gameHandler.GetPaymentInfo)
		private.POST("/games/:id/reward", gameHandler.AddToRewardPool)
		private.POST("/games/:id/activate", gameHandler.ActivateGame)
		private.POST("/games/:id/deactivate", gameHandler.DeactivateGame)
		
		// Вступление в игру (join)
		private.POST("/games/join", gameHandler.JoinGame)

		// Лобби
		private.POST("/lobbies", lobbyHandler.JoinGame)
		private.GET("/lobbies", lobbyHandler.GetUserLobbies)
		private.GET("/lobbies/active", lobbyHandler.GetActiveLobby)
		private.GET("/lobbies/:id", lobbyHandler.GetLobby)
		private.POST("/lobbies/:id/attempt", lobbyHandler.MakeAttempt)
		private.GET("/lobbies/:id/attempts", lobbyHandler.GetAttempts)
		private.POST("/lobbies/:id/extend", lobbyHandler.ExtendLobbyTime)
		private.POST("/lobbies/:id/cancel", lobbyHandler.CancelLobby)

		// Транзакции
		private.GET("/transactions", transactionHandler.GetUserTransactions)
		private.GET("/transactions/:id", transactionHandler.GetTransaction)
		private.GET("/transactions/by-type", transactionHandler.GetTransactionsByType)
		private.GET("/transactions/stats", transactionHandler.GetTransactionStats)
		private.POST("/transactions/deposit", transactionHandler.CreateDepositTransaction)
		private.POST("/transactions/verify", transactionHandler.VerifyDeposit)

		// Блокчейн операции (legacy)
		private.GET("/wallet/address", transactionHandler.GetDepositAddress)
		private.POST("/wallet/withdraw/prepare", transactionHandler.PrepareWithdraw)
	}

	return router
}

// SetupRouterWithServices настраивает роутер используя структуру Services
func SetupRouterWithServices(services Services, config RouterConfig) *gin.Engine {
	return SetupRouterWithConfig(
		services.AuthService,
		services.UserService,
		services.GameService,
		services.LobbyService,
		services.TransactionService,
		services.TONService,
		config,
	)
}
