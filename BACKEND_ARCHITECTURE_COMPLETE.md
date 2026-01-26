# Backend Code Structure - Complete Overview

## Date: 2026-01-26

---

## Project Architecture

```
backend/
├── cmd/                          # Command-line applications
│   ├── server/                   # Main API server
│   ├── check_arrived/            # Helper utilities
│   ├── cleanup_expired/
│   ├── cleanup_ingredients/
│   ├── drop_fk/
│   ├── init_treasury/
│   ├── migrate/
│   ├── test_saved_recipes/
│   └── test_ws/
│
├── internal/                     # Core business logic (private)
│   ├── app/                      # Application setup & routes
│   │   ├── app.go               # Main app initialization
│   │   ├── routes_modular.go    # Modular route registration
│   │   └── server_setup.go      # Server configuration
│   │
│   ├── middleware/               # HTTP middleware
│   │   ├── auth.go              # JWT, role-based middleware
│   │   ├── cors.go              # CORS handling
│   │   └── logging.go           # Request logging
│   │
│   ├── models/                   # Data models (ORM)
│   │   ├── user.go              # User model with roles
│   │   ├── recipe_catalog.go    # Recipe with multilingual support
│   │   ├── ingredient.go        # Ingredient model
│   │   ├── dish.go              # Dish model (menu cards) ← MULTILINGUAL
│   │   ├── user_menu_items.go   # Kitchen pipeline tracking
│   │   ├── user_fridge.go       # Fridge storage
│   │   ├── token_bank.go        # Token/balance system
│   │   └── ... (other models)
│   │
│   ├── modules/                  # Feature modules (DDD)
│   │   │
│   │   ├── admin/               # Admin panel operations
│   │   │   ├── service/
│   │   │   │   ├── service.go           # AdminService interface
│   │   │   │   ├── dish_ai.go           # Dish AI generation ← MULTILINGUAL
│   │   │   │   ├── dish_crud.go         # Dish CRUD operations
│   │   │   │   ├── recipe_ai.go         # Recipe AI generation
│   │   │   │   ├── canonical_ingredient_service.go
│   │   │   │   └── policy.go            # Admin policy rules
│   │   │   │
│   │   │   ├── transport/http/
│   │   │   │   ├── handlers.go          # User, token, treasury handlers
│   │   │   │   ├── dish_handlers.go     # Dish API endpoints ← NEW
│   │   │   │   └── recipe_handlers.go   # Recipe management
│   │   │   │
│   │   │   ├── repository/              # Data access layer
│   │   │   ├── dto/                     # Data transfer objects
│   │   │   └── module.go                # Module registration
│   │   │
│   │   ├── auth/                 # Authentication
│   │   │   ├── service/
│   │   │   │   ├── auth_service.go
│   │   │   │   └── jwt_service.go
│   │   │   ├── transport/http/
│   │   │   └── module.go
│   │   │
│   │   ├── menu/                 # Kitchen pipeline (active menu)
│   │   │   ├── service/
│   │   │   │   ├── menu_service.go
│   │   │   │   └── (includes GetHistory for completed items) ← NEW
│   │   │   ├── transport/http/
│   │   │   │   └── menu_handler.go (GET /api/menu/history) ← NEW
│   │   │   ├── repository/
│   │   │   └── module.go
│   │   │
│   │   ├── fridge/              # User storage management
│   │   ├── history/             # Event tracking & analytics
│   │   ├── notifications/       # Real-time updates
│   │   ├── websocket/           # WebSocket connections
│   │   └── ... (other modules)
│   │
│   ├── platform/                # Infrastructure
│   │   ├── database/
│   │   │   ├── database.go      # Connection & migrations
│   │   │   └── seeds.go         # Test data
│   │   ├── logger/              # Logging setup
│   │   ├── event_bus/           # Event publishing
│   │   └── ai_core/             # AI client (Groq, OpenAI)
│   │
│   └── cron/                     # Scheduled tasks
│       ├── fridge_expiry_checker.go
│       └── ... (other jobs)
│
├── pkg/                         # Shared utilities
│   ├── utils/                   # Helper functions
│   │   ├── response.go          # API response formatting
│   │   ├── validator.go         # Input validation
│   │   └── ... (utilities)
│   └── ... (other packages)
│
├── migrations/                  # Database migrations (SQL)
│   ├── 20260122_recreate_user_menu_items_fixed.sql  (Kitchen pipeline)
│   ├── 20260123_add_menu_history.sql                (History separation) ← NEW
│   ├── 20260126_add_multilingual_to_dishes.sql      (Multilingual dishes) ← NEW
│   └── ... (other migrations)
│
├── docker/                      # Container setup
│   └── Dockerfile
│
├── sql/                         # SQL utilities & backups
│
├── scripts/                     # Helper scripts
│   ├── test_*.sh               # Test scripts
│   └── ... (other scripts)
│
├── go.mod                       # Go modules
├── go.sum                       # Go dependencies lock
├── Makefile                     # Build targets
├── README.md                    # Documentation
│
└── docs/                        # Architecture documentation
    ├── *_COMPLETE.md           # Feature documentation
    └── ... (guides)
```

