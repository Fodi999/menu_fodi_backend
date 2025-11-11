# 📊 Admin Endpoints - Complete Data Structure Reference

Полное описание всех данных, которые возвращают админ-эндпоинты.

---

## 1️⃣ GET /api/admin/users

**Описание:** Получить список всех пользователей системы

**Метод:** GET  
**Требует:** JWT токен + роль admin  
**Возвращает:** Массив объектов User

### Request

```bash
curl -X GET https://api.example.com/api/admin/users \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json"
```

### Response (200 OK)

```json
[
  {
    "id": "7ec8aba4-8195-4be1-a9a8-067c30aae306",
    "email": "admin@example.com",
    "name": "System Administrator",
    "password": "$2a$10$...", // hashed
    "role": "admin",
    "createdAt": "2024-01-01T00:00:00Z"
  },
  {
    "id": "user-id-123",
    "email": "user@example.com",
    "name": "John Doe",
    "password": "$2a$10$...", // hashed
    "role": "user",
    "createdAt": "2024-10-15T14:30:00Z"
  }
]
```

### Data Structure

```typescript
interface User {
  id: string;                // UUID уникальный идентификатор
  email: string;             // Email пользователя
  name: string;              // Имя пользователя
  password: string;          // Хеш пароля (bcrypt)
  role: "user" | "admin";    // Роль в системе
  createdAt: string;         // ISO 8601 дата создания
}
```

### Фильтрация и сортировка

Текущая реализация возвращает ВСЕ пользователей. Можно расширить:

```typescript
// Возможные параметры в будущем
interface GetUsersQuery {
  limit?: number;      // Количество записей
  offset?: number;     // Смещение
  role?: "user" | "admin"; // Фильтр по роли
  search?: string;     // Поиск по email/name
  sortBy?: "createdAt" | "name" | "email"; // Сортировка
  order?: "asc" | "desc"; // Порядок сортировки
}
```

---

## 2️⃣ GET /api/admin/stats

**Описание:** Получить статистику системы (количество пользователей и заказов)

**Метод:** GET  
**Требует:** JWT токен + роль admin  
**Возвращает:** Объект с статистикой

### Request

```bash
curl -X GET https://api.example.com/api/admin/stats \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```

### Response (200 OK)

```json
{
  "totalUsers": 156,
  "totalOrders": 2340
}
```

### Data Structure

```typescript
interface AdminStats {
  totalUsers: number;   // Количество всех пользователей в системе
  totalOrders: number;  // Количество всех заказов в системе
}
```

### Возможные расширения

```typescript
interface AdminStatsExtended {
  totalUsers: number;
  activeUsers: number;           // Вошедшие за последние 30 дней
  totalOrders: number;
  completedOrders: number;       // Завершённые заказы
  pendingOrders: number;         // Заказы в обработке
  totalRevenue: number;          // Общая выручка
  averageOrderValue: number;     // Средняя стоимость заказа
  adminCount: number;            // Количество админов
  lastUpdated: string;           // Когда обновлены данные
}
```

---

## 3️⃣ PUT /api/admin/users/{id}

**Описание:** Обновить данные пользователя

**Метод:** PUT  
**Требует:** JWT токен + роль admin  
**URL параметры:** `id` - UUID пользователя  
**Возвращает:** Обновлённый объект User

### Request

```bash
curl -X PUT https://api.example.com/api/admin/users/user-id-123 \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Updated Name",
    "email": "newemail@example.com"
  }'
```

### Request Body

```typescript
interface UpdateUserRequest {
  name?: string;     // Новое имя (опционально)
  email?: string;    // Новый email (опционально)
}
```

### Response (200 OK)

```json
{
  "id": "user-id-123",
  "email": "newemail@example.com",
  "name": "Updated Name",
  "password": "$2a$10$...",
  "role": "user",
  "createdAt": "2024-10-15T14:30:00Z"
}
```

### Possible Errors

- **400 Bad Request** - Неверный формат запроса
- **401 Unauthorized** - Нет токена или истёк
- **403 Forbidden** - Недостаточно прав (не админ)
- **404 Not Found** - Пользователь не найден
- **500 Internal Server Error** - Ошибка сервера

---

