# Admin API Quick Reference

## 🔐 Authentication Required

**All endpoints require JWT token in Authorization header:**
```
Authorization: Bearer {JWT_TOKEN}
```

**User must have role = "admin"**

---

## 📊 Dashboard

### Get Statistics
```bash
GET /api/admin/stats
```
```json
{
  "totalUsers": 1250,
  "totalOrders": 4580
}
```

---

## 👥 Users Management

### List All Users
```bash
GET /api/admin/users
```

### Get User (by ID)
```bash
GET /api/admin/users/{userId}
```

### Update User
```bash
PUT /api/admin/users/{userId}
Content-Type: application/json

{
  "name": "New Name",
  "email": "newemail@example.com"
}
```

### Delete User
```bash
DELETE /api/admin/users/{userId}
```

### Change User Role
```bash
PATCH /api/admin/users/update-role
Content-Type: application/json

{
  "user_id": "uuid",
  "role": "admin"
}
```

**Available Roles**: `user`, `admin`, `moderator`, `chef`

---

## 📦 Orders Management

### List All Orders (sorted by date DESC)
```bash
GET /api/admin/orders
```

### Get Recent Orders (last 10)
```bash
GET /api/admin/orders/recent
```

### Update Order Status
```bash
PUT /api/admin/orders/{orderId}/status
Content-Type: application/json

{
  "status": "completed"
}
```

**Available Statuses**: `pending`, `processing`, `shipped`, `delivered`, `cancelled`, `failed`

---

## 🔑 Making Admin Requests

### cURL Example
```bash
TOKEN="your_jwt_token_here"

# Get stats
curl -H "Authorization: Bearer $TOKEN" \
  https://api.example.com/api/admin/stats

# List users
curl -H "Authorization: Bearer $TOKEN" \
  https://api.example.com/api/admin/users

# Update order status
curl -X PUT -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"status":"completed"}' \
  https://api.example.com/api/admin/orders/{orderId}/status
```

### JavaScript Example
```javascript
const token = localStorage.getItem('token');

// Get stats
const stats = await fetch('/api/admin/stats', {
  headers: { 'Authorization': `Bearer ${token}` }
}).then(r => r.json());

// Update user
await fetch('/api/admin/users/{userId}', {
  method: 'PUT',
  headers: {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${token}`
  },
  body: JSON.stringify({
    name: 'New Name',
    email: 'new@example.com'
  })
});
```

---

## ❌ Error Codes

| Code | Error | Meaning |
|------|-------|---------|
| 200 | OK | Success |
| 400 | Bad Request | Invalid input |
| 401 | Unauthorized | Invalid/missing token |
| 403 | Forbidden | Not admin role |
| 404 | Not Found | Resource not found |
| 500 | Server Error | Database error |

---

## 🛡️ Security Notes

✅ Always use HTTPS in production  
✅ Keep JWT token secret (don't commit to repo)  
✅ Set short token expiration (15-30 min)  
✅ Implement token refresh mechanism  
✅ Log all admin actions  

⚠️ Deleting users is permanent  
⚠️ Role changes are immediate  
⚠️ Never expose passwords in API responses  

---

## 📚 Full Documentation

See **ADMIN_PANEL_GUIDE.md** for complete implementation details.
