# User Module

User Module отвечает за управление профилями пользователей, прогрессом обучения, дашбордом и достижениями.

## Архитектура

```
user/
├── dto/                    # Data Transfer Objects
│   └── requests.go        # Request/Response DTOs
├── repo/                   # Repository Layer (Database)
│   └── repository.go      # User data access
├── service/               # Business Logic Layer
│   └── service.go         # User business rules
├── transport/             # Transport Layer
│   └── http/
│       └── handlers.go    # HTTP handlers
├── module.go              # Module initialization
└── README.md             # This file
```

## Функциональность

### 1. Profile Management
- **GetProfile** - получение профиля пользователя
- **UpdateProfile** - обновление имени, аватара, языка

### 2. Learning Progress
- **GetProgress** - получение прогресса по курсам
- **GetDashboard** - комплексный дашборд с агрегированными данными

### 3. Achievements
- **GetAchievements** - получение разблокированных достижений

## API Endpoints

### Get User Profile
```http
GET /api/user/profile
Authorization: Bearer {token}
```

**Response:**
```json
{
  "userId": "uuid",
  "name": "John Doe",
  "email": "john@example.com",
  "level": 5,
  "stars": 120,
  "xp": 2400,
  "role": "user",
  "language": "ru",
  "avatarUrl": "https://...",
  "completedCourses": 8,
  "walletBalance": 150.50
}
```

### Update User Profile
```http
PUT /api/user/profile
Authorization: Bearer {token}
Content-Type: application/json

{
  "name": "John Doe Updated",
  "avatarUrl": "https://new-avatar.com/john.jpg",
  "language": "en"
}
```

**Response:**
```json
{
  "message": "profile updated successfully"
}
```

### Get User Progress
```http
GET /api/user/progress
Authorization: Bearer {token}
```

**Response:**
```json
[
  {
    "id": "uuid",
    "userId": "uuid",
    "courseId": "uuid",
    "completedLessons": 5,
    "totalLessons": 10,
    "progress": 50.0,
    "lastAccessedAt": "2025-01-15T10:30:00Z"
  }
]
```

### Get User Dashboard
```http
GET /api/user/dashboard
Authorization: Bearer {token}
```

**Response:**
```json
{
  "profile": {
    "level": 5,
    "stars": 120,
    "xp": 2400,
    "completedCourses": 8,
    "walletBalance": 150.50,
    "name": "John Doe",
    "avatarUrl": "https://...",
    "language": "ru"
  },
  "progressToNextLevel": 60.0,
  "nextLevelXP": 2500,
  "totalCourses": 50,
  "courseProgress": [
    {
      "courseId": "uuid",
      "courseName": "Основы кулинарии",
      "completedLessons": 5,
      "totalLessons": 10,
      "progress": 50.0,
      "lastAccessed": "2025-01-15T10:30:00Z"
    }
  ],
  "recentActivity": [
    {
      "type": "quiz",
      "course": "Итальянская кухня",
      "stars": 3,
      "score": 95,
      "timestamp": "2025-01-15T10:30:00Z"
    }
  ],
  "recommendations": [
    {
      "courseId": "uuid",
      "title": "Продвинутая паста",
      "description": "Изучите сложные техники приготовления пасты",
      "level": 6,
      "match": 85,
      "imageUrl": "https://..."
    }
  ],
  "recentTransactions": [
    {
      "id": "uuid",
      "amount": 50.0,
      "type": "purchase",
      "createdAt": "2025-01-14T15:00:00Z"
    }
  ],
  "activeRecipes": [
    {
      "id": "uuid",
      "name": "Карбонара",
      "progress": 60,
      "status": "in_progress"
    }
  ]
}
```

### Get User Achievements
```http
GET /api/user/achievements
Authorization: Bearer {token}
```

**Response:**
```json
[
  {
    "id": "uuid",
    "code": "first_course",
    "title": "Первый курс",
    "description": "Завершите ваш первый курс",
    "iconUrl": "https://...",
    "category": "learning",
    "unlockedAt": "2025-01-10T12:00:00Z"
  }
]
```

## Business Logic

### Profile Creation
- При первом запросе профиля автоматически создается профиль по умолчанию:
  - Level: 1
  - Stars: 0
  - XP: 0
  - WalletBalance: 0

