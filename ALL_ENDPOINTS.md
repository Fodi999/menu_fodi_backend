# 🎓 Complete API Endpoints List - Culinary Academy

**Production Server:** `https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app`  
**Version:** Marketplace Evolution v1.0 + Recipe Feed  
**Total Endpoints:** 88 (83 + 5 new Recipe Feed endpoints)  
**Date:** 5 November 2025

---

## � Recipe Feed (5) - NEW! 🆕

### 1. Get All Recipes (Main Feed)
```http
GET /api/posts
```
**Response:** List of all recipes from all users with author details
```json
{
  "status": "success",
  "data": [
    {
      "id": "recipe-001",
      "title": "Fresh Salmon Nigiri",
      "description": "Autentyczne nigiri z łososiem",
      "imageUrl": "https://images.unsplash.com/...",
      "authorId": "ef03cd81-71fd-429f-bb5f-8be5c9172ca8",
      "author": {
        "id": "ef03cd81-71fd-429f-bb5f-8be5c9172ca8",
        "name": "Dima Fomin",
        "email": "dima@example.com"
      },
      "createdAt": "2025-11-05T11:36:50.584057+01:00"
    }
  ]
}
```

### 2. Get User Recipes (Profile)
```http
GET /api/users/{userId}/posts
```
**Response:** List of recipes posted by specific user

### 3. Create Recipe
```http
POST /api/recipes
Content-Type: application/json

{
  "title": "California Roll",
  "description": "Classic California roll with crab and avocado",
  "imageUrl": "https://images.unsplash.com/photo-1617196034796-ca11959d7f34?w=800",
  "authorId": "ef03cd81-71fd-429f-bb5f-8be5c9172ca8"
}
```

### 4. Update Recipe
```http
PUT /api/recipes/{recipeId}
Content-Type: application/json

{
  "title": "Premium California Roll",
  "description": "Luxury California roll with king crab"
}
```

### 5. Delete Recipe
```http
DELETE /api/recipes/{recipeId}
```

📚 **Full Documentation:** [RECIPE_FEED_API.md](RECIPE_FEED_API.md)

---

## �🔐 Authentication (2)

### 1. Register User
```http
POST /api/auth/register
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "securepass123",
  "name": "John Doe"
}
```

### 2. Login User
```http
POST /api/auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "securepass123"
}
```

**Response:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "userId": "ef03cd81-71fd-429f-bb5f-8be5c9172ca8"
}
```

---

## 👨‍🍳 User Profile (10)

### 3. Get User Profile
```http
GET /api/user/{userId}/profile
```

**Example:**
```bash
curl https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/user/ef03cd81-71fd-429f-bb5f-8be5c9172ca8/profile
```

### 4. Update User Profile
```http
PUT /api/user/{userId}/profile
Content-Type: application/json

{
  "name": "New Name",
  "avatarUrl": "https://...",
  "language": "pl"
}
```

### 5. Get User Dashboard ✨
```http
GET /api/user/{userId}/dashboard
```

**Returns:** Profile, courses, recipes, achievements, certificates, kitchen simulator, wallet, ranking

**Example:**
```bash
curl https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/user/ef03cd81-71fd-429f-bb5f-8be5c9172ca8/dashboard
```

### 6. Get User Progress
```http
GET /api/user/{userId}/progress
```

### 7. Get User Certificates
```http
GET /api/user/{userId}/certificates
```

### 8. Get User Wallet
```http
GET /api/user/{userId}/wallet
```

### 9. Get User Recipes
```http
GET /api/user/{userId}/recipes
```

### 10. Create User Recipe
```http
POST /api/user/{userId}/recipes
Content-Type: application/json

