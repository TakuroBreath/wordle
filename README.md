Техническое задание: Бэкенд on-chain Wordle Game (Telegram Mini App)

Цель

Разработать backend-приложение на Go, обслуживающее игру Wordle с криптоставками, реализованную как Telegram Mini App. Бэкенд отвечает за хранение данных, валидацию логики игры, обработку ставок, интеграцию с блокчейном и выдачу наград.

⸻

Технологии
	•	Язык: Go 1.24+
	•	Хранилище: PostgreSQL, In-Memory (для сессий и кэша)
	•	Blockchain: TON (через SDK / gRPC / REST)
	•	API: REST (JSON over HTTPS), Gin
	•	Аутентификация: через Telegram Mini App авторизацию
	•	Observability:
		• Метрики: Prometheus + Grafana
		• Логирование: Loki + Promtail + Grafana

⸻

Основные сущности

1. User (пользователь)
	•	telegram_id: uint64 (уникальный)
	•	username: string
	•	first_name: string
	•	last_name: string
	•	wallet: string
	•	balance_ton: float64
	•	balance_usdt: float64
	•	wins: int
	•	losses: int
	•	created_at: datetime
	•	updated_at: datetime

2. Game (игра)
	•	id: UUID
	•	creator_id: uint64 (Telegram ID создателя)
	•	word: string
	•	length: int
	•	difficulty: string
	•	max_tries: int
	•	title: string
	•	description: text
	•	min_bet: float64
	•	max_bet: float64
	•	reward_multiplier: float64
	•	currency: string ("TON" или "USDT")
	•	reward_pool_ton: float64
	•	reward_pool_usdt: float64
	•	status: string ("active" или "inactive")
	•	created_at: datetime
	•	updated_at: datetime

3. Lobby (игровая сессия)
	•	id: UUID
	•	game_id: UUID
	•	user_id: uint64 (Telegram ID игрока)
	•	max_tries: int
	•	tries_used: int
	•	bet_amount: float64
	•	potential_reward: float64
	•	status: string ("active" или "inactive")
	•	created_at: datetime
	•	updated_at: datetime
	•	expires_at: datetime
	•	attempts: []Attempt

4. Attempt (попытка)
	•	id: UUID
	•	game_id: UUID
	•	user_id: uint64 (Telegram ID игрока)
	•	word: string
	•	result: []int (0 - нет, 1 - есть в слове, 2 - на месте)
	•	created_at: datetime
	•	updated_at: datetime

⸻

Бизнес-логика

Создание пользователя
	•	При первом входе через Telegram Mini App пользователь создаётся на основе telegram_id.
	•	Баланс = 0, побед и поражений нет.

Создание игры
	•	Игрок задаёт слово, макс. количество попыток, множитель награды, минимальную и максимальную ставки.
	•	Высчитывается max_reward = max_bet * multiplier.
	•	Создатель переводит депозит(reward pool) = max_reward (На фронтенде)
	•	После успешной транзакции игра становится активной и доступной другим игрокам.

Присоединение к игре
	•	Игрок указывает свою ставку (>= min_bet, <= max_bet).
	•	Создаётся Lobby, высчитывается потенциальный выигрыш.
	•	Игрок начинает игру с пустым полем попыток.

Попытка
	•	Игрок отправляет слово (client-side валидация длины).
	•	Бэкенд возвращает массив из int:
	•	0 — буквы нет
	•	1 — буква есть, не на месте
	•	2 — буква есть и на месте
	•	Попытка сохраняется.
	•	Если слово угадано — статус лобби = success, начисляется выигрыш на баланс игрока.
	•	Если попытки исчерпаны или время вышло — статус = failed.
	•	Если игра проиграна, ставка начисляется в reward pool игры.

Выплата наград
	•	Победа: сумма начисляется игроку, вычитается у создателя.
	•	Проигрыш: остаётся у создателя.
	•	Комиссия: % остаётся в сервисе (константа, например 5%).
	•	Создатель игры может удалить игру, ее reward pool переходит к нему на баланс.
	•	Игрок может запросить выплату наград со своего баланса. Его баланс проверяется.

⸻

API (эндпоинты)

Auth
	•	POST /auth/telegram — регистрация по данным Telegram Mini App
	•	GET /auth/verify — проверка подписи Telegram Mini App

