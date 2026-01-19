# DRY RUN: Canonical Names Migration Analysis
## Дата: 2026-01-18, 15:45 UTC

## 📊 Текущее состояние БД (22 рецепта)

### Проблемы по категориям:

#### 1️⃣ ПУСТЫЕ canonical names (3 рецепта)
```
id: 107040a6-0614-4332-b94d-dc2d28c620e5 → canonicalName: ""
id: 4fc60295-... (из предыдущего анализа)
id: 56c18b33-... (из предыдущего анализа)
```
**Действие:** Установить значение на основе localName или titles

#### 2️⃣ ЛОКАЛИЗОВАННЫЕ (кириллица) canonical names (7 рецептов)
```
"жареный_лосось" → должно быть "fried_salmon"
"лосось_жареный" → должно быть "fried_salmon"
"жареный_лосось_(микроскопический_тест)" → должно быть "fried_salmon"
"жареный_лосось_(реалистичный_тест)" → должно быть "fried_salmon"
"жареный_лосось_с_хрустящей_кожей" → должно быть "fried_salmon"
"яичница" → должно быть "scrambled_eggs"
"паста_карбонара_(авторский_рецепт)" → должно быть "pasta_carbonara"
```
**Действие:** Заменить на English slugs (5 вариантов лосося → 1 canonical)

#### 3️⃣ НЕ SLUG формат (пробелы, заглавные) (7 рецептов)
```
"Greek Salad" → должно быть "greek_salad"
"Scrambled Eggs" → должно быть "scrambled_eggs"
"Polish Meat Dumplings" → должно быть "polish_meat_dumplings"
"Pierogi Ruskie" → должно быть "pierogi_ruskie"
"Spaghetti Carbonara" → должно быть "spaghetti_carbonara"
"Polish Chicken Soup" → должно быть "polish_chicken_soup"
"Polish Breaded Pork Chop" → должно быть "polish_breaded_pork_chop"
"Polish Hunters Stew" → должно быть "polish_hunters_stew"
"Polish Potato Pancakes" → должно быть "polish_potato_pancakes"
"Pizza Margherita" → должно быть "pizza_margherita"
```
**Действие:** Конвертировать в lowercase + underscores

#### 4️⃣ ПРАВИЛЬНЫЙ формат (5 рецептов)
```
Остальные рецепты уже имеют правильный slug формат
```

---

## 🎯 План миграции

### Фаза 1: UPDATE существующих рецептов

1. **Кириллица → English slugs**
   - 5× жареный лосось → `fried_salmon` (объединение)
   - 1× яичница → `scrambled_eggs`
   - 1× паста карбонара → `pasta_carbonara`

2. **Пробелы + заглавные → slugs**
   - 10 рецептов: lowercase + replace spaces with underscores

3. **Пустые canonical names**
   - 3 рецепта: генерация на основе titles

### Фаза 2: Деупликация

После UPDATE проверить:
```sql
SELECT canonicalName, COUNT(*)
FROM "Recipe"
GROUP BY canonicalName
HAVING COUNT(*) > 1;
```

**Ожидаемые дубликаты после UPDATE:**
- `scrambled_eggs`: 2 рецепта (merge required)
- `fried_salmon`: 5 рецептов (merge required)

**Стратегия merge:**
- Оставить 1 с лучшими данными
- Удалить остальные
- Обновить references (если есть)

### Фаза 3: Constraints

```sql
-- NOT NULL constraint
ALTER TABLE "Recipe"
ALTER COLUMN "canonicalName" SET NOT NULL;

-- UNIQUE constraint
CREATE UNIQUE INDEX recipes_canonical_name_unique
ON "Recipe" ("canonicalName");
```

---

## ⚠️ РИСКИ

1. **ПОТЕРЯ ДАННЫХ при merge:**
   - 5 рецептов лосося → 1 (потеря 4 рецептов)
   - 2 рецепта яичницы → 1 (потеря 1 рецепта)
   
   **Митигация:** Сохранить backup, проверить что удаляемые рецепты не используются

2. **BROKEN REFERENCES:**
   - Если есть связи Recipe → RecipeIngredient
   - Если есть связи Recipe → UserFavorites
   
   **Митигация:** Проверить foreign keys перед удалением

3. **API BREAKING CHANGES:**
   - Frontend может использовать старые canonicalName
   
   **Митигация:** Backend теперь генерирует canonicalName, frontend не должен зависеть

---

## ✅ СЛЕДУЮЩИЕ ШАГИ

1. ✅ **Commit pushed** (4f94695)
2. ⏳ **Backup БД** (ждём подтверждения доступа к Neon Console)
3. ⏳ **Выполнить DRY_RUN_CANONICAL_MIGRATION.sql** через Neon SQL Editor
4. ⏳ **Проверить результаты dry-run**
5. ⏳ **Выполнить NORMALIZE_CANONICAL_NAMES.sql**
6. ⏳ **Добавить constraints**
7. ⏳ **Тестирование API**

---

## 📝 ВОПРОСЫ К ПОЛЬЗОВАТЕЛЮ

1. **Есть ли доступ к Neon Console для создания backup branch?**
   - Да → создать branch "backup_before_canonical_migration"
   - Нет → нужен DATABASE_URL для pg_dump

2. **Готовы ли удалить дубликаты рецептов?**
   - 4 рецепта "жареный лосось" (оставить 1)
   - 1 рецепт "яичница" дубликат (оставить 1)

3. **Нужна ли проверка foreign keys перед удалением?**
   - RecipeIngredient ссылается на Recipe.id?
   - UserFavorites ссылается на Recipe.id?
