# 🚀 Admin API Quick Cheat Sheet

Быстрая справка по всем админ-эндпоинтам - скопируй и используй!

---

## 🔐 Аутентификация

```bash
# 1. Получить токен
TOKEN=$(curl -s -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.com",
    "password": "admin_password_123"
  }' | jq -r '.token')

echo "Токен: $TOKEN"

# 2. Сохранить в переменную для всех запросов
ADMIN_API="http://localhost:8080/api/admin"
HEADERS="-H 'Authorization: Bearer $TOKEN' -H 'Content-Type: application/json'"
```

---

## 👥 Users Endpoints

### Получить всех пользователей
```bash
curl -s -X GET $ADMIN_API/users \
  -H "Authorization: Bearer $TOKEN" | jq '.'
```

### Обновить пользователя
```bash
curl -s -X PUT $ADMIN_API/users/USER_ID \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "New Name",
    "email": "newemail@example.com"
  }' | jq '.'
```

### Удалить пользователя ⚠️
```bash
curl -s -X DELETE $ADMIN_API/users/USER_ID \
  -H "Authorization: Bearer $TOKEN"
```

### Изменить роль пользователя
```bash
# user → admin
curl -s -X PATCH $ADMIN_API/users/update-role \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "USER_ID",
    "role": "admin"
  }' | jq '.'
```

---

## 📦 Orders Endpoints

### Получить ВСЕ заказы
```bash
curl -s -X GET $ADMIN_API/orders \
  -H "Authorization: Bearer $TOKEN" | jq '.'
```

### Получить последние 10 заказов (быстрее)
```bash
curl -s -X GET $ADMIN_API/orders/recent \
  -H "Authorization: Bearer $TOKEN" | jq '.'
```

### Изменить статус заказа
```bash
# Статусы: pending, processing, shipped, delivered, completed, cancelled
curl -s -X PUT $ADMIN_API/orders/ORDER_ID/status \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"status": "shipped"}' | jq '.'
```

---

## 📊 Statistics & Profile

### Получить статистику
```bash
curl -s -X GET $ADMIN_API/stats \
  -H "Authorization: Bearer $TOKEN" | jq '.'

# Response: {totalUsers: 156, totalOrders: 2340}
```

### Получить профиль админа
```bash
curl -s -X GET $ADMIN_API/profile \
  -H "Authorization: Bearer $TOKEN" | jq '.'

# Response: {id, name, email, role, managedUsers, managedOrders, ...}
```

---

## 🔍 Полезные jq фильтры

```bash
# Получить только имена и emails пользователей
curl -s $ADMIN_API/users -H "Authorization: Bearer $TOKEN" | \
  jq '.[] | {name, email}'

# Получить количество пользователей
curl -s $ADMIN_API/users -H "Authorization: Bearer $TOKEN" | \
  jq 'length'

# Получить только админов
curl -s $ADMIN_API/users -H "Authorization: Bearer $TOKEN" | \
  jq '.[] | select(.role == "admin")'

# Получить последний заказ
curl -s $ADMIN_API/orders -H "Authorization: Bearer $TOKEN" | \
  jq '.[0]'

# Получить заказы со статусом "pending"
curl -s $ADMIN_API/orders -H "Authorization: Bearer $TOKEN" | \
  jq '.[] | select(.status == "pending")'
```

---

## 📝 TypeScript Types

```typescript
// User
interface User {
  id: string;
  email: string;
  name: string;
  role: "user" | "admin";
  createdAt: string;
}

// Order
interface Order {
  id: string;
  userId: string;
  status: "pending" | "processing" | "shipped" | "delivered" | "completed" | "cancelled";
  totalPrice: number;
  itemsCount: number;
  createdAt: string;
  updatedAt: string;
}

// AdminStats
interface AdminStats {
  totalUsers: number;
  totalOrders: number;
}

// AdminProfile
interface AdminProfile {
  id: string;
  name: string;
  email: string;
  role: "admin";
  createdAt: string;
  managedUsers: number;
  managedOrders: number;
  totalStats: { users: number; orders: number };
}
```

---

## ❌ Error Codes

| Code | Meaning | Fix |
|------|---------|-----|
| 200 | ✅ Success | Хорошо! |
| 400 | Invalid request | Проверь JSON синтаксис |
| 401 | Unauthorized | Токен отсутствует или истёк |
| 403 | Forbidden | Недостаточно прав (нужна роль admin) |
| 404 | Not found | User/Order не найден |
| 500 | Server error | Ошибка на сервере |

---

## 🧪 Полный тестовый сценарий

```bash
#!/bin/bash

# Сохраняем токен
TOKEN=$(curl -s -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"admin_password_123"}' \
  | jq -r '.token')

ADMIN_API="http://localhost:8080/api/admin"

echo "=== Admin API Testing ==="

# 1. Получить статистику
echo "1️⃣ Stats:"
curl -s -X GET $ADMIN_API/stats -H "Authorization: Bearer $TOKEN" | jq '.'

# 2. Получить пользователей
echo "2️⃣ Users:"
curl -s -X GET $ADMIN_API/users -H "Authorization: Bearer $TOKEN" | jq '.[] | {name, email}' | head -10

# 3. Получить последние заказы
echo "3️⃣ Recent Orders:"
curl -s -X GET $ADMIN_API/orders/recent -H "Authorization: Bearer $TOKEN" | jq '.[] | {id, status, totalPrice}' | head -5

# 4. Получить профиль админа
echo "4️⃣ Admin Profile:"
curl -s -X GET $ADMIN_API/profile -H "Authorization: Bearer $TOKEN" | jq '.'

echo "=== Testing Complete ==="
```

---

## 🎯 Частые операции

### Найти пользователя по email
```bash
curl -s -X GET $ADMIN_API/users \
  -H "Authorization: Bearer $TOKEN" | \
  jq '.[] | select(.email == "user@example.com")'
```

### Повысить пользователя до админа
```bash
# 1. Найти ID пользователя
USER_ID=$(curl -s -X GET $ADMIN_API/users \
  -H "Authorization: Bearer $TOKEN" | \
  jq -r '.[] | select(.email == "user@example.com") | .id')

# 2. Изменить роль
curl -s -X PATCH $ADMIN_API/users/update-role \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"user_id\": \"$USER_ID\", \"role\": \"admin\"}"
```

### Переслать заказ
```bash
curl -s -X PUT $ADMIN_API/orders/ORDER_ID/status \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"status": "shipped"}'
```

### Получить статистику по пользователям
```bash
curl -s -X GET $ADMIN_API/users \
  -H "Authorization: Bearer $TOKEN" | \
  jq '{total: length, admins: [.[] | select(.role=="admin")] | length}'
```

---

## 📚 Дополнительно

Полная документация: `ADMIN_ENDPOINTS_DATA_STRUCTURE.md`

Все эндпоинты требуют:
- ✅ JWT токен в заголовке `Authorization: Bearer $TOKEN`
- ✅ Роль `admin` (кроме /api/user/profile)
- ✅ Content-Type: application/json (для PUT/PATCH/DELETE)

