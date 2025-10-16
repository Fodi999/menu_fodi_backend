# 🏗️ Структура кода Menu Fodi Backend

## 📊 Визуальная архитектура

```
┌─────────────────────────────────────────────────────────────────┐
│                        CLIENT REQUESTS                          │
│          (Web, Mobile, Elixir Bot, Admin Panel)                 │
└────────────────────────┬────────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────────┐
│                      HTTP SERVER :8080                          │
│                    (Gorilla Mux Router)                         │
│                         + CORS                                  │
└────────────────────────┬────────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────────┐
│                      MIDDLEWARE LAYER                           │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐         │
│  │  API Key     │  │   JWT Auth   │  │  Admin Auth  │         │
│  │ (apikey.go)  │  │  (auth.go)   │  │  (auth.go)   │         │
│  └──────────────┘  └──────────────┘  └──────────────┘         │
└────────────────────────┬────────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────────┐
│                      HANDLERS LAYER                             │
│                    (HTTP Controllers)                           │
│                                                                 │
│  ┌────────────────┐  ┌────────────────┐  ┌────────────────┐   │
│  │  auth.go       │  │  products.go   │  │  orders.go     │   │
│  │  - Register    │  │  - GetProducts │  │  - CreateOrder │   │
│  │  - Login       │  │  - GetByID     │  │  - GetOrders   │   │
│  │  - VerifyToken │  │  - CRUD (admin)│  │  - UpdateStatus│   │
│  └────────────────┘  └────────────────┘  └────────────────┘   │
│                                                                 │
│  ┌────────────────┐  ┌────────────────┐  ┌────────────────┐   │
│  │ ingredients.go │  │semi_finished.go│  │   admin.go     │   │
│  │  - CRUD        │  │  - CRUD        │  │  - UserMgmt    │   │
│  │  - Movements   │  │  - Recipes     │  │  - Stats       │   │
│  └────────────────┘  └────────────────┘  └────────────────┘   │
│                                                                 │
│  ┌────────────────┐  ┌────────────────┐  ┌────────────────┐   │
│  │   hint.go      │  │  health.go     │  │ websocket.go   │   │
│  │  - BotSearch   │  │  - HealthCheck │  │  - Hub         │   │
│  │  - Suggestions │  │  - DBStatus    │  │  - Broadcast   │   │
│  └────────────────┘  └────────────────┘  └────────────────┘   │
└────────────────────────┬────────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────────┐
│                     BUSINESS LOGIC LAYER                        │
│                                                                 │
│  ┌────────────────────────────────────────────────────────┐    │
│  │              AUTH MODULE (auth/jwt.go)                 │    │
│  │  - GenerateToken(userID, role) -> JWT                  │    │
│  │  - ValidateToken(token) -> claims, error               │    │
│  │  - Secret: SECRET_KEY env, Expiry: 72h                 │    │
│  └────────────────────────────────────────────────────────┘    │
│                                                                 │
│  ┌────────────────────────────────────────────────────────┐    │
│  │           REPOSITORIES (database/)                     │    │
│  │  - user_repository.go: FindByEmail, Create, Update     │    │
│  │  - ingredient_repository.go: CRUD, Search, Movements   │    │
│  └────────────────────────────────────────────────────────┘    │
└────────────────────────┬────────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────────┐
│                      MODELS LAYER (GORM)                        │
│                                                                 │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐            │
│  │   User      │  │  Product    │  │   Order     │            │
│  │  - ID       │  │  - ID       │  │  - ID       │            │
│  │  - Email    │  │  - Name     │  │  - UserID   │            │
│  │  - Password │  │  - Price    │  │  - Total    │            │
│  │  - Role     │  │  - Category │  │  - Status   │            │
│  └─────────────┘  └─────────────┘  └─────────────┘            │
│                                                                 │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐            │
│  │ Ingredient  │  │SemiFinished │  │ OrderItem   │            │
│  │  - ID       │  │  - ID       │  │  - ID       │            │
│  │  - Name     │  │  - Name     │  │  - OrderID  │            │
│  │  - Quantity │  │  - Cost     │  │  - Quantity │            │
│  │  - Unit     │  │  - Recipe   │  │  - Price    │            │
│  └─────────────┘  └─────────────┘  └─────────────┘            │
└────────────────────────┬────────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────────┐
│                    DATABASE LAYER (db.go)                       │
│                                                                 │
│  ┌────────────────────────────────────────────────────────┐    │
│  │              PostgreSQL (Neon Cloud)                   │    │
│  │  - Connection Pool                                     │    │
│  │  - Auto Migration (GORM)                               │    │
│  │  - SQL Logging                                         │    │
│  │  - Tables: User, Product, Order, Ingredient, etc.     │    │
│  └────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────┘
```

