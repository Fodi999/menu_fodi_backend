# 🎯 Backend as Single Source of Truth

> **Status:** ✅ Implementation Guide  
> **Date:** 2026-01-11  
> **Purpose:** Establish backend as the single source of truth with unified response format

---

## 📜 Core Principles

### ❌ Frontend MUST NOT:
- Know anything about database structure
- Guess API behavior or response formats
- Implement business logic (validation, authorization, localization)
- Make assumptions about data shape

### ✅ Backend MUST:
- Be the single source of truth for ALL business logic
- Validate all inputs
- Handle all localization (Accept-Language → backend decides)
- Enforce all authorization rules (JWT → backend validates)
- Return unified, predictable response format

---

## 🔧 Implementation Plan

### Phase 1: Unified Response Format ✅
**Status:** Ready to implement  
**Files to change:** `internal/models/response.go` (create), all handlers

### Phase 2: Request ID Middleware ✅
**Status:** Ready to implement  
**Files to change:** `internal/middleware/request_id.go` (create), `cmd/server/main.go`

### Phase 3: Error Code System ✅
**Status:** Ready to implement  
**Files to change:** `internal/models/errors.go` (create), all error handling

---

## 📋 Unified Response Format

### Standard Response Structure

```go
type APIResponse struct {
    Data      interface{}    `json:"data"`
    Error     *APIError      `json:"error"`
    Meta      *ResponseMeta  `json:"meta"`
}

type APIError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Details string `json:"details,omitempty"`
}

type ResponseMeta struct {
    RequestID string `json:"request_id"`
    Timestamp string `json:"timestamp"`
    Version   string `json:"version,omitempty"`
}
```

### Examples

#### ✅ Success Response
```json
{
  "data": {
    "id": "uuid",
    "name": "Tomato",
    "category": "vegetable"
  },
  "error": null,
  "meta": {
    "request_id": "9f4c7a8b-1234-5678-90ab-cdef12345678",
    "timestamp": "2026-01-11T10:30:00Z",
    "version": "v1"
  }
}
```

#### ❌ Error Response
```json
{
  "data": null,
  "error": {
    "code": "INGREDIENT_NOT_FOUND",
    "message": "Ingredient with ID 'abc123' not found",
    "details": "Please check the ingredient ID and try again"
  },
  "meta": {
    "request_id": "9f4c7a8b-1234-5678-90ab-cdef12345678",
    "timestamp": "2026-01-11T10:30:00Z"
  }
}
```

#### 📄 Paginated Response
```json
{
  "data": {
    "items": [...],
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 100,
      "total_pages": 5
    }
  },
  "error": null,
  "meta": {
    "request_id": "9f4c7a8b-1234-5678-90ab-cdef12345678",
    "timestamp": "2026-01-11T10:30:00Z"
  }
}
```

---

## 🔍 Error Codes System

### Error Code Structure: `DOMAIN_ERROR_TYPE`

### Authentication Errors (AUTH_*)
- `AUTH_INVALID_TOKEN` - Invalid or expired JWT token
- `AUTH_MISSING_TOKEN` - Authorization header missing
- `AUTH_INVALID_CREDENTIALS` - Wrong email/password
- `AUTH_USER_EXISTS` - User already registered
- `AUTH_INSUFFICIENT_PERMISSIONS` - User lacks required role

### Ingredient Errors (INGREDIENT_*)
- `INGREDIENT_NOT_FOUND` - Ingredient not found
- `INGREDIENT_INVALID_INPUT` - Invalid ingredient data
- `INGREDIENT_ALREADY_EXISTS` - Ingredient already exists

### Recipe Errors (RECIPE_*)
- `RECIPE_NOT_FOUND` - Recipe not found
- `RECIPE_INVALID_INPUT` - Invalid recipe data
- `RECIPE_AI_GENERATION_FAILED` - AI recipe generation failed
- `RECIPE_VALIDATION_FAILED` - Recipe validation failed

### Fridge Errors (FRIDGE_*)
- `FRIDGE_ITEM_NOT_FOUND` - Fridge item not found
- `FRIDGE_INSUFFICIENT_QUANTITY` - Not enough quantity in fridge
- `FRIDGE_INVALID_INPUT` - Invalid fridge item data

### Token Economy Errors (TOKEN_*)
- `TOKEN_INSUFFICIENT_BALANCE` - Not enough tokens
- `TOKEN_INVALID_AMOUNT` - Invalid token amount
- `TOKEN_TRANSACTION_FAILED` - Token transaction failed

### General Errors (GENERAL_*)
- `GENERAL_INVALID_INPUT` - Invalid input data
- `GENERAL_INTERNAL_ERROR` - Internal server error
- `GENERAL_NOT_FOUND` - Resource not found
- `GENERAL_DATABASE_ERROR` - Database operation failed

---

## 🆔 Request ID Middleware

### Implementation

