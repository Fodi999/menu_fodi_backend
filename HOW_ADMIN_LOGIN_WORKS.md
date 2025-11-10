# 🔐 Authentication & Admin Access - How It Works

**Date**: November 10, 2025

---

## ❌ Просто написать "admin" НЕ поможет!

Вы **не можете** просто написать "admin" при входе и получить доступ администратора.

Вот почему:

### Система входа требует:

1. **Email** (корректный email адрес)
2. **Password** (пароль пользователя)

```json
POST /api/auth/login
{
  "email": "user@example.com",
  "password": "password123"
}
```

### Что происходит при входе:

```
1. Backend получает email и password
   ↓
2. Ищет пользователя в БД по email
   ↓
3. Проверяет, что пароль совпадает (bcrypt сравнение)
   ↓
4. Генерирует JWT токен с user.Role из БД
   ↓
5. Возвращает токен
```

### Где берется роль администратора?

Роль администратора хранится в **базе данных** в таблице `User`:

```sql
SELECT id, email, name, role FROM "User" WHERE email = 'user@example.com';

Result:
id | email | name | role
----+-------+------+------
1  | user@example.com | John | user      ← обычный пользователь
2  | admin@example.com | Admin | admin   ← администратор
```

---

## ✅ Как ПРАВИЛЬНО получить доступ администратора

### Способ 1: Существует аккаунт администратора

Если админ уже существует:

```bash
# Войти с email админа
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.com",
    "password": "admin_password"
  }'

# Ответ:
{
  "success": true,
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "user": {
    "id": "uuid",
    "email": "admin@example.com",
    "name": "Admin User",
    "role": "admin"        ← роль в токене!
  }
}
```

### Способ 2: Назначить админа текущему пользователю

Если админ существует, он может назначить админа другому пользователю:

```bash
TOKEN="admin_jwt_token"

curl -X PATCH http://localhost:8080/api/admin/users/update-role \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user-uuid-to-promote",
    "role": "admin"
  }'
```

Затем этот пользователь может войти:

```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "newly.promoted@example.com",
    "password": "their_password"
  }'
```

### Способ 3: Прямое изменение БД (для первого админа)

Если нет администраторов в системе:

```sql
-- 1. Создать обычного пользователя (через регистрацию)
-- 2. Затем прямо в БД назначить админа:

UPDATE "User" SET role = 'admin' WHERE email = 'first.admin@example.com';
```

Затем войти обычным образом:

```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "first.admin@example.com",
    "password": "password123"
  }'
```

---

## 🔐 Процесс входа: Подробно

### Шаг 1: Отправка учетных данных

```javascript
const response = await fetch('/api/auth/login', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    email: 'admin@example.com',
    password: 'password123'
  })
});
```

### Шаг 2: Валидация на сервере

**Handler** (`internal/modules/auth/transport/http/handlers.go`):
```go
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
    var req dto.LoginRequest
    json.NewDecoder(r.Body).Decode(&req)
    
    // req.Email = "admin@example.com"
    // req.Password = "password123"
    
    response, err := h.service.Login(req)
    // ...
}
```

**Service** (`internal/modules/auth/service/service.go`):
```go
func (s *AuthService) Login(req dto.LoginRequest) (*dto.AuthResponse, error) {
    // 1. Найти пользователя по email
    user, err := s.repo.FindByEmail(req.Email)
    // SELECT * FROM "User" WHERE email = 'admin@example.com'
    // Result: id=1, email=admin@example.com, role=admin
    
    // 2. Проверить пароль (bcrypt сравнение)
    err := bcrypt.CompareHashAndPassword(user.Password, req.Password)
    // bcrypt сравнивает хэш из БД с введенным паролем
    
    // 3. Генерировать JWT токен с ролью из БД
    token, err := GenerateToken(user.ID, user.Email, user.Role)
    // JWT содержит: user_id, email, role="admin"
    
    return &dto.AuthResponse{
        Token: token,
        User: dto.ToUserInfo(user),  // role="admin"
    }
}
```

### Шаг 3: Получение токена с ролью

```json
{
  "success": true,
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMSIsImVtYWlsIjoiYWRtaW5AZXhhbXBsZS5jb20iLCJyb2xlIjoiYWRtaW4iLCJleHAiOjE3MzExNDE2MDB9.xyz...",
  "user": {
    "role": "admin"
  }
}
```

### Шаг 4: Использование токена для admin endpoints

```javascript
// Сохранить токен
localStorage.setItem('token', token);

// Использовать в админ запросе
const response = await fetch('/api/admin/stats', {
  headers: {
    'Authorization': `Bearer ${token}`
  }
});
```

### Шаг 5: Проверка роли на сервере

**AdminMiddleware** (`internal/middleware/auth.go`):
```go
func AdminMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // 1. Получить claims из JWT (в контексте)
        claims := r.Context().Value(UserContextKey).(*authservice.Claims)
        
        // 2. Проверить роль
        if claims.Role != "admin" {
            return 403 Forbidden: "Admin access required"
        }
        
        // 3. Если роль == "admin", выполнить handler
        next.ServeHTTP(w, r)
    })
}
```

---

## 📊 Диаграмма процесса

