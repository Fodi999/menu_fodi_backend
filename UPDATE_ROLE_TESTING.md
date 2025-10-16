# Update User Role Endpoint - Testing Guide

## 📋 Overview

Эндпоинт для обновления ролей пользователей. Доступен только администраторам.

**Endpoint:** `PATCH /api/admin/users/update-role`  
**Auth Required:** ✅ Yes (Admin JWT token)  
**Admin Only:** ✅ Yes

---

## 🎯 Available Roles

1. **user** - обычный пользователь (default)
2. **admin** - администратор системы
3. **business_owner** - владелец бизнеса
4. **investor** - инвестор

---

## 🔧 Database Setup

### Add Role Enum Values (если еще не добавлены)

```sql
-- Add business_owner role
ALTER TYPE "Role" ADD VALUE 'business_owner';

-- Add investor role
ALTER TYPE "Role" ADD VALUE 'investor';

-- Verify enum values
SELECT enumlabel FROM pg_enum 
WHERE enumtypid = (SELECT oid FROM pg_type WHERE typname = 'Role') 
ORDER BY enumsortorder;
```

**Migration file:** `migrations/012_add_role_enum_values.sql`

---

## 🧪 Testing

### 1. Get Admin Token

```bash
ADMIN_TOKEN=$(curl -s -X POST http://127.0.0.1:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@fodi.com","password":"admin123"}' \
  | jq -r '.token')

echo "Admin token: $ADMIN_TOKEN"
```

### 2. Update User Role

```bash
curl -X PATCH http://127.0.0.1:8080/api/admin/users/update-role \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "ea506c38-0c50-4f6a-a401-7789be03e3bc",
    "role": "business_owner"
  }' | jq
```

### 3. Expected Response

```json
{
  "message": "✅ User role updated successfully",
  "user_id": "ea506c38-0c50-4f6a-a401-7789be03e3bc",
  "old_role": "user",
  "new_role": "business_owner",
  "name": "Admin User",
  "email": "admin@fodi.com",
  "updated_by": "ea506c38-0c50-4f6a-a401-7789be03e3bc"
}
```

---

## 🔒 Security Testing

### Test 1: Without Authorization Header

```bash
curl -s -X PATCH http://127.0.0.1:8080/api/admin/users/update-role \
  -H "Content-Type: application/json" \
  -d '{"user_id":"test-id", "role":"admin"}' | jq
```

**Expected:** `{"error": "Authorization header required"}`

### Test 2: With Invalid Token

```bash
curl -s -X PATCH http://127.0.0.1:8080/api/admin/users/update-role \
  -H "Authorization: Bearer invalid_token" \
  -H "Content-Type: application/json" \
  -d '{"user_id":"test-id", "role":"admin"}' | jq
```

**Expected:** `{"error": "Invalid or expired token"}`

### Test 3: With Non-Admin User Token

```bash
# Login as regular user
USER_TOKEN=$(curl -s -X POST http://127.0.0.1:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password123"}' \
  | jq -r '.token')

# Try to update role
curl -s -X PATCH http://127.0.0.1:8080/api/admin/users/update-role \
  -H "Authorization: Bearer $USER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"user_id":"test-id", "role":"admin"}' | jq
```

**Expected:** `{"error": "Only admins can update user roles"}`

---

## 🎭 Complete Test Suite

Test all 4 roles in sequence:

```bash
echo "🧪 Testing all roles..."
echo ""

# 1. investor
echo "1️⃣ Changing to investor:"
curl -s -X PATCH http://127.0.0.1:8080/api/admin/users/update-role \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"user_id":"ea506c38-0c50-4f6a-a401-7789be03e3bc", "role":"investor"}' \
  | jq -r '"✅ Role: " + .new_role'
echo ""

# 2. business_owner
echo "2️⃣ Changing to business_owner:"
curl -s -X PATCH http://127.0.0.1:8080/api/admin/users/update-role \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"user_id":"ea506c38-0c50-4f6a-a401-7789be03e3bc", "role":"business_owner"}' \
  | jq -r '"✅ Role: " + .new_role'
echo ""

# 3. user
echo "3️⃣ Changing to user:"
curl -s -X PATCH http://127.0.0.1:8080/api/admin/users/update-role \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"user_id":"ea506c38-0c50-4f6a-a401-7789be03e3bc", "role":"user"}' \
  | jq -r '"✅ Role: " + .new_role'
echo ""

# 4. admin (restore)
echo "4️⃣ Restoring to admin:"
curl -s -X PATCH http://127.0.0.1:8080/api/admin/users/update-role \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"user_id":"ea506c38-0c50-4f6a-a401-7789be03e3bc", "role":"admin"}' \
  | jq -r '"✅ Role: " + .new_role'
```

---

## ❌ Error Cases

### Invalid Role

```bash
curl -s -X PATCH http://127.0.0.1:8080/api/admin/users/update-role \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"user_id":"test-id", "role":"superuser"}' | jq
```

**Expected:** `{"error": "Invalid role. Must be: user, admin, business_owner, or investor"}`

### User Not Found

```bash
curl -s -X PATCH http://127.0.0.1:8080/api/admin/users/update-role \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"user_id":"non-existent-id", "role":"user"}' | jq
```

**Expected:** `{"error": "User not found"}`

### Invalid Request Body

```bash
curl -s -X PATCH http://127.0.0.1:8080/api/admin/users/update-role \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d 'invalid json' | jq
```

**Expected:** `{"error": "Invalid request body"}`

---

## 📊 Verification

Verify role change in database:

```bash
source .env && psql "$DATABASE_URL" -c \
  "SELECT id, email, name, role FROM \"User\" WHERE email = 'admin@fodi.com';"
```

---

## ✅ Test Results (16.10.2025)

All tests passed successfully:

- ✅ Role update: user → investor
- ✅ Role update: investor → business_owner
- ✅ Role update: business_owner → user
- ✅ Role update: user → admin
- ✅ Authorization check (no token)
- ✅ Authorization check (invalid token)
- ✅ Admin role verification
- ✅ Invalid role rejection
- ✅ User not found handling
- ✅ Invalid request body handling

---

**Status:** ✅ Fully Functional  
**Last Updated:** 16 октября 2025 г.
