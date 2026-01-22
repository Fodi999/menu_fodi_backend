# 🔧 MENU UNIQUE CONSTRAINT FIX

## 🐛 ПРОБЛЕМА

При попытке выполнить Шаг 5 (завершить готовку), получаю ошибку:

```
ERROR: duplicate key value violates unique constraint "unique_user_recipe_today"
```

### Корневая причина

**Неправильный UNIQUE constraint в миграции:**

```sql
UNIQUE (user_id, recipe_id, planned_for, status)
```

Этот constraint позволял ОДНУ комбинацию `(user_id, recipe_id, planned_for, status)`.

**Проблема в workflow:**
1. ✅ `/api/menu/{id}/start` → status: planned → cooking (UPDATE)
2. ❌ `/api/menu/{id}/complete` → status: cooking → completed (UPDATE → INSERT?)

Когда status меняется, система пытается создать новую запись, но это нарушает UNIQUE!

## ✅ РЕШЕНИЕ

### 1. Исправить constraint (только на recipe per day, независимо от status)

```sql
UNIQUE (user_id, recipe_id, planned_for)  -- ✅ Правильно!
```

Это позволяет:
- ✅ Одного юзера + один рецепт + один день = уникально
- ✅ Менять status как угодно (planned → cooking → completed)
- ✅ Предотвратить добавление одного рецепта дважды в один день

### 2. Обновить GORM AutoMigrate

Добавить `&models.UserMenuItem{}` в список моделей в `db.go`:

```go
// Kitchen Pipeline (user menu)
&models.UserMenuItem{},
```

## 📁 Файлы изменены

1. **migrations/20260122_recreate_user_menu_items_fixed.sql** - Новая миграция с правильным constraint
2. **internal/database/db.go** - Добавлена модель в AutoMigrate

## 🚀 Как применить

### Вариант 1: Через GORM AutoMigrate (автоматический)
```bash
cd backend
make build
make migrate
```

GORM автоматически пересоздаст таблицу с правильной структурой.

### Вариант 2: Ручная SQL миграция (если нужен контроль)
```bash
psql $DATABASE_URL < migrations/20260122_recreate_user_menu_items_fixed.sql
```

## ✅ Проверка после исправления

После применения миграции тестируем весь цикл:

```bash
# STEP 1: Добавить рецепт в меню
curl -X POST "https://api.example.com/api/menu" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{"recipe_id":"0f153c77-d554-44c5-94b1-9b1171b854d0","servings":2}'

# STEP 2: Начать готовку
curl -X POST "https://api.example.com/api/menu/{ITEM_ID}/start" \
  -H "Authorization: Bearer YOUR_TOKEN"

# STEP 3: Завершить готовку
curl -X POST "https://api.example.com/api/menu/{ITEM_ID}/complete" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{"actual_servings":2}'
```

### Ожидаемый результат

```json
✅ STEP 3: status → completed
   completed_at: 2026-01-22T20:15:30Z ✅
```

## 📊 Сравнение constraints

| Constraint | Позволяет дублировать? | Позволяет менять status? |
|---|---|---|
| **ДО** (с status) | Да, разные статусы | ❌ НЕТ |
| **ПОСЛЕ** (без status) | ❌ НЕТ | ✅ ДА |

## 🔍 Debugging

Если проблемы все еще есть:

```bash
# Проверить constraint
psql $DATABASE_URL -c "\d user_menu_items"

# Проверить данные
psql $DATABASE_URL -c "SELECT * FROM user_menu_items WHERE user_id = '407582be-59d5-4d21-873b-1a72d31b0d42';"

# Очистить тестовые данные
psql $DATABASE_URL -c "DELETE FROM user_menu_items WHERE user_id = '407582be-59d5-4d21-873b-1a72d31b0d42';"
```

## 📝 Примечания

- ✅ Таблица остается безопасной: один рецепт один раз в день
- ✅ Status теперь может меняться без ошибок
- ✅ Комментарии в коде помогают понять логику
- ✅ Добавлены дополнительные CHECK constraints для валидации
