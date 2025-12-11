# 🏦 Treasury System Documentation

**Status:** ✅ Production Ready  
**Created:** December 12, 2025  
**Total Supply:** 1,000,000,000 tokens

## Overview

The Treasury system is a centralized token management system with a fixed supply of 1 billion tokens. All token allocations to users come from this central treasury, ensuring controlled token economics and transparent distribution tracking.

## Architecture

### Treasury Account
- **User ID:** `TREASURY`
- **Email:** `treasury@system.internal`
- **Role:** `admin` (system account, cannot login)
- **Initial Balance:** 1,000,000,000 tokens
- **Type:** Special system user with token bank

### Database Schema

```sql
-- Treasury User
INSERT INTO "User" (id, email, name, password, role, "createdAt")
VALUES (
    'TREASURY',
    'treasury@system.internal',
    'Platform Treasury',
    '$2a$10$TREASURY_SYSTEM_ACCOUNT_NOT_FOR_LOGIN',
    'admin',
    NOW()
);

-- Treasury Token Bank
INSERT INTO token_bank (id, user_id, balance, total_allocated, total_used)
VALUES (
    gen_random_uuid(),
    'TREASURY',
    1000000000,  -- 1 billion initial supply
    1000000000,  -- Total supply allocated at creation
    0            -- No tokens distributed yet
);
```

## Token Flow

```
Treasury (1B tokens) 
    ↓
    ├─> User Registration Bonus (100 tokens)
    ├─> Quest Rewards (variable)
    ├─> Achievement Rewards (variable)
    └─> Admin Manual Allocation (variable)
```

### Transaction Mechanics

1. **Atomic Transactions:** All allocations use database transactions to ensure consistency
2. **Balance Tracking:**
   - Treasury: `balance` decreases, `total_used` increases
   - User: `balance` increases, `total_allocated` increases
3. **Validation:** Checks treasury balance before allocation
4. **Rollback:** Any failure rolls back entire transaction

## API Endpoints

### 1. Get Treasury Info
```http
GET /api/admin/treasury
Authorization: Bearer {admin_token}
```

**Response:**
```json
{
  "id": "uuid",
  "user_id": "TREASURY",
  "balance": 999994000,
  "total_allocated": 1000000000,
  "total_used": 6000,
  "total_supply": 1000000000,
  "distributed": 6000,
  "remaining": 999994000,
  "created_at": "2025-12-10T23:43:21.226714Z",
  "updated_at": "2025-12-10T23:52:52.000755Z"
}
```

### 2. Allocate from Treasury
```http
POST /api/admin/treasury/allocate
Authorization: Bearer {admin_token}
Content-Type: application/json

{
  "user_id": "user-uuid",
  "amount": 1000,
  "reason": "Quest completion reward"
}
```

**Response:**
```json
{
  "message": "Tokens allocated from treasury successfully",
  "user_id": "user-uuid",
  "amount": 1000,
  "source": "TREASURY"
}
```

## Repository Methods

### Core Methods

#### `AllocateFromTreasury(userID string, amount int64) error`
Main allocation method. Used by all token distribution mechanisms.

```go
// Example: Award quest completion
repo.AllocateFromTreasury(userID, 50)
```

#### `GetTreasuryBalance() (int64, error)`
Check available treasury balance.

```go
balance, err := repo.GetTreasuryBalance()
```

#### `GetTreasuryInfo() (*models.TokenBank, error)`
Get complete treasury information.

```go
treasury, err := repo.GetTreasuryInfo()
```

### Helper Methods

#### `AllocateWelcomeBonus(userID string, bonusAmount int64) error`
Automatically allocate welcome bonus to new users (default: 100 tokens).

```go
// In registration handler
repo.AllocateWelcomeBonus(newUserID, 100)
```

#### `AllocateQuestReward(userID, questID string, rewardAmount int64) error`
Award tokens for quest completion.

```go
// When user completes quest
repo.AllocateQuestReward(userID, "quest-123", 50)
```

#### `AllocateAchievementReward(userID, achievementID string, rewardAmount int64) error`
Award tokens for achievement unlocked.

```go
// When user unlocks achievement
repo.AllocateAchievementReward(userID, "chef-master", 200)
```

#### `CheckTreasuryBalance(requiredAmount int64) (bool, error)`
Verify treasury has sufficient balance.

```go
hasBalance, err := repo.CheckTreasuryBalance(1000)
```

## Integration Guide

### 1. User Registration with Welcome Bonus

```go
// In auth/service/service.go - Register method
func (s *authService) Register(req RegisterRequest) (*User, error) {
    // Create user...
    newUser, err := s.repo.CreateUser(user)
    if err != nil {
        return nil, err
    }

    // Initialize token bank
    tokenRepo := &database.TokenBankRepository{}
    if err := tokenRepo.InitializeTokenBankForUser(newUser.ID); err != nil {
        return nil, err
    }

    // Award welcome bonus from treasury
    if err := tokenRepo.AllocateWelcomeBonus(newUser.ID, 100); err != nil {
        log.Printf("Failed to allocate welcome bonus: %v", err)
        // Non-fatal - user can still use the app
    }

    return newUser, nil
}
```

### 2. Quest Completion Reward

