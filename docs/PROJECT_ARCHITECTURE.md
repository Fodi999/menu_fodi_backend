# 🏗️ Архитектура проекта Chef Academy Backend

## 📋 Общая информация

**Стек технологий:**
- **Язык:** Go 1.24.3
- **Фреймворк:** Chi Router v5
- **База данных:** PostgreSQL (GORM ORM)
- **Хостинг:** Neon (DB) + Koyeb (Server)
- **Архитектурный паттерн:** Domain-Driven Design (DDD) + Clean Architecture

---

## 🎯 Архитектурные принципы

### 1. **Модульная структура (DDD)**
Каждый модуль — это независимый домен с собственной бизнес-логикой.

### 2. **Слоистая архитектура (Clean Architecture)**
```
Transport (HTTP) → Service (Business Logic) → Repository (Data Access)
```

### 3. **Dependency Injection**
Зависимости передаются через конструкторы, а не создаются внутри.

### 4. **Разделение ответственности**
- **Transport:** HTTP handlers, request/response
- **Service:** Бизнес-логика, валидация
- **Repository:** Работа с БД

---

## 🏛️ Структура проекта

```
backend/
│
├── cmd/                          # Entry points (команды)
│   ├── server/                   # Основной HTTP сервер
│   │   └── main.go              # 🚀 Точка входа приложения
│   ├── migrate/                  # Миграции БД
│   ├── check_arrived/            # Утилиты проверки
│   └── cleanup_ingredients/      # Чистка данных
│
├── internal/                     # Внутренняя логика (приватная)
│   │
│   ├── app/                      # 🎛️ Ядро приложения
│   │   ├── server.go            # Инициализация сервера, graceful shutdown
│   │   └── routes_modular.go    # Регистрация всех модулей и маршрутов
│   │
│   ├── database/                 # 💾 Подключение к БД
│   │   └── database.go          # GORM connection pool
│   │
│   ├── middleware/               # 🛡️ HTTP Middleware
│   │   └── auth.go              # JWT Auth, AdminMiddleware, SuperAdminMiddleware
│   │
│   ├── models/                   # 📦 Общие модели данных
│   │   ├── user.go              # User, Role (super_admin, admin, home_chef, etc.)
│   │   ├── ingredient.go        # Ingredient, Category
│   │   ├── recipe.go            # Recipe, RecipeIngredient
│   │   ├── fridge.go            # UserFridgeItem
│   │   ├── business.go          # Business, Marketplace
│   │   └── ...                  # Другие доменные модели
│   │
│   ├── modules/                  # 🧩 Бизнес-модули (DDD)
│   │   │
│   │   ├── auth/                # 🔐 Аутентификация
│   │   │   ├── module.go       # Инициализация модуля
│   │   │   ├── service/        # AuthService (login, register, verify)
│   │   │   ├── repo/           # AuthRepository (user queries)
│   │   │   └── transport/http/ # HTTP handlers (POST /auth/login, etc.)
│   │   │
│   │   ├── user/                # 👤 Управление пользователями
│   │   │   ├── module.go
│   │   │   ├── service/        # UserService (profile, settings)
│   │   │   ├── repo/           # UserRepository
│   │   │   └── transport/http/ # GET /api/users/me, PATCH /api/users/profile
│   │   │
│   │   ├── admin/               # 👑 Админ-панель
│   │   │   ├── module.go
│   │   │   ├── service/        # AdminService (user management, stats)
│   │   │   │   └── service.go  # GetUsersWithFilters, UpdateUserRole, GetAllIngredients
│   │   │   ├── repo/           # AdminRepository
│   │   │   └── transport/http/ # GET /api/admin/users, DELETE /api/admin/users/:id
│   │   │       └── handlers.go # GetAllUsers, GetIngredientsStats, etc.
│   │   │
│   │   ├── fridge/              # 🧊 Холодильник пользователя
│   │   │   ├── module.go
│   │   │   ├── service/        # FridgeService (add/remove items, check expiry)
│   │   │   ├── repo/           # FridgeRepository
│   │   │   └── transport/http/ # GET /api/fridge, POST /api/fridge/items
│   │   │
│   │   ├── recipes/             # 📖 Рецепты
│   │   │   ├── module.go
│   │   │   ├── service/        # RecipeService (catalog, match, recommendations)
│   │   │   │   ├── service.go          # Core recipe service
│   │   │   │   ├── catalog_service.go  # Recipe catalog with filters
│   │   │   │   ├── match_service.go    # Match recipes with fridge
│   │   │   │   └── exclude_service.go  # User exclusions
│   │   │   ├── repo/           # RecipeRepository
│   │   │   └── transport/http/ # GET /api/recipes, GET /api/recipes/match
│   │   │
│   │   ├── ingredients/         # 🥕 Каталог ингредиентов
│   │   │   ├── module.go
│   │   │   ├── service/        # IngredientsService (search, catalog)
│   │   │   ├── repo/           # IngredientsRepository
│   │   │   └── transport/http/ # GET /api/catalog/ingredients?search=говядина
│   │   │
│   │   ├── ai/                  # 🤖 AI-интеграция
│   │   │   ├── module.go
│   │   │   ├── service/        # AIService (GROQ API, recipe adaptation)
│   │   │   │   ├── service.go          # Main AI service
│   │   │   │   ├── groq_service.go     # GROQ API client
│   │   │   │   └── adaptation_service.go # Recipe adaptation logic
│   │   │   └── transport/http/ # POST /api/ai/adapt, POST /api/ai/suggestions
│   │   │
│   │   ├── ai_recommendations/  # 🧠 AI Рекомендации (Decision Engine)
│   │   │   ├── module.go
│   │   │   ├── service/        # AI decision logic (что готовить)
│   │   │   └── transport/http/ # GET /api/ai/recommendations
│   │   │
│   │   ├── budget/              # 💰 Бюджет пользователя
│   │   │   ├── module.go
│   │   │   ├── service/        # BudgetService (weekly tracking)
│   │   │   └── transport/http/ # GET /api/budget, POST /api/budget/weekly
│   │   │
│   │   ├── meal_plan/           # 📅 План питания
│   │   │   ├── module.go
│   │   │   ├── service/        # MealPlanService (weekly/daily plans)
│   │   │   └── transport/http/ # GET /api/meal-plan, POST /api/meal-plan/generate
│   │   │
│   │   ├── marketplace/         # 🛒 Marketplace (B2B)
│   │   │   ├── module.go
│   │   │   ├── service/        # MarketplaceService (businesses, orders)
│   │   │   └── transport/http/ # GET /api/marketplace/businesses
│   │   │
│   │   ├── business/            # 🏢 Бизнес-аккаунты
│   │   │   ├── module.go
│   │   │   ├── service/        # BusinessService (registration, products)
│   │   │   └── transport/http/ # POST /api/business/register
│   │   │
│   │   ├── prepared_dishes/     # 🍽️ Готовые блюда (после готовки)
│   │   │   ├── module.go
│   │   │   ├── service/        # PreparedDishesService (tracking cooked recipes)
│   │   │   └── transport/http/ # POST /api/prepared-dishes
│   │   │
│   │   ├── wallet/              # 💳 Кошелёк и токены
│   │   │   ├── module.go
│   │   │   ├── service/        # WalletService (balance, transactions)
│   │   │   └── transport/http/ # GET /api/wallet, POST /api/wallet/topup
│   │   │
│   │   ├── leaderboard/         # 🏆 Таблица лидеров
│   │   │   ├── module.go
│   │   │   ├── service/        # LeaderboardService (top users)
│   │   │   └── transport/http/ # GET /api/leaderboard
│   │   │
│   │   ├── history/             # 📊 История активности
│   │   │   ├── module.go
│   │   │   ├── service/        # HistoryService (user activity analytics)
│   │   │   └── transport/http/ # GET /api/history/activity
│   │   │
│   │   ├── stats/               # 📈 Статистика системы
│   │   │   ├── module.go
│   │   │   ├── service/        # StatsService (global metrics)
│   │   │   └── transport/http/ # GET /api/stats/overview
│   │   │
│   │   ├── health/              # ❤️ Health checks
│   │   │   └── transport/http/ # GET /health, GET /health/db
│   │   │
│   │   ├── contact/             # 📧 Контактная форма
│   │   │   └── transport/http/ # POST /api/contact
│   │   │
│   │   ├── hint/                # 💡 Подсказки пользователям
│   │   │   └── transport/http/ # GET /api/hints
│   │   │
│   │   ├── metrics/             # 📊 Метрики (Prometheus-style)
│   │   │   └── transport/http/ # GET /metrics
│   │   │
│   │   ├── nutrition/           # 🥗 Нутриционный анализ
│   │   │   └── transport/http/ # GET /api/nutrition/analyze
│   │   │
│   │   └── websocket/           # 🔄 WebSocket (real-time)
│   │       ├── module.go
│   │       └── transport/http/ # WS /api/ws
│   │
│   ├── platform/                 # 🔧 Платформенные утилиты
│   │   ├── config/              # Конфигурация (.env)
│   │   └── logger/              # Структурированное логирование (zap)
│   │
│   └── pkg/                      # 📦 Переиспользуемые пакеты
│       └── utils/               # Хелперы (RespondWithJSON, ValidateEmail, etc.)
│
├── migrations/                   # 🗄️ SQL миграции
│   ├── 001_create_users_table.sql
│   ├── 058_add_russian_translations.sql
│   ├── 059_add_user_status_tracking.sql
│   ├── 060_add_super_admin_role.sql
│   └── ...
│
├── docs/                         # 📚 Документация
│   ├── PROJECT_ARCHITECTURE.md  # Этот файл
│   ├── RECIPE_SYSTEM_SUMMARY.md
│   ├── ECONOMY_CALCULATION.md
│   └── ...
│
├── Dockerfile                    # 🐳 Docker образ
├── Makefile                      # ⚙️ Команды для сборки
├── go.mod                        # 📦 Go зависимости
└── .env                          # 🔐 Переменные окружения

```

