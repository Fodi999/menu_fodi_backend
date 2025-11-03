# 🛒 Marketplace Evolution API

## Обзор

Marketplace Evolution - полноценная экономическая система с покупкой рецептов, AI-оценкой, статистикой продавцов и PDF-сертификатами.

**Реализованные модули:**
- ✅ AI Recipe Review - оценка рецептов с рейтингом 0-10
- ✅ Marketplace Feed - просмотр публичных рецептов с фильтрами
- ✅ Recipe Purchase - покупка рецептов за ChefToken (комиссия 10%)
- ✅ Seller Statistics - статистика продаж, дохода, топ-рецептов
- ✅ AI Certificate Generator - PDF-сертификаты с персонализацией

---

## 🧠 AI Recipe Review

### POST `/api/ai/review-recipe`
**AI оценивает рецепт ученика** с рейтингом, комментариями и рекомендациями.

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

---

## 🛒 Marketplace Feed

### GET `/api/market/recipes`
**Просмотр всех публичных рецептов** с фильтрацией и сортировкой.

**Query Parameters:**
- `category` - фильтр по категории (sushi, soup, dessert, ...)
- `difficulty` - easy, medium, hard
- `maxPrice` - максимальная цена в ChefToken
- `minRating` - минимальный AI-рейтинг (0-10)
- `sortBy` - popular, newest, rating, price

**Example:**
```bash
GET /api/market/recipes?category=sushi&minRating=7&sortBy=rating
```

**Response:**
```json
{
  "status": "ok",
  "data": {
    "recipes": [
      {
        "id": "fadafe77-3a6b-4b70-906a-a0f035bd579f",
        "title": "Philadelphia Roll Deluxe",
        "description": "Mój autorski przepis na Philadelphia",
        "category": "sushi",
        "difficulty": "medium",
        "cookingTime": 45,
        "servings": 4,
        "rating": 8.0,
        "price": 25,
        "purchases": 1,
        "authorName": "Dima Fomin",
        "authorLevel": 1,
        "authorAvatar": "",
        "reviewCount": 0,
        "avgReview": 0
      }
    ],
    "total": 1
  }
}
```

---

## 💸 Recipe Purchase

### POST `/api/market/purchase`
**Покупка рецепта за ChefToken** с автоматическими транзакциями.

**Request:**
```json
{
  "recipeId": "fadafe77-3a6b-4b70-906a-a0f035bd579f",
  "buyerId": "fba50be3-e3c5-4d73-8ed8-cfb6422f7034"
}
```

**Response:**
```json
{
  "status": "ok",
  "data": {
    "purchaseId": "12345-uuid",
    "recipe": "Philadelphia Roll Deluxe",
    "price": 25.0,
    "commission": 2.5,
    "sellerReceived": 22.5,
    "buyerBalance": 75.0
  }
}
```

**Экономическая модель:**
- 🛒 Покупатель платит полную цену
- 💰 Продавец получает 90% (комиссия платформы 10%)
- 📊 Автоматические транзакции в `WalletTransaction`
- 🔒 Проверка: нельзя купить свой рецепт
- ✅ Проверка: рецепт не куплен ранее

---

## 📊 Seller Statistics

### GET `/api/market/stats/{userId}`
**Статистика продавца на маркетплейсе.**

**Response:**
```json
{
  "status": "ok",
  "data": {
    "sellerId": "ef03cd81-71fd-429f-bb5f-8be5c9172ca8",
    "totalSales": 1,
    "totalRevenue": 22.5,
    "averageRating": 7.25,
    "topRecipe": "Philadelphia Roll Deluxe",
    "topRecipeSales": 1
  }
}
```

**Метрики:**
- `totalSales` - количество проданных рецептов
- `totalRevenue` - чистый доход (после комиссии 10%)
- `averageRating` - средний AI-рейтинг рецептов продавца
- `topRecipe` - самый популярный рецепт по продажам

---

## 📜 AI Certificate Generator

### POST `/api/academy/certificate/{courseId}`
**Генерация PDF-сертификата** с AI-персонализацией.

**Request:**
```json
{
  "userId": "ef03cd81-71fd-429f-bb5f-8be5c9172ca8"
}
```

**Response:**
```json
{
  "status": "ok",
  "data": {
    "certificateId": "604c880f-76ec-445c-b69a-21caa344ab94",
    "pdfUrl": "certificates/certificate_Dima_Fomin_1762171106.pdf",
    "studentName": "Dima Fomin",
    "courseName": "Podstawy Sushi - Kurs dla Początkujących",
    "issuedAt": "2025-11-03T12:58:26.306694+01:00",
    "message": "Gratulacje! Twoja pasja i determinacja są inspirujące."
  }
}
```

**Требования:**
- ✅ Курс должен быть завершён (`UserProgress.isCompleted = true`)
- ✅ Повторный запрос вернёт существующий сертификат

**Дизайн PDF:**
- 📄 Формат: A4 Landscape
- 🏆 Золотая рамка с двойной границей
- 🎨 AI-персонализированное сообщение от шеф-повара
- 📊 Уровень, % теста, количество звёзд
- ✍️ Подпись: "Chef Dima Fomin"
- 🌐 Многоязычность: PL, UA, EN

---

## 🎯 User Purchases

### GET `/api/user/{userId}/purchases`
**Список купленных рецептов пользователя.**

**Response:**
```json
{
  "status": "ok",
  "data": [
    {
      "purchaseId": "uuid",
      "recipeId": "fadafe77-3a6b-4b70-906a-a0f035bd579f",
      "title": "Philadelphia Roll Deluxe",
      "category": "sushi",
      "imageUrl": "https://...",
      "price": 25.0,
      "purchasedAt": "2025-11-03T12:55:00Z"
    }
  ]
}
```

