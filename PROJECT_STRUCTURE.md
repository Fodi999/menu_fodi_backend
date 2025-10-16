# 📂 Структура проекта Menu Fodi Backend

## 🗂️ Полное описание всех файлов

```
backend/
│
├── 📁 bin/                                    # Скомпилированные бинарные файлы
│   └── server                                 # Исполняемый файл сервера (после go build)
│
├── 📁 cmd/                                    # Точки входа приложения (commands)
│   ├── 📁 migrate/
│   │   └── main.go                            # Утилита для запуска миграций БД
│   │                                          # Команда: go run cmd/migrate/main.go
│   │
│   └── 📁 server/
│       └── main.go                            # 🚀 ГЛАВНЫЙ ФАЙЛ СЕРВЕРА
│                                              # - Инициализация БД
│                                              # - Настройка CORS (localhost:3000, 3001, 4000, Vercel)
│                                              # - Регистрация всех 30 роутов
│                                              # - Запуск WebSocket Hub
│                                              # - Старт HTTP сервера на :8080
│
├── 📁 config/                                 # Конфигурационные файлы (пусто)
│                                              # Можно добавить: config.yaml, .env loader
│
├── 📁 internal/                               # Внутренняя бизнес-логика (не экспортируется)
│   │
│   ├── 📁 auth/                               # 🔐 Модуль аутентификации
│   │   └── jwt.go                             # JWT токены
│   │                                          # - GenerateToken(userID, role string) -> token
│   │                                          # - ValidateToken(tokenString) -> claims, error
│   │                                          # - Использует SECRET_KEY из env
│   │                                          # - Срок действия токена: 72 часа
│   │
│   ├── 📁 database/                           # 💾 Работа с базой данных
│   │   ├── db.go                              # Подключение к БД и миграции
│   │   │                                      # - InitDB() -> *gorm.DB
│   │   │                                      # - Подключение к PostgreSQL (Neon)
│   │   │                                      # - AutoMigrate всех моделей
│   │   │                                      # - Настройка логирования SQL
│   │   │
│   │   ├── ingredient_repository.go           # Репозиторий для работы с ингредиентами
│   │   │                                      # - CRUD операции
│   │   │                                      # - Поиск по имени, категории
│   │   │                                      # - Проверка наличия на складе
│   │   │
│   │   └── user_repository.go                 # Репозиторий для работы с пользователями
│   │                                          # - FindByEmail(email) -> User
│   │                                          # - CreateUser(user) -> error
│   │                                          # - UpdateUser(user) -> error
│   │
│   ├── 📁 handlers/                           # 🎯 HTTP обработчики (контроллеры)
│   │   │
│   │   ├── admin.go                           # Админ панель
│   │   │                                      # - GetAllUsers() - список пользователей
│   │   │                                      # - UpdateUser() - обновить пользователя
│   │   │                                      # - DeleteUser() - удалить пользователя
│   │   │                                      # Требует: Admin JWT токен
│   │   │
│   │   ├── auth.go                            # Аутентификация
│   │   │                                      # - Register(POST) - регистрация нового пользователя
│   │   │                                      # - Login(POST) - вход в систему
│   │   │                                      # - VerifyToken(POST) - проверка валидности токена
│   │   │                                      # Возвращает: JWT токен + данные пользователя
│   │   │
│   │   ├── health.go                          # Health checks
│   │   │                                      # - HealthHandler() - GET /health
│   │   │                                      # - APIHealthHandler() - GET /api/health
│   │   │                                      # Response: {status: "ok", data: {service, database}}
│   │   │
│   │   ├── hint.go                            # 🤖 Endpoint для Elixir бота
│   │   │                                      # - HintHandler(POST) - /api/hint
│   │   │                                      # - Поиск продуктов по запросу (question)
│   │   │                                      # - Поиск: LOWER(name) LIKE OR LOWER(category) LIKE
│   │   │                                      # - Limit: 5 продуктов
│   │   │                                      # - Response: {status, data: {hint, suggested_products}}
│   │   │                                      # Требует: X-API-Key header
│   │   │
│   │   ├── ingredients.go                     # Управление ингредиентами
│   │   │                                      # - GetIngredients(GET) - все ингредиенты
│   │   │                                      # - CreateIngredient(POST) - создать
│   │   │                                      # - UpdateIngredient(PUT) - обновить
│   │   │                                      # - DeleteIngredient(DELETE) - удалить
│   │   │                                      # - GetIngredientMovements(GET) - история склада
│   │   │                                      # Требует: Admin JWT токен
│   │   │
│   │   ├── orders.go                          # Заказы
│   │   │                                      # - CreateOrder(POST) - создать заказ
│   │   │                                      # - GetUserOrders(GET) - заказы пользователя
│   │   │                                      # - GetAllOrders(GET) - все заказы (admin)
│   │   │                                      # - GetRecentOrders(GET) - последние заказы (admin)
│   │   │                                      # - UpdateOrderStatus(PUT) - изменить статус (admin)
│   │   │                                      # - Отправка WebSocket уведомлений
│   │   │
│   │   ├── products.go                        # Продукты
│   │   │                                      # PUBLIC:
│   │   │                                      # - GetProducts(GET) - видимые продукты
│   │   │                                      # - GetProductByID(GET) - продукт по ID
│   │   │                                      # ADMIN:
│   │   │                                      # - GetAllProducts(GET) - все (+ скрытые)
│   │   │                                      # - CreateProduct(POST) - создать
│   │   │                                      # - UpdateProduct(PUT) - обновить
│   │   │                                      # - DeleteProduct(DELETE) - удалить
│   │   │
│   │   ├── semi_finished.go                   # Полуфабрикаты
│   │   │                                      # - GetSemiFinished(GET) - все
│   │   │                                      # - CreateSemiFinished(POST) - создать
│   │   │                                      # - GetSemiFinishedByID(GET) - по ID
│   │   │                                      # - UpdateSemiFinished(PUT) - обновить
│   │   │                                      # - DeleteSemiFinished(DELETE) - удалить
│   │   │                                      # Требует: Admin JWT токен
│   │   │
│   │   ├── stats.go                           # Статистика
│   │   │                                      # - GetStats(GET) - /api/admin/stats
│   │   │                                      # Возвращает:
│   │   │                                      # - total_users, total_orders, total_products
│   │   │                                      # - total_ingredients, total_revenue
│   │   │                                      # - pending_orders, completed_orders
│   │   │                                      # Требует: Admin JWT токен
│   │   │
│   │   └── websocket.go                       # WebSocket для real-time
│   │                                          # - Hub структура (clients, broadcast, register)
│   │                                          # - NewHub() - создать хаб
│   │                                          # - Run() - запустить горутину хаба
│   │                                          # - HandleWebSocket() - подключение клиента
│   │                                          # - BroadcastNewOrder() - отправка уведомлений
│   │                                          # Endpoint: WS /api/admin/ws?token={jwt}
│   │
│   ├── 📁 middleware/                         # 🛡️ Промежуточные обработчики
│   │   │
│   │   ├── apikey.go                          # API Key middleware
│   │   │                                      # - APIKeyMiddleware() - проверка X-API-Key
│   │   │                                      # - Использует ELIXIR_API_KEY env (default: "supersecret")
│   │   │                                      # - Пропускает OPTIONS для CORS preflight
│   │   │                                      # - 401 если ключ невалидный
│   │   │                                      # Используется: /api/hint endpoint
│   │   │
│   │   └── auth.go                            # JWT middleware
│   │                                          # - AuthMiddleware() - проверка JWT токена
│   │                                          # - AdminMiddleware() - проверка admin роли
│   │                                          # - Извлекает токен из Authorization header
│   │                                          # - Валидирует через auth.ValidateToken()
│   │                                          # - Добавляет userID в context
│   │
│   └── 📁 models/                             # 📊 Модели данных (GORM)
│       │
│       ├── ingredient.go                      # Модель Ingredient
│       │                                      # Поля:
│       │                                      # - ID (uuid)
│       │                                      # - Name, Unit (кг, л, шт)
│       │                                      # - Quantity, MinQuantity (float64)
│       │                                      # - PricePerUnit (float64)
│       │                                      # - Supplier (поставщик)
│       │                                      # - CreatedAt
│       │
│       ├── order.go                           # Модель Order и OrderItem
│       │                                      # Order:
│       │                                      # - ID (uuid), UserID
│       │                                      # - Name, Status (pending/confirmed/etc)
│       │                                      # - Total (float64)
│       │                                      # - Address, Phone, Comment
│       │                                      # - Items []OrderItem (has many)
│       │                                      # - CreatedAt, UpdatedAt
│       │                                      #
│       │                                      # OrderItem:
│       │                                      # - ID, OrderID, ProductID
│       │                                      # - Quantity (int)
│       │                                      # - Price (float64)
│       │
│       ├── product.go                         # Модель Product
│       │                                      # Поля:
│       │                                      # - ID (uuid)
│       │                                      # - Name, Description
│       │                                      # - Price (float64)
│       │                                      # - ImageURL (string)
│       │                                      # - Weight (int, граммы)
│       │                                      # - Category (Суши, Роллы, Супы, etc)
│       │                                      # - IsVisible (bool)
│       │                                      # - Ingredients []ProductIngredient
│       │                                      # - SemiFinished []ProductSemiFinished
│       │                                      # - CreatedAt
│       │
│       ├── semi_finished.go                   # Модель SemiFinished (полуфабрикаты)
│       │                                      # Поля:
│       │                                      # - ID (uuid)
│       │                                      # - Name, Description
│       │                                      # - OutputQuantity, OutputUnit
│       │                                      # - CostPerUnit, TotalCost (float64)
│       │                                      # - Category, IsVisible, IsArchived
│       │                                      # - Ingredients []SemiFinishedIngredient
│       │                                      # - CreatedAt, UpdatedAt, DeletedAt
│       │
│       └── user.go                            # Модель User
│                                              # Поля:
│                                              # - ID (uuid)
│                                              # - Email (unique)
│                                              # - Name
│                                              # - Password (hashed bcrypt)
│                                              # - Role (user/admin)
│                                              # - CreatedAt
│                                              # Методы:
│                                              # - HashPassword() - хеширование пароля
│                                              # - CheckPassword() - проверка пароля
│
├── 📁 migrations/                             # 🔄 SQL миграции
│   ├── 003_add_semi_finished_fields.sql       # Добавление полей в semi_finished
│   ├── add_price_per_unit.sql                 # Добавление price_per_unit
│   ├── create_semi_finished_ingredients.sql   # Создание связи полуфабрикаты-ингредиенты
│   └── create_semi_finished_tables.sql        # Создание таблиц полуфабрикатов
│
├── 📁 pkg/                                    # 📦 Публичные пакеты (можно переиспользовать)
│   └── 📁 utils/
│       └── response.go                        # Утилиты для HTTP ответов
│                                              # - JSONResponse() - стандартный JSON ответ
│                                              # - ErrorResponse() - JSON с ошибкой
│                                              # - SuccessResponse() - успешный ответ
│
├── 📄 .gitignore                              # Git ignore файл
│                                              # Исключает: bin/, *.db, .env, node_modules
│
├── 📄 add_semifinished_columns.sql            # SQL скрипт для добавления колонок
│
├── 📄 API_DOCUMENTATION.md                    # 📚 ПОЛНАЯ ДОКУМЕНТАЦИЯ API
│                                              # - 30 endpoints с примерами
│                                              # - cURL, JavaScript, Axios, React, Vue.js
│                                              # - Описание всех request/response
│                                              # - Коды ошибок
│                                              # - CORS настройки
│                                              # - Quick Start гайды
│
├── 📄 check_db.sql                            # SQL скрипт для проверки БД
│
├── 📄 database.db                             # SQLite БД (для локальной разработки)
│                                              # ⚠️ В production используется PostgreSQL
│
├── 📄 DEPLOY.md                               # 🚀 Инструкции по деплою
│                                              # - Настройка окружения
│                                              # - Environment variables
│                                              # - Docker команды
│                                              # - Production tips
│
├── 📄 Dockerfile                              # 🐳 Docker конфигурация
│                                              # - Multi-stage build
│                                              # - Go 1.x base image
│                                              # - Expose :8080
│                                              # - Оптимизация размера образа
│
├── 📄 ELIXIR_INTEGRATION.md                   # 🤖 Документация интеграции с Elixir
│                                              # - Настройка Elixir/Phoenix проекта
│                                              # - Примеры использования /api/hint
│                                              # - HTTPoison код
│                                              # - API key конфигурация
│                                              # - Тестирование
│
├── 📄 go.mod                                  # Go модули (зависимости)
│                                              # Основные:
│                                              # - github.com/gorilla/mux (роутинг)
│                                              # - gorm.io/gorm (ORM)
│                                              # - gorm.io/driver/postgres (PostgreSQL)
│                                              # - github.com/golang-jwt/jwt/v5 (JWT)
│                                              # - github.com/google/uuid (UUID)
│                                              # - github.com/rs/cors (CORS)
│                                              # - golang.org/x/crypto (bcrypt)
│                                              # - github.com/gorilla/websocket (WebSocket)
│
├── 📄 go.sum                                  # Checksums зависимостей
│                                              # Автоматически управляется Go
│
├── 📄 PROJECT_STRUCTURE.md                    # 📂 ЭТОТ ФАЙЛ
│                                              # Полное описание структуры проекта
│
└── 📄 server.log                              # Логи сервера
                                               # - Запросы HTTP
                                               # - SQL запросы
                                               # - Ошибки
                                               # - Время выполнения

```

