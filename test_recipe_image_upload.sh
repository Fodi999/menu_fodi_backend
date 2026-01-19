#!/bin/bash

# Цвета для вывода
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

BASE_URL="https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api"
EMAIL="admin@example.com"
PASSWORD="admin_password_123"

echo -e "${YELLOW}=== Тест загрузки изображения для рецепта ===${NC}\n"

# 1. Авторизация
echo -e "${YELLOW}1. Авторизация...${NC}"
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\"}")

TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.data.token // .data.accessToken')

if [ "$TOKEN" == "null" ] || [ -z "$TOKEN" ]; then
  echo -e "${RED}❌ Ошибка авторизации${NC}"
  echo "$LOGIN_RESPONSE" | jq '.'
  exit 1
fi

echo -e "${GREEN}✅ Токен получен: ${TOKEN:0:20}...${NC}\n"

# 2. Получаем список рецептов
echo -e "${YELLOW}2. Получаем список рецептов...${NC}"
RECIPES_RESPONSE=$(curl -s "$BASE_URL/recipes")
RECIPE_ID=$(echo "$RECIPES_RESPONSE" | jq -r '.data.recipes[0].id')

if [ "$RECIPE_ID" == "null" ] || [ -z "$RECIPE_ID" ]; then
  echo -e "${RED}❌ Нет рецептов в каталоге${NC}"
  exit 1
fi

echo -e "${GREEN}✅ Выбран рецепт: $RECIPE_ID${NC}"

# Показываем текущее состояние
RECIPE_NAME=$(echo "$RECIPES_RESPONSE" | jq -r '.data.recipes[0].canonicalName')
CURRENT_IMAGE=$(echo "$RECIPES_RESPONSE" | jq -r '.data.recipes[0].imageUrl // "null"')
echo -e "   Название: ${BLUE}$RECIPE_NAME${NC}"
echo -e "   Текущее изображение: ${BLUE}$CURRENT_IMAGE${NC}\n"

# 3. Создаём тестовое изображение (1x1 PNG)
echo -e "${YELLOW}3. Создаём тестовое изображение...${NC}"
TEST_IMAGE="/tmp/test_recipe_image.png"

# Создаём минимальный PNG (1x1 красный пиксель)
# PNG header + IHDR + IDAT + IEND
printf '\x89\x50\x4e\x47\x0d\x0a\x1a\x0a\x00\x00\x00\x0d\x49\x48\x44\x52\x00\x00\x00\x01\x00\x00\x00\x01\x08\x02\x00\x00\x00\x90\x77\x53\xde\x00\x00\x00\x0c\x49\x44\x41\x54\x08\xd7\x63\xf8\xcf\xc0\x00\x00\x03\x01\x01\x00\x18\xdd\x8d\xb4\x00\x00\x00\x00\x49\x45\x4e\x44\xae\x42\x60\x82' > "$TEST_IMAGE"

if [ ! -f "$TEST_IMAGE" ]; then
  echo -e "${RED}❌ Не удалось создать тестовое изображение${NC}"
  exit 1
fi

FILE_SIZE=$(wc -c < "$TEST_IMAGE")
echo -e "${GREEN}✅ Изображение создано: $FILE_SIZE bytes${NC}\n"

# 4. Загружаем изображение
echo -e "${YELLOW}4. Загружаем изображение...${NC}"

UPLOAD_RESPONSE=$(curl -s -X POST "$BASE_URL/admin/recipes/$RECIPE_ID/image" \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@$TEST_IMAGE")

echo -e "${BLUE}Response:${NC}"
echo "$UPLOAD_RESPONSE" | jq '.'

# 5. Проверяем результат
SUCCESS=$(echo "$UPLOAD_RESPONSE" | jq -r '.success')

if [ "$SUCCESS" == "true" ]; then
  echo -e "\n${GREEN}✅ УСПЕХ: Изображение загружено!${NC}"
  
  IMAGE_URL=$(echo "$UPLOAD_RESPONSE" | jq -r '.data.imageUrl')
  PUBLIC_ID=$(echo "$UPLOAD_RESPONSE" | jq -r '.data.publicId')
  WIDTH=$(echo "$UPLOAD_RESPONSE" | jq -r '.data.width')
  HEIGHT=$(echo "$UPLOAD_RESPONSE" | jq -r '.data.height')
  FORMAT=$(echo "$UPLOAD_RESPONSE" | jq -r '.data.format')
  SIZE=$(echo "$UPLOAD_RESPONSE" | jq -r '.data.size')
  
  echo -e "\n${BLUE}📊 Результат загрузки:${NC}"
  echo -e "   URL: ${GREEN}$IMAGE_URL${NC}"
  echo -e "   Public ID: ${GREEN}$PUBLIC_ID${NC}"
  echo -e "   Размер: ${GREEN}${WIDTH}x${HEIGHT}${NC}"
  echo -e "   Формат: ${GREEN}$FORMAT${NC}"
  echo -e "   Размер файла: ${GREEN}$SIZE bytes${NC}"
  
  # Thumbnails
  echo -e "\n${BLUE}🖼️  Thumbnails:${NC}"
  echo "$UPLOAD_RESPONSE" | jq -r '.data.thumbnails | to_entries[] | "   \(.key): \(.value)"'
  
  # 6. Проверяем что изображение доступно
  echo -e "\n${YELLOW}6. Проверяем доступность изображения...${NC}"
  HTTP_STATUS=$(curl -s -o /dev/null -w "%{http_code}" "$IMAGE_URL")
  
  if [ "$HTTP_STATUS" == "200" ]; then
    echo -e "${GREEN}✅ Изображение доступно (HTTP $HTTP_STATUS)${NC}"
  else
    echo -e "${RED}⚠️  Изображение недоступно (HTTP $HTTP_STATUS)${NC}"
  fi
  
  # 7. Проверяем обновление в БД
  echo -e "\n${YELLOW}7. Проверяем обновление рецепта в БД...${NC}"
  UPDATED_RECIPE=$(curl -s "$BASE_URL/recipes" | jq ".data.recipes[] | select(.id == \"$RECIPE_ID\")")
  UPDATED_IMAGE=$(echo "$UPDATED_RECIPE" | jq -r '.imageUrl')
  
  if [ "$UPDATED_IMAGE" == "$IMAGE_URL" ]; then
    echo -e "${GREEN}✅ База данных обновлена корректно${NC}"
  else
    echo -e "${RED}❌ ОШИБКА: imageUrl в БД не совпадает${NC}"
    echo -e "   Expected: $IMAGE_URL"
    echo -e "   Got: $UPDATED_IMAGE"
  fi
  
else
  echo -e "\n${RED}❌ ОШИБКА: Загрузка не удалась${NC}"
  ERROR=$(echo "$UPLOAD_RESPONSE" | jq -r '.error // .message // "Unknown error"')
  echo -e "   Причина: $ERROR"
fi

# Cleanup
rm -f "$TEST_IMAGE"

echo -e "\n${GREEN}=== Тест завершён ===${NC}"
