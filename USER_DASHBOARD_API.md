# User Dashboard & Culinary Academy API

## 📊 Обзор

User Dashboard + Culinary Academy - комплексная система обучения кулинарии с AI-наставником, сертификатами и виртуальной экономикой ChefToken.

**Новое в v1.0:**
- ✅ AI Recipe Review - оценка рецептов с рейтингом и советами
- Dashboard с Kitchen Simulator функциями
- 12 Achievements (достижения)
- Chef Ranking система (PL/UA/EN)

---

## 🧠 AI Recipe Review API

### POST `/api/ai/review-recipe`
**AI оценивает рецепт ученика** - использует Groq API для анализа рецепта с рейтингом, комментариями и рекомендациями.

**Request:**
```json
{
  "recipeId": "fadafe77-3a6b-4b70-906a-a0f035bd579f",
  "language": "pl"  // pl, ua, en
}
```

**Response:**
```json
{
  "status": "ok",
  "data": {
    "recipeId": "fadafe77-3a6b-4b70-906a-a0f035bd579f",
    "rating": 8.0,
    "chefComment": "Doskonałe połączenie kremowego sera z świeżym łososiem",
    "tasteBalance": "creamy-fresh",
    "difficulty": "easy",
    "improvements": [
      "Użyj świeżego soku z cytryny dla lekkości",
      "Dodaj odrobinę wasabi dla pikantności"
    ],
    "estimatedPrice": 45.50
  }
}
```

**Особенности:**
- ✅ Автоматически сохраняет рейтинг в `PersonalRecipe.rating`
- ✅ Многоязычность: PL, UA, EN
- ✅ AI анализ: вкусовой баланс, сложность, стоимость
- ✅ Fallback на 7.0 если AI недоступен

**Пример тестирования:**
```bash
# Польский рецепт
curl -X POST http://localhost:8080/api/ai/review-recipe \
  -H 'Content-Type: application/json' \
  -d '{
    "recipeId": "fadafe77-3a6b-4b70-906a-a0f035bd579f",
    "language": "pl"
  }'

# Украинский рецепт
curl -X POST http://localhost:8080/api/ai/review-recipe \
  -H 'Content-Type: application/json' \
  -d '{
    "recipeId": "44b75c8e-591d-49fe-b3bb-39bc3564916e",
    "language": "ua"
  }'
```

---

## 👨‍🍳 User Dashboard API (`/api/user`)

### GET `/api/user/{userId}/profile`
Получить профиль ученика. Автоматически создаёт профиль, если не существует.

**Response:**
```json
{
  "id": "uuid",
  "userId": "uuid",
  "name": "Дмитрий",
  "email": "user@example.com",
  "avatarUrl": "https://...",
  "level": 3,
  "stars": 25,
  "role": "student",
  "language": "pl",
  "xp": 1250,
  "completedCourses": 2,
  "favoriteRecipes": ["uuid1", "uuid2"],
  "walletBalance": 150.00,
  "createdAt": "2025-11-03T10:00:00Z"
}
```

### PUT `/api/user/{userId}/profile`
Обновить профиль ученика.

**Request:**
```json
{
  "name": "Новое имя",
  "avatarUrl": "https://cloudinary.com/...",
  "language": "ua"
}
```

### GET `/api/user/{userId}/progress`
Получить прогресс по всем курсам.

**Response:**
```json
[
  {
    "id": "uuid",
    "userId": "uuid",
    "courseId": "uuid",
    "completedLessons": 5,
    "totalLessons": 10,
    "quizScore": 85,
    "starsEarned": 4,
    "isCompleted": false,
    "lastAccessedAt": "2025-11-03T10:00:00Z"
  }
]
```

### GET `/api/user/{userId}/certificates`
Получить все сертификаты ученика.

**Response:**
```json
[
  {
    "id": "uuid",
    "userId": "uuid",
    "courseId": "uuid",
    "courseName": "Sushi Master",
    "userName": "Дмитрий Фомин",
    "level": 5,
    "stars": 25,
    "pdfUrl": "https://cloudinary.com/certificate.pdf",
    "signature": "Dima Fomin AI Academy",
    "issuedAt": "2025-11-01T15:00:00Z"
  }
]
```

### GET `/api/user/{userId}/recipes`
Получить личные рецепты ученика.

