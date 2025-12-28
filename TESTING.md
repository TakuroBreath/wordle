# Тестирование Wordle TON Game

## Обзор

Проект содержит unit-тесты и интеграционные тесты для проверки работоспособности всех компонентов.

## Структура тестов

```
internal/
├── api/handlers/
│   └── handlers_test.go      # Тесты HTTP handlers
├── mocks/
│   ├── repositories.go       # Моки для репозиториев
│   └── redis.go              # Мок для Redis
└── service/
    ├── auth_service_test.go        # Тесты аутентификации
    ├── game_service_test.go        # Тесты игровой логики
    ├── lobby_service_test.go       # Тесты лобби
    ├── transaction_service_test.go # Тесты транзакций
    └── user_service_test.go        # Тесты пользователей
```

## Запуск тестов

### Все тесты
```bash
make test
# или
go test ./...
```

### Только unit тесты
```bash
make test-unit
# или
go test -short ./...
```

### Тесты с покрытием
```bash
make test-coverage
# или
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

### Benchmark тесты
```bash
make test-bench
# или
go test -bench=. -benchmem ./...
```

### Тесты конкретного пакета
```bash
go test -v ./internal/service/...
go test -v ./internal/api/handlers/...
```

## Что тестируется

### Handler тесты (`handlers_test.go`)
- Health check endpoint
- JSON валидация
- Параметры пагинации
- UUID валидация
- CORS заголовки
- Ответы с ошибками

### Service тесты

#### `auth_service_test.go`
- Валидация срока действия токена
- Парсинг Telegram InitData
- Формат JWT токена
- Извлечение user ID из токена

#### `game_service_test.go`
- Проверка слова (CheckWord)
- Расчёт награды с учётом комиссии
- Валидация параметров игры
- Генерация короткого ID

#### `lobby_service_test.go`
- Проверка слова (CheckWord)
- Проверка корректности слова
- Расчёт награды

#### `transaction_service_test.go`
- Валидация суммы транзакции
- Валидация валюты
- Валидация типа транзакции
- Расчёт сетевой комиссии
- Валидация хеша транзакции
- Форматирование суммы

#### `user_service_test.go`
- Валидация баланса
- Расчёт доступного баланса
- Валидация адреса кошелька
- Валидация суммы вывода

## Моки

### MockUserRepository
Полная имплементация UserRepository для тестирования без БД.

### MockGameRepository
Полная имплементация GameRepository.

### MockLobbyRepository
Полная имплементация LobbyRepository.

### MockTransactionRepository
Полная имплементация TransactionRepository.

### MockAttemptRepository
Полная имплементация AttemptRepository.

### MockHistoryRepository
Полная имплементация HistoryRepository.

### MockRedisRepository
Полная имплементация RedisRepository.

## Примеры использования моков

```go
package mytest

import (
    "context"
    "testing"
    
    "github.com/TakuroBreath/wordle/internal/mocks"
    "github.com/TakuroBreath/wordle/internal/models"
)

func TestMyFunction(t *testing.T) {
    // Создаём моки
    userRepo := mocks.NewMockUserRepository()
    gameRepo := mocks.NewMockGameRepository()
    
    // Добавляем тестовые данные
    testUser := &models.User{
        TelegramID: 123456789,
        Username:   "testuser",
        BalanceTon: 100.0,
    }
    _ = userRepo.Create(context.Background(), testUser)
    
    // Используем в тестах
    user, err := userRepo.GetByTelegramID(context.Background(), 123456789)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if user.Username != "testuser" {
        t.Errorf("expected username 'testuser', got '%s'", user.Username)
    }
}
```

## CI/CD интеграция

Тесты автоматически запускаются при:
- Push в main/develop ветки
- Создании Pull Request
- Ручном запуске

### GitHub Actions пример

```yaml
name: Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.22'
      - run: go test -v ./...
```

## Рекомендации

1. **Запускайте тесты перед коммитом**
   ```bash
   make test
   ```

2. **Проверяйте покрытие при добавлении нового кода**
   ```bash
   make test-coverage
   ```

3. **Используйте benchmark для критичных функций**
   ```bash
   go test -bench=BenchmarkCheckWord -benchmem ./internal/service/...
   ```

4. **Пишите тесты для новых функций**
   - Покрывайте edge cases
   - Используйте table-driven tests
   - Добавляйте benchmark для критичных путей