User
	•	GET /me — получение данных текущего пользователя
	•	GET /user/{id} — получение данных пользователя по ID
	•	GET /user/balance — получение баланса пользователя
	•	POST /user/withdraw — запрос на вывод средств
	•	GET /user/withdraw/history — история выводов

Game
	•	POST /game — создание новой игры
	•	GET /game/{id} — получение информации об игре
	•	GET /games/active — список активных игр
	•	GET /games/my — список игр, созданных пользователем
	•	DELETE /game/{id} — удаление игры (только для создателя)
	•	POST /game/{id}/reward-pool — пополнение reward pool игры

Lobby
	•	POST /lobby/join — присоединение к игре
	•	GET /lobby/{id} — получение информации о лобби
	•	GET /lobby/active — получение активного лобби пользователя
	•	POST /lobby/{id}/attempt — отправка попытки угадать слово
	•	GET /lobby/{id}/attempts — получение истории попыток

Admin / Service
	•	GET /stats — общая статистика
	•	GET /stats/games — статистика по играм
	•	GET /stats/users — статистика по пользователям
	•	GET /rewards/pending — список ожидающих выплат
	•	POST /rewards/process — обработка выплаты
	•	GET /rewards/history — история выплат

⸻

Безопасность
	•	Telegram ID подписывается на фронте, проверяется на бэке.
	•	Все транзакции блокчейна подписываются пользователем.
	•	Игра не активируется до подтверждения депозита от создателя.
	•	Все балансы хранятся on-chain, либо периодически синхронизируются.

⸻

Потенциальные доработки (V2+)
	•	NFT-награды
	•	Турниры
	•	История игр
	•	Реплей попыток
	•	Механика "подсказок за комиссию"

## Запуск проекта

### Локальный запуск в Docker (со всеми сервисами)

Для запуска проекта локально со всеми сервисами (backend, frontend, PostgreSQL, Prometheus, Grafana, Loki):

1. **Клонируйте репозиторий:**
```bash
git clone https://github.com/TakuroBreath/wordle.git
cd wordle
```

2. **Запустите все сервисы:**
```bash
docker-compose -f docker-compose.full.yml up -d
```

3. **Проверьте статус контейнеров:**
```bash
docker-compose -f docker-compose.full.yml ps
```

4. **Просмотр логов:**
```bash
# Все логи
docker-compose -f docker-compose.full.yml logs -f

# Только backend
docker-compose -f docker-compose.full.yml logs -f app
```

#### Доступные сервисы (локально)

После запуска будут доступны следующие сервисы:

- **API Backend**: http://localhost:8080
- **Frontend**: http://localhost:3000
- **Grafana** (мониторинг): http://localhost:3000 (admin/secret)
- **Prometheus** (метрики): http://localhost:9090
- **Loki** (логи): http://localhost:3100
- **PostgreSQL**: localhost:5432

#### Структура контейнеров

- **wordle-api** - основное приложение (Go)
- **wordle-postgres** - база данных PostgreSQL
- **prometheus** - сбор метрик
- **grafana** - визуализация метрик и логов
- **loki** - хранилище логов
- **promtail** - сборщик логов

#### Варианты локального запуска

**Только backend + базы данных (без мониторинга):**
```bash
docker-compose up -d
```

**Backend + Frontend (для разработки):**
```bash
docker-compose -f docker-compose.dev.yml up -d
```

**Все сервисы включая мониторинг:**
```bash
docker-compose -f docker-compose.full.yml up -d
```

### Запуск в продакшене

Для запуска в продакшене выполните следующие шаги:

1. **Скопируйте шаблон конфигурации:**
```bash
cp configs/config.docker.prod.yaml configs/config.docker.local.yaml
```

2. **Заполните секреты в `configs/config.docker.local.yaml`:**
   - `postgres.password` - сильный пароль для PostgreSQL
   - `auth.jwt_secret` - JWT секрет (минимум 32 символа, сгенерируйте: `openssl rand -hex 32`)
   - `auth.bot_token` - токен Telegram бота (получите у @BotFather)
   - `blockchain.ton.api_key` - API ключ от toncenter.com
   - `blockchain.ton.master_wallet` - адрес кастодиального кошелька
   - `blockchain.ton.master_wallet_seed` - seed фраза кошелька (24 слова)