---

## 🔄 Архитектурный Flow

### 1. **Запуск приложения**

```
cmd/server/main.go
    ↓
internal/app/server.go (New)
    ↓
- Load config (.env)
- Init logger (zap)
- Connect to DB (GORM)
- Setup routes (routes_modular.go)
    ↓
HTTP Server starts on :8080
```

### 2. **Регистрация модулей**

```go
// internal/app/routes_modular.go

func (a *App) setupModularRoutes() http.Handler {
    r := chi.NewRouter()
    
    // Global middleware
    r.Use(chimiddleware.Logger)
    r.Use(cors.Handler(...))
    
    // Initialize modules
    authModule := auth.NewModule()
    adminModule := admin.NewModule()
    recipesModule := recipes.NewModule(a.db)
    
    // Register routes
    authModule.RegisterRoutes(r)
    adminModule.RegisterRoutes(r, middleware.AuthMiddleware, middleware.AdminMiddleware)
    recipesModule.RegisterRoutes(r, middleware.AuthMiddleware)
    
    return r
}
```

### 3. **Типичный HTTP запрос (пример: Admin Panel)**

```
1. HTTP Request: GET /api/admin/users?role=super_admin&limit=1000
   ↓
2. Chi Router → /api/admin/users handler
   ↓
3. Middleware Chain:
   - AuthMiddleware (JWT validation)
   - AdminMiddleware (check role = admin OR super_admin)
   ↓
4. Handler: AdminHandlers.GetAllUsers(w, r)
   ↓
5. Service: adminService.GetUsersWithFilters(role, status, search, limit)
   ↓
6. Repository: db.Where(...).Find(&users)
   ↓
7. Database: PostgreSQL query
   ↓
8. Response: utils.RespondWithJSON(w, 200, {data: [...], meta: {total: 54}})
   ↓
9. HTTP Response → Frontend
```

