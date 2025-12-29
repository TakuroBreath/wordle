# Инструкция по запуску в Docker

## Быстрый старт

### Локальный запуск со всеми сервисами

Для запуска проекта локально со всеми сервисами (backend, frontend, PostgreSQL, Prometheus, Grafana, Loki):

```bash
docker-compose -f docker-compose.full.yml up -d
```

**Доступные сервисы:**
- API: http://localhost:8080
- Frontend: http://localhost:3000
- Grafana: http://localhost:3000 (admin/secret)
- Prometheus: http://localhost:9090
- Loki: http://localhost:3100

### Запуск в продакшене

1. **Подготовьте конфигурацию:**
   ```bash
   cp configs/config.docker.prod.yaml configs/config.docker.local.yaml
   ```

2. **Заполните секреты в `configs/config.docker.local.yaml`:**
   - Пароль для PostgreSQL
   - JWT секрет: `openssl rand -hex 32`
   - Telegram Bot Token (получите у @BotFather)
   - TON API ключ и данные кошелька

3. **Обновите `docker-compose.yml`:**
   - Измените путь к конфигу на `config.docker.local.yaml`
   - Обновите пароль для PostgreSQL

4. **Запустите:**
   ```bash
   docker-compose up -d
   ```

## Варианты запуска

### Только backend + базы данных
```bash
docker-compose up -d
```

### Backend + Frontend (для разработки)
```bash
docker-compose -f docker-compose.dev.yml up -d
```

### Все сервисы включая мониторинг
```bash
docker-compose -f docker-compose.full.yml up -d
```

## Управление

**Просмотр логов:**
```bash
docker-compose logs -f app
```

**Остановка:**
```bash
docker-compose down
```

**Перезапуск:**
```bash
docker-compose restart app
```

**Удаление всех данных (⚠️ осторожно!):**
```bash
docker-compose down -v
```

## Структура конфигов

- `configs/config.local.yaml.example` - шаблон для локальной разработки (без Docker)
- `configs/config.docker.yaml` - конфиг для Docker dev (используется автоматически)
- `configs/config.docker.prod.yaml` - шаблон для продакшена

⚠️ **Важно:** Никогда не коммитьте файлы с реальными секретами (`config.docker.local.yaml`, `config.local.yaml`)!
