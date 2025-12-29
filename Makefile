.PHONY: run dev prod build test test-coverage clean docker-up docker-down docker-build help lint

# По умолчанию - dev режим
run: dev

# Dev режим (без авторизации, mock блокчейн)
dev:
	@if [ ! -f configs/config.local.yaml ]; then \
		echo "⚠️  configs/config.local.yaml не найден!"; \
		echo "Скопируйте шаблон: cp configs/config.local.yaml.example configs/config.local.yaml"; \
		exit 1; \
	fi
	@ln -sf configs/config.local.yaml config.yaml
	@echo "Running in DEV mode..."
	go run cmd/api/main.go

# Prod режим (с авторизацией, реальный блокчейн)
prod:
	@if [ ! -f configs/config.local.yaml ]; then \
		echo "⚠️  configs/config.local.yaml не найден!"; \
		echo "Скопируйте шаблон: cp configs/config.local.yaml.example configs/config.local.yaml"; \
		echo "И установите environment: prod в конфиге"; \
		exit 1; \
	fi
	@ln -sf configs/config.local.yaml config.yaml
	@echo "Running in PROD mode..."
	go run cmd/api/main.go

# Сборка бинарника
build:
	go build -ldflags="-s -w" -o wordle cmd/api/main.go

# Запуск тестов
test:
	@echo "Running tests..."
	go test -v ./...

# Запуск тестов с покрытием
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Запуск только unit тестов
test-unit:
	@echo "Running unit tests..."
	go test -v -short ./...

# Запуск benchmark тестов
test-bench:
	@echo "Running benchmark tests..."
	go test -bench=. -benchmem ./...

# Линтинг
lint:
	@echo "Running linters..."
	golangci-lint run ./...

# Docker - запуск всех сервисов
docker-up:
	@echo "Starting all services..."
	docker-compose up -d

# Docker - остановка всех сервисов
docker-down:
	@echo "Stopping all services..."
	docker-compose down

# Docker - сборка образов
docker-build:
	@echo "Building docker images..."
	docker-compose build

# Docker - dev режим с фронтендом
docker-dev:
	@echo "Starting development environment..."
	docker-compose -f docker-compose.dev.yml up --build

# Docker - dev режим в фоне
docker-dev-d:
	@echo "Starting development environment in background..."
	docker-compose -f docker-compose.dev.yml up -d --build

# Docker - остановка dev режима
docker-dev-down:
	@echo "Stopping development environment..."
	docker-compose -f docker-compose.dev.yml down

# Docker - просмотр логов
docker-logs:
	docker-compose logs -f

# Docker - логи бэкенда
docker-logs-backend:
	docker-compose logs -f app

# Миграции - применить
migrate-up:
	@echo "Applying migrations..."
	migrate -path ./migrations -database "postgres://postgres:postgres@localhost:5432/wordle?sslmode=disable" up

# Миграции - откатить
migrate-down:
	@echo "Rolling back migrations..."
	migrate -path ./migrations -database "postgres://postgres:postgres@localhost:5432/wordle?sslmode=disable" down 1

# Миграции - статус
migrate-status:
	@echo "Migration status..."
	migrate -path ./migrations -database "postgres://postgres:postgres@localhost:5432/wordle?sslmode=disable" version

# Очистка
clean:
	rm -f wordle config.yaml coverage.out coverage.html
	docker-compose down -v --remove-orphans 2>/dev/null || true

# Генерация моков (если используется mockgen)
generate-mocks:
	@echo "Generating mocks..."
	go generate ./...

# Помощь
help:
	@echo "Доступные команды:"
	@echo ""
	@echo "  Разработка:"
	@echo "    make dev           - запуск в dev режиме"
	@echo "    make prod          - запуск в prod режиме"
	@echo "    make build         - сборка бинарника"
	@echo ""
	@echo "  Тестирование:"
	@echo "    make test          - запуск всех тестов"
	@echo "    make test-unit     - запуск unit тестов"
	@echo "    make test-coverage - запуск тестов с отчётом о покрытии"
	@echo "    make test-bench    - запуск benchmark тестов"
	@echo "    make lint          - запуск линтеров"
	@echo ""
	@echo "  Docker:"
	@echo "    make docker-up     - запуск всех сервисов"
	@echo "    make docker-down   - остановка всех сервисов"
	@echo "    make docker-build  - сборка docker образов"
	@echo "    make docker-dev    - запуск dev окружения с фронтендом"
	@echo "    make docker-logs   - просмотр логов"
	@echo ""
	@echo "  Миграции:"
	@echo "    make migrate-up    - применить миграции"
	@echo "    make migrate-down  - откатить последнюю миграцию"
	@echo "    make migrate-status - статус миграций"
	@echo ""
	@echo "  Прочее:"
	@echo "    make clean         - очистка временных файлов"
	@echo "    make help          - эта справка"
