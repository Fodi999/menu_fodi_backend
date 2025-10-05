# 🍣 FODI SUSHI - Go Backend API

REST API для интернет-магазина доставки суши, построенный на Go с JWT аутентификацией и PostgreSQL.

![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-316192?style=for-the-badge&logo=postgresql)
![JWT](https://img.shields.io/badge/JWT-000000?style=for-the-badge&logo=jsonwebtokens)

## 🚀 Возможности

- 🔐 **JWT Аутентификация** - Безопасная авторизация с токенами
- 👥 **Управление пользователями** - Регистрация, вход, профили
- 🛡️ **Защищённые роуты** - Middleware для проверки прав доступа
- 🔧 **Админ-панель API** - Управление пользователями и заказами
- 💾 **PostgreSQL** - Интеграция через GORM ORM
- 🌐 **CORS** - Настроенная поддержка CORS
- 📊 **Статистика** - Endpoints для админ-панели

## 📁 Структура проекта

\`\`\`
backend/
├── cmd/
│   └── server/
│       └── main.go              # Точка входа приложения
├── internal/
│   ├── auth/
│   │   └── jwt.go               # JWT генерация и валидация
│   ├── handlers/
│   │   ├── auth.go              # Регистрация, вход, профиль
│   │   └── admin.go             # Админ endpoints
│   ├── middleware/
│   │   └── auth.go              # Auth & Admin middleware
│   ├── models/
│   │   ├── user.go              # User модель
│   │   └── order.go             # Order модель
│   └── database/
│       ├── db.go                # Подключение к PostgreSQL
│       └── user_repository.go   # User репозиторий
├── pkg/
│   └── utils/
│       └── response.go          # JSON response helpers
├── .env                         # Переменные окружения
├── .env.example                 # Пример конфигурации
├── go.mod                       # Go зависимости
└── README.md                    # Документация
\`\`\`

## 🛠️ Технологии

- **Go** 1.21+
- **gorilla/mux** - HTTP роутер
- **GORM** - ORM для PostgreSQL
- **golang-jwt/jwt** - JWT токены
- **bcrypt** - Хеширование паролей
- **godotenv** - Переменные окружения
- **rs/cors** - CORS middleware

## 🚀 Быстрый старт

### Установка зависимостей

\`\`\`bash
go mod tidy
\`\`\`

### Настройка переменных окружения

\`\`\`bash
cp .env.example .env
# Отредактируйте .env файл
\`\`\`

Пример `.env`:

\`\`\`env
# Server
PORT=8080

# JWT Secret
JWT_SECRET=your-super-secret-jwt-key

# PostgreSQL Database
DATABASE_URL=postgresql://user:password@host:5432/dbname?sslmode=require

# CORS
ALLOWED_ORIGINS=http://localhost:3000,https://your-frontend.vercel.app
\`\`\`

### Запуск сервера

\`\`\`bash
go run cmd/server/main.go
\`\`\`

Сервер запустится на `http://localhost:8080`

## 🔌 API Endpoints

### 🔓 Публичные (без авторизации)

| Метод | Endpoint | Описание |
|-------|----------|----------|
| POST | `/api/auth/register` | Регистрация нового пользователя |
| POST | `/api/auth/login` | Вход в систему |

### 🔒 Защищённые (требуют JWT токен)

| Метод | Endpoint | Описание |
|-------|----------|----------|
| GET | `/api/user/profile` | Получение профиля пользователя |
| PUT | `/api/user/profile` | Обновление профиля |

### 🛡️ Админ (требуют роль admin)

| Метод | Endpoint | Описание |
|-------|----------|----------|
| GET | `/api/admin/users` | Список всех пользователей |
| PUT | `/api/admin/users/:id` | Обновление пользователя |
| DELETE | `/api/admin/users/:id` | Удаление пользователя |
| GET | `/api/admin/orders` | Список всех заказов |
| GET | `/api/admin/orders/recent` | Последние заказы |
| GET | `/api/admin/stats` | Статистика для дашборда |

## 📝 Примеры запросов

### Регистрация

\`\`\`bash
curl -X POST http://localhost:8080/api/auth/register \\
  -H "Content-Type: application/json" \\
  -d '{
    "email": "user@example.com",
    "name": "John Doe",
    "password": "password123"
  }'
\`\`\`

**Ответ:**
\`\`\`json
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "user": {
    "id": "uuid",
    "email": "user@example.com",
    "name": "John Doe",
    "role": "user",
    "createdAt": "2025-10-05T21:00:00Z"
  }
}
\`\`\`

### Вход

\`\`\`bash
curl -X POST http://localhost:8080/api/auth/login \\
  -H "Content-Type: application/json" \\
  -d '{
    "email": "user@example.com",
    "password": "password123"
  }'
\`\`\`

### Получение профиля

\`\`\`bash
curl http://localhost:8080/api/user/profile \\
  -H "Authorization: Bearer <your-token>"
\`\`\`

### Получение списка пользователей (админ)

\`\`\`bash
curl http://localhost:8080/api/admin/users \\
  -H "Authorization: Bearer <admin-token>"
\`\`\`

## 🔐 Аутентификация

Используется JWT (JSON Web Token). После успешного входа/регистрации сервер возвращает токен.

Для доступа к защищённым endpoints добавьте заголовок:

\`\`\`
Authorization: Bearer <token>
\`\`\`

## 🗄️ База данных

Проект использует PostgreSQL через GORM ORM. Схема совместима с Prisma (Next.js frontend).

### Модели:

- **User** - Пользователи (id, email, name, password, role, createdAt)
- **Order** - Заказы (TODO)
- **Product** - Продукты (TODO)

## 🔧 Разработка

### Структура кода

- `cmd/server/main.go` - Entry point, роутинг, CORS
- `internal/handlers/` - HTTP handlers
- `internal/middleware/` - Middleware (auth, logging)
- `internal/auth/` - JWT логика
- `internal/database/` - Database layer, repositories
- `internal/models/` - Data models
- `pkg/utils/` - Utilities (JSON responses)

### Добавление нового endpoint

1. Создайте handler в `internal/handlers/`
2. Добавьте роут в `cmd/server/main.go`
3. При необходимости добавьте middleware

## 🚢 Деплой

### Heroku

\`\`\`bash
heroku create your-app-name
heroku addons:create heroku-postgresql:hobby-dev
git push heroku main
\`\`\`

### Railway

\`\`\`bash
railway init
railway add
railway up
\`\`\`

### Docker

\`\`\`dockerfile
FROM golang:1.21-alpine
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o server cmd/server/main.go
CMD ["./server"]
\`\`\`

## 🤝 Интеграция с Frontend

Backend работает с Next.js frontend:
- Frontend: https://github.com/Fodi999/menu-fodifood
- API URL: `http://localhost:8080` (dev) или `https://your-api.com` (prod)

## 📄 Лицензия

© 2025 FODI SUSHI. All rights reserved.

## 👨‍💻 Автор

Dmitrij Fomin

---

Сделано с ❤️ для любителей суши 🍣
