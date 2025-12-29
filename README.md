Техническое задание: Бэкенд on-chain Wordle Game (Telegram Mini App)

Цель

Разработать backend-приложение на Go, обслуживающее игру Wordle с криптоставками, реализованную как Telegram Mini App. Бэкенд отвечает за хранение данных, валидацию логики игры, обработку ставок, интеграцию с блокчейном и выдачу наград.

⸻

Технологии
	•	Язык: Go 1.24+
	•	Хранилище: PostgreSQL, Redis
	•	Blockchain: TON (через SDK / gRPC / REST)
	•	API: REST (JSON over HTTPS), Gin
	•	Очередь: Redis / BullMQ (для обработки событий транзакций)
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

## Запуск с Docker

Для запуска проекта с использованием Docker выполните следующие шаги:

1. Клонируйте репозиторий:
```bash
git clone https://github.com/TakuroBreath/wordle.git
cd wordle
```

2. Создайте файл `.env` с необходимыми переменными окружения или используйте значения по умолчанию:
```bash
cp .env.example .env  # Если есть пример файла окружения
```

3. Запустите приложение с помощью Docker Compose:
```bash
docker-compose up -d
```

4. Проверьте запущенные контейнеры:
```bash
docker-compose ps
```

5. Для просмотра логов:
```bash
docker-compose logs -f app
```

### Структура контейнеров

- **wordle-api** - основное приложение (Go)
- **postgres** - база данных PostgreSQL
- **redis** - кэш Redis
- **prometheus** - сбор метрик
- **grafana** - визуализация метрик и логов
- **loki** - хранилище логов
- **promtail** - сборщик логов

### Варианты запуска

- **Полная система (включая мониторинг и логирование)**:
  ```bash
  docker-compose -f docker-compose.full.yml up -d
  ```

- **Только бэкенд**:
  ```bash
  docker-compose up -d
  ```

- **Только мониторинг**:
  ```bash
  docker-compose -f docker-compose.monitoring.yml up -d
  ```

- **Только логирование**:
  ```bash
  docker-compose -f docker-compose.logging.yml up -d
  ```

### Доступные сервисы

- **API**: http://localhost:8080
- **Grafana**: http://localhost:3000 (admin/secret)
- **Prometheus**: http://localhost:9090
- **Loki**: http://localhost:3100

### Управление

- Остановка контейнеров: `docker-compose down`
- Перезапуск: `docker-compose restart [service]`
- Удаление данных: `docker-compose down -v` (будьте осторожны - это удалит все тома!)

## Миграции

Приложение использует [golang-migrate](https://github.com/golang-migrate/migrate) для управления миграциями базы данных.

### Ручной запуск миграций

```bash
# Запуск всех миграций
docker-compose exec app migrate -path=/app/migrations -database postgres://postgres:postgres@postgres:5432/wordle?sslmode=disable up

# Откат последней миграции
docker-compose exec app migrate -path=/app/migrations -database postgres://postgres:postgres@postgres:5432/wordle?sslmode=disable down 1
```

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