3. **Обновите docker-compose.yml для использования продакшн конфига:**
   
   Отредактируйте `docker-compose.yml` и измените путь к конфигу:
   ```yaml
   volumes:
     - ./configs/config.docker.local.yaml:/app/config.yaml:ro
   ```

4. **Настройте переменные окружения для PostgreSQL:**
   
   Отредактируйте `docker-compose.yml` и обновите пароль:
   ```yaml
   postgres:
     environment:
       POSTGRES_PASSWORD: YOUR_STRONG_PASSWORD
   ```

5. **Запустите в продакшене:**
```bash
docker-compose up -d
```

6. **Проверьте логи:**
```bash
docker-compose logs -f app
```

#### Важные замечания для продакшена

- ⚠️ **Никогда не коммитьте** `configs/config.docker.local.yaml` в git (он уже в `.gitignore`)
- 🔒 Используйте сильные пароли для всех сервисов
- 🔐 Храните seed фразу кошелька в безопасном месте (используйте секреты Docker или внешний менеджер секретов)
- 🌐 Для продакшена рекомендуется использовать внешнюю базу данных
- 📊 Настройте регулярные бэкапы базы данных
- 🛡️ Используйте SSL/TLS для подключения к PostgreSQL в продакшене (`ssl_mode: require`)

### Управление контейнерами

**Остановка:**
```bash
docker-compose down
# или для полного стека
docker-compose -f docker-compose.full.yml down
```

**Перезапуск сервиса:**
```bash
docker-compose restart app
```

**Пересборка и запуск:**
```bash
docker-compose up -d --build
```

**Удаление всех данных (⚠️ осторожно!):**
```bash
docker-compose down -v
```

## Миграции

