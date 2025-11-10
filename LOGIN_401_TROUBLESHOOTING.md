# 🔧 Troubleshooting: Login 401 Error - Invalid Credentials

**Problem**: Frontend получает 401 Unauthorized с ошибкой "Invalid credentials"

```
📡 Attempting login with email: user@example.com
📤 Request body: {"email":"user@example.com","password":"password123"}
❌ POST https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/auth/login 401 (Unauthorized)
❌ API Error 401: Invalid credentials
```

---

## 🔍 Что это означает?

**401 Unauthorized** с сообщением **"Invalid credentials"** значит:

❌ Пользователь с email `user@example.com` **НЕ СУЩЕСТВУЕТ в БД**  
❌ ИЛИ пароль неправильный для существующего пользователя  
❌ ИЛИ пользователь не был зарегистрирован в production БД  

---

## 🚨 Production Issue: БД может быть пуста

Если вы развернули backend на production (Koyeb) и БД только что создана:

**Production Server**: `https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app`

**Проблема**: Production БД может быть пустой (без пользователей)

**Решение**: Нужно создать тестовых пользователей!

---

## ✅ Шаги для решения

### Шаг 1: Проверить, что пользователь зарегистрирован

Сначала зарегистрируйте пользователя:

```bash
POST https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/auth/register
Content-Type: application/json

{
  "email": "test@example.com",
  "password": "password123",
  "name": "Test User"
}
```

**Ответ (201 Created)**:
```json
{
  "success": true,
  "message": "Registration successful",
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "user": {
    "id": "user-uuid",
    "email": "test@example.com",
    "name": "Test User",
    "role": "user"
  }
}
```

### Шаг 2: Использовать те же учетные данные для входа

Используйте **точно те же** email и пароль что при регистрации:

```bash
POST https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/auth/login
Content-Type: application/json

{
  "email": "test@example.com",
  "password": "password123"
}
```

**Ответ (200 OK)**:
```json
{
  "success": true,
  "message": "Login successful",
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "user": {
    "id": "user-uuid",
    "email": "test@example.com",
    "name": "Test User",
    "role": "user"
  }
}
```

### Шаг 3: Сохранить токен

```javascript
// Frontend:
const response = await fetch('/api/auth/login', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    email: 'test@example.com',
    password: 'password123'
  })
});

const data = await response.json();

if (data.token) {
  // ✅ Сохранить токен
  localStorage.setItem('token', data.token);
  // ✅ Сохранить роль
  localStorage.setItem('role', data.user.role);
  // ✅ Перенаправить
  window.location.href = '/dashboard';
} else {
  // ❌ Ошибка входа
  console.error('Login failed:', data.message);
}
```

---

## 🐛 Частые причины ошибки

### Причина 1: Неправильный email

```javascript
❌ НЕПРАВИЛЬНО:
email: "test@example.com"   // Но в БД "test@example.com" не существует!

✅ ПРАВИЛЬНО:
email: "admin@example.com"  // Используйте email из регистрации
```

**Решение**: Проверьте, что email зарегистрирован:

```bash
# Сначала зарегистрируйте:
POST /api/auth/register
{
  "email": "myemail@example.com",
  "password": "mypassword",
  "name": "My Name"
}

# Потом используйте эот же email при входе
POST /api/auth/login
{
  "email": "myemail@example.com",
  "password": "mypassword"
}
```

### Причина 2: Неправильный пароль

```javascript
❌ НЕПРАВИЛЬНО:
password: "wrongpassword"

✅ ПРАВИЛЬНО:
password: "password123"  // Точно такой же как при регистрации
```

**Решение**: Используйте точно такой же пароль что при регистрации!

### Причина 3: Опечатка в email

```javascript
❌ НЕПРАВИЛЬНО:
email: "test@example.com"   // Но регистрировались как "test@example.com"
                            // (может быть пробел в конце)

email: "test@example.com "  // ← пробел в конце!

✅ ПРАВИЛЬНО:
email: "test@example.com"   // Без пробелов
```

