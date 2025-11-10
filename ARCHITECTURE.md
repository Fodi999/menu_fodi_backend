# Fridge-Chat Integration Architecture

## System Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                         FRONTEND / CLIENT                         │
├─────────────────────────────────────────────────────────────────┤
│                                                                   │
│  1. User starts: "I want to make pasta carbonara"               │
│     ↓                                                             │
│  2. AI provides guidance and suggestions                         │
│     ↓                                                             │
│  3. Recipe building through conversation                        │
│     ↓                                                             │
│  4. When complete: Shows action buttons                         │
│     ├─ Save Recipe                                              │
│     ├─ Save Ingredients to Fridge ← USER SELECTS THIS         │
│     └─ Generate Meal Plan                                      │
│                                                                   │
└─────────────────────────────────────────────────────────────────┘
                              ↑
                              │
                   HTTP/HTTPS Requests
                              │
                              ↓
┌─────────────────────────────────────────────────────────────────┐
│                        BACKEND API LAYER                          │
├─────────────────────────────────────────────────────────────────┤
│                                                                   │
│  ┌─────────────────────────────────────────────────────┐       │
│  │ HTTP Request: POST /api/ai/save-ingredients        │       │
│  │ Headers: Authorization: Bearer {JWT_TOKEN}         │       │
│  │ Body: {"ingredients": [...]}                        │       │
│  └─────────────────────────────────────────────────────┘       │
│                        ↓                                         │
│  ┌─────────────────────────────────────────────────────┐       │
│  │        JWT MIDDLEWARE (jwtMiddleware)               │       │
│  │  ✓ Validates token                                 │       │
│  │  ✓ Extracts user_id from token                     │       │
│  │  ✓ Adds to request context                         │       │
│  │                                                      │       │
│  │  If invalid: Return 401 Unauthorized               │       │
│  └─────────────────────────────────────────────────────┘       │
│                        ↓                                         │
│  ┌─────────────────────────────────────────────────────┐       │
│  │   AIHandlers.SaveRecipeIngredientsToFridge()       │       │
│  │                                                      │       │
│  │   1. Extract user_id from context                  │       │
│  │   2. Decode SaveIngredientsRequest from JSON       │       │
│  │   3. Validate ingredients not empty (400 if empty) │       │
│  │   4. For each ingredient:                          │       │
│  │      - Create UserFridge model                     │       │
│  │      - Set product, quantity, unit                 │       │
│  │      - Set available = true                        │       │
│  │      - Set user_id from JWT                        │       │
│  │   5. Call database repository                      │       │
│  │   6. Return JSON: {success, message, count}        │       │
│  └─────────────────────────────────────────────────────┘       │
│                        ↓                                         │
│  ┌─────────────────────────────────────────────────────┐       │
│  │        DATABASE LAYER (GORM Repository)            │       │
│  │                                                      │       │
│  │   For each ingredient:                             │       │
│  │   ├─ Create(UserFridge{                            │       │
│  │   │   ID: uuid.New(),                              │       │
│  │   │   UserID: user_id,                             │       │
│  │   │   Product: ingredient.Name,                    │       │
│  │   │   Quantity: ingredient.Amount,                 │       │
│  │   │   Unit: ingredient.Unit,                       │       │
│  │   │   Available: true,                             │       │
│  │   │   AddedAt: time.Now(),                         │       │
│  │   │   UpdatedAt: time.Now()                        │       │
│  │   })                                                │       │
│  │   └─ If error: Return 500 error                    │       │
│  │                                                      │       │
│  │   Count successfully inserted records              │       │
│  └─────────────────────────────────────────────────────┘       │
│                        ↓                                         │
│  ┌─────────────────────────────────────────────────────┐       │
│  │         HTTP Response: 200 OK                       │       │
│  │                                                      │       │
│  │  {                                                  │       │
│  │    "success": true,                                │       │
│  │    "message": "ingredients saved to fridge",       │       │
│  │    "count": 4                                      │       │
│  │  }                                                  │       │
│  └─────────────────────────────────────────────────────┘       │
│                                                                   │
└─────────────────────────────────────────────────────────────────┘
                              ↑
                              │
                   HTTP Response (200)
                              │
                              ↓
