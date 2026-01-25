# BaseApp Mini App (React) — ТЗ для клиентского агента

Документ описывает, как собрать клиентскую часть (React) для мини‑приложения в Base App. Цель — передать этот файл другому агенту, который создаст полноценный фронтенд, интегрированный с нашим бэкендом.

Источники по Base Mini Apps: https://docs.base.org/mini-apps/llms-full.txt (см. разделы MiniKit, OnchainKit, Authentication, Manifest).

---

## 0) Ключевые принципы

1. **Гостевой режим** — показывать список игр и карточки без авторизации.  
2. **Прогрессивная авторизация** — просить auth только при действиях: создание игры, вступление, депозит, вывод.  
3. **Не доверять контексту клиента** — любые подписи/токены проверяются на сервере.  
4. **Base Mini App** — использовать MiniKit + OnchainKit, учитывать UI‑паттерны Base App (короткие флоу, минимум кликов).

---

## 1) Сетевые требования и интеграция Base Mini App

### 1.1 Библиотеки
- `@coinbase/minikit`
- `@coinbase/onchainkit`
- `wagmi` + `viem` (если нужно для подписей/балансов)

### 1.2 Провайдеры (минимально)
```tsx
import { OnchainKitProvider } from '@coinbase/onchainkit'
import { base } from 'wagmi/chains'

export function Providers(props: { children: React.ReactNode }) {
  return (
    <OnchainKitProvider apiKey={import.meta.env.VITE_ONCHAINKIT_API_KEY} chain={base}>
      {props.children}
    </OnchainKitProvider>
  )
}
```

### 1.3 Контекст MiniKit
```tsx
import { useMiniKit } from '@coinbase/minikit'
const { context } = useMiniKit()
// context.user — только для подсказок UI, НЕ для auth
```

### 1.4 Манифест Base App
Подготовить `/.well-known/farcaster.json` с данными приложения (пример в Base docs). Это нужно для discovery и корректного запуска в Base App.

---

## 2) Аутентификация (Base)

### 2.1 Что нужно на клиенте
Использовать `useAuthenticate` (OnchainKit) для получения подписанного payload.  
В результате получить:
- `address` (адрес, которым подписан message)
- `message` (SIWF / Wallet message)
- `signature`
- `fid` (Farcaster ID)
- `username`, `display_name`, `pfp_url`
- `custody_address`, `smart_wallet_address` (если доступны)

### 2.2 Серверный контракт
**POST** `/api/v1/auth/base`

**Body:**
```json
{
  "address": "0x...",
  "message": "string",
  "signature": "0x...",
  "fid": 12345,
  "username": "alice",
  "display_name": "Alice",
  "pfp_url": "https://...",
  "custody_address": "0x...",
  "smart_wallet_address": "0x...",
  "auth_method": "siwf"
}
```

**Response:**
```json
{ "token": "JWT" }
```

**Требование:** `fid` обязателен. В БД `telegram_id` используется как универсальный user_id и равен `fid` для Base.

### 2.3 Использование токена
Передавать в заголовке:
```
Authorization: Bearer <JWT>
```

### 2.4 Проверка токена
**POST** `/api/v1/auth/verify`
```json
{ "token": "JWT" }
```

Ответ содержит профиль пользователя и Base‑метаданные.

---

## 3) Пользователь и профиль

### 3.1 Получить текущего пользователя
**GET** `/api/v1/users/me`

**Response (пример):**
```json
{
  "telegram_id": 12345,
  "auth_provider": "base",
  "username": "alice",
  "first_name": "Alice",
  "last_name": "",
  "wallet": "0x...",
  "balance_ton": 0,
  "balance_usdt": 0,
  "balance_usdc": 12.5,
  "pending_withdrawal": 0,
  "wins": 2,
  "losses": 1,
  "total_deposited": 50,
  "total_withdrawn": 10,
  "can_withdraw": true,
  "base": {
    "fid": 12345,
    "username": "alice",
    "display_name": "Alice",
    "pfp_url": "https://...",
    "custody_address": "0x...",
    "smart_wallet_address": "0x..."
  }
}
```

### 3.2 Привязать кошелек (если нужно)
**POST** `/api/v1/users/wallet`
```json
{ "wallet": "0x..." }
```

---

## 4) Баланс, депозит, вывод

### 4.1 Баланс
**GET** `/api/v1/users/balance`

### 4.2 Депозит (Base / USDC)
1) **GET** `/api/v1/users/deposit?currency=USDC`  
Ответ содержит адрес получателя, сеть и инструкцию.

2) Пользователь отправляет USDC на Base.

3) **POST** `/api/v1/transactions/deposit`
```json
{
  "amount": 10.5,
  "currency": "USDC",
  "tx_hash": "0x...",
  "network": "base"
}
```