---

## Core Modules Deep Dive

### 1. Admin Module

**Purpose:** Administrative operations (recipes, ingredients, dishes, users, tokens)

**Key Files:**
```
admin/
├── service/
│   ├── service.go              # AdminService interface (1300+ lines)
│   │   - Users management
│   │   - Orders handling
│   │   - Token bank operations
│   │   - Treasury management
│   │   - Ingredients catalog (AI-powered)
│   │   - Recipes catalog management
│   │   - GetUserByID ← NEW
│   │   - UpdateUserRole(userID, role, adminID) ← UPDATED
│   │   - Dishes CRUD ← NEW
│   │
│   ├── recipe_ai.go            # AI recipe generation
│   ├── dish_ai.go              # AI dish generation (now multilingual) ← UPDATED
│   ├── dish_crud.go            # Dish create/update/delete
│   ├── canonical_ingredient_service.go
│   └── policy.go               # Access control policies
│
├── transport/http/
│   ├── handlers.go             # Main admin handlers (1275+ lines)
│   │   - GetAllUsers, GetUserByID ← NEW
│   │   - UpdateUser, DeleteUser
│   │   - UpdateUserRole ← UPDATED (now logs admin ID)
│   │   - GetAdminStats, GetAdminProfile
│   │   - Token bank operations
│   │   - Treasury endpoints
│   │   - Ingredients management
│   │   - Recipe management
│   │   - Dishes endpoints ← NEW
│   │
│   ├── dish_handlers.go        # Dedicated dish handlers ← NEW
│   │   - GenerateDishFromRecipe
│   │   - GetDishes, GetDishByID
│   │   - UpdateDish
│   │   - ApproveDish, PublishDish, UnpublishDish
│   │   - DeleteDish
│   │
│   └── recipe_handlers.go      # Recipe management
│
└── module.go                    # Route registration
```

**Key Service Methods (AdminService Interface):**

```go
type AdminService interface {
    // Users
    GetAllUsers() ([]models.User, error)
    GetUsersWithFilters(params GetUsersParams) (*UserListResponse, error)
    GetUsersStats() (map[string]interface{}, error)
    GetUserByID(userID string) (*models.User, error)                    // ← NEW
    UpdateUser(userID string, name, email string) (*models.User, error)
    DeleteUser(userID string) error
    UpdateUserRole(userID, role, adminID string) error                 // ← UPDATED
    
    // Dishes ← NEW
    GenerateDishWithAI(req GenerateDishRequest, adminID string) (*models.Dish, error)
    GetDishes(params GetDishesParams) ([]models.Dish, int64, error)
    GetDishByID(dishID string) (*models.Dish, error)
    UpdateDish(dishID string, req UpdateDishRequest, adminID string) (*models.Dish, error)
    ApproveDish(dishID, adminID string) error
    PublishDish(dishID, adminID string) error
    UnpublishDish(dishID, adminID string) error
    DeleteDish(dishID, adminID string) error
    
    // ... (other methods for ingredients, recipes, tokens, treasury)
}
```

---

### 2. Menu Module

**Purpose:** Kitchen pipeline - managing active cooking items

**Files:**
```
menu/
├── service/
│   └── menu_service.go
│       - GetTodayMenu() → returns only active (planned + cooking) ← FIXED
│       - GetHistory() → returns completed items ← NEW
│       - AddToMenu, StartCooking, CompleteCooking
│
├── transport/http/
│   └── menu_handler.go
│       - GetTodayMenu handler
│       - GetHistory handler ← NEW (GET /api/menu/history?limit=30)
│       - AddToMenu, StartCooking, CompleteCooking
│
├── repository/
│   └── menu_repository.go
│       - GetTodayMenu query (only planned + cooking) ← FIXED
│       - GetHistory query (only completed) ← ALREADY EXISTS
│
└── module.go
    - Route registration with GetHistory endpoint
```

**Key Changes:**
- ✅ Fixed GetTodayMenu to return only active items (planned + cooking)
- ✅ Added GetHistory endpoint to get completed items
- ✅ Completed items no longer clutter active menu

**API Endpoints:**
```
GET  /api/menu/today                    # Active items only (planned + cooking)
GET  /api/menu/history?limit=30         # Completed items
POST /api/menu/today                    # Add recipe to menu
POST /api/menu/{id}/start               # Start cooking
POST /api/menu/{id}/complete            # Mark as completed
```

