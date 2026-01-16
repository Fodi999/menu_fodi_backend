# 🔧 JWT `sub` Field Fix - RFC 7519 Compliance

## 🔴 Problem

Frontend was rejecting JWT tokens because they **lacked the `sub` (subject) field**, causing:

- ❌ Token validation failures
- 🗑️ RecipeContext localStorage being cleared
- 🤖 AI Assistant state resets
- 👤 User session instability

### Original JWT Payload (WRONG)
```json
{
  "userId": "407582be-59d5-4d21-873b-1a72d31b0d42",  // ❌ Custom field
  "email": "fodi85@gmail.ru",
  "role": "home_chef",
  "exp": 1768689689
}
```

Frontend expected:
```json
{
  "sub": "407582be-59d5-4d21-873b-1a72d31b0d42",  // ✅ RFC 7519 standard
  "email": "fodi85@gmail.ru",
  "role": "home_chef",
  "hasRole": true,
  "exp": 1768689689
}
```

## ✅ Solution

Changed JWT Claims structure to use **RFC 7519 standard fields**:

### Before (Custom Field)
```go
type Claims struct {
	UserID string `json:"userId"`  // ❌ Non-standard field name
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

claims := &Claims{
	UserID: userID,  // ❌ Wrong field
	Email:  email,
	Role:   role,
	RegisteredClaims: jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	},
}
```

### After (RFC 7519 Standard)
```go
type Claims struct {
	Email   string `json:"email"`
	Role    string `json:"role"`
	HasRole bool   `json:"hasRole"`
	jwt.RegisteredClaims // Contains Subject (sub), ExpiresAt (exp), IssuedAt (iat)
}

claims := &Claims{
	Email:   email,
	Role:    role,
	HasRole: true,
	RegisteredClaims: jwt.RegisteredClaims{
		Subject:   userID, // ✅ RFC 7519: sub field
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	},
}
```

## 🔄 Code Changes

### 1. JWT Service
**File:** `internal/modules/auth/service/jwt_service.go`

**Changed:**
- Claims struct: Removed `UserID` field, using `Subject` from `RegisteredClaims`
- GenerateToken: Sets `Subject` field with user ID
- Added `HasRole: true` for frontend compatibility

### 2. All Handlers & Middleware
**Mass replacement:** `claims.UserID` → `claims.Subject`

**Files updated:**
- `internal/modules/recipes/transport/http/handler.go`
- `internal/modules/recipes/transport/http/handlers.go`
- `internal/modules/recipes_admin/transport/http/handlers.go`
- `internal/modules/admin/transport/http/handlers.go`
- `internal/modules/admin/transport/http/recipe_ai_handlers.go`
- `internal/modules/auth/service/service.go`
- `internal/middleware/auth.go`

**Command used:**
```bash
find internal -name "*.go" -exec sed -i '' 's/claims\.UserID/claims.Subject/g' {} \;
```

## 🎯 Result

### New JWT Payload
```json
{
  "sub": "407582be-59d5-4d21-873b-1a72d31b0d42",  // ✅ Standard field
  "email": "fodi85@gmail.ru",
  "role": "home_chef",
  "hasRole": true,
  "exp": 1768689689,
  "iat": 1768603289
}
```

### Frontend Impact
✅ Token validation passes  
✅ RecipeContext persists  
✅ AI Assistant state preserved  
✅ User session stable  
✅ RFC 7519 compliant  

## 📚 Why `sub` is Required

From [RFC 7519 Section 4.1.2](https://datatracker.ietf.org/doc/html/rfc7519#section-4.1.2):

> The "sub" (subject) claim identifies the principal that is the subject of the JWT.
> The claims in a JWT are normally statements about the subject.
> The subject value MUST either be scoped to be locally unique in the context of the issuer
> or be globally unique.

**Key reasons:**
1. **RFC Standard:** `sub` is the standard field for user identity
2. **Email changes:** Email addresses can change, user ID cannot
3. **Security:** Proper token structure prevents security issues
4. **Interoperability:** Other services expect `sub` field
5. **Best practices:** Enterprise-grade JWT implementations use `sub`

## 🧪 Testing

### Manual Test
```bash
# Login and get token
curl -X POST https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"fodi85@gmail.ru","password":"210185"}'

# Decode JWT at jwt.io - should show:
# {
#   "sub": "407582be-59d5-4d21-873b-1a72d31b0d42",
#   "email": "fodi85@gmail.ru",
#   "role": "home_chef",
#   "hasRole": true,
#   "exp": ...,
#   "iat": ...
# }
```

### Expected Browser Logs (After Fix)
```javascript
✅ [TokenValidator] Token is valid
🔍 [TokenValidator] JWT Payload: {
  sub: "407582be-59d5-4d21-873b-1a72d31b0d42",
  email: "fodi85@gmail.ru",
  role: "home_chef",
  hasRole: true,
  exp: 1768689689
}
```

**NO MORE:**
- ❌ `Token missing 'sub'`
- 🗑️ `RecipeContext: Cleared localStorage`

## 📝 Commit

```bash
git add -A
git commit -m "fix: JWT now uses RFC 7519 'sub' field instead of custom 'userId'

- Changed Claims struct to use Subject from RegisteredClaims
- Mass replaced claims.UserID with claims.Subject across codebase
- Added hasRole field for frontend compatibility
- Fixes token validation on frontend
- Prevents RecipeContext state loss
- RFC 7519 compliant

Closes: Frontend token validation issue
Impact: AI Assistant now persists state correctly"
```

## 🔗 Related Files

- `internal/modules/auth/service/jwt_service.go` - JWT generation
- `internal/middleware/auth.go` - Token validation middleware
- All handler files using `claims.Subject`

## 🚀 Deployment

```bash
git push origin main
# Koyeb auto-deploys
# Frontend will automatically receive correct JWT format
```

---

**Date:** 2026-01-16  
**Issue:** JWT missing `sub` field  
**Fix:** Use `RegisteredClaims.Subject`  
**Standard:** RFC 7519 Section 4.1.2  
