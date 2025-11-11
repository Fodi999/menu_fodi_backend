# ⚠️ Frontend API Configuration Fix

## Problem
Frontend получает `404 Not Found` при попытке входа:
```
Failed to load resource: yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/login:1 404
```

## Root Cause
Frontend пытается обратиться к `/api/login`, но correct endpoint это `/api/auth/login`

## Solution

### Backend API Endpoints
```
✅ CORRECT endpoints:
POST /api/auth/login      ← Login
POST /api/auth/register   ← Registration
GET /api/user/profile     ← User Profile
GET /api/admin/profile    ← Admin Profile
GET /api/admin/stats      ← Admin Stats
```

### Frontend Configuration

**File:** `src/lib/api.ts` или похожий файл с API конфигурацией

**BEFORE (НЕПРАВИЛЬНО):**
```typescript
const API_BASE = 'https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app';

export async function login(email: string, password: string) {
  const res = await fetch(`${API_BASE}/api/login`, {  // ❌ НЕПРАВИЛЬНО
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password })
  });
  // ...
}
```

**AFTER (ПРАВИЛЬНО):**
```typescript
const API_BASE = 'https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app';

export async function login(email: string, password: string) {
  const res = await fetch(`${API_BASE}/api/auth/login`, {  // ✅ ПРАВИЛЬНО
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password })
  });
  // ...
}
```

### All Frontend API Routes That Need Fixing

| Frontend Function | Current (❌) | Correct (✅) |
|------------------|-------------|-----------|
| login() | `/api/login` | `/api/auth/login` |
| register() | `/api/register` | `/api/auth/register` |
| getProfile() | `/api/profile` | `/api/user/profile` (user) or `/api/admin/profile` (admin) |
| getAdminStats() | `/api/stats` | `/api/admin/stats` |
| getUsers() | `/api/users` | `/api/admin/users` |

### AuthContext Fix

**File:** `src/context/AuthContext.tsx` (или UserContext.tsx)

```typescript
// BEFORE ❌
async function login(email: string, password: string) {
  const res = await fetch('/api/login', {  // ❌ НЕПРАВИЛЬНО
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password })
  });
  // ...
}

// AFTER ✅
async function login(email: string, password: string) {
  const res = await fetch('/api/auth/login', {  // ✅ ПРАВИЛЬНО
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password })
  });
  // ...
}
```

### checkAuth() Fix

```typescript
// BEFORE ❌
async function checkAuth() {
  const token = localStorage.getItem("token");
  const res = await fetch('/api/profile', {  // ❌ НЕПРАВИЛЬНО
    headers: { Authorization: `Bearer ${token}` }
  });
  // ...
}

// AFTER ✅
async function checkAuth() {
  const token = localStorage.getItem("token");
  const role = localStorage.getItem("role");
  
  // Select correct profile endpoint based on role
  const profileUrl = role === "admin" 
    ? "/api/admin/profile" 
    : "/api/user/profile";
  
  const res = await fetch(profileUrl, {  // ✅ ПРАВИЛЬНО
    headers: { Authorization: `Bearer ${token}` }
  });
  // ...
}
```

## Quick Test

After fixing, test with curl:

```bash
# Test login
curl -X POST https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"admin_password_123"}' | jq '.'

# Expected response:
{
  "token": "eyJhbGc...",
  "role": "admin",
  "message": "Login successful"
}
```

## Checklist

- [ ] Update `/api/login` → `/api/auth/login`
- [ ] Update `/api/register` → `/api/auth/register`
- [ ] Update `/api/profile` → `/api/user/profile` or `/api/admin/profile`
- [ ] Update `/api/stats` → `/api/admin/stats`
- [ ] Update `/api/users` → `/api/admin/users`
- [ ] Test login with admin account
- [ ] Test user profile fetch
- [ ] Test admin profile fetch

## Backend Ready ✅

All endpoints are working on the backend:
- ✅ POST /api/auth/login
- ✅ POST /api/auth/register
- ✅ GET /api/user/profile
- ✅ GET /api/admin/profile
- ✅ GET /api/admin/stats
- ✅ GET /api/admin/users
- ✅ All other admin endpoints

Just update frontend API paths!
