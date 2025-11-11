# 🏦 Token Bank Setup & Deployment Guide

Полное руководство по развёртыванию и использованию функции Token Bank в админ-панели.

---

## 📋 Table of Contents

1. [Overview](#overview)
2. [Architecture](#architecture)
3. [Database Setup](#database-setup)
4. [API Endpoints](#api-endpoints)
5. [Testing](#testing)
6. [Production Deployment](#production-deployment)
7. [Troubleshooting](#troubleshooting)

---

## Overview

**Token Bank** — это система управления токинами пользователей через админ-панель. Позволяет администраторам:

- ✅ Просмотреть все токин-банки пользователей
- ✅ Выделить токины пользователям
- ✅ Отозвать токины (если нарушены правила)
- ✅ Установить точный баланс
- ✅ Просмотреть статистику по токинам в системе

---

## Architecture

### Database Schema

```sql
CREATE TABLE token_bank (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL UNIQUE REFERENCES "User"(id) ON DELETE CASCADE,
  balance BIGINT NOT NULL DEFAULT 0,                -- Текущий доступный баланс
  total_allocated BIGINT NOT NULL DEFAULT 0,       -- Всего выделено админом
  total_used BIGINT NOT NULL DEFAULT 0,            -- Всего использовано пользователем
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Автоматическое создание записи для новых пользователей
CREATE TRIGGER trigger_create_token_bank
AFTER INSERT ON "User"
FOR EACH ROW
EXECUTE FUNCTION create_token_bank_for_new_user();
```

### Code Structure

```
internal/
├── models/
│   └── token_bank.go              # TokenBank model & request/response types
├── database/
│   └── token_bank_repository.go   # CRUD operations & business logic
└── modules/
    └── admin/
        ├── service/
        │   └── service.go         # AdminService interface with token methods
        └── transport/http/
            └── handlers.go        # HTTP handlers for token endpoints
```

### API Layer

```
Routes registered in admin/module.go:
- GET    /api/admin/token-bank              → GetAllTokenBanks
- GET    /api/admin/token-bank/stats        → GetTokenBankStats
- GET    /api/admin/token-bank/{userID}     → GetUserTokenBank
- POST   /api/admin/token-bank/allocate     → AllocateTokens
- POST   /api/admin/token-bank/revoke       → RevokeTokens
- PUT    /api/admin/token-bank/balance      → SetTokenBalance

All endpoints:
- Require: Valid JWT token
- Require: Admin role (role = "admin")
```

---

## Database Setup

### Step 1: Ensure Migrations are Applied

The token bank table is created by two migrations:

1. **`013_create_token_bank.sql`** - Creates the `token_bank` table with constraints
2. **`014_initialize_token_banks.sql`** - Initializes records for existing users and adds trigger

### Step 2: Manual SQL Execution (if needed)

If migrations aren't automatically applied, execute manually:

```bash
# Connect to production database
psql $DATABASE_URL < migrations/013_create_token_bank.sql
psql $DATABASE_URL < migrations/014_initialize_token_banks.sql
```

### Step 3: Verify Table Creation

```sql
-- Check if table exists
SELECT * FROM token_bank LIMIT 5;

-- Check table structure
\d token_bank

-- Count records
SELECT COUNT(*) FROM token_bank;

-- Check if trigger exists
SELECT * FROM pg_trigger WHERE tgname LIKE '%token_bank%';
```

---

## API Endpoints

### 1. GET /api/admin/token-bank

**Get all token banks for all users**

```bash
curl -X GET https://api.example.com/api/admin/token-bank \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```

**Response:**
```json
[
  {
    "id": "bank-id-001",
    "user_id": "user-id-123",
    "balance": 500,
    "total_allocated": 1000,
    "total_used": 500,
    "created_at": "2024-11-01T10:00:00Z",
    "updated_at": "2024-11-10T15:30:00Z",
    "user": {
      "id": "user-id-123",
      "email": "user@example.com",
      "name": "John Doe",
      "role": "user",
      "createdAt": "2024-10-15T14:30:00Z"
    }
  }
]
```

---

### 2. GET /api/admin/token-bank/stats

**Get system-wide token statistics**

```bash
curl -X GET https://api.example.com/api/admin/token-bank/stats \
  -H "Authorization: Bearer $ADMIN_TOKEN"
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

### 3. GET /api/admin/token-bank/{userID}

**Get token bank for specific user**

```bash
curl -X GET https://api.example.com/api/admin/token-bank/user-id-123 \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```

**Response:**
```json
{
  "id": "bank-id-001",
  "user_id": "user-id-123",
  "balance": 500,
  "total_allocated": 1000,
  "total_used": 500,
  "created_at": "2024-11-01T10:00:00Z",
  "updated_at": "2024-11-10T15:30:00Z",
  "user": {
    "id": "user-id-123",
    "email": "user@example.com",
    "name": "John Doe",
    "role": "user",
    "createdAt": "2024-10-15T14:30:00Z"
  }
}
```

---

### 4. POST /api/admin/token-bank/allocate

**Allocate tokens to a user**

```bash
curl -X POST https://api.example.com/api/admin/token-bank/allocate \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user-id-123",
    "amount": 500,
    "reason": "Monthly allocation"
  }'
```

**Response:**
```json
{
  "message": "Tokens allocated successfully",
  "user_id": "user-id-123",
  "amount": 500
}
```

**Important:**
- `amount` must be positive
- Increases both `balance` and `total_allocated`
- User's token bank must exist

---

### 5. POST /api/admin/token-bank/revoke

**Revoke tokens from a user**

```bash
curl -X POST https://api.example.com/api/admin/token-bank/revoke \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user-id-123",
    "amount": 100,
    "reason": "Policy violation"
  }'
```

**Response:**
```json
{
  "message": "Tokens revoked successfully",
  "user_id": "user-id-123",
  "amount": 100
}
```

**Important:**
- `amount` must be positive
- User must have sufficient balance
- Only decreases `balance` (not `total_allocated`)
- Can only revoke available tokens

---

### 6. PUT /api/admin/token-bank/balance

**Set exact token balance**

```bash
curl -X PUT https://api.example.com/api/admin/token-bank/balance \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user-id-123",
    "balance": 1000
  }'
```

**Response:**
```json
{
  "message": "Token balance set successfully",
  "user_id": "user-id-123",
  "balance": 1000
}
```

**Important:**
- Sets exact balance (replaces current value)
- `balance` must be non-negative
- Does not affect `total_allocated` or `total_used`
- Use for corrections or special cases

---

## Testing

### Local Testing

```bash
# 1. Start the server locally
go run cmd/server/main.go

# 2. Login as admin and get token
TOKEN=$(curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"admin_password_123"}' \
  | jq -r '.token')

# 3. Get a test user ID
USER_ID=$(curl -X GET http://localhost:8080/api/admin/users \
  -H "Authorization: Bearer $TOKEN" \
  | jq -r '.[0].id')

# 4. Test token bank endpoints
ADMIN_TOKEN=$TOKEN TEST_USER_ID=$USER_ID bash test_token_bank_api.sh
```

### Production Testing

```bash
# Set variables
export ADMIN_TOKEN="your_production_admin_token"
export TEST_USER_ID="production_user_id"
export API_URL="https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app"

# Run tests
bash test_token_bank_api.sh
```

---

## Production Deployment

### Step 1: Code Changes

All code changes are already in place:
- ✅ `internal/models/token_bank.go` - Model definitions
- ✅ `internal/database/token_bank_repository.go` - Repository layer
- ✅ `internal/modules/admin/service/service.go` - Service layer (updated)
- ✅ `internal/modules/admin/transport/http/handlers.go` - HTTP handlers
- ✅ `internal/modules/admin/module.go` - Routes registered
- ✅ `internal/database/db.go` - AutoMigrate updated

### Step 2: Deploy to Koyeb

```bash
# 1. Commit all changes
git add -A
git commit -m "✨ feat: Add Token Bank admin panel feature"

# 2. Push to GitHub (triggers Koyeb deployment)
git push origin main

# 3. Wait for Koyeb to build and deploy (usually 2-5 minutes)
# Monitor: https://app.koyeb.com/

# 4. Verify deployment
curl -X GET https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/admin/token-bank \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```

### Step 3: Initialize Token Banks (if needed)

If token banks weren't automatically created, run initialization:

```bash
# Option 1: Using psql directly
psql $PRODUCTION_DATABASE_URL < migrations/014_initialize_token_banks.sql

# Option 2: Via Go migration script
cd cmd/migrate && go run main.go
```

### Step 4: Verify Data

```sql
-- Check token_bank table
SELECT COUNT(*) as total_records FROM token_bank;

-- Check if all users have token banks
SELECT u.id, u.email, t.id as token_bank_id
FROM "User" u
LEFT JOIN token_bank t ON u.id = t.user_id
WHERE t.id IS NULL;

-- Check sample balances
SELECT u.email, t.balance, t.total_allocated, t.total_used
FROM "User" u
JOIN token_bank t ON u.id = t.user_id
LIMIT 10;
```

---

## Troubleshooting

### Issue 1: 404 Errors on Token Bank Endpoints

**Symptoms:**
```
GET /api/admin/token-bank 404 (Not Found)
GET /api/admin/token-bank/stats 404 (Not Found)
```

**Causes & Solutions:**

1. **TokenBank not registered in AutoMigrate**
   ```go
   // Check: internal/database/db.go
   // Should include: &models.TokenBank{}
   ```
   ✅ Already fixed in db.go

2. **Migration not applied**
   ```sql
   -- Manually run:
   psql $DATABASE_URL < migrations/013_create_token_bank.sql
   ```

3. **Server not rebuilt**
   ```bash
   go build ./cmd/server
   ```

### Issue 2: Foreign Key Errors

**Error:** `violates foreign key constraint "token_bank_user_id_fkey"`

**Solution:** Ensure user exists before allocating tokens:
```bash
# Verify user exists
curl -X GET /api/admin/users/$USER_ID -H "Authorization: Bearer $TOKEN"
```

### Issue 3: "Insufficient Tokens" Error

**Error:** `400 Bad Request: Insufficient tokens to revoke`

**Cause:** User's balance is less than the revoke amount

**Solution:** Check current balance first:
```bash
curl -X GET /api/admin/token-bank/$USER_ID -H "Authorization: Bearer $TOKEN"
```

### Issue 4: Trigger Not Working for New Users

**Problem:** New users don't automatically get token_bank entries

**Solution:** Manually run initialization:
```sql
INSERT INTO token_bank (user_id, balance, total_allocated, total_used)
SELECT id, 0, 0, 0 FROM "User" u
WHERE NOT EXISTS (SELECT 1 FROM token_bank t WHERE t.user_id = u.id)
ON CONFLICT (user_id) DO NOTHING;
```

---

## Monitoring & Maintenance

### Regular Health Checks

```bash
#!/bin/bash
# Check token bank status

ADMIN_TOKEN="$1"
API_URL="${2:-https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app}"

echo "📊 Token Bank Health Check"
echo "=========================="

# 1. Check API availability
echo -n "1. API availability: "
if curl -s -X GET "$API_URL/api/admin/token-bank" \
  -H "Authorization: Bearer $ADMIN_TOKEN" > /dev/null 2>&1; then
  echo "✅ Online"
else
  echo "❌ Offline"
fi

# 2. Get statistics
echo -n "2. Fetching statistics: "
STATS=$(curl -s -X GET "$API_URL/api/admin/token-bank/stats" \
  -H "Authorization: Bearer $ADMIN_TOKEN")
echo "✅"
echo "   $STATS" | jq '.'

# 3. Count records
echo -n "3. Database check: "
TOTAL=$(echo "$STATS" | jq '.total_users_with_tokens')
echo "✅ Found $TOTAL users with token banks"
```

### Performance Optimization

If you notice slow queries on large datasets:

```sql
-- Create indexes for faster queries
CREATE INDEX IF NOT EXISTS idx_token_bank_balance 
  ON token_bank(balance) 
  WHERE balance > 0;

CREATE INDEX IF NOT EXISTS idx_token_bank_created_at 
  ON token_bank(created_at DESC);
```

---

## Next Steps

1. ✅ Deploy to Koyeb
2. ✅ Verify endpoints return data
3. ✅ Test token allocation workflow
4. ✅ Monitor production logs for errors
5. 📋 (Optional) Add token transaction history endpoint
6. 📋 (Optional) Add token expiration feature
7. 📋 (Optional) Add audit logging for token changes

---

## Support

For issues or questions:
1. Check the [Troubleshooting](#troubleshooting) section
2. Review logs: `docker logs <container_id>`
3. Check database: `psql $DATABASE_URL`
4. Review source code in `internal/modules/admin/`
