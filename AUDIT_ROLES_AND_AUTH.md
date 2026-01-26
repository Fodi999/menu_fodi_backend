# 🔍 АУДИТ СИСТЕМЫ РОЛЕЙ И АУТЕНТИФИКАЦИИ

**Дата:** 2026-01-26  
**Статус:** ✅ Аудит завершен

---

## 📋 1. МОДЕЛЬ ПОЛЬЗОВАТЕЛЯ (`/internal/models/user.go`)

### ✅ **Есть ли `role`?**
**ДА** ✅
```go
Role string `gorm:"column:role;default:home_chef" json:"role"`
```

### ✅ **Есть ли `status`?**
**ДА** ✅
```go
Status string `gorm:"column:status;default:active" json:"status"`
```

### ✅ **Константы ролей:**
```go
const (
    RoleHomeChef   = "home_chef"   // Домашний повар
    RoleProChef    = "pro_chef"    // Профессиональный повар / ресторан
    RoleAdmin      = "admin"       // Администратор
    RoleSuperAdmin = "super_admin" // Супер администратор
)
```

### ✅ **Константы статусов:**
```go
const (
    UserStatusActive  = "active"  // Normal user - can login and use all features
    UserStatusBlocked = "blocked" // Blocked by admin - cannot login
    UserStatusPending = "pending" // Unverified / limited access
)
```

### ❌ **Лишние поля:**
**НЕТ** ✅
- ❌ `isChef` - **НЕТ в модели**
- ❌ `mode` - **НЕТ в модели**
- ❌ `accountType` - **НЕТ в модели**
- ❌ `permissions` - **НЕТ в модели**
- ❌ `selectedRole` - **НЕТ в модели**

**ВЫВОД:** Модель чистая, без лишних полей ✅

---

## 🔐 2. JWT CLAIMS (`/internal/modules/auth/service/jwt_service.go`)

### ✅ **Структура Claims:**
```go
type Claims struct {
    Email   string `json:"email"`
    Role    string `json:"role"`
    HasRole bool   `json:"hasRole"`  // ⚠️ Лишнее поле (можно убрать)
    jwt.RegisteredClaims  // Contains Subject (sub), ExpiresAt (exp), IssuedAt (iat)
}
```

### ✅ **Что кладется в JWT при логине:**
```go
// internal/modules/auth/service/service.go:120
token, err := GenerateToken(user.ID, user.Email, user.Role)
```

**В JWT кладется:**
- ✅ `sub` (Subject) = `user.ID`
- ✅ `email` = `user.Email`
- ✅ `role` = `user.Role`
- ✅ `exp` (ExpiresAt) = 24 часа
- ✅ `iat` (IssuedAt) = текущее время
- ⚠️ `hasRole` = `true` (лишнее поле, можно убрать)

### ❌ **Есть ли `mode` в JWT?**
**НЕТ** ✅ - `mode` отсутствует в Claims

### ✅ **Есть ли `role` в JWT?**
**ДА** ✅ - `role` присутствует в Claims

**ВЫВОД:** JWT чистый, без UI-данных ✅ (кроме `hasRole`, которое можно убрать)

---

## 🛡️ 3. MIDDLEWARE (`/internal/middleware/auth.go`)

### ❌ **Проверяется ли `status`?**
**НЕТ** ❌ - **КРИТИЧЕСКАЯ ПРОБЛЕМА**

**Текущий код:**
```go
func AuthMiddleware(next http.Handler) http.Handler {
    // ... проверка JWT ...
    // ❌ НЕТ проверки user.Status
    next.ServeHTTP(w, r.WithContext(ctx))
}
```

**Проблема:** Пользователь с `status = "blocked"` может использовать API, если у него валидный JWT.

**Решение:** Нужно добавить проверку статуса после валидации JWT.

### ✅ **Проверяется ли `role`?**
**ДА** ✅ - Проверка роли есть в:
- `AdminMiddleware` - проверяет `admin` или `super_admin`
- `SuperAdminMiddleware` - проверяет только `super_admin`
- `RequireRole(role)` - проверяет конкретную роль

**ВЫВОД:** 
- ✅ Проверка роли работает
- ❌ **КРИТИЧНО:** Нет проверки статуса в `AuthMiddleware`

---

## 🛣️ 4. МАРШРУТЫ (`/internal/app/routes_modular.go`)

### ❌ **Есть ли разделение по зонам `/customer`, `/chef`, `/admin`?**
**НЕТ** ❌ - Все маршруты под `/api/*`

**Структура маршрутов:**
```
/api/auth/*          - Аутентификация (публичные)
/api/user/*          - Профиль пользователя
/api/admin/*         - Админ-панель (с AdminMiddleware)
/api/fridge/*        - Холодильник
/api/recipes/*       - Рецепты
/api/ingredients/*   - Ингредиенты
...
```

**НЕТ разделения:**
- ❌ `/customer/*` - нет
- ❌ `/chef/*` - нет
- ✅ `/admin/*` - есть (с middleware)

**ВЫВОД:** Архитектура единая для всех ролей, разделение через middleware ✅

---

## 🔗 5. FRONT-BACK КОНТРАКТ

### ❌ **Фронт вызывает `/customer/profile` - существует ли?**
**НЕТ** ❌

**Реальный эндпоинт:**
```
GET /api/user/profile  ✅ (существует)
```

**Админский профиль:**
```
GET /api/admin/profile  ✅ (существует)
```

