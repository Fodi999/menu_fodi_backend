# User Roles & Access Control - Summary

## Date: 2026-01-24

---

## Database Table Structure

**Table Name:** `"User"` (Prisma naming convention with capital U)

**Important:** Not `users` (lowercase) - that's a different/old table with only 3 test users!

```sql
-- Correct table:
SELECT * FROM "User" WHERE ...

-- Wrong table (old/unused):
SELECT * FROM users WHERE ...
```

---

## Role System Architecture

### Defined Roles (from `internal/models/user.go`)

```go
const (
    RoleHomeChef   = "home_chef"   // Домашний повар (default)
    RoleProChef    = "pro_chef"    // Профессиональный повар / ресторан
    RoleAdmin      = "admin"       // Администратор
    RoleSuperAdmin = "super_admin" // Супер администратор
)
```

### User Status

```go
const (
    UserStatusActive  = "active"  // Can login and use all features
    UserStatusBlocked = "blocked" // Blocked by admin - cannot login
    UserStatusPending = "pending" // Unverified / limited access
)
```

---

## Current Production Database State

### Users by Role (Total: 40 users)

| Role | Status | Count | Description |
|------|--------|-------|-------------|
| `home_chef` | active | 36 | Regular users (домашние повара) |
| `admin` | active | 2 | Administrators |
| `super_admin` | active | 1 | Super Administrator (highest privileges) |
| `investor` | active | 1 | Special role (investor access) |

### Admin Users

| Email | Name | Role | User ID | Created |
|-------|------|------|---------|---------|
| `admin@fodisushi.com` | Администратор | **admin** | cmgap6c140000ujeubf0gfc4z | 2025-10-03 |
| `admin@fodi.com` | Admin User | **admin** | ea506c38-0c50-4f6a-a401-7789be03e3bc | 2025-10-16 |
| `admin@example.com` | System Administrator | **super_admin** | 7ec8aba4-8195-4be1-a9a8-067c30aae306 | 2025-11-10 |

### Your User

| Email | Name | Role | User ID | Status |
|-------|------|------|---------|--------|
| `fodi85@gmail.ru` | Dima Fomin | **home_chef** | 407582be-59d5-4d21-873b-1a72d31b0d42 | active |

---

## Access Control Middleware

### Location: `internal/middleware/auth.go`

#### 1. AdminMiddleware
```go
func AdminMiddleware(next http.Handler) http.Handler {
    // Allows: admin, super_admin
    // Blocks: home_chef, pro_chef, investor, etc.
}
```

**Usage:**
```go
r.With(middleware.AdminMiddleware).Get("/admin/users", handler.GetUsers)
```

**Allowed Roles:**
- ✅ `admin`
- ✅ `super_admin`

---

#### 2. SuperAdminMiddleware
```go
func SuperAdminMiddleware(next http.Handler) http.Handler {
    // Allows: ONLY super_admin
    // Blocks: all other roles
}
```

**Usage:**
```go
r.With(middleware.SuperAdminMiddleware).Post("/admin/roles", handler.ChangeRole)
```

**Allowed Roles:**
- ✅ `super_admin` ONLY

---

## Role Permissions Matrix

| Feature | home_chef | pro_chef | admin | super_admin |
|---------|-----------|----------|-------|-------------|
| **Basic Features** |
| Login / Register | ✅ | ✅ | ✅ | ✅ |
| View Recipes | ✅ | ✅ | ✅ | ✅ |
| Manage Fridge | ✅ | ✅ | ✅ | ✅ |
| Kitchen Pipeline | ✅ | ✅ | ✅ | ✅ |
| Save Recipes | ✅ | ✅ | ✅ | ✅ |
| **Pro Features** |
| Stock Management | ❌ | ✅ | ✅ | ✅ |
| Inventory System | ❌ | ✅ | ✅ | ✅ |
| Business Analytics | ❌ | ✅ | ✅ | ✅ |
| **Admin Features** |
| View All Users | ❌ | ❌ | ✅ | ✅ |
| Manage Recipes | ❌ | ❌ | ✅ | ✅ |
| Manage Canonical Ingredients | ❌ | ❌ | ✅ | ✅ |
| View System Stats | ❌ | ❌ | ✅ | ✅ |
| Token Management | ❌ | ❌ | ✅ | ✅ |
| **Super Admin Features** |
| Change User Roles | ❌ | ❌ | ❌ | ✅ |
| Block/Unblock Users | ❌ | ❌ | ❌ | ✅ |
| System Configuration | ❌ | ❌ | ❌ | ✅ |
| Database Migrations | ❌ | ❌ | ❌ | ✅ |

---

## API Endpoints by Role

### Public Endpoints (No Auth Required)
```
POST /api/auth/register
POST /api/auth/login
GET  /api/recipes (public catalog)
```

### Home Chef Endpoints (Default Role)
```
GET  /api/fridge
POST /api/fridge/items
GET  /api/menu/today
POST /api/menu/today
GET  /api/menu/history
GET  /api/recipes/saved
GET  /api/recipes/recommendations
```

