# ✅ БД — Единственный источник истины (2026)

**Дата:** 2026-01-26  
**Статус:** ✅ Реализовано  
**Принцип:** Database as Single Source of Truth

---

## 🎯 Ключевой принцип

### ❌ НЕПРАВИЛЬНО (старый подход)

```typescript
// ❌ Хранить роль только в JWT
const token = jwt.sign({ 
  id: user.id, 
  role: user.role,  // Роль в токене
  status: user.status 
});

// ❌ Полагаться на role из токена после изменения роли
const role = jwt.decode(token).role; // Устаревшая роль!
```

**Проблемы:**
- Токен остаётся валидным после изменения роли
- Пользователь может иметь устаревшие права
- Нужно принудительно инвалидировать токены
- Риск безопасности

### ✅ ПРАВИЛЬНО (2026)

```go
// ✅ JWT содержит только ID пользователя
token := jwt.Sign(Claims{
    Subject: user.ID,  // Только ID!
    Email:   user.Email,
    Role:    user.Role, // Для удобства, НО не источник истины
})

// ✅ AuthMiddleware ВСЕГДА читает актуальные данные из БД
user := userRepo.GetByID(claims.Subject)
if user.Status != "active" {
    return 403 // Проверка актуального статуса
}

// ✅ /api/auth/me — всегда читает актуальные данные
func GetCurrentUser(userID) {
    return db.FindByID(userID) // Актуальная роль и статус
}
```

**Преимущества:**
- ✅ Изменения роли применяются мгновенно
- ✅ Не нужно инвалидировать токены
- ✅ Актуальные данные всегда из БД
- ✅ Безопасность на уровне middleware

---

## 📋 Реализация в проекте

### 1️⃣ AuthMiddleware — проверка статуса из БД

```go
// internal/middleware/auth.go

func AuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // 1. Валидируем JWT
        token := extractToken(r)
        claims, err := validateToken(token)
        
        // 2. КРИТИЧЕСКИ ВАЖНО: Загружаем пользователя из БД
        userRepo := &database.UserRepository{}
        user, err := userRepo.FindByID(claims.Subject)
        
        // 3. Проверяем актуальный статус из БД (не из JWT!)
        if user.Status != models.UserStatusActive {
            utils.WriteError(w, http.StatusForbidden, "User is not active")
            return
        }
        
        // 4. Кладём актуального пользователя в контекст
        ctx := context.WithValue(r.Context(), middleware.CtxUser, user)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

**Что это даёт:**
- ✅ Заблокированный пользователь не может вызывать API
- ✅ Изменение статуса применяется мгновенно
- ✅ Не нужно ждать истечения токена

### 2️⃣ UpdateUserRole — атомарное изменение роли и статуса

```go
// internal/modules/admin/service/service.go

func (s *adminService) UpdateUserRole(userID, role, adminID string) error {
    // 1️⃣ ЗАПРЕТ НА НАЗНАЧЕНИЕ SUPER_ADMIN ЧЕРЕЗ UI
    if role == models.RoleSuperAdmin {
        return errors.New("super_admin role cannot be assigned via API")
    }
    
    // 2️⃣ Валидация роли
    validRoles := map[string]bool{
        models.RoleCustomer:  true,
        models.RoleHomeChef:  true,
        models.RoleChefStaff: true,
        models.RoleAdmin:     true,
    }
    if !validRoles[role] {
        return errors.New("invalid role")
    }
    
    // 3️⃣ Получаем текущие данные
    var user models.User
    s.db.Where("id = ?", userID).First(&user)
    oldRole := user.Role
    oldStatus := user.Status
    
    // 4️⃣ АВТОМАТИЧЕСКАЯ ЛОГИКА СТАТУСА
    newStatus := models.UserStatusActive
    
    switch role {
    case models.RoleChefStaff:
        // Персонал требует дополнительной проверки
        newStatus = models.UserStatusPending
    case models.RoleAdmin, models.RoleHomeChef, models.RoleCustomer:
        // Остальные роли сразу активны
        newStatus = models.UserStatusActive
    }
    
    // 5️⃣ АТОМАРНОЕ ОБНОВЛЕНИЕ роли и статуса
    updates := map[string]interface{}{
        "role":   role,
        "status": newStatus,
    }
    s.db.Model(&models.User{}).Where("id = ?", userID).Updates(updates)
    
    // 6️⃣ АУДИТ: Логируем изменение
    metadata := map[string]interface{}{
        "old_role":     oldRole,
        "new_role":     role,
        "old_status":   oldStatus,
        "new_status":   newStatus,
        "changed_by":   adminID,
        "auto_status":  oldStatus != newStatus,
    }
    s.history.Log(userID, "role_changed", metadata)
    
    return nil
}
```

**Что это даёт:**
- ✅ Роль и статус меняются атомарно
- ✅ Логика статуса автоматическая (не забыть установить)
- ✅ Полный аудит изменений
- ✅ Запрет на назначение super_admin через UI

### 3️⃣ GetCurrentUser — актуальные данные из БД

```go
// internal/modules/auth/service/service.go

