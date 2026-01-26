# ✅ Ролевая модель 2026 — ИСТОЧНИК ИСТИНЫ

**Дата:** 2026-01-26  
**Статус:** ✅ Реализовано и задокументировано

---

## 🎯 Принципы

### Правило 2026 (КРИТИЧНО)

1. ❌ **Пользователь НИКОГДА сам не выбирает роль**
2. ✅ **Роль назначает ТОЛЬКО admin / super_admin**
3. ✅ **Backend — единственный источник истины**

---

## 👥 Роли (Enum)

```go
// internal/models/user.go

type UserRole string

const (
    RoleCustomer   = "customer"    // Покупатель (обычный пользователь)
    RoleHomeChef   = "home_chef"   // Домашний повар
    RoleChefStaff  = "chef_staff"  // Младший повар / персонал
    RoleAdmin      = "admin"       // Администратор
    RoleSuperAdmin = "super_admin" // Супер администратор (владелец системы)
)
```

### Описание ролей

| Роль | Константа | Описание | Назначение |
|------|-----------|----------|------------|
| **Customer** | `customer` | Покупатель, обычный пользователь | По умолчанию при регистрации |
| **Home Chef** | `home_chef` | Домашний повар | Админ может назначить |
| **Chef Staff** | `chef_staff` | Младший повар / персонал ресторана | Админ может назначить |
| **Admin** | `admin` | Администратор системы | Super admin может назначить |
| **Super Admin** | `super_admin` | Владелец системы (полный доступ) | Только вручную в БД |

---

## 📊 Статусы (Enum)

```go
// internal/models/user.go

type UserStatus string

const (
    UserStatusPending   = "pending"   // Зарегистрирован, но не активирован
    UserStatusActive    = "active"    // Активен - может использовать все функции
    UserStatusSuspended = "suspended" // Временно отключён
    UserStatusBlocked   = "blocked"   // Заблокирован администратором
)
```

### Описание статусов

| Статус | Константа | Описание | Может логиниться | Может использовать API |
|--------|-----------|----------|------------------|------------------------|
| **Pending** | `pending` | Зарегистрирован, ожидает активации | ❌ | ❌ |
| **Active** | `active` | Активен | ✅ | ✅ |
| **Suspended** | `suspended` | Временно отключён | ❌ | ❌ |
| **Blocked** | `blocked` | Заблокирован навсегда | ❌ | ❌ |

---

## 🔐 Регистрация

### POST /api/auth/register

**Правило:** При регистрации всегда назначается роль `customer` и статус `active`.

```go
// internal/modules/auth/service/service.go

user := &models.User{
    ID:        uuid.New().String(),
    Email:     req.Email,
    Name:      req.Name,
    Password:  string(hashedPassword),
    Role:      models.RoleCustomer,      // ✅ Всегда customer
    Status:    models.UserStatusActive,   // ✅ Активен сразу
    CreatedAt: time.Now(),
}
```

**Примечание:** Если нужна email-активация, установите `Status: models.UserStatusPending`.

---

## 🔑 JWT Claims (минимальный и чистый)

```json
{
  "sub": "407582be-59d5-4d21-873b-1a72d31b0d42",
  "email": "user@example.com",
  "role": "customer",
  "exp": 1738080000,
  "iat": 1738000000
}
```

### ✅ Что ЕСТЬ в JWT:
- `sub` — User ID (UUID)
- `email` — Email пользователя
- `role` — Роль пользователя
- `exp` — Срок действия токена
- `iat` — Время создания токена

### ❌ Чего НЕТ в JWT (и не должно быть):
- ❌ `hasRole` — удалено
- ❌ `mode` — не существует
- ❌ `permissions` — не существует
- ❌ UI флаги — не существует

---

## 🛡️ AuthMiddleware — Проверка статуса

```go
// internal/middleware/auth.go

// Проверяем статус пользователя
if user.Status != models.UserStatusActive {
    // pending / suspended / blocked → 403 Forbidden
    utils.WriteError(w, http.StatusForbidden, "Account is not active")
    return
}
```

### Поведение по статусам:

| Статус | HTTP Code | Сообщение | Действие на фронте |
|--------|-----------|-----------|-------------------|
| `pending` | 403 | "Account is pending activation" | Показать "Активируйте email" |
| `active` | ✅ 200 | — | Пропустить запрос |
| `suspended` | 403 | "Account is temporarily suspended" | Показать "Аккаунт приостановлен" |
| `blocked` | 403 | "Account is blocked" | Показать "Аккаунт заблокирован" |

---

## 🔧 Назначение ролей (ТОЛЬКО админка)

### PATCH /api/admin/users/update-role

**Требования:**
- Авторизация: `Bearer <token>`
- Роль вызывающего: `super_admin` (только супер админ может менять роли)

**Request:**
```json
{
  "user_id": "407582be-59d5-4d21-873b-1a72d31b0d42",
  "role": "home_chef"
}
```

**Response (200 OK):**
```json
{
  "message": "Role updated successfully"
}
```

