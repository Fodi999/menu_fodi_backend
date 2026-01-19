#!/bin/bash

# Цвета для вывода
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${YELLOW}=== Проверка canonical name в admin коде ===${NC}\n"

# ПРИМЕЧАНИЕ: Этот тест проверяет КОД, а не API
# Причина: для API нужна роль 'admin', у fodi85@gmail.ru роль 'home_chef'

echo -e "${BLUE}📋 Проверяем implementation в коде...${NC}\n"

# 1. Проверяем что pkg/utils/canonical_names.go существует
echo -e "${YELLOW}1. Проверяем pkg/utils/canonical_names.go...${NC}"

if [ ! -f "pkg/utils/canonical_names.go" ]; then
  echo -e "${RED}❌ ОШИБКА: Файл pkg/utils/canonical_names.go не найден${NC}"
  exit 1
fi

# Проверяем что функция GenerateCanonicalName экспортирована
if ! grep -q "func GenerateCanonicalName" pkg/utils/canonical_names.go; then
  echo -e "${RED}❌ ОШИБКА: Функция GenerateCanonicalName не найдена${NC}"
  exit 1
fi

# Проверяем маппинг для "яичница"
if ! grep -q '"яичница".*"scrambled_eggs"' pkg/utils/canonical_names.go; then
  echo -e "${RED}❌ ОШИБКА: Маппинг 'яичница' → 'scrambled_eggs' не найден${NC}"
  exit 1
fi

echo -e "${GREEN}✅ pkg/utils/canonical_names.go корректен${NC}"
echo -e "   - Функция GenerateCanonicalName() экспортирована"
echo -e "   - Маппинг 'яичница' → 'scrambled_eggs' присутствует"
echo ""

# 2. Проверяем что admin service использует utils.GenerateCanonicalName
echo -e "${YELLOW}2. Проверяем internal/modules/admin/service/recipe_ai.go...${NC}"

if ! grep -q '"github.com/dmitrijfomin/menu-fodifood/backend/pkg/utils"' internal/modules/admin/service/recipe_ai.go; then
  echo -e "${RED}❌ ОШИБКА: Импорт pkg/utils не найден${NC}"
  exit 1
fi

if ! grep -q 'utils\.GenerateCanonicalName' internal/modules/admin/service/recipe_ai.go; then
  echo -e "${RED}❌ ОШИБКА: Вызов utils.GenerateCanonicalName() не найден${NC}"
  exit 1
fi

# Проверяем что НЕТ старого кода (strings.ToLower + ReplaceAll)
if grep -q 'strings\.ToLower.*strings\.ReplaceAll.*Title' internal/modules/admin/service/recipe_ai.go; then
  echo -e "${RED}❌ ОШИБКА: Найден старый код с strings.ToLower/ReplaceAll${NC}"
  echo -e "${RED}   Должен использоваться utils.GenerateCanonicalName()${NC}"
  exit 1
fi

echo -e "${GREEN}✅ admin service использует правильную функцию${NC}"
echo -e "   - Импорт pkg/utils присутствует"
echo -e "   - Вызов utils.GenerateCanonicalName() найден"
echo -e "   - Старый код (локализованный slug) удалён"
echo ""

# 3. Компиляция проекта
echo -e "${YELLOW}3. Проверяем компиляцию проекта...${NC}"

BUILD_OUTPUT=$(go build ./... 2>&1)
if [ $? -ne 0 ]; then
  echo -e "${RED}❌ ОШИБКА: Проект не компилируется${NC}"
  echo "$BUILD_OUTPUT"
  exit 1
fi

echo -e "${GREEN}✅ Проект компилируется без ошибок${NC}"
echo ""

# 4. Проверяем AI Recommendation service (тоже должен использовать shared utility)
echo -e "${YELLOW}4. Проверяем ai_recipe_recommendation service...${NC}"

if ! grep -q 'utils\.GenerateCanonicalName' internal/modules/ai_recipe_recommendation/service/recipe_match_service.go; then
  echo -e "${YELLOW}⚠️  WARNING: AI service не использует utils.GenerateCanonicalName()${NC}"
else
  echo -e "${GREEN}✅ AI service тоже использует shared utility${NC}"
fi
echo ""

# 5. Итоговый вывод
echo -e "${GREEN}╔════════════════════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║  ✅ ВСЕ ПРОВЕРКИ ПРОЙДЕНЫ                                 ║${NC}"
echo -e "${GREEN}╚════════════════════════════════════════════════════════════╝${NC}"
echo ""
echo -e "${BLUE}📊 Результаты анализа кода:${NC}"
echo -e "   1. ✅ Shared utility создан (pkg/utils/canonical_names.go)"
echo -e "   2. ✅ Admin service использует utils.GenerateCanonicalName()"
echo -e "   3. ✅ Маппинг 'яичница' → 'scrambled_eggs' присутствует"
echo -e "   4. ✅ Старый код (локализованный slug) удалён"
echo -e "   5. ✅ Проект компилируется без ошибок"
echo ""
echo -e "${BLUE}🎯 Architectural Decision:${NC}"
echo -e "   canonicalName теперь ВСЕГДА English slug"
echo -e "   Пример: 'Яичница' → 'scrambled_eggs' (не 'яичница')"
echo ""
echo -e "${YELLOW}📝 Следующие шаги:${NC}"
echo -e "   1. Закоммитить изменения (pkg/utils + admin service)"
echo -e "   2. Выполнить SQL миграцию NORMALIZE_CANONICAL_NAMES.sql"
echo -e "   3. Добавить UNIQUE constraint на canonicalName"