## 📁 Детальная структура файлов

### 🚀 **cmd/** - Точки входа приложения

```
cmd/
├── server/main.go          # 🎯 ГЛАВНЫЙ ФАЙЛ - запуск HTTP сервера
│   ├── InitDB()           # Инициализация базы данных
│   ├── CORS setup         # Настройка CORS для фронтенда
│   ├── Router setup       # Регистрация всех 30 endpoints
│   ├── WebSocket Hub      # Запуск real-time уведомлений
│   └── http.ListenAndServe(":8080")
│
└── migrate/main.go         # 🔄 Утилита миграций БД
    └── Запуск: go run cmd/migrate/main.go
```

### 🔐 **internal/auth/** - Аутентификация

```
auth/
└── jwt.go
    ├── GenerateToken(userID string, role string) -> string, error
    │   ├── Создаёт JWT с claims: userID, role
    │   ├── Срок действия: 72 часа
    │   └── Подпись: SECRET_KEY из env
    │
    └── ValidateToken(tokenString string) -> jwt.MapClaims, error
        ├── Проверяет подпись
        ├── Проверяет срок действия
        └── Возвращает claims (userID, role)
```

### 💾 **internal/database/** - Работа с БД

```
database/
├── db.go
│   ├── InitDB() -> *gorm.DB
│   │   ├── Подключение к PostgreSQL (Neon)
│   │   ├── AutoMigrate всех моделей
│   │   └── Настройка логирования SQL
│   │
│   └── Global: database.DB (*gorm.DB)
│
├── user_repository.go
│   ├── FindByEmail(email string) -> User, error
│   ├── CreateUser(user *User) -> error
│   └── UpdateUser(user *User) -> error
│
└── ingredient_repository.go
    ├── GetAllIngredients() -> []Ingredient
    ├── SearchByName(name string) -> []Ingredient
    └── GetMovements() -> []Movement
```

### 🎯 **internal/handlers/** - HTTP обработчики

```
handlers/
├── auth.go                    # Аутентификация пользователей
│   ├── Register(POST)        # Регистрация нового пользователя
│   │   └── Request: {email, password, name}
│   ├── Login(POST)           # Вход в систему
│   │   └── Returns: {token, user}
│   └── VerifyToken(POST)     # Проверка валидности токена
│
├── products.go               # Управление продуктами
│   ├── GetProducts(GET)      # PUBLIC: список видимых продуктов
│   ├── GetProductByID(GET)   # PUBLIC: продукт по ID
│   ├── GetAllProducts(GET)   # ADMIN: все продукты + скрытые
│   ├── CreateProduct(POST)   # ADMIN: создать продукт
│   ├── UpdateProduct(PUT)    # ADMIN: обновить продукт
│   └── DeleteProduct(DELETE) # ADMIN: удалить продукт
│
├── orders.go                 # Управление заказами
│   ├── CreateOrder(POST)     # Создать заказ (гость или user)
│   │   ├── Auto-calculate total = Σ(price × quantity)
│   │   ├── Round to 2 decimals
│   │   └── Send WebSocket notification
│   ├── GetUserOrders(GET)    # USER: свои заказы
│   ├── GetAllOrders(GET)     # ADMIN: все заказы
│   ├── GetRecentOrders(GET)  # ADMIN: последние заказы
│   └── UpdateOrderStatus(PUT)# ADMIN: изменить статус
│
├── ingredients.go            # Управление ингредиентами (ADMIN)
│   ├── GetIngredients(GET)
│   ├── CreateIngredient(POST)
│   ├── UpdateIngredient(PUT)
│   ├── DeleteIngredient(DELETE)
│   └── GetIngredientMovements(GET)
│
├── semi_finished.go          # Полуфабрикаты (ADMIN)
│   ├── GetSemiFinished(GET)
│   ├── CreateSemiFinished(POST)
│   ├── GetSemiFinishedByID(GET)
│   ├── UpdateSemiFinished(PUT)
│   └── DeleteSemiFinished(DELETE)
│
├── admin.go                  # Админ панель
│   ├── GetAllUsers(GET)      # Список всех пользователей
│   ├── UpdateUser(PUT)       # Обновить пользователя
│   └── DeleteUser(DELETE)    # Удалить пользователя
│
├── stats.go                  # Статистика (ADMIN)
│   └── GetStats(GET)         # Общая статистика
│       └── Returns: {total_users, total_orders, revenue, etc.}
│
├── hint.go                   # 🤖 Endpoint для Elixir бота
│   └── HintHandler(POST)     # /api/hint
│       ├── Requires: X-API-Key header
│       ├── Search: LOWER(name) LIKE OR LOWER(category) LIKE
│       ├── Limit: 5 products
│       └── Response: {hint, suggested_products[]}
│
├── health.go                 # Health checks
│   ├── HealthHandler(GET)    # /health
│   └── APIHealthHandler(GET) # /api/health
│       └── Response: {status: "ok", data: {service, database}}
│
└── websocket.go              # Real-time уведомления
    ├── Hub struct            # Управление WebSocket клиентами
    ├── NewHub()              # Создать хаб
    ├── Run()                 # Запустить горутину хаба
    ├── HandleWebSocket()     # Подключение клиента
    └── BroadcastNewOrder()   # Отправка уведомлений о заказах
```