```go
package middleware

import (
    "context"
    "net/http"
    "github.com/google/uuid"
)

type contextKey string

const RequestIDKey contextKey = "request_id"

func RequestIDMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Check if client sent X-Request-ID
        requestID := r.Header.Get("X-Request-ID")
        
        // If not, generate one
        if requestID == "" {
            requestID = uuid.New().String()
        }
        
        // Add to context
        ctx := context.WithValue(r.Context(), RequestIDKey, requestID)
        
        // Add to response header
        w.Header().Set("X-Request-ID", requestID)
        
        // Log the request
        logger.Info("Incoming request",
            "method", r.Method,
            "path", r.URL.Path,
            "request_id", requestID,
        )
        
        // Call next handler
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

// GetRequestID retrieves request ID from context
func GetRequestID(ctx context.Context) string {
    if id, ok := ctx.Value(RequestIDKey).(string); ok {
        return id
    }
    return "unknown"
}
```

### Usage in Handlers

```go
func (h *Handler) GetIngredient(w http.ResponseWriter, r *http.Request) {
    requestID := middleware.GetRequestID(r.Context())
    
    logger.Info("Processing ingredient request",
        "request_id", requestID,
        "ingredient_id", ingredientID,
    )
    
    // ... handler logic
}
```

---

## 🎯 Migration Strategy

### Step 1: Create Response Models (Day 1)
1. Create `internal/models/response.go`
2. Define `APIResponse`, `APIError`, `ResponseMeta`
3. Create helper functions: `SuccessResponse()`, `ErrorResponse()`

### Step 2: Add Request ID Middleware (Day 1)
1. Create `internal/middleware/request_id.go`
2. Add middleware to router in `cmd/server/main.go`
3. Update logger to include request_id in all logs

### Step 3: Create Error Codes (Day 2)
1. Create `internal/models/errors.go`
2. Define all error codes as constants
3. Create `NewAPIError()` helper

### Step 4: Migrate Handlers (Day 3-5)
1. Start with high-priority modules: `auth`, `ingredients`, `recipes`
2. Replace all `json.NewEncoder().Encode()` with `SuccessResponse()`
3. Replace all error responses with `ErrorResponse()`
4. Test each module thoroughly

### Step 5: Update Documentation (Day 6)
1. Update `API_CONTRACT_COMPLETE.md` with new format
2. Add error codes to each endpoint
3. Add request_id examples

---

## 📊 Before vs After

### ❌ BEFORE (Inconsistent)

**Auth Handler:**
```go
json.NewEncoder(w).Encode(map[string]interface{}{
    "success": true,
    "data": map[string]interface{}{
        "token": token,
        "user": user,
    },
})
```

**Ingredient Handler:**
```go
json.NewEncoder(w).Encode(ingredient)
```

**Error Handler:**
```go
http.Error(w, "Ingredient not found", http.StatusNotFound)
```

### ✅ AFTER (Consistent)

**All Handlers:**
```go
// Success
models.SuccessResponse(w, r, ingredient)

// Error
models.ErrorResponse(w, r, 
    http.StatusNotFound,
    "INGREDIENT_NOT_FOUND",
    "Ingredient not found",
    "Please check the ingredient ID",
)
```

---

## 🚀 Implementation Code

### File: `internal/models/response.go`

```go
package models

import (
    "encoding/json"
    "net/http"
    "time"
    "context"
)

// APIResponse is the standard response format for all endpoints
type APIResponse struct {
    Data  interface{}    `json:"data"`
    Error *APIError      `json:"error"`
    Meta  *ResponseMeta  `json:"meta"`
}

// APIError represents an error in the API
type APIError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Details string `json:"details,omitempty"`
}

// ResponseMeta contains metadata about the response
type ResponseMeta struct {
    RequestID string `json:"request_id"`
    Timestamp string `json:"timestamp"`
    Version   string `json:"version,omitempty"`
}

// SuccessResponse sends a successful response
func SuccessResponse(w http.ResponseWriter, r *http.Request, data interface{}) {
    response := APIResponse{
        Data:  data,
        Error: nil,
        Meta:  buildMeta(r.Context()),
    }
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(response)
}

// SuccessResponseWithStatus sends a successful response with custom status
func SuccessResponseWithStatus(w http.ResponseWriter, r *http.Request, status int, data interface{}) {
    response := APIResponse{
        Data:  data,
        Error: nil,
        Meta:  buildMeta(r.Context()),
    }
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(response)
}

// ErrorResponse sends an error response
func ErrorResponse(w http.ResponseWriter, r *http.Request, status int, code, message, details string) {
    response := APIResponse{
        Data: nil,
        Error: &APIError{
            Code:    code,
            Message: message,
            Details: details,
        },
        Meta: buildMeta(r.Context()),
    }
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(response)
}

// buildMeta creates response metadata
func buildMeta(ctx context.Context) *ResponseMeta {
    requestID := "unknown"
    if id := ctx.Value("request_id"); id != nil {
        if idStr, ok := id.(string); ok {
            requestID = idStr
        }
    }
    
    return &ResponseMeta{
        RequestID: requestID,
        Timestamp: time.Now().UTC().Format(time.RFC3339),
        Version:   "v1",
    }
}
```