---

### 3. Auth Module

**Purpose:** User authentication and JWT token management

**Files:**
```
auth/
├── service/
│   ├── auth_service.go          # Registration, login, password reset
│   ├── jwt_service.go           # JWT creation/validation
│   └── claims.go                # JWT claims structure
│
├── transport/http/
│   └── auth_handlers.go         # Register, login, verify endpoints
│
└── module.go
```

---

## Data Models (ORM)

### User Model
```go
type User struct {
    ID     string  // TEXT (not UUID!)
    Email  string
    Name   string
    Role   string  // home_chef | chef_staff | admin | super_admin | customer | investor
    Status string  // active | blocked | pending
    // ... timestamps
}
```

### Recipe Model
```go
type RecipeCatalog struct {
    ID            uuid.UUID
    CanonicalName string
    
    // Multilingual (PL, EN, RU)
    NamePl, NameEn, NameRu             *string
    DescriptionPl, DescriptionEn, DescriptionRu *string
    StepsPl, StepsEn, StepsRu          datatypes.JSON
    
    // Metadata
    Category, Difficulty string
    TimeMinutes, Servings int
    // ... timestamps, relations
}
```

### Dish Model (NEW - Multilingual)
```go
type Dish struct {
    ID       uuid.UUID
    RecipeID uuid.UUID
    
    // Multilingual (PL, EN, RU) ← NEW
    Title,       *TitlePl,       *TitleEn,       *TitleRu       string
    Description, *DescriptionPl, *DescriptionEn, *DescriptionRu string
    
    // Finance
    Cost   float64  // Ingredient costs
    Price  float64  // Selling price
    Margin float64  // Profit margin %
    
    // Status lifecycle
    Status DishStatus // draft → approved → published
    
    // Metadata
    CreatedBy  string     // Admin UUID
    ApprovedBy *string    // Admin UUID
    CreatedAt  time.Time
    UpdatedAt  time.Time
    
    // Relations
    Recipe   RecipeCatalog
    Creator  User
    Approver *User
}
```

### UserMenuItem Model (Kitchen Pipeline)
```go
type UserMenuItem struct {
    ID         uuid.UUID
    UserID     string    // FK to User.id
    RecipeID   uuid.UUID // FK to Recipe.id
    
    Servings   int
    Status     string    // planned | cooking | completed | cancelled
    PlannedFor string    // YYYY-MM-DD
    
    CreatedAt       time.Time
    StartedCookingAt *time.Time
    CompletedAt     *time.Time
}
```

---

## Database Schema Highlights

### User Tables

```sql
-- Main users table
"User" (TEXT id, email, name, role, status, ...)  [Prisma naming - capital U]

-- Legacy (to be removed)
users (UUID id, email, name, role, ...)  [DEPRECATED]
```

### Menu Tables

```sql
user_menu_items(
    id UUID PRIMARY KEY,
    user_id TEXT FK,
    recipe_id UUID FK,
    status VARCHAR (planned|cooking|completed|cancelled),
    planned_for DATE,
    started_cooking_at TIMESTAMP,
    completed_at TIMESTAMP,
    
    UNIQUE(user_id, recipe_id, planned_for)  -- Fixed constraint
)
```

### Recipe/Ingredient Tables

```sql
Recipe (
    id UUID PRIMARY KEY,
    canonical_name VARCHAR UNIQUE,
    
    -- Multilingual
    name_pl, name_en, name_ru VARCHAR,
    description_pl, description_en, description_ru TEXT,
    steps_pl, steps_en, steps_ru JSONB,
    
    category, difficulty, time_minutes, servings INT,
    ...
)

Ingredient (
    id UUID PRIMARY KEY,
    name_pl, name_en, name_ru VARCHAR,
    category, unit VARCHAR,
    ...
)
```

### Dishes Table (NEW)

```sql
dishes (
    id UUID PRIMARY KEY,
    recipe_id UUID FK,
    
    -- Multilingual (PL, EN, RU)
    title VARCHAR NOT NULL,
    title_pl, title_en, title_ru VARCHAR,
    description TEXT,
    description_pl, description_en, description_ru TEXT,
    
    -- Finance
    cost DECIMAL(10,2),
    price DECIMAL(10,2),
    margin DECIMAL(5,2),
    
    -- Lifecycle
    status VARCHAR (draft|approved|published),
    is_available BOOLEAN,
    
    created_by TEXT FK,
    approved_by TEXT FK,
    approved_at TIMESTAMP,
    
    created_at, updated_at TIMESTAMP
)
```

---

## API Routes

