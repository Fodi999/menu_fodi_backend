# JWT `sub` Field Fix - COMPLETE ✅

## 🎯 Problem
JWT tokens were missing the `sub` (subject) field, causing:
- Frontend token validation failures
- RecipeContext clearing on every page load
- AI Assistant losing user context
- Recommendations not working

## 🔧 Solution Applied

### Backend Changes (Go)

**File:** `internal/modules/auth/service/jwt_service.go`

#### 1. Updated Claims Structure
```go
type Claims struct {
	Email   string `json:"email"`
	Role    string `json:"role"`
	HasRole bool   `json:"hasRole"`
	jwt.RegisteredClaims // Contains Subject field
}
```

#### 2. Token Generation (RFC 7519 Compliant)
```go
claims := &Claims{
	Email:   email,
	Role:    role,
	HasRole: true,
	RegisteredClaims: jwt.RegisteredClaims{
		Subject:   userID, // ✅ RFC 7519: sub field for user ID
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	},
}
```

#### 3. Global Replacement
Replaced all occurrences of `claims.UserID` → `claims.Subject`:
- `internal/middleware/auth.go`
- `internal/modules/user/service/user_service.go`
- `internal/modules/recipes/service/recipe_service.go`
- `internal/modules/recipes/service/match_service.go`
- `internal/modules/fridge/service/fridge_service.go`
- `internal/modules/notifications/service/notification_service.go`
- `internal/modules/admin/service/recipe_ai.go`

Total: **26 files updated**

## ✅ Validation

### New JWT Token (POST /api/auth/login)
```json
{
  "email": "fodi85@gmail.ru",
  "role": "home_chef",
  "hasRole": true,
  "sub": "407582be-59d5-4d21-873b-1a72d31b0d42",
  "exp": 1768685347,
  "iat": 1768598947
}
```

### Test Command
```bash
curl -X POST https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"fodi85@gmail.ru","password":"210185"}' | jq .

# Decode JWT payload:
# Take token from response, split by '.', decode middle part (base64)
```

## 📋 Frontend Action Required

**⚠️ Users MUST re-login or clear cache to get new tokens!**

### Option 1: Clear localStorage (Quick)
```javascript
localStorage.clear();
location.reload();
```

### Option 2: Re-login
1. Logout
2. Login again
3. Verify in console:
```javascript
// Should show:
{
  sub: "407582be-59d5-4d21-873b-1a72d31b0d42",
  email: "fodi85@gmail.ru",
  role: "home_chef",
  hasRole: true
}
```

## 🎉 Expected Results

After re-login:
- ✅ No more "Token missing 'sub'" errors
- ✅ No more "RecipeContext: Cleared localStorage"
- ✅ AI Assistant maintains user context
- ✅ Recommendations work correctly
- ✅ User-specific data persists
- ✅ RFC 7519 compliant JWT tokens

## 🔗 Related Changes

- **Commit:** `fix: add RFC 7519 compliant sub field to JWT tokens`
- **Date:** 2026-01-16
- **Files Changed:** 27
- **Lines Changed:** +85 / -85

## 📚 References

- [RFC 7519: JSON Web Token (JWT)](https://tools.ietf.org/html/rfc7519)
- [JWT Claims: sub (Subject)](https://tools.ietf.org/html/rfc7519#section-4.1.2)
- [golang-jwt/jwt v5 Documentation](https://github.com/golang-jwt/jwt)

## ✅ Status

- **Backend:** DEPLOYED ✅
- **Frontend:** REQUIRES RE-LOGIN ⏳
- **Testing:** VALIDATED ✅
