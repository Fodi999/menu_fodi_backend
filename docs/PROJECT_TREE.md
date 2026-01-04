# 🌲 Структура проекта Chef Academy Backend

## 📦 Полное дерево проекта

```
backend/
│
├── 📂 cmd/                           # Точки входа (команды)
│   ├── server/main.go               # 🚀 Основной HTTP сервер
│   ├── migrate/main.go              # Миграции БД
│   ├── check_arrived/               # Утилита проверки
│   ├── cleanup_expired/             # Авто-чистка просроченных продуктов
│   └── cleanup_ingredients/         # Чистка дубликатов
│
├── 📂 internal/                      # Внутренняя логика (приватная)
│   │
│   ├── 📂 app/                      # ⚙️ Ядро приложения
│   │   ├── server.go               # Инициализация сервера
│   │   └── routes_modular.go       # Регистрация всех модулей
│   │
│   ├── 📂 database/                 # 💾 Репозитории (Data Access Layer)
│   │   ├── db.go                   # GORM connection pool
│   │   ├── user_repository.go
│   │   ├── ingredient_repository.go
│   │   └── ...                     # Другие репозитории
│   │
│   ├── 📂 middleware/               # 🛡️ HTTP Middleware
│   │   ├── auth.go                 # JWT Auth, AdminMiddleware, SuperAdminMiddleware
│   │   └── apikey.go               # API Key validation
│   │
│   ├── 📂 models/                   # 📦 Модели данных (GORM)
│   │   ├── user.go                 # User, Role enum
│   │   ├── ingredient.go           # Ingredient, Category
│   │   ├── recipe.go               # Recipe, RecipeIngredient
│   │   ├── user_fridge.go          # UserFridgeItem
│   │   ├── business.go             # Business entities
│   │   └── ...                     # 25+ моделей
│   │
│   ├── 📂 modules/                  # 🧩 Бизнес-модули (DDD)
│   │   │
│   │   ├── 📂 auth/                # 🔐 Аутентификация
│   │   │   ├── module.go          # Инициализация
│   │   │   ├── service/           # Бизнес-логика
│   │   │   ├── repo/              # Data access
│   │   │   └── transport/http/    # HTTP handlers
│   │   │
│   │   ├── 📂 admin/               # 👑 Админ-панель (NEW!)
│   │   │   ├── module.go
│   │   │   ├── service/
│   │   │   │   └── service.go     # GetUsersWithFilters, GetAllIngredients
│   │   │   └── transport/http/
│   │   │       └── handlers.go    # GetAllUsers, DeleteUser
│   │   │
│   │   ├── 📂 user/                # 👤 Пользователи
│   │   ├── 📂 fridge/              # 🧊 Холодильник
│   │   ├── 📂 recipes/             # 📖 Рецепты (catalog, match, exclude)
│   │   ├── 📂 ingredients/         # 🥕 Каталог ингредиентов
│   │   ├── 📂 ai/                  # 🤖 AI-адаптация рецептов
│   │   ├── 📂 ai_recommendations/  # 🧠 AI рекомендации
│   │   ├── 📂 budget/              # 💰 Бюджет
│   │   ├── 📂 meal_plan/           # 📅 План питания
│   │   ├── 📂 marketplace/         # 🛒 Marketplace (B2B)
│   │   ├── 📂 business/            # 🏢 Бизнес-аккаунты
│   │   ├── 📂 prepared_dishes/     # 🍽️ Готовые блюда
│   │   ├── 📂 wallet/              # 💳 Кошелёк
│   │   ├── 📂 leaderboard/         # 🏆 Таблица лидеров
│   │   ├── 📂 history/             # 📊 История активности
│   │   ├── 📂 stats/               # 📈 Статистика
│   │   ├── 📂 health/              # ❤️ Health checks
│   │   ├── 📂 contact/             # 📧 Контактная форма
│   │   ├── 📂 hint/                # 💡 Подсказки
│   │   ├── 📂 metrics/             # 📊 Метрики
│   │   ├── 📂 nutrition/           # 🥗 Нутриционный анализ
│   │   └── 📂 websocket/           # 🔄 WebSocket (real-time)
│   │
│   └── 📂 platform/                 # 🔧 Платформенные утилиты
│       ├── config/                 # .env конфигурация
│       ├── logger/                 # zap логирование
│       └── httpx/                  # HTTP helpers
│
├── 📂 pkg/                          # 📦 Переиспользуемые пакеты
│   └── utils/
│       └── response.go             # RespondWithJSON, RespondWithError
│
├── 📂 migrations/                   # 🗄️ SQL миграции (60 файлов)
│   ├── 001_create_users_table.sql
│   ├── 058_add_russian_translations.sql
│   ├── 059_add_user_status_and_activity.sql
│   ├── 060_add_super_admin_role.sql
│   └── ...
│
├── 📂 docs/                         # 📚 Документация (35+ файлов)
│   ├── PROJECT_ARCHITECTURE.md     # Полная архитектура
│   ├── PROJECT_TREE.md             # Этот файл
│   ├── RECIPE_SYSTEM_SUMMARY.md
│   ├── SUPER_ADMIN_IMPLEMENTATION.md
│   └── ...
│
├── 📂 sql/                          # 📊 SQL утилиты
│   ├── check_user_activity.sql
│   └── diagnostic_price_flow.sql
│
├── 📂 certificates/                 # 🎓 Сертификаты
│
├── 📄 .env                          # 🔐 Переменные окружения
├── 📄 go.mod                        # 📦 Go зависимости
├── 📄 go.sum                        # 🔒 Lockfile
├── 📄 Dockerfile                    # 🐳 Docker образ
├── 📄 Makefile                      # ⚙️ Команды сборки
├── 📄 server.log                    # 📝 Логи сервера
│
└── 📜 Скрипты утилиты:
    ├── check_activity.sh           # Проверка активности пользователей
    ├── test_admin_filters.sh       # Тест фильтров админ-панели
    ├── login_as_super_admin.sh     # Логин как super_admin
    └── make_dima_admin.sh          # Назначить роль admin
```