{
  "title": "My Sushi Roll",
  "description": "Unique recipe",
  "ingredients": ["Rice", "Salmon", "Nori"],
  "steps": ["Step 1", "Step 2"],
  "category": "sushi",
  "difficulty": "easy",
  "cookingTime": 30,
  "servings": 2
}
```

### 11. Delete User Recipe
```http
DELETE /api/user/{userId}/recipes/{recipeId}
```

### 12. Get User Market Purchases
```http
GET /api/user/{userId}/market/purchases
```

---

## �� Culinary Academy (7)

### 13. Get All Courses
```http
GET /api/academy/courses?language=pl&category=sushi
```

**Example:**
```bash
curl 'https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/academy/courses?language=pl'
```

### 14. Get Course Details
```http
GET /api/academy/courses/{courseId}
```

### 15. Get Course Lessons
```http
GET /api/academy/courses/{courseId}/lessons
```

### 16. Get Lesson Details
```http
GET /api/academy/lessons/{lessonId}
```

### 17. Get Quiz Questions
```http
GET /api/academy/quiz/{courseId}
```

**Returns:** 10 random questions (without correct answers)

### 18. Submit Quiz
```http
POST /api/academy/quiz/{courseId}/submit
Content-Type: application/json

{
  "userId": "ef03cd81-71fd-429f-bb5f-8be5c9172ca8",
  "answers": [0, 2, 1, 3, 0, 1, 2, 0, 3, 1]
}
```

**Returns:** Score, stars, ChefToken reward

### 19. Generate Certificate ✨
```http
POST /api/academy/certificate/{courseId}
Content-Type: application/json

{
  "userId": "ef03cd81-71fd-429f-bb5f-8be5c9172ca8"
}
```

**Returns:** PDF certificate with AI personalization

---

## 🛒 Marketplace (5)

### 20. Get Market Recipes ✨
```http
GET /api/market/recipes?category=sushi&minRating=7&sortBy=rating&maxPrice=30
```

**Query Parameters:**
- `category` - sushi, soup, dessert, etc.
- `difficulty` - easy, medium, hard
- `maxPrice` - maximum price in ChefToken
- `minRating` - minimum AI rating (0-10)
- `sortBy` - popular, newest, rating, price

**Example:**
```bash
curl 'https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/market/recipes?category=sushi&sortBy=rating'
```

### 21. Purchase Recipe ✨
```http
POST /api/market/purchase
Content-Type: application/json

{
  "recipeId": "fadafe77-3a6b-4b70-906a-a0f035bd579f",
  "buyerId": "fba50be3-e3c5-4d73-8ed8-cfb6422f7034"
}
```

**Commission:** 10% platform fee (seller receives 90%)

### 22. Get User Purchases
```http
GET /api/user/{userId}/purchases
```

**Example:**
```bash
curl https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/user/fba50be3-e3c5-4d73-8ed8-cfb6422f7034/purchases
```

### 23. Get Seller Statistics ✨
```http
GET /api/market/stats/{userId}
```

**Returns:** totalSales, totalRevenue, averageRating, topRecipe

**Example:**
```bash
curl https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/market/stats/ef03cd81-71fd-429f-bb5f-8be5c9172ca8
```

### 24. Global Chef Leaderboard 🏆 NEW
```http
GET /api/leaderboard?sortBy=xp&language=pl&limit=10
```

**Sort Options:**
- `xp` - by experience points (default)
- `sales` - by total recipe sales
- `rating` - by average AI rating
- `revenue` - by total ChefToken revenue

**Example:**
```bash
curl 'https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/leaderboard?sortBy=xp&limit=5'
```

**Response:**
```json
{
  "data": {
    "leaders": [
      {
        "rank": 1,
        "userId": "ef03cd81-71fd-429f-bb5f-8be5c9172ca8",
        "name": "Dima Fomin",
        "level": 1,
        "totalXp": 100,
        "totalSales": 1,
        "totalRevenue": 22.5,
        "averageRating": 7.25,
        "recipeCount": 2,
        "achievementCount": 2
      }
    ],
    "total": 2,
    "sortBy": "xp"
  }
}
```

---

## 🧠 AI Features (7)

### 25. Analyze Recipe
```http
POST /api/ai/analyze
Content-Type: application/json