### 🛡️ **internal/middleware/** - Промежуточные обработчики

```
middleware/
├── apikey.go                          # API Key аутентификация
│   └── APIKeyMiddleware()
│       ├── Извлекает X-API-Key header
│       ├── Сравнивает с ELIXIR_API_KEY env
│       ├── Пропускает OPTIONS (CORS)
│       └── Returns 401 если invalid
│       └── Используется для: /api/hint
│
└── auth.go                            # JWT аутентификация
    ├── AuthMiddleware()
    │   ├── Извлекает Bearer token
    │   ├── Валидирует через auth.ValidateToken()
    │   └── Добавляет userID в request context
    │
    └── AdminMiddleware()
        ├── Проверяет role == "admin"
        └── Returns 403 если не admin
```

### 📊 **internal/models/** - Модели данных (GORM)

```
models/
├── user.go
│   └── User struct
│       ├── ID        uuid (PK)
│       ├── Email     string (unique)
│       ├── Name      string
│       ├── Password  string (bcrypt hashed)
│       ├── Role      string ("user" | "admin")
│       └── CreatedAt time.Time
│       └── Methods:
│           ├── HashPassword() -> error
│           └── CheckPassword(password) -> error
│
├── product.go
│   └── Product struct
│       ├── ID          uuid (PK)
│       ├── Name        string
│       ├── Description string
│       ├── Price       float64
│       ├── ImageURL    string
│       ├── Weight      int (граммы)
│       ├── Category    string
│       ├── IsVisible   bool
│       ├── Ingredients []ProductIngredient (relations)
│       └── CreatedAt   time.Time
│
├── order.go
│   ├── Order struct
│   │   ├── ID        uuid (PK)
│   │   ├── UserID    *string (nullable - гостевые заказы)
│   │   ├── Name      string
│   │   ├── Status    string ("pending", "confirmed", etc.)
│   │   ├── Total     float64 (auto-calculated)
│   │   ├── Address   string
│   │   ├── Phone     string
│   │   ├── Comment   string
│   │   ├── Items     []OrderItem (has many)
│   │   ├── CreatedAt time.Time
│   │   └── UpdatedAt time.Time
│   │
│   └── OrderItem struct
│       ├── ID        uuid (PK)
│       ├── OrderID   string (FK)
│       ├── ProductID string (FK)
│       ├── Quantity  int
│       └── Price     float64
│
├── ingredient.go
│   └── Ingredient struct
│       ├── ID           uuid (PK)
│       ├── Name         string
│       ├── Unit         string ("кг", "л", "шт")
│       ├── Quantity     float64
│       ├── MinQuantity  float64
│       ├── PricePerUnit float64
│       ├── Supplier     string
│       └── CreatedAt    time.Time
│
└── semi_finished.go
    └── SemiFinished struct
        ├── ID            uuid (PK)
        ├── Name          string
        ├── Description   string
        ├── OutputQuantity float64
        ├── OutputUnit    string
        ├── CostPerUnit   float64
        ├── TotalCost     float64
        ├── Category      string
        ├── IsVisible     bool
        ├── IsArchived    bool
        ├── Ingredients   []SemiFinishedIngredient (relations)
        ├── CreatedAt     time.Time
        ├── UpdatedAt     time.Time
        └── DeletedAt     *time.Time (soft delete)
```

