# 📚 Complete API Endpoints Reference

**Culinary Academy & Marketplace Evolution v1.0**

Total Endpoints: **83** (including Leaderboard)

---

## 🔐 Authentication (2)

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/auth/register` | Регистрация пользователя |
| POST | `/api/auth/login` | Авторизация |

---

## 👨‍🍳 User Profile (10)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/user/{userId}/profile` | Получить профиль (auto-create) |
| PUT | `/api/user/{userId}/profile` | Обновить профиль |
| GET | `/api/user/{userId}/dashboard` | Kitchen Simulator Dashboard (9 разделов) |
| GET | `/api/user/{userId}/progress` | Прогресс по курсам |
| GET | `/api/user/{userId}/certificates` | Все сертификаты |
| GET | `/api/user/{userId}/wallet` | История ChefToken |
| GET | `/api/user/{userId}/recipes` | Личные рецепты |
| POST | `/api/user/{userId}/recipes` | Создать рецепт |
| DELETE | `/api/user/{userId}/recipes/{recipeId}` | Удалить рецепт |
| GET | `/api/user/{userId}/market/purchases` | История покупок (legacy) |

---

## 🎓 Culinary Academy (8)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/academy/courses` | Список курсов (filters: language, category) |
| GET | `/api/academy/courses/{courseId}` | Детали курса |
| GET | `/api/academy/courses/{courseId}/lessons` | Уроки курса |
| GET | `/api/academy/lessons/{lessonId}` | Детали урока |
| GET | `/api/academy/quiz/{courseId}` | Вопросы теста |
| POST | `/api/academy/quiz/{courseId}/submit` | Отправить ответы |
| POST | `/api/academy/certificate/{courseId}` | ✨ Генерация PDF-сертификата с AI |

---

## 🛒 Marketplace (5)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/market/recipes` | Все рецепты (filters: category, difficulty, price, rating, sortBy) |
| POST | `/api/market/purchase` | Покупка рецепта (10% комиссия) |
| GET | `/api/user/{userId}/purchases` | Купленные рецепты |
| GET | `/api/market/stats/{userId}` | Статистика продавца |
| GET | `/api/leaderboard` | ✨ Глобальный рейтинг поваров (sortBy: xp/sales/rating) |

---

## 🧠 AI Features (7)

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/ai/analyze` | Анализ рецепта с категоризацией |
| POST | `/api/ai/review-recipe` | ✨ AI оценка рецепта (rating + tips) |
| POST | `/api/ai/critique` | ✨ Глубокий анализ (5 критериев: taste, presentation, technique, creativity, health) |
| POST | `/api/ai/estimate-price` | Оценка стоимости ингредиентов |
| POST | `/api/mentor/chat` | Одиночный вопрос AI-ментору |
| POST | `/api/mentor/analyze-step` | Анализ шага приготовления |
| GET | `/api/mentor/history` | История сообщений сессии |

---

## 📡 WebSocket (3)

| Protocol | Endpoint | Description |
|----------|----------|-------------|
| WS | `/ws` | Admin order notifications |
| WS | `/ws/mentor?userId={uuid}&language=pl&topic=sushi` | ✨ AI Mentor Chat (real-time) |
| GET | `/api/user/{userId}/mentor/sessions` | Все mentor сессии |

---

## 📸 Image Upload (1)

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/upload/image` | Cloudinary signed upload (SHA1) |

---

## 🏪 Business Management (15)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/businesses` | Список всех бизнесов |
| POST | `/api/businesses` | Создать бизнес |
| GET | `/api/businesses/{id}` | Детали бизнеса |
| PUT | `/api/businesses/{id}` | Обновить бизнес |
| DELETE | `/api/businesses/{id}` | Удалить бизнес |
| GET | `/api/businesses/owner/{ownerId}` | Бизнесы владельца |
| GET | `/api/businesses/{id}/tokens` | Токены бизнеса |
| POST | `/api/businesses/{id}/tokens` | Создать токен |
| GET | `/api/businesses/{id}/subscriptions` | Подписки |
| POST | `/api/businesses/{id}/subscriptions` | Создать подписку |
| GET | `/api/businesses/{id}/menu` | Меню бизнеса |
| POST | `/api/businesses/{id}/menu/products` | Добавить продукт |
| PUT | `/api/businesses/{id}/menu/products/{productId}` | Обновить продукт |
| DELETE | `/api/businesses/{id}/menu/products/{productId}` | Удалить продукт |
| GET | `/api/metrics/{businessId}` | AI-метрики бизнеса |

---

## 🍱 Products & Ingredients (12)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/ingredients` | Список ингредиентов |
| POST | `/api/ingredients` | Создать ингредиент |
| GET | `/api/ingredients/{id}` | Детали ингредиента |
| PUT | `/api/ingredients/{id}` | Обновить ингредиент |
| DELETE | `/api/ingredients/{id}` | Удалить ингредиент |
| GET | `/api/products` | Список продуктов |
| POST | `/api/products` | Создать продукт |
| GET | `/api/products/{id}` | Детали продукта |
| PUT | `/api/products/{id}` | Обновить продукт |
| DELETE | `/api/products/{id}` | Удалить продукт |
| GET | `/api/semi-finished` | Полуфабрикаты |
| POST | `/api/semi-finished` | Создать полуфабрикат |

---

