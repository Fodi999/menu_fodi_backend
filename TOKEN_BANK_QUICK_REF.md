# 🏦 Token Bank Admin Panel - Quick Reference

> Полная документация для управления банком токинов через админ-панель

## 📋 Overview

**Что это?** Система управления токинами пользователей через админ-панель.

**Для чего?** Администраторы могут:
- ✅ Выделять токины пользователям
- ✅ Отзывать токины при нарушении политики
- ✅ Просматривать баланс токинов
- ✅ Анализировать использование токинов в системе

**Где находится?** Вкладка "Банк Токинов" (Token Bank) в админ-панели

---

## 🔐 Authorization

Все эндпоинты требуют:
- ✅ Valid JWT токен в header: `Authorization: Bearer {token}`
- ✅ Роль пользователя должна быть `admin`

**Пример:**
```bash
curl -X GET http://api.example.com/api/admin/token-bank \
  -H "Authorization: Bearer eyJhbGc..." \
  -H "Content-Type: application/json"
```

---

## 🚀 API Endpoints

### 1. Получить все токин-банки (All Users)
```
GET /api/admin/token-bank
```
**Возвращает:** Список всех пользователей с их балансами

```bash
curl -X GET http://localhost:8080/api/admin/token-bank \
  -H "Authorization: Bearer $TOKEN"
```

**Response:**
```json
[
  {
    "id": "bank-id-1",
    "user_id": "user-1",
    "balance": 500,
    "total_allocated": 1000,
    "total_used": 500,
    "user": {
      "email": "user@example.com",
      "name": "John Doe"
    }
  }
]
```

---

### 2. Получить статистику токинов
```
GET /api/admin/token-bank/stats
```
**Возвращает:** Агрегированная статистика по всем токинам

```bash
curl -X GET http://localhost:8080/api/admin/token-bank/stats \
  -H "Authorization: Bearer $TOKEN"
```

**Response:**
```json
{
  "total_tokens_allocated": 10000,
  "total_tokens_used": 3500,
  "total_users_with_tokens": 156,
  "average_balance_per_user": 41.67
}
```

---

### 3. Получить баланс конкретного пользователя
```
GET /api/admin/token-bank/{userID}
```
**URL параметры:**
- `userID` - UUID пользователя

```bash
# Пример с ID пользователя
curl -X GET http://localhost:8080/api/admin/token-bank/7ec8aba4-8195-4be1-a9a8-067c30aae306 \
  -H "Authorization: Bearer $TOKEN"
```

**Response:**
```json
{
  "id": "bank-id-1",
  "user_id": "7ec8aba4-8195-4be1-a9a8-067c30aae306",
  "balance": 500,
  "total_allocated": 1000,
  "total_used": 500,
  "created_at": "2024-11-01T10:00:00Z",
  "updated_at": "2024-11-10T15:30:00Z",
  "user": {
    "id": "7ec8aba4-8195-4be1-a9a8-067c30aae306",
    "email": "john@example.com",
    "name": "John Doe",
    "role": "user"
  }
}
```

---

### 4. Выделить токины пользователю
```
POST /api/admin/token-bank/allocate
```
**Request Body:**
```json
{
  "user_id": "7ec8aba4-8195-4be1-a9a8-067c30aae306",
  "amount": 500,
  "reason": "Monthly allocation"
}
```

```bash
curl -X POST http://localhost:8080/api/admin/token-bank/allocate \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "7ec8aba4-8195-4be1-a9a8-067c30aae306",
    "amount": 500,
    "reason": "Monthly allocation"
  }'
```

**Response:**
```json
{
  "message": "Tokens allocated successfully",
  "user_id": "7ec8aba4-8195-4be1-a9a8-067c30aae306",
  "amount": 500
}
```

---

### 5. Отозвать токины у пользователя
```
POST /api/admin/token-bank/revoke
```
**Request Body:**
```json
{
  "user_id": "7ec8aba4-8195-4be1-a9a8-067c30aae306",
  "amount": 100,
  "reason": "Policy violation"
}
```

```bash
curl -X POST http://localhost:8080/api/admin/token-bank/revoke \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "7ec8aba4-8195-4be1-a9a8-067c30aae306",
    "amount": 100,
    "reason": "Policy violation"
  }'
```

**Response:**
```json
{
  "message": "Tokens revoked successfully",
  "user_id": "7ec8aba4-8195-4be1-a9a8-067c30aae306",
  "amount": 100
}
```

---

### 6. Установить точный баланс
```
PUT /api/admin/token-bank/balance
```
**Request Body:**
```json
{
  "user_id": "7ec8aba4-8195-4be1-a9a8-067c30aae306",
  "balance": 1000
}
```

```bash
curl -X PUT http://localhost:8080/api/admin/token-bank/balance \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "7ec8aba4-8195-4be1-a9a8-067c30aae306",
    "balance": 1000
  }'
```

**Response:**
```json
{
  "message": "Token balance set successfully",
  "user_id": "7ec8aba4-8195-4be1-a9a8-067c30aae306",
  "balance": 1000
}
```

---

## 📊 Data Types & Examples

### TokenBank Object
```typescript
interface TokenBank {
  id: string;                // Unique bank ID
  user_id: string;           // User UUID
  balance: number;           // Current available tokens
  total_allocated: number;   // Total tokens allocated by admin
  total_used: number;        // Total tokens used by user
  created_at: string;        // ISO 8601 creation date
  updated_at: string;        // ISO 8601 last update date
  user?: {
    id: string;
    email: string;
    name: string;
    role: string;
  };
}
```

