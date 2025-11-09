# Wallet Module

Модуль управления виртуальным кошельком пользователя (токены).

## Структура

```
wallet/
├── dto/              # Data Transfer Objects
│   └── requests.go   # Request/Response DTOs
├── repo/             # Repository (Database layer)
│   └── repository.go # Database operations
├── service/          # Business Logic layer
│   └── service.go    # Business logic
├── transport/        # Transport layer
│   └── http/         # HTTP handlers
│       └── handlers.go
└── module.go         # Module registration
```

## API Endpoints

### Protected Routes (require authentication)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/wallet/balance` | Получить баланс кошелька |
| POST | `/api/wallet/purchase` | Купить токены |
| POST | `/api/wallet/spend` | Потратить токены |
| GET | `/api/wallet/transactions?limit=50` | История транзакций |
| GET | `/api/user/{userId}/wallet` | Информация о кошельке пользователя |
| POST | `/api/user/{userId}/wallet/grant-welcome` | Выдать приветственные токены |

## Request/Response Examples

### Purchase Tokens
```bash
POST /api/wallet/purchase
Authorization: Bearer {token}
Content-Type: application/json

{
  "amount": 100,
  "paymentMethod": "card",
  "description": "Monthly token purchase"
}
```

Response:
```json
{
  "success": true,
  "message": "Tokens purchased successfully",
  "transaction": {
    "id": "uuid",
    "userId": "uuid",
    "amount": 100,
    "type": "purchase",
    "status": "completed",
    "newBalance": 200,
    "previousBalance": 100,
    "createdAt": "2025-11-09T10:30:00Z"
  }
}
```

### Spend Tokens
```bash
POST /api/wallet/spend
Authorization: Bearer {token}
Content-Type: application/json

{
  "amount": 50,
  "description": "Purchase premium recipe",
  "relatedId": "recipe-uuid"
}
```

Response:
```json
{
  "success": true,
  "message": "Tokens spent successfully",
  "transaction": {
    "id": "uuid",
    "userId": "uuid",
    "amount": 50,
    "type": "spend",
    "status": "completed",
    "newBalance": 150,
    "previousBalance": 200,
    "createdAt": "2025-11-09T10:35:00Z"
  }
}
```

### Get Balance
```bash
GET /api/wallet/balance
Authorization: Bearer {token}
```

Response:
```json
{
  "success": true,
  "data": {
    "userId": "uuid",
    "balance": 150
  }
}
```

### Transaction History
```bash
GET /api/wallet/transactions?limit=10
Authorization: Bearer {token}
```

Response:
```json
{
  "success": true,
  "transactions": [
    {
      "id": "uuid",
      "userId": "uuid",
      "amount": 50,
      "type": "spend",
      "status": "completed",
      "createdAt": "2025-11-09T10:35:00Z"
    },
    {
      "id": "uuid",
      "userId": "uuid",
      "amount": 100,
      "type": "purchase",
      "status": "completed",
      "createdAt": "2025-11-09T10:30:00Z"
    }
  ],
  "count": 2
}
```

## Business Logic

### Service Layer (`service/service.go`)

Основные методы:
- `GetBalance(userID)` - получить баланс
- `PurchaseTokens(userID, amount, method)` - покупка токенов
- `SpendTokens(userID, amount, description)` - списание токенов
- `GetTransactionHistory(userID, limit)` - история транзакций
- `GrantWelcomeTokens(userID, amount)` - выдача приветственных токенов

Бизнес-правила:
- ✅ Проверка достаточности средств перед списанием
- ✅ Транзакционность операций (rollback при ошибке)
- ✅ Валидация суммы (amount > 0)
- ✅ Создание записи транзакции для каждой операции
- ✅ Обновление баланса в UserProfile

### Repository Layer (`repo/repository.go`)

Операции с БД:
- `GetBalance(userID)` - получить баланс из UserProfile
- `UpdateBalance(userID, newBalance)` - обновить баланс
- `CreateTransaction(transaction)` - создать запись транзакции
- `GetTransactions(userID, limit)` - получить историю

Таблицы:
- `UserProfile.wallet_balance` - текущий баланс пользователя
- `WalletTransaction` - история всех транзакций

## Error Handling

Типы ошибок:
- `ErrInsufficientBalance` - недостаточно средств
- `ErrInvalidAmount` - неверная сумма (<=0)
- `ErrUserNotFound` - пользователь не найден

HTTP статусы:
- `200 OK` - успешная операция
- `400 Bad Request` - неверные данные запроса
- `401 Unauthorized` - отсутствует авторизация
- `500 Internal Server Error` - ошибка сервера

## Dependencies

- `github.com/google/uuid` - UUID generation
- `gorm.io/gorm` - ORM для работы с БД
- `internal/database` - подключение к БД
- `internal/middleware` - AuthMiddleware
- `internal/platform/httpx` - HTTP response helpers
- `internal/platform/logger` - Structured logging

## Testing

```bash
# Unit tests для service layer
go test ./internal/modules/wallet/service/...

# Integration tests для HTTP handlers
go test ./internal/modules/wallet/transport/http/...

# Run all tests
go test ./internal/modules/wallet/...
```

## Usage in Application

```go
// В internal/app/routes.go
walletModule := wallet.NewModule()
walletModule.RegisterRoutes(r.Route("/api", ...))
```

## Future Improvements

- [ ] Добавить поддержку валют (USD, EUR, UAH)
- [ ] Реализовать систему промокодов
- [ ] Добавить лимиты на операции
- [ ] Реализовать возвраты (refunds)
- [ ] Добавить webhooks для payment providers
- [ ] Интеграция с Stripe/PayPal
- [ ] Аналитика расходов пользователя
- [ ] Подписки и автоплатежи
