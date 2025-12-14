# 🚀 БЫСТРАЯ ИНСТРУКЦИЯ: Очистка каталога через Render

## Шаг 1: Открыть базу данных
1. Зайти на https://dashboard.render.com
2. Выбрать PostgreSQL Database: **menu_fodi_backend**
3. Нажать **"Connect"** → **"External Connection"**
4. Или нажать кнопку **"Shell"** для прямого доступа

## Шаг 2: Выполнить SQL скрипты

### 📝 Скрипт 1: Удаление дубликатов (ОБЯЗАТЕЛЬНО)

```sql
-- КОПИРОВАТЬ И ВСТАВИТЬ В RENDER SHELL --

-- Удаление тестовых продуктов
DELETE FROM "Ingredient" 
WHERE "name" LIKE '%Тестов%' OR "name" LIKE '%тест%';

-- Удаление русских дубликатов лосося
DELETE FROM "Ingredient" 
WHERE "name" IN ('Лосось свежий', 'Лосось норвежский', 'Лосось Фермерский', 'Лосось фермерский', 'Лосось чилийский');

-- Удаление русских дубликатов креветок
DELETE FROM "Ingredient" 
WHERE "name" IN ('Креветки Королевские', 'Креветки тигровые');

-- Удаление русских дубликатов тунца
DELETE FROM "Ingredient" 
WHERE "name" IN ('Лещь', 'Тунец', 'Тунец Желтопёрый', 'Тунец желтоперый', 'Тунец свежий');

-- Удаление русских базовых продуктов
DELETE FROM "Ingredient" 
WHERE "name" IN ('Минеральная вода', 'Мука', 'Соль', 'Яица');

-- ПРОВЕРКА: Показать сколько осталось
SELECT COUNT(*) as total FROM "Ingredient";
```

### 📝 Скрипт 2: Установка значений по умолчанию (ОБЯЗАТЕЛЬНО)

```sql
-- КОПИРОВАТЬ И ВСТАВИТЬ В RENDER SHELL --

-- Установка defaultShelfLifeDays
UPDATE "Ingredient" SET "defaultShelfLifeDays" = 3 WHERE "defaultShelfLifeDays" IS NULL AND "category" = 'protein';
UPDATE "Ingredient" SET "defaultShelfLifeDays" = 7 WHERE "defaultShelfLifeDays" IS NULL AND "category" = 'vegetable';
UPDATE "Ingredient" SET "defaultShelfLifeDays" = 14 WHERE "defaultShelfLifeDays" IS NULL AND "category" = 'dairy';
UPDATE "Ingredient" SET "defaultShelfLifeDays" = 365 WHERE "defaultShelfLifeDays" IS NULL AND "category" = 'grain';
UPDATE "Ingredient" SET "defaultShelfLifeDays" = 365 WHERE "defaultShelfLifeDays" IS NULL AND "category" = 'condiment';
UPDATE "Ingredient" SET "defaultShelfLifeDays" = 180 WHERE "defaultShelfLifeDays" IS NULL AND "category" = 'other';

-- Установка defaultPricePerUnit
UPDATE "Ingredient" SET "defaultPricePerUnit" = 0.02 WHERE "defaultPricePerUnit" IS NULL AND "category" = 'protein';
UPDATE "Ingredient" SET "defaultPricePerUnit" = 0.01 WHERE "defaultPricePerUnit" IS NULL AND "category" IN ('vegetable', 'grain', 'dairy');
UPDATE "Ingredient" SET "defaultPricePerUnit" = 0.03 WHERE "defaultPricePerUnit" IS NULL AND "category" = 'condiment';
UPDATE "Ingredient" SET "defaultPricePerUnit" = 0.01 WHERE "defaultPricePerUnit" IS NULL AND "category" = 'other';

-- ПРОВЕРКА: Показать статистику
SELECT 
    "category",
    COUNT(*) as total,
    COUNT(CASE WHEN "defaultShelfLifeDays" IS NULL THEN 1 END) as missing_shelf_life,
    COUNT(CASE WHEN "defaultPricePerUnit" IS NULL THEN 1 END) as missing_price
FROM "Ingredient"
GROUP BY "category"
ORDER BY "category";
```

## Шаг 3: Проверить результат

### Через API:
```bash
curl -X GET "https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/catalog/ingredients" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### Через SQL:
```sql
SELECT COUNT(*) as total, 
       COUNT(CASE WHEN "defaultShelfLifeDays" IS NULL THEN 1 END) as missing_shelf_life,
       COUNT(CASE WHEN "defaultPricePerUnit" IS NULL THEN 1 END) as missing_price
FROM "Ingredient";
```

## ✅ Ожидаемый результат:
- ✅ Всего продуктов: **~180-190** (было 210)
- ✅ Нет русских названий (только PL)
- ✅ У всех продуктов есть `defaultShelfLifeDays`
- ✅ У всех продуктов есть `defaultPricePerUnit`
- ✅ `missing_shelf_life = 0`
- ✅ `missing_price = 0`

## 🎯 Что будет удалено:
- ~20 тестовых и русских дубликатов
- Все "Лосось" → остается только "Łosoś"
- Все "Тунец" → остается только "Tuńczyk"
- Все "Креветки" → остается только "Krewetki"

---

**⏱ Время выполнения:** ~2-3 минуты  
**⚠️ ВАЖНО:** Выполнять скрипты по порядку!