func (s *AuthService) GetCurrentUser(userID uuid.UUID) (*dto.CurrentUserResponse, error) {
    // КРИТИЧЕСКИ ВАЖНО: Читаем из БД, НЕ из JWT!
    user, err := s.repo.FindByID(userID.String())
    if err != nil {
        return nil, ErrUserNotFound
    }
    
    // Возвращаем актуальные данные
    return &dto.CurrentUserResponse{
        ID:     user.ID,
        Email:  user.Email,
        Name:   user.Name,
        Role:   user.Role,   // ✅ Актуальная роль из БД
        Status: user.Status, // ✅ Актуальный статус из БД
    }, nil
}
```

**Что это даёт:**
- ✅ Фронтенд всегда получает актуальные данные
- ✅ Изменения роли видны сразу после вызова `/api/auth/me`
- ✅ Не нужно обновлять токен

---

## 🔄 Жизненный цикл изменения роли

```
1. Super admin меняет роль пользователя
   PATCH /api/admin/users/{id}/role
   { "role": "chef_staff" }
   ↓

2. Backend обновляет БД атомарно
   UPDATE User SET role='chef_staff', status='pending' WHERE id=?
   ↓

3. Backend логирует изменение в history
   INSERT INTO history_events (event_type='role_changed', ...)
   ↓

4. Пользователь продолжает работать со старым токеном
   (токен не инвалидируется)
   ↓

5. При следующем запросе AuthMiddleware проверяет БД
   user := db.FindByID(claims.Subject)
   if user.Status != "active" → 403 Forbidden
   ↓

6. Фронтенд вызывает /api/auth/me
   GET /api/auth/me
   → { role: "chef_staff", status: "pending" }
   ↓

7. Фронтенд обновляет UI
   showMessage("Вы теперь повар. Ожидается подтверждение.")
```

---

## 🔐 Безопасность

### Проверка прав на каждом запросе

```go
// Middleware проверяет актуальный статус
func AuthMiddleware(next http.Handler) http.Handler {
    // ...
    user := db.FindByID(claims.Subject) // ← Актуальные данные из БД
    if user.Status != "active" {
        return 403
    }
    // ...
}

