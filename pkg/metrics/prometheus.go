package metrics

import (
	"log"

	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MetricsRegistry содержит все счетчики и метрики проекта
type MetricsRegistry struct {
	// HTTP метрики
	RequestDuration *prometheus.HistogramVec
	RequestTotal    *prometheus.CounterVec
	ErrorsTotal     *prometheus.CounterVec

	// Бизнес-метрики
	GameStartTotal      prometheus.Counter
	GameCompleteTotal   prometheus.Counter
	GameAbandonedTotal  prometheus.Counter
	ActiveGamesGauge    prometheus.Gauge
	WordGuessedCounters *prometheus.CounterVec
}

// NewMetricsRegistry создает и регистрирует все метрики
func NewMetricsRegistry() *MetricsRegistry {
	registry := &MetricsRegistry{
		// HTTP метрики
		RequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "Длительность HTTP запросов",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "endpoint", "status"},
		),
		RequestTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Количество HTTP запросов",
			},
			[]string{"method", "endpoint", "status"},
		),
		ErrorsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "errors_total",
				Help: "Количество ошибок",
			},
			[]string{"type"},
		),

		// Бизнес-метрики
		GameStartTotal: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "wordle_game_start_total",
				Help: "Количество начатых игр",
			},
		),
		GameCompleteTotal: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "wordle_game_complete_total",
				Help: "Количество завершенных игр",
			},
		),
		GameAbandonedTotal: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "wordle_game_abandoned_total",
				Help: "Количество брошенных игр",
			},
		),
		ActiveGamesGauge: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "wordle_active_games",
				Help: "Количество активных игр",
			},
		),
		WordGuessedCounters: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "wordle_word_guessed_total",
				Help: "Статистика угаданных слов",
			},
			[]string{"attempts"},
		),
	}

	return registry
}

// InitMetrics инициализирует систему метрик Prometheus
func InitMetrics(enabled bool, port string) *MetricsRegistry {
	if !enabled {
		log.Printf("Metrics disabled")
		return nil
	}

	if port == "" {
		port = "9090"
	}

	registry := NewMetricsRegistry()
	SetRegistry(registry)

	// Запуск HTTP сервера для Prometheus метрик
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		log.Printf("Starting metrics server on :%s", port)
		if err := http.ListenAndServe(":"+port, nil); err != nil {
			log.Printf("Error starting metrics server: %v", err)
		}
	}()

	return registry
}
