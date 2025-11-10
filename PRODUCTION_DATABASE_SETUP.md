# 🚀 Production Database Setup Guide

**Production URL**: `https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app`

---

## ⚠️ Проблема: Production БД пуста

После развертывания backend на Koyeb, база данных может быть пустой:
- Нет пользователей
- Нет тестовых данных
- Frontend получает 401 при попытке входа

---

## ✅ Решение: Создать тестовых пользователей

### Метод 1: Быстрая регистрация через API

Создайте пользователя через API:

```bash
# 1. Создать обычного пользователя
curl -X POST https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123",
    "name": "Test User"
  }'

# Ответ должен быть 201 Created
```

### Метод 2: Создать администратора

Если нужна админ панель:

```bash
# 1. Зарегистрировать пользователя
curl -X POST https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.com",
    "password": "admin_password_123",
    "name": "System Admin"
  }'

# 2. Получить ID пользователя из ответа
# (сохранить для следующего шага)

# 3. Обновить роль в БД (если есть доступ)
# (нужно подключиться к БД на Koyeb)
```

---

## 🔐 Вход с созданными учетными данными

После регистрации используйте эти учетные данные для входа:

```bash
curl -X POST https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123"
  }'

# Ответ: 200 OK с token
```

---

## 📝 Batch Script для создания нескольких пользователей

Создайте script `setup_production_db.sh`:

```bash
#!/bin/bash

API_URL="https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app"

echo "🚀 Creating test users in production..."

# Массив с пользователями [email, password, name]
declare -a users=(
    "user@example.com:password123:Test User"
    "admin@example.com:admin123:Admin User"
    "demo@example.com:demo123:Demo User"
    "chef@example.com:chef123:Chef User"
)

# Создать каждого пользователя
for user_data in "${users[@]}"; do
    IFS=':' read -r email password name <<< "$user_data"
    
    echo "📝 Registering: $email"
    
    response=$(curl -s -X POST "$API_URL/api/auth/register" \
      -H "Content-Type: application/json" \
      -d "{
        \"email\": \"$email\",
        \"password\": \"$password\",
        \"name\": \"$name\"
      }")
    
    if echo "$response" | grep -q '"success":true'; then
        echo "✅ $email created successfully"
    else
        echo "❌ Failed to create $email"
        echo "Response: $response"
    fi
done

echo "🎉 Done!"
```

Запустите:

```bash
chmod +x setup_production_db.sh
./setup_production_db.sh
```

---

## 🗄️ Подключение к Production БД (если нужно)

Если у вас есть доступ к PostgreSQL на Koyeb:

```bash
# Получить connection string из Koyeb dashboard

# Подключиться:
psql "postgresql://user:password@host:port/dbname"

# Посмотреть пользователей:
SELECT id, email, role FROM "User";

# Назначить админа:
UPDATE "User" SET role = 'admin' WHERE email = 'admin@example.com';

# Проверить:
SELECT email, role FROM "User" WHERE email = 'admin@example.com';
```

---

## 📋 Чек-лист для Production Setup

- [ ] Backend развернут на Koyeb
- [ ] Database создана и миграции запущены
- [ ] Протестировать регистрацию: `POST /api/auth/register`
- [ ] Создать тестового пользователя
- [ ] Протестировать вход: `POST /api/auth/login`
- [ ] Получить JWT token
- [ ] Протестировать защищенный endpoint с token
- [ ] (Опционально) Создать админа и назначить роль
- [ ] (Опционально) Протестировать admin endpoints

---

## 🔍 Проверка статуса Production

### 1. Проверить что сервер работает

```bash
curl https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/health
```

### 2. Проверить что БД подключена

Попробуйте зарегистрировать пользователя:

```bash
curl -X POST https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@test.com","password":"test123","name":"Test"}'

# 201 Created → БД работает
# 500 Error → проблема с БД
```

### 3. Проверить Logs на Koyeb

В Koyeb dashboard → Logs → смотрите ошибки

---

## 🎯 Быстрый старт

1. **Зарегистрируйте пользователя**:
```bash
curl -X POST https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password123","name":"Test User"}'
```

2. **Войдите**:
```bash
curl -X POST https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password123"}'
```

3. **Сохраните token**:
```bash
TOKEN="eyJhbGciOiJIUzI1NiIs..."
```

4. **Используйте в запросах**:
```bash
curl -H "Authorization: Bearer $TOKEN" \
  https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/admin/stats
```

---

## 📞 Troubleshooting Production

| Проблема | Решение |
|---------|---------|
| 401 Invalid credentials | Зарегистрируйте пользователя через `/api/auth/register` |
| 500 Internal Server Error | Проверьте logs на Koyeb, проверьте подключение к БД |
| 404 Not Found | Проверьте что endpoint существует (see API_ENDPOINTS_FOR_FRONTEND.md) |
| Timeout | Проверьте что Koyeb deployment запущен |
| CORS ошибка | Проверьте CORS конфиг в backend |

---

## 🔗 Полезные ссылки

- **Production URL**: https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app
- **API Endpoints**: см. `API_ENDPOINTS_FOR_FRONTEND.md`
- **Admin Panel**: см. `ADMIN_PANEL_GUIDE.md`
- **Login Guide**: см. `HOW_ADMIN_LOGIN_WORKS.md`
- **Troubleshooting**: см. `LOGIN_401_TROUBLESHOOTING.md`

---

**Главное**: После развертывания нужно создать тестовых пользователей в production БД!