## 📦 Orders (6)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/orders` | Список заказов |
| POST | `/api/orders` | Создать заказ |
| GET | `/api/orders/{id}` | Детали заказа |
| PUT | `/api/orders/{id}` | Обновить заказ |
| DELETE | `/api/orders/{id}` | Удалить заказ |
| GET | `/api/orders/business/{businessId}` | Заказы бизнеса |

---

## 💳 Subscriptions & Tokens (8)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/subscriptions` | Все подписки |
| POST | `/api/subscriptions` | Создать подписку |
| GET | `/api/subscriptions/{id}` | Детали подписки |
| PUT | `/api/subscriptions/{id}` | Обновить подписку |
| DELETE | `/api/subscriptions/{id}` | Удалить подписку |
| GET | `/api/tokens` | Список токенов |
| POST | `/api/tokens/buy` | Купить токены |
| POST | `/api/tokens/burn` | Сжечь токены |

---

## 💰 Transactions (2)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/transactions` | История транзакций |
| GET | `/api/transactions/business/{businessId}` | Транзакции бизнеса |

---

## 📊 Statistics & Health (3)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/health` | Health check (database status) |
| GET | `/api/stats` | Общая статистика |
| GET | `/api/metrics/{businessId}` | AI-метрики бизнеса |

---

## 🎯 NEW in v1.0 (Marketplace Evolution)

### ✨ AI-Powered Features
- **AI Recipe Review** - автоматическая оценка рецептов с советами
- **AI Recipe Critique** - глубокий анализ по 5 критериям
- **AI Certificate Generator** - PDF с персонализированным сообщением от AI
- **AI Mentor WebSocket** - реальное время чат для обучения

### 💰 Marketplace Economy
- **Recipe Marketplace** - покупка/продажа рецептов за ChefToken
- **Комиссия 10%** - платформа берет 10%, продавец получает 90%
- **Seller Stats** - статистика продаж, доходов, рейтингов
- **Purchase History** - история покупок с полной информацией

### 🎓 Learning Platform
- **3 Test Courses** - Podstawy Sushi, Zaawansowane Techniki, Майстерність Ножа
- **Quiz System** - автоматические награды (stars → ChefToken → XP)
- **12 Achievements** - система достижений по 4 категориям
- **Chef Ranking** - динамические ранги по уровню (4 тира на язык)

### 📜 Certification System
- **PDF Generation** - gofpdf для создания сертификатов
- **AI Personalization** - уникальное сообщение для каждого студента
- **Premium Design** - золотая рамка, подпись шефа
- **Multi-language** - PL/UA/EN

---

## 🌍 Multi-Language Support

**Применяется к:**
- Course content (PL/UA/EN)
- Quiz questions
- AI responses
- Certificates
- Mentor chat
- Dashboard tips

**Default:** `pl` (Polski)

---

## 📈 Performance Stats

- **Total Endpoints:** 82
- **Database Tables:** 31
- **Active AI Modules:** 4
- **Supported Languages:** 3
- **Achievements:** 12
- **Test Data:** 3 courses, 5 lessons, 10 questions
- **WebSocket Hubs:** 2 (Admin + Mentor)

---

## 🔧 Query Parameters Summary

### Filters
- `language` - pl, ua, en
- `category` - sushi, soup, dessert, etc.
- `difficulty` - easy, medium, hard
- `maxPrice` - number
- `minRating` - 0-10
- `sortBy` - popular, newest, rating, price

### Pagination
- `limit` - max results (default varies by endpoint)
- `offset` - skip N records

---

## 🚀 Quick Start Examples

### 1. Complete Student Journey
```bash
# Register
curl -X POST http://localhost:8080/api/auth/register \
  -d '{"email":"student@test.pl","password":"pass","name":"Jan"}'

# Browse courses
curl http://localhost:8080/api/academy/courses?language=pl

# Create recipe
curl -X POST http://localhost:8080/api/user/{userId}/recipes \
  -d '{"title":"My Recipe","price":25,"isPublic":true}'

# AI Review
curl -X POST http://localhost:8080/api/ai/review-recipe \
  -d '{"recipeId":"uuid","language":"pl"}'

# Browse marketplace
curl 'http://localhost:8080/api/market/recipes?sortBy=rating'

# Check dashboard
curl http://localhost:8080/api/user/{userId}/dashboard
```

### 2. WebSocket AI Mentor
```javascript
const ws = new WebSocket('ws://localhost:8080/ws/mentor?userId=uuid&language=pl');

ws.onopen = () => console.log('Connected to AI Mentor');

ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  if (data.type === 'session_start') {
    console.log('Welcome:', data.content);
  } else if (data.type === 'ai_response') {
    console.log('AI:', data.content);
  }
};

// Ask question
ws.send(JSON.stringify({
  type: 'user_message',
  content: 'Jak przygotować ryż sushi?'
}));
```

### 3. Marketplace Transaction
```bash
# Browse
curl 'http://localhost:8080/api/market/recipes?category=sushi&minRating=8'

# Purchase
curl -X POST http://localhost:8080/api/market/purchase \
  -d '{"recipeId":"uuid","buyerId":"uuid"}'

# Check seller stats
curl http://localhost:8080/api/market/stats/{sellerId}
```

---

## 📝 Response Format

**Success:**
```json
{
  "status": "ok",
  "data": { ... }
}
```

**Error:**
```json
{
  "status": "error",
  "message": "Description"
}
```

---

**Created:** 2025-11-03  
**Version:** v1.0  
**Server:** http://localhost:8080  
**Documentation:** MARKETPLACE_EVOLUTION.md, USER_DASHBOARD_API.md