---

## 🧩 Модульная структура (на примере Admin)

```
internal/modules/admin/
│
├── module.go                     # Инициализация модуля
│   - NewModule() *Module
│   - RegisterRoutes(r chi.Router, middlewares...)
│
├── service/                      # Бизнес-логика
│   ├── interface.go             # AdminService interface
│   └── service.go               # AdminService implementation
│       - GetUsersWithFilters()
│       - UpdateUserRole()
│       - GetAllIngredients()
│       - GetIngredientsStats()
│
├── repo/                         # Data Access Layer
│   └── repository.go
│       - FindUserByID()
│       - UpdateUser()
│
└── transport/http/               # HTTP Layer
    └── handlers.go
        - GetAllUsers(w, r)
        - DeleteUser(w, r)
        - GetAllIngredients(w, r)
```

### Пример кода модуля:

```go
// module.go
type Module struct {
    service  service.AdminService
    handlers *http.AdminHandlers
}

func NewModule() *Module {
    repo := repo.NewAdminRepository()
    svc := service.NewAdminService(database.DB, repo)
    handlers := http.NewAdminHandlers(svc)
    
    return &Module{
        service:  svc,
        handlers: handlers,
    }
}

func (m *Module) RegisterRoutes(r chi.Router, authMW, adminMW, superAdminMW func(http.Handler) http.Handler) {
    r.Route("/api/admin", func(r chi.Router) {
        r.Use(authMW)     // JWT required
        r.Use(adminMW)    // Admin role required
        
        // Public admin routes
        r.Get("/users", m.handlers.GetAllUsers)
        r.Get("/users/stats", m.handlers.GetUsersStats)
        r.Get("/ingredients", m.handlers.GetAllIngredients)
        
        // Super admin only
        r.With(superAdminMW).Delete("/users/{id}", m.handlers.DeleteUser)
        r.With(superAdminMW).Patch("/users/update-role", m.handlers.UpdateUserRole)
    })
}
```

