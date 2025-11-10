# 🔑 User Roles & Permissions System

**Date**: November 10, 2025  
**System**: Role-Based Access Control (RBAC)

---

## Admin Role

### Role String
```
"admin"
```

### Where It's Used

**Model Definition** (`internal/models/user.go`):
```go
Role string `gorm:"column:role;default:user" json:"role"` // "user" или "admin"
```

**Default Value**: `"user"` (users are not admins by default)

### Checking Admin Role

**Middleware** (`internal/middleware/auth.go`):
```go
func AdminMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        claims := r.Context().Value(UserContextKey).(*authservice.Claims)
        
        // Check if role is exactly "admin"
        if claims.Role != "admin" {
            utils.WriteError(w, http.StatusForbidden, "Admin access required")
            return
        }
        
        next.ServeHTTP(w, r)
    })
}
```

### Admin Permissions

Users with role = `"admin"` can access:

```
✅ GET    /api/admin/users              - List all users
✅ PUT    /api/admin/users/{id}         - Update user
✅ DELETE /api/admin/users/{id}         - Delete user
✅ PATCH  /api/admin/users/update-role  - Change user role
✅ GET    /api/admin/orders             - List all orders
✅ GET    /api/admin/orders/recent      - Get recent orders
✅ PUT    /api/admin/orders/{id}/status - Update order status
✅ GET    /api/admin/stats              - Get dashboard stats
```

---

## How to Create an Admin User

### Method 1: Direct Database Update

```sql
UPDATE "User" SET role = 'admin' WHERE email = 'admin@example.com';
```

### Method 2: Using Admin API (if another admin exists)

```bash
TOKEN="admin_jwt_token"

curl -X PATCH http://localhost:8080/api/admin/users/update-role \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user-uuid-to-make-admin",
    "role": "admin"
  }'
```

### Method 3: During User Registration

When user registers, they get role = `"user"` by default. An existing admin must promote them.

---

## JWT Claims Structure

When admin logs in, JWT token contains:

```json
{
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "admin@example.com",
  "role": "admin",
  "exp": 1731141600
}
```

The `role` field is checked by `AdminMiddleware`.

---

## Role Values in Database

| Role Value | Description | Can Access Admin Endpoints |
|-----------|-------------|---------------------------|
| `"user"` | Regular user | ❌ No |
| `"admin"` | Administrator | ✅ Yes |
| `"student"` | Student (in user_profile) | ❌ No |
| `"mentor"` | Mentor (in user_profile) | ❌ No |

---

## Code Flow: Admin Access Check

```
1. User logs in with email/password
   ↓
2. Backend verifies credentials
   ↓
3. JWT token generated with role from database
   ↓
4. Frontend stores token in localStorage
   ↓
5. Frontend makes request to admin endpoint with token
   ↓
6. AuthMiddleware validates JWT signature
   ↓
7. AdminMiddleware checks: claims.Role == "admin"
   ↓
8. If true → Handler executes
   If false → Return 403 Forbidden "Admin access required"
   ↓
9. Response sent to frontend
```

---

## Checking Current User Role (Frontend)

```javascript
// Decode JWT token to check role
function getUserRole(token) {
  const payload = token.split('.')[1];
  const decoded = JSON.parse(atob(payload));
  return decoded.role;  // returns "admin", "user", etc.
}

// Usage
const token = localStorage.getItem('token');
const role = getUserRole(token);

if (role === 'admin') {
  // Show admin panel
  window.location.href = '/admin/dashboard';
} else {
  // Show user panel
  window.location.href = '/dashboard';
}
```

---

## Security Notes

⚠️ **Important**: The role check happens on the backend, not the frontend

- Frontend can decode JWT to show/hide admin UI (UX optimization)
- But backend MUST verify role on every admin endpoint
- Never trust client-side role checks for security

✅ **Protected**: All admin endpoints require:
1. Valid JWT token (AuthMiddleware)
2. role = "admin" (AdminMiddleware)

❌ **Not Protected**: Changing your own role
- User cannot change their role via API
- Only admins can change roles
- But admins can change any user's role (including their own)

---

## Available Roles in System

### User Model (main)
- `"user"` - Regular user
- `"admin"` - Administrator

### UserProfile Model (extended)
- `"student"` - Student profile
- `"mentor"` - Mentor profile

---

## Related Files

| File | Purpose |
|------|---------|
| `internal/models/user.go` | User model with role field |
| `internal/middleware/auth.go` | AdminMiddleware implementation |
| `internal/modules/admin/module.go` | Admin routes registration |
| `internal/modules/admin/transport/http/handlers.go` | Admin handlers |

---

## Testing Admin Access

### Get Admin Token

```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.com",
    "password": "password123"
  }' | jq .token
```

### Test Admin Endpoint

```bash
TOKEN="eyJhbGciOiJIUzI1NiIs..."

curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/admin/stats
```

### Test Non-Admin Access

```bash
# User token (role = "user")
TOKEN="user_jwt_token"

curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/admin/stats

# Result: 403 Forbidden - "Admin access required"
```

---

## Summary

**Admin Role = `"admin"`**

- Stored in `User.Role` column in database
- Checked by `AdminMiddleware` on every request
- Included in JWT token claims
- Cannot be changed by user themselves
- Can only be set by existing admin or database direct update
- All admin endpoints protected by role check
