.PHONY: run dev prod build test clean

# По умолчанию - dev режим
run: dev

# Dev режим (без авторизации, mock блокчейн)
dev:
	@ln -sf configs/config.dev.yaml config.yaml
	@echo "Running in DEV mode..."
	go run cmd/api/main.go

# Prod режим (с авторизацией, реальный блокчейн)
prod:
	@ln -sf configs/config.prod.yaml config.yaml
	@echo "Running in PROD mode..."
	go run cmd/api/main.go

# Сборка бинарника
build:
	go build -o wordle cmd/api/main.go

# Запуск тестов
test:
	go test ./...

# Очистка
clean:
	rm -f wordle config.yaml

# Помощь
help:
	@echo "Доступные команды:"
	@echo "  make dev   - запуск в dev режиме (по умолчанию)"
	@echo "  make prod  - запуск в prod режиме"
	@echo "  make build - сборка бинарника"
	@echo "  make test  - запуск тестов"
	@echo "  make clean - очистка"