// Middleware проверяет актуальную роль
func RequireRole(role string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            user := getUserFromContext(r) // ← Уже актуальные данные из БД
            if user.Role != role {
                return 403
            }
            next.ServeHTTP(w, r)
        })
    }
}
```

**Что это даёт:**
- ✅ Изменения применяются мгновенно
- ✅ Невозможно использовать устаревшие права
- ✅ Блокировка работает без перелогина

### Запрет на назначение super_admin через UI

```go
if role == models.RoleSuperAdmin {
    logger.Warn("Attempt to assign super_admin role via API (blocked)")
    return errors.New("super_admin role cannot be assigned via API")
}
```

**Почему:**
- ✅ Super admin — критическая роль
- ✅ Назначается только через миграцию БД
- ✅ Невозможно случайно назначить через UI
- ✅ Защита от атак и ошибок

---

## 📊 Автоматическая логика статуса

### Правила

| Роль          | Статус по умолчанию | Причина                           |
|---------------|---------------------|-----------------------------------|
| `customer`    | `active`            | Обычный пользователь              |
| `home_chef`   | `active`            | Домашний повар                    |
| `chef_staff`  | `pending`           | Требует проверки администратором  |
| `admin`       | `active`            | Администратор                     |
| `super_admin` | (нельзя назначить)  | Только через миграцию БД          |

### Реализация

```go
switch role {
case models.RoleChefStaff:
    // Персонал требует дополнительной проверки
    newStatus = models.UserStatusPending
    logger.Info("User role changed to chef_staff, status set to pending")
    
case models.RoleAdmin:
    // Админ должен быть сразу активен
    newStatus = models.UserStatusActive
    
case models.RoleHomeChef, models.RoleCustomer:
    // Обычные роли остаются активными
    newStatus = models.UserStatusActive
}
```

---

## 🧪 Тестирование

### Тест 1: Изменение роли и мгновенное применение

```bash
# 1. Логин пользователя
TOKEN=$(curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password"}' \
  | jq -r '.token')

# 2. Проверка текущей роли
curl -X GET http://localhost:8080/api/auth/me \
  -H "Authorization: Bearer $TOKEN"
# → { "role": "customer", "status": "active" }

# 3. Super admin меняет роль на chef_staff
curl -X PATCH http://localhost:8080/api/admin/users/{id}/role \
  -H "Authorization: Bearer $SUPER_ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"role":"chef_staff"}'

# 4. Пользователь проверяет роль (с ТЕМ ЖЕ токеном!)
curl -X GET http://localhost:8080/api/auth/me \
  -H "Authorization: Bearer $TOKEN"
# → { "role": "chef_staff", "status": "pending" }

# ✅ Роль изменилась мгновенно без перелогина!
# ✅ Статус автоматически установлен в pending
```

### Тест 2: Блокировка пользователя через статус

```bash
# 1. Пользователь работает с токеном
curl -X GET http://localhost:8080/api/recipes \
  -H "Authorization: Bearer $TOKEN"
# → 200 OK

# 2. Admin блокирует пользователя
psql $DATABASE_URL -c "UPDATE \"User\" SET status='blocked' WHERE id='user-id';"

# 3. Пользователь пытается сделать запрос (с ТЕМ ЖЕ токеном!)
curl -X GET http://localhost:8080/api/recipes \
  -H "Authorization: Bearer $TOKEN"
# → 403 Forbidden: "User is not active"

# ✅ Блокировка сработала мгновенно без перелогина!
```

### Тест 3: Попытка назначить super_admin через UI

```bash
curl -X PATCH http://localhost:8080/api/admin/users/{id}/role \
  -H "Authorization: Bearer $SUPER_ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"role":"super_admin"}'

# → 403 Forbidden: "Super admin role cannot be assigned via API"

# ✅ Защита от назначения super_admin через UI работает!
```

---

## 📝 Frontend интеграция

### Правильный подход

```typescript
// ❌ НЕПРАВИЛЬНО: Полагаться на роль из токена
const token = localStorage.getItem('token');
const decoded = jwt.decode(token);
const role = decoded.role; // Может быть устаревшим!

// ✅ ПРАВИЛЬНО: Получать актуальные данные из /api/auth/me
async function getCurrentUser() {
  const response = await fetch('/api/auth/me', {
    headers: {
      'Authorization': `Bearer ${getToken()}`
    }
  });
  return response.json(); // { role, status, ... } из БД
}

// Использование
const user = await getCurrentUser();
if (user.status !== 'active') {
  redirectTo('/account/status');
}
if (user.role === 'chef_staff' && user.status === 'pending') {
  showMessage('Ваш аккаунт ожидает подтверждения');
}
```

### После изменения роли

```typescript
// Super admin изменил роль пользователя
await changeUserRole(userId, 'chef_staff');

// Пользователь обновляет свои данные (не перелогинивается!)
const updatedUser = await getCurrentUser();
console.log(updatedUser);
// { role: 'chef_staff', status: 'pending', ... }

// UI обновляется автоматически
updateUserInterface(updatedUser);
```

---

## ✅ Итог

### Что реализовано

1. ✅ **AuthMiddleware проверяет статус из БД** — изменения применяются мгновенно
2. ✅ **UpdateUserRole атомарно меняет роль и статус** — консистентность данных
3. ✅ **GetCurrentUser читает актуальные данные из БД** — не полагается на JWT
4. ✅ **Запрет на назначение super_admin через UI** — безопасность
5. ✅ **Автоматическая логика статуса** — не забыть установить статус
6. ✅ **Полный аудит изменений** — логирование в history

### Принципы 2026

| Принцип                                  | Статус |
|------------------------------------------|--------|
| БД — единственный источник истины        | ✅ Да  |
| JWT — только идентификатор сессии        | ✅ Да  |
| /api/auth/me — всегда читает из БД       | ✅ Да  |
| Изменения применяются мгновенно          | ✅ Да  |
| Не нужно инвалидировать токены           | ✅ Да  |
| Роль и статус меняются атомарно          | ✅ Да  |
| Super admin нельзя назначить через UI    | ✅ Да  |

### Преимущества

- ✅ **Безопасность**: Актуальные права на каждом запросе
- ✅ **Простота**: Не нужно управлять инвалидацией токенов
- ✅ **Консистентность**: БД — единственный источник истины
- ✅ **UX**: Изменения видны сразу без перелогина
- ✅ **Аудит**: Полная история изменений ролей

---

**Статус:** ✅ Реализовано  
**Принцип:** Database as Single Source of Truth  
**Безопасность:** ✅ Защищено
