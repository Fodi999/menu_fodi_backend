# ✅ Canonical Ingredients System - DONE

## Что сделано

### 1. База данных (SQL Migration)
- `CanonicalIngredient` - каноническая таблица продуктов
- `IngredientAlias` - таблица алиасов (все языки, синонимы, опечатки)
- **Unique indexes** защищают от дублей на DB level
- Триггеры для `updatedAt`

### 2. Модели Go
- `models.CanonicalIngredient` - канонический продукт
- `models.IngredientAlias` - алиас
- DTOs для API
- Методы `GetNameForLanguage()`, `ToDTO()`

### 3. Нормализация
- `utils.NormalizeName()` - для поиска дублей
- `utils.GenerateCanonicalKey()` - уникальный ключ
- Удаление диакритики (ą→a, ó→o)
- Levenshtein distance для similarity

### 4. CRUD Сервис
- `CreateOrFindIngredient()` - **главная функция** (всегда проверяет дубли)
- `SearchByQuery()` - автокомплит через aliases
- `MergeIngredients()` - объединение дублей
- `ArchiveIngredient()` - архивация (не удаление!)

### 5. Документация
- `CANONICAL_INGREDIENTS_GUIDE.md` - полное руководство по внедрению

---

## Как это работает

### Пример: "Лук"

**Было (старая система):**
```
Ingredient { id: 1, name: "Лук" }
Ingredient { id: 2, name: "лук" }
Ingredient { id: 3, name: "Onion" }
Ingredient { id: 4, name: "Cebula" }
→ 4 дубликата!
```

**Стало (каноническая система):**
```
CanonicalIngredient { 
  id: uuid-123,
  canonicalKey: "onion",
  canonicalName: "Onion"
}

IngredientAlias { name: "Лук", normalizedName: "лук", language: "ru" }
IngredientAlias { name: "лук", normalizedName: "лук", language: "ru" }
IngredientAlias { name: "Onion", normalizedName: "onion", language: "en" }
IngredientAlias { name: "Cebula", normalizedName: "cebula", language: "pl" }
→ 1 продукт + 4 алиаса!
```

---

## Защита от дублей

### 1. Уникальный индекс на DB
```sql
CREATE UNIQUE INDEX ON "IngredientAlias"("normalizedName");
```

**Результат:** Попытка создать "лук" дважды → DB ERROR

### 2. CreateOrFind pattern
```go
// AI НЕ создаёт напрямую
ingredient, created, err := service.CreateOrFindIngredient(input)

// created = false если найден существующий
// created = true если создан новый
```

**Результат:** AI физически не может создать дубликат

---

## Что дальше

### Следующие шаги:

1. **Применить миграцию** в production DB
   ```bash
   psql $DATABASE_URL < migrations/20260115_add_canonical_ingredients.sql
   ```

2. **Мигрировать существующие данные** (скрипт в гайде)
   - Старые ингредиенты → канонические
   - Создать алиасы для всех языков
   - Найти и объединить дубликаты

3. **Обновить API endpoints**
   - `POST /api/admin/ingredients/v2` - новый эндпоинт
   - `GET /api/admin/ingredients/v2/search` - поиск через aliases
   - `POST /api/admin/ingredients/v2/merge` - объединение дублей

4. **Интегрировать с AI**
   - AI → `CreateOrFindIngredient()` вместо прямого создания
   - Автоматическая проверка дублей

5. **Обновить фронтенд**
   - Использовать новый API
   - Показывать все алиасы
   - UI для merge дублей

---

## Преимущества

✅ **Физически невозможно** создать дубликат (unique index)  
✅ **AI не создаёт мусор** (через CreateOrFind)  
✅ **Все языки** - одна запись  
✅ **Merge вместо delete** - история сохраняется  
✅ **Быстрый поиск** - индексы на normalizedName  
✅ **Автоматическая нормализация** - без человека  

---

## Мониторинг

```sql
-- Сколько продуктов
SELECT COUNT(*) FROM "CanonicalIngredient" WHERE status='active';

-- Сколько алиасов
SELECT COUNT(*) FROM "IngredientAlias";

-- Топ по алиасам
SELECT ci."canonicalName", COUNT(ia.id) as aliases
FROM "CanonicalIngredient" ci
JOIN "IngredientAlias" ia ON ia."canonicalIngredientId" = ci.id
GROUP BY ci.id
ORDER BY aliases DESC
LIMIT 10;

-- Проверка дублей (должно быть 0!)
SELECT "normalizedName", COUNT(*) 
FROM "IngredientAlias" 
GROUP BY "normalizedName" 
HAVING COUNT(*) > 1;
```

---

## Файлы

```
migrations/
  20260115_add_canonical_ingredients.sql     ← SQL миграция

internal/models/
  canonical_ingredient.go                    ← Модели

internal/modules/admin/service/
  canonical_ingredient_service.go            ← CRUD сервис

pkg/utils/
  normalize.go                               ← Нормализация

CANONICAL_INGREDIENTS_GUIDE.md               ← Полный гайд
CANONICAL_INGREDIENTS_SUMMARY.md             ← Это резюме
```

---

**Статус:** ✅ Готово к деплою  
**Коммит:** `acb6540`  
**Pushed:** ✅ GitHub  
**Koyeb:** ⏳ Ожидает деплоя с первым исправлением

---

## FAQ

**Q: Что будет со старыми ингредиентами?**  
A: Они остаются! Миграция не удаляет ничего. Работают параллельно.

**Q: Когда удалять старую таблицу?**  
A: После полной миграции и проверки (недели через 2-3).

**Q: AI создаст дубликат если я забуду?**  
A: Нет! Unique index на DB не даст.

**Q: Можно ли откатить?**  
A: Да, просто удалить новые таблицы. Старые не трогались.

**Q: Что если нашли дубликат?**  
A: `POST /api/admin/ingredients/v2/merge` - объединит в один.

---

**🎯 Цель достигнута: дубликаты физически невозможны!**
