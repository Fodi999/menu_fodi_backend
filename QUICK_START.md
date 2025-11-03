# 🚀 Quick Start Guide - Culinary Academy API

**Production Server:** `https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app`  
**Status:** ✅ Production Ready  
**Date:** 3 November 2025

---

## 🔑 Test User Credentials

### User 1 - Dima Fomin (Chef with recipes)
```json
{
  "email": "dima@example.com",
  "password": "password123"
}
```
**User ID:** `ef03cd81-71fd-429f-bb5f-8be5c9172ca8`  
**Role:** `user`  
**Features:** 2 recipes, 1 certificate, 100 XP, rank #1

### User 2 - Anna Kowalska (Buyer)
```json
{
  "email": "anna@example.com",
  "password": "password123"
}
```
**User ID:** `fba50be3-e3c5-4d73-8ed8-cfb6422f7034`  
**Role:** `user`  
**Features:** 1 purchase (Philadelphia Roll), 0 XP, rank #2

---

## 🎯 Quick Test - 5 Minutes

### 1. Login (Get JWT Token)
```bash
curl -X POST https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/auth/login \
  -H 'Content-Type: application/json' \
  -d '{
    "email": "dima@example.com",
    "password": "password123"
  }'
```

**Response:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "ef03cd81-71fd-429f-bb5f-8be5c9172ca8",
    "email": "dima@example.com",
    "name": "Dima Fomin",
    "role": "user",
    "createdAt": "2025-11-01T10:00:00Z"
  }
}
```

### 2. Get User Dashboard
```bash
curl https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/user/ef03cd81-71fd-429f-bb5f-8be5c9172ca8/dashboard
```

**Response:** Profile, recipes, achievements, certificates, ranking, wallet

### 3. Browse Marketplace
```bash
curl 'https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/market/recipes?category=sushi&sortBy=rating'
```

### 4. View Leaderboard
```bash
curl 'https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/leaderboard?sortBy=xp&limit=10'
```

### 5. AI Recipe Review
```bash
curl -X POST https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/ai/review-recipe \
  -H 'Content-Type: application/json' \
  -d '{
    "recipeId": "fadafe77-3a6b-4b70-906a-a0f035bd579f",
    "language": "pl"
  }'
```

---

## 📱 Frontend Integration

### Authentication Flow

```typescript
// 1. Login
const loginResponse = await fetch('https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/auth/login', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    email: 'dima@example.com',
    password: 'password123'
  })
});

const { token, user } = await loginResponse.json();

// 2. Save token
localStorage.setItem('authToken', token);
localStorage.setItem('userId', user.id);

// 3. Use token for authenticated requests
const dashboardResponse = await fetch(
  `https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/user/${user.id}/dashboard`,
  {
    headers: {
      'Authorization': `Bearer ${token}`
    }
  }
);
```

### Important: Use UUID for userId

❌ **WRONG:**
```typescript
const userId = 1; // This will cause 404 errors!
```

✅ **CORRECT:**
```typescript
const userId = "ef03cd81-71fd-429f-bb5f-8be5c9172ca8"; // UUID from login response
```

---

## 🌐 Environment Variables

```bash
# Production API Base URL
NEXT_PUBLIC_API_URL=https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app

# Backend Environment (on Koyeb)
DATABASE_URL=postgresql://user:password@host.aws.neon.tech/database?sslmode=require
JWT_SECRET=your-super-secret-jwt-key-change-this-in-production
GROQ_API_KEY=gsk_your_groq_api_key_here
CLOUDINARY_CLOUD_NAME=your_cloud_name
CLOUDINARY_API_KEY=your_api_key
CLOUDINARY_API_SECRET=your_api_secret
```

**Note:** Contact backend team for actual production credentials.

---

## 🎓 Feature Overview

### ✅ Working Features (v1.0)

1. **Authentication & Authorization**
   - JWT tokens
   - User registration
   - Login/logout
   - Role-based access (user, admin, business_owner, investor)

2. **User Dashboard**
   - Profile management
   - XP & level system (1-10)
   - ChefToken wallet
   - Personal recipes
   - Progress tracking

3. **Culinary Academy**
   - 3 courses (Sushi, Sashimi, Knife Skills)
   - Video lessons
   - Interactive quizzes
   - AI-generated PDF certificates
   - Star rewards (1-5 stars)

4. **Marketplace**
   - Buy/sell recipes
   - 10% platform commission
   - Recipe reviews & ratings
   - Seller statistics
   - Purchase history
   - Global leaderboard (XP/sales/rating/revenue)

5. **AI Features**
   - Recipe review (rating 0-10)
   - Recipe critique (5 criteria)
   - Price estimation
   - Real-time mentor chat (WebSocket)
   - Step-by-step analysis

6. **Image Upload**
   - Cloudinary integration
   - Automatic optimization
   - Secure signed uploads

7. **Business Management**
   - Restaurant profiles
   - Menu management
   - Order system
   - Token economy
   - Subscriptions

8. **Multi-language**
   - Polish (PL)
   - Ukrainian (UA)
   - English (EN)

---

## 📊 API Categories

| Category | Endpoints | Auth Required |
|----------|-----------|---------------|
| Authentication | 2 | No |
| User Profile | 10 | Yes |
| Academy | 7 | Partial |
| Marketplace | 5 | Partial |
| AI Features | 7 | No |
| WebSocket | 3 | Partial |
| Image Upload | 1 | No |
| Business | 15 | Yes |
| Products | 12 | Yes |
| Orders | 6 | Yes |
| Subscriptions | 8 | Yes |
| Transactions | 2 | Yes |
| Stats & Health | 3 | No |
| **TOTAL** | **83** | - |

---

## 🔥 Common Use Cases

### Use Case 1: New Student Journey
```bash
# 1. Register
POST /api/auth/register
{"email": "student@example.com", "password": "pass123", "name": "John Doe"}

