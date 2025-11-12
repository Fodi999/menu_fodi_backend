# 👤 User Endpoints - Complete API Reference

Полное описание всех эндпоинтов для обычных пользователей (без admin прав).

---

## 📋 Quick Reference

| # | Endpoint | Method | Описание |
|---|----------|--------|---------|
| 1 | `/api/user/profile` | GET | Получить профиль |
| 2 | `/api/user/profile` | PUT | Обновить профиль |
| 3 | `/api/user/avatar` | POST | Загрузить аватар |
| 4 | `/api/user/settings` | GET | Получить настройки |
| 5 | `/api/user/settings` | PUT | Обновить настройки |
| 6 | `/api/user/courses` | GET | Список курсов |
| 7 | `/api/user/progress` | GET | Прогресс обучения |
| 8 | `/api/user/stats` | GET | Статистика |
| 9 | `/api/user/wallet` | GET | Кошелёк |
| 10 | `/api/user/tokens` | GET | Токины |

---

## 1️⃣ GET /api/user/profile

**Получить профиль текущего пользователя**

### Request

```bash
curl -X GET http://localhost:8080/api/user/profile \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..."
```

### Response (200 OK)

```json
{
  "userId": "7ec8aba4-8195-4be1-a9a8-067c30aae306",
  "name": "John Doe",
  "email": "john@example.com",
  "level": 5,
  "stars": 150,
  "xp": 2500,
  "role": "user",
  "language": "en",
  "avatarUrl": "https://example.com/avatars/user123.jpg",
  "completedCourses": 3,
  "walletBalance": 1500,
  "createdAt": "2024-01-15T10:30:00Z",
  "updatedAt": "2024-11-10T15:45:00Z"
}
```

### Параметры

| Параметр | Тип | Описание |
|----------|-----|---------|
| Authorization | Header | JWT токен (обязателен) |

### Ошибки

- **401 Unauthorized** - Нет токена или он неверный
- **500 Internal Server Error** - Ошибка сервера

---

## 2️⃣ PUT /api/user/profile

**Обновить профиль пользователя**

### Request

```bash
curl -X PUT http://localhost:8080/api/user/profile \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..." \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Updated",
    "email": "newemail@example.com",
    "language": "pl"
  }'
```

### Request Body

```json
{
  "name": "string (опционально)",
  "email": "string (опционально)",
  "language": "string (опционально)"
}
```

### Response (200 OK)

```json
{
  "message": "profile updated successfully"
}
```

### Допустимые языки

- `en` - English
- `pl` - Polski
- `ru` - Русский
- `de` - Deutsch
- `fr` - Français

### Ошибки

- **400 Bad Request** - Неверные данные (email некорректен, и т.д.)
- **401 Unauthorized** - Нет токена
- **500 Internal Server Error** - Ошибка сервера

---

## 3️⃣ POST /api/user/avatar

**Загрузить аватар пользователя**

### Request

```bash
curl -X POST http://localhost:8080/api/user/avatar \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..." \
  -F "avatar=@/path/to/avatar.jpg"
```

### Request Body

- `avatar` (file) - Изображение (JPG, PNG, GIF)
- Max size: 5MB

### Response (200 OK)

```json
{
  "message": "Avatar uploaded successfully",
  "avatarUrl": "https://example.com/avatars/user123_1699627200.jpg",
  "fileName": "user123_1699627200.jpg"
}
```

### Поддерживаемые форматы

- JPG / JPEG
- PNG
- GIF
- WebP

### Ошибки

- **400 Bad Request** - Размер файла > 5MB или неверный формат
- **401 Unauthorized** - Нет токена
- **500 Internal Server Error** - Ошибка при загрузке

---

## 4️⃣ GET /api/user/settings

**Получить настройки пользователя**

### Request

```bash
curl -X GET http://localhost:8080/api/user/settings \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..."
```

### Response (200 OK)