---

## 🗂️ Группировка по функциональности

### 🔐 Аутентификация & Авторизация
- `internal/auth/jwt.go` - генерация и валидация JWT
- `internal/handlers/auth.go` - register, login, verify
- `internal/middleware/auth.go` - JWT middleware
- `internal/middleware/apikey.go` - API key для Elixir
- `internal/models/user.go` - модель пользователя

### 💾 База данных
- `internal/database/db.go` - подключение и миграции
- `internal/database/*_repository.go` - репозитории
- `internal/models/*.go` - все модели GORM
- `migrations/*.sql` - SQL миграции

### 🎯 API Handlers (Endpoints)
- `internal/handlers/health.go` - health checks
- `internal/handlers/products.go` - продукты (public + admin)
- `internal/handlers/orders.go` - заказы
- `internal/handlers/ingredients.go` - ингредиенты (admin)
- `internal/handlers/semi_finished.go` - полуфабрикаты (admin)
- `internal/handlers/admin.go` - управление пользователями
- `internal/handlers/stats.go` - статистика
- `internal/handlers/hint.go` - рекомендации для бота
- `internal/handlers/websocket.go` - WebSocket

### 📡 Real-time & Интеграции
- `internal/handlers/websocket.go` - WebSocket Hub
- `internal/handlers/hint.go` - Elixir bot endpoint
- `ELIXIR_INTEGRATION.md` - документация интеграции