{
  "title": "Philadelphia Roll",
  "ingredients": ["salmon", "cream cheese", "avocado"],
  "steps": ["Prepare rice", "Cut salmon", "Roll maki"],
  "language": "pl"
}
```

### 26. Review Recipe ✨
```http
POST /api/ai/review-recipe
Content-Type: application/json

{
  "recipeId": "fadafe77-3a6b-4b70-906a-a0f035bd579f",
  "language": "pl"
}
```

**Returns:** AI rating (0-10), chef comment, taste balance, improvements, estimated price

**Example:**
```bash
curl -X POST https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/ai/review-recipe \
  -H 'Content-Type: application/json' \
  -d '{"recipeId":"fadafe77-3a6b-4b70-906a-a0f035bd579f","language":"pl"}'
```

### 27. Critique Recipe ✨
```http
POST /api/ai/critique
Content-Type: application/json

{
  "recipeId": "fadafe77-3a6b-4b70-906a-a0f035bd579f",
  "language": "pl"
}
```

**5 Criteria:** Taste, Presentation, Technique, Creativity, Health

### 28. Estimate Price
```http
POST /api/ai/estimate-price
Content-Type: application/json

{
  "ingredients": ["salmon 200g", "cream cheese 100g"],
  "servings": 4,
  "difficulty": "medium"
}
```

### 29. Mentor Chat
```http
POST /api/mentor/chat
Content-Type: application/json

{
  "userId": "ef03cd81-71fd-429f-bb5f-8be5c9172ca8",
  "message": "How to cut salmon for sushi?",
  "language": "pl"
}
```

### 30. Analyze Step
```http
POST /api/mentor/analyze-step
Content-Type: application/json

{
  "step": "Cut salmon into thin slices",
  "language": "pl"
}
```

### 31. Get Mentor History
```http
GET /api/mentor/history?sessionId={sessionId}
```

---

## 📡 WebSocket (3)

### 32. Admin WebSocket
```
WS /ws
```

**For admin order notifications**

### 33. AI Mentor WebSocket ✨
```
WS /ws/mentor?userId={userId}&language=pl&topic=sushi
```

**Real-time AI chat with session management**

**Example:**
```javascript
const ws = new WebSocket('wss://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/ws/mentor?userId=ef03cd81-71fd-429f-bb5f-8be5c9172ca8&language=pl&topic=sushi');

ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  console.log('AI Chef:', data.content);
};

ws.send(JSON.stringify({
  type: 'user_message',
  content: 'How to make perfect sushi rice?'
}));
```

### 34. Get User Mentor Sessions
```http
GET /api/user/{userId}/mentor/sessions
```

**Returns:** Last 20 mentor chat sessions

---

## 📸 Image Upload (1)

### 35. Upload Image ✨
```http
POST /api/upload/image
Content-Type: application/json

{
  "imageUrl": "https://images.unsplash.com/photo-1579584425555-c3ce17fd4351?w=500"
}
```

**Or multipart:**
```http
POST /api/upload/image
Content-Type: multipart/form-data

image: [file]
```

**Example:**
```bash
curl -X POST https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/upload/image \
  -H 'Content-Type: application/json' \
  -d '{"imageUrl":"https://images.unsplash.com/photo-1579584425555-c3ce17fd4351?w=500"}'
```

**Response:**
```json
{
  "data": {
    "url": "https://res.cloudinary.com/dwrn0ohbp/image/upload/v1762173109/culinary-academy/b03bd1a7-e040-44ae-9165-67ef6ef51610.jpg",
    "publicId": "culinary-academy/b03bd1a7-e040-44ae-9165-67ef6ef51610",
    "width": 500,
    "height": 888,
    "format": "jpg",
    "bytes": 144642
  }
}
```

---

## 🏪 Business Management (15)

### 36. Get All Businesses
```http
GET /api/businesses?category=restaurant&city=Warsaw
```

### 37. Create Business
```http
POST /api/businesses
Content-Type: application/json

