# Конфигурация Wordle Betting

## Быстрый старт

```bash
# Dev режим (по умолчанию)
make dev

# Или просто
make

# Prod режим
make prod
```

## Структура

```
configs/
├── config.dev.yaml     # Dev - без авторизации, mock блокчейн
└── config.prod.yaml    # Prod - с авторизацией, реальный блокчейн
```

Приложение ищет `config.yaml` в корне проекта. Makefile автоматически создаёт symlink на нужный конфиг.

## Основные параметры

| Параметр | Dev | Prod |
|----------|-----|------|
| `environment` | dev | prod |
| `network` | ton | ton |
| `use_mock_provider` | true | false |
| Авторизация | ❌ | ✅ |

## Режимы

### Dev (`config.dev.yaml`)
- Авторизация отключена
- Используется mock блокчейн провайдер
- Debug логи в консоль
- Метрики отключены

### Prod (`config.prod.yaml`)  
- Авторизация через Telegram
- Реальный блокчейн (TON)
- JSON логи
- Метрики включены

## Настройка для Prod

Отредактируйте `configs/config.prod.yaml`:

```yaml
auth:
  jwt_secret: ваш_секретный_ключ
  bot_token: "123456789:ABCdefGHIjklMNOpqrsTUVwxyz"

blockchain:
  ton:
    api_key: "ваш_tonapi_ключ"
    master_wallet: "EQxxxx..."
```

## Сеть

| network | Описание | Валюты |
|---------|----------|--------|
| `ton` | Telegram Mini App | TON, USDT |
| `evm` | Web (в разработке) | ETH, USDT, USDC |
