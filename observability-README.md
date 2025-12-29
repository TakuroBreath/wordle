# Система наблюдаемости Wordle Betting

## Обзор

Система наблюдаемости (Observability) для Wordle Betting состоит из двух основных компонентов:

1. **Метрики** - Prometheus + Grafana
2. **Логирование** - Loki + Promtail + Grafana

Это лёгкое и эффективное решение обеспечивает полную видимость работы приложения:
- Отслеживание производительности и доступности сервисов
- Анализ ошибок и проблем в режиме реального времени
- Визуализация технических и бизнес-метрик
- Централизованный просмотр логов

## Архитектура системы

```
                    ┌─────────────────────────────────────────────┐
                    │                 GRAFANA                      │
                    │  ┌─────────────┐    ┌─────────────────────┐ │
                    │  │ Dashboards  │    │    Log Explorer     │ │
                    │  └─────────────┘    └─────────────────────┘ │
                    └─────────────────────────────────────────────┘
                              ▲                      ▲
                              │                      │
                    ┌─────────┴─────────┐  ┌────────┴────────┐
                    │    PROMETHEUS     │  │       LOKI      │
                    │   (Time Series)   │  │  (Log Storage)  │
                    └───────────────────┘  └─────────────────┘
                              ▲                      ▲
                              │                      │
                    ┌─────────┴─────────┐  ┌────────┴────────┐
                    │  /metrics (HTTP)  │  │    PROMTAIL     │
                    │                   │  │  (Log Shipper)  │
                    └───────────────────┘  └─────────────────┘
                              ▲                      ▲
                              │                      │
                    ┌─────────┴──────────────────────┴────────┐
                    │                                          │
                    │              WORDLE API                   │
                    │                                          │
                    │   stdout ────► Console logs (human)      │
                    │   file ──────► JSON logs (Loki)          │
                    │   /metrics ──► Prometheus metrics        │
                    │                                          │
                    └──────────────────────────────────────────┘
```

## Компоненты

### Метрики (Prometheus + Grafana)

Prometheus собирает метрики из приложения через HTTP-эндпоинт `/metrics`.

#### HTTP-метрики

| Метрика | Тип | Описание |
|---------|-----|----------|
| `http_request_duration_seconds` | Histogram | Время выполнения HTTP-запросов |
| `http_requests_total` | Counter | Общее количество HTTP-запросов |
| `errors_total` | Counter | Количество ошибок по типам |

#### Игровые метрики

| Метрика | Тип | Описание |
|---------|-----|----------|
| `wordle_game_start_total` | Counter | Количество начатых игр |
| `wordle_game_complete_total` | Counter | Количество успешно завершённых игр |
| `wordle_game_abandoned_total` | Counter | Количество брошенных игр |
| `wordle_active_games` | Gauge | Текущее количество активных игр |
| `wordle_word_guessed_total` | Counter | Статистика угаданных слов (по кол-ву попыток) |

#### Финансовые метрики (Business)

| Метрика | Тип | Labels | Описание |
|---------|-----|--------|----------|
| `wordle_deposits_total_amount` | Counter | currency | Общая сумма депозитов |
| `wordle_deposits_count` | Counter | currency | Количество депозитов |
| `wordle_withdrawals_total_amount` | Counter | currency | Общая сумма выводов |
| `wordle_withdrawals_count` | Counter | currency | Количество выводов |
| `wordle_commissions_total_amount` | Counter | currency | Комиссии (revenue проекта) |
| `wordle_bets_total_amount` | Counter | currency | Общая сумма ставок |
| `wordle_bets_count` | Counter | currency | Количество ставок |
| `wordle_rewards_total_amount` | Counter | currency | Общая сумма выплат |
| `wordle_rewards_count` | Counter | currency | Количество выплат |
| `wordle_pending_withdrawals` | Gauge | - | Ожидающие обработки выводы |
| `wordle_total_users_balance` | Gauge | currency | Общий баланс пользователей |

### Логирование (Loki + Promtail)

Loki - это лёгкая система агрегации логов от Grafana Labs. В отличие от ELK Stack, Loki не индексирует содержимое логов, а только метаданные (labels), что делает его значительно легче и быстрее.

#### Конфигурация логгера

Приложение использует `zap` логгер с двойным выводом:

1. **Console (stdout)** - человекочитаемый формат для просмотра в контейнере
2. **JSON (файл)** - структурированные логи для Loki

```yaml
# Пример конфигурации
logging:
  level: info
  isProduction: true
  lokiFilePath: /var/log/wordle/app.json  # JSON логи для Loki
  disableConsole: false                     # Console в stdout
  disableLokiFile: false                    # JSON в файл
```

#### Структура JSON логов