---

## 🔐 Middleware цепочка

```
Request
  ↓
[RequestID]         → Генерирует уникальный ID запроса
  ↓
[Logger]            → Логирует HTTP запросы
  ↓
[Recoverer]         → Ловит панику
  ↓
[Timeout(60s)]      → Таймаут запроса
  ↓
[CORS]              → Разрешает запросы с фронтенда
  ↓
[AuthMiddleware]    → Проверяет JWT токен (если требуется)
  ↓
[AdminMiddleware]   → Проверяет role = admin OR super_admin (если требуется)
  ↓
[SuperAdminMiddleware] → Проверяет role = super_admin (для критических операций)
  ↓
Handler             → Обработчик запроса
```

---

## 📊 База данных (основные таблицы)

```sql
-- 👤 Users (54 total)
users:
  - id UUID PRIMARY KEY
  - email VARCHAR UNIQUE
  - role VARCHAR (super_admin, admin, home_chef, pro_chef, investor)
  - status VARCHAR (active, inactive, pending)
  - last_login TIMESTAMP
  - created_at TIMESTAMP

-- 🥕 Ingredients (211 total)
ingredients:
  - id UUID PRIMARY KEY
  - name VARCHAR
  - name_en VARCHAR
  - name_ru VARCHAR
  - category VARCHAR (protein, vegetable, dairy, grain, condiment, other)
  - price_per_unit NUMERIC

-- 📖 Recipes (70 total)
recipes:
  - id UUID PRIMARY KEY
  - title VARCHAR
  - description TEXT
  - cook_time INTEGER
  - difficulty VARCHAR (easy, medium, hard)
  - cuisine VARCHAR
  - category VARCHAR

-- 🔗 Recipe Ingredients (many-to-many)
recipe_ingredients:
  - recipe_id UUID
  - ingredient_id UUID
  - quantity NUMERIC
  - unit VARCHAR

-- 🧊 User Fridge
user_fridge_items:
  - id UUID PRIMARY KEY
  - user_id UUID
  - ingredient_id UUID
  - quantity NUMERIC
  - expiry_date DATE
```

---

## 🎯 Ключевые возможности системы