{
  "name": "Sushi Bar Tokyo",
  "ownerId": "ef03cd81-71fd-429f-bb5f-8be5c9172ca8",
  "category": "restaurant",
  "city": "Warsaw"
}
```

### 38. Get Business
```http
GET /api/businesses/{id}
```

### 39. Update Business
```http
PUT /api/businesses/{id}
```

### 40. Delete Business
```http
DELETE /api/businesses/{id}
```

### 41. Get Businesses by Owner
```http
GET /api/businesses/owner/{ownerId}
```

### 42. Get Business Tokens
```http
GET /api/businesses/{id}/tokens
```

### 43. Add Business Tokens
```http
POST /api/businesses/{id}/tokens
Content-Type: application/json

{
  "tokens": 1000,
  "reason": "Monthly subscription"
}
```

### 44. Get Business Subscriptions
```http
GET /api/businesses/{id}/subscriptions
```

### 45. Subscribe to Business
```http
POST /api/businesses/{id}/subscriptions
Content-Type: application/json

{
  "userId": "ef03cd81-71fd-429f-bb5f-8be5c9172ca8"
}
```

### 46. Get Business Menu
```http
GET /api/businesses/{id}/menu
```

### 47. Add Product to Menu
```http
POST /api/businesses/{id}/menu/products
Content-Type: application/json

{
  "name": "Philadelphia Roll",
  "price": 45.00,
  "category": "sushi"
}
```

### 48. Update Menu Product
```http
PUT /api/businesses/{id}/menu/products/{productId}
```

### 49. Delete Menu Product
```http
DELETE /api/businesses/{id}/menu/products/{productId}
```

### 50. Get Business Metrics
```http
GET /api/metrics/{businessId}
```

---

## 🍱 Products & Ingredients (12)

### 51. Get All Ingredients
```http
GET /api/ingredients
```

### 52. Create Ingredient
```http
POST /api/ingredients
Content-Type: application/json

{
  "name": "Salmon",
  "category": "fish",
  "unit": "g",
  "pricePerUnit": 0.05
}
```

### 53. Get Ingredient
```http
GET /api/ingredients/{id}
```

### 54. Update Ingredient
```http
PUT /api/ingredients/{id}
```

### 55. Delete Ingredient
```http
DELETE /api/ingredients/{id}
```

### 56. Get All Products
```http
GET /api/products
```

### 57. Create Product
```http
POST /api/products
```

### 58. Get Product
```http
GET /api/products/{id}
```

### 59. Update Product
```http
PUT /api/products/{id}
```

### 60. Delete Product
```http
DELETE /api/products/{id}
```

### 61. Get Semi-Finished Products
```http
GET /api/semi-finished
```

### 62. Create Semi-Finished Product
```http
POST /api/semi-finished
```

---

## 📦 Orders (6)

### 63. Get All Orders
```http
GET /api/orders
```

### 64. Create Order
```http
POST /api/orders
Content-Type: application/json

{
  "businessId": "uuid",
  "items": [
    {
      "productId": "uuid",
      "quantity": 2
    }
  ]
}
```

### 65. Get Order
```http
GET /api/orders/{id}
```

### 66. Update Order
```http
PUT /api/orders/{id}
Content-Type: application/json

{
  "status": "completed"
}
```

### 67. Delete Order
```http
DELETE /api/orders/{id}
```

### 68. Get Business Orders
```http
GET /api/orders/business/{businessId}
```

---

## 💳 Subscriptions & Tokens (8)

### 69. Get All Subscriptions
```http
GET /api/subscriptions
```

### 70. Create Subscription
```http
POST /api/subscriptions
Content-Type: application/json

{
  "name": "Premium Plan",
  "price": 99.99,
  "duration": 30
}
```

### 71. Get Subscription
```http
GET /api/subscriptions/{id}
```

### 72. Update Subscription
```http
PUT /api/subscriptions/{id}
```

### 73. Delete Subscription
```http
DELETE /api/subscriptions/{id}
```

### 74. Get Tokens
```http
GET /api/tokens
```

### 75. Buy Tokens
```http
POST /api/tokens/buy
Content-Type: application/json

{
  "amount": 1000,
  "paymentMethod": "card"
}
```

### 76. Burn Tokens
```http
POST /api/tokens/burn
Content-Type: application/json