┌─────────────────────────────────────────────────────────────────┐
│                      DATABASE LAYER                               │
├─────────────────────────────────────────────────────────────────┤
│                                                                   │
│  PostgreSQL - user_fridge Table                                 │
│  ┌──────────────────────────────────────────────────────┐      │
│  │ id (UUID)      │ 550e8400-e29b-41d4-a716-446655440000 │      │
│  │ user_id (UUID) │ a5a8c0e0-... (from JWT)           │      │
│  │ product        │ Pasta                                │      │
│  │ quantity       │ 400.00                               │      │
│  │ unit           │ g                                    │      │
│  │ available      │ true                                 │      │
│  │ added_at       │ 2024-11-10 12:34:56                 │      │
│  │ updated_at     │ 2024-11-10 12:34:56                 │      │
│  ├──────────────────────────────────────────────────────┤      │
│  │ id             │ 550e8400-e29b-41d4-a716-446655440001 │      │
│  │ user_id        │ a5a8c0e0-... (from JWT)            │      │
│  │ product        │ Eggs                                 │      │
│  │ quantity       │ 3.00                                 │      │
│  │ unit           │ pcs                                  │      │
│  │ available      │ true                                 │      │
│  │ added_at       │ 2024-11-10 12:34:56                 │      │
│  │ updated_at     │ 2024-11-10 12:34:56                 │      │
│  ├──────────────────────────────────────────────────────┤      │
│  │ ... 2 more records (Bacon, Parmesan)                 │      │
│  └──────────────────────────────────────────────────────┘      │
│                                                                   │
│  Now user can:                                                  │
│  ├─ GET /api/fridge/ → See all ingredients                     │
│  ├─ POST /api/ai/fridge-recommendations → Get recipes          │
│  ├─ POST /api/ai/meal-plan → Plan meals                        │
│  └─ PUT/DELETE /api/fridge/{id} → Manage items                 │
│                                                                   │
└─────────────────────────────────────────────────────────────────┘
```

## Data Flow Sequence Diagram

```
User                 Frontend              Backend API              Database
 │                      │                      │                       │
 │ "Make pasta"         │                      │                       │
 ├─────────────────────>│                      │                       │
 │                      │  POST /chef-mentor   │                       │
 │                      ├─────────────────────>│                       │
 │                      │                      │  Call Groq AI         │
 │                      │  {message, recipe}   │                       │
 │                      │<─────────────────────┤                       │
 │<─────────────────────┤                      │                       │
 │ Shows: AI guidance   │                      │                       │
 │ + suggested actions  │                      │                       │
 │                      │                      │                       │
 │ Clicks "Save to      │                      │                       │
 │ Fridge" button       │                      │                       │
 ├─────────────────────>│                      │                       │
 │                      │ POST /save-ingredients (JWT)                 │
 │                      │ {ingredients: [...]} │                       │
 │                      ├─────────────────────>│                       │
 │                      │                      │ 1. Validate JWT       │
 │                      │                      │ 2. Extract user_id    │
 │                      │                      │                       │
 │                      │                      │ For each ingredient:  │
 │                      │                      │  ├─ Create record    │
 │                      │                      ├──────────────────────>
 │                      │                      │  │ INSERT user_fridge │
 │                      │                      │<──────────────────────
 │                      │                      │                       │
 │                      │ 200 OK               │                       │
 │                      │ {success, count}     │                       │
 │                      │<─────────────────────┤                       │
 │                      │                      │                       │
 │ "4 ingredients       │                      │                       │
 │ saved!" ✓            │                      │                       │
 │<─────────────────────┤                      │                       │
 │                      │                      │                       │
 │ Clicks "View Fridge" │                      │                       │
 ├─────────────────────>│                      │                       │
 │                      │ GET /api/fridge/ (JWT)                       │
 │                      ├─────────────────────>│                       │
 │                      │                      │ SELECT * FROM user_fridge
 │                      │                      │ WHERE user_id = ?    │
 │                      │                      ├──────────────────────>
 │                      │                      │<──────────────────────
 │                      │ [ingredients]        │                       │
 │                      │<─────────────────────┤                       │
 │<─────────────────────┤                      │                       │
 │ Shows fridge items   │                      │                       │
 │ - Pasta 400g ✓       │                      │                       │
 │ - Eggs 3pcs ✓        │                      │                       │
 │ - Bacon 200g ✓       │                      │                       │
 │ - Parmesan 100g ✓    │                      │                       │
 │                      │                      │                       │
