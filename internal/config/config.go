package config

import (
	"time"

	"github.com/TakuroBreath/wordle/internal/logger"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

// Config представляет конфигурацию приложения
type Config struct {
	HTTP       HTTPConfig
	Postgres   PostgresConfig
	Redis      RedisConfig
	Auth       AuthConfig
	Metrics    MetricsConfig
	Logging    logger.Config
	Blockchain BlockchainConfig
}

// HTTPConfig представляет конфигурацию HTTP-сервера
type HTTPConfig struct {
	Port         string        `envconfig:"HTTP_PORT" default:"8080"`
	ReadTimeout  time.Duration `envconfig:"HTTP_READ_TIMEOUT" default:"10s"`
	WriteTimeout time.Duration `envconfig:"HTTP_WRITE_TIMEOUT" default:"10s"`
	IdleTimeout  time.Duration `envconfig:"HTTP_IDLE_TIMEOUT" default:"60s"`
}

// PostgresConfig представляет конфигурацию PostgreSQL
type PostgresConfig struct {
	Host     string `envconfig:"POSTGRES_HOST" default:"localhost"`
	Port     string `envconfig:"POSTGRES_PORT" default:"5432"`
	User     string `envconfig:"POSTGRES_USER" default:"postgres"`
	Password string `envconfig:"POSTGRES_PASSWORD" default:"postgres"`
	DBName   string `envconfig:"POSTGRES_DB" default:"wordle"`
	SSLMode  string `envconfig:"POSTGRES_SSLMODE" default:"disable"`
}

// RedisConfig представляет конфигурацию Redis
type RedisConfig struct {
	Host     string `envconfig:"REDIS_HOST" default:"localhost"`
	Port     string `envconfig:"REDIS_PORT" default:"6379"`
	Password string `envconfig:"REDIS_PASSWORD" default:""`
	DB       int    `envconfig:"REDIS_DB" default:"0"`
}

// AuthConfig представляет конфигурацию аутентификации
type AuthConfig struct {
	JWTSecret string `envconfig:"JWT_SECRET" default:"super_secret_key"`
	BotToken  string `envconfig:"BOT_TOKEN" default:"bot_token"`
}

// MetricsConfig представляет конфигурацию для метрик Prometheus
type MetricsConfig struct {
	Enabled bool   `envconfig:"METRICS_ENABLED" default:"true"`
	Port    string `envconfig:"METRICS_PORT" default:"9090"`
}

// BlockchainConfig представляет конфигурацию блокчейна
type BlockchainConfig struct {
	// Provider тип провайдера: "ton", "ethereum", "mock"
	Provider string `envconfig:"BLOCKCHAIN_PROVIDER" default:"mock"`

	// TON configuration
	TON TONConfig

	// Ethereum configuration
	Ethereum EthereumConfig
}

// TONConfig конфигурация для TON блокчейна
type TONConfig struct {
	// APIEndpoint URL API ноды TON
	APIEndpoint string `envconfig:"TON_API_ENDPOINT" default:"https://toncenter.com/api/v2"`
	// APIKey ключ API для доступа к ноде
	APIKey string `envconfig:"TON_API_KEY" default:""`
	// MasterWallet адрес мастер-кошелька для выводов
	MasterWallet string `envconfig:"TON_MASTER_WALLET" default:""`
	// MasterWalletSecret секрет мастер-кошелька
	MasterWalletSecret string `envconfig:"TON_MASTER_WALLET_SECRET" default:""`
	// MinWithdrawTON минимальная сумма вывода в TON
	MinWithdrawTON float64 `envconfig:"TON_MIN_WITHDRAW" default:"0.1"`
	// WithdrawFeeTON комиссия за вывод в TON
	WithdrawFeeTON float64 `envconfig:"TON_WITHDRAW_FEE" default:"0.01"`
	// RequiredConfirmations количество подтверждений
	RequiredConfirmations int `envconfig:"TON_REQUIRED_CONFIRMATIONS" default:"1"`
	// Testnet использовать тестовую сеть
	Testnet bool `envconfig:"TON_TESTNET" default:"false"`
}

// EthereumConfig конфигурация для Ethereum блокчейна
type EthereumConfig struct {
	// RPCURL URL RPC ноды Ethereum
	RPCURL string `envconfig:"ETH_RPC_URL" default:""`
	// ChainID ID сети (1 = mainnet, 5 = goerli, 11155111 = sepolia)
	ChainID int64 `envconfig:"ETH_CHAIN_ID" default:"1"`
	// MasterWallet адрес мастер-кошелька
	MasterWallet string `envconfig:"ETH_MASTER_WALLET" default:""`
	// PrivateKey приватный ключ мастер-кошелька
	PrivateKey string `envconfig:"ETH_PRIVATE_KEY" default:""`
	// MinWithdrawETH минимальная сумма вывода в ETH
	MinWithdrawETH float64 `envconfig:"ETH_MIN_WITHDRAW" default:"0.01"`
	// WithdrawFeeETH комиссия за вывод в ETH
	WithdrawFeeETH float64 `envconfig:"ETH_WITHDRAW_FEE" default:"0.001"`
	// RequiredConfirmations количество подтверждений
	RequiredConfirmations int `envconfig:"ETH_REQUIRED_CONFIRMATIONS" default:"12"`
	// USDTContractAddress адрес контракта USDT (ERC-20)
	USDTContractAddress string `envconfig:"ETH_USDT_CONTRACT" default:"0xdAC17F958D2ee523a2206206994597C13D831ec7"`
}

// New создает новую конфигурацию из переменных окружения
func New() (*Config, error) {
	_ = godotenv.Load()

	var config Config
	err := envconfig.Process("", &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

// DSN возвращает строку подключения к PostgreSQL
func (p PostgresConfig) DSN() string {
	return "host=" + p.Host +
		" port=" + p.Port +
		" user=" + p.User +
		" password=" + p.Password +
		" dbname=" + p.DBName +
		" sslmode=" + p.SSLMode
}

// Addr возвращает адрес сервера Redis
func (r RedisConfig) Addr() string {
	return r.Host + ":" + r.Port
}
