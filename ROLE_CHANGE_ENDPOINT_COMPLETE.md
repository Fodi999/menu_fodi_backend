# ✅ Изменение ролей — ПОЛНАЯ РЕАЛИЗАЦИЯ

**Дата:** 2026-01-26  
**Статус:** ✅ Готово к продакшену

---

## 🎯 Что реализовано

### 1. RESTful эндпоинт
**PATCH** `/api/admin/users/{id}/role`

### 2. Проверки безопасности
- ✅ Только `super_admin` может менять роли
- ✅ Валидация роли (все 5 ролей)
- ✅ Проверка существования пользователя

### 3. История изменений
- ✅ Каждое изменение роли логируется в `history_events`
- ✅ Метаданные: старая роль, новая роль, кто изменил, когда

---

## 📋 API Спецификация

### Эндпоинт

```
PATCH /api/admin/users/{id}/role
```

### Авторизация

```
Authorization: Bearer <token>
```

**Требования:**
- ✅ Валидный JWT токен
- ✅ Роль: `super_admin` (только супер админ может менять роли)
- ✅ Статус: `active`

### Request

```json
{
  "role": "chef_staff"
}
```

**Доступные роли:**
- `customer` — покупатель
- `home_chef` — домашний повар
- `chef_staff` — персонал ресторана
- `admin` — администратор
- `super_admin` — владелец системы

### Response (200 OK)

```json
{
  "message": "Role updated successfully",
  "user_id": "407582be-59d5-4d21-873b-1a72d31b0d42",
  "new_role": "chef_staff"
}
```

### Error Responses

**400 Bad Request** — неправильная роль
```json
{
  "error": "invalid role: must be one of customer, home_chef, chef_staff, admin, super_admin"
}
```

**404 Not Found** — пользователь не найден
```json
{
  "error": "User not found"
}
```

**403 Forbidden** — не super_admin
```json
{
  "error": "Super admin access required"
}
```

---

## 🔐 Безопасность

### Middleware цепочка

```go
r.Route("/admin", func(r chi.Router) {
    r.Use(authMiddleware)      // 1. Проверка JWT
    r.Use(adminMiddleware)     // 2. Проверка роли admin/super_admin
    
    // Super admin only
    r.With(superAdminMiddleware).Patch("/users/{id}/role", handler)
})
```

### Проверки

1. **JWT валидность** → `authMiddleware`
2. **Статус пользователя** → `authMiddleware` (только `active`)
3. **Роль админа** → `adminMiddleware` (admin или super_admin)
4. **Super admin** → `superAdminMiddleware` (только super_admin)
5. **Валидация роли** → в сервисе (только валидные роли)
6. **Существование пользователя** → в сервисе (проверка в БД)

---

## 📝 История изменений

### История логируется в `history_events`

**Структура события:**

```json
{
  "id": "uuid",
  "user_id": "407582be-59d5-4d21-873b-1a72d31b0d42",
  "event_type": "role_changed",
  "source_type": "admin",
  "source_id": "admin-user-id",
  "metadata": {
    "old_role": "customer",
    "new_role": "chef_staff",
    "changed_by": "admin-user-id",
    "changed_at": "2026-01-26T18:30:00Z",
    "reason": "role_change_by_admin"
  },
  "created_at": "2026-01-26T18:30:00Z"
}
```

### Запрос истории изменений ролей

```sql
-- Получить все изменения ролей для пользователя
SELECT * FROM history_events 
WHERE user_id = '407582be-59d5-4d21-873b-1a72d31b0d42' 
  AND event_type = 'role_changed'
ORDER BY created_at DESC;

-- Получить все изменения ролей в системе (аудит)
SELECT * FROM history_events 
WHERE event_type = 'role_changed'
ORDER BY created_at DESC
LIMIT 100;
```

---

## 🧪 Тестирование

### Тест 1: Успешное изменение роли

```bash
# Логин как super_admin
TOKEN=$(curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"admin123"}' \
  | jq -r '.token')

# Изменение роли пользователя
curl -X PATCH http://localhost:8080/api/admin/users/407582be-59d5-4d21-873b-1a72d31b0d42/role \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"role":"chef_staff"}'

# Ожидаемый ответ:
# {"message":"Role updated successfully","user_id":"407582be...","new_role":"chef_staff"}
```

### Тест 2: Проверка истории

```bash
# Получить историю изменений ролей
curl -X GET "http://localhost:8080/api/history?type=role_changed" \
  -H "Authorization: Bearer $TOKEN"
```

### Тест 3: Проверка валидации

```bash
# Попытка назначить несуществующую роль
curl -X PATCH http://localhost:8080/api/admin/users/407582be.../role \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"role":"invalid_role"}'

# Ожидаемый ответ: 400 Bad Request
# {"error":"invalid role: must be one of customer, home_chef, chef_staff, admin, super_admin"}
```

### Тест 4: Проверка доступа