## 4️⃣ DELETE /api/admin/users/{id}

**Описание:** Удалить пользователя из системы

**Метод:** DELETE  
**Требует:** JWT токен + роль admin  
**URL параметры:** `id` - UUID пользователя  
**Возвращает:** Сообщение об успехе

### Request

```bash
curl -X DELETE https://api.example.com/api/admin/users/user-id-123 \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```

### Response (200 OK)

```json
{
  "message": "User deleted successfully"
}
```

### ⚠️ Важно

- Это **необратимая операция** - данные удаляются полностью
- Удаляются все связанные данные (профиль, заказы, история)
- Рекомендуется использовать мягкое удаление (soft delete) в production

---

## 5️⃣ PATCH /api/admin/users/update-role

**Описание:** Изменить роль пользователя (user ↔ admin)

**Метод:** PATCH  
**Требует:** JWT токен + роль admin  
**Возвращает:** Сообщение об успехе

### Request

```bash
curl -X PATCH https://api.example.com/api/admin/users/update-role \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user-id-123",
    "role": "admin"
  }'
```

### Request Body

```typescript
interface UpdateRoleRequest {
  user_id: string;         // UUID пользователя
  role: "user" | "admin";  // Новая роль
}
```

### Response (200 OK)

```json
{
  "message": "Role updated successfully"
}
```

### Possible Errors

- **400 Bad Request** - Неверная роль (допустимы только "user" и "admin")
- **404 Not Found** - Пользователь не найден
- **401 Unauthorized** - Нет токена
- **403 Forbidden** - Не админ

---

## 6️⃣ GET /api/admin/orders

**Описание:** Получить все заказы системы, отсортированные по дате (новые в начале)

**Метод:** GET  
**Требует:** JWT токен + роль admin  
**Возвращает:** Массив объектов Order

### Request

```bash
curl -X GET https://api.example.com/api/admin/orders \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```

### Response (200 OK)

```json
[
  {
    "id": "order-id-001",
    "userId": "user-id-123",
    "status": "completed",
    "totalPrice": 49.99,
    "itemsCount": 5,
    "createdAt": "2024-11-10T15:30:00Z",
    "updatedAt": "2024-11-10T16:45:00Z"
  },
  {
    "id": "order-id-002",
    "userId": "user-id-456",
    "status": "pending",
    "totalPrice": 125.50,
    "itemsCount": 3,
    "createdAt": "2024-11-09T10:15:00Z",
    "updatedAt": "2024-11-09T10:15:00Z"
  }
]
```

### Data Structure

```typescript
interface Order {
  id: string;              // UUID заказа
  userId: string;          // UUID пользователя, создавшего заказ
  status: OrderStatus;     // Статус заказа
  totalPrice: number;      // Общая сумма
  itemsCount: number;      // Количество товаров
  createdAt: string;       // ISO 8601 дата создания
  updatedAt: string;       // ISO 8601 дата последнего обновления
}

type OrderStatus = 
  | "pending"      // В ожидании
  | "processing"   // В обработке
  | "shipped"      // Отправлен
  | "delivered"    // Доставлен
  | "completed"    // Завершён
  | "cancelled";   // Отменён
```

---

## 7️⃣ GET /api/admin/orders/recent

**Описание:** Получить последние 10 заказов (самые новые)

**Метод:** GET  
**Требует:** JWT токен + роль admin  
**Возвращает:** Массив из максимум 10 заказов

### Request

```bash
curl -X GET https://api.example.com/api/admin/orders/recent \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```

### Response (200 OK)

```json
[
  {
    "id": "order-id-001",
    "userId": "user-id-123",
    "status": "completed",
    "totalPrice": 49.99,
    "itemsCount": 5,
    "createdAt": "2024-11-10T15:30:00Z",
    "updatedAt": "2024-11-10T16:45:00Z"
  },
  {
    "id": "order-id-002",
    "userId": "user-id-456",
    "status": "pending",
    "totalPrice": 125.50,
    "itemsCount": 3,
    "createdAt": "2024-11-09T10:15:00Z",
    "updatedAt": "2024-11-09T10:15:00Z"
  },
  // ... до 10 записей
]
```

