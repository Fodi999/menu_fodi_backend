# 🗑️ Удаление пользователя — DELETE /api/admin/users/{id}

**Дата:** 2026-01-26  
**Статус:** ✅ Реализовано  
**Доступ:** 🔴 Только Super Admin

---

## 🎯 Описание

Эндпоинт для **безвозвратного удаления** пользователя из базы данных.

⚠️ **КРИТИЧЕСКИ ВАЖНО:**
- Это **необратимая операция**
- Удаляет пользователя и **все связанные данные** (через CASCADE)
- Доступно **только супер-админу** (`super_admin`)
- Обычный `admin` **НЕ МОЖЕТ** удалять пользователей

---

## 📋 API Спецификация

### Эндпоинт

```
DELETE /api/admin/users/{id}
```

### Авторизация

```
Authorization: Bearer <super_admin_token>
```

**Требования:**
- ✅ Валидный JWT токен
- ✅ Роль: `super_admin` (только супер админ)
- ✅ Статус: `active`

### URL Parameters

| Параметр | Тип    | Обязательный | Описание                |
|----------|--------|--------------|-------------------------|
| `id`     | string | ✅ Да        | UUID пользователя       |

### Request Body

**Нет** — удаление происходит через URL параметр `{id}`

### Response (200 OK)

```json
{
  "message": "User deleted successfully"
}
```

### Error Responses

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

**401 Unauthorized** — нет токена или токен невалидный
```json
{
  "error": "Unauthorized"
}
```

**500 Internal Server Error** — ошибка базы данных
```json
{
  "error": "Failed to delete user"
}
```

---

## 🔐 Безопасность

### Middleware цепочка

```go
r.Route("/admin", func(r chi.Router) {
    r.Use(authMiddleware)       // 1. Проверка JWT
    r.Use(adminMiddleware)      // 2. Проверка роли admin/super_admin
    
    // Super admin only
    r.With(superAdminMiddleware).Delete("/users/{id}", m.handlers.DeleteUser)
})
```

### Проверки

1. **JWT валидность** → `authMiddleware`
2. **Статус пользователя** → `authMiddleware` (только `active`)
3. **Роль админа** → `adminMiddleware` (admin или super_admin)
4. **Super admin** → `superAdminMiddleware` (только super_admin)
5. **Существование пользователя** → в сервисе

---

## 💾 Реализация

### Handler

```go
// internal/modules/admin/transport/http/handlers.go

func (h *AdminHandlers) DeleteUser(w http.ResponseWriter, r *http.Request) {
    userID := chi.URLParam(r, "id")

    err := h.service.DeleteUser(userID)
    if err != nil {
        if err.Error() == "user not found" {
            utils.RespondWithError(w, http.StatusNotFound, "User not found")
        } else {
            utils.RespondWithError(w, http.StatusInternalServerError, "Failed to delete user")
        }
        return
    }

    utils.RespondWithJSON(w, http.StatusOK, map[string]string{
        "message": "User deleted successfully"
    })
}
```

### Service

```go
// internal/modules/admin/service/service.go

func (s *adminService) DeleteUser(userID string) error {
    result := s.db.Delete(&models.User{}, "id = ?", userID)
    if result.Error != nil {
        return result.Error
    }
    if result.RowsAffected == 0 {
        return errors.New("user not found")
    }
    return nil
}
```

---

## ⚠️ Что удаляется

### Прямое удаление

```sql
DELETE FROM "User" WHERE id = '{user_id}';
```

### Каскадное удаление (ON DELETE CASCADE)

При удалении пользователя **автоматически удаляются**:

1. **fridge_items** — элементы холодильника
2. **notifications** — уведомления
3. **token_bank** — токен банк
4. **user_fridge_items** — элементы холодильника
5. **user_menu_items** — элементы меню
6. **user_recipe_sessions** — сессии рецептов
7. **user_saved_recipes** — сохранённые рецепты
8. **RecipeCookLog** — история приготовления

### Связанные данные с SET NULL (ON DELETE SET NULL)

1. **Recipe.author_id** — рецепты остаются, но автор = NULL
2. **token_transactions** — транзакции остаются, но пользователь = NULL