```go
// In quest/service/service.go - CompleteQuest method
func (s *questService) CompleteQuest(userID, questID string) error {
    // Mark quest as completed...
    
    // Award tokens from treasury
    tokenRepo := &database.TokenBankRepository{}
    rewardAmount := quest.RewardTokens // e.g., 50 tokens
    
    if err := tokenRepo.AllocateQuestReward(userID, questID, rewardAmount); err != nil {
        log.Printf("Failed to allocate quest reward: %v", err)
        return err
    }

    return nil
}
```

### 3. Achievement Unlock Reward

```go
// In achievement/service/service.go - UnlockAchievement method
func (s *achievementService) UnlockAchievement(userID, achievementID string) error {
    // Record achievement...
    
    // Award tokens from treasury
    tokenRepo := &database.TokenBankRepository{}
    rewardAmount := achievement.TokenReward // e.g., 200 tokens
    
    if err := tokenRepo.AllocateAchievementReward(userID, achievementID, rewardAmount); err != nil {
        log.Printf("Failed to allocate achievement reward: %v", err)
        return err
    }

    return nil
}
```

## Error Handling

### Common Errors

| Error | Cause | Solution |
|-------|-------|----------|
| `treasury not found` | Treasury not initialized | Run migration 016 |
| `insufficient treasury balance` | Not enough tokens left | Check total supply usage |
| `user token bank not found` | User's token bank not created | Initialize user token bank first |
| `amount must be positive` | Invalid amount parameter | Validate input |

### Best Practices

1. **Always check treasury balance** before large allocations
2. **Log all allocations** for audit trail
3. **Handle errors gracefully** - don't block user actions if token allocation fails
4. **Use transactions** for multiple operations
5. **Validate user token bank exists** before allocation

## Monitoring

### Key Metrics to Track

```sql
-- Total tokens distributed
SELECT total_used FROM token_bank WHERE user_id = 'TREASURY';

-- Remaining supply
SELECT balance FROM token_bank WHERE user_id = 'TREASURY';

-- Distribution rate (tokens/day)
SELECT 
    DATE(updated_at) as date,
    total_used,
    total_used - LAG(total_used) OVER (ORDER BY DATE(updated_at)) as daily_distribution
FROM token_bank 
WHERE user_id = 'TREASURY'
ORDER BY date DESC;

-- Top recipients
SELECT 
    u.name,
    u.email,
    tb.balance,
    tb.total_allocated
FROM token_bank tb
JOIN "User" u ON tb.user_id = u.id
WHERE u.id != 'TREASURY'
ORDER BY tb.total_allocated DESC
LIMIT 10;
```

## Security

1. **Treasury Account:** Cannot be used for login (invalid bcrypt hash)
2. **Admin Only:** Treasury endpoints require admin role
3. **Atomic Transactions:** Prevent double-spending and inconsistencies
4. **Audit Trail:** All transactions logged with `total_used` tracking
5. **Balance Validation:** Checks before each allocation

## Migration

**File:** `migrations/016_create_treasury_token_bank.sql`

```bash
# Apply migration
psql $DATABASE_URL < migrations/016_create_treasury_token_bank.sql

# Verify
psql $DATABASE_URL -c "SELECT balance, total_allocated FROM token_bank WHERE user_id = 'TREASURY';"
```

## Testing

### Local Testing

```bash
# 1. Start server
export DATABASE_URL="your-db-url"
export JWT_SECRET="your-secret"
./bin/server

# 2. Get admin token
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"admin_password_123"}'

# 3. Check treasury
curl http://localhost:8080/api/admin/treasury \
  -H "Authorization: Bearer $TOKEN"

# 4. Allocate tokens
curl -X POST http://localhost:8080/api/admin/treasury/allocate \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"user_id":"user-uuid","amount":100,"reason":"Test"}'
```

### Production Testing

```bash
# Check treasury balance
curl https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/admin/treasury \
  -H "Authorization: Bearer $PROD_TOKEN"

# Allocate from treasury
curl -X POST https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/admin/treasury/allocate \
  -H "Authorization: Bearer $PROD_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"user_id":"uuid","amount":100}'
```

## Current Status (Production)

**Date:** December 12, 2025

```json
{
  "total_supply": 1000000000,
  "distributed": 6000,
  "remaining": 999994000,
  "utilization": "0.0006%",
  "status": "healthy"
}
```

## Roadmap

- [ ] Add treasury withdrawal mechanism (for burning tokens)
- [ ] Implement treasury recharge system (if needed)
- [ ] Add treasury transaction history endpoint
- [ ] Create treasury analytics dashboard
- [ ] Add treasury alert system (low balance warning)
- [ ] Implement treasury backup system

## Files Modified

1. `migrations/016_create_treasury_token_bank.sql` - Treasury creation migration
2. `internal/database/token_bank_repository.go` - Treasury methods
3. `internal/modules/admin/service/service.go` - Treasury service layer
4. `internal/modules/admin/transport/http/handlers.go` - Treasury HTTP handlers
5. `internal/modules/admin/module.go` - Treasury routes registration

## Conclusion

The Treasury system provides a robust, secure, and scalable foundation for token distribution across the platform. With a fixed supply of 1 billion tokens and atomic transaction guarantees, it ensures controlled token economics and transparent distribution tracking.

**Status:** ✅ Production Ready and Operational