Приложение использует [golang-migrate](https://github.com/golang-migrate/migrate) для управления миграциями базы данных.

### Миграции базы данных

Миграции запускаются автоматически при старте контейнера через `entrypoint.sh`.

**Ручной запуск миграций:**

```bash
# Запуск всех миграций
docker-compose exec app migrate -path=/app/migrations -database postgres://postgres:postgres@postgres:5432/wordle?sslmode=disable up

# Откат последней миграции
docker-compose exec app migrate -path=/app/migrations -database postgres://postgres:postgres@postgres:5432/wordle?sslmode=disable down 1

# Статус миграций
docker-compose exec app migrate -path=/app/migrations -database postgres://postgres:postgres@postgres:5432/wordle?sslmode=disable version
```

**Для продакшена** используйте соответствующие учетные данные из конфига.

## Система наблюдаемости (Observability)

### Метрики (Prometheus + Grafana)

Prometheus собирает метрики из приложения через HTTP-эндпоинт `/metrics`. Grafana используется для визуализации.

#### Собираемые метрики

1. **HTTP-метрики**:
   - `http_request_duration_seconds` - время выполнения запросов
   - `http_requests_total` - количество запросов
   - `errors_total` - количество ошибок

2. **Игровые метрики**:
   - `wordle_game_start_total` - количество начатых игр
   - `wordle_game_complete_total` - количество завершенных игр
   - `wordle_game_abandoned_total` - количество брошенных игр
   - `wordle_active_games` - текущее количество активных игр
   - `wordle_word_guessed_total` - статистика угаданных слов по количеству попыток

3. **Финансовые метрики (Business)**:
   - `wordle_deposits_total_amount` - общая сумма депозитов
   - `wordle_withdrawals_total_amount` - общая сумма выводов
   - `wordle_commissions_total_amount` - комиссии (revenue проекта)
   - `wordle_bets_total_amount` - общая сумма ставок
   - `wordle_rewards_total_amount` - общая сумма выплат
   - `wordle_pending_withdrawals` - ожидающие обработки выводы

### Логирование (Loki + Promtail)

Loki - лёгкая система агрегации логов. Приложение использует zap логгер с двойным выводом:
- **Console (stdout)** - человекочитаемый формат для `docker logs`
- **JSON (файл)** - структурированные логи для Loki

Просмотр логов доступен через Grafana Log Explorer.

## Документация

- [Полная система наблюдаемости](./observability-README.md)
- [Система мониторинга](./monitoring-README.md)

## Бизнес-логика и сценарии

### Основные сущности

1.  **User (Пользователь)**: Участник системы, идентифицируемый через Telegram ID. Имеет балансы в TON и USDT, статистику побед/поражений и привязанный кошелек.
2.  **Game (Игра)**: Шаблон игры, созданный пользователем. Содержит загаданное слово, параметры сложности, лимиты ставок и мультипликатор награды. Каждая игра имеет свой "Reward Pool" (пул наград), из которого выплачиваются выигрыши.
3.  **Lobby (Игровая сессия)**: Конкретная сессия игры игрока. Связывает игрока с игрой, отслеживает ставку, потенциальную награду, количество использованных попыток и статус.
4.  **Attempt (Попытка)**: Одиночный ввод слова игроком в рамках лобби. Хранит результат проверки каждой буквы (0 - нет, 1 - не на месте, 2 - на месте).

### Бизнес-сценарии

#### 1. Регистрация и вход
Пользователь заходит через Telegram Mini App. Бэкенд проверяет подпись Telegram и создает запись пользователя, если её нет. Балансы изначально нулевые.

#### 2. Создание игры (Creator)
- Создатель указывает слово, сложность, лимиты ставок (min/max) и мультипликатор.
- Бэкенд вычисляет необходимый депозит: `max_bet * reward_multiplier`.
- Игра создается в статусе `pending`. Создатель должен отправить депозит на мастер-кошелек сервиса.
- `BlockchainWorker` отслеживает транзакцию с нужным комментарием и переводит игру в статус `active`.

#### 3. Участие в игре (Player)
- Игрок выбирает активную игру и указывает ставку.
- Игрок может оплатить ставку с внутреннего баланса или напрямую через блокчейн (TON).
- При создании лобби потенциальная награда (`bet * multiplier`) резервируется в пуле игры (`ReservedAmount`). Это гарантирует выплату выигрыша.
- Нельзя играть в свою собственную игру.

#### 4. Игровой процесс
- Игрок отправляет попытки угадать слово.
- Бэкенд проверяет слово и возвращает результат для каждой буквы.
- Если слово угадано полностью или попытки исчерпаны/время вышло, лобби завершается.

#### 5. Финансовые расчеты и награды
- **Победа игрока**:
  - Игрок получает `net_reward = (bet * multiplier) * (1 - commission)`.
  - Сумма списывается с баланса текущего пользователя и из пула игры (creator).
  - Комиссия (5%) удерживается сервисом.
- **Проигрыш игрока**:
  - Ставка игрока (за вычетом комиссии сервиса) добавляется в пул наград игры (`RewardPool`). Таким образом, создатель зарабатывает на проигрышах игроков.
  - Комиссия (5%) удерживается сервисом.

### Оценка готовности к продакшену (Production Readiness)

На текущий момент проект имеет высокую степень готовности, но требует внимания к следующим аспектам:

1.  **Положительные стороны**:
    - Полная реализация ядра игровой логики.
    - Интеграция с TON блокчейном (входящие платежи, выводы TON).
    - Система фоновых воркеров для обработки транзакций.
    - Настроенный стек наблюдаемости (Prometheus, Grafana, Loki).
    - Миграции базы данных и контейнеризация.

2.  **Что нужно доработать перед полноценным запуском**:
    - **USDT Jetton**: В текущей версии реализованы только TON-транзакции. Метод `SendUSDT` в `TONService` требует реализации для полноценной поддержки USDT.
    - **Безопасность кошелька**: Seed-фраза мастер-кошелька передается через конфиг. В продакшене рекомендуется использовать Secret Manager (HashiCorp Vault, AWS Secrets Manager) или хотя бы Docker Secrets.
    - **Лимиты и Anti-fraud**: Требуется внедрение rate-limiting на API и базовых проверок на подозрительную активность (например, слишком быстрые попытки).
    - **Интеграционные тесты**: Желательно добавить тесты, имитирующие работу с реальным блокчейном (Testnet) для воркера.
    - **Фронтенд**: Требуется финальная проверка интеграции с актуальными эндпоинтами бэкенда.

**Итоговая оценка**: 8/10. Проект готов к запуску в режиме "Beta" в Testnet. Для полноценного Mainnet-запуска необходимо завершить поддержку USDT и усилить безопасность управления ключами.