**Решение**: Проверьте что нет пробелов в email!

### Причина 4: БД не содержит пользователей

Если база данных пуста (после пересоздания БД), нужно:

1. Сначала зарегистрировать пользователя
2. Потом делать вход

```bash
# 1. Регистрация (создает пользователя в БД)
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "newuser@example.com",
    "password": "password123",
    "name": "New User"
  }'

# 2. Вход (с теми же учетными данными)
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "newuser@example.com",
    "password": "password123"
  }'
```

---

## 🔒 Безопасность пароля

### Как хранится пароль?

Пароль **никогда** не хранится в открытом виде!

```
При регистрации:
"password123" → bcrypt hash → "$2a$10$..." → сохраняется в БД

При входе:
"password123" (введенный пароль) 
    ↓
bcrypt.CompareHashAndPassword(dbHash, inputPassword)
    ↓
match? → YES ✅ → вход успешен
```

### Почему 401 если пароль неправильный?

**Потому что это безопасно!** Сервер не говорит:
- "Email не найден" → раскрыл бы какие email в системе
- "Пароль неправильный" → раскрыл бы что email существует

Вместо этого просто:
- "Invalid credentials" → могло быть что угодно

---

## 📋 Чек-лист решения

- [ ] Зарегистрировал пользователя (GET 201 Created при регистрации)
- [ ] Использую ТЕ ЖЕ email и пароль что при регистрации
- [ ] Проверил что нет опечаток в email
- [ ] Проверил что нет пробелов в email
- [ ] Пароль не содержит опечаток
- [ ] Попробовал с разными данными (новый email/пароль)
- [ ] Проверил что backend работает (не 504 Gateway Timeout)
- [ ] Проверил консоль browser (F12 → Console)
- [ ] Попробовал с curl команды

---

## 🚀 Полный пример: От регистрации к админ панели

### 1. Регистрация первого администратора

```bash
# Зарегистрировать пользователя
curl -X POST https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.com",
    "password": "strong_password_123",
    "name": "System Admin"
  }'

# Сохранить ID и email из ответа
```

### 2. Назначить роль администратора (в БД)

```bash
# Прямое обновление БД (если есть доступ к БД)
psql -d database_name -c "UPDATE \"User\" SET role = 'admin' WHERE email = 'admin@example.com';"
```

### 3. Вход с учетными данными администратора

```bash
curl -X POST https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.com",
    "password": "strong_password_123"
  }'

# Ответ:
{
  "success": true,
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "user": {
    "role": "admin"
  }
}
```

### 4. Использовать админ endpoints

```bash
TOKEN="eyJhbGciOiJIUzI1NiIs..."

curl -H "Authorization: Bearer $TOKEN" \
  https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/admin/stats
```

---

## 🔗 Связанные файлы

- `internal/modules/auth/service/service.go` - Логика входа
- `internal/modules/auth/transport/http/handlers.go` - Auth handler
- `HOW_ADMIN_LOGIN_WORKS.md` - Как работает вход

---

## 💡 Советы

1. **Используйте правильный endpoint**:
   - Регистрация: `/api/auth/register`
   - Вход: `/api/auth/login`

2. **Проверьте статус кода**:
   - 201 Created → успешная регистрация
   - 200 OK → успешный вход
   - 400 Bad Request → ошибка в данных
   - 401 Unauthorized → неправильные учетные данные
   - 500 Internal Error → ошибка сервера

3. **Проверьте консоль**:
   - Откройте DevTools (F12 → Network)
   - Посмотрите Request/Response при логине
   - Проверьте что отправляется правильный JSON

4. **Тестируйте с curl**:
   ```bash
   # Сначала тестируйте через curl (проще)
   # Потом фиксьте фронтенд
   ```

---

**Главное**: Сначала **зарегистрируйте** пользователя, потом **входите** с теми же учетными данными!