**Отличие от /api/admin/orders:**
- Возвращает ТОЛЬКО последние 10 заказов
- Быстрее для получения свежих данных
- Идеально для dashboard

---

## 8️⃣ PUT /api/admin/orders/{id}/status

**Описание:** Изменить статус заказа

**Метод:** PUT  
**Требует:** JWT токен + роль admin  
**URL параметры:** `id` - UUID заказа  
**Возвращает:** Сообщение об успехе

### Request

```bash
curl -X PUT https://api.example.com/api/admin/orders/order-id-001/status \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "status": "shipped"
  }'
```

### Request Body

```typescript
interface UpdateOrderStatusRequest {
  status: "pending" | "processing" | "shipped" | "delivered" | "completed" | "cancelled";
}
```

### Response (200 OK)

```json
{
  "message": "Order status updated"
}
```

### Workflow Пример

```
Создание заказа
      ↓
pending → processing → shipped → delivered → completed
      ↓
      cancelled (если отменить)
```

---

## 9️⃣ GET /api/admin/profile

**Описание:** Получить профиль текущего администратора с управляемыми ресурсами

**Метод:** GET  
**Требует:** JWT токен + роль admin  
**Возвращает:** Объект AdminProfile

### Request

```bash
curl -X GET https://api.example.com/api/admin/profile \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```

### Response (200 OK)

```json
{
  "id": "7ec8aba4-8195-4be1-a9a8-067c30aae306",
  "name": "System Administrator",
  "email": "admin@example.com",
  "role": "admin",
  "createdAt": "2024-01-01T00:00:00Z",
  "managedUsers": 156,
  "managedOrders": 2340,
  "totalStats": {
    "users": 156,
    "orders": 2340
  }
}
```

### Data Structure

```typescript
interface AdminProfile {
  id: string;                    // UUID администратора
  name: string;                  // Имя администратора
  email: string;                 // Email администратора
  role: "admin";                 // Всегда "admin"
  createdAt: string;             // ISO 8601 дата создания аккаунта
  managedUsers: number;          // Количество управляемых пользователей
  managedOrders: number;         // Количество управляемых заказов
  totalStats: {
    users: number;               // Всего пользователей
    orders: number;              // Всего заказов
  };
}
```

---

## 📊 Summary Table

| # | Endpoint | Method | Возвращает | Фильтры |
|----|----------|--------|-----------|---------|
| 1 | `/api/admin/users` | GET | Все пользователи | Нет |
| 2 | `/api/admin/stats` | GET | Статистика | Нет |
| 3 | `/api/admin/users/{id}` | PUT | Обновлённый пользователь | По ID |
| 4 | `/api/admin/users/{id}` | DELETE | Сообщение | По ID |
| 5 | `/api/admin/users/update-role` | PATCH | Сообщение | По user_id |
| 6 | `/api/admin/orders` | GET | Все заказы | DESC по дате |
| 7 | `/api/admin/orders/recent` | GET | 10 последних заказов | DESC по дате |
| 8 | `/api/admin/orders/{id}/status` | PUT | Сообщение | По ID |
| 9 | `/api/admin/profile` | GET | Профиль админа | По JWT |

---

## 🔐 Authorization Matrix

```
┌─────────────────────────────┬────────┬──────────┐
│ Endpoint                    │ User   │ Admin    │
├─────────────────────────────┼────────┼──────────┤
│ /api/admin/users            │ ❌     │ ✅       │
│ /api/admin/stats            │ ❌     │ ✅       │
│ /api/admin/users/{id}       │ ❌     │ ✅       │
│ /api/admin/users/update-role│ ❌     │ ✅       │
│ /api/admin/orders           │ ❌     │ ✅       │
│ /api/admin/orders/recent    │ ❌     │ ✅       │
│ /api/admin/orders/{id}/status│ ❌     │ ✅       │
│ /api/admin/profile          │ ❌     │ ✅       │
│ /api/user/profile           │ ✅     │ ✅       │
└─────────────────────────────┴────────┴──────────┘
```

---

## 💻 Frontend Integration Example

