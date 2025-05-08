#!/bin/bash
set -e

# Цвета для вывода
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Starting Wordle Game server...${NC}"

# Ждем доступности PostgreSQL
echo -e "${YELLOW}Waiting for PostgreSQL...${NC}"
until PGPASSWORD=$POSTGRES_PASSWORD psql -h $POSTGRES_HOST -U $POSTGRES_USER -d $POSTGRES_DB -c '\q'; do
  echo -e "${YELLOW}PostgreSQL is unavailable - sleeping${NC}"
  sleep 1
done

echo -e "${GREEN}PostgreSQL is up - continuing${NC}"

# Запускаем миграции
echo -e "${YELLOW}Running database migrations...${NC}"
/usr/local/bin/migrate -path=/app/migrations -database postgres://$POSTGRES_USER:$POSTGRES_PASSWORD@$POSTGRES_HOST:$POSTGRES_PORT/$POSTGRES_DB?sslmode=disable up

if [ $? -ne 0 ]; then
  echo -e "${RED}Migration failed${NC}"
  exit 1
fi

echo -e "${GREEN}Migrations completed successfully${NC}"

# Запускаем приложение
echo -e "${GREEN}Starting application...${NC}"
exec /app/wordle 