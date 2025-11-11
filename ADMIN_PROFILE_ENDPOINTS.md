# Admin & User Profile Endpoints 👤

Разделённые эндпоинты профиля для администраторов и обычных пользователей.

## Overview

| Endpoint | Метод | Роль | Назначение |
|----------|-------|------|-----------|
| `/api/user/profile` | GET | `user`, `admin` | Получить профиль обычного пользователя |
| `/api/admin/profile` | GET | `admin` | Получить профиль администратора с управляемыми ресурсами |

---

## 1️⃣ GET /api/user/profile

**Описание:** Получить профиль обычного пользователя с его статистикой

**Требует:** JWT токен с любой ролью

**Доступно:** Всем авторизованным пользователям (user, admin)

### Request

```bash
curl -X GET https://api.example.com/api/user/profile \
  -H "Authorization: Bearer eyJhbGc..."
```

### Response (200 OK)

```json
{
  "id": "7ec8aba4-8195-4be1-a9a8-067c30aae306",
  "name": "John Doe",
  "email": "john@example.com",
  "role": "user",
  "level": 5,
  "stars": 120,
  "xp": 5500,
  "walletBalance": 1250,
  "achievements": 15,
  "coursesCompleted": 8,
  "createdAt": "2024-10-15T14:30:00Z"
}
```

### Errors

- **401 Unauthorized** - Отсутствует или невалидный токен
- **500 Internal Server Error** - Ошибка при получении профиля

---

## 2️⃣ GET /api/admin/profile

**Описание:** Получить профиль администратора с информацией об управляемых ресурсах

**Требует:** JWT токен с ролью `admin`

**Доступно:** Только администраторам

### Request

```bash
curl -X GET https://api.example.com/api/admin/profile \
  -H "Authorization: Bearer eyJhbGc..."
```

### Response (200 OK)

```json
{
  "id": "test-admin-id",
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

### Errors

- **401 Unauthorized** - Отсутствует или невалидный токен
- **403 Forbidden** - Пользователь не является администратором
- **404 Not Found** - Администратор не найден
- **500 Internal Server Error** - Ошибка сервера

---

## 📱 Frontend Integration

### checkAuth() функция для выбора нужного профиля

```typescript
async function checkAuth() {
  const token = localStorage.getItem("token");
  const role = localStorage.getItem("role");

  if (!token) {
    redirectToLogin();
    return;
  }

  // Выбираем эндпоинт в зависимости от роли
  const profileUrl = role === "admin" 
    ? "/api/admin/profile" 
    : "/api/user/profile";

  try {
    const res = await fetch(profileUrl, {
      method: "GET",
      headers: {
        "Authorization": `Bearer ${token}`,
        "Content-Type": "application/json"
      }
    });

    if (!res.ok) {
      if (res.status === 401) {
        redirectToLogin();
        return;
      }
      throw new Error(`Failed to fetch profile: ${res.status}`);
    }

    const profileData = await res.json();
    console.log("Profile data:", profileData);

    // Сохраняем профиль в контексте
    setUser(profileData);

    // Если администратор - переходим в админ панель
    if (role === "admin" && profileData.role === "admin") {
      redirectToAdminDashboard();
    } else {
      redirectToDashboard();
    }
  } catch (error) {
    console.error("Error fetching profile:", error);
    redirectToLogin();
  }
}
```

### UserContext пример

```typescript
// UserContext.tsx
export async function checkAuth() {
  const token = localStorage.getItem("token");
  const role = localStorage.getItem("role");

  if (!token) {
    setIsAuthenticated(false);
    return;
  }

  // Определяем нужный эндпоинт
  const endpoint = role === "admin" 
    ? "/api/admin/profile" 
    : "/api/user/profile";

  try {
    const response = await fetch(endpoint, {
      headers: { Authorization: `Bearer ${token}` }
    });

    if (!response.ok) {
      // Если не удалось получить профиль, используем JWT данные
      const decoded = jwtDecode(token);
      setUser({
        id: decoded.userId,
        email: decoded.email,
        role: decoded.role,
        name: "User"
      });
      return;
    }

    const profileData = await response.json();
    setUser(profileData);
  } catch (error) {
    console.error("Error fetching profile:", error);
    setIsAuthenticated(false);
  }
}
```

---

## 🔐 Authorization Matrix

| Эндпоинт | user | admin | Аноним |
|----------|------|-------|---------|
| GET /api/user/profile | ✅ | ✅ | ❌ |
| GET /api/admin/profile | ❌ | ✅ | ❌ |
| GET /api/admin/stats | ❌ | ✅ | ❌ |
| GET /api/admin/users | ❌ | ✅ | ❌ |

---

## 💾 Backend Architecture

### Service Layer

```go
// internal/modules/admin/service/service.go

type AdminService interface {
    // ... other methods
    GetAdminProfile(adminID string) (map[string]interface{}, error)
}

func (s *adminService) GetAdminProfile(adminID string) (map[string]interface{}, error) {
    // Получить данные админа из БД
    // Подсчитать управляемые ресурсы
    // Вернуть полный профиль
}
```

### Handler Layer

```go
// internal/modules/admin/transport/http/handlers.go

func (h *AdminHandlers) GetAdminProfile(w http.ResponseWriter, r *http.Request) {
    adminID := r.Context().Value("user_id").(string)
    profile, err := h.service.GetAdminProfile(adminID)
    // ... error handling ...
    utils.RespondWithJSON(w, http.StatusOK, profile)
}
```

### Routes

```go
// internal/modules/admin/module.go

r.Route("/admin", func(r chi.Router) {
    r.Use(authMiddleware)
    r.Use(adminMiddleware)
    
    r.Get("/profile", m.handlers.GetAdminProfile)
    r.Get("/stats", m.handlers.GetAdminStats)
    // ... other routes ...
})
```

---

## 🧪 Testing

### Test Admin Profile with JWT

```bash
# Генерируем JWT токен для админа
TOKEN=$(curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"admin_password_123"}' \
  | jq -r '.token')

# Получаем профиль админа
curl -X GET http://localhost:8080/api/admin/profile \
  -H "Authorization: Bearer $TOKEN" \
  | jq '.'
```

### Test User Profile with JWT

```bash
# Генерируем JWT токен для обычного пользователя
TOKEN=$(curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password123"}' \
  | jq -r '.token')

# Получаем профиль пользователя
curl -X GET http://localhost:8080/api/user/profile \
  -H "Authorization: Bearer $TOKEN" \
  | jq '.'
```

---

## 📊 Data Models

### User Profile Response

```typescript
interface UserProfile {
  id: string;
  name: string;
  email: string;
  role: "user" | "admin";
  level: number;
  stars: number;
  xp: number;
  walletBalance: number;
  achievements: number;
  coursesCompleted: number;
  createdAt: string;
}
```

### Admin Profile Response

```typescript
interface AdminProfile {
  id: string;
  name: string;
  email: string;
  role: "admin";
  createdAt: string;
  managedUsers: number;
  managedOrders: number;
  totalStats: {
    users: number;
    orders: number;
  };
}
```

---

## ✅ Summary

- ✅ `/api/user/profile` - для всех авторизованных пользователей
- ✅ `/api/admin/profile` - только для администраторов
- ✅ Оба эндпоинта требуют JWT токен
- ✅ Frontend выбирает нужный эндпоинт в зависимости от роли
- ✅ Архитектура следует Clean Architecture паттерну
- ✅ Защищены middleware и role-based access control
