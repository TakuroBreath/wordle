Техническое задание: Бэкенд on-chain Wordle Game (Telegram Mini App)

Цель

Разработать backend-приложение на Go, обслуживающее игру Wordle с криптоставками, реализованную как Telegram Mini App. Бэкенд отвечает за хранение данных, валидацию логики игры, обработку ставок, интеграцию с блокчейном и выдачу наград.

⸻

Технологии
	•	Язык: Go 1.24+
	•	Хранилище: PostgreSQL
	•	Blockchain: TON (через SDK / gRPC / REST)
	•	API: REST (JSON over HTTPS)
	•	Очередь: Redis / BullMQ (для обработки событий транзакций)
	•	Аутентификация: через Telegram Mini App авторизацию

⸻

Основные сущности

1. User (пользователь)
	•	telegram_id: string (уникальный)
	•	username: string
	•	first_name: string
	•	last_name: string
	•	balance: decimal (в TON/USDT)
	•	wins: int
	•	losses: int
	•	created_at: datetime
	•	updated_at: datetime

2. Game (игра)
	•	id: UUID
	•	creator_id: telegram_id (User)
	•	word: string (хэшируется/шифруется)
	•	title: string
	•	description: text
	•	min_bet: decimal
	•	max_bet: decimal (рассчитывается)
	•	reward_multiplier: float64 (1.5 - 10)
	•	expires_at: datetime
	•	status: enum: [active, inactive]
	•	max_reward: decimal (max_bet * multiplier)
	•	created_at: datetime
	•	updated_at: datetime

3. Lobby (игровая сессия)
	•	id: UUID
	•	game_id: UUID
	•	player_id: telegram_id
	•	attempts_max: int
	•	attempts_used: int
	•	attempts: []Attempt
	•	potential_reward: decimal
	•	status: enum: [active, success, failed]
	•	created_at: datetime
	•	updated_at: datetime
	•	expires_at: datetime

4. Attempt (попытка)
	•	id: UUID
	•	game_id: UUID
	•	lobby_id: UUID
	•	word_guessed: string
	•	result: []int (0/1/2 как в Wordle)
	•	created_at: datetime

⸻

Бизнес-логика

Создание пользователя
	•	При первом входе через Telegram Mini App пользователь создаётся на основе telegram_id.
	•	Баланс = 0, побед и поражений нет.

Создание игры
	•	Игрок задаёт слово, макс. количество попыток, множитель награды, минимальную ставку.
	•	Высчитывается max_bet = min_bet * 2, max_reward = max_bet * multiplier.
	•	Создатель переводит депозит = max_reward.
	•	После успешной транзакции игра становится активной.

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
	•	Если слово угадано — статус лобби = success, начисляется выигрыш.
	•	Если попытки исчерпаны — статус = failed.

Выплата наград
	•	Победа: сумма начисляется игроку, вычитается у создателя.
	•	Проигрыш: остаётся у создателя.
	•	Комиссия: % остаётся в сервисе (константа, например 5%).

⸻

API (примерные эндпоинты)

Auth
	•	POST /auth/telegram — регистрация по данным Telegram Mini App

User
	•	GET /me
	•	GET /user/{id}

Game
	•	POST /game — создание игры
	•	GET /game/{id}
	•	GET /games/active

Lobby
	•	POST /lobby/join — присоединение к игре
	•	GET /lobby/{id}
	•	POST /lobby/{id}/attempt — новая попытка

Admin / Service
	•	GET /stats
	•	GET /rewards/pending
	•	POST /rewards/withdraw

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
	•	Механика “подсказок за комиссию”