```

## Component Interaction Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                   AI MODULE COMPONENTS                            │
├─────────────────────────────────────────────────────────────────┤
│                                                                   │
│  ┌─────────────────────┐                                        │
│  │  HTTP Handlers      │                                        │
│  │  (transport layer)  │                                        │
│  │                     │                                        │
│  │ - ChefMentor()      │                                        │
│  │ - GenerateRecipe()  │                                        │
│  │ - SaveRecipe...() ◄─────── NEW                              │
│  │ - MealPlan()        │                                        │
│  │ - Recomm...()       │                                        │
│  └──────────┬──────────┘                                        │
│             │                                                    │
│             │ calls                                              │
│             ↓                                                    │
│  ┌─────────────────────┐                                        │
│  │  DTOs               │                                        │
│  │  (data structures)  │                                        │
│  │                     │                                        │
│  │ - ChefMentorReq     │                                        │
│  │ - ChefMentorRes     │                                        │
│  │ - SaveIngredients...◄─────── NEW                           │
│  │ - RecipeGenerationReq│                                       │
│  │ - etc.              │                                        │
│  └──────────┬──────────┘                                        │
│             │                                                    │
│             │ uses                                               │
│             ↓                                                    │
│  ┌─────────────────────┐                                        │
│  │  Service Layer      │                                        │
│  │  (business logic)   │                                        │
│  │                     │                                        │
│  │ - ChefMentor() ◄────────── ENHANCED with actions            │
│  │ - GenerateRecipe()  │                                        │
│  │ - MealPlan()        │                                        │
│  │ - Recomm...()       │                                        │
│  └──────────┬──────────┘                                        │
│             │                                                    │
│             │ calls                                              │
│             ↓                                                    │
│  ┌─────────────────────┐                                        │
│  │  Groq AI Client     │                                        │
│  │  (external API)     │                                        │
│  │                     │                                        │
│  │ - Chat()            │                                        │
│  │ - Stream()          │                                        │
│  └─────────────────────┘                                        │
│                                                                   │
└─────────────────────────────────────────────────────────────────┘
         ↑                          ↑                      ↑
         │                          │                      │
         │                          │                      │
    ┌────┴──────────────┬───────────┴──────────┬──────────┴─────┐
    │                   │                      │                │
    ↓                   ↓                      ↓                ↓
┌────────────┐  ┌──────────────┐    ┌─────────────────┐  ┌──────────┐
│ Middleware │  │   Router     │    │ Database Repo.  │  │  Models  │
│            │  │              │    │                 │  │          │
│ - JWT Auth │  │ /chef-mentor │    │ - UserFridge    │  │ Recipe   │
│ - CORS     │  │ /save-ingred.│    │ - Create()      │  │ User     │
│ - Logging  │  │ /meal-plan   │    │ - Read()        │  │ Order    │
│            │  │ /fridge-reco │    │ - Update()      │  │ etc.     │
│            │  │              │    │ - Delete()      │  │          │
│            │  │ ◄─── NEW ────│    │                 │  │          │
│            │  │              │    │                 │  │          │
└────────────┘  └──────────────┘    └─────────────────┘  └──────────┘
```