```json
{
  "userId": "7ec8aba4-8195-4be1-a9a8-067c30aae306",
  "language": "en",
  "notifications": true,
  "emailNotifications": true,
  "darkMode": false,
  "privateProfile": false,
  "twoFactorEnabled": false,
  "lastLoginAt": "2024-11-10T15:45:00Z"
}
```

### Параметры

| Параметр | Тип | Описание |
|----------|-----|---------|
| Authorization | Header | JWT токен (обязателен) |

---

## 5️⃣ PUT /api/user/settings

**Обновить настройки пользователя**

### Request

```bash
curl -X PUT http://localhost:8080/api/user/settings \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..." \
  -H "Content-Type: application/json" \
  -d '{
    "language": "pl",
    "notifications": false,
    "darkMode": true
  }'
```

### Request Body

```json
{
  "language": "string (опционально)",
  "notifications": "boolean (опционально)",
  "emailNotifications": "boolean (опционально)",
  "darkMode": "boolean (опционально)",
  "privateProfile": "boolean (опционально)"
}
```

### Response (200 OK)

```json
{
  "message": "Settings updated successfully"
}
```

### Поля

| Поле | Тип | Default | Описание |
|------|-----|---------|---------|
| language | string | en | Язык интерфейса |
| notifications | boolean | true | Push уведомления |
| emailNotifications | boolean | true | Email уведомления |
| darkMode | boolean | false | Тёмная тема |
| privateProfile | boolean | false | Приватный профиль |

---

## 6️⃣ GET /api/user/courses

**Получить список пройденных курсов**

### Request

```bash
curl -X GET "http://localhost:8080/api/user/courses?limit=20&offset=0&status=completed" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..."
```

### Query Parameters

| Параметр | Тип | Default | Описание |
|----------|-----|---------|---------|
| limit | int | 10 | Количество результатов |
| offset | int | 0 | Смещение |
| status | string | - | 'completed', 'in_progress', 'not_started' |

### Response (200 OK)

```json
{
  "userId": "7ec8aba4-8195-4be1-a9a8-067c30aae306",
  "totalCourses": 3,
  "completedCourses": 3,
  "inProgressCourses": 2,
  "courses": [
    {
      "courseId": "course-001",
      "title": "Go Programming Basics",
      "description": "Learn Go from scratch",
      "instructor": "John Smith",
      "status": "completed",
      "progress": 100,
      "enrolledAt": "2024-01-15T10:30:00Z",
      "completedAt": "2024-03-20T14:22:00Z",
      "certificateUrl": "https://example.com/certificates/course001.pdf"
    },
    {
      "courseId": "course-002",
      "title": "REST APIs with Go",
      "description": "Build RESTful APIs using Go",
      "instructor": "Jane Doe",
      "status": "in_progress",
      "progress": 65,
      "enrolledAt": "2024-05-10T09:15:00Z",
      "completedAt": null
    }
  ]
}
```

### Статусы курса

- `not_started` - Зарегистрирован но не начат
- `in_progress` - Проходит в данный момент
- `completed` - Завершён

---

## 7️⃣ GET /api/user/progress

**Получить прогресс обучения**

### Request

```bash
curl -X GET http://localhost:8080/api/user/progress \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..."
```

### Response (200 OK)

```json
{
  "userId": "7ec8aba4-8195-4be1-a9a8-067c30aae306",
  "overallProgress": 68,
  "currentLevel": 5,
  "levelProgress": 45,
  "nextLevelXP": 3000,
  "currentXP": 2500,
  "totalXPEarned": 5000,
  "streakDays": 12,
  "lastActivityDate": "2024-11-10T15:45:00Z",
  "milestones": [
    {
      "id": "milestone-001",
      "title": "First Course Completed",
      "description": "Complete your first course",
      "completed": true,
      "completedAt": "2024-03-20T14:22:00Z",
      "reward": 100
    }
  ],
  "badges": [
    {
      "id": "badge-001",
      "name": "Fast Learner",
      "description": "Complete a course in less than 30 days",
      "icon": "https://example.com/badges/fast-learner.png",
      "earnedAt": "2024-03-20T14:22:00Z"
    }
  ]
}
```

---