{
  "amount": 100,
  "reason": "service_usage"
}
```

---

## 💰 Transactions (2)

### 77. Get All Transactions
```http
GET /api/transactions
```

### 78. Get Business Transactions
```http
GET /api/transactions/business/{businessId}
```

---

## 📊 Statistics & Health (3)

### 79. Health Check
```http
GET /api/health
```

**Example:**
```bash
curl https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/health
```

**Response:**
```json
{
  "status": "ok",
  "data": {
    "database": "connected",
    "service": "menu-fodifood-backend"
  }
}
```

### 80. Get Statistics
```http
GET /api/stats
```

### 81. Get Business Metrics
```http
GET /api/metrics/{businessId}
```

### 82-83. Subscription Management
```http
DELETE /api/businesses/{id}/unsubscribe
GET /api/businesses/{id}/subscribers
```

---

## 🎯 Quick Test Examples

### Test Health
```bash
curl https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/health
```

### Test Leaderboard
```bash
curl 'https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/leaderboard?sortBy=xp&limit=5'
```

### Test Marketplace
```bash
curl 'https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/market/recipes?category=sushi'
```

### Test Dashboard
```bash
curl https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/user/ef03cd81-71fd-429f-bb5f-8be5c9172ca8/dashboard
```

### Test AI Review
```bash
curl -X POST https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/ai/review-recipe \
  -H 'Content-Type: application/json' \
  -d '{"recipeId":"fadafe77-3a6b-4b70-906a-a0f035bd579f","language":"pl"}'
```

### Test Image Upload
```bash
curl -X POST https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/upload/image \
  -H 'Content-Type: application/json' \
  -d '{"imageUrl":"https://images.unsplash.com/photo-1579584425555-c3ce17fd4351?w=500"}'
```

---

## 🔑 Test User Credentials

**User 1 - Dima Fomin:**
```
userId: ef03cd81-71fd-429f-bb5f-8be5c9172ca8
email: dima@example.com
password: password123
level: 1
xp: 100
language: pl
```

**User 2 - Anna Kowalska:**
```
userId: fba50be3-e3c5-4d73-8ed8-cfb6422f7034
email: anna@example.com
password: password123
level: 1
xp: 0
language: pl
```

**Login Example:**
```bash
curl -X POST https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email": "dima@example.com", "password": "password123"}'
```

---

## 📚 Summary

| Category | Endpoints | New in v1.0 |
|----------|-----------|-------------|
| Authentication | 2 | - |
| User Profile | 10 | Dashboard ✨ |
| Academy | 7 | Certificate ✨ |
| Marketplace | 5 | Leaderboard 🏆 |
| AI Features | 7 | Review, Critique ✨ |
| WebSocket | 3 | Mentor Chat ✨ |
| Image Upload | 1 | Cloudinary ✨ |
| Business | 15 | - |
| Products | 12 | - |
| Orders | 6 | - |
| Subscriptions | 8 | - |
| Transactions | 2 | - |
| Stats & Health | 3 | - |
| **TOTAL** | **83** | **8 NEW** |

---

## 🚀 Features

✅ **Authentication & Authorization** - JWT tokens  
✅ **User Profiles** - Levels, XP, ChefToken wallet  
✅ **Culinary Academy** - Courses, lessons, quizzes, certificates  
✅ **AI Integration** - Recipe review, mentor chat, price estimation  
✅ **Marketplace** - Buy/sell recipes, 10% commission  
✅ **Leaderboard** - Global chef ranking (XP/sales/rating)  
✅ **Image Upload** - Cloudinary integration  
✅ **WebSocket** - Real-time AI chat  
✅ **Business Management** - Restaurants, menus, orders  
✅ **Multi-language** - Polish (PL), Ukrainian (UA), English (EN)  

---

**Production Server:** https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app  
**Documentation:** API_ENDPOINTS.md, MARKETPLACE_EVOLUTION.md, USER_DASHBOARD_API.md  
**Tested:** 3 November 2025 ✅  
**Status:** Production Ready 🚀