### Menu API (Active)
```
GET  /api/menu/today              # Active items (planned + cooking)
GET  /api/menu/history?limit=30   # Completed items ← NEW
POST /api/menu/today              # Add to menu
POST /api/menu/{id}/start         # Start cooking
POST /api/menu/{id}/complete      # Mark completed
POST /api/menu/{id}/cancel        # Cancel item
DELETE /api/menu/{id}             # Delete item
```

### Admin API (Dishes) ← NEW
```
POST   /api/admin/dishes/generate-from-recipe   # AI generation
GET    /api/admin/dishes?status=draft&limit=20
GET    /api/admin/dishes/{id}
PATCH  /api/admin/dishes/{id}                   # Edit draft/approved
POST   /api/admin/dishes/{id}/approve           # draft → approved
POST   /api/admin/dishes/{id}/publish           # approved → published
POST   /api/admin/dishes/{id}/unpublish         # published → approved
DELETE /api/admin/dishes/{id}                   # Delete draft only
```

### Admin API (Users) ← UPDATED
```
GET    /api/admin/users?page=1&limit=20&role=home_chef
GET    /api/admin/users/{id}                    # ← NEW
POST   /api/admin/users/{id}
PATCH  /api/admin/users/{id}
DELETE /api/admin/users/{id}
PATCH  /api/admin/users/{id}/role               # ← UPDATED (logs admin ID)
```

---

## Middleware Stack

### CORS Middleware
```go
// Allows specific origins (not wildcard *)
allowedOrigins := map[string]bool{
    "https://dima-fomin.pl": true,
    "http://localhost:3000": true,
}

// Returns Access-Control-Allow-Origin header only for allowed origins
```

### Auth Middleware
```go
// Validates JWT token
// Extracts user ID and role from claims
// Stores in context for handlers
```

### Role-Based Middleware
```go
AdminMiddleware()       // Requires admin or super_admin role
SuperAdminMiddleware()  // Requires super_admin role only
```

---

## Migrations (SQL)

```
migrations/
├── 20250101_initial_schema.sql
├── 20260122_recreate_user_menu_items_fixed.sql
│   - Fixes unique constraint (allows status transitions)
│   - UNIQUE(user_id, recipe_id, planned_for) -- removed status
│
├── 20260123_add_menu_history.sql
│   - Creates GetHistory view/helper
│
└── 20260126_add_multilingual_to_dishes.sql ← NEW
    - Adds title_pl, title_en, title_ru
    - Adds description_pl, description_en, description_ru
    - Creates language-specific indexes
```

---

## Key Architectural Principles

### 1. DDD (Domain-Driven Design)
- Each feature is a module with service, repository, transport layers
- Routes organized by module
- Clear separation of concerns

### 2. Multilingual First
- Recipes, Dishes, Ingredients all support PL/EN/RU
- Fallback logic ensures no missing content
- Language negotiation via headers

### 3. Type Safety
- Strong Go typing prevents errors
- Database migrations typed and versioned
- Foreign key constraints enforced

### 4. Audit Trail
- Admin actions logged (who changed role, when)
- Event bus for system-wide notifications
- History tracking for kitchen pipeline

### 5. API Consistency
- Camel case in JSON (titlePl, not title_pl)
- Standard error responses
- Pagination with limit/offset
- Query parameter validation

---

## Build & Run

```bash
# Build
go build -o server cmd/server/main.go

# Run
./server

# Run tests
go test ./...

# Database migrations
go run cmd/migrate/main.go

# Docker
docker build -t backend .
docker run -e DATABASE_URL="..." backend
```

---

## Development Tools

- **Database:** PostgreSQL (Neon)
- **ORM:** GORM
- **Router:** Chi (Go HTTP router)
- **Logging:** Zap
- **AI:** Groq API (via ai_core module)
- **Testing:** Go testing + custom test scripts
- **Deployment:** Koyeb (backend), Vercel (frontend)

---

## Recent Fixes & Additions (2026-01-22 to 2026-01-26)

✅ Kitchen Pipeline constraint fix (unique key includes status)  
✅ Menu history separation (completed items in /history endpoint)  
✅ CORS fix for https://dima-fomin.pl (whitelist instead of wildcard)  
✅ Legacy `users` table cleanup (migrated to "User" table)  
✅ Admin service improvements (GetUserByID, UpdateUserRole audit logging)  
✅ Dishes multilingual support (PL, EN, RU titles & descriptions)  

---

## Status

✅ **Production Ready**

All core modules functioning, multilingual support complete, architecture scalable.

---

**Last Updated:** 2026-01-26  
**Version:** 1.0 (ChefOS Backend)  
**Deployment:** Koyeb (continuous deployment from main branch)
