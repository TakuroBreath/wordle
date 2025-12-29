# Конфигурация Wordle Betting

## Структура файлов

```
configs/
├── config.dev.yaml           # Dev - локальная разработка без Docker
├── config.prod.yaml          # Prod - шаблон с плейсхолдерами
├── config.local.yaml.example # Шаблон для создания config.local.yaml
├── config.docker.yaml        # Docker dev - для docker-compose
├── config.docker.prod.yaml   # Docker prod - шаблон с плейсхолдерами
└── README.md
```

## Приоритет загрузки

1. `CONFIG_PATH` env variable (используется в Docker)
2. `./config.yaml` (symlink в корне проекта)
3. `./configs/config.local.yaml` (ваши секреты)
4. `./configs/config.dev.yaml` (дефолт)

## Быстрый старт

### Локальная разработка (без Docker)

```bash
# Используется config.dev.yaml автоматически
make run

# Или с symlink на конкретный конфиг
make dev   # symlink на config.dev.yaml
make prod  # symlink на config.prod.yaml
```

### Docker разработка

```bash
# Использует configs/config.docker.yaml
docker-compose -f docker-compose.dev.yml up -d
```

### Production

1. Скопируйте шаблон:
   ```bash
   cp configs/config.docker.prod.yaml configs/config.docker.local.yaml
   ```

2. Заполните секреты в `config.docker.local.yaml`

3. Обновите docker-compose.yml:
   ```yaml
   volumes:
     - ./configs/config.docker.local.yaml:/app/config.yaml:ro
   ```

4. Запустите:
   ```bash
   docker-compose up -d
   ```

## Секреты

⚠️ **Никогда не коммитьте файлы с реальными секретами!**

Файлы с секретами добавлены в `.gitignore`:
- `configs/config.local.yaml`
- `configs/config.docker.local.yaml`

### Как получить секреты

| Секрет | Где получить |
|--------|--------------|
| `jwt_secret` | Генерация: `openssl rand -hex 32` |
| `bot_token` | Telegram: @BotFather |
| `ton.api_key` | https://toncenter.com/ |
| `ton.master_wallet` | Ваш TON кошелёк |
| `ton.master_wallet_seed` | 24 слова при создании кошелька |

## Режимы

### Dev (`environment: dev`)
- Авторизация отключена автоматически
- Mock blockchain провайдер
- Debug логи в консоль
- Метрики опционально

### Prod (`environment: prod`)
- Авторизация включена автоматически
- Реальный блокчейн (TON/EVM)
- JSON логи
- Метрики включены

## Основные параметры

| Параметр | Dev | Prod |
|----------|-----|------|
| `environment` | dev | prod |
| `network` | ton | ton |
| `use_mock_provider` | true | false |
| Авторизация | ❌ | ✅ |
| `logging.level` | debug | info |
| `logging.format` | console | json |

## Сеть

| network | Описание | Валюты |
|---------|----------|--------|
| `ton` | Telegram Mini App | TON, USDT |
| `evm` | Web (в разработке) | ETH, USDT, USDC |

## Пример полной конфигурации

```yaml
environment: prod
network: ton
use_mock_provider: false

http:
  port: "8080"
  read_timeout: 15s
  write_timeout: 15s
  idle_timeout: 120s

postgres:
  host: localhost
  port: "5432"
  user: wordle_user
  password: your_strong_password
  db_name: wordle
  ssl_mode: require

redis:
  host: localhost
  port: "6379"
  password: redis_password
  db: 0

auth:
  enabled: true
  jwt_secret: your_32_char_jwt_secret_here
  bot_token: "123456789:ABCdefGHIjklMNOpqrsTUVwxyz"
  token_ttl: 24h

metrics:
  enabled: true
  port: "9090"

logging:
  level: info
  format: json
  output: stdout

blockchain:
  ton:
    api_endpoint: https://toncenter.com/api/v3
    api_key: your_api_key
    master_wallet: EQDxxxxxxx
    master_wallet_seed: "word1 word2 ... word24"
    min_withdraw_ton: 0.5
    min_withdraw_usdt: 5.0
    withdraw_fee_ton: 0.05
    withdraw_fee_usdt: 0.5
    required_confirmations: 1
    testnet: false
    usdt_master_address: EQCxE6mUtQJKFnGfaROTKOt1lZbDiiX1kCixRv7Nw2Id_sDs
    worker_poll_interval: 5
    commission_rate: 0.05
```
