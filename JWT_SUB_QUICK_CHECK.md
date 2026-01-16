# ✅ JWT `sub` Field - Quick Check

## 🔍 Verify Token Has `sub`

### 1. Login and Get Token
```bash
curl -X POST https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"fodi85@gmail.ru","password":"210185"}' | jq -r '.data.token'
```

### 2. Decode at jwt.io
- Copy token from step 1
- Go to https://jwt.io
- Paste token in "Encoded" section
- **Verify payload has:**
  ```json
  {
    "sub": "407582be-59d5-4d21-873b-1a72d31b0d42",  ✅
    "email": "fodi85@gmail.ru",                    ✅
    "role": "home_chef",                            ✅
    "hasRole": true,                                ✅
    "exp": 1768689689,                              ✅
    "iat": 1768603289                               ✅
  }
  ```

### 3. Browser Console Check
Open https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app and check logs:

**✅ Expected (CORRECT):**
```javascript
✅ [TokenValidator] Token is valid
🔍 [TokenValidator] JWT Payload: {
  sub: "407582be-59d5-4d21-873b-1a72d31b0d42",
  email: "fodi85@gmail.ru",
  role: "home_chef",
  hasRole: true,
  exp: "2026-01-17T..."
}
```

**❌ OLD (WRONG):**
```javascript
❌ Token missing 'sub'
🗑️ RecipeContext: Cleared localStorage
```

## 🎯 What Fixed

| Before | After |
|--------|-------|
| `"userId": "..."` | `"sub": "..."` |
| Custom field name | RFC 7519 standard |
| Frontend rejects | Frontend accepts |
| State resets | State persists |

## 📝 Files Changed

- `internal/modules/auth/service/jwt_service.go` - Claims structure
- All handlers: `claims.UserID` → `claims.Subject`

## 🚀 Deployment

Commit: `9b70a06`  
Status: ✅ Deployed to production  
Auto-deploy: Koyeb  

## 🧪 E2E Test

```bash
# Should work now without token errors
cd /Users/dmitrijfomin/Desktop/backend
make test-egg
```

Expected:
```
✅ ✅ ✅ ТЕСТ ПРОЙДЕН! ✅ ✅ ✅
```

---

**Quick Fix Summary:** JWT now has `sub` field = user ID (RFC 7519 standard)
