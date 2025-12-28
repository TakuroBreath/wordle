package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/TakuroBreath/wordle/internal/logger"
	"gopkg.in/yaml.v3"
)

// Environment тип окружения
type Environment string

const (
	EnvDev  Environment = "dev"
	EnvProd Environment = "prod"
)

// Network тип сети
type Network string

const (
	NetworkTON Network = "ton"
	NetworkEVM Network = "evm"
)

// Config представляет конфигурацию приложения
type Config struct {
	// Основные настройки
	Environment     Environment `yaml:"environment"`
	Network         Network     `yaml:"network"`
	UseMockProvider bool        `yaml:"use_mock_provider"`

	// Компоненты
	HTTP       HTTPConfig       `yaml:"http"`
	Postgres   PostgresConfig   `yaml:"postgres"`
	Redis      RedisConfig      `yaml:"redis"`
	Auth       AuthConfig       `yaml:"auth"`
	Metrics    MetricsConfig    `yaml:"metrics"`
	Logging    logger.Config    `yaml:"logging"`
	Blockchain BlockchainConfig `yaml:"blockchain"`
}

// HTTPConfig представляет конфигурацию HTTP-сервера
type HTTPConfig struct {
	Port         string        `yaml:"port"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
	IdleTimeout  time.Duration `yaml:"idle_timeout"`
}

// PostgresConfig представляет конфигурацию PostgreSQL
type PostgresConfig struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"db_name"`
	SSLMode  string `yaml:"ssl_mode"`
}

// RedisConfig представляет конфигурацию Redis
type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

// AuthConfig представляет конфигурацию аутентификации
type AuthConfig struct {
	Enabled   bool          `yaml:"enabled"`
	JWTSecret string        `yaml:"jwt_secret"`
	BotToken  string        `yaml:"bot_token"`
	TokenTTL  time.Duration `yaml:"token_ttl"`
}

// MetricsConfig представляет конфигурацию для метрик Prometheus
type MetricsConfig struct {
	Enabled bool   `yaml:"enabled"`
	Port    string `yaml:"port"`
}

// BlockchainConfig представляет конфигурацию блокчейна
type BlockchainConfig struct {
	TON      TONConfig      `yaml:"ton"`
	Ethereum EthereumConfig `yaml:"ethereum"`
}

// TONConfig конфигурация для TON блокчейна
type TONConfig struct {
	APIEndpoint           string  `yaml:"api_endpoint"`
	APIKey                string  `yaml:"api_key"`
	MasterWallet          string  `yaml:"master_wallet"`
	MasterWalletSeed      string  `yaml:"master_wallet_seed"`      // 24 слова seed фразы
	MasterWalletSecret    string  `yaml:"master_wallet_secret"`    // Deprecated: используй master_wallet_seed
	MinWithdrawTON        float64 `yaml:"min_withdraw_ton"`
	MinWithdrawUSDT       float64 `yaml:"min_withdraw_usdt"`
	WithdrawFeeTON        float64 `yaml:"withdraw_fee_ton"`
	WithdrawFeeUSDT       float64 `yaml:"withdraw_fee_usdt"`
	RequiredConfirmations int     `yaml:"required_confirmations"`
	Testnet               bool    `yaml:"testnet"`
	USDTMasterAddress     string  `yaml:"usdt_master_address"`     // Адрес USDT Jetton контракта
	WorkerPollInterval    int     `yaml:"worker_poll_interval"`    // Интервал опроса воркера в секундах
	CommissionRate        float64 `yaml:"commission_rate"`         // Комиссия сервиса (0.05 = 5%)
}

// EthereumConfig конфигурация для Ethereum блокчейна
type EthereumConfig struct {
	RPCURL                string  `yaml:"rpc_url"`
	ChainID               int64   `yaml:"chain_id"`
	MasterWallet          string  `yaml:"master_wallet"`
	PrivateKey            string  `yaml:"private_key"`
	MinWithdrawETH        float64 `yaml:"min_withdraw"`
	WithdrawFeeETH        float64 `yaml:"withdraw_fee"`
	RequiredConfirmations int     `yaml:"required_confirmations"`
	USDTContractAddress   string  `yaml:"usdt_contract"`
}

