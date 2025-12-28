# Конфигурация Wordle Betting

## Структура конфигурации

Проект использует YAML файлы для конфигурации. Доступны следующие файлы:

```
configs/
├── config.dev.yaml     # Конфигурация для разработки
├── config.prod.yaml    # Конфигурация для production
└── config.local.yaml   # Локальная конфигурация (не коммитится в git)
```

## Выбор конфигурации

Приложение ищет конфиг в следующем порядке:

1. `CONFIG_PATH` - переменная окружения с путём к файлу
2. `./config.yaml` - корень проекта
3. `./configs/config.local.yaml` - локальный конфиг
4. `./configs/config.dev.yaml` - dev конфиг

### Запуск с конкретным конфигом

```bash
# Через переменную окружения
CONFIG_PATH=./configs/config.prod.yaml ./wordle

# Или создайте symlink
ln -s configs/config.dev.yaml config.yaml
./wordle
```

## Основные параметры

### Окружение (environment)

| Значение | Описание |
|----------|----------|
| `dev` | Режим разработки. Авторизация отключена, подробное логирование |
| `prod` | Production режим. Авторизация включена, JSON логи |

### Сеть (network)

| Значение | Описание | Авторизация |
|----------|----------|-------------|
| `ton` | TON блокчейн, Telegram Mini App | Через Telegram (в prod) |
| `evm` | Ethereum/EVM совместимые сети | Отключена (в разработке) |

### Mock провайдер (use_mock_provider)

| Значение | Описание |
|----------|----------|
| `true` | Использовать тестовый провайдер (не взаимодействует с блокчейном) |
| `false` | Использовать реальный провайдер блокчейна |

## Правила окружений

### Dev режим (`environment: dev`)
- Авторизация **всегда отключена**
- Используется mock пользователь для тестирования
- Подробные логи в консоль

### Prod режим (`environment: prod`)
- Авторизация **всегда включена**
- Mock провайдер **отключён** (use_mock_provider игнорируется)
- JSON логи для ELK

### TON сеть (`network: ton`)
- Приложение работает как Telegram Mini App
- Авторизация через Telegram initData
- Поддерживаемые валюты: TON, USDT

### EVM сеть (`network: evm`)
- Авторизация **отключена** (в разработке - будет через кошелёк)
- Поддерживаемые валюты: ETH, USDT, USDC

## Пример конфигурации

```yaml
# Окружение и сеть
environment: dev
network: ton
use_mock_provider: true

# HTTP сервер
http:
  port: "8080"
  read_timeout: 10s
  write_timeout: 10s
  idle_timeout: 60s

# PostgreSQL
postgres:
  host: localhost
  port: "5432"
  user: postgres
  password: postgres
  db_name: wordle
  ssl_mode: disable

# Redis
redis:
  host: localhost
  port: "6379"
  password: ""
  db: 0

# Авторизация
auth:
  enabled: false          # Управляется автоматически на основе environment
  jwt_secret: secret_key
  bot_token: ""
  token_ttl: 24h

# Метрики
metrics:
  enabled: true
  port: "9090"

# Логирование
logging:
  level: debug            # debug, info, warn, error
  format: console         # console, json
  output: stdout

# Блокчейн
blockchain:
  ton:
    api_endpoint: https://toncenter.com/api/v2
    api_key: ${TON_API_KEY}
    master_wallet: ${TON_MASTER_WALLET}
    min_withdraw: 0.1
    withdraw_fee: 0.01
    testnet: false

  ethereum:
    rpc_url: ${ETH_RPC_URL}
    chain_id: 1
    master_wallet: ${ETH_MASTER_WALLET}
    min_withdraw: 0.01
    withdraw_fee: 0.001
```

## Подстановка переменных окружения

В YAML файлах можно использовать синтаксис `${VAR}` для подстановки переменных окружения:

```yaml
postgres:
  password: ${POSTGRES_PASSWORD}
  
auth:
  jwt_secret: ${JWT_SECRET}
```

Если переменная не найдена, значение остаётся как `${VAR}`.

## Валидация конфигурации

В production режиме проверяются:

1. `jwt_secret` - должен быть установлен и не равен `dev_secret_key`
2. `bot_token` - обязателен для сети TON
3. Настройки блокчейна (если не используется mock):
   - TON: `api_key`, `master_wallet`
   - EVM: `rpc_url`, `master_wallet`

## Быстрый старт

### Для локальной разработки

```bash
# 1. Скопируйте локальный конфиг
cp configs/config.local.yaml config.yaml

# 2. Отредактируйте config.yaml при необходимости

# 3. Запустите приложение
go run cmd/api/main.go
```

### Для production

```bash
# 1. Установите переменные окружения
export CONFIG_PATH=./configs/config.prod.yaml
export JWT_SECRET=your_secure_secret
export BOT_TOKEN=your_telegram_bot_token
export POSTGRES_HOST=db.example.com
export POSTGRES_PASSWORD=secure_password
export TON_API_KEY=your_tonapi_key
export TON_MASTER_WALLET=EQ...

# 2. Запустите приложение
./wordle
```

### Docker

```dockerfile
ENV CONFIG_PATH=/app/configs/config.prod.yaml
COPY configs/config.prod.yaml /app/configs/
```
