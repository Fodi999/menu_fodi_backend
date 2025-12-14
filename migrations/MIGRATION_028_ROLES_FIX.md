# 🔧 Миграция 028: Исправление ролей пользователей

## Проблема
Новые пользователи регистрируются с ролью `user`, которая:
- ❌ Не существует в enum `Role` 
- ❌ Не даёт доступа к `/api/fridge/*` endpoints
- ❌ Вызывает ошибку: `Access denied: insufficient permissions`

## Решение

### Шаг 1: Открой Render Dashboard SQL Shell
1. https://dashboard.render.com/
2. Выбери базу **menu_fodi_backend**
3. Перейди на вкладку **Query**

### Шаг 2: Выполни миграцию

Скопируй и выполни:

```sql
-- Добавляем значения в enum если их нет
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_enum WHERE enumlabel = 'home_chef' AND enumtypid = 'Role'::regtype) THEN
        ALTER TYPE "Role" ADD VALUE 'home_chef';
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM pg_enum WHERE enumlabel = 'pro_chef' AND enumtypid = 'Role'::regtype) THEN
        ALTER TYPE "Role" ADD VALUE 'pro_chef';
    END IF;
END $$;

-- Обновляем всех пользователей с ролью 'user' на 'home_chef'
UPDATE "User"
SET role = 'home_chef'
WHERE role = 'user';

-- Проверка
SELECT role, COUNT(*) as user_count
FROM "User"
GROUP BY role
ORDER BY role;
```

### Шаг 3: Ожидаемый результат

После выполнения должны увидеть:

| role       | user_count |
|------------|------------|
| admin      | 1-2        |
| home_chef  | 10-50      |
| pro_chef   | 0-5        |

**НЕ должно быть строки с `user`!**

### Шаг 4: Проверка в коде

После деплоя (Koyeb автоматически задеплоит коммит `3199d58`):

```bash
# Регистрация нового пользователя
curl -X POST https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "newuser@example.com",
    "password": "TestPass123!",
    "name": "New User"
  }'
```

Ответ должен содержать:
```json
{
  "data": {
    "user": {
      "role": "home_chef"  // ✅ Правильная роль!
    },
    "token": "..."
  }
}
```

### Шаг 5: Тест fridge API

```bash
# Получаем токен
TOKEN=$(curl -s -X POST https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"dima@example.com","password":"password123"}' | jq -r '.data.token')

# Проверяем fridge
curl -X GET https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/fridge/items \
  -H "Authorization: Bearer $TOKEN"
```

Ожидаемый ответ:
```json
{
  "data": {
    "items": [],
    "count": 0
  },
  "success": true
}
```

## 📝 Что исправлено

### В коде (коммит 3199d58):
- ✅ Дефолтная роль изменена с `"user"` на `models.RoleHomeChef`
- ✅ RegisterRequest не принимает роль из запроса
- ✅ Роль назначается только сервером

### В базе данных (миграция 028):
- ✅ Добавлены значения `home_chef` и `pro_chef` в enum
- ✅ Все существующие пользователи с `role='user'` обновлены на `home_chef`

## 🎯 Результат

После миграции:
- ✅ Новые пользователи сразу получают доступ к fridge API
- ✅ Существующие пользователи получают доступ после обновления
- ✅ Ошибка "Access denied" исчезает