### Profile Update
- Валидация: хотя бы одно поле должно быть заполнено
- Поддерживаемые поля: name, avatarUrl, language
- Остальные поля (level, stars, xp) обновляются через другие модули

### Dashboard Aggregation
Дашборд агрегирует данные из множества таблиц:
- **UserProfile** - основной профиль
- **Course** - информация о курсах
- **UserProgress** - прогресс по курсам (последние 5)
- **UserQuiz** - завершенные квизы (последние 5)
- **WalletTransaction** - транзакции кошелька (последние 5)
- **PersonalRecipe** - активные рецепты (последние 3)

### XP Calculation
- `nextLevelXP = currentLevel * 500`
- `progressToNextLevel = (currentXP / nextLevelXP) * 100`

### Recommendations
- Выбираются курсы на языке пользователя
- Уровень курса: текущий уровень пользователя ± 2
- Match percentage:
  - 100% - курс на текущем уровне
  - 85% - курс выше текущего уровня

## Database Schema

### UserProfile
```sql
CREATE TABLE user_profiles (
  user_id UUID PRIMARY KEY,
  name VARCHAR(255),
  email VARCHAR(255),
  level INT DEFAULT 1,
  stars INT DEFAULT 0,
  xp INT DEFAULT 0,
  role VARCHAR(50),
  language VARCHAR(10),
  avatar_url TEXT,
  completed_courses INT DEFAULT 0,
  wallet_balance DECIMAL(10,2) DEFAULT 0
);
```

### UserProgress
```sql
CREATE TABLE user_progress (
  id UUID PRIMARY KEY,
  user_id UUID REFERENCES users(id),
  course_id UUID REFERENCES courses(id),
  completed_lessons INT,
  total_lessons INT,
  progress DECIMAL(5,2),
  last_accessed_at TIMESTAMP
);
```

### UserAchievements
```sql
CREATE TABLE user_achievements (
  id UUID PRIMARY KEY,
  user_id UUID REFERENCES users(id),
  achievement_id UUID REFERENCES achievements(id),
  unlocked_at TIMESTAMP DEFAULT NOW()
);
```

## Error Handling

### Custom Errors
- `ErrProfileNotFound` - профиль не найден (автоматически создается)
- `ErrInvalidUpdateData` - нет данных для обновления

### HTTP Status Codes
- **200 OK** - успешный запрос
- **400 Bad Request** - неверные данные
- **401 Unauthorized** - не авторизован
- **500 Internal Server Error** - ошибка сервера

## Dependencies

### External
- `github.com/google/uuid` - UUID support
- `go.uber.org/zap` - структурированное логирование
- `gorm.io/gorm` - ORM для работы с БД

### Internal
- `backend/internal/models` - модели данных
- `backend/internal/middleware` - JWT middleware
- `backend/internal/platform/httpx` - HTTP helpers
- `backend/internal/platform/logger` - логгер

## Usage Example

```go
import (
    "gorm.io/gorm"
    "backend/internal/modules/user"
)

// Initialize module
userModule := user.NewModule(db)

// Register routes
userModule.RegisterRoutes(router, jwtMiddleware)
```

## Testing

```bash
# Get profile
curl -X GET http://localhost:8080/api/user/profile \
  -H "Authorization: Bearer YOUR_TOKEN"

# Update profile
curl -X PUT http://localhost:8080/api/user/profile \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"New Name","language":"en"}'

# Get dashboard
curl -X GET http://localhost:8080/api/user/dashboard \
  -H "Authorization: Bearer YOUR_TOKEN"

# Get achievements
curl -X GET http://localhost:8080/api/user/achievements \
  -H "Authorization: Bearer YOUR_TOKEN"
```

## Security

- ✅ Все эндпоинты требуют JWT аутентификацию
- ✅ User ID извлекается из JWT токена (не из запроса)
- ✅ Пользователь может редактировать только свой профиль
- ✅ Валидация всех входных данных
- ✅ Структурированное логирование всех операций

## Future Improvements

- [ ] Добавить кэширование dashboard data
- [ ] Пагинация для списка достижений
- [ ] Фильтры для course progress
- [ ] Webhooks при разблокировке достижений
- [ ] Rate limiting для update операций