## Authentication & Authorization Flow

```
┌──────────────────────────────────────────────────────────┐
│         Protected Endpoint: /api/ai/save-ingredients      │
└──────────────────────────────────────────────────────────┘

Request Header Check:
┌────────────────────────────────────────────────────────┐
│ Authorization: Bearer eyJhbGciOiJIUzI1NiIs...         │
└────────────────────────────────────────────────────────┘
                          ↓
                   [JWT Middleware]
                          ↓
        ┌─────────────────────────────────────┐
        │ 1. Extract token from header        │
        │ 2. Verify signature                 │
        │ 3. Decode claims                    │
        │ 4. Check expiration                 │
        └─────────────────────────────────────┘
                          ↓
                 ┌────────┴────────┐
                 │                 │
              Valid           Invalid
                 │                 │
                 ↓                 ↓
        ┌──────────────┐  ┌─────────────────┐
        │ Extract data:│  │ Return 401      │
        │ - user_id    │  │ Unauthorized    │
        │ - email      │  │                 │
        │ - roles      │  │ No further      │
        │              │  │ processing      │
        │ Add to       │  └─────────────────┘
        │ context      │
        └──────────────┘
                 │
                 ↓
        ┌──────────────────────┐
        │ Handler executes     │
        │ Uses user_id from    │
        │ context to:          │
        │ - Create records     │
        │ - Verify ownership   │
        │ - Return responses   │
        └──────────────────────┘
```

## Error Handling Flow

```
POST /api/ai/save-ingredients

┌─────────────────────────────────────────┐
│ 1. Check Authorization Header           │
└─────────────────────────────────────────┘
         │ No header              │ Valid header
         ↓                        ↓
      401 Auth           ┌──────────────────┐
      Required           │ Validate JWT     │
                         └──────────────────┘
                         │ Invalid    │ Valid
                         ↓            ↓
                      401 Unauth  ┌─────────────────┐
                                  │ Parse JSON Body │
                                  └─────────────────┘
                                  │ Invalid    │ Valid
                                  ↓            ↓
                               400 Bad Req ┌────────────────┐
                                           │ Validate       │
                                           │ Ingredients    │
                                           └────────────────┘
                                           │ Empty      │ Valid
                                           ↓            ↓
                                        400 Bad Req ┌─────────────┐
                                                    │ Save to DB  │
                                                    └─────────────┘
                                                    │ Error  │ Success
                                                    ↓       ↓
                                                 500 Error 200 OK
                                                           {success}
```

## Technology Stack Integration

```
┌────────────────────────────────────────────────────────┐
│                   GO BACKEND STACK                       │
├────────────────────────────────────────────────────────┤
│                                                          │
│  ┌──────────────────────────────────────────┐         │
│  │ Framework: Chi v5 (HTTP Router)          │         │
│  │ ORM: GORM (Object-Relational Mapping)    │         │
│  │ DB: PostgreSQL                           │         │
│  │ Auth: JWT (JSON Web Tokens)              │         │
│  │ AI: Groq API (Chat Completions)          │         │
│  │ UUID: Google UUID Library                │         │
│  └──────────────────────────────────────────┘         │
│                                                          │
│  POST /api/ai/save-ingredients                        │
│     ├─ Chi Router (matches request)                   │
│     ├─ JWT Middleware (validates token)               │
│     ├─ Handler (SaveRecipeIngredientsToFridge)       │
│     ├─ GORM (executes SQL)                            │
│     └─ PostgreSQL (stores data)                       │
│                                                          │
└────────────────────────────────────────────────────────┘
```

---

This architecture ensures:
- ✅ Security through JWT authentication
- ✅ Scalability through separation of concerns
- ✅ Reliability through proper error handling
- ✅ Maintainability through clear component structure
- ✅ Performance through optimized database operations
