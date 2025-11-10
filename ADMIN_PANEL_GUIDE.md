# 🔐 Admin Panel - Complete Implementation Guide

**Date**: November 10, 2025  
**Status**: Production Ready  
**Version**: 1.0.0

---

## 📋 Table of Contents

1. [Overview](#overview)
2. [Architecture](#architecture)
3. [Authentication & Authorization](#authentication--authorization)
4. [API Endpoints](#api-endpoints)
5. [User Management](#user-management)
6. [Order Management](#order-management)
7. [Dashboard & Statistics](#dashboard--statistics)
8. [Frontend Integration](#frontend-integration)
9. [Security Considerations](#security-considerations)
10. [Deployment](#deployment)

---

## Overview

Admin panel is a role-based access control system implemented in the backend with middleware-based authentication. Only users with `admin` role can access admin endpoints.

### Key Features

✅ User Management (view, update, delete)  
✅ Role Management (change user roles)  
✅ Order Management (view, update status)  
✅ Dashboard Statistics (user count, order count)  
✅ JWT-based authentication  
✅ Role-based authorization  

---

## Architecture

### Component Structure

```
internal/modules/admin/
├── module.go              # Module initialization & routing
├── service/               # (empty - could be expanded)
└── transport/
    └── http/
        └── handlers.go    # HTTP handlers
```

### Middleware Stack

```
Request
  ↓
[Router]
  ↓
[AuthMiddleware] → Validates JWT token
  ↓
[AdminMiddleware] → Checks for admin role
  ↓
[Handler] → Processes request
  ↓
Response
```

### Data Flow

```
Frontend
  ↓ (Request with JWT token)
Admin API (/api/admin/*)
  ↓
AuthMiddleware (validate token)
  ↓
AdminMiddleware (check role = "admin")
  ↓
AdminHandlers (process request)
  ↓
Database (GORM)
  ↓ (Response JSON)
Frontend
```

---

## Authentication & Authorization

### JWT Token Structure

```json
{
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "admin@example.com",
  "role": "admin",
  "exp": 1731141600
}
```

### Admin Role

Users must have `role = "admin"` to access admin endpoints.

**Setting Admin Role:**

```go
// In database
UPDATE users SET role = 'admin' WHERE id = 'user-uuid'

// Via API
PATCH /api/admin/users/update-role
{
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "role": "admin"
}
```

### Middleware Implementation

**AuthMiddleware** (`internal/middleware/auth.go`):
```go
func AuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // 1. Get Authorization header
        authHeader := r.Header.Get("Authorization")
        
        // 2. Extract Bearer token
        tokenString := strings.TrimPrefix(authHeader, "Bearer ")
        
        // 3. Validate token with JWT library
        claims, err := authservice.ValidateToken(tokenString)
        if err != nil {
            // Return 401 Unauthorized
            return
        }
        
        // 4. Add claims to context
        ctx := context.WithValue(r.Context(), UserContextKey, claims)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

**AdminMiddleware** (`internal/middleware/auth.go`):
```go
func AdminMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // 1. Get claims from context (set by AuthMiddleware)
        claims := r.Context().Value(UserContextKey).(*authservice.Claims)
        
        // 2. Check if role is "admin"
        if claims.Role != "admin" {
            // Return 403 Forbidden
            return
        }
        
        // 3. Allow request to proceed
        next.ServeHTTP(w, r)
    })
}
```

---

## API Endpoints

### Base URL
```
https://api.example.com/api/admin
```

### Authentication Header (Required for all endpoints)
```
Authorization: Bearer {JWT_TOKEN}
```

---

### User Management

#### 1. Get All Users

```http
GET /api/admin/users
Authorization: Bearer {JWT_TOKEN}
```

**Response** (200 OK):
```json
[
  {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "John Doe",
    "email": "john@example.com",
    "role": "user",
    "created_at": "2024-11-01T10:00:00Z"
  },
  {
    "id": "550e8400-e29b-41d4-a716-446655440001",
    "name": "Jane Smith",
    "email": "jane@example.com",
    "role": "admin",
    "created_at": "2024-11-02T10:00:00Z"
  }
]
```

**Error Responses**:
- `401 Unauthorized` - Invalid or missing token
- `403 Forbidden` - User is not admin
- `500 Internal Server Error` - Database error

---

#### 2. Update User

```http
PUT /api/admin/users/{userId}
Authorization: Bearer {JWT_TOKEN}
Content-Type: application/json
```

**Request Body**:
```json
{
  "name": "Updated Name",
  "email": "newemail@example.com"
}
```

**Response** (200 OK):
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "Updated Name",
  "email": "newemail@example.com",
  "role": "user",
  "updated_at": "2024-11-10T15:30:00Z"
}
```

**Error Responses**:
- `404 Not Found` - User not found
- `400 Bad Request` - Invalid request body
- `401 Unauthorized` - Invalid token
- `403 Forbidden` - Not admin

---

#### 3. Delete User

```http
DELETE /api/admin/users/{userId}
Authorization: Bearer {JWT_TOKEN}
```

**Response** (200 OK):
```json
{
  "message": "User deleted successfully"
}
```

**Error Responses**:
- `404 Not Found` - User not found
- `401 Unauthorized` - Invalid token
- `403 Forbidden` - Not admin

---

#### 4. Update User Role

```http
PATCH /api/admin/users/update-role
Authorization: Bearer {JWT_TOKEN}
Content-Type: application/json
```

**Request Body**:
```json
{
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "role": "admin"
}
```

**Supported Roles**:
- `user` - Regular user
- `admin` - Administrator
- `moderator` - Moderator (if implemented)
- `chef` - Professional chef (if implemented)

**Response** (200 OK):
```json
{
  "message": "Role updated successfully"
}
```

**Error Responses**:
- `400 Bad Request` - Invalid request body
- `401 Unauthorized` - Invalid token
- `403 Forbidden` - Not admin
- `500 Internal Server Error` - Database error

---

### Order Management

#### 1. Get All Orders

```http
GET /api/admin/orders
Authorization: Bearer {JWT_TOKEN}
```

**Response** (200 OK):
```json
[
  {
    "id": "order-uuid-1",
    "user_id": "user-uuid-1",
    "items": [...],
    "total_price": 45.50,
    "status": "pending",
    "created_at": "2024-11-10T10:00:00Z"
  },
  {
    "id": "order-uuid-2",
    "user_id": "user-uuid-2",
    "items": [...],
    "total_price": 120.00,
    "status": "completed",
    "created_at": "2024-11-09T14:30:00Z"
  }
]
```

Orders are sorted by `created_at DESC` (newest first).

---

#### 2. Get Recent Orders

```http
GET /api/admin/orders/recent
Authorization: Bearer {JWT_TOKEN}
```

Returns the last 10 orders (same format as GET /orders but limited to 10).

---

#### 3. Update Order Status

```http
PUT /api/admin/orders/{orderId}/status
Authorization: Bearer {JWT_TOKEN}
Content-Type: application/json
```

**Request Body**:
```json
{
  "status": "completed"
}
```

**Supported Statuses**:
- `pending` - Awaiting processing
- `processing` - Being prepared
- `shipped` - Sent to customer
- `delivered` - Received by customer
- `cancelled` - Order cancelled
- `failed` - Payment failed

**Response** (200 OK):
```json
{
  "message": "Order status updated"
}
```

**Error Responses**:
- `400 Bad Request` - Invalid request body
- `401 Unauthorized` - Invalid token
- `403 Forbidden` - Not admin
- `500 Internal Server Error` - Database error

---

### Dashboard & Statistics

#### Get Admin Dashboard Stats

```http
GET /api/admin/stats
Authorization: Bearer {JWT_TOKEN}
```

**Response** (200 OK):
```json
{
  "totalUsers": 1250,
  "totalOrders": 4580
}
```

**Statistics Included**:
- `totalUsers` - Total count of registered users
- `totalOrders` - Total count of orders

---

## User Management

### User Model

```go
type User struct {
    ID        uuid.UUID
    Name      string
    Email     string
    Password  string
    Role      string    // "user", "admin", "moderator", etc.
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

### Roles

| Role | Permissions | Description |
|------|-------------|-------------|
| `user` | Read own data, create recipes | Regular user |
| `admin` | All endpoints, manage users & orders | Administrator |
| `moderator` | Moderate content (if implemented) | Content moderator |
| `chef` | Create recipes, sell on marketplace | Professional chef |

### User Operations

**View Users**: Admins can see all users with their details
**Update User**: Admins can modify name and email
**Delete User**: Admins can permanently delete users (use with caution!)
**Change Role**: Admins can promote/demote users

### Important Notes

⚠️ **Deleting a user is permanent** - All associated data may be deleted
⚠️ **Role changes are immediate** - User needs to re-login to update their permissions
⚠️ **Cannot delete self** - Admins should not be able to delete their own account (implement this check)

---

## Order Management

### Order Model

```go
type Order struct {
    ID         uuid.UUID
    UserID     uuid.UUID
    Items      []OrderItem    // Products in order
    TotalPrice float64
    Status     string         // pending, processing, shipped, etc.
    CreatedAt  time.Time
    UpdatedAt  time.Time
}
```

### Order Statuses

| Status | Description | Next Status |
|--------|-------------|------------|
| `pending` | Order received, awaiting processing | processing |
| `processing` | Order being prepared | shipped |
| `shipped` | Order sent to customer | delivered |
| `delivered` | Order received by customer | - |
| `cancelled` | Order cancelled by user/admin | - |
| `failed` | Payment failed | pending |

### Workflow Example

```
pending → processing → shipped → delivered ✓
   ↓
 cancelled (can cancel anytime before shipped)
   ↓
 failed → pending (can retry)
```

---

## Dashboard & Statistics

### Available Metrics

**User Statistics**:
- Total users registered
- New users (last 24h, 7d, 30d)
- Users by role
- Active users

**Order Statistics**:
- Total orders
- Orders by status
- Pending orders
- Revenue (sum of completed orders)

### Current Implementation

Currently returns:
- `totalUsers` - All users count
- `totalOrders` - All orders count

**Future Enhancements** (to implement):
- Revenue tracking
- User activity tracking
- Order completion rate
- Popular products
- Peak hours analysis

---

## Frontend Integration

### Login Flow

```javascript
// 1. Login with credentials
const loginRes = await fetch('/api/auth/login', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    email: 'admin@example.com',
    password: 'password123'
  })
});

const { token } = await loginRes.json();
localStorage.setItem('token', token);

// 2. Check if user is admin (parse JWT)
const payload = JSON.parse(atob(token.split('.')[1]));
if (payload.role !== 'admin') {
  // Redirect to regular user dashboard
}

// 3. Redirect to admin panel
window.location.href = '/admin/dashboard';
```

### Admin Dashboard Component (React)

```jsx
import { useEffect, useState } from 'react';

export function AdminDashboard() {
  const [stats, setStats] = useState(null);
  const [users, setUsers] = useState([]);
  const [orders, setOrders] = useState([]);
  const token = localStorage.getItem('token');

  // Fetch dashboard stats
  useEffect(() => {
    const fetchStats = async () => {
      const res = await fetch('/api/admin/stats', {
        headers: { 'Authorization': `Bearer ${token}` }
      });
      setStats(await res.json());
    };
    fetchStats();
  }, []);

  // Fetch all users
  useEffect(() => {
    const fetchUsers = async () => {
      const res = await fetch('/api/admin/users', {
        headers: { 'Authorization': `Bearer ${token}` }
      });
      setUsers(await res.json());
    };
    fetchUsers();
  }, []);

  return (
    <div className="admin-dashboard">
      <h1>Admin Panel</h1>
      <div className="stats">
        <div className="stat-card">
          <h3>Total Users</h3>
          <p>{stats?.totalUsers || 0}</p>
        </div>
        <div className="stat-card">
          <h3>Total Orders</h3>
          <p>{stats?.totalOrders || 0}</p>
        </div>
      </div>
      
      <div className="users-section">
        <h2>Users</h2>
        <table>
          <thead>
            <tr>
              <th>Name</th>
              <th>Email</th>
              <th>Role</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            {users.map(user => (
              <tr key={user.id}>
                <td>{user.name}</td>
                <td>{user.email}</td>
                <td>
                  <select 
                    value={user.role}
                    onChange={(e) => updateUserRole(user.id, e.target.value)}
                  >
                    <option value="user">User</option>
                    <option value="admin">Admin</option>
                  </select>
                </td>
                <td>
                  <button onClick={() => deleteUser(user.id)}>Delete</button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );

  async function updateUserRole(userId, role) {
    await fetch('/api/admin/users/update-role', {
      method: 'PATCH',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      },
      body: JSON.stringify({ user_id: userId, role })
    });
  }

  async function deleteUser(userId) {
    if (confirm('Are you sure?')) {
      await fetch(`/api/admin/users/${userId}`, {
        method: 'DELETE',
        headers: { 'Authorization': `Bearer ${token}` }
      });
    }
  }
}
```

### Admin Panel Features Checklist

- [ ] Dashboard with key metrics
- [ ] User management (list, edit, delete, change role)
- [ ] Order management (list, update status)
- [ ] Search and filter users
- [ ] Search and filter orders
- [ ] Pagination for large datasets
- [ ] Confirm dialogs for destructive actions
- [ ] Real-time stats updates
- [ ] Audit logs (who changed what)
- [ ] User activity history

---

## Security Considerations

### 1. Authentication

✅ JWT token-based authentication  
✅ Token validation on every request  
✅ Authorization header required  

**Best Practices**:
- Use HTTPS in production
- Set short token expiration (15-30 minutes)
- Implement token refresh mechanism
- Store tokens securely (not in localStorage for sensitive data)

### 2. Authorization

✅ Role-based access control (RBAC)  
✅ Admin middleware checks role  
✅ Separate endpoints for admin operations  

**Best Practices**:
- Never trust client-side role checks
- Always validate role on backend
- Implement granular permissions (not just admin/user)
- Log all admin actions

### 3. Data Protection

⚠️ **Missing**: Field-level permissions
⚠️ **Missing**: Data encryption at rest
⚠️ **Missing**: Audit logs

**Recommendations**:
- Hash sensitive data
- Don't return passwords in API responses
- Implement soft deletes (don't permanently delete data)
- Log all admin operations
- Implement rate limiting

### 4. Input Validation

❌ **Limited**: No validation on role field
❌ **Limited**: No validation on status field

**To Add**:
- Validate email format
- Validate role against allowed values
- Validate order status against allowed values
- Sanitize text inputs
- Check for SQL injection vulnerabilities

### 5. Common Vulnerabilities

**Privilege Escalation**: ⚠️ Risk
- User could change own role via API (if frontend sends it)
- Solution: Never trust user input for sensitive fields

**Unauthorized Access**: ✅ Protected
- AdminMiddleware ensures only admins access endpoints
- JWT validation prevents token tampering

**Data Exposure**: ⚠️ Risk
- Passwords could be exposed in GET /users response
- Solution: Don't include password hash in responses

---

## Deployment

### Prerequisites

- Backend running with all modules initialized
- Database with users table and role column
- JWT secret configured in environment

### Setup Steps

1. **Create Admin User**

```sql
-- Method 1: Direct database update
UPDATE users SET role = 'admin' WHERE email = 'admin@example.com';

-- Method 2: Using API (if another admin exists)
PATCH /api/admin/users/update-role
{
  "user_id": "user-uuid",
  "role": "admin"
}
```

2. **Configure Middleware**

Check that middleware is registered in routes:
```go
adminModule.RegisterRoutes(r, middleware.AuthMiddleware, middleware.AdminMiddleware)
```

3. **Test Endpoints**

```bash
# Get token
TOKEN=$(curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"..."}' | jq .token)

# Test admin endpoint
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/admin/stats
```

4. **Deploy to Production**

```bash
# 1. Build
go build -o bin/server ./cmd/server

# 2. Test locally
./bin/server

# 3. Push to production
git add .
git commit -m "✨ feat: Admin panel implementation"
git push origin main

# 4. Restart production server
# (depends on your deployment platform)
```

---

## Future Enhancements

### Phase 1: Core Analytics
- [ ] Revenue tracking
- [ ] User growth charts
- [ ] Order completion rates
- [ ] Popular products analysis

### Phase 2: Advanced Management
- [ ] User segments
- [ ] Marketing campaigns
- [ ] Email notifications to users
- [ ] Bulk user actions

### Phase 3: Content Management
- [ ] Manage recipes/marketplace items
- [ ] Content moderation
- [ ] Review management
- [ ] Category management

### Phase 4: Reporting
- [ ] PDF reports generation
- [ ] Scheduled email reports
- [ ] Custom analytics dashboards
- [ ] Data export functionality

### Phase 5: Security
- [ ] Two-factor authentication
- [ ] Audit logs
- [ ] IP whitelist
- [ ] Activity monitoring

---

## Troubleshooting

### Issue: "Admin access required" Error

**Cause**: User role is not 'admin'

**Solution**:
```sql
UPDATE users SET role = 'admin' WHERE id = 'user-uuid';
```

### Issue: "Unauthorized" Error

**Cause**: Invalid or missing JWT token

**Solution**:
1. Check Authorization header format: `Bearer {token}`
2. Verify token is not expired
3. Re-login to get new token

### Issue: Database Errors

**Cause**: User/Order not found

**Solution**:
1. Verify UUID format is correct
2. Check if user/order exists in database
3. Check database connection

### Issue: Middleware Not Applied

**Cause**: Routes not registered with middleware

**Solution**:
```go
// Ensure this is called in routes_modular.go
adminModule.RegisterRoutes(r, middleware.AuthMiddleware, middleware.AdminMiddleware)
```

---

## Support & Contact

For admin panel questions:
1. Check this documentation
2. Review code in `internal/modules/admin/`
3. Check middleware implementation in `internal/middleware/auth.go`
4. Review routes in `internal/app/routes_modular.go`

---

**Last Updated**: November 10, 2025  
**Maintained By**: Backend Team  
**Status**: Production Ready
