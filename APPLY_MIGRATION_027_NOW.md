# ⚡ СРОЧНО: Применение миграции 027 через Render Dashboard

## 🎯 Проблема
Backend код обновлён (коммит `f902bf4`), но **миграция 027 не применена в production БД**.

Результат: API `/api/fridge/items` возвращает `500 Internal Server Error`.

---

## ✅ РЕШЕНИЕ (3 минуты)

### Шаг 1: Открой Render Dashboard
1. Перейди: **https://dashboard.render.com/**
2. Найди базу: **menu_fodi_backend**
3. Кликни на неё
4. Перейди на вкладку **"Query"** (вверху)

### Шаг 2: Выполни миграцию

**Скопируй весь SQL ниже и вставь в Query Editor:**

```sql
-- =====================================================
-- МИГРАЦИЯ 027: Пересоздание user_fridge_items (MVP)
-- =====================================================

-- Удаляем старые таблицы
DROP TABLE IF EXISTS "UserFridgeItem" CASCADE;
DROP TABLE IF EXISTS "user_fridge_items" CASCADE;
DROP TABLE IF EXISTS user_fridge_items CASCADE;

-- Создаем новую таблицу с правильной MVP структурой
CREATE TABLE user_fridge_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    user_id UUID NOT NULL
        REFERENCES "User"(id) ON DELETE CASCADE,
    
    ingredient_id UUID NOT NULL
        REFERENCES "Ingredient"(id) ON DELETE CASCADE,
    
    quantity DOUBLE PRECISION NOT NULL DEFAULT 0,
    unit VARCHAR(50) NOT NULL DEFAULT 'шт',
    expires_at TIMESTAMP,  -- ✅ NULLABLE - может быть NULL
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(user_id, ingredient_id)
);

-- Создаем индексы для оптимизации
CREATE INDEX idx_user_fridge_items_user_id 
    ON user_fridge_items(user_id);

CREATE INDEX idx_user_fridge_items_ingredient_id 
    ON user_fridge_items(ingredient_id);

CREATE INDEX idx_user_fridge_items_expires_at 
    ON user_fridge_items(expires_at);

CREATE INDEX idx_user_fridge_items_user_expires 
    ON user_fridge_items(user_id, expires_at);

-- Добавляем комментарий
COMMENT ON TABLE user_fridge_items 
IS 'MVP: User fridge items with simplified structure';

-- ✅ ПРОВЕРКА: Смотрим структуру таблицы
SELECT column_name, data_type, is_nullable
FROM information_schema.columns
WHERE table_name = 'user_fridge_items'
ORDER BY ordinal_position;
```

### Шаг 3: Нажми "Run Query"

Ожидаемый результат:
```
DROP TABLE
DROP TABLE
DROP TABLE
CREATE TABLE
CREATE INDEX
CREATE INDEX
CREATE INDEX
CREATE INDEX
COMMENT
```

### Шаг 4: Проверка (ОБЯЗАТЕЛЬНО!)

Выполни эту проверку:
```sql
-- Должно быть 7 колонок
SELECT column_name, data_type, is_nullable
FROM information_schema.columns
WHERE table_name = 'user_fridge_items'
ORDER BY ordinal_position;
```

**Ожидаемый результат:**

| column_name    | data_type          | is_nullable |
|----------------|-------------------|-------------|
| id             | uuid              | NO          |
| user_id        | uuid              | NO          |
| ingredient_id  | uuid              | NO          |
| quantity       | double precision  | NO          |
| unit           | character varying | NO          |
| **expires_at** | **timestamp**     | **YES** ✅  |
| created_at     | timestamp         | NO          |

⚠️ **КРИТИЧНО**: `expires_at` должен быть **YES** (nullable)!

---

## 🔄 Шаг 5: Redeploy в Koyeb

После применения миграции:

1. Открой **https://app.koyeb.com/**
2. Найди сервис **yeasty-madelaine-fodi999-671ccdf5**
3. Нажми **"Redeploy"** или **"Trigger Deploy"**
4. Подожди 2-3 минуты

---

## 🧪 Шаг 6: Проверка работоспособности

После деплоя выполни:

```bash
# 1. Регистрация нового пользователя
curl -X POST https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test_migration@example.com","password":"Test123!","name":"Test Migration"}'

# 2. Получи токен из ответа выше, затем:
curl -X GET https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/fridge/items \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

**Ожидаемый ответ:**
```json
{
  "data": {
    "items": [],
    "count": 0
  },
  "success": true
}
```

✅ **УСПЕХ!** Если видишь пустой массив - всё работает!

---

## 📊 Что было исправлено

### В коде (коммиты f902bf4 + 3199d58 + 79e93a9):
- ✅ `ExpiresAt *time.Time` - nullable pointer
- ✅ Nil-check в сервисах
- ✅ Роль по умолчанию `home_chef`
- ✅ AI handlers обновлены

### В базе данных (миграция 027):
- ✅ Таблица `user_fridge_items` создана
- ✅ `expires_at TIMESTAMP` - nullable
- ✅ Foreign keys на `User` и `Ingredient`
- ✅ Индексы для производительности

---

## 🎯 Финальный чек-лист

После всех шагов должно работать:

- [ ] Регистрация даёт роль `home_chef`
- [ ] `/api/fridge/items` возвращает `200 OK`
- [ ] Пустой холодильник показывает `[]`
- [ ] Frontend показывает "Twoja lodówka jest pusta"

---

## ❓ Если всё ещё не работает

Проверь логи в Koyeb:
1. https://app.koyeb.com/
2. Service → Logs
3. Ищи ошибки типа `failed to get items` или `scan error`

Если видишь ошибки - напиши в чат, разберём!