### Связанные данные с RESTRICT (НЕ ПОЗВОЛЯЕТ УДАЛИТЬ)

1. **ChatMessage** — если у пользователя есть сообщения, удаление заблокировано

---

## 🧪 Тестирование

### Тест 1: Успешное удаление пользователя

```bash
# Логин как super_admin
SUPER_ADMIN_TOKEN=$(curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"superadmin@example.com","password":"admin123"}' \
  | jq -r '.token')

# Удалить пользователя
curl -X DELETE http://localhost:8080/api/admin/users/407582be-59d5-4d21-873b-1a72d31b0d42 \
  -H "Authorization: Bearer $SUPER_ADMIN_TOKEN"

# Ожидаемый ответ: 200 OK
# {"message":"User deleted successfully"}
```

### Тест 2: Попытка удалить несуществующего пользователя

```bash
curl -X DELETE http://localhost:8080/api/admin/users/00000000-0000-0000-0000-000000000000 \
  -H "Authorization: Bearer $SUPER_ADMIN_TOKEN"

# Ожидаемый ответ: 404 Not Found
# {"error":"User not found"}
```

### Тест 3: Попытка удалить пользователя обычным админом

```bash
# Логин как обычный admin (не super_admin)
ADMIN_TOKEN=$(curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"admin123"}' \
  | jq -r '.token')

# Попытка удалить
curl -X DELETE http://localhost:8080/api/admin/users/407582be... \
  -H "Authorization: Bearer $ADMIN_TOKEN"

# Ожидаемый ответ: 403 Forbidden
# {"error":"Super admin access required"}
```

### Тест 4: Проверка каскадного удаления

```bash
# Проверить, что все связанные данные удалены
psql $DATABASE_URL -c "
  SELECT 
    (SELECT COUNT(*) FROM fridge_items WHERE user_id = '407582be...') as fridge_items,
    (SELECT COUNT(*) FROM notifications WHERE user_id = '407582be...') as notifications,
    (SELECT COUNT(*) FROM token_bank WHERE user_id = '407582be...') as token_bank;
"

# Ожидаемый результат: все 0
```

---

## 🚨 Важные предупреждения

### 1. Необратимая операция

❌ **НЕЛЬЗЯ** восстановить удалённого пользователя
- Все данные удаляются безвозвратно
- Резервное копирование — единственный способ восстановления

### 2. Проверка перед удалением

Рекомендуется проверить, что пользователь действительно должен быть удалён:

```bash
# Проверить данные пользователя перед удалением
curl -X GET http://localhost:8080/api/admin/users/407582be-59d5-4d21-873b-1a72d31b0d42 \
  -H "Authorization: Bearer $SUPER_ADMIN_TOKEN"

# Проверить связанные данные
psql $DATABASE_URL -c "
  SELECT 
    u.id,
    u.email,
    u.role,
    COUNT(f.id) as fridge_items_count,
    COUNT(r.id) as recipes_count,
    COUNT(n.id) as notifications_count
  FROM \"User\" u
  LEFT JOIN fridge_items f ON f.user_id = u.id
  LEFT JOIN \"Recipe\" r ON r.author_id = u.id
  LEFT JOIN notifications n ON n.user_id = u.id
  WHERE u.id = '407582be-59d5-4d21-873b-1a72d31b0d42'
  GROUP BY u.id, u.email, u.role;
"
```

### 3. Альтернатива удалению — блокировка

Вместо удаления рекомендуется **заблокировать** пользователя:

```bash
# Заблокировать пользователя (обратимая операция)
curl -X PATCH http://localhost:8080/api/admin/users/407582be.../status \
  -H "Authorization: Bearer $SUPER_ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"status":"blocked"}'
```

**Преимущества блокировки:**
- ✅ Обратимая операция
- ✅ Данные сохраняются
- ✅ Можно восстановить доступ
- ✅ История остаётся

---

## 📊 Диаграмма потока

```
1. Запрос → DELETE /api/admin/users/{id}
   ↓
2. AuthMiddleware → Проверка JWT + статуса
   ↓
3. AdminMiddleware → Проверка роли (admin/super_admin)
   ↓
4. SuperAdminMiddleware → Проверка super_admin
   ↓
5. Handler → Извлечение userID
   ↓
6. Service → DELETE FROM "User" WHERE id = ?
   ↓
7. Database → Каскадное удаление связанных данных
   ↓
8. Response → {"message": "User deleted successfully"}
```