### 1. **Многоуровневая ролевая система**
```
super_admin → Полный доступ (удаление пользователей, изменение ролей)
admin       → Управление пользователями (просмотр, статистика)
home_chef   → Базовые функции (рецепты, холодильник)
pro_chef    → Расширенные функции
investor    → Специальные права
```

### 2. **Умный поиск ингредиентов**
- Поддержка кириллицы (говядина, молоко)
- Нормализация запросов (LOWER, TRIM)
- Полнотекстовый поиск (GIN индексы)

### 3. **AI-рекомендации**
- Адаптация рецептов под холодильник
- GROQ API интеграция
- Self-repair паттерн для надёжности

### 4. **Фильтрация и пагинация**
- Динамическая генерация SQL запросов
- Поддержка множественных фильтров (role, status, search)
- Limit до 1000 записей

### 5. **Real-time обновления**
- WebSocket для live уведомлений
- Event-driven architecture

---

## 🚀 Deployment Pipeline

```
Local Development
  ↓
Git push to main
  ↓
Koyeb auto-deploy
  ↓
Neon PostgreSQL (production DB)
  ↓
Server running on Koyeb (port 8080)
  ↓
Frontend: menu-fodi.vercel.app
```

---

## 📦 Основные зависимости

```go
// HTTP Router
github.com/go-chi/chi/v5

// Database ORM
gorm.io/gorm
gorm.io/driver/postgres

// Authentication
github.com/golang-jwt/jwt/v5
golang.org/x/crypto (bcrypt)

// Logging
go.uber.org/zap

// Configuration
github.com/joho/godotenv

// UUID
github.com/google/uuid

// PDF Generation
github.com/jung-kurt/gofpdf
```

---

## 🔧 Команды для разработки

```bash
# Сборка
make build
go build -o bin/server ./cmd/server

# Запуск
make run
./bin/server

# Миграции
make migrate-up
go run cmd/migrate/main.go

# Тесты
make test
go test ./...

# Проверка активности пользователей
./check_activity.sh
```

---

## 📈 Метрики системы (текущее состояние)

```
👥 Users:        54 total (1 super_admin, 3 admin, 49 home_chef, 1 investor)
🥕 Ingredients:  211 total (6 categories)
📖 Recipes:      70 total
🧊 Fridge Items: Active user data
📊 Active Users: 25.9% (last 30 days)
```

---

## 🎓 Лучшие практики

### 1. **Dependency Injection**
```go
// ❌ BAD: Creating dependencies inside
func NewService() *Service {
    db := database.Connect()  // Hard coupling
    return &Service{db: db}
}

// ✅ GOOD: Injecting dependencies
func NewService(db *gorm.DB) *Service {
    return &Service{db: db}
}
```

### 2. **Interface-based design**
```go
// Define interface in service package
type AdminService interface {
    GetUsers() ([]User, error)
    DeleteUser(id string) error
}

// Implementation in service package
type adminService struct {
    db *gorm.DB
}

func (s *adminService) GetUsers() ([]User, error) { ... }
```

### 3. **Error handling**
```go
// Always wrap errors with context
if err := s.db.Find(&users).Error; err != nil {
    return nil, fmt.Errorf("failed to fetch users: %w", err)
}
```

### 4. **Graceful shutdown**
```go
// Handle SIGINT/SIGTERM
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

if err := server.Shutdown(ctx); err != nil {
    logger.Error("Server forced to shutdown", zap.Error(err))
}
```

---

## 🔮 Будущие улучшения

1. **Кэширование** (Redis для hot data)
2. **Rate Limiting** (защита от DDoS)
3. **Мониторинг** (Prometheus + Grafana)
4. **CI/CD** (GitHub Actions)
5. **Тесты** (unit + integration)
6. **Документация API** (Swagger/OpenAPI)

---

## 📞 Контакты

**Разработчик:** Dmitrij Fomin  
**Проект:** Chef Academy Backend  
**GitHub:** github.com/Fodi999/menu_fodi_backend

---

**Последнее обновление:** 4 января 2026 г.
