# 📚 Backend API Contract - Complete Reference

> **Last Updated:** 2026-01-11  
> **Backend:** Go (Chi Router)  
> **Base URL:** `http://localhost:8080/api` (dev) | `https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api` (production)  
> **Response Format:** Unified (v1) - All responses follow the same structure

---

## 🎯 Backend as Single Source of Truth

### Core Principles
- ❌ Frontend does NOT know database structure
- ❌ Frontend does NOT guess API behavior
- ✅ Backend validates ALL inputs
- ✅ Backend handles ALL localization
- ✅ Backend enforces ALL business logic
- ✅ Backend returns unified, predictable responses

---

## 📦 Unified Response Format

### All Success Responses
```json
{
  "data": { ... },
  "error": null,
  "meta": {
    "request_id": "9f4c7a8b-1234-5678-90ab-cdef12345678",
    "timestamp": "2026-01-11T10:30:00Z",
    "version": "v1"
  }
}
```

### All Error Responses
```json
{
  "data": null,
  "error": {
    "code": "INGREDIENT_NOT_FOUND",
    "message": "Ingredient not found",
    "details": "Ingredient with ID 'abc123' does not exist"
  },
  "meta": {
    "request_id": "9f4c7a8b-1234-5678-90ab-cdef12345678",
    "timestamp": "2026-01-11T10:30:00Z"
  }
}
```

### Paginated Responses
```json
{
  "data": {
    "items": [...],
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 100,
      "total_pages": 5
    }
  },
  "error": null,
  "meta": {
    "request_id": "9f4c7a8b-1234-5678-90ab-cdef12345678",
    "timestamp": "2026-01-11T10:30:00Z"
  }
}
```

---

## 🆔 Request ID Tracking

### Client-Provided Request ID
```bash
curl -H "X-Request-ID: my-custom-id-123" http://localhost:8080/api/ingredients
```

### Server-Generated Request ID
If no `X-Request-ID` header is provided, server generates UUID automatically.

### Response Header
Every response includes:
```
X-Request-ID: 9f4c7a8b-1234-5678-90ab-cdef12345678
```

### Benefits
- Track request through entire flow (frontend → backend → database → logs)
- Correlate errors with Sentry
- Debug production issues faster

---

## 🔐 Authentication

All protected endpoints require JWT token in header:
```
Authorization: Bearer <jwt_token>
```

### Roles
- **user**: Regular user
- **admin**: Administrator (can manage users, recipes, ingredients)
- **super_admin**: Super administrator (can change roles, delete users)

---

## 📋 Table of Contents