---

## 🔄 Frontend пример

```typescript
// utils/admin.ts
export async function deleteUser(userId: string): Promise<void> {
  const confirmed = window.confirm(
    'Are you sure you want to delete this user? This action CANNOT be undone!'
  );
  
  if (!confirmed) {
    return;
  }
  
  const response = await fetch(`/api/admin/users/${userId}`, {
    method: 'DELETE',
    headers: {
      'Authorization': `Bearer ${getToken()}`
    }
  });
  
  if (!response.ok) {
    if (response.status === 403) {
      throw new Error('Only super admin can delete users');
    }
    if (response.status === 404) {
      throw new Error('User not found');
    }
    throw new Error('Failed to delete user');
  }
  
  return;
}

// Использование
try {
  await deleteUser('407582be-59d5-4d21-873b-1a72d31b0d42');
  alert('User deleted successfully');
  // Обновить список пользователей
} catch (error) {
  alert(error.message);
}
```

---

## 🗄️ SQL запросы

### Проверить, сколько данных будет удалено

```sql
-- Проверить пользователя и связанные данные
SELECT 
    u.id,
    u.email,
    u.role,
    u.status,
    COUNT(DISTINCT f.id) as fridge_items,
    COUNT(DISTINCT n.id) as notifications,
    COUNT(DISTINCT tb.id) as token_banks,
    COUNT(DISTINCT r.id) as recipes_authored
FROM "User" u
LEFT JOIN fridge_items f ON f.user_id = u.id
LEFT JOIN notifications n ON n.user_id = u.id
LEFT JOIN token_bank tb ON tb.user_id = u.id
LEFT JOIN "Recipe" r ON r.author_id = u.id
WHERE u.id = '407582be-59d5-4d21-873b-1a72d31b0d42'
GROUP BY u.id, u.email, u.role, u.status;
```

### Удалить пользователя напрямую через SQL

```sql
-- Удалить пользователя (каскадное удаление сработает автоматически)
DELETE FROM "User" WHERE id = '407582be-59d5-4d21-873b-1a72d31b0d42';

-- Проверить результат
SELECT 
    (SELECT COUNT(*) FROM "User" WHERE id = '407582be...') as user_exists,
    (SELECT COUNT(*) FROM fridge_items WHERE user_id = '407582be...') as fridge_items,
    (SELECT COUNT(*) FROM notifications WHERE user_id = '407582be...') as notifications;
```

---

## 📝 Рекомендации

### Когда удалять пользователя

✅ **Удалять можно:**
- Спам-аккаунты
- Тестовые аккаунты
- Дублирующиеся аккаунты
- Аккаунты, нарушающие правила (после предупреждения)

❌ **Не рекомендуется удалять:**
- Активные пользователи с данными
- Пользователи с историей заказов
- Пользователи с рецептами (используются другими)

### Альтернативные действия

Вместо удаления рассмотрите:

1. **Блокировка** — `status = blocked`
2. **Приостановка** — `status = suspended`
3. **Архивация** — добавить поле `archived = true`
4. **Анонимизация** — заменить личные данные на `[deleted]`

---

## ✅ Итог

### Что работает

- ✅ Эндпоинт: `DELETE /api/admin/users/{id}`
- ✅ Только `super_admin` может удалять
- ✅ Каскадное удаление связанных данных
- ✅ Проверка существования пользователя
- ✅ Корректные коды ошибок (404, 403, 401, 500)

### Безопасность

- ✅ Требуется JWT токен
- ✅ Требуется роль `super_admin`
- ✅ Обычный `admin` не может удалять
- ✅ Проверка статуса пользователя (только `active` может выполнять запросы)

### Рекомендации

- ⚠️ Используйте с осторожностью
- ⚠️ Проверяйте данные перед удалением
- ⚠️ Рассмотрите блокировку вместо удаления
- ⚠️ Делайте резервные копии базы данных

---

**Статус:** ✅ Реализовано  
**Доступ:** 🔴 Только Super Admin  
**Безопасность:** ✅ Защищено