```json
{
  "timestamp": "2024-01-15T12:34:56.789Z",
  "level": "info",
  "message": "Processing deposit",
  "caller": "transaction_service.go:123",
  "user_id": "abc-123",
  "amount": 100.5,
  "currency": "TON"
}
```

#### Просмотр логов в Grafana

В Grafana доступен Log Explorer с поддержкой LogQL запросов:

```logql
# Все логи приложения
{job="wordle"}

# Только ошибки
{job="wordle"} |= "error"

# Логи определённого уровня
{job="wordle"} | json | level="error"

# Поиск по сообщению
{job="wordle"} |= "deposit" | json
```

## Запуск системы

### Минимальный вариант (только приложение)

```bash
docker-compose up -d
```

### С мониторингом (Prometheus + Grafana + Loki)

```bash
docker-compose -f docker-compose.full.yml up -d
```

### Только мониторинг (добавить к запущенному приложению)

```bash
docker-compose -f docker-compose.monitoring.yml up -d
```

## Доступ к компонентам

| Компонент | URL | Логин/Пароль |
|-----------|-----|--------------|
| Grafana | http://localhost:3000 | admin / secret |
| Prometheus | http://localhost:9090 | - |
| Loki API | http://localhost:3100 | - |
| App Metrics | http://localhost:9091/metrics | - |

## Grafana Dashboards

Предустановленный дашборд **Wordle Betting Dashboard** включает:

### HTTP Metrics
- Requests per second (RPS)
- Request duration (p95)

### Game Metrics
- Active games
- Games started / completed / abandoned
- Word guessed distribution

### Financial Metrics (Business)
- Total deposits / withdrawals
- Revenue (commissions)
- Pending withdrawals
- Deposits vs Withdrawals by currency
- Bets vs Rewards
- Total users balance

### Application Logs
- Live log stream
- Error logs filter
- Log level distribution

## Просмотр логов в контейнере

Благодаря console выводу, логи можно читать прямо в контейнере:

```bash
# Посмотреть логи в реальном времени
docker logs -f wordle-api

# Пример вывода (человекочитаемый формат):
# 12:34:56.789  INFO  Processing deposit  {"user_id": "abc", "amount": 100}
# 12:34:56.890  ERROR Failed to process  {"error": "insufficient funds"}
```

## Файловая структура

```
configs/
├── grafana/
│   ├── dashboards/
│   │   └── wordle.json           # Дашборд
│   └── provisioning/
│       ├── dashboards/
│       │   └── dashboards.yml    # Автозагрузка дашбордов
│       └── datasources/
│           ├── prometheus.yml    # Datasource Prometheus
│           └── loki.yml          # Datasource Loki
├── loki/
│   └── loki-config.yaml          # Конфигурация Loki
├── prometheus/
│   └── prometheus.yml            # Конфигурация Prometheus
└── promtail/
    └── promtail-config.yaml      # Конфигурация Promtail
```

## Конфигурация

### Prometheus targets

```yaml
# configs/prometheus/prometheus.yml
scrape_configs:
  - job_name: 'wordle-app'
    static_configs:
      - targets: ['app:9090']
    metrics_path: '/metrics'
```

### Promtail scrape

```yaml
# configs/promtail/promtail-config.yaml
scrape_configs:
  - job_name: wordle_app
    static_configs:
      - targets: [localhost]
        labels:
          job: wordle
          __path__: /var/log/wordle/*.json
```

## Retention (хранение данных)

| Компонент | Retention | Настройка |
|-----------|-----------|-----------|
| Prometheus | 15 дней | `--storage.tsdb.retention.time` |
| Loki | 7 дней | `compactor.retention_period` |

## Рекомендации

1. **Мониторинг в реальном времени** - используйте Grafana дашборды
2. **Анализ ошибок** - используйте Log Explorer в Grafana с фильтром по level="error"
3. **Бизнес-аналитика** - отслеживайте финансовые метрики для понимания unit-экономики
4. **Отладка в контейнере** - используйте `docker logs -f` для быстрого просмотра

## Почему Loki вместо ELK?

| Характеристика | ELK Stack | Loki |
|----------------|-----------|------|
| Потребление RAM | ~2-4 GB | ~200-500 MB |
| Потребление диска | Высокое (индексы) | Низкое |
| Сложность настройки | Высокая | Низкая |
| Интеграция с Grafana | Через плагин | Нативная |
| Полнотекстовый поиск | Да | Нет (но есть grep-like) |
| Стоимость | Бесплатно / Платно | Бесплатно |

Для данного проекта Loki является оптимальным выбором, так как:
- Проект небольшой, не требует полнотекстового поиска
- Grafana уже используется для метрик
- Минимальное потребление ресурсов