// Load загружает конфигурацию из YAML файла
func Load(configPath string) (*Config, error) {
	// Читаем файл
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Подставляем переменные окружения
	configStr := expandEnvVariables(string(data))

	// Парсим YAML
	var config Config
	if err := yaml.Unmarshal([]byte(configStr), &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Применяем правила окружения
	config.applyEnvironmentRules()

	// Валидируем конфигурацию
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &config, nil
}

// New создает конфигурацию (обратная совместимость)
// Ищет конфиг файл по приоритету:
// 1. CONFIG_PATH env variable
// 2. ./config.yaml
// 3. ./configs/config.local.yaml
// 4. ./configs/config.dev.yaml
func New() (*Config, error) {
	configPath := os.Getenv("CONFIG_PATH")

	if configPath == "" {
		// Пробуем найти конфиг файл
		candidates := []string{
			"config.yaml",
			"configs/config.local.yaml",
			"configs/config.dev.yaml",
		}

		for _, candidate := range candidates {
			if _, err := os.Stat(candidate); err == nil {
				configPath = candidate
				break
			}
		}
	}

	// Если файл не найден, используем дефолтную конфигурацию
	if configPath == "" {
		return defaultConfig(), nil
	}

	return Load(configPath)
}

// defaultConfig возвращает конфигурацию по умолчанию (dev режим)
func defaultConfig() *Config {
	return &Config{
		Environment:     EnvDev,
		Network:         NetworkTON,
		UseMockProvider: true,
		HTTP: HTTPConfig{
			Port:         getEnvOrDefault("HTTP_PORT", "8080"),
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
		Postgres: PostgresConfig{
			Host:     getEnvOrDefault("POSTGRES_HOST", "localhost"),
			Port:     getEnvOrDefault("POSTGRES_PORT", "5432"),
			User:     getEnvOrDefault("POSTGRES_USER", "postgres"),
			Password: getEnvOrDefault("POSTGRES_PASSWORD", "postgres"),
			DBName:   getEnvOrDefault("POSTGRES_DB", "wordle"),
			SSLMode:  getEnvOrDefault("POSTGRES_SSLMODE", "disable"),
		},
		Redis: RedisConfig{
			Host:     getEnvOrDefault("REDIS_HOST", "localhost"),
			Port:     getEnvOrDefault("REDIS_PORT", "6379"),
			Password: getEnvOrDefault("REDIS_PASSWORD", ""),
			DB:       0,
		},
		Auth: AuthConfig{
			Enabled:   false,
			JWTSecret: getEnvOrDefault("JWT_SECRET", "dev_secret_key"),
			BotToken:  getEnvOrDefault("BOT_TOKEN", ""),
			TokenTTL:  24 * time.Hour,
		},
		Metrics: MetricsConfig{
			Enabled: true,
			Port:    getEnvOrDefault("METRICS_PORT", "9090"),
		},
		Logging: logger.Config{
			Level: "debug",
		},
		Blockchain: BlockchainConfig{
			TON: TONConfig{
				APIEndpoint:           "https://testnet.toncenter.com/api/v2",
				Testnet:               true,
				MinWithdrawTON:        0.1,
				WithdrawFeeTON:        0.01,
				RequiredConfirmations: 1,
			},
			Ethereum: EthereumConfig{
				ChainID:               11155111, // Sepolia
				MinWithdrawETH:        0.01,
				WithdrawFeeETH:        0.001,
				RequiredConfirmations: 3,
			},
		},
	}
}

// applyEnvironmentRules применяет правила на основе окружения и сети
func (c *Config) applyEnvironmentRules() {
	// В dev режиме авторизация всегда выключена
	if c.Environment == EnvDev {
		c.Auth.Enabled = false
	}

	// В prod режиме авторизация всегда включена
	if c.Environment == EnvProd {
		c.Auth.Enabled = true
		// В prod не используем mock провайдер
		c.UseMockProvider = false
	}

	// Для EVM пока авторизация отключена (в будущем - через кошелёк)
	if c.Network == NetworkEVM {
		c.Auth.Enabled = false
	}
}

// Validate валидирует конфигурацию
func (c *Config) Validate() error {
	// Базовая валидация - ничего критичного не проверяем
	// Для пет-проекта это нормально
	return nil
}

// IsDev проверяет, является ли окружение dev
func (c *Config) IsDev() bool {
	return c.Environment == EnvDev
}

// IsProd проверяет, является ли окружение prod
func (c *Config) IsProd() bool {
	return c.Environment == EnvProd
}

// IsTON проверяет, используется ли TON сеть
func (c *Config) IsTON() bool {
	return c.Network == NetworkTON
}

// IsEVM проверяет, используется ли EVM сеть
func (c *Config) IsEVM() bool {
	return c.Network == NetworkEVM
}

// IsAuthEnabled проверяет, включена ли авторизация
func (c *Config) IsAuthEnabled() bool {
	return c.Auth.Enabled
}

// GetBlockchainProviderType возвращает тип провайдера блокчейна
func (c *Config) GetBlockchainProviderType() string {
	if c.UseMockProvider {
		return "mock"
	}

	switch c.Network {
	case NetworkTON:
		return "ton"
	case NetworkEVM:
		return "ethereum"
	default:
		return "mock"
	}
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

// expandEnvVariables заменяет ${VAR} на значения переменных окружения
func expandEnvVariables(content string) string {
	return os.Expand(content, func(key string) string {
		if value, exists := os.LookupEnv(key); exists {
			return value
		}
		return "${" + key + "}"
	})
}

// getEnvOrDefault возвращает значение переменной окружения или дефолтное значение
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// String возвращает строковое представление конфигурации (для логов)
func (c *Config) String() string {
	return fmt.Sprintf(
		"Config{Environment: %s, Network: %s, UseMockProvider: %v, AuthEnabled: %v}",
		c.Environment, c.Network, c.UseMockProvider, c.Auth.Enabled,
	)
}

// GetSupportedCurrencies возвращает поддерживаемые валюты для текущей сети
func (c *Config) GetSupportedCurrencies() []string {
	switch c.Network {
	case NetworkTON:
		return []string{"TON", "USDT"}
	case NetworkEVM:
		return []string{"ETH", "USDT", "USDC"}
	default:
		return []string{"TON", "USDT"}
	}
}

// ParseEnvironment парсит строку в Environment
func ParseEnvironment(s string) Environment {
	switch strings.ToLower(s) {
	case "prod", "production":
		return EnvProd
	default:
		return EnvDev
	}
}

// ParseNetwork парсит строку в Network
func ParseNetwork(s string) Network {
	switch strings.ToLower(s) {
	case "evm", "ethereum", "eth":
		return NetworkEVM
	default:
		return NetworkTON
	}
}