## 8️⃣ GET /api/user/stats

**Получить личную статистику пользователя**

### Request

```bash
curl -X GET http://localhost:8080/api/user/stats \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..."
```

### Response (200 OK)

```json
{
  "userId": "7ec8aba4-8195-4be1-a9a8-067c30aae306",
  "stats": {
    "totalCoursesEnrolled": 5,
    "totalCoursesCompleted": 3,
    "totalLessonsCompleted": 45,
    "totalQuizzesCompleted": 23,
    "quizzAverageScore": 88.5,
    "totalStudyHours": 125,
    "averageStudyTimePerDay": 1.5,
    "totalXPEarned": 5000,
    "totalStarsEarned": 150,
    "currentLevel": 5,
    "currentRank": 42,
    "totalUsersInSystem": 1000,
    "streakCurrentDays": 12,
    "streakMaxDays": 25,
    "certificatesEarned": 3,
    "lastActivityTime": "2024-11-10T15:45:00Z",
    "joinDate": "2024-01-15T10:30:00Z"
  }
}
```

### Объяснение полей

| Поле | Описание |
|------|---------|
| totalCoursesEnrolled | Всего зарегистрировано в курсах |
| totalCoursesCompleted | Всего завершено курсов |
| totalLessonsCompleted | Всего уроков пройдено |
| totalQuizzesCompleted | Всего тестов пройдено |
| quizzAverageScore | Средний результат тестов |
| totalStudyHours | Всего часов учёбы |
| averageStudyTimePerDay | Среднее время в день |
| totalXPEarned | Всего XP заработано |
| totalStarsEarned | Всего звёзд получено |
| currentLevel | Текущий уровень |
| currentRank | Место в рейтинге |
| streakCurrentDays | Текущая последовательность дней |
| streakMaxDays | Максимальная последовательность |

---

## 9️⃣ GET /api/user/wallet

**Получить информацию о кошельке пользователя**

### Request

```bash
curl -X GET http://localhost:8080/api/user/wallet \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..."
```

### Response (200 OK)

```json
{
  "userId": "7ec8aba4-8195-4be1-a9a8-067c30aae306",
  "wallet": {
    "balance": 1500,
    "currency": "tokens",
    "lastTransaction": "2024-11-10T14:30:00Z",
    "totalEarned": 2000,
    "totalSpent": 500
  },
  "earnings": {
    "coursesCompleted": 1000,
    "quizzesCompleted": 600,
    "bonuses": 400,
    "referrals": 0
  },
  "spending": {
    "courseEnrollments": 300,
    "premiumFeatures": 150,
    "rewards": 50
  },
  "transactionHistory": [
    {
      "id": "txn-001",
      "date": "2024-11-10T14:30:00Z",
      "type": "credit",
      "amount": 100,
      "description": "Course completion: Go Basics",
      "balance": 1500
    },
    {
      "id": "txn-002",
      "date": "2024-11-09T10:15:00Z",
      "type": "debit",
      "amount": 50,
      "description": "Enrollment in REST APIs course",
      "balance": 1400
    }
  ]
}
```

### Типы транзакций

- `credit` - Пополнение баланса (заработок)
- `debit` - Уменьшение баланса (трата)

---

## 🔟 GET /api/user/tokens

**Получить информацию о токинах пользователя**

### Request

```bash
curl -X GET http://localhost:8080/api/user/tokens \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..."
```

### Response (200 OK)

```json
{
  "userId": "7ec8aba4-8195-4be1-a9a8-067c30aae306",
  "tokenBank": {
    "totalBalance": 1500,
    "earned": 2000,
    "spent": 500,
    "lastUpdated": "2024-11-10T15:45:00Z"
  },
  "recentTransactions": [
    {
      "id": "txn-001",
      "type": "earned",
      "amount": 100,
      "reason": "Course completion",
      "relatedItem": "course-001",
      "timestamp": "2024-11-10T14:30:00Z"
    },
    {
      "id": "txn-002",
      "type": "spent",
      "amount": 50,
      "reason": "Course enrollment",
      "relatedItem": "course-003",
      "timestamp": "2024-11-09T10:15:00Z"
    }
  ],
  "availableRewards": [
    {
      "id": "reward-001",
      "title": "Premium Course Access",
      "description": "Get access to premium courses",
      "tokensRequired": 200,
      "icon": "https://example.com/rewards/premium.png"
    }
  ]
}
```