```
┌─────────────────────────────────────────────────────────────┐
│ FRONTEND                                                    │
│ ┌──────────────────────────────────────────────────────┐   │
│ │ Input: email="admin@example.com", password="pass123"│   │
│ └──────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│ BACKEND - AuthHandler.Login()                               │
│ Parse JSON request → LoginRequest                           │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│ BACKEND - AuthService.Login()                               │
│ 1. FindByEmail("admin@example.com")                         │
│    → SELECT FROM User WHERE email=...                       │
│    → User{id=1, email=admin@..., role="admin", pwd=hash}   │
│                                                              │
│ 2. CompareHashAndPassword(dbHash, inputPassword)            │
│    → bcrypt: does hash match?                               │
│    → YES ✓                                                  │
│                                                              │
│ 3. GenerateToken(id, email, role="admin")                   │
│    → JWT: {user_id:1, email:admin@..., role:"admin"}        │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│ RESPONSE                                                    │
│ {                                                           │
│   "token": "eyJ...xyz...",                                  │
│   "user": { "role": "admin" }                               │
│ }                                                           │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│ FRONTEND                                                    │
│ localStorage.setItem('token', token)                        │
│ Redirect to /admin/dashboard                                │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│ ADMIN REQUEST                                               │
│ GET /api/admin/stats                                        │
│ Headers: Authorization: Bearer eyJ...xyz...                 │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│ BACKEND - AuthMiddleware                                    │
│ 1. Validate JWT signature                                   │
│ 2. Extract claims: {user_id:1, role:"admin"}                │
│ 3. Add to context                                           │
│ 4. Continue to next middleware                              │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│ BACKEND - AdminMiddleware                                   │
│ 1. Get claims from context                                  │
│ 2. Check: claims.Role == "admin"?                           │
│    → YES ✓                                                  │
│ 3. Continue to handler                                      │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│ BACKEND - AdminHandler.GetAdminStats()                      │
│ Execute handler → Get stats from DB                         │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│ RESPONSE                                                    │
│ { "totalUsers": 100, "totalOrders": 500 }                  │
└─────────────────────────────────────────────────────────────┘
```

---

## 🔑 Ключевые моменты

### 1. Роль идет ИЗ БД, НЕ из input

```
❌ НЕПРАВИЛЬНО:
{
  "email": "admin@example.com",
  "password": "password123",
  "role": "admin"  ← это никак не влияет!
}

✅ ПРАВИЛЬНО:
{
  "email": "admin@example.com",
  "password": "password123"
}

Сервер берет роль из БД:
SELECT role FROM User WHERE email = 'admin@example.com'
```

### 2. Пароль ДОЛЖЕН совпадать

```
❌ НЕПРАВИЛЬНО:
{
  "email": "admin@example.com",
  "password": "любой текст"  ← ошибка, не совпадает
}

✅ ПРАВИЛЬНО:
{
  "email": "admin@example.com",
  "password": "правильный_пароль_админа"
}
```

### 3. Роль проверяется на сервере, не на клиенте

```
❌ НЕПРАВИЛЬНО (попытка обхода):
// Frontend: изменить декодированный JWT
localStorage.setItem('token', fakeTokenWithAdminRole)
→ Server: не валидирует сигнатуру → 401 Unauthorized

✅ ПРАВИЛЬНО:
// Frontend: использовать реальный токен от сервера
localStorage.setItem('token', tokenFromServer)
→ Server: валидирует сигнатуру → Success
```

---

## 🚀 Практический пример

### Создать первого администратора

```bash
# 1. Регистрировать пользователя
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.com",
    "password": "strong_password_123",
    "name": "System Admin"
  }'
# Role будет "user" по умолчанию

# 2. Обновить роль прямо в БД
# (если нет других админов)
psql -d database_name -c "UPDATE \"User\" SET role = 'admin' WHERE email = 'admin@example.com';"

# 3. Войти с правильными учетными данными
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.com",
    "password": "strong_password_123"
  }'

# 4. Получить токен админа
TOKEN="eyJhbGciOiJIUzI1NiIs..."
echo $TOKEN

# 5. Использовать админ endpoints
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/admin/stats
```

---

## 📋 Проверка списка

Для доступа администратора нужно:

- [ ] Пользователь зарегистрирован с email и паролем
- [ ] Роль пользователя в БД = "admin"
- [ ] Пользователь вводит правильный email и пароль
- [ ] Server валидирует пароль (bcrypt)
- [ ] Server генерирует JWT с role="admin"
- [ ] Frontend отправляет JWT в Authorization header
- [ ] Backend валидирует JWT сигнатуру
- [ ] AdminMiddleware проверяет role == "admin"
- [ ] Handler выполняет админ операцию

Если пропустить любой из этих шагов → **403 Forbidden: Admin access required**

---

## 🔗 Связанные файлы

- `internal/modules/auth/service/service.go` - Логика входа
- `internal/middleware/auth.go` - AdminMiddleware
- `internal/models/user.go` - Модель User с role
- `internal/modules/admin/module.go` - Админ routes с middleware

---

**Вывод**: Роль администратора определяется **в базе данных**, а не вводом при входе!