```typescript
// adminService.ts

class AdminService {
  private baseURL = process.env.REACT_APP_API_URL;
  private token = localStorage.getItem("token");

  private getHeaders() {
    return {
      "Authorization": `Bearer ${this.token}`,
      "Content-Type": "application/json"
    };
  }

  // Получить всех пользователей
  async getAllUsers(): Promise<User[]> {
    const res = await fetch(`${this.baseURL}/api/admin/users`, {
      headers: this.getHeaders()
    });
    return res.json();
  }

  // Получить статистику
  async getStats(): Promise<AdminStats> {
    const res = await fetch(`${this.baseURL}/api/admin/stats`, {
      headers: this.getHeaders()
    });
    return res.json();
  }

  // Обновить пользователя
  async updateUser(userId: string, data: UpdateUserRequest) {
    const res = await fetch(`${this.baseURL}/api/admin/users/${userId}`, {
      method: "PUT",
      headers: this.getHeaders(),
      body: JSON.stringify(data)
    });
    return res.json();
  }

  // Удалить пользователя
  async deleteUser(userId: string) {
    const res = await fetch(`${this.baseURL}/api/admin/users/${userId}`, {
      method: "DELETE",
      headers: this.getHeaders()
    });
    return res.json();
  }

  // Изменить роль
  async updateUserRole(userId: string, role: "user" | "admin") {
    const res = await fetch(`${this.baseURL}/api/admin/users/update-role`, {
      method: "PATCH",
      headers: this.getHeaders(),
      body: JSON.stringify({ user_id: userId, role })
    });
    return res.json();
  }

  // Получить все заказы
  async getAllOrders(): Promise<Order[]> {
    const res = await fetch(`${this.baseURL}/api/admin/orders`, {
      headers: this.getHeaders()
    });
    return res.json();
  }

  // Получить последние заказы
  async getRecentOrders(): Promise<Order[]> {
    const res = await fetch(`${this.baseURL}/api/admin/orders/recent`, {
      headers: this.getHeaders()
    });
    return res.json();
  }

  // Обновить статус заказа
  async updateOrderStatus(orderId: string, status: string) {
    const res = await fetch(`${this.baseURL}/api/admin/orders/${orderId}/status`, {
      method: "PUT",
      headers: this.getHeaders(),
      body: JSON.stringify({ status })
    });
    return res.json();
  }

  // Получить профиль админа
  async getAdminProfile(): Promise<AdminProfile> {
    const res = await fetch(`${this.baseURL}/api/admin/profile`, {
      headers: this.getHeaders()
    });
    return res.json();
  }
}

export default new AdminService();
```

---

## ✅ Testing with cURL

```bash
# Установка переменной токена
TOKEN="your_jwt_token_here"

# 1. Получить всех пользователей
curl -X GET http://localhost:8080/api/admin/users \
  -H "Authorization: Bearer $TOKEN"

# 2. Получить статистику
curl -X GET http://localhost:8080/api/admin/stats \
  -H "Authorization: Bearer $TOKEN"

# 3. Обновить пользователя
curl -X PUT http://localhost:8080/api/admin/users/user-id \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name": "New Name"}'

# 4. Удалить пользователя
curl -X DELETE http://localhost:8080/api/admin/users/user-id \
  -H "Authorization: Bearer $TOKEN"

# 5. Изменить роль
curl -X PATCH http://localhost:8080/api/admin/users/update-role \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"user_id": "user-id", "role": "admin"}'

# 6. Получить все заказы
curl -X GET http://localhost:8080/api/admin/orders \
  -H "Authorization: Bearer $TOKEN"

# 7. Получить последние заказы
curl -X GET http://localhost:8080/api/admin/orders/recent \
  -H "Authorization: Bearer $TOKEN"

# 8. Обновить статус заказа
curl -X PUT http://localhost:8080/api/admin/orders/order-id/status \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"status": "shipped"}'

# 9. Получить профиль админа
curl -X GET http://localhost:8080/api/admin/profile \
  -H "Authorization: Bearer $TOKEN"
```

---

## 📋 Notes

- Все эндпоинты требуют валидный JWT токен
- Все эндпоинты (кроме /api/user/profile) требуют роль `admin`
- Все даты возвращаются в ISO 8601 формате
- Пароли всегда хешированы (bcrypt) и безопасны
- Пустые поля могут быть `null` или опущены