### Admin Endpoints
```
GET    /api/admin/users
POST   /api/admin/ingredients/canonical
PATCH  /api/admin/ingredients/canonical/:id
DELETE /api/admin/ingredients/canonical/:id
GET    /api/admin/stats
POST   /api/admin/treasury/allocate
```

### Super Admin Endpoints
```
POST   /api/admin/users/:id/role
POST   /api/admin/users/:id/block
POST   /api/admin/users/:id/unblock
DELETE /api/admin/users/:id
```

---

## Code Examples

### Check User Role in Handler

```go
func (h *Handler) AdminOnlyEndpoint(w http.ResponseWriter, r *http.Request) {
    userRole, ok := r.Context().Value("userRole").(string)
    if !ok || (userRole != models.RoleAdmin && userRole != models.RoleSuperAdmin) {
        utils.RespondError(w, http.StatusForbidden, "forbidden", "admin access required")
        return
    }
    
    // Admin logic here
}
```

### Using Middleware in Routes

```go
// Admin routes
r.Route("/admin", func(r chi.Router) {
    r.Use(middleware.AdminMiddleware) // Allows: admin, super_admin
    
    r.Get("/users", adminHandler.GetUsers)
    r.Get("/stats", adminHandler.GetStats)
    
    // Super admin only
    r.With(middleware.SuperAdminMiddleware).Post("/users/{id}/role", adminHandler.ChangeRole)
})
```

---

## How to Change User Role

### Via Database (Direct)

```sql
-- Promote user to admin
UPDATE "User" 
SET role = 'admin' 
WHERE email = 'user@example.com';

-- Promote to super_admin
UPDATE "User" 
SET role = 'super_admin' 
WHERE id = '407582be-59d5-4d21-873b-1a72d31b0d42';

-- Demote to home_chef
UPDATE "User" 
SET role = 'home_chef' 
WHERE email = 'admin@example.com';
```

### Via API (Super Admin Only)

```bash
# Endpoint (when implemented):
POST /api/admin/users/{userId}/role
Authorization: Bearer {super_admin_token}
Content-Type: application/json

{
  "role": "admin"
}
```

---

## Common Queries

### List All Admins
```sql
SELECT id, email, name, role, "createdAt" 
FROM "User" 
WHERE role IN ('admin', 'super_admin') 
ORDER BY role, "createdAt";
```

### Count Users by Role
```sql
SELECT role, status, COUNT(*) as count 
FROM "User" 
GROUP BY role, status 
ORDER BY role;
```

### Find User by Email
```sql
SELECT id, email, name, role, status 
FROM "User" 
WHERE email = 'fodi85@gmail.ru';
```

### Check User Permissions
```sql
SELECT id, email, role, status, "lastLogin" 
FROM "User" 
WHERE id = '407582be-59d5-4d21-873b-1a72d31b0d42';
```

---

## Security Notes

1. ⚠️ **Super Admin Access** - Only 1 super_admin exists. Protect this account!
2. 🔒 **Role Changes** - Only super_admin can change user roles
3. 🚫 **Blocked Users** - status='blocked' prevents login via JWT validation
4. ✅ **Default Role** - New registrations get `home_chef` by default
5. 🔑 **Token Claims** - JWT includes `role` claim for quick permission checks

---

## Testing Role-Based Access

### Test Script Template

```bash
#!/bin/bash

# Get tokens for different roles
HOME_CHEF_TOKEN="..." # fodi85@gmail.ru
ADMIN_TOKEN="..."     # admin@fodi.com
SUPER_ADMIN_TOKEN="..." # admin@example.com

# Test home_chef access
curl -H "Authorization: Bearer $HOME_CHEF_TOKEN" \
  https://api.example.com/api/menu/today
# Expected: 200 OK

# Test admin access (should fail for home_chef)
curl -H "Authorization: Bearer $HOME_CHEF_TOKEN" \
  https://api.example.com/api/admin/users
# Expected: 403 Forbidden

# Test super admin access (should fail for regular admin)
curl -H "Authorization: Bearer $ADMIN_TOKEN" \
  -X POST https://api.example.com/api/admin/users/123/role \
  -d '{"role":"admin"}'
# Expected: 403 Forbidden
```

---

## Related Files

- **User Model:** `internal/models/user.go`
- **Auth Middleware:** `internal/middleware/auth.go`
- **JWT Service:** `internal/modules/auth/service/auth_service.go`
- **Admin Routes:** `internal/modules/admin/module.go`

---

## Summary

✅ **Total Users:** 40 (36 home_chef, 2 admin, 1 super_admin, 1 investor)  
✅ **Role System:** 4 distinct roles with hierarchical permissions  
✅ **Your Account:** home_chef (fodi85@gmail.ru)  
✅ **Database Table:** `"User"` (capital U, Prisma convention)  
✅ **Middleware:** AdminMiddleware, SuperAdminMiddleware implemented  

---

**Last Updated:** 2026-01-24  
**Database:** Neon PostgreSQL (production)  
**Status:** ✅ Active & Working
