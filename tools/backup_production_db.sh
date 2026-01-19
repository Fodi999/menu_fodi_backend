#!/bin/bash

# Цвета
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${YELLOW}=== Backup Production Database (Neon PostgreSQL) ===${NC}\n"

# Переменные подключения (из Koyeb env)
DB_HOST="ep-soft-mud-agon8wu3.c-2.eu-central-1.aws.neon.tech"
DB_NAME="recipe_matching_db"
DB_USER="fodi999"

# ПРИМЕЧАНИЕ: Требуется пароль из Koyeb environment variables
# Найти можно в: Koyeb Dashboard → Service → Settings → Environment

echo -e "${BLUE}📋 Информация о базе данных:${NC}"
echo -e "   Host: $DB_HOST"
echo -e "   Database: $DB_NAME"
echo -e "   User: $DB_USER"
echo ""

echo -e "${YELLOW}⚠️  ВНИМАНИЕ: Требуется DATABASE_URL из Koyeb environment${NC}"
echo -e "${YELLOW}   Формат: postgresql://user:password@host/database${NC}"
echo ""

# Проверка наличия DATABASE_URL
if [ -z "$DATABASE_URL" ]; then
  echo -e "${RED}❌ ERROR: DATABASE_URL не установлен${NC}"
  echo -e "${YELLOW}Установите переменную окружения:${NC}"
  echo -e "   export DATABASE_URL='postgresql://...'"
  echo ""
  echo -e "${YELLOW}Или используйте Neon Console для создания backup:${NC}"
  echo -e "   1. Откройте https://console.neon.tech"
  echo -e "   2. Выберите проект 'ep-soft-mud-agon8wu3'"
  echo -e "   3. Branches → Create Branch (snapshot создастся автоматически)"
  exit 1
fi

# Создаём backup
BACKUP_FILE="backup_before_canonical_migration_$(date +%Y%m%d_%H%M%S).sql"

echo -e "${BLUE}📦 Создаём backup...${NC}"
pg_dump "$DATABASE_URL" > "$BACKUP_FILE"

if [ $? -eq 0 ]; then
  BACKUP_SIZE=$(du -h "$BACKUP_FILE" | cut -f1)
  echo -e "${GREEN}✅ Backup создан успешно!${NC}"
  echo -e "   Файл: $BACKUP_FILE"
  echo -e "   Размер: $BACKUP_SIZE"
  echo ""
  
  # Проверяем что в backup есть таблица Recipe
  if grep -q "CREATE TABLE.*Recipe" "$BACKUP_FILE"; then
    echo -e "${GREEN}✅ Таблица Recipe найдена в backup${NC}"
  else
    echo -e "${RED}⚠️  WARNING: Таблица Recipe не найдена в backup${NC}"
  fi
  
  echo ""
  echo -e "${GREEN}Backup готов. Теперь можно выполнять миграцию.${NC}"
else
  echo -e "${RED}❌ ОШИБКА: Backup не создан${NC}"
  exit 1
fi
