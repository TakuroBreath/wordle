# Стадия сборки
FROM golang:1.23-alpine AS builder

# Установка зависимостей для сборки
RUN apk add --no-cache gcc musl-dev git

# Создаем рабочую директорию
WORKDIR /app

# Копируем файлы модулей
COPY go.mod go.sum ./

# Скачиваем зависимости
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/wordle ./cmd/api

# Стадия миграции
FROM golang:1.23-alpine AS migrate

# Установка migrate
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Финальная стадия
FROM alpine:3.19

# Устанавливаем необходимые зависимости
RUN apk add --no-cache ca-certificates tzdata postgresql-client curl bash

# Создаем директорию для приложения
WORKDIR /app

# Копируем бинарный файл и миграции
COPY --from=builder /app/wordle /app/wordle
COPY --from=migrate /go/bin/migrate /usr/local/bin/migrate
COPY migrations /app/migrations
COPY scripts/entrypoint.sh /app/entrypoint.sh

# Проверяем содержимое директории migrations
RUN ls -la /app/migrations

# Делаем entrypoint.sh исполняемым
RUN chmod +x /app/entrypoint.sh

# Открываем порты
EXPOSE 8080 9090

# Указываем entrypoint
ENTRYPOINT ["/app/entrypoint.sh"]