### TokenBankStats Object
```typescript
interface TokenBankStats {
  total_tokens_allocated: number;  // Sum of all allocations
  total_tokens_used: number;       // Sum of all usage
  total_users_with_tokens: number; // Count of users
  average_balance_per_user: number; // Mean balance
}
```

---

## 🛠️ Use Cases & Examples

### Use Case 1: Monthly Token Allocation

**Scenario:** Allocate 100 tokens to all active users monthly

```bash
#!/bin/bash

# Get all users
USERS=$(curl -s -X GET http://localhost:8080/api/admin/users \
  -H "Authorization: Bearer $TOKEN" | jq -r '.[] | select(.role=="user") | .id')

# Allocate tokens to each user
for user_id in $USERS; do
  curl -X POST http://localhost:8080/api/admin/token-bank/allocate \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "{
      \"user_id\": \"$user_id\",
      \"amount\": 100,
      \"reason\": \"Monthly allocation\"
    }"
  echo "Allocated 100 tokens to $user_id"
done
```

---

### Use Case 2: Check Low Balance Users

**Scenario:** Find users with less than 50 tokens and send notification

```bash
#!/bin/bash

# Get all token banks
curl -s -X GET http://localhost:8080/api/admin/token-bank \
  -H "Authorization: Bearer $TOKEN" | \
  jq '.[] | select(.balance < 50) | {user: .user.email, balance: .balance}'
```

Output:
```json
{
  "user": "john@example.com",
  "balance": 25
}
{
  "user": "jane@example.com",
  "balance": 45
}
```

---

### Use Case 3: Reset User's Balance

**Scenario:** User exceeded their quota, reset balance to default

```bash
curl -X PUT http://localhost:8080/api/admin/token-bank/balance \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "7ec8aba4-8195-4be1-a9a8-067c30aae306",
    "balance": 0
  }'
```

---

### Use Case 4: View System-wide Stats

**Scenario:** Monitor token usage across the platform

```bash
curl -X GET http://localhost:8080/api/admin/token-bank/stats \
  -H "Authorization: Bearer $TOKEN" | jq '.'
```

Perfect for:
- 📊 Dashboard reports
- 📈 Capacity planning
- 💹 Usage trends

---

## ⚠️ Error Handling

### Common Errors

| Error | HTTP Code | Solution |
|-------|-----------|----------|
| Missing token | 401 | Add `Authorization: Bearer {token}` header |
| Not admin | 403 | User role must be `admin` |
| User not found | 404 | Check if user_id is valid UUID |
| Invalid amount | 400 | Amount must be positive integer |
| Insufficient tokens | 400 | Can't revoke more than balance |
| Invalid balance | 400 | Balance must be non-negative |

### Example Error Response
```json
{
  "error": "Failed to revoke tokens: insufficient tokens"
}
```

---

## 📈 Best Practices

### ✅ DO:
- Use environment variables for tokens: `export TOKEN="..."`
- Validate user IDs before API calls
- Set reasonable token amounts (not too high)
- Track reasons for allocations/revocations
- Monitor stats regularly
- Batch allocations via script

### ❌ DON'T:
- Hardcode tokens in scripts
- Share token in logs or version control
- Allocate negative amounts
- Set very high balances (cause inflation)
- Forget to document manual balance changes
- Use allocate/revoke for wrong users

---

## 🔄 Workflow Example

```
1. Admin logs in → gets JWT token
2. Admin opens Token Bank tab
3. Views current stats via GET /api/admin/token-bank/stats
4. Selects user from list via GET /api/admin/token-bank
5. Checks user's balance via GET /api/admin/token-bank/{userID}
6. Makes decision:
   - Allocate? → POST /api/admin/token-bank/allocate
   - Revoke? → POST /api/admin/token-bank/revoke
   - Set exact? → PUT /api/admin/token-bank/balance
7. Confirms action completed
8. Checks updated stats
```

---

## 📝 Database Schema

```sql
CREATE TABLE token_bank (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL UNIQUE REFERENCES "User"(id),
  balance BIGINT NOT NULL DEFAULT 0,           -- Current balance
  total_allocated BIGINT NOT NULL DEFAULT 0,   -- Historical allocated
  total_used BIGINT NOT NULL DEFAULT 0,        -- Historical used
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

---

## 🧪 Testing

Run test script:
```bash
./test_token_bank.sh
```

Manual test:
```bash
# 1. Set token variable
TOKEN="your_admin_token_here"

# 2. Test get all
curl http://localhost:8080/api/admin/token-bank \
  -H "Authorization: Bearer $TOKEN"

# 3. Test allocate
curl -X POST http://localhost:8080/api/admin/token-bank/allocate \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"user_id":"test-id","amount":100}'

# 4. Verify
curl http://localhost:8080/api/admin/token-bank/stats \
  -H "Authorization: Bearer $TOKEN"
```

---

## 📚 Related Documentation

- 📖 [Admin Endpoints - Full Reference](ADMIN_ENDPOINTS_DATA_STRUCTURE.md)
- 🔑 [Admin Authentication](HOW_ADMIN_LOGIN_WORKS.md)
- 👥 [Admin Role Guide](ADMIN_ROLE_GUIDE.md)
- 🏗️ [Admin Panel Architecture](ADMIN_PANEL_GUIDE.md)

---

**Last Updated:** November 11, 2024  
**Version:** 1.0
