# Business Investment Module - API Documentation

## 📋 Обзор

Модуль инвестиций в бизнесы позволяет пользователям:
- Создавать бизнесы за $19 (получают 1 токен)
- Покупать токены существующих бизнесов (инвестировать)
- Отслеживать свои инвестиции и долю владения
- Просматривать историю транзакций

## 🔗 Эндпоинты (13 новых)

### Публичные эндпоинты (не требуют авторизации)

#### 1. GET `/api/businesses`
Получить список всех активных бизнесов

**Response:**
```json
{
  "status": "ok",
  "data": [
    {
      "id": "uuid",
      "owner_id": "user_id",
      "name": "Coffee Shop",
      "description": "Best coffee in town",
      "category": "Food & Beverage",
      "city": "New York",
      "is_active": true,
      "created_at": "2024-10-15T11:00:00Z",
      "owner": {
        "id": "user_id",
        "name": "John Doe"
      }
    }
  ]
}
```

#### 2. GET `/api/businesses/{id}`
Получить детальную информацию о бизнесе

**Response:**
```json
{
  "status": "ok",
  "data": {
    "business": {
      "id": "uuid",
      "name": "Coffee Shop",
      "tokens": [...],
      "subscriptions": [...]
    },
    "totalInvested": 1000.50,
    "totalTokensOwned": 52,
    "investorsCount": 5
  }
}
```

#### 3. GET `/api/businesses/{id}/holders`
Получить список держателей токенов (инвесторов)

**Response:**
```json
{
  "status": "ok",
  "data": {
    "businessId": "uuid",
    "businessName": "Coffee Shop",
    "totalSupply": 100,
    "holdersCount": 5,
    "holders": [
      {
        "userId": "user_id",
        "userName": "John Doe",
        "tokensOwned": 52,
        "invested": 988.00,
        "sharePercent": 52.0,
        "subscribedAt": "2024-10-15T11:00:00Z"
      }
    ]
  }
}
```

#### 4. GET `/api/businesses/{id}/transactions`
Получить историю транзакций бизнеса

**Response:**
```json
{
  "status": "ok",
  "data": {
    "transactions": [...],
    "stats": {
      "totalVolume": 5000.00,
      "totalCount": 25,
      "buyCount": 20,
      "sellCount": 3,
      "transferCount": 2,
      "burnCount": 0
    }
  }
}
```

---

### Защищённые эндпоинты (требуют JWT токен в заголовке `Authorization: Bearer <token>`)

#### 5. POST `/api/businesses`
Создать новый бизнес (стоимость: $19, получаете 1 токен)

**Request:**
```json
{
  "name": "My Coffee Shop",
  "description": "Best coffee in town",
  "category": "Food & Beverage",
  "city": "New York"
}
```

**Response:**
```json
{
  "status": "ok",
  "message": "Business created successfully",
  "data": {
    "businessId": "uuid",
    "tokenId": "uuid",
    "symbol": "MYCOTKN"
  }
}
```

**Что происходит:**
1. Создаётся бизнес с вами как владельцем
2. Создаётся токен с символом (первые 3 буквы названия + TKN)
3. Вам выдаётся 1 токен (initial supply)
4. Создаётся подписка (вы владеете 100%)
5. Записывается транзакция "buy"

#### 6. PUT `/api/businesses/{id}`
Обновить бизнес (только владелец)

**Request:**
```json
{
  "name": "Updated Name",
  "description": "New description",
  "category": "New Category",
  "city": "Los Angeles"
}
```

#### 7. DELETE `/api/businesses/{id}`
Деактивировать бизнес (только владелец, не удаляет, а устанавливает `is_active = false`)

**Response:**
```json
{
  "status": "ok",
  "message": "Business deactivated successfully"
}
```

#### 8. POST `/api/businesses/{id}/subscribe`
Купить токены бизнеса (инвестировать)

**Request:**
```json
{
  "tokens": 10,
  "amount": 190.00
}
```