### 📝 Документация
- `API_DOCUMENTATION.md` - полная API документация
- `ELIXIR_INTEGRATION.md` - интеграция с Elixir
- `DEPLOY.md` - деплой инструкции
- `PROJECT_STRUCTURE.md` - структура проекта

### 🚀 Запуск & Сборка
- `cmd/server/main.go` - главная точка входа
- `cmd/migrate/main.go` - миграции
- `Dockerfile` - Docker образ
- `go.mod` - зависимости

---

## 📊 Статистика проекта

- **Всего endpoints:** 30
- **Публичных:** 9
- **Защищённых (JWT):** 3
- **Админских:** 18
- **WebSocket:** 1

- **Моделей данных:** 5 (User, Product, Order, Ingredient, SemiFinished)
- **Handlers файлов:** 10
- **Middleware:** 2
- **Репозиториев:** 2

---

## 🔧 Технологический стек

```
Go 1.x
├── Gorilla Mux (HTTP роутинг)
├── GORM (ORM для PostgreSQL)
├── JWT (аутентификация)
├── WebSocket (real-time)
├── CORS (rs/cors)
├── Bcrypt (хеширование паролей)
└── UUID (уникальные ID)
```

---

## 🌐 База данных

### PostgreSQL (Production)
- **Хостинг:** Neon Cloud
- **База:** neondb
- **Подключение:** через DATABASE_URL env variable

