# Auth Module

Модуль аутентификации и авторизации пользователей.

## Структура

```
auth/
├── dto/              # Data Transfer Objects
│   └── requests.go   # Request/Response DTOs
├── repo/             # Repository (Database layer)
│   └── repository.go # Database operations
├── service/          # Business Logic layer
│   └── service.go    # Authentication logic
├── transport/        # Transport layer
│   └── http/         # HTTP handlers
│       └── handlers.go
└── module.go         # Module registration
```

## API Endpoints

### Public Routes

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/auth/register` | Регистрация нового пользователя |
| POST | `/api/auth/login` | Вход в систему |
| POST | `/api/auth/verify` | Верификация JWT токена |

### Protected Routes

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/auth/me` | Получить текущего пользователя |

## Request/Response Examples

### Register User
```bash
POST /api/auth/register
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password123",
  "name": "John Doe"
}
```

Response:
```json
{
  "success": true,
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "uuid",
    "email": "user@example.com",
    "name": "John Doe",
    "role": "user",
    "createdAt": "2025-11-09T10:00:00Z"
  },
  "message": "Registration successful"
}
```

### Login
```bash
POST /api/auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password123"
}
```

Response:
```json
{
  "success": true,
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "uuid",
    "email": "user@example.com",
    "name": "John Doe",
    "role": "user",
    "createdAt": "2025-11-09T10:00:00Z"
  },
  "message": "Login successful"
}
```

### Verify Token
```bash
POST /api/auth/verify
Content-Type: application/json

{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

Response (Valid):
```json
{
  "success": true,
  "data": {
    "valid": true,
    "user_id": "uuid",
    "role": "user",
    "name": "John Doe",
    "email": "user@example.com"
  }
}
```

Response (Invalid):
```json
{
  "success": true,
  "data": {
    "valid": false
  }
}
```

### Get Current User
```bash
GET /api/auth/me
Authorization: Bearer {token}
```

Response:
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "email": "user@example.com",
    "name": "John Doe",
    "role": "user",
    "createdAt": "2025-11-09T10:00:00Z",
    "walletBalance": 100,
    "profile": {
      "level": 5,
      "stars": 120,
      "xp": 1500,
      "avatarUrl": "https://...",
      "language": "ua",
      "completedCourses": 3
    }
  }
}
```

## Business Logic

### Service Layer (`service/service.go`)

Основные методы:
- `Register(req)` - Регистрация нового пользователя
- `Login(req)` - Аутентификация пользователя
- `VerifyToken(req)` - Верификация JWT токена
- `GetCurrentUser(userID)` - Получение текущего пользователя с профилем

Бизнес-правила:
- ✅ Проверка уникальности email при регистрации
- ✅ Хеширование пароля с bcrypt
- ✅ Минимальная длина пароля: 6 символов
- ✅ Генерация JWT токена после успешной аутентификации
- ✅ Роль по умолчанию: "user"
- ✅ Проверка существования пользователя при верификации токена

### Repository Layer (`repo/repository.go`)

Операции с БД:
- `FindByEmail(email)` - Поиск пользователя по email
- `FindByID(id)` - Поиск пользователя по ID
- `Create(user)` - Создание нового пользователя
- `Update(user)` - Обновление данных пользователя
- `GetUserProfile(userID)` - Получение профиля пользователя

Таблицы:
- `User` - основная информация о пользователе
- `UserProfile` - дополнительные данные профиля

## Error Handling

Типы ошибок:
- `ErrUserExists` - Пользователь уже существует
- `ErrInvalidCredentials` - Неверные учетные данные
- `ErrUserNotFound` - Пользователь не найден
- `ErrInvalidToken` - Недействительный токен
- `ErrWeakPassword` - Слабый пароль (< 6 символов)

HTTP статусы:
- `200 OK` - Успешная операция
- `201 Created` - Пользователь создан
- `400 Bad Request` - Неверные данные запроса
- `401 Unauthorized` - Отсутствует авторизация
- `404 Not Found` - Пользователь не найден
- `500 Internal Server Error` - Ошибка сервера

## Security

### Password Hashing
Используется bcrypt с DefaultCost для хеширования паролей.

```go
hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
```

### JWT Tokens
- Алгоритм: HS256
- Claims: userID, email, role
- Срок действия: настраивается в auth package

### Validation
- Email: проверка формата
- Password: минимум 6 символов
- Name: минимум 2 символа

## Dependencies

- `golang.org/x/crypto/bcrypt` - Password hashing
- `github.com/google/uuid` - UUID generation
- `gorm.io/gorm` - ORM для работы с БД
- `internal/auth` - JWT generation and validation
- `internal/middleware` - AuthMiddleware
- `internal/platform/httpx` - HTTP response helpers
- `internal/platform/logger` - Structured logging

## Testing

```bash
# Unit tests для service layer
go test ./internal/modules/auth/service/...

# Integration tests для HTTP handlers
go test ./internal/modules/auth/transport/http/...

# Run all tests
go test ./internal/modules/auth/...
```

## Usage in Application

```go
// В internal/app/routes_modular.go
authModule := auth.NewModule()
authModule.RegisterRoutes(r)
```

## Integration with Other Modules

### Wallet Module
После регистрации можно выдать приветственные токены:
```go
// После успешной регистрации
walletService.GrantWelcomeTokens(userID, 100)
```

### User Module
Профиль пользователя создается после регистрации и может быть обновлен через User Module.

## Future Improvements

- [ ] OAuth2 integration (Google, Facebook)
- [ ] Two-factor authentication (2FA)
- [ ] Email verification
- [ ] Password reset functionality
- [ ] Refresh tokens
- [ ] Session management
- [ ] Rate limiting для login attempts
- [ ] Account lockout after failed attempts
- [ ] Audit logging для security events