**Response:**
```json
{
  "status": "ok",
  "message": "Successfully subscribed to business",
  "data": {
    "subscriptionId": "uuid",
    "tokensOwned": 15,
    "totalInvested": 285.00,
    "sharePercent": 15.5
  }
}
```

**Что происходит:**
1. Проверяется, что amount ≥ tokens × price
2. Увеличивается total_supply токена
3. Создаётся или обновляется подписка пользователя
4. Записывается транзакция "buy"

#### 9. GET `/api/subscriptions`
Получить все инвестиции текущего пользователя

**Response:**
```json
{
  "status": "ok",
  "data": {
    "totalInvested": 1500.00,
    "totalTokens": 78,
    "subscriptionsCount": 3,
    "subscriptions": [
      {
        "subscriptionId": "uuid",
        "businessId": "uuid",
        "businessName": "Coffee Shop",
        "category": "Food & Beverage",
        "tokensOwned": 52,
        "invested": 988.00,
        "sharePercent": 52.0,
        "subscribedAt": "2024-10-15T11:00:00Z"
      }
    ]
  }
}
```

#### 10. GET `/api/transactions`
Получить все транзакции текущего пользователя

**Response:**
```json
{
  "status": "ok",
  "data": {
    "transactions": [...],
    "totalInvested": 1500.00,
    "totalReceived": 0.00,
    "netInvestment": 1500.00,
    "buyCount": 15,
    "sellCount": 0,
    "transactionsCount": 15
  }
}
```

---

## 🗂️ Структура данных

### Business (Бизнес)
```go
{
  "id": "uuid",
  "owner_id": "text",       // FK на User
  "name": "text",
  "description": "text",
  "category": "text",       // Категория бизнеса
  "city": "text",           // Город
  "is_active": boolean,     // Активен ли бизнес
  "created_at": "timestamp",
  "updated_at": "timestamp"
}
```

### BusinessToken (Токен)
```go
{
  "id": "uuid",
  "business_id": "uuid",    // FK на Business
  "symbol": "text",         // MYCOTKN
  "total_supply": int64,    // Общее количество токенов
  "price": decimal(10,2),   // Цена одного токена ($19.00 по умолчанию)
  "created_at": "timestamp"
}
```

**Методы:**
- `GetMarketCap()` → total_supply × price

### BusinessSubscription (Подписка/Инвестиция)
```go
{
  "id": "uuid",
  "user_id": "text",        // FK на User
  "business_id": "uuid",    // FK на Business
  "tokens_owned": int64,    // Количество токенов
  "invested": decimal(10,2),// Сумма инвестиций
  "created_at": "timestamp"
}
```

**Методы:**
- `GetSharePercentage(totalSupply)` → (tokens_owned / totalSupply) × 100

**Constraints:**
- UNIQUE(user_id, business_id) - один пользователь может иметь только одну подписку на бизнес

### Transaction (Транзакция)
```go
{
  "id": "uuid",
  "business_id": "uuid",    // FK на Business
  "from_user": "text",      // Отправитель (пустой для "buy")
  "to_user": "text",        // Получатель
  "tokens": int64,          // Количество токенов
  "amount": decimal(10,2),  // Сумма
  "tx_type": "text",        // buy, sell, transfer, burn
  "created_at": "timestamp"
}
```

**Типы транзакций:**
- `buy` - Покупка токенов (from_user пустой, to_user = покупатель)
- `sell` - Продажа токенов (from_user = продавец, to_user пустой)
- `transfer` - Перевод между пользователями
- `burn` - Сжигание токенов (уменьшение supply)

---

## 📊 Примеры использования

### Создание бизнеса
```bash
curl -X POST http://localhost:8080/api/businesses \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "My Coffee Shop",
    "description": "Best coffee in NYC",
    "category": "Food & Beverage",
    "city": "New York"
  }'
```

