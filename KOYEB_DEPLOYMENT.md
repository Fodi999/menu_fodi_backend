# 🚀 Koyeb Deployment Guide

Полное руководство по развертыванию бэкэнда на Koyeb.

---

## 📋 Содержание

1. [Koyeb Setup](#koyeb-setup)
2. [Environment Variables](#environment-variables)
3. [Build & Deploy](#build--deploy)
4. [Troubleshooting](#troubleshooting)
5. [API URLs](#api-urls)

---

## Koyeb Setup

### 1️⃣ Создать приложение на Koyeb

```bash
# 1. Зайти на https://app.koyeb.com
# 2. Нажать "Create App"
# 3. Выбрать "Docker"
# 4. Указать GitHub repo
# 5. Выбрать branch: main
```

### 2️⃣ Dockerfile

Убедитесь, что есть `Dockerfile` в корне проекта:

```dockerfile
FROM golang:1.24.3-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server cmd/server/main.go

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/server .

# Expose port
EXPOSE 8000

# Run
CMD ["./server"]
```

### 3️⃣ .dockerignore

```
.git
.gitignore
README.md
*.md
.env
.env.local
bin/
migrations/
tests/
```

---

## Environment Variables

### На Koyeb установите эти переменные:

| Переменная | Значение | Пример |
|-----------|----------|--------|
| `HTTP_PORT` | `8000` | (default для Koyeb) |
| `APP_ENV` | `prod` | или `staging` |
| `DATABASE_URL` | PostgreSQL URL | `postgresql://user:pass@host/db` |
| `JWT_SECRET` | Секретный ключ | `your-super-secret-key` |
| `GROQ_API_KEY` | API ключ | (если нужен AI) |

### Как установить:

```bash
# 1. На странице приложения Koyeb
# Settings → Environment Variables
# 2. Добавить переменные
# 3. Redeploy приложение
```

### ⚠️ Важные переменные

```bash
# DATABASE_URL (Neon PostgreSQL)
DATABASE_URL=postgresql://user:password@ep-xxx.eu-central-1.neon.tech/dbname?sslmode=require

# JWT_SECRET (генерируйте безопасный ключ)
JWT_SECRET=$(openssl rand -base64 32)

# APP_ENV
APP_ENV=prod
```

---

## Build & Deploy

### 1️⃣ Локально тестировать Docker образ

```bash
# Собрать образ
docker build -t menu-fodi:latest .

# Запустить с env переменными
docker run -it \
  -e HTTP_PORT=8000 \
  -e APP_ENV=prod \
  -e DATABASE_URL="postgresql://..." \
  -e JWT_SECRET="your-secret" \
  -p 8000:8000 \
  menu-fodi:latest
```

### 2️⃣ Автоматический deploy на Koyeb

```bash
# Просто push в main branch
git add .
git commit -m "Deploy to Koyeb"
git push origin main

# Koyeb автоматически:
# 1. Скачает код
# 2. Запустит `docker build`
# 3. Задеплоит на https://your-app.koyeb.app
```

### 3️⃣ Ручной deploy (если нужен)

```bash
# Через Koyeb Web UI:
# 1. Settings → Deployments
# 2. Нажать "Trigger Deployment"
```

---

## API URLs

### Локально (dev)

```
Base URL: http://localhost:8080
Endpoints:
- GET    http://localhost:8080/health
- POST   http://localhost:8080/api/auth/login
- GET    http://localhost:8080/api/user/profile
```

### На Koyeb (prod)

```
Base URL: https://your-app.koyeb.app
Endpoints:
- GET    https://your-app.koyeb.app/health
- POST   https://your-app.koyeb.app/api/auth/login
- GET    https://your-app.koyeb.app/api/user/profile
```

⚠️ **ВАЖНО:** На Koyeb используйте **https://** а не http://

---

## Testing API

### 1️⃣ Health Check

```bash
# Локально
curl http://localhost:8080/health

# На Koyeb
curl https://your-app.koyeb.app/health

# Ответ: "ok"
```

### 2️⃣ Login

```bash
# Локально
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123"
  }'

# На Koyeb
curl -X POST https://your-app.koyeb.app/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123"
  }'
```

### 3️⃣ Get Profile

```bash
# Скопируйте токен из login ответа

curl -X GET https://your-app.koyeb.app/api/user/profile \
  -H "Authorization: Bearer YOUR_TOKEN"
```

---

## Troubleshooting

### 404 ошибки

**Проблема:** Фронт получает 404 на /api/user/profile

**Решения:**

1. **Проверить HTTP_PORT**
   ```bash
   # Koyeb использует порт 8000
   # Убедитесь что в config используется PORT из env
   HTTP_PORT=${HTTP_PORT:-8080}
   ```

2. **Проверить DATABASE_URL**
   ```bash
   # На Koyeb должен быть postgresql URL
   # Тестировать локально с тем же URL:
   DATABASE_URL="postgresql://..." go run cmd/server/main.go
   ```

3. **Проверить routes регистрацию**
   ```bash
   # Все модули регистрируются в setupModularRoutes()
   # Убедитесь что userModule.RegisterRoutes вызывается
   ```

4. **CORS проблема**
   ```bash
   # На Koyeb фронт на другом домене
   # CORS должен разрешать все origins:
   AllowedOrigins: []string{"*"}
   ```

### 500 ошибки

**Проблема:** Internal Server Error

**Решения:**

1. **Проверить логи на Koyeb**
   ```bash
   # Settings → Logs
   # Смотреть какая именно ошибка
   ```

2. **Проверить DATABASE_URL**
   ```bash
   # Неверный URL → ошибка подключения к БД
   # Убедитесь что password правильный
   ```

3. **Проверить JWT_SECRET**
   ```bash
   # Пустой JWT_SECRET → ошибка при login
   # Установить в Environment Variables
   ```

### 502 Bad Gateway

**Проблема:** Сервер не запускается

**Решения:**

1. **Проверить Dockerfile**
   ```bash
   docker build -t test .
   docker run test
   # Должно запуститься без ошибок
   ```

2. **Проверить HTTP_PORT**
   ```bash
   # Должен быть 8000 (default) или из env
   # Не используйте hardcoded порты
   ```

3. **Проверить зависимости**
   ```bash
   go mod tidy
   git add go.mod go.sum
   git commit -m "Update dependencies"
   git push
   ```

---

## Как использовать на фронте

### Development (localhost)

```typescript
// src/lib/api.ts
const BASE_URL = process.env.REACT_APP_API_URL || 'http://localhost:8080';

export async function apiFetch(endpoint: string, options: any) {
  const url = `${BASE_URL}${endpoint}`;
  // ...
}
```

### Production (Koyeb)

```bash
# .env.production
REACT_APP_API_URL=https://your-app.koyeb.app
```

### Оба окружения

```bash
# .env
REACT_APP_API_URL=http://localhost:8080

# .env.production
REACT_APP_API_URL=https://your-app.koyeb.app

# Или через переменные окружения при build
npm run build -- --env APP_API_URL=https://your-app.koyeb.app
```

---

## Полезные команды

```bash
# Проверить что Dockerfile работает
docker build -t menu-fodi:test .
docker run --rm menu-fodi:test ./server --version

# Проверить размер образа
docker build -t menu-fodi .
docker images menu-fodi

# Посмотреть логи локально
docker run -it \
  -e HTTP_PORT=8000 \
  -e DATABASE_URL="postgresql://..." \
  -e JWT_SECRET="secret" \
  menu-fodi

# На Koyeb смотреть логи
# Settings → Logs (Web UI)
```

---

## Масштабирование на Koyeb

### Настройки производительности

```
Settings → Scaling
- Min instances: 1
- Max instances: 3
- CPU: 500m
- Memory: 512Mi
```

### Рекомендации

- **Development:** 1 instance, 256Mi memory
- **Staging:** 1-2 instances, 512Mi memory
- **Production:** 2-3 instances, 1Gi memory

---

## CI/CD на Koyeb

Koyeb автоматически:

1. ✅ Слушает push в main branch
2. ✅ Скачивает код
3. ✅ Запускает docker build
4. ✅ Запускает docker run
5. ✅ Заменяет старый контейнер на новый

**Не нужны дополнительные GitHub Actions!**

---

## Мониторинг

### Health Check

Koyeb проверяет здоровье приложения:

```bash
# Koyeb делает периодический запрос
GET /health

# Должен быть статус 200 и ответ "ok"
w.WriteHeader(http.StatusOK)
w.Write([]byte("ok"))
```

Убедитесь что `/health` работает без authentication!

---

## Backup БД

Используйте **Neon Console** для backup:

```bash
# https://console.neon.tech
# Projects → Backups
# Можно восстановить любую версию БД за 1 клик
```

---

## Заключение

**Что нужно сделать:**

1. ✅ Проверить `Dockerfile` в корне проекта
2. ✅ Установить Environment Variables на Koyeb
3. ✅ Убедитесь что `DATABASE_URL` правильный
4. ✅ Push код в main branch
5. ✅ Koyeb автоматически задеплоит
6. ✅ Проверить здоровье: `curl https://your-app.koyeb.app/health`

**Если остались проблемы:**

- Смотреть логи на Koyeb: Settings → Logs
- Тестировать локально с Docker
- Проверить Environment Variables