### Таблицы:
1. **User** - пользователи системы
2. **Ingredient** - ингредиенты на складе
3. **semi_finished** - полуфабрикаты
4. **semi_finished_ingredients** - связь полуфабрикаты-ингредиенты
5. **Product** - продукты меню
6. **product_ingredients** - связь продукты-ингредиенты
7. **product_semi_finished** - связь продукты-полуфабрикаты
8. **Order** - заказы клиентов
9. **OrderItem** - позиции в заказе

---

## 🚀 Быстрый старт

### Запуск сервера
```bash
# Из корня проекта
go run cmd/server/main.go

# Или скомпилировать
go build -o bin/server cmd/server/main.go
./bin/server
```

### Миграции
```bash
go run cmd/migrate/main.go
```

### Docker
```bash
docker build -t menu-fodi-backend .
docker run -p 8080:8080 menu-fodi-backend
```

---

## 🔐 Environment Variables

```env
DATABASE_URL=postgresql://user:password@host/neondb
SECRET_KEY=your-secret-key-for-jwt
ELIXIR_API_KEY=supersecret
PORT=8080
```

---

## 📞 API Endpoints Summary

### 🟢 Public (9)
- Health checks (2)
- Auth (3): register, login, verify
- Products (2): list, get by ID
- Orders (1): create
- Hint (1): Elixir bot recommendations

### 🔒 Protected (3)
- User profile: get, update
- User orders: list

### 🔴 Admin (18)
- Users management (3)
- Orders management (3)
- Ingredients CRUD (5)
- Semi-finished CRUD (5)
- Products CRUD (5)
- Stats (1)
- WebSocket (1)

---

**Версия:** 1.0  
**Дата создания:** 15 октября 2025 г.  
**Автор:** Menu Fodi Backend Team