**Response:**
```json
[
  {
    "id": "uuid",
    "userId": "uuid",
    "title": "Philadelphia Roll Deluxe",
    "description": "Мой улучшенный Philadelphia",
    "ingredients": ["Łosoś 200g", "Ser Philadelphia 100g", "Awokado"],
    "steps": ["Przygotuj ryż", "Posiekaj łososia", "Zwiń maki"],
    "imageUrl": "https://cloudinary.com/...",
    "category": "sushi",
    "difficulty": "medium",
    "cookingTime": 45,
    "servings": 4,
    "rating": 8.5,
    "isPublic": true,
    "price": 25.00,
    "purchases": 12,
    "createdAt": "2025-11-01T10:00:00Z"
  }
]
```

### POST `/api/user/{userId}/recipes`
Создать новый личный рецепт.

**Request:**
```json
{
  "title": "My Sushi Roll",
  "description": "Unique recipe",
  "ingredients": ["Rice", "Salmon", "Nori"],
  "steps": ["Step 1", "Step 2"],
  "imageUrl": "https://...",
  "category": "sushi",
  "difficulty": "easy",
  "cookingTime": 30,
  "servings": 2
}
```

### DELETE `/api/user/{userId}/recipes/{recipeId}`
Удалить личный рецепт.

### GET `/api/user/{userId}/wallet`
Получить баланс ChefToken и историю транзакций.

**Response:**
```json
{
  "balance": 150.00,
  "transactions": [
    {
      "id": "uuid",
      "userId": "uuid",
      "amount": 50.00,
      "type": "reward",
      "description": "Quiz completion reward: 5 stars",
      "relatedId": "courseId",
      "createdAt": "2025-11-03T10:00:00Z"
    }
  ]
}
```

### GET `/api/user/{userId}/market/purchases`
Получить купленные рецепты в marketplace.

**Response:**
```json
{
  "purchases": [
    {
      "id": "uuid",
      "buyerId": "uuid",
      "sellerId": "uuid",
      "recipeId": "uuid",
      "price": 25.00,
      "createdAt": "2025-11-02T14:00:00Z"
    }
  ],
  "recipes": [ /* полные данные рецептов */ ]
}
```

---

## 🎓 Culinary Academy API (`/api/academy`)

### GET `/api/academy/courses`
Получить список всех опубликованных курсов.

**Query Parameters:**
- `language` (optional): `pl`, `ua`, `en`
- `category` (optional): `sushi`, `sashimi`, `knife-skills`, etc.

**Response:**
```json
[
  {
    "id": "uuid",
    "title": "Sushi Master Course",
    "description": "Learn professional sushi making",
    "imageUrl": "https://...",
    "level": 3,
    "category": "sushi",
    "duration": 120,
    "lessonsCount": 10,
    "language": "pl",
    "isPublished": true,
    "instructor": "Chef Nakamura",
    "stars": 50,
    "createdAt": "2025-10-01T00:00:00Z"
  }
]
```

### GET `/api/academy/courses/{courseId}`
Получить детали курса.

### GET `/api/academy/courses/{courseId}/lessons`
Получить уроки курса (отсортированы по порядку).

**Response:**
```json
[
  {
    "id": "uuid",
    "courseId": "uuid",
    "title": "Wprowadzenie do sushi",
    "description": "Podstawy przygotowania sushi",
    "videoUrl": "https://youtube.com/...",
    "order": 1,
    "duration": 15,
    "content": "Tekst lekcji...",
    "steps": ["Krok 1", "Krok 2"],
    "isPublished": true
  }
]
```

### GET `/api/academy/lessons/{lessonId}`
Получить детали урока.

### GET `/api/academy/quiz/{courseId}`
Получить случайные 10 вопросов для теста курса.

**Response:**
```json
[
  {
    "id": "uuid",
    "courseId": "uuid",
    "question": "Jaka temperatura jest idealna dla ryżu sushi?",
    "options": ["20°C", "30°C", "40°C", "50°C"],
    "explanation": "Ryż sushi najlepiej smakuje w temperaturze pokojowej (20-25°C)",
    "difficulty": "medium",
    "language": "pl"
  }
]
```

**Note:** `correctAnswer` не возвращается клиенту!

### POST `/api/academy/quiz/{courseId}/submit`
Отправить ответы на тест и получить результат.

**Request:**
```json
{
  "userId": "uuid",
  "answers": [0, 2, 1, 3, 0, 1, 2, 0, 3, 1]
}
```

**Response:**
```json
{
  "score": 80,
  "correctAnswers": 8,
  "totalQuestions": 10,
  "stars": 4,
  "reward": 40
}
```