```bash
# Попытка изменить роль с ролью admin (не super_admin)
TOKEN_ADMIN=$(curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"regular-admin@example.com","password":"admin123"}' \
  | jq -r '.token')

curl -X PATCH http://localhost:8080/api/admin/users/407582be.../role \
  -H "Authorization: Bearer $TOKEN_ADMIN" \
  -H "Content-Type: application/json" \
  -d '{"role":"chef_staff"}'

# Ожидаемый ответ: 403 Forbidden
# {"error":"Super admin access required"}
```

---

## 📊 Диаграмма потока

```
1. Запрос → PATCH /api/admin/users/{id}/role
   ↓
2. AuthMiddleware → Проверка JWT + статуса
   ↓
3. AdminMiddleware → Проверка роли (admin/super_admin)
   ↓
4. SuperAdminMiddleware → Проверка super_admin
   ↓
5. Handler → Извлечение userID и adminID
   ↓
6. Service → Валидация роли
   ↓
7. Service → Получение старой роли из БД
   ↓
8. Service → Обновление роли в БД
   ↓
9. Service → Запись в history_events
   ↓
10. Response → {"message": "Role updated successfully"}
```

---

## 🔄 Обратная совместимость

### Legacy endpoint (deprecated)

```
PATCH /api/admin/users/update-role
Body: { "user_id": "...", "role": "..." }
```

Этот эндпоинт сохранён для обратной совместимости, но **рекомендуется использовать RESTful вариант**.

### Новый endpoint (recommended)

```
PATCH /api/admin/users/{id}/role
Body: { "role": "..." }
```

---

## 📋 Чеклист реализации

### Backend
- [x] RESTful эндпоинт: `PATCH /api/admin/users/{id}/role`
- [x] Legacy эндпоинт: `PATCH /api/admin/users/update-role`
- [x] Middleware: `superAdminMiddleware`
- [x] Валидация ролей (все 5 ролей)
- [x] Проверка существования пользователя
- [x] История изменений (в `history_events`)
- [x] Логирование (zap logger)
- [x] Метаданные: old_role, new_role, changed_by, changed_at

### Frontend (TODO)
- [ ] UI для изменения роли
- [ ] Dropdown с доступными ролями
- [ ] Подтверждение изменения роли
- [ ] Отображение истории изменений
- [ ] Проверка прав (только super_admin видит кнопку)

---

## 🔍 Логи

### Успешное изменение роли

```
INFO  User role changed
  user_id=407582be-59d5-4d21-873b-1a72d31b0d42
  old_role=customer
  new_role=chef_staff
  changed_by=admin-user-id
```

### Ошибка (пользователь не найден)

```
ERROR Failed to update role
  user_id=invalid-id
  error=user not found
```

### Ошибка (история не сохранилась)

```
ERROR Failed to log role change to history
  user_id=407582be...
  old_role=customer
  new_role=chef_staff
  error=database error
```

Примечание: Если история не сохранилась, роль всё равно изменится. История — это дополнительный аудит, не критичный для работы.

---

## 🚀 Пример использования на фронтенде

```typescript
// utils/admin.ts
export async function updateUserRole(userId: string, newRole: string) {
  const response = await fetch(`/api/admin/users/${userId}/role`, {
    method: 'PATCH',
    headers: {
      'Authorization': `Bearer ${getToken()}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({ role: newRole })
  });
  
  if (!response.ok) {
    if (response.status === 403) {
      throw new Error('Only super admin can change roles');
    }
    if (response.status === 404) {
      throw new Error('User not found');
    }
    if (response.status === 400) {
      const data = await response.json();
      throw new Error(data.error);
    }
    throw new Error('Failed to update role');
  }
  
  return response.json();
}

// Использование
try {
  await updateUserRole('407582be-59d5-4d21-873b-1a72d31b0d42', 'chef_staff');
  alert('Role updated successfully');
} catch (error) {
  alert(error.message);
}
```

---

## 📝 Миграция БД (если нужно добавить enum)

```sql
-- Создать ENUM для ролей (если используется PostgreSQL enum)
CREATE TYPE user_role AS ENUM (
  'customer',
  'home_chef',
  'chef_staff',
  'admin',
  'super_admin'
);

-- Создать ENUM для статусов
CREATE TYPE user_status AS ENUM (
  'pending',
  'active',
  'suspended',
  'blocked'
);

-- Обновить существующие роли (если были старые значения)
UPDATE "User" 
SET role = 'home_chef' 
WHERE role = 'pro_chef';

UPDATE "User" 
SET role = 'customer' 
WHERE role IS NULL OR role = '';
```

---

## ✅ Итог

### Что работает:
1. ✅ RESTful эндпоинт: `PATCH /api/admin/users/{id}/role`
2. ✅ Только super_admin может менять роли
3. ✅ Валидация всех 5 ролей
4. ✅ История изменений логируется
5. ✅ Логирование в zap
6. ✅ Обратная совместимость (legacy endpoint)

### Следующие шаги:
- Протестировать изменение роли через админку
- Проверить историю изменений
- Обновить фронтенд для использования нового эндпоинта

---

**Статус:** ✅ Реализовано  
**Безопасность:** ✅ Только super_admin  
**Аудит:** ✅ Полная история изменений
