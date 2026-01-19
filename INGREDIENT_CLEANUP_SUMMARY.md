# Очистка базы ингредиентов - Краткая памятка

## ✅ Что сделано (19.01.2026)

### Проблема
- 79 ингредиентов с **маленькой буквы** (bacon, бекон, łosoś)
- 4 **дубликата** (ice cream x2, tomato x2, cucumber x2, test ingredient)

### Решение
```bash
# 1. Капитализация (79 updates)
bacon → Bacon
бекон → Бекон
łosoś → Łosoś

# 2. Удаление дубликатов (4 deletions)
- Ice cream (other/g) ❌
- Pomidor→pomidor ❌  
- Огурец обычный ❌
- Test ingredient ❌
```

### Результат
- **До:** 218 ингредиентов (79 lowercase)
- **После:** 214 ингредиентов (0 lowercase) ✅

---

## 📁 SQL Скрипты

```bash
export DATABASE_URL="postgresql://neondb_owner:...@.../neondb?sslmode=require"

# Анализ
psql "$DATABASE_URL" -f sql/analyze_lowercase_ingredients.sql

# Капитализация
psql "$DATABASE_URL" -f sql/capitalize_now.sql

# Просмотр дубликатов
psql "$DATABASE_URL" -f sql/cleanup_duplicate_ingredients.sql

# Удаление дубликатов
psql "$DATABASE_URL" -f sql/delete_duplicates_final.sql
```

---

## 🛡️ Защита на будущее

Backend автоматически капитализирует новые ингредиенты:

```go
// internal/modules/admin/service/service.go
namePL := utils.Capitalize(classification.NamePL)
nameEN := utils.Capitalize(classification.NameEN)
nameRU := utils.Capitalize(classification.NameRU)
```

**Статус:** ✅ Реализовано (commit `923e145`)

---

## 📊 Статистика

| Метрика | Значение |
|---------|----------|
| Total ingredients | 214 |
| Lowercase PL | 0 ✅ |
| Lowercase EN | 0 ✅ |
| Lowercase RU | 0 ✅ |
| Auto-translated | 21 |

---

## 📝 Документы

- `docs/INGREDIENT_CLEANUP_REPORT.md` - Полный отчет
- `docs/INGREDIENT_NAME_CAPITALIZATION.md` - Капитализация
- `docs/WHO_ADDS_INGREDIENT_TRANSLATIONS.md` - AI переводы
- `sql/*.sql` - SQL скрипты (5 файлов)

---

**Commit:** `afd7cd8`  
**Status:** ✅ Complete  
**Database:** Clean & consistent