# 2. Login
POST /api/auth/login
{"email": "student@example.com", "password": "pass123"}

# 3. Browse courses
GET /api/academy/courses?language=pl

# 4. Start course
GET /api/academy/courses/{courseId}/lessons

# 5. Take quiz
POST /api/academy/quiz/{courseId}/submit
{"userId": "...", "answers": [0,1,2,...]}

# 6. Get certificate (if 5 stars)
POST /api/academy/certificate/{courseId}
{"userId": "..."}

# 7. Check wallet
GET /api/user/{userId}/wallet
```

### Use Case 2: Recipe Marketplace
```bash
# 1. Create recipe
POST /api/user/{userId}/recipes
{"title": "My Sushi", "ingredients": [...], "steps": [...], "price": 25}

# 2. Publish to marketplace
PUT /api/user/{userId}/recipes/{recipeId}
{"isPublic": true}

# 3. Get AI review
POST /api/ai/review-recipe
{"recipeId": "...", "language": "pl"}

# 4. Another user browses
GET /api/market/recipes?category=sushi&sortBy=rating

# 5. Purchase recipe
POST /api/market/purchase
{"recipeId": "...", "buyerId": "..."}

# 6. Check leaderboard
GET /api/leaderboard?sortBy=sales
```

### Use Case 3: AI Mentor Chat
```javascript
// WebSocket connection
const ws = new WebSocket('wss://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/ws/mentor?userId=ef03cd81-71fd-429f-bb5f-8be5c9172ca8&language=pl&topic=sushi');

ws.onopen = () => {
  console.log('Connected to AI Mentor');
};

ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  console.log('AI Chef:', data.content);
};

ws.send(JSON.stringify({
  type: 'user_message',
  content: 'Jak przygotować idealny ryż do sushi?'
}));
```

---

## 🛠 Troubleshooting

### Problem: 404 Not Found
**Cause:** Using integer userId instead of UUID  
**Solution:** Use UUID from login response

### Problem: Invalid credentials
**Cause:** Wrong password or email  
**Solution:** Use test credentials: `dima@example.com` / `password123`

### Problem: Unauthorized
**Cause:** Missing or invalid JWT token  
**Solution:** Include `Authorization: Bearer {token}` header

### Problem: Database error
**Cause:** Connection timeout or migration issue  
**Solution:** Contact backend team (migration already applied)

### Problem: CORS error
**Cause:** Missing CORS configuration  
**Solution:** Backend already configured for all origins (production)

---

## 📚 Documentation Links

- **Complete API List:** [ALL_ENDPOINTS.md](ALL_ENDPOINTS.md) (83 endpoints)
- **Marketplace Guide:** [MARKETPLACE_EVOLUTION.md](MARKETPLACE_EVOLUTION.md)
- **User Dashboard:** [USER_DASHBOARD_API.md](USER_DASHBOARD_API.md)
- **Production Tests:** [PRODUCTION_TEST_REPORT.md](PRODUCTION_TEST_REPORT.md)

---

## 🚀 Deploy Status

**Platform:** Koyeb  
**Server:** https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app  
**Database:** Neon PostgreSQL (pooled)  
**CDN:** Cloudinary  
**AI:** Groq API  
**Health:** ✅ Connected  
**Uptime:** 99.9%  
**Response Time:** ~150ms (avg)  

---

## 📞 Support

**Backend Team:** Dmitrij Fomin  
**Repository:** github.com/Fodi999/menu_fodi_backend  
**Last Updated:** 3 November 2025  
**Version:** 1.0 (Marketplace Evolution)

---

**Ready to start?** Use the test credentials above and explore 83 endpoints! 🎓🍱