---

## 📊 Database Schema

### RecipePurchase
```sql
CREATE TABLE "RecipePurchase" (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    recipe_id UUID NOT NULL REFERENCES "PersonalRecipe"(id),
    buyer_id UUID NOT NULL REFERENCES "User"(id),
    seller_id UUID NOT NULL REFERENCES "User"(id),
    price NUMERIC(10,2) NOT NULL,
    commission NUMERIC(10,2) DEFAULT 0,
    net_amount NUMERIC(10,2) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
```

### RecipeReview
```sql
CREATE TABLE "RecipeReview" (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    recipe_id UUID NOT NULL REFERENCES "PersonalRecipe"(id),
    user_id UUID NOT NULL REFERENCES "User"(id),
    rating NUMERIC(3,1) CHECK (rating >= 0 AND rating <= 10),
    comment TEXT,
    would_buy_again BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
```

---

## 🧪 Testing Examples

### 1. AI Review Recipe
```bash
curl -X POST http://localhost:8080/api/ai/review-recipe \
  -H 'Content-Type: application/json' \
  -d '{
    "recipeId": "fadafe77-3a6b-4b70-906a-a0f035bd579f",
    "language": "pl"
  }'
```

### 2. Browse Marketplace
```bash
# Все sushi рецепты с рейтингом 7+
curl 'http://localhost:8080/api/market/recipes?category=sushi&minRating=7&sortBy=rating'

# Дешёвые рецепты до 20 ChefToken
curl 'http://localhost:8080/api/market/recipes?maxPrice=20&sortBy=price'

# Новые рецепты
curl 'http://localhost:8080/api/market/recipes?sortBy=newest'
```

### 3. Purchase Recipe
```bash
curl -X POST http://localhost:8080/api/market/purchase \
  -H 'Content-Type: application/json' \
  -d '{
    "recipeId": "fadafe77-3a6b-4b70-906a-a0f035bd579f",
    "buyerId": "fba50be3-e3c5-4d73-8ed8-cfb6422f7034"
  }'
```

### 4. Generate Certificate
```bash
curl -X POST http://localhost:8080/api/academy/certificate/e37cb669-9bc3-4688-b723-5af965a57f20 \
  -H 'Content-Type: application/json' \
  -d '{
    "userId": "ef03cd81-71fd-429f-bb5f-8be5c9172ca8"
  }'
```

### 5. Check Seller Stats
```bash
curl http://localhost:8080/api/market/stats/ef03cd81-71fd-429f-bb5f-8be5c9172ca8
```

---

## 💡 Business Logic

### Recipe Purchase Flow
1. ✅ Проверка: рецепт существует и публичный
2. ✅ Проверка: покупатель ≠ продавец
3. ✅ Проверка: рецепт не куплен ранее
4. ✅ Проверка: достаточно ChefToken на балансе
5. 💰 Транзакция: buyer -= price
6. 💸 Транзакция: seller += (price * 0.9)
7. 📝 Создание RecipePurchase
8. 📊 Обновление счётчика `PersonalRecipe.purchases`
9. 💳 2 записи в WalletTransaction (purchase + sale)

### Certificate Generation Flow
1. ✅ Проверка: курс завершён (UserProgress.isCompleted)
2. 🔍 Проверка: сертификат не существует
3. 🧠 AI генерирует персонализированное сообщение
4. 📄 PDF создаётся с золотым дизайном
5. 💾 Сохранение в `certificates/` директорию
6. 📝 Запись в таблицу Certificate
7. 📊 Dashboard показывает сертификат

---

## 🚀 Production Ready

**Environment Variables:**
```bash
GROQ_API_KEY="gsk_..."  # для AI-персонализации
CLOUDINARY_*            # для загрузки PDF (опционально)
```

**Dependencies:**
```bash
go get github.com/jung-kurt/gofpdf  # PDF generation
```

**File Structure:**
```
backend/
├── certificates/           # Generated PDFs
├── internal/
│   ├── services/
│   │   └── certificate_service.go
│   ├── handlers/
│   │   ├── marketplace.go
│   │   └── academy.go
│   └── models/
│       └── marketplace.go
```

---

## 📈 Future Enhancements

**Planned Features:**
- [ ] **Chef Leaderboard** - GET /api/leaderboard
- [ ] **Recipe Reviews** - POST /api/market/review
- [ ] **Trending Recipes** - AI-powered recommendations
- [ ] **Certificate Upload** - Cloudinary integration
- [ ] **Email Delivery** - Send certificates via email
- [ ] **Social Sharing** - Share achievements

---

## ✅ Testing Report

| Feature | Endpoint | Status | Test Result |
|---------|----------|--------|-------------|
| AI Recipe Review | POST /api/ai/review-recipe | ✅ | Rating: 8/10 |
| Marketplace Feed | GET /api/market/recipes | ✅ | 2 recipes |
| Recipe Purchase | POST /api/market/purchase | ✅ | 25 ChefToken |
| Seller Stats | GET /api/market/stats/{id} | ✅ | Revenue: 22.5 |
| Certificate PDF | POST /api/academy/certificate/{id} | ✅ | 3 KB PDF |
| User Purchases | GET /api/user/{id}/purchases | ✅ | 1 purchase |

**Test Users:**
- **Seller**: ef03cd81-71fd-429f-bb5f-8be5c9172ca8 (Dima Fomin)
- **Buyer**: fba50be3-e3c5-4d73-8ed8-cfb6422f7034 (Anna Kowalska)

---

**Marketplace Evolution v1.0** 🎉
Developed with ❤️ by AI-Powered Culinary Academy
