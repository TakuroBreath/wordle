package config

import (
	"time"

	"github.com/TakuroBreath/wordle/internal/logger"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

// Config представляет конфигурацию приложения
type Config struct {
	HTTP     HTTPConfig
	Postgres PostgresConfig
	Redis    RedisConfig
	Auth     AuthConfig
	Metrics  MetricsConfig
	Logging  logger.Config
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
