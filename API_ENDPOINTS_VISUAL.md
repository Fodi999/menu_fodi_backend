# API Endpoints Visual Guide

```
┌─────────────────────────────────────────────────────────────────────────┐
│                       FODI FOOD API - ALL ENDPOINTS                      │
└─────────────────────────────────────────────────────────────────────────┘

────────────────────────────────────────────────────────────────────────────
 🔐 AUTHENTICATION (No Auth Required)
────────────────────────────────────────────────────────────────────────────

  POST /api/auth/login
  ├─ Input: { email, password }
  └─ Output: { token, user }
  
  POST /api/auth/register
  ├─ Input: { email, password, name }
  └─ Output: { token, user }
  
  POST /api/auth/refresh
  ├─ Input: { token }
  └─ Output: { token }

────────────────────────────────────────────────────────────────────────────
 🤖 AI & CHEF MENTOR (No Auth Required - except save-ingredients)
────────────────────────────────────────────────────────────────────────────

  POST /api/ai/chef-mentor
  ├─ No Auth
  ├─ Input: { message, language, currentRecipe, conversationHistory }
  └─ Output: { message, recipe, isComplete, suggestedActions }
  
  POST /api/ai/recipe-generator
  ├─ No Auth
  ├─ Input: { title, language }
  └─ Output: { recipe }
  
  ⭐ POST /api/ai/save-ingredients  [NEW FEATURE]
  ├─ Auth Required: ✅ YES (JWT)
  ├─ Input: { ingredients: [{name, amount, unit}] }
  └─ Output: { success, message, count }
  
  POST /api/ai/fridge-recommendations
  ├─ Auth Required: ✅ YES (JWT)
  ├─ Input: { cuisine, maxTime, language }
  └─ Output: { recommendations: [{recipeName, matchPercentage, ...}] }
  
  POST /api/ai/meal-plan
  ├─ Auth Required: ✅ YES (JWT)
  ├─ Input: { days, targetCalories, language }
  └─ Output: { plan: [{day, breakfast, lunch, dinner}], totalCalories }

────────────────────────────────────────────────────────────────────────────
 🧊 FRIDGE MANAGEMENT (All Require Auth)
────────────────────────────────────────────────────────────────────────────

  GET /api/fridge/
  ├─ Auth Required: ✅ YES
  ├─ Query: ?page=1&limit=20&available=true
  └─ Output: { data: [...items], pagination }
  
  GET /api/fridge/available
  ├─ Auth Required: ✅ YES
  └─ Output: { data: [...items] }
  
  POST /api/fridge/
  ├─ Auth Required: ✅ YES
  ├─ Input: { product, quantity, unit, category, available }
  └─ Output: { data: {...item} }
  
  PUT /api/fridge/{id}
  ├─ Auth Required: ✅ YES
  ├─ Input: { quantity, available, ... }
  └─ Output: { data: {...item} }
  
  DELETE /api/fridge/{id}
  ├─ Auth Required: ✅ YES
  └─ Output: { success, message }

────────────────────────────────────────────────────────────────────────────
 🏥 HEALTH & STATUS (No Auth Required)
────────────────────────────────────────────────────────────────────────────

  GET /health
  └─ Output: { status, uptime, timestamp }
  
  GET /api/status
  └─ Output: { status, version, services }


┌─────────────────────────────────────────────────────────────────────────┐
│                         AUTHENTICATION FLOW                              │
└─────────────────────────────────────────────────────────────────────────┘

  User App                          API Server
     │                                 │
     │ POST /api/auth/login            │
     │ {email, password}               │
     ├────────────────────────────────>│
     │                                 │ ✓ Validate
     │                                 │ ✓ Generate JWT
     │                                 │
     │ {token: "jwt..."}               │
     │<────────────────────────────────┤
     │                                 │
     │ Save token in localStorage      │
     │                                 │
     │ GET /api/fridge/                │
     │ Authorization: Bearer {token}   │
     ├────────────────────────────────>│
     │                                 │ ✓ Validate token
     │                                 │ ✓ Extract user_id
     │                                 │ ✓ Return data
     │ [...fridge items...]            │
     │<────────────────────────────────┤
     │                                 │


┌─────────────────────────────────────────────────────────────────────────┐
│               CHAT → FRIDGE WORKFLOW (New Feature)                       │
└─────────────────────────────────────────────────────────────────────────┘

  Frontend UI                  API Server
     │                            │
     ├─ User: "Make pasta"        │
     │                            │
     ├─ POST /chef-mentor ───────>│
     │                            │ AI generates response
     │ Response: AI message       │
     │           + recipe data    │
     │<─────────────────────────── ├─ Show next question
     │                            │
     ├─ User: "I have eggs..."    │
     │                            │
     ├─ POST /chef-mentor ───────>│
     │ (with updated recipe)      │
     │                            │ AI updates recipe
     │ Response: isComplete=true  │
     │           suggestedActions │
     │<─────────────────────────── ├─ Show action buttons
     │                            │
     ├─ User clicks "Save Fridge" │
     │                            │
     ├─ POST /save-ingredients ──>│
     │ (JWT + ingredients)        │
     │                            │ ✓ Validate JWT
     │ {success, count: 4}        │ ✓ Create records
     │<─────────────────────────── │ ✓ Return result
     │                            │
     ├─ Show "4 saved!"           │
     │ Refresh fridge view        │
     │                            │
     ├─ GET /api/fridge/ ───────>│
     │ (JWT)                      │
     │                            │ ✓ Return user items
     │ [...items...]              │
     │<─────────────────────────── │
     │                            │


┌─────────────────────────────────────────────────────────────────────────┐
│                    ERROR HANDLING GUIDE                                  │
└─────────────────────────────────────────────────────────────────────────┘

  Status Code          Error                       Solution
  ───────────────────────────────────────────────────────────────────────
  400 Bad Request      Invalid input              Check JSON format
                       Missing fields             Add required fields
                       Empty array                Ensure ingredients []
  
  401 Unauthorized     No token                   Login first
                       Invalid token             Refresh token
                       Expired token             Get new token
  
  403 Forbidden        Insufficient permissions  User doesn't own resource
  
  404 Not Found        Resource doesn't exist    Check ID/parameter
  
  500 Server Error     Database error            Retry later
                       Service unavailable       Contact support


┌─────────────────────────────────────────────────────────────────────────┐
│                     ENDPOINT SUMMARY TABLE                               │
└─────────────────────────────────────────────────────────────────────────┘

  #  Endpoint                      Method  Auth  Purpose
  ──────────────────────────────────────────────────────────────────────
  1  /api/auth/login               POST    ❌   Login user
  2  /api/auth/register            POST    ❌   Register user
  3  /api/auth/refresh             POST    ✅   Refresh token
  
  4  /api/ai/chef-mentor           POST    ❌   Chat with AI
  5  /api/ai/recipe-generator      POST    ❌   Generate recipe
  6  /api/ai/save-ingredients      POST    ✅   Save to fridge ⭐
  7  /api/ai/fridge-recommendations POST    ✅   Get recommendations
  8  /api/ai/meal-plan             POST    ✅   Generate meal plan
  
  9  /api/fridge/                  GET     ✅   Get all items
  10 /api/fridge/available         GET     ✅   Get available items
  11 /api/fridge/                  POST    ✅   Add item
  12 /api/fridge/{id}              PUT     ✅   Update item
  13 /api/fridge/{id}              DELETE  ✅   Delete item
  
  14 /health                       GET     ❌   Health check
  15 /api/status                   GET     ❌   API status


┌─────────────────────────────────────────────────────────────────────────┐
│                    REQUEST/RESPONSE EXAMPLES                             │
└─────────────────────────────────────────────────────────────────────────┘

  1. LOGIN
  ─────────────────────────────────────────────────────────────────────
  POST /api/auth/login
  
  Request:
    {
      "email": "user@example.com",
      "password": "password123"
    }
  
  Response (200):
    {
      "success": true,
      "data": {
        "token": "eyJhbGciOiJIUzI1NiIs...",
        "user": {"id": "uuid", "email": "user@example.com"}
      }
    }


  2. CHAT WITH CHEF
  ─────────────────────────────────────────────────────────────────────
  POST /api/ai/chef-mentor
  
  Request:
    {
      "message": "I want to make pasta carbonara",
      "language": "en",
      "currentRecipe": null,
      "conversationHistory": []
    }
  
  Response (200):
    {
      "message": "Great! Let me help...",
      "recipe": {
        "title": "Pasta Carbonara",
        "ingredients": [...],
        "steps": []
      },
      "isComplete": false,
      "suggestedActions": null
    }


  3. SAVE INGREDIENTS ⭐ NEW
  ─────────────────────────────────────────────────────────────────────
  POST /api/ai/save-ingredients
  Authorization: Bearer {token}
  
  Request:
    {
      "ingredients": [
        {"name": "Pasta", "amount": 400, "unit": "g"},
        {"name": "Eggs", "amount": 3, "unit": "pcs"},
        {"name": "Bacon", "amount": 200, "unit": "g"}
      ]
    }
  
  Response (200):
    {
      "success": true,
      "message": "ingredients saved to fridge",
      "count": 3
    }


  4. GET FRIDGE ITEMS
  ─────────────────────────────────────────────────────────────────────
  GET /api/fridge/
  Authorization: Bearer {token}
  
  Response (200):
    {
      "success": true,
      "data": [
        {
          "id": "uuid",
          "product": "Pasta",
          "quantity": 400,
          "unit": "g",
          "available": true,
          "addedAt": "2024-11-10T12:34:56Z"
        },
        ...
      ],
      "pagination": {
        "page": 1,
        "limit": 20,
        "total": 3
      }
    }


┌─────────────────────────────────────────────────────────────────────┐
│                         QUICK REFERENCE                              │
└─────────────────────────────────────────────────────────────────────┘

Headers:
  Content-Type: application/json
  Authorization: Bearer {token}  (for protected endpoints)

Base URL:
  https://api.example.com

Token Storage:
  localStorage.setItem('token', response.data.token)
  
Token Usage:
  headers: {
    'Authorization': `Bearer ${localStorage.getItem('token')}`
  }

Common Units:
  g, kg, mg, ml, l, cl, pcs, dozen, tbsp, tsp, cup
```

---

## Endpoint Groups

### 📊 Statistics
- **Total Endpoints**: 15
- **Requires Auth**: 8 endpoints
- **No Auth Required**: 7 endpoints
- **New Endpoints**: 1 (save-ingredients) ⭐

### 🔄 Most Used Workflow
1. Login → Get token
2. Chat with Chef → Build recipe
3. Save ingredients → New feature!
4. View fridge → See saved items
5. Get recommendations → Plan meals

### ⚡ Performance Tips
- Cache token in localStorage
- Implement request debouncing for chat
- Load fridge items with pagination
- Show loading states during requests

### 🛡️ Security Best Practices
- Never expose token in logs
- Always use HTTPS in production
- Implement token refresh before expiry
- Clear token on logout
- Validate all inputs on frontend
