# 🚀 ИНСТРУКЦИЯ: Запуск миграции 027

## ✅ Шаг 1: Открой Render Dashboard

1. Перейди на https://dashboard.render.com/
2. Найди свою базу данных: **menu_fodi_backend**
3. Кликни на неё

## ✅ Шаг 2: Открой SQL Shell

1. В верхнем меню найди вкладку **"Query"** или **"Shell"**
2. Откроется SQL редактор

## ✅ Шаг 3: Скопируй и выполни миграцию

Скопируй весь код ниже и вставь в SQL редактор:

```sql
-- MIGRATION 027: Recreate user_fridge_items (MVP)
-- Drop old tables
DROP TABLE IF EXISTS "UserFridgeItem" CASCADE;
DROP TABLE IF EXISTS "user_fridge_items" CASCADE;
DROP TABLE IF EXISTS user_fridge_items CASCADE;

-- Create new table with MVP structure
CREATE TABLE user_fridge_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    user_id UUID NOT NULL
        REFERENCES "User"(id) ON DELETE CASCADE,
    
    ingredient_id UUID NOT NULL
        REFERENCES "Ingredient"(id) ON DELETE CASCADE,
    
    quantity DOUBLE PRECISION NOT NULL DEFAULT 0,
    unit VARCHAR(50) NOT NULL DEFAULT 'шт',
    expires_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(user_id, ingredient_id)
);

-- Create indexes
CREATE INDEX idx_user_fridge_items_user_id 
    ON user_fridge_items(user_id);

CREATE INDEX idx_user_fridge_items_ingredient_id 
    ON user_fridge_items(ingredient_id);

CREATE INDEX idx_user_fridge_items_expires_at 
    ON user_fridge_items(expires_at);

CREATE INDEX idx_user_fridge_items_user_expires 
    ON user_fridge_items(user_id, expires_at);

-- Add comment
COMMENT ON TABLE user_fridge_items 
IS 'MVP: User fridge items with simplified structure';
```

## ✅ Шаг 4: Нажми "Run Query"

Ожидаемый результат:

```
DROP TABLE
CREATE TABLE
CREATE INDEX (4 times)
COMMENT
```

**Без ошибок!**

## ✅ Шаг 5: Проверка (ОБЯЗАТЕЛЬНО!)

Выполни эту проверку в том же SQL редакторе:

```sql
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
| expires_at     | timestamp         | YES         |
| created_at     | timestamp         | NO          |

## ✅ Шаг 6: Проверка Foreign Keys

```sql
SELECT 
    tc.constraint_name,
    kcu.column_name,
    ccu.table_name AS foreign_table
FROM information_schema.table_constraints tc
JOIN information_schema.key_column_usage kcu 
    ON tc.constraint_name = kcu.constraint_name
JOIN information_schema.constraint_column_usage ccu 
    ON ccu.constraint_name = tc.constraint_name
WHERE tc.table_name = 'user_fridge_items'
AND tc.constraint_type = 'FOREIGN KEY';
```

**Ожидаемый результат:**

- `user_fridge_items.user_id` → `User`
- `user_fridge_items.ingredient_id` → `Ingredient`

## 🔥 Шаг 7: После миграции

### 7.1 Получи новый JWT токен

```bash
curl -X POST https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "твой_email@example.com",
    "password": "твой_пароль"
  }'
```

### 7.2 Проверь API напрямую

```bash
curl -X GET https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/fridge/items \
  -H "Authorization: Bearer ТВОЙ_НОВЫЙ_TOKEN"
```

**Ожидаемый ответ:**

```json
{
  "data": {
    "items": []
  },
  "success": true
}
```

### 7.3 Обнови страницу /fridge

Перейди на http://localhost:3000/fridge и обнови страницу.

**Ошибка должна исчезнуть!** ✅

---

## ❓ Если всё ещё 500 ошибка после миграции

Проверь в Koyeb, что backend перезапустился с новым кодом:

1. Открой https://app.koyeb.com/
2. Найди сервис `yeasty-madelaine-fodi999-671ccdf5`
3. Проверь, что последний деплой содержит commit `2e55261`
4. Если нет - нажми "Redeploy"

---

## 📝 Резюме

✅ Миграция создаёт таблицу `user_fridge_items` с правильной структурой  
✅ Foreign Keys указывают на `"User"` и `"Ingredient"` (с кавычками!)  
✅ `expires_at` nullable (может быть NULL)  
✅ `quantity` может быть 0  
✅ `unit` VARCHAR(50) поддерживает кириллицу 'шт'  

После выполнения всех шагов API `/api/fridge/items` будет возвращать `200 OK` с пустым массивом `items: []`.