### File: `internal/models/errors.go`

```go
package models

// Error codes for the API
const (
    // Authentication errors
    ErrorAuthInvalidToken          = "AUTH_INVALID_TOKEN"
    ErrorAuthMissingToken          = "AUTH_MISSING_TOKEN"
    ErrorAuthInvalidCredentials    = "AUTH_INVALID_CREDENTIALS"
    ErrorAuthUserExists            = "AUTH_USER_EXISTS"
    ErrorAuthInsufficientPermissions = "AUTH_INSUFFICIENT_PERMISSIONS"
    
    // Ingredient errors
    ErrorIngredientNotFound      = "INGREDIENT_NOT_FOUND"
    ErrorIngredientInvalidInput  = "INGREDIENT_INVALID_INPUT"
    ErrorIngredientAlreadyExists = "INGREDIENT_ALREADY_EXISTS"
    
    // Recipe errors
    ErrorRecipeNotFound          = "RECIPE_NOT_FOUND"
    ErrorRecipeInvalidInput      = "RECIPE_INVALID_INPUT"
    ErrorRecipeAIGenerationFailed = "RECIPE_AI_GENERATION_FAILED"
    ErrorRecipeValidationFailed  = "RECIPE_VALIDATION_FAILED"
    
    // Fridge errors
    ErrorFridgeItemNotFound        = "FRIDGE_ITEM_NOT_FOUND"
    ErrorFridgeInsufficientQuantity = "FRIDGE_INSUFFICIENT_QUANTITY"
    ErrorFridgeInvalidInput        = "FRIDGE_INVALID_INPUT"
    
    // Token economy errors
    ErrorTokenInsufficientBalance = "TOKEN_INSUFFICIENT_BALANCE"
    ErrorTokenInvalidAmount       = "TOKEN_INVALID_AMOUNT"
    ErrorTokenTransactionFailed   = "TOKEN_TRANSACTION_FAILED"
    
    // General errors
    ErrorGeneralInvalidInput   = "GENERAL_INVALID_INPUT"
    ErrorGeneralInternalError  = "GENERAL_INTERNAL_ERROR"
    ErrorGeneralNotFound       = "GENERAL_NOT_FOUND"
    ErrorGeneralDatabaseError  = "GENERAL_DATABASE_ERROR"
)
```

---

## 🎓 Benefits

### For Frontend Developers
✅ **Predictable API**: Every response has the same structure  
✅ **Easy Debugging**: Request ID tracks the entire flow  
✅ **Clear Errors**: Error codes tell exactly what went wrong  
✅ **Auto-Generation**: Can generate TypeScript SDK automatically

### For Backend Developers
✅ **Consistent Code**: All handlers use same response pattern  
✅ **Easy Logging**: Request ID in every log message  
✅ **Better Monitoring**: Sentry/error tracking works perfectly  
✅ **Easy Testing**: Predictable responses = easy tests

### For DevOps
✅ **Faster Debugging**: Request ID traces across logs  
✅ **Better Alerts**: Error codes trigger specific alerts  
✅ **Easy Monitoring**: Standard format = easy parsing

---

## 📚 Next Steps

1. **Implement Response Models** ✅ (Code ready above)
2. **Add Request ID Middleware** ✅ (Code ready above)
3. **Migrate Auth Module** (Start here - highest priority)
4. **Migrate Admin Module** (Ingredients, Recipes)
5. **Migrate User Modules** (Fridge, Recipes, Wallet)
6. **Update Tests** (Expect new response format)
7. **Update Frontend** (Parse new response format)

---

## 🔥 Quick Win: Migrate One Handler

### Before:
```go
func (h *Handler) GetIngredient(w http.ResponseWriter, r *http.Request) {
    ingredient, err := h.service.GetByID(id)
    if err != nil {
        http.Error(w, "Not found", 404)
        return
    }
    json.NewEncoder(w).Encode(ingredient)
}
```

### After:
```go
func (h *Handler) GetIngredient(w http.ResponseWriter, r *http.Request) {
    ingredient, err := h.service.GetByID(id)
    if err != nil {
        models.ErrorResponse(w, r,
            http.StatusNotFound,
            models.ErrorIngredientNotFound,
            "Ingredient not found",
            fmt.Sprintf("Ingredient with ID '%s' does not exist", id),
        )
        return
    }
    models.SuccessResponse(w, r, ingredient)
}
```

**Result:** ✅ Consistent, ✅ Debuggable, ✅ Professional

---

**Status:** 📋 Ready for Implementation  
**Estimated Time:** 3-5 days  
**Priority:** 🔥 HIGH - Foundation for all future development