**Награды:**
- 90%+ → 5 звёзд (50 ChefToken)
- 80-89% → 4 звезды (40 ChefToken)
- 70-79% → 3 звезды (30 ChefToken)
- 60-69% → 2 звезды (20 ChefToken)
- 50-59% → 1 звезда (10 ChefToken)
- <50% → 0 звёзд

---

## 🧠 AI Mentor API

### POST `/api/mentor/analyze-step`
Анализ шага рецепта с советами от AI Chef.

**Request:**
```json
{
  "step": "Obierz i pokrój łososia.",
  "language": "pl"
}
```

**Response:**
```json
{
  "status": "ok",
  "data": {
    "comment": "Upewnij się, że nóż jest bardzo ostry – to klucz do perfekcyjnych plasterków. Łosoś powinien być schłodzony, ale nie zamrożony..."
  }
}
```

**Supported Languages:** `pl`, `ua`, `en`

---

## 📊 Database Schema

### UserProfile
- `id` (uuid, PK)
- `user_id` (uuid, unique) - связь с User
- `name`, `email`, `avatar_url`
- `level` (1-10), `stars`, `xp`
- `role` (student, mentor, admin)
- `language` (pl, ua, en)
- `completed_courses`, `favorite_recipes[]`
- `wallet_balance` (decimal)

### PersonalRecipe
- `id` (uuid, PK)
- `user_id` (uuid)
- `title`, `description`, `image_url`
- `ingredients[]`, `steps[]`
- `category`, `difficulty`, `cooking_time`, `servings`
- `rating` (AI оценка 0-10)
- `is_public`, `price`, `purchases`

### Course
- `id` (uuid, PK)
- `title`, `description`, `image_url`
- `level` (1-10), `category`, `language`
- `duration`, `lessons_count`
- `instructor`, `stars`, `is_published`

### Lesson
- `id` (uuid, PK)
- `course_id` (uuid)
- `title`, `description`, `video_url`, `content`
- `order`, `duration`, `steps[]`

### QuizQuestion
- `id` (uuid, PK)
- `course_id` (uuid)
- `question`, `options[]`, `correct_answer` (index)
- `explanation`, `difficulty`, `language`

### Certificate
- `id` (uuid, PK)
- `user_id`, `course_id`
- `course_name`, `user_name`, `level`, `stars`
- `pdf_url`, `signature`, `issued_at`

### WalletTransaction
- `id` (uuid, PK)
- `user_id` (uuid)
- `amount` (+/-)
- `type` (reward, purchase, refund, bonus)
- `description`, `related_id`

---

## 🎯 Использование

### Сценарий 1: Ученик проходит курс

1. **GET** `/api/academy/courses?language=pl` - выбирает курс
2. **GET** `/api/academy/courses/{id}/lessons` - смотрит уроки
3. **GET** `/api/academy/lessons/{id}` - изучает урок
4. **GET** `/api/academy/quiz/{courseId}` - получает тест
5. **POST** `/api/academy/quiz/{courseId}/submit` - сдаёт тест → получает звёзды + ChefToken
6. **GET** `/api/user/{userId}/wallet` - проверяет баланс

### Сценарий 2: AI анализ рецепта

1. **POST** `/api/ai/analyze` - общий анализ рецепта (рейтинг, баланс)
2. **POST** `/api/mentor/analyze-step` - детальный анализ каждого шага
3. **POST** `/api/ai/estimate-price` - оценка стоимости

### Сценарий 3: Marketplace

1. **POST** `/api/user/{userId}/recipes` - создаёт рецепт
2. **PUT** `/api/user/{userId}/recipes/{id}` - устанавливает `isPublic: true, price: 25`
3. Другой пользователь покупает → создаётся `MarketPurchase`
4. Продавец получает ChefToken в `WalletTransaction`

---

## 📈 Statystyki

**Всего endpoints:** 75
- User Dashboard: 9 новых
- Culinary Academy: 6 новых
- AI Mentor: 1 новый (analyze-step)

**Всего таблиц:** 27
- 11 новых (UserProfile, PersonalRecipe, Certificate, etc.)

**AI Features:**
- Recipe Analyzer (рейтинг 0-10)
- Mentor Chat (PL/UA/EN)
- Price Estimator
- Step Analyzer (новый!)

---

## 🚀 Deployment

```bash
# Build
go build -o bin/server cmd/server/main.go

# Run
./bin/server

# Test
curl http://localhost:8080/api/academy/courses
```

---

**Created:** November 3, 2025  
**Version:** 2.0 (User Dashboard + Academy)  
**AI Engine:** Groq API (openai/gpt-oss-20b)  
**Image Storage:** Cloudinary (signed upload)