### Инвестирование в бизнес
```bash
curl -X POST http://localhost:8080/api/businesses/{business_id}/subscribe \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "tokens": 10,
    "amount": 190.00
  }'
```

### Просмотр своих инвестиций
```bash
curl -X GET http://localhost:8080/api/subscriptions \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

---

## 🎯 Бизнес-логика

### Создание бизнеса
1. Пользователь платит $19
2. Создаётся бизнес с owner_id = user_id
3. Создаётся токен (symbol = первые 3 буквы + "TKN", price = $19, total_supply = 1)
4. Создаётся подписка (tokens_owned = 1, invested = $19)
5. Записывается транзакция "buy"
6. **Результат:** Пользователь владеет 100% бизнеса

### Покупка токенов
1. Пользователь указывает количество токенов и сумму
2. Проверка: amount ≥ tokens × price
3. Увеличивается total_supply токена
4. Создаётся или обновляется подписка пользователя
5. Записывается транзакция "buy"
6. **Результат:** Доля владения пересчитывается автоматически

### Расчёт доли владения
```
sharePercent = (tokens_owned / total_supply) × 100
```

Пример:
- Total supply = 100 токенов
- User A owns = 52 токена
- Share = (52 / 100) × 100 = 52%

---

## 🔐 Права доступа

| Эндпоинт | Публичный | Требует JWT | Только Owner |
|----------|-----------|-------------|--------------|
| GET /businesses | ✅ | ❌ | ❌ |
| GET /businesses/{id} | ✅ | ❌ | ❌ |
| GET /businesses/{id}/holders | ✅ | ❌ | ❌ |
| GET /businesses/{id}/transactions | ✅ | ❌ | ❌ |
| POST /businesses | ❌ | ✅ | ❌ |
| PUT /businesses/{id} | ❌ | ✅ | ✅ |
| DELETE /businesses/{id} | ❌ | ✅ | ✅ |
| POST /businesses/{id}/subscribe | ❌ | ✅ | ❌ |
| GET /subscriptions | ❌ | ✅ | ❌ |
| GET /transactions | ❌ | ✅ | ❌ |

---

## 📈 Статистика

### По бизнесу:
- Total invested (сумма всех инвестиций)
- Total tokens owned (сумма токенов у всех инвесторов)
- Investors count (количество инвесторов)
- Market cap (total_supply × price)

### По пользователю:
- Total invested (сколько вложил)
- Total tokens (сколько токенов владеет)
- Subscriptions count (в скольких бизнесах инвестировал)
- Net investment (invested - received)

---

## 🚀 Deployment

Модуль уже интегрирован и работает на порту 8080.

**Важно:** 
- Все 4 таблицы созданы через AutoMigrate
- Индексы будут созданы отдельно через SQL миграцию (010_create_businesses.sql)
- Транзакции используют ACID для целостности данных
- CASCADE удаление для поддержания связей

---

## ✅ Что реализовано

- ✅ 4 модели: Business, BusinessToken, BusinessSubscription, Transaction
- ✅ 13 новых эндпоинтов (4 публичных + 9 защищённых)
- ✅ Полный CRUD для бизнесов
- ✅ Система инвестирования через токены
- ✅ Подсчёт долей владения
- ✅ История транзакций
- ✅ Статистика по бизнесам и пользователям
- ✅ Защита прав доступа (только владелец может редактировать)
- ✅ Транзакционность операций
- ✅ AutoMigrate для создания таблиц

## 📝 Что можно добавить позже

- 🔄 Продажа токенов (sell)
- 🔄 Перевод токенов между пользователями (transfer)
- 🔄 Сжигание токенов (burn)
- 🔄 Изменение цены токена владельцем
- 🔄 Выплата дивидендов инвесторам
- 🔄 Лимит на максимальное количество токенов
- 🔄 KYC/верификация для крупных инвестиций
- 🔄 Уведомления о новых транзакциях (WebSocket)
- 🔄 Графики роста бизнеса
- 🔄 Рейтинг бизнесов