**Валидация:**
```go
// internal/modules/admin/service/service.go

validRoles := map[string]bool{
    models.RoleCustomer:   true,
    models.RoleHomeChef:   true,
    models.RoleChefStaff:  true,
    models.RoleAdmin:      true,
    models.RoleSuperAdmin: true,
}

if !validRoles[role] {
    return errors.New("invalid role")
}
```

---

## 📦 Структура БД (users)

```sql
CREATE TABLE "User" (
    id           TEXT PRIMARY KEY,
    email        TEXT UNIQUE NOT NULL,
    name         TEXT NOT NULL,
    password     TEXT NOT NULL,
    role         TEXT NOT NULL DEFAULT 'customer',
    status       TEXT NOT NULL DEFAULT 'active',
    settings     JSONB,
    last_login   TIMESTAMP,
    "createdAt"  TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Индексы для производительности
CREATE INDEX idx_user_role ON "User"(role);
CREATE INDEX idx_user_status ON "User"(status);
CREATE INDEX idx_user_email ON "User"(email);
```

### ✅ Что ЕСТЬ в таблице:
- `id` — UUID пользователя
- `email` — Email
- `name` — Имя
- `password` — Хеш пароля
- `role` — Роль (enum)
- `status` — Статус (enum)
- `settings` — Настройки (JSONB)
- `last_login` — Последний вход
- `createdAt` — Дата создания

### ❌ Чего НЕТ (и не должно быть):
- ❌ `isChef` — не существует
- ❌ `mode` — не существует
- ❌ `accountType` — не существует
- ❌ `permissions` — не существует
- ❌ `selectedRole` — не существует

---

## 🔄 Миграция существующих пользователей

### SQL для обновления ролей:

```sql
-- Если у вас есть старые пользователи с role = 'pro_chef'
UPDATE "User" 
SET role = 'home_chef' 
WHERE role = 'pro_chef';

-- Если нужно установить customer для пользователей без роли
UPDATE "User" 
SET role = 'customer' 
WHERE role IS NULL OR role = '';

-- Если нужно установить active для пользователей без статуса
UPDATE "User" 
SET status = 'active' 
WHERE status IS NULL OR status = '';
```

---

## 🎨 Фронтенд интеграция

### Проверка роли

```typescript
// ✅ ПРАВИЛЬНО: Backend решает роль
const user = await fetch('/api/auth/me');
const { role, status } = user.data;

if (status !== 'active') {
  // Показать сообщение о статусе
  return <AccountStatusPage status={status} />;
}

switch (role) {
  case 'customer':
    return <CustomerDashboard />;
  case 'home_chef':
    return <ChefDashboard />;
  case 'admin':
    return <AdminDashboard />;
  default:
    return <DefaultDashboard />;
}
```

### ❌ НЕПРАВИЛЬНО: Фронт не решает роль

```typescript
// ❌ НЕПРАВИЛЬНО
const [role, setRole] = useState('customer');

// Пользователь выбирает роль
<select onChange={(e) => setRole(e.target.value)}>
  <option value="customer">Customer</option>
  <option value="chef">Chef</option>
</select>
```

---

## 📋 Чеклист реализации

### Backend
- [x] Роли определены как enum константы
- [x] Статусы определены как enum константы
- [x] Регистрация назначает `customer` роль
- [x] JWT содержит только необходимые поля
- [x] AuthMiddleware проверяет статус
- [x] Эндпоинт назначения ролей только для super_admin
- [x] Валидация ролей при назначении
- [x] БД содержит только необходимые поля

### Фронтенд
- [ ] Фронт получает роль из `/api/auth/me`
- [ ] Фронт не позволяет выбирать роль
- [ ] Обработка статусов (pending, suspended, blocked)
- [ ] Редирект на соответствующий dashboard по роли
- [ ] Админка для назначения ролей

---

## 🚀 Примеры использования

### Проверка роли в коде

```go
// Проверка конкретной роли
if user.Role == models.RoleAdmin {
    // Админские действия
}

// Проверка нескольких ролей
allowedRoles := []string{models.RoleAdmin, models.RoleSuperAdmin}
if contains(allowedRoles, user.Role) {
    // Действие
}

// Middleware для проверки роли
r.With(middleware.RequireRole(models.RoleHomeChef)).Get("/chef-only", handler)
```

### Проверка статуса

```go
// Проверка активности
if user.Status != models.UserStatusActive {
    return errors.New("user is not active")
}

// Проверка конкретного статуса
if user.Status == models.UserStatusBlocked {
    return errors.New("user is blocked")
}
```

---

## 📝 Лог изменений

### 2026-01-26

#### Добавлено:
- ✅ Новые роли: `customer`, `chef_staff`
- ✅ Новый статус: `suspended`
- ✅ Проверка статуса в AuthMiddleware
- ✅ Эндпоинт назначения ролей
- ✅ Валидация ролей

#### Изменено:
- ✅ Роль по умолчанию: `home_chef` → `customer`
- ✅ Удалена роль: `pro_chef`
- ✅ JWT: удалено поле `hasRole`

#### Удалено:
- ✅ Возможность пользователя выбирать роль
- ✅ UI флаги в JWT
- ✅ Лишние поля в User модели

---

**Статус:** ✅ Реализовано  
**Источник истины:** Backend  
**Документация:** Актуальна
