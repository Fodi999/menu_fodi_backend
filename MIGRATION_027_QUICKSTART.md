# ⚡ БЫСТРЫЙ СТАРТ: Миграция 027

## Проблема
```
GET /api/fridge/items → 500 Internal Server Error
message: "failed to get items"
```

**Причина**: Таблица `user_fridge_items` не существует в базе данных.

## ✅ Решение (3 минуты)

### Шаг 1: Открой Render Dashboard SQL Shell
1. Перейди: https://dashboard.render.com/d/dpg-csekngjtq21c73ctcqh0-a
2. Кликни на базу **menu_fodi_backend**
3. Открой вкладку **"Query"** (вверху)

### Шаг 2: Скопируй и выполни миграцию

```sql
DROP TABLE IF EXISTS "UserFridgeItem" CASCADE;
DROP TABLE IF EXISTS "user_fridge_items" CASCADE;
DROP TABLE IF EXISTS user_fridge_items CASCADE;

CREATE TABLE user_fridge_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES "User"(id) ON DELETE CASCADE,
    ingredient_id UUID NOT NULL REFERENCES "Ingredient"(id) ON DELETE CASCADE,
    quantity DOUBLE PRECISION NOT NULL DEFAULT 0,
    unit VARCHAR(50) NOT NULL DEFAULT 'шт',
    expires_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, ingredient_id)
);

CREATE INDEX idx_user_fridge_items_user_id ON user_fridge_items(user_id);
CREATE INDEX idx_user_fridge_items_ingredient_id ON user_fridge_items(ingredient_id);
CREATE INDEX idx_user_fridge_items_expires_at ON user_fridge_items(expires_at);
CREATE INDEX idx_user_fridge_items_user_expires ON user_fridge_items(user_id, expires_at);

COMMENT ON TABLE user_fridge_items IS 'MVP: User fridge items with simplified structure';
```

### Шаг 3: Нажми "Run Query"

Ожидаемый результат:
```
DROP TABLE
CREATE TABLE
CREATE INDEX (4 times)
COMMENT
```

### Шаг 4: Проверка (обязательно!)

Выполни эту проверку:
```sql
SELECT column_name, data_type, is_nullable
FROM information_schema.columns
WHERE table_name = 'user_fridge_items'
ORDER BY ordinal_position;
```

Должно быть **7 колонок**: id, user_id, ingredient_id, quantity, unit, expires_at, created_at

### Шаг 5: Обнови страницу /fridge

Просто обнови страницу в браузере:
- http://localhost:3000/fridge

**Ошибка 500 должна исчезнуть! ✅**

---

## 📌 Почему не работает psql из терминала?

Render блокирует прямые SSL соединения. Нужно использовать Web UI (Query tab) или настроить SSL сертификаты.

## 🎯 После миграции

API будет возвращать пустой массив:
```json
{
  "data": {
    "items": []
  },
  "success": true
}
```

Это нормально - холодильник пока пустой!