### 📦 **pkg/utils/** - Утилиты

```
utils/
└── response.go
    ├── JSONResponse(w, statusCode, data)
    ├── ErrorResponse(w, statusCode, message)
    └── SuccessResponse(w, data)
    └── Стандартный формат: {status: "ok/error", data/message: ...}
```

### 🔄 **migrations/** - SQL миграции

```
migrations/
├── 003_add_semi_finished_fields.sql
├── add_price_per_unit.sql
├── create_semi_finished_ingredients.sql
└── create_semi_finished_tables.sql
```

## 🔄 Поток данных

### Создание заказа (CreateOrder)

```
1. Client Request
   POST /api/orders
   Body: {name, phone, address, items: [{productId, quantity, price}]}
   
   ↓

2. Handler: orders.go::CreateOrder()
   ├── Валидация input (name, phone, address не пустые)
   ├── Проверка items.length > 0
   ├── Извлечение userID из context (опционально)
   │
   ├── 🧮 Расчёт total:
   │   var total float64
   │   for item in items:
   │       total += item.Price × item.Quantity
   │   total = math.Round(total × 100) / 100  // округление
   │
   ├── Создание Order объекта
   │   ├── ID: uuid.New()
   │   ├── UserID: из context или nil (гость)
   │   ├── Total: calculated total
   │   └── Status: "pending"
   │
   ├── 💾 Transaction:
   │   ├── tx.Create(&order)
   │   ├── for each item:
   │   │   └── tx.Create(&orderItem)
   │   └── tx.Commit()
   │
   ├── 📡 WebSocket Broadcast
   │   └── BroadcastOrderNotification("new_order", orderData)
   │
   └── Response
       {
         message: "Order created successfully",
         orderId: uuid,
         total: 350.00,
         status: "pending"
       }
```

### JWT Authentication Flow

```
1. User Registration/Login
   ↓
2. auth.go::Login()
   ├── Validate credentials
   ├── auth.GenerateToken(userID, role)
   └── Return: {token: "eyJhbG...", user: {...}}
   
   ↓

3. Client stores token
   └── localStorage.setItem('token', token)
   
   ↓

4. Protected Request
   GET /api/user/orders
   Headers: Authorization: Bearer eyJhbG...
   
   ↓

5. middleware.AuthMiddleware()
   ├── Extract token from header
   ├── auth.ValidateToken(token)
   ├── Add userID to context
   └── next()
   
   ↓

6. Handler: orders.go::GetUserOrders()
   ├── userID := context.Value("userID")
   ├── db.Where("user_id = ?", userID).Find(&orders)
   └── Return orders
```

## 🔌 Endpoints Flow Map

