# Используем официальный образ Go
FROM golang:1.24-alpine AS builder

# Создаем рабочую директорию
WORKDIR /app

# Копируем файлы модулей
COPY go.mod go.sum ./

# Скачиваем зависимости
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -o /wordle ./cmd/wordle

# Используем минимальный образ для запуска
FROM alpine:3.18

# Копируем бинарный файл из builder
COPY --from=builder /wordle /wordle

# Указываем точку входа
ENTRYPOINT ["/wordle"]