1. [Authentication](#1-authentication)
2. [Admin Panel](#2-admin-panel)
3. [User Management](#3-user-management)
4. [Fridge Management](#4-fridge-management)
5. [Recipe Catalog](#5-recipe-catalog)
6. [Ingredients](#6-ingredients)
7. [AI Services](#7-ai-services)
8. [Token Economy](#8-token-economy)
9. [Academy](#9-academy)
10. [Health & Stats](#10-health--stats)

---

## 1. Authentication

### POST `/api/auth/register`
Register new user

**Request:**
```json
{
  "email": "user@example.com",
  "password": "password123",
  "name": "John Doe"
}
```

**Response:** `201 Created`
```json
{
  "success": true,
  "data": {
    "token": "eyJhbGci...",
    "user": {
      "id": "uuid",
      "email": "user@example.com",
      "name": "John Doe",
      "role": "user"
    }
  }
}
```

---

### POST `/api/auth/login`
Login user

**Request:**
```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

**Response:** `200 OK`
```json
{
  "success": true,
  "data": {
    "token": "eyJhbGci...",
    "user": {
      "id": "uuid",
      "email": "user@example.com",
      "name": "John Doe",
      "role": "user"
    },
    "message": "Login successful"
  }
}
```

---

### POST `/api/auth/verify`
Verify JWT token validity

**Request:**
```json
{
  "token": "eyJhbGci..."
}
```

**Response:** `200 OK`
```json
{
  "valid": true,
  "user": {
    "id": "uuid",
    "email": "user@example.com",
    "role": "user"
  }
}
```

---

### GET `/api/auth/me`
Get current user profile

**Headers:**
```
Authorization: Bearer <token>
```

**Response:** `200 OK`
```json
{
  "id": "uuid",
  "email": "user@example.com",
  "name": "John Doe",
  "role": "user",
  "createdAt": "2025-01-01T00:00:00Z"
}
```

---

## 2. Admin Panel

**All endpoints require `admin` or `super_admin` role**

### Users

#### GET `/api/admin/users`
Get all users

**Query Params:**
- `page` (optional): Page number (default: 1)
- `limit` (optional): Items per page (default: 50)

**Response:** `200 OK`
```json
{
  "data": [
    {
      "id": "uuid",
      "email": "user@example.com",
      "name": "John Doe",
      "role": "user",
      "createdAt": "2025-01-01T00:00:00Z"
    }
  ],
  "total": 100,
  "page": 1,
  "limit": 50
}
```

---

#### GET `/api/admin/users/stats`
Get user statistics

**Response:** `200 OK`
```json
{
  "totalUsers": 1000,
  "activeUsers": 750,
  "newUsersThisMonth": 50,
  "usersByRole": {
    "user": 980,
    "admin": 15,
    "super_admin": 5
  }
}
```

---

#### PUT `/api/admin/users/{id}`
Update user

**Request:**
```json
{
  "name": "New Name",
  "email": "newemail@example.com"
}
```

**Response:** `200 OK`
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "name": "New Name",
    "email": "newemail@example.com"
  }
}
```

---

#### PATCH `/api/admin/users/update-role` ⚠️ **super_admin only**
Change user role

**Request:**
```json
{
  "userId": "uuid",
  "newRole": "admin"
}
```

**Response:** `200 OK`
```json
{
  "success": true,
  "message": "User role updated successfully"
}
```

---

#### DELETE `/api/admin/users/{id}` ⚠️ **super_admin only**
Delete user

**Response:** `200 OK`
```json
{
  "success": true,
  "message": "User deleted successfully"
}
```

---

### Ingredients

#### GET `/api/admin/ingredients`
Get all ingredients

**Query Params:**
- `page` (optional): Page number (default: 1)
- `limit` (optional): Items per page (default: 50)

**Headers:**
```
Accept-Language: pl|en|ru (optional, default: en)
```

**Response:** `200 OK`
```json
{
  "data": [
    {
      "id": "uuid",
      "name": "Tomato",
      "namePl": "Pomidor",
      "nameEn": "Tomato",
      "nameRu": "Помидор",
      "category": "vegetable",
      "nutritionGroup": "vegetable",
      "unit": "g"
    }
  ]
}
```

---

#### GET `/api/admin/ingredients/suggest`
🔥 **Autocomplete search** (fast, no AI)

**Query Params:**
- `q` (required): Search query (min 2 chars)
- `limit` (optional): Max results (default: 10)

**Headers:**
```
Accept-Language: pl|en|ru (optional, default: en)
```

**Response:** `200 OK`
```json
{
  "data": [
    {
      "id": "uuid",
      "name": "Tomato",
      "namePl": "Pomidor",
      "nameEn": "Tomato",
      "nameRu": "Помидор",
      "category": "vegetable",
      "unit": "g"
    }
  ]
}
```

**Example:**
```bash
curl "http://localhost:8080/api/admin/ingredients/suggest?q=tom&limit=5" \
  -H "Authorization: Bearer <token>" \
  -H "Accept-Language: pl"
```

---

#### GET `/api/admin/ingredients/stats`
Get ingredient statistics

**Response:** `200 OK`
```json
{
  "totalIngredients": 500,
  "byCategory": {
    "vegetable": 150,
    "fruit": 80,
    "protein": 100
  }
}
```

---

#### POST `/api/admin/ingredients`
Create ingredient (with AI classification)

**Request:**
```json
{
  "name": "Tomato",
  "namePl": "Pomidor",
  "nameEn": "Tomato",
  "nameRu": "Помидор",
  "unit": "g"
}
```

**Response:** `201 Created`
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "name": "Tomato",
    "category": "vegetable",
    "nutritionGroup": "vegetable"
  }
}
```

---

#### POST `/api/admin/ingredients/hint`
Get AI hint for ingredient classification

**Request:**
```json
{
  "name": "Tomato"
}
```

**Response:** `200 OK`
```json
{
  "category": "vegetable",
  "nutritionGroup": "vegetable",
  "confidence": 0.95
}
```

---

### Recipes

#### GET `/api/admin/recipes`
Get all recipes

**Query Params:**
- `page` (optional): Page number (default: 1)
- `limit` (optional): Items per page (default: 50)

**Response:** `200 OK`
```json
{
  "data": [
    {
      "id": "uuid",
      "title": "Tomato Soup",
      "difficulty": "easy",
      "timeMinutes": 30,
      "servings": 4,
      "descriptionPl": "...",
      "descriptionEn": "...",
      "descriptionRu": "..."
    }
  ]
}
```

---

#### GET `/api/admin/recipes/stats`
Get recipe statistics

**Response:** `200 OK`
```json
{
  "totalRecipes": 200,
  "byDifficulty": {
    "easy": 100,
    "medium": 70,
    "hard": 30
  }
}
```

---

### 🤖 AI Recipe Creation

#### POST `/api/admin/recipes/preview-ai`
Preview AI-generated recipe (no save)

**Request:**
```json
{
  "title": "Grilled Salmon with Rice",
  "language": "en",
  "ingredients": [
    {
      "ingredientId": "uuid",
      "quantity": 200,
      "amount": 200,
      "unit": "g"
    }
  ],
  "rawCookingText": "Grill salmon for 5 minutes on each side. Cook rice..."
}
```

**Response:** `200 OK`
```json
{
  "data": {
    "title": "Grilled Salmon with Rice",
    "language": "en",
    "description": "A delicious and healthy dish...",
    "servings": 2,
    "time_minutes": 25,
    "difficulty": "easy",
    "calories": 520,
    "steps": [
      {
        "order": 1,
        "text": "Grill salmon for 5 minutes",
        "time": 5
      }
    ],
    "ingredients": [
      {
        "ingredientId": "uuid",
        "name": "Salmon",
        "amount": 200,
        "unit": "g"
      }
    ]
  },
  "message": "Recipe preview generated",
  "success": true
}
```

---

#### POST `/api/admin/recipes/create-ai`
Create recipe via AI and save to database

**Request:** Same as preview

**Response:** `201 Created`
```json
{
  "data": {
    "id": "uuid",
    "title": "Grilled Salmon with Rice",
    "difficulty": "easy",
    "timeMinutes": 25,
    "servings": 2,
    "descriptionEn": "A delicious and healthy dish..."
  },
  "success": true
}
```

---

### Token Economy

#### GET `/api/admin/token-bank`
Get all user token banks

**Response:** `200 OK`
```json
{
  "data": [
    {
      "userId": "uuid",
      "balance": 100,
      "totalEarned": 150,
      "totalSpent": 50
    }
  ]
}
```

---

#### GET `/api/admin/token-bank/{userID}`
Get specific user token bank

**Response:** `200 OK`
```json
{
  "userId": "uuid",
  "balance": 100,
  "totalEarned": 150,
  "totalSpent": 50
}
```

---

#### POST `/api/admin/token-bank/allocate`
Allocate tokens to user

**Request:**
```json
{
  "userId": "uuid",
  "amount": 50,
  "reason": "Bonus reward"
}
```

**Response:** `200 OK`
```json
{
  "success": true,
  "newBalance": 150
}
```

---

#### POST `/api/admin/token-bank/revoke`
Revoke tokens from user

**Request:**
```json
{
  "userId": "uuid",
  "amount": 20,
  "reason": "Penalty"
}
```

**Response:** `200 OK`
```json
{
  "success": true,
  "newBalance": 130
}
```

---

### Treasury

#### GET `/api/public/treasury`
Get public treasury info (no auth required)

**Response:** `200 OK`
```json
{
  "totalTokens": 1000000,
  "allocatedTokens": 50000,
  "availableTokens": 950000
}
```

---

#### GET `/api/admin/treasury`
Get detailed treasury info

**Response:** `200 OK`
```json
{
  "totalTokens": 1000000,
  "allocatedTokens": 50000,
  "availableTokens": 950000,
  "reservedTokens": 100000
}
```

---

#### POST `/api/admin/treasury/allocate`
Allocate tokens from treasury

**Request:**
```json
{
  "amount": 1000,
  "purpose": "Monthly rewards"
}
```

**Response:** `200 OK`
```json
{
  "success": true,
  "newTreasuryBalance": 949000
}
```

---

## 3. User Management

### GET `/api/user/profile`
Get current user profile

**Headers:**
```
Authorization: Bearer <token>
```

**Response:** `200 OK`
```json
{
  "id": "uuid",
  "email": "user@example.com",
  "name": "John Doe",
  "role": "user",
  "tokenBalance": 100
}
```

---

### PUT `/api/user/profile`
Update current user profile

**Request:**
```json
{
  "name": "New Name",
  "email": "newemail@example.com"
}
```

**Response:** `200 OK`
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "name": "New Name",
    "email": "newemail@example.com"
  }
}
```

---

## 4. Fridge Management

**All endpoints require authentication**

### GET `/api/fridge/items`
Get user's fridge items

**Response:** `200 OK`
```json
{
  "data": [
    {
      "id": "uuid",
      "ingredientId": "uuid",
      "ingredientName": "Tomato",
      "quantity": 500,
      "unit": "g",
      "expiresAt": "2026-01-15T00:00:00Z",
      "status": "fresh",
      "price": 2.50,
      "totalValue": 1.25
    }
  ]
}
```

---

### POST `/api/fridge/items`
Add item to fridge

**Request:**
```json
{
  "ingredientId": "uuid",
  "quantity": 500,
  "unit": "g",
  "expiresAt": "2026-01-15T00:00:00Z",
  "price": 2.50
}
```

**Response:** `201 Created`
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "ingredientId": "uuid",
    "quantity": 500,
    "unit": "g",
    "price": 2.50
  }
}
```

---

### PATCH `/api/fridge/items/{id}`
Update fridge item quantity

**Request:**
```json
{
  "quantity": 300,
  "expiresAt": "2026-01-20T00:00:00Z"
}
```

**Response:** `200 OK`
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "quantity": 300
  }
}
```

---

### DELETE `/api/fridge/items/{id}`
Delete fridge item

**Response:** `200 OK`
```json
{
  "success": true,
  "message": "Item deleted"
}
```

---

### POST `/api/fridge/items/{id}/price`
Add price event (event sourcing for price history)

**Request:**
```json
{
  "price": 3.00,
  "quantity": 500,
  "eventType": "purchase"
}
```

**Response:** `201 Created`
```json
{
  "success": true,
  "message": "Price event added"
}
```

---

### GET `/api/fridge/items/{id}/price/history`
Get price history for item

**Response:** `200 OK`
```json
{
  "data": [
    {
      "id": "uuid",
      "price": 2.50,
      "quantity": 500,
      "eventType": "purchase",
      "createdAt": "2026-01-01T10:00:00Z"
    },
    {
      "id": "uuid",
      "price": 3.00,
      "quantity": 500,
      "eventType": "price_change",
      "createdAt": "2026-01-05T14:30:00Z"
    }
  ]
}
```

---

### POST `/api/fridge/add-missing`
Add missing ingredients from recipe to fridge

**Request:**
```json
{
  "recipeId": "uuid",
  "servings": 2
}
```

**Response:** `200 OK`
```json
{
  "success": true,
  "data": {
    "addedIngredients": [
      {
        "ingredientId": "uuid",
        "name": "Tomato",
        "quantity": 200,
        "unit": "g"
      }
    ],
    "totalCost": 5.50
  }
}
```

---

## 5. Recipe Catalog

### GET `/api/recipes`
Get public recipes (catalog) with filters

**Query Params:**
- `page` (optional): Page number (default: 1)
- `limit` (optional): Items per page (default: 20)
- `difficulty` (optional): easy|medium|hard
- `maxTime` (optional): Max cooking time in minutes
- `category` (optional): breakfast|lunch|dinner|dessert
- `diet` (optional): vegetarian|vegan|gluten-free
- `search` (optional): Search by title

**Response:** `200 OK`
```json
{
  "data": [
    {
      "id": "uuid",
      "title": "Tomato Soup",
      "difficulty": "easy",
      "timeMinutes": 30,
      "servings": 4,
      "calories": 200,
      "category": "lunch",
      "matchPercentage": 85
    }
  ],
  "total": 100,
  "page": 1,
  "limit": 20
}
```

---

### GET `/api/recipes/stats`
Get recipe statistics (public)

**Response:** `200 OK`
```json
{
  "totalRecipes": 500,
  "byDifficulty": {
    "easy": 200,
    "medium": 200,
    "hard": 100
  },
  "byCategory": {
    "breakfast": 100,
    "lunch": 150,
    "dinner": 150,
    "dessert": 100
  }
}
```

---

### GET `/api/recipes/{id}`
Get recipe details with optional fridge matching

**Headers:**
```
Authorization: Bearer <token> (optional)
Accept-Language: pl|en|ru (optional)
```

**Response:** `200 OK`
```json
{
  "id": "uuid",
  "title": "Tomato Soup",
  "titlePl": "Zupa pomidorowa",
  "titleEn": "Tomato Soup",
  "titleRu": "Томатный суп",
  "description": "...",
  "difficulty": "easy",
  "timeMinutes": 30,
  "servings": 4,
  "calories": 200,
  "category": "lunch",
  "steps": [
    {
      "order": 1,
      "textPl": "Pokrój pomidory",
      "textEn": "Chop tomatoes",
      "textRu": "Нарежьте помидоры",
      "time": 5
    }
  ],
  "ingredients": [
    {
      "ingredientId": "uuid",
      "name": "Tomato",
      "quantity": 500,
      "unit": "g",
      "inFridge": true,
      "availableQuantity": 600
    }
  ],
  "matchPercentage": 85,
  "missingIngredients": [
    {
      "ingredientId": "uuid",
      "name": "Onion",
      "quantity": 100,
      "unit": "g"
    }
  ]
}
```

---

### POST `/api/recipes/{id}/view`
Increment recipe view count (public)

**Response:** `200 OK`
```json
{
  "success": true,
  "views": 101
}
```

---

### 🔍 GET `/api/recipes/match`
Find recipes matching your fridge (requires auth)

**Headers:**
```
Authorization: Bearer <token>
```

**Query Params:**
- `minMatch` (optional): Minimum match percentage (default: 50)
- `limit` (optional): Max results (default: 20)

**Response:** `200 OK`
```json
{
  "data": [
    {
      "recipeId": "uuid",
      "title": "Tomato Soup",
      "matchPercentage": 95,
      "missingIngredients": [],
      "availableIngredients": 5,
      "totalIngredients": 5,
      "canCookNow": true
    },
    {
      "recipeId": "uuid",
      "title": "Pasta Carbonara",
      "matchPercentage": 60,
      "missingIngredients": [
        {"name": "Bacon", "quantity": 100, "unit": "g"}
      ],
      "availableIngredients": 3,
      "totalIngredients": 5,
      "canCookNow": false
    }
  ]
}
```

---

### 🎯 GET `/api/recipes/available`
Get recipes categorized by cooking feasibility (requires auth)

**Headers:**
```
Authorization: Bearer <token>
```

**Response:** `200 OK`
```json
{
  "canCookNow": [
    {
      "recipeId": "uuid",
      "title": "Tomato Soup",
      "matchPercentage": 100,
      "missingIngredients": []
    }
  ],
  "almostReady": [
    {
      "recipeId": "uuid",
      "title": "Pasta Carbonara",
      "matchPercentage": 80,
      "missingIngredients": [
        {"name": "Bacon", "quantity": 100, "unit": "g"}
      ]
    }
  ],
  "needMoreIngredients": [
    {
      "recipeId": "uuid",
      "title": "Beef Stew",
      "matchPercentage": 40,
      "missingIngredients": [
        {"name": "Beef", "quantity": 500, "unit": "g"},
        {"name": "Carrots", "quantity": 200, "unit": "g"}
      ]
    }
  ]
}
```

---

### 🌟 POST `/api/recipes/recommendations`
Get best recipe recommendation for UI (requires auth)

**Headers:**
```
Authorization: Bearer <token>
Accept-Language: pl|en|ru
```

**Request:**
```json
{
  "preferences": {
    "maxTime": 30,
    "difficulty": "easy",
    "category": "lunch"
  }
}
```

**Response:** `200 OK`
```json
{
  "recommendation": {
    "recipeId": "uuid",
    "title": "Tomato Soup",
    "matchPercentage": 95,
    "difficulty": "easy",
    "timeMinutes": 30,
    "servings": 4,
    "reason": "You have all ingredients and it matches your preferences",
    "canCookNow": true
  }
}
```

---

### 💾 POST `/api/user/recipes/save`
Save recipe to user's collection (requires auth)

**Request:**
```json
{
  "recipeId": "uuid"
}
```

**Response:** `200 OK`
```json
{
  "success": true,
  "message": "Recipe saved successfully"
}
```

---

### 📚 GET `/api/user/recipes/saved`
Get user's saved recipes (requires auth)

**Response:** `200 OK`
```json
{
  "data": [
    {
      "id": "uuid",
      "recipeId": "uuid",
      "title": "Tomato Soup",
      "savedAt": "2026-01-01T10:00:00Z"
    }
  ]
}
```

---

### 🍳 POST `/api/recipes/{id}/cook`
Cook recipe (deducts ingredients from fridge) (requires auth)

**Request:**
```json
{
  "servings": 2
}
```

**Response:** `200 OK`
```json
{
  "success": true,
  "data": {
    "recipeId": "uuid",
    "title": "Tomato Soup",
    "servings": 2,
    "deductedIngredients": [
      {
        "ingredientId": "uuid",
        "name": "Tomato",
        "quantity": 250,
        "unit": "g"
      }
    ],
    "remainingFridge": [
      {
        "ingredientId": "uuid",
        "name": "Tomato",
        "remainingQuantity": 250,
        "unit": "g"
      }
    ]
  }
}
```

---

### 🤖 POST `/api/recipes/{id}/adapt`
AI adapts recipe to available ingredients (requires auth)

**Request:**
```json
{
  "servings": 2,
  "language": "en"
}
```

**Response:** `200 OK`
```json
{
  "success": true,
  "data": {
    "recipeId": "uuid",
    "originalTitle": "Tomato Soup",
    "adaptedTitle": "Tomato & Carrot Soup",
    "adaptedDescription": "Adapted version using available ingredients...",
    "adaptedSteps": [
      {
        "order": 1,
        "text": "Chop tomatoes and carrots",
        "time": 5
      }
    ],
    "adaptedIngredients": [
      {
        "ingredientId": "uuid",
        "name": "Tomato",
        "quantity": 250,
        "unit": "g",
        "substitution": null
      },
      {
        "ingredientId": "uuid",
        "name": "Carrot",
        "quantity": 100,
        "unit": "g",
        "substitution": "replaces onion"
      }
    ],
    "adaptationNotes": "Replaced onion with carrot for sweeter flavor"
  }
}
```

---

### 📝 POST `/api/recipes` (OLD - User-Created Recipes)
Create user recipe (requires auth) - **NOT USED BY FRONTEND**

**Request:**
```json
{
  "title": "My Recipe",
  "description": "...",
  "steps": ["Step 1", "Step 2"],
  "ingredients": [
    {"name": "Tomato", "quantity": 500, "unit": "g"}
  ]
}
```

**Response:** `201 Created`

---

### ✏️ PUT `/api/recipes/{id}` (OLD - User-Created Recipes)
Update user recipe (requires auth) - **NOT USED BY FRONTEND**

---

### 🗑️ DELETE `/api/recipes/{id}` (OLD - User-Created Recipes)
Delete user recipe (requires auth) - **NOT USED BY FRONTEND**

---

## 6. Ingredients

### GET `/api/ingredients`
Get public ingredients catalog

**Response:** `200 OK`
```json
{
  "data": [
    {
      "id": "uuid",
      "name": "Tomato",
      "category": "vegetable",
      "unit": "g"
    }
  ]
}
```

---

## 7. AI Services

### POST `/api/ai/chat`
Chat with AI chef mentor

**Request:**
```json
{
  "message": "How do I cook salmon?",
  "language": "en"
}
```

**Response:** `200 OK`
```json
{
  "response": "To cook salmon, first...",
  "tokensUsed": 50
}
```

---

## 8. Token Economy

### GET `/api/wallet/balance`
Get user token balance

**Response:** `200 OK`
```json
{
  "balance": 100,
  "totalEarned": 150,
  "totalSpent": 50
}
```

---

### GET `/api/wallet/transactions`
Get user transaction history

**Response:** `200 OK`
```json
{
  "data": [
    {
      "id": "uuid",
      "type": "earn",
      "amount": 10,
      "reason": "Daily login",
      "createdAt": "2026-01-01T00:00:00Z"
    }
  ]
}
```

---

## 9. Academy

### GET `/api/academy/courses`
Get all courses

**Response:** `200 OK`
```json
{
  "data": [
    {
      "id": "uuid",
      "title": "Basics of Cooking",
      "description": "...",
      "lessonsCount": 10
    }
  ]
}
```

---

### POST `/api/academy/enroll`
Enroll in course

**Request:**
```json
{
  "courseId": "uuid"
}
```

**Response:** `200 OK`
```json
{
  "success": true,
  "message": "Enrolled successfully"
}
```

---

## 10. Marketplace

### GET `/api/marketplace/recipes`
Get recipes from marketplace (public)

**Query Params:**
- `page` (optional): Page number
- `limit` (optional): Items per page
- `sort` (optional): price|rating|newest

**Response:** `200 OK`
```json
{
  "data": [
    {
      "id": "uuid",
      "title": "Premium Pasta Recipe",
      "description": "...",
      "price": 50,
      "rating": 4.8,
      "sellerId": "uuid",
      "sellerName": "Chef John",
      "purchases": 120
    }
  ]
}
```

---

### GET `/api/marketplace/leaderboard`
Get marketplace sellers leaderboard (public)

**Response:** `200 OK`
```json
{
  "data": [
    {
      "userId": "uuid",
      "userName": "Chef John",
      "totalSales": 5000,
      "totalRecipes": 20,
      "averageRating": 4.8,
      "rank": 1
    }
  ]
}
```

---

### GET `/api/marketplace/stats/{userId}`
Get seller statistics (public)

**Response:** `200 OK`
```json
{
  "userId": "uuid",
  "userName": "Chef John",
  "totalSales": 5000,
  "totalRecipes": 20,
  "totalPurchases": 100,
  "averageRating": 4.8,
  "bestSellingRecipe": {
    "id": "uuid",
    "title": "Premium Pasta",
    "sales": 50
  }
}
```

---

### POST `/api/marketplace/purchase` 🔒
Purchase recipe from marketplace (requires auth)

**Request:**
```json
{
  "recipeId": "uuid"
}
```

**Response:** `200 OK`
```json
{
  "success": true,
  "data": {
    "recipeId": "uuid",
    "title": "Premium Pasta Recipe",
    "price": 50,
    "transactionId": "uuid",
    "newTokenBalance": 450
  }
}
```

---

### GET `/api/marketplace/purchases` 🔒
Get user's purchased recipes (requires auth)

**Response:** `200 OK`
```json
{
  "data": [
    {
      "recipeId": "uuid",
      "title": "Premium Pasta Recipe",
      "price": 50,
      "purchasedAt": "2026-01-01T10:00:00Z",
      "sellerId": "uuid",
      "sellerName": "Chef John"
    }
  ]
}
```

---

### POST `/api/upload/image` 🔒
Upload image (requires auth)

**Request:** `multipart/form-data`
```
file: <binary>
```

**Response:** `200 OK`
```json
{
  "success": true,
  "url": "https://cdn.example.com/images/uuid.jpg"
}
```

---

## 11. Health & Stats

### GET `/health`
Health check

**Response:** `200 OK`
```json
{
  "status": "healthy",
  "database": "connected",
  "timestamp": "2026-01-08T10:00:00Z"
}
```

---

### GET `/api/stats`
Get public statistics

**Response:** `200 OK`
```json
{
  "totalUsers": 1000,
  "totalRecipes": 500,
  "totalIngredients": 300
}
```

---

## 🌍 Localization

Many endpoints support `Accept-Language` header:

```
Accept-Language: pl|en|ru
```

**Supported Languages:**
- `pl` - Polish (Bazylia, Pomidor, Łosoś)
- `en` - English (Basil, Tomato, Salmon)
- `ru` - Russian (Базилик, Помидор, Лосось)

**Example:**
```bash
curl "http://localhost:8080/api/admin/ingredients/suggest?q=tom" \
  -H "Authorization: Bearer <token>" \
  -H "Accept-Language: pl"
```

---

## 🔒 Error Responses

All error responses follow the unified format with specific error codes.

### Error Response Structure
```json
{
  "data": null,
  "error": {
    "code": "ERROR_CODE",
    "message": "Human-readable message",
    "details": "Additional context (optional)"
  },
  "meta": {
    "request_id": "9f4c7a8b-1234-5678-90ab-cdef12345678",
    "timestamp": "2026-01-11T10:30:00Z"
  }
}
```

---

### Authentication Errors (AUTH_*)

#### 400 Bad Request
```json
{
  "data": null,
  "error": {
    "code": "AUTH_INVALID_CREDENTIALS",
    "message": "Invalid credentials",
    "details": "Please check your email and password"
  },
  "meta": { ... }
}
```

#### 401 Unauthorized
```json
{
  "data": null,
  "error": {
    "code": "AUTH_INVALID_TOKEN",
    "message": "Invalid or expired token",
    "details": "Please log in again"
  },
  "meta": { ... }
}
```

**Error Codes:**
- `AUTH_INVALID_TOKEN` - Invalid or expired JWT token
- `AUTH_MISSING_TOKEN` - Authorization header missing
- `AUTH_INVALID_CREDENTIALS` - Wrong email/password
- `AUTH_USER_EXISTS` - User already registered
- `AUTH_INSUFFICIENT_PERMISSIONS` - User lacks required role

---

### Ingredient Errors (INGREDIENT_*)

#### 404 Not Found
```json
{
  "data": null,
  "error": {
    "code": "INGREDIENT_NOT_FOUND",
    "message": "Ingredient not found",
    "details": "Ingredient with ID 'abc123' does not exist"
  },
  "meta": { ... }
}
```

**Error Codes:**
- `INGREDIENT_NOT_FOUND` - Ingredient not found
- `INGREDIENT_INVALID_INPUT` - Invalid ingredient data
- `INGREDIENT_ALREADY_EXISTS` - Ingredient already exists
- `INGREDIENT_INVALID_UNIT` - Invalid unit (must be g, ml, piece, etc.)

---

### Recipe Errors (RECIPE_*)

**Error Codes:**
- `RECIPE_NOT_FOUND` - Recipe not found
- `RECIPE_INVALID_INPUT` - Invalid recipe data
- `RECIPE_AI_GENERATION_FAILED` - AI recipe generation failed
- `RECIPE_VALIDATION_FAILED` - Recipe validation failed
- `RECIPE_INSUFFICIENT_INGREDIENTS` - Not enough ingredients in fridge

---

### Fridge Errors (FRIDGE_*)

**Error Codes:**
- `FRIDGE_ITEM_NOT_FOUND` - Fridge item not found
- `FRIDGE_INSUFFICIENT_QUANTITY` - Not enough quantity in fridge
- `FRIDGE_INVALID_INPUT` - Invalid fridge item data
- `FRIDGE_ITEM_EXPIRED` - Fridge item has expired

---

### Token Economy Errors (TOKEN_*)

**Error Codes:**
- `TOKEN_INSUFFICIENT_BALANCE` - Not enough tokens
- `TOKEN_INVALID_AMOUNT` - Invalid token amount
- `TOKEN_TRANSACTION_FAILED` - Token transaction failed

---

### General Errors (GENERAL_*)

#### 400 Bad Request
```json
{
  "data": null,
  "error": {
    "code": "GENERAL_INVALID_JSON",
    "message": "Invalid request body",
    "details": "The request body must be valid JSON"
  },
  "meta": { ... }
}
```

#### 500 Internal Server Error
```json
{
  "data": null,
  "error": {
    "code": "GENERAL_INTERNAL_ERROR",
    "message": "Internal server error",
    "details": "An unexpected error occurred"
  },
  "meta": { ... }
}
```

**Error Codes:**
- `GENERAL_INVALID_INPUT` - Invalid input data
- `GENERAL_INVALID_JSON` - Invalid JSON in request body
- `GENERAL_INTERNAL_ERROR` - Internal server error
- `GENERAL_NOT_FOUND` - Resource not found
- `GENERAL_DATABASE_ERROR` - Database operation failed

---

### HTTP Status Codes

- **200 OK** - Request successful
- **201 Created** - Resource created successfully
- **400 Bad Request** - Invalid input or validation error
- **401 Unauthorized** - Authentication required or failed
- **403 Forbidden** - Insufficient permissions
- **404 Not Found** - Resource not found
- **500 Internal Server Error** - Server error

---

## 📝 Notes

1. **Pagination**: Most list endpoints support `page` and `limit` query params
2. **Localization**: Use `Accept-Language` header for multilingual content
3. **AI Endpoints**: May take longer to respond (5-30 seconds)
4. **Token Economy**: All token operations are logged for audit
5. **Validation**: Input data is validated before processing

---

## 🚀 Quick Start Examples

### Login and Get Ingredients
```bash
# Login
TOKEN=$(curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"admin_password_123"}' \
  | jq -r '.data.token')

# Get ingredients
curl http://localhost:8080/api/admin/ingredients \
  -H "Authorization: Bearer $TOKEN" \
  -H "Accept-Language: pl"
```

### Search Ingredients
```bash
curl "http://localhost:8080/api/admin/ingredients/suggest?q=pom&limit=5" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Accept-Language: pl"
```

### Create AI Recipe
```bash
curl -X POST http://localhost:8080/api/admin/recipes/preview-ai \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -H "Accept-Language: en" \
  -d '{
    "title": "Grilled Salmon",
    "language": "en",
    "ingredients": [
      {"ingredientId": "uuid", "amount": 200, "unit": "g"}
    ],
    "rawCookingText": "Grill salmon for 5 minutes..."
  }'
```

---

**Generated:** 2026-01-08  
**Version:** 1.0.0  
**Status:** ✅ Complete and Tested