```
PUBLIC ENDPOINTS (без auth):
┌──────────────────────────────────────────┐
│ GET  /health                              │ → health.go::HealthHandler
│ GET  /api/health                          │ → health.go::APIHealthHandler
│ POST /api/register                        │ → auth.go::Register
│ POST /api/login                           │ → auth.go::Login
│ POST /api/verify-token                    │ → auth.go::VerifyToken
│ GET  /api/products                        │ → products.go::GetProducts
│ GET  /api/products/{id}                   │ → products.go::GetProductByID
│ POST /api/orders                          │ → orders.go::CreateOrder
│ POST /api/hint                            │ → hint.go::HintHandler (API Key)
└──────────────────────────────────────────┘

PROTECTED ENDPOINTS (JWT required):
┌──────────────────────────────────────────┐
│ GET  /api/user/profile                    │ → auth.go::GetProfile
│ PUT  /api/user/profile                    │ → auth.go::UpdateProfile
│ GET  /api/user/orders                     │ → orders.go::GetUserOrders
└──────────────────────────────────────────┘

ADMIN ENDPOINTS (JWT + Admin role):
┌──────────────────────────────────────────┐
│ GET  /api/admin/users                     │ → admin.go::GetAllUsers
│ PUT  /api/admin/users/{id}                │ → admin.go::UpdateUser
│ DELETE /api/admin/users/{id}              │ → admin.go::DeleteUser
│ GET  /api/admin/orders                    │ → orders.go::GetAllOrders
│ GET  /api/admin/orders/recent             │ → orders.go::GetRecentOrders
│ PUT  /api/admin/orders/{id}/status        │ → orders.go::UpdateOrderStatus
│ GET  /api/admin/ingredients               │ → ingredients.go::GetIngredients
│ POST /api/admin/ingredients               │ → ingredients.go::CreateIngredient
│ PUT  /api/admin/ingredients/{id}          │ → ingredients.go::UpdateIngredient
│ DELETE /api/admin/ingredients/{id}        │ → ingredients.go::DeleteIngredient
│ GET  /api/admin/ingredients/movements     │ → ingredients.go::GetMovements
│ GET  /api/admin/semi-finished             │ → semi_finished.go::GetSemiFinished
│ POST /api/admin/semi-finished             │ → semi_finished.go::Create
│ GET  /api/admin/semi-finished/{id}        │ → semi_finished.go::GetByID
│ PUT  /api/admin/semi-finished/{id}        │ → semi_finished.go::Update
│ DELETE /api/admin/semi-finished/{id}      │ → semi_finished.go::Delete
│ GET  /api/admin/products                  │ → products.go::GetAllProducts
│ POST /api/admin/products                  │ → products.go::CreateProduct
│ PUT  /api/admin/products/{id}             │ → products.go::UpdateProduct
│ DELETE /api/admin/products/{id}           │ → products.go::DeleteProduct
│ GET  /api/admin/stats                     │ → stats.go::GetStats
│ WS   /api/admin/ws                        │ → websocket.go::HandleWebSocket
└──────────────────────────────────────────┘
```

## 📈 Database Schema Relations

```
┌──────────┐
│   User   │
└────┬─────┘
     │ 1
     │
     │ *
┌────┴─────┐
│  Order   │ ←──────────────┐
└────┬─────┘                │
     │ 1                    │
     │                      │ WebSocket
     │ *               ┌────┴────┐
┌────┴─────────┐       │   Hub   │
│  OrderItem   │       └─────────┘
└────┬─────────┘
     │ *
     │
     │ 1
┌────┴─────────┐
│   Product    │
└────┬────┬────┘
     │ *  │ *
     │    │
     │    └──────────────────────┐
     │                           │
┌────┴──────────────┐  ┌─────────┴──────────┐
│ProductIngredient  │  │ProductSemiFinished │
└────┬──────────────┘  └─────────┬──────────┘
     │ *                         │ *
     │ 1                         │ 1
┌────┴──────────┐      ┌─────────┴──────────┐
│  Ingredient   │      │   SemiFinished     │
└───────────────┘      └────────┬───────────┘
                                │ *
                                │ 1
                       ┌────────┴───────────────┐
                       │SemiFinishedIngredient  │
                       └────────┬───────────────┘
                                │ *
                                │ 1
                       ┌────────┴──────────┐
                       │   Ingredient      │
                       └───────────────────┘
```

## 🚀 Как запустить

```bash
# 1. Запуск сервера (разработка)
go run cmd/server/main.go

# 2. Запуск миграций
go run cmd/migrate/main.go

# 3. Компиляция
go build -o bin/server cmd/server/main.go
./bin/server

# 4. Docker
docker build -t menu-fodi-backend .
docker run -p 8080:8080 menu-fodi-backend
```

## 📝 Environment Variables

```env
DATABASE_URL="postgresql://..."
SECRET_KEY="your-jwt-secret"
ELIXIR_API_KEY="supersecret"
PORT=8080
```

---

**Автор:** Menu Fodi Backend Team  
**Версия:** 1.0  
**Дата:** 15 октября 2025 г.