**Проблема:** Фронт ожидает `/customer/profile`, но бекенд предоставляет `/api/user/profile`.

**ВЫВОД:** Несоответствие между фронтом и бекендом ❌

---

## 📊 6. ИСТОРИЯ (`/internal/models/history_event.go`)

### ✅ **Какие события логируются:**
```go
const (
    EventTypeCook         = "cook"
    EventTypeConsume      = "consume"
    EventTypeWaste        = "waste"
    EventTypeManual       = "manual"
    EventTypeFridgeAdd    = "fridge_add"
    EventTypeFridgeRemove = "fridge_remove"
    EventTypeExpired      = "waste"
)
```

### ❌ **Есть ли события для ролей/статусов?**
**НЕТ** ❌
- ❌ `ROLE_CHANGED` - нет
- ❌ `STATUS_CHANGED` - нет
- ❌ `LOGIN` - нет
- ❌ `LOGOUT` - нет

**ВЫВОД:** История не отслеживает изменения ролей/статусов ❌

---

## ✅ 7. ЧЕКЛИСТ

| Вопрос | Ответ | Комментарий |
|--------|-------|-------------|
| В БД нет `mode` | ✅ ДА | Модель чистая |
| Role — enum (константы) | ✅ ДА | Есть константы `RoleHomeChef`, `RoleProChef`, etc. |
| Status — enum (константы) | ✅ ДА | Есть константы `UserStatusActive`, `UserStatusBlocked`, etc. |
| JWT не содержит UI данных | ⚠️ ЧАСТИЧНО | Есть `hasRole` (можно убрать), но нет `mode` |
| Есть `/auth/me` | ✅ ДА | `GET /api/auth/me` |
| Middleware проверяет `status` | ❌ НЕТ | **КРИТИЧНО: нужно добавить** |
| Роли меняет только admin | ✅ ДА | `UpdateUserRole` только для `super_admin` |
| Фронт не решает "кто я" | ❌ НЕТ | Фронт вызывает `/customer/profile` вместо `/api/user/profile` |

---

## 🚨 КРИТИЧЕСКИЕ ПРОБЛЕМЫ

### 1. ❌ **Нет проверки `status` в `AuthMiddleware`**
**Проблема:** Заблокированные пользователи могут использовать API.

**Решение:**
```go
func AuthMiddleware(next http.Handler) http.Handler {
    // ... валидация JWT ...
    
    // ✅ ДОБАВИТЬ: Проверка статуса пользователя
    user, err := userRepo.FindByID(claims.Subject)
    if err != nil || user.Status == models.UserStatusBlocked {
        utils.WriteError(w, http.StatusForbidden, "Account is blocked")
        return
    }
    
    next.ServeHTTP(w, r.WithContext(ctx))
}
```

### 2. ❌ **Несоответствие фронт-бек контракта**
**Проблема:** Фронт вызывает `/customer/profile`, бекенд предоставляет `/api/user/profile`.

**Решение:**
- Вариант 1: Добавить редирект `/customer/profile` → `/api/user/profile`
- Вариант 2: Исправить фронт (предпочтительно)

### 3. ⚠️ **Лишнее поле `hasRole` в JWT Claims**
**Проблема:** Не используется, можно убрать.

**Решение:** Удалить поле `HasRole` из структуры `Claims`.

---

## ✅ ЧТО РАБОТАЕТ ХОРОШО

1. ✅ **Модель пользователя чистая** - нет лишних полей (`mode`, `isChef`, etc.)
2. ✅ **Роли через константы** - избегаем опечаток
3. ✅ **Статусы через константы** - типобезопасность
4. ✅ **JWT чистый** - только необходимые данные (кроме `hasRole`)
5. ✅ **Проверка ролей работает** - `AdminMiddleware`, `SuperAdminMiddleware`, `RequireRole`
6. ✅ **Изменение ролей защищено** - только `super_admin` может менять роли
7. ✅ **Проверка статуса при логине** - заблокированные не могут войти

---

## 📝 РЕКОМЕНДАЦИИ

### Приоритет 1 (КРИТИЧНО):
1. ✅ Добавить проверку `status` в `AuthMiddleware`
2. ✅ Исправить фронт-бек контракт (`/customer/profile` → `/api/user/profile`)

### Приоритет 2 (ВАЖНО):
3. ✅ Убрать поле `hasRole` из JWT Claims
4. ✅ Добавить события в историю: `ROLE_CHANGED`, `STATUS_CHANGED`, `LOGIN`, `LOGOUT`

### Приоритет 3 (УЛУЧШЕНИЯ):
5. ✅ Добавить валидацию ролей при создании/обновлении пользователя
6. ✅ Добавить миграцию для проверки существующих данных в БД

---

## 🎯 ПЛАН ДЕЙСТВИЙ

### Шаг 1: Исправить `AuthMiddleware` (5 минут)
```go
// Добавить проверку статуса после валидации JWT
```

### Шаг 2: Исправить фронт-бек контракт (10 минут)
```go
// Добавить редирект или исправить фронт
```

### Шаг 3: Убрать `hasRole` из JWT (5 минут)
```go
// Удалить поле из Claims
```

### Шаг 4: Добавить события в историю (30 минут)
```go
// Добавить константы и логирование
```

---

**Статус аудита:** ✅ Завершен  
**Следующий шаг:** Исправить критичные проблемы (приоритет 1)