---

## 🎯 Ключевые директории

### 1. **cmd/** - Точки входа
```
cmd/server/main.go  → Основной сервер (порт 8080)
cmd/migrate/        → Применение миграций
cmd/cleanup_*       → Утилиты очистки данных
```

### 2. **internal/app/** - Ядро приложения
```
server.go          → Инициализация, graceful shutdown
routes_modular.go  → Регистрация 28 модулей
```

### 3. **internal/modules/** - Бизнес-логика (28 модулей)
```
Каждый модуль содержит:
├── module.go              # Инициализация и регистрация роутов
├── service/               # Бизнес-логика
│   ├── interface.go      # Интерфейс сервиса
│   └── service.go        # Реализация
├── repo/                  # Data Access Layer (опционально)
└── transport/http/        # HTTP handlers
    └── handlers.go
```

### 4. **internal/models/** - Модели данных (GORM)
```
user.go         → User, Role enum (super_admin, admin, home_chef, etc.)
ingredient.go   → Ingredient, Category (211 records)
recipe.go       → Recipe (70 records)
user_fridge.go  → Холодильник пользователя
business.go     → B2B сущности
```

### 5. **migrations/** - База данных (60 миграций)
```
001-010: Базовые таблицы (users, recipes, businesses)
011-020: Токены, задачи, роли
021-030: Холодильник, ингредиенты, цены
031-040: Каталог рецептов, логи приготовления
041-050: Нормализация, сохранённые рецепты
051-060: Мультиязычность, активность, super_admin
```

---

## 🔗 Взаимосвязи между слоями

```
HTTP Request
    ↓
[Chi Router] → routes_modular.go
    ↓
[Middleware Chain]
    ├── AuthMiddleware       (JWT validation)
    ├── AdminMiddleware      (role check)
    └── SuperAdminMiddleware (super_admin only)
    ↓
[Module Handler] → transport/http/handlers.go
    ↓
[Service Layer] → service/service.go
    ↓
[Repository Layer] → database/*_repository.go
    ↓
[GORM] → PostgreSQL (Neon)
    ↓
HTTP Response
```

---

## 📊 Статистика проекта

```
📁 Модулей:        28 (DDD architecture)
📄 Миграций:       60 SQL файлов
📝 Документов:     35+ в /docs
🗃️ Моделей:        25+ GORM models
👥 Пользователей:  54 (1 super_admin, 3 admin, 49 home_chef, 1 investor)
🥕 Ингредиентов:   211 (6 категорий)
📖 Рецептов:       70
```

---

## 🚀 Быстрый старт

```bash
# Клонировать репозиторий
git clone https://github.com/Fodi999/menu_fodi_backend.git

# Установить зависимости
go mod download

# Настроить .env
cp .env.example .env

# Применить миграции
go run cmd/migrate/main.go

# Запустить сервер
go run cmd/server/main.go
# или
make run

# Сервер доступен на http://localhost:8080
```

---

## 🔧 Основные команды

```bash
# Сборка
make build
go build -o bin/server ./cmd/server

# Запуск
./bin/server

# Миграции
go run cmd/migrate/main.go

# Проверка активности пользователей
./check_activity.sh

# Тестирование админ фильтров
./test_admin_filters.sh

# Логин как super_admin
./login_as_super_admin.sh
```

---

## 🌐 API Endpoints (основные)

```
🔐 Auth:
POST   /api/auth/register
POST   /api/auth/login
GET    /api/auth/me

👤 Users:
GET    /api/users/me
PATCH  /api/users/profile

👑 Admin (super_admin + admin):
GET    /api/admin/users?role=X&status=Y&search=Z
GET    /api/admin/users/stats
GET    /api/admin/ingredients
DELETE /api/admin/users/:id (super_admin only)

🧊 Fridge:
GET    /api/fridge
POST   /api/fridge/items
DELETE /api/fridge/items/:id

📖 Recipes:
GET    /api/recipes/catalog?category=X&difficulty=Y
GET    /api/recipes/match (match с холодильником)
POST   /api/recipes/exclude

🥕 Ingredients:
GET    /api/catalog/ingredients?search=говядина

🤖 AI:
POST   /api/ai/adapt
GET    /api/ai/recommendations

💰 Budget:
GET    /api/budget/weekly

📊 Stats:
GET    /api/stats/overview
```

---

## 📖 Полная документация

Смотрите `docs/PROJECT_ARCHITECTURE.md` для детальной архитектуры.

---

**Последнее обновление:** 4 января 2026 г.