---

## 🔐 Authentication

Все эндпоинты требуют JWT токена:

```bash
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

### Как получить токен

```bash
# 1. Зарегистрироваться
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123",
    "name": "John Doe"
  }'

# 2. Войти
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123"
  }'

# Response:
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "user": { ... }
}

# 3. Использовать токен
curl -X GET http://localhost:8080/api/user/profile \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..."
```

---

## 📊 Error Responses

### 401 Unauthorized

```json
{
  "error": "Unauthorized",
  "message": "Invalid or expired token"
}
```

### 400 Bad Request

```json
{
  "error": "Bad Request",
  "message": "Invalid request parameters"
}
```

### 404 Not Found

```json
{
  "error": "Not Found",
  "message": "Resource not found"
}
```

### 500 Server Error

```json
{
  "error": "Internal Server Error",
  "message": "An unexpected error occurred"
}
```

---

## 🧪 Testing Script

```bash
#!/bin/bash

# Set variables
TOKEN="your_jwt_token_here"
USER_ID="your_user_id_here"
API_URL="http://localhost:8080"

echo "🧪 Testing User Endpoints"

# 1. Get profile
echo "1. Get profile"
curl -X GET $API_URL/api/user/profile \
  -H "Authorization: Bearer $TOKEN" | jq '.'

# 2. Update profile
echo "2. Update profile"
curl -X PUT $API_URL/api/user/profile \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"New Name"}' | jq '.'

# 3. Get settings
echo "3. Get settings"
curl -X GET $API_URL/api/user/settings \
  -H "Authorization: Bearer $TOKEN" | jq '.'

# 4. Get courses
echo "4. Get courses"
curl -X GET "$API_URL/api/user/courses?limit=10" \
  -H "Authorization: Bearer $TOKEN" | jq '.'

# 5. Get progress
echo "5. Get progress"
curl -X GET $API_URL/api/user/progress \
  -H "Authorization: Bearer $TOKEN" | jq '.'

# 6. Get stats
echo "6. Get stats"
curl -X GET $API_URL/api/user/stats \
  -H "Authorization: Bearer $TOKEN" | jq '.'

# 7. Get wallet
echo "7. Get wallet"
curl -X GET $API_URL/api/user/wallet \
  -H "Authorization: Bearer $TOKEN" | jq '.'

# 8. Get tokens
echo "8. Get tokens"
curl -X GET $API_URL/api/user/tokens \
  -H "Authorization: Bearer $TOKEN" | jq '.'
```

---

## 📋 Summary Table

| Endpoint | Method | Auth | Body | Кеш |
|----------|--------|------|------|-----|
| `/api/user/profile` | GET | ✅ | - | 5 мин |
| `/api/user/profile` | PUT | ✅ | ✅ | Очистить |
| `/api/user/avatar` | POST | ✅ | File | Очистить |
| `/api/user/settings` | GET | ✅ | - | 10 мин |
| `/api/user/settings` | PUT | ✅ | ✅ | Очистить |
| `/api/user/courses` | GET | ✅ | - | 15 мин |
| `/api/user/progress` | GET | ✅ | - | 5 мин |
| `/api/user/stats` | GET | ✅ | - | 1 час |
| `/api/user/wallet` | GET | ✅ | - | 5 мин |
| `/api/user/tokens` | GET | ✅ | - | 5 мин |

---

## Performance Tips

1. **Кеширование:** Профиль редко меняется - кешируйте на 5 минут
2. **Пагинация:** Курсы и транзакции - используйте limit/offset
3. **Индексы:** На user_id для быстрого поиска
4. **Batch запросы:** Загружайте несколько данных за раз
5. **CDN:** Аватары и иконки служите через CDN