Бэкенд валидирует tx_hash и проверяет, что `from` совпадает с `wallet` пользователя.

4) Проверка статуса:  
**POST** `/api/v1/transactions/verify`
```json
{ "tx_hash": "0x..." }
```

### 4.3 Вывод средств
**POST** `/api/v1/users/withdraw`
```json
{
  "amount": 5,
  "currency": "USDC",
  "to_address": "0x..." // опционально
}
```

Ответ:
```json
{
  "transaction_id": "uuid",
  "status": "pending",
  "amount": 5,
  "fee": 0.5,
  "net_amount": 4.5,
  "currency": "USDC"
}
```

История: **GET** `/api/v1/users/withdrawals`

---

## 5) Игры и лобби

### 5.1 Список игр
**GET** `/api/v1/games?limit=10&offset=0`

### 5.2 Поиск
**GET** `/api/v1/games/search?min_bet=&max_bet=&difficulty=&currency=USDC`

### 5.3 Создание игры (Base)
**POST** `/api/v1/games`
```json
{
  "word": "apple",
  "difficulty": "easy",
  "max_tries": 6,
  "time_limit": 5,
  "title": "Easy Apple",
  "description": "Try to guess",
  "min_bet": 1,
  "max_bet": 5,
  "reward_multiplier": 2,
  "currency": "USDC"
}
```

Для Base депозит резервируется **с баланса** (USDC), без on-chain платежа.

### 5.4 Вступление в игру
**POST** `/api/v1/games/join`
```json
{
  "game_id": "<uuid|short_id>",
  "bet_amount": 2,
  "use_balance": true
}
```
На Base `use_balance` автоматически = true.

### 5.5 Лобби
**POST** `/api/v1/lobbies` — вступить (баланс)  
**GET** `/api/v1/lobbies` — список лобби пользователя  
**GET** `/api/v1/lobbies/active?game_id=...` — активное лобби  
**GET** `/api/v1/lobbies/:id` — состояние  
**POST** `/api/v1/lobbies/:id/attempt` — попытка  
**GET** `/api/v1/lobbies/:id/attempts` — история  
**POST** `/api/v1/lobbies/:id/extend` — продление времени  
**POST** `/api/v1/lobbies/:id/cancel` — отмена

---

## 6) Транзакции

### 6.1 История транзакций
**GET** `/api/v1/transactions?limit=20&offset=0&type=all`

### 6.2 Статистика
**GET** `/api/v1/transactions/stats`

### 6.3 Транзакция по id
**GET** `/api/v1/transactions/:id`

---

## 7) UX‑сценарии (пайплайны)

### 7.1 Онбординг
1. Экран Welcome + список активных игр (без auth).
2. CTA: «Начать игру» / «Создать игру».
3. При попытке действия — показать auth (SIWF).

### 7.2 Создание игры (Base)
1. Проверка баланса USDC (GET /users/balance).
2. Форма создания.
3. POST /games — автоматическое резервирование депозита.
4. Показать подтверждение.

### 7.3 Вступление в игру
1. Проверка доступного баланса.
2. POST /games/join (use_balance=true).
3. Переход к экрану лобби.

### 7.4 Игра
1. Получить активное лобби.
2. Сабмит попыток.
3. После завершения — показать награду и обновить баланс.

### 7.5 Депозит
1. Получить адрес для пополнения.
2. Открыть transaction tray для перевода USDC.
3. Отправить tx_hash на сервер.
4. Показать статус (pending → completed).

### 7.6 Вывод
1. Проверить can_withdraw.
2. POST /users/withdraw.
3. Показать pending и обновлять статус.

---

## 8) Ошибки и статус‑коды (общие)

| HTTP | Причина |
|------|--------|
| 400  | Ошибка валидации / недостаточно баланса |
| 401  | Нет/невалидный токен |
| 403  | Нет доступа |
| 404  | Объект не найден |
| 409  | Конфликт (например, активное лобби) |
| 500  | Ошибка сервера |

---

## 9) Важные заметки

1. `telegram_id` в ответах = `fid` для Base.  
2. `wallet` — основной кошелек пользователя для депозитов/выводов.  
3. Для Base активна валюта **USDC**, не USDT/TON.  
4. `/games/:id/payment` не используется на Base (резервируем с баланса).  
5. Минимизировать лишние шаги: показывать auth только при необходимости.  

---

## 10) Минимальный список экранов

- Home (лента игр)
- Game Details
- Create Game
- Lobby (игровая сессия)
- Profile (баланс, история)
- Deposit / Withdraw

---

## 11) Что должен сделать клиент‑агент

1. Реализовать UI на React + MiniKit + OnchainKit.
2. Настроить Base auth: собрать payload и вызвать `/auth/base`.
3. Обеспечить работу всех эндпоинтов и сценариев выше.
4. Покрыть ошибки/статусы и показать понятные уведомления.
