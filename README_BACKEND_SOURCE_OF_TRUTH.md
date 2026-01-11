# 🎯 Backend as Single Source of Truth - Complete Package

> **Date:** 2026-01-11  
> **Status:** ✅ Ready to implement  
> **Priority:** 🔥 CRITICAL

---

## 📦 What's Included

This package establishes backend as the single source of truth with:

1. **Unified Response Format** - All endpoints return same structure
2. **Error Code System** - Machine-readable error codes (DOMAIN_ERROR_TYPE)
3. **Request ID Tracking** - Track requests through entire flow
4. **Complete Documentation** - API contracts, migration guides, examples
5. **TypeScript Types** - Frontend integration types

---

## 🚀 Quick Start (5 minutes)

### Step 1: Add Middleware

**File:** `cmd/server/main.go`

```go
import "github.com/dmitrijfomin/menu-fodifood/backend/internal/middleware"

func main() {
    r := chi.NewRouter()
    
    // IMPORTANT: Add FIRST in middleware chain
    r.Use(middleware.RequestIDMiddleware)
    
    r.Use(middleware.Logger)
    r.Use(middleware.CORS)
    // ... rest
}
```

### Step 2: Test It Works

```bash
# Build and run
go build -o bin/server ./cmd/server
./bin/server

# Test in another terminal
curl -v http://localhost:8080/health
# Look for: X-Request-ID: 9f4c7a8b-1234-5678-90ab-cdef12345678
```

### Step 3: Migrate One Handler

See: `docs/MIGRATION_EXAMPLES.md` - Example 2 (GetIngredientByID)

**Time:** 10 minutes for first handler

---

## 📁 Files Created

### Core Implementation
- ✅ `internal/models/response.go` - Response helpers (SuccessResponse, ErrorResponse, etc.)
- ✅ `internal/models/errors.go` - Error code constants (80+ codes)
- ✅ `internal/middleware/request_id.go` - Request ID tracking middleware

### Documentation
- ✅ `docs/BACKEND_AS_SOURCE_OF_TRUTH.md` - Philosophy & principles
- ✅ `docs/MIGRATION_EXAMPLES.md` - 6 detailed migration examples
- ✅ `docs/API_TYPES_TYPESCRIPT.ts` - TypeScript types for frontend
- ✅ `API_CONTRACT_COMPLETE.md` - Updated API contracts (all endpoints)
- ✅ `IMPLEMENTATION_SUMMARY.md` - High-level overview
- ✅ `IMPLEMENTATION_CHECKLIST.md` - Task-by-task checklist
- ✅ `QUICKSTART_REQUEST_ID.md` - 5-minute setup guide
- ✅ `README_BACKEND_SOURCE_OF_TRUTH.md` - This file

---

## 🎯 Benefits

### For Backend Team
- ✅ Consistent code across all handlers
- ✅ Easy to add new endpoints (copy-paste pattern)
- ✅ Better logging with request_id
- ✅ Easier debugging in production

### For Frontend Team
- ✅ Every response has same structure
- ✅ Error codes tell exactly what went wrong
- ✅ Request ID for tracking bugs
- ✅ Can auto-generate TypeScript SDK

### For DevOps
- ✅ Request ID in logs = faster debugging
- ✅ Error codes = specific alerts
- ✅ Standard format = easy parsing
- ✅ Correlate Sentry errors with backend logs

---

## 📋 Response Format

### Success Response
```json
{
  "data": { ... },
  "error": null,
  "meta": {
    "request_id": "9f4c7a8b-1234-5678-90ab-cdef12345678",
    "timestamp": "2026-01-11T10:30:00Z",
    "version": "v1"
  }
}
```

### Error Response
```json
{
  "data": null,
  "error": {
    "code": "INGREDIENT_NOT_FOUND",
    "message": "Ingredient not found",
    "details": "Ingredient with ID 'abc123' does not exist"
  },
  "meta": {
    "request_id": "9f4c7a8b-1234-5678-90ab-cdef12345678",
    "timestamp": "2026-01-11T10:30:00Z"
  }
}
```

---

## 🔧 Usage in Handlers

### Before (Inconsistent)
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

### After (Unified)
```go
func (h *Handler) GetIngredient(w http.ResponseWriter, r *http.Request) {
    requestID := middleware.GetRequestID(r.Context())
    
    ingredient, err := h.service.GetByID(id)
    if err != nil {
        logger.Log.Error("Ingredient not found",
            zap.String("request_id", requestID),
            zap.String("id", id),
        )
        models.NotFoundError(w, r,
            models.ErrorIngredientNotFound,
            "Ingredient not found",
            fmt.Sprintf("ID: %s", id),
        )
        return
    }
    
    models.SuccessResponse(w, r, ingredient)
}
```

---

## 📊 Error Codes

### Format: `DOMAIN_ERROR_TYPE`

**Categories:**
- `AUTH_*` - Authentication (invalid token, credentials, etc.)
- `INGREDIENT_*` - Ingredients (not found, invalid input)
- `RECIPE_*` - Recipes (AI generation, validation)
- `FRIDGE_*` - Fridge (insufficient quantity, expired)
- `TOKEN_*` - Token economy (insufficient balance)
- `USER_*` - User management
- `MARKETPLACE_*` - Marketplace
- `ACADEMY_*` - Academy
- `GENERAL_*` - General errors (invalid JSON, database)
- `VALIDATION_*` - Validation errors
- `UPLOAD_*` - File upload errors

**See:** `internal/models/errors.go` for complete list (80+ codes)

---

## 📚 Documentation

### For Developers

1. **Start Here:** `IMPLEMENTATION_SUMMARY.md`
   - High-level overview
   - Quick examples
   - Team communication templates

2. **Migration Guide:** `docs/MIGRATION_EXAMPLES.md`
   - 6 detailed examples (auth, ingredients, recipes, fridge, etc.)
   - Before/after comparisons
   - Best practices

3. **Philosophy:** `docs/BACKEND_AS_SOURCE_OF_TRUTH.md`
   - Why unified responses?
   - Why error codes?
   - Why request ID?
   - Implementation plan

4. **Checklist:** `IMPLEMENTATION_CHECKLIST.md`
   - Task-by-task breakdown
   - Progress tracking
   - Success criteria

### For Frontend

1. **TypeScript Types:** `docs/API_TYPES_TYPESCRIPT.ts`
   - Complete type definitions
   - Helper functions (isSuccess, isError, unwrapResponse)
   - API client with request ID support
   - Error handling examples

2. **API Contracts:** `API_CONTRACT_COMPLETE.md`
   - All endpoints documented
   - Request/response examples
   - Error codes per endpoint
   - Localization support

---

## ⏱️ Implementation Timeline

### Day 1: Foundation (5 minutes) ✅
- [x] Create response models ✅
- [x] Create error codes ✅
- [x] Create request ID middleware ✅
- [ ] Add middleware to router
- [ ] Test middleware works

### Days 2-3: High-Priority Modules (2 days)
- [ ] Migrate Auth module (4 handlers)
- [ ] Migrate Admin Ingredients (5 handlers)
- [ ] Migrate Admin Recipes (4 handlers)

### Days 4-5: User-Facing Modules (2 days)
- [ ] Migrate Fridge module (7 handlers)
- [ ] Migrate Recipe Catalog (10 handlers)
- [ ] Migrate User module (2 handlers)

### Day 6: Supporting Modules (1 day)
- [ ] Migrate Admin Users (5 handlers)
- [ ] Migrate Token Economy (7 handlers)
- [ ] Migrate Marketplace (6 handlers)
- [ ] Migrate Academy (2 handlers)

### Day 7: Testing & Docs (1 day)
- [ ] Update tests
- [ ] Update documentation
- [ ] Frontend integration

### Day 8: Deployment (1 day)
- [ ] Local testing
- [ ] Production deployment
- [ ] Monitoring

**Total:** 7-8 days for complete migration

---

## ✅ Success Criteria

- [x] All code files created ✅
- [x] Code compiles successfully ✅
- [x] Documentation complete ✅
- [ ] Request ID middleware active
- [ ] High-priority handlers migrated
- [ ] All tests passing
- [ ] Frontend integrated
- [ ] Production deployment successful

**Current Status:** 20% complete (foundation ready, migration pending)

---

## 🎓 Team Communication

### For Backend Team

**Slack Message:**
```
🎯 New Backend Standard: Unified Response Format

We're implementing a unified response format across all endpoints:

✅ Every response: { data, error, meta: {request_id, timestamp} }
✅ Error codes: INGREDIENT_NOT_FOUND, AUTH_INVALID_TOKEN, etc.
✅ Request ID tracking for debugging

📚 Docs: /docs/BACKEND_AS_SOURCE_OF_TRUTH.md
🚀 Quick Start: /QUICKSTART_REQUEST_ID.md (5 min)
📝 Examples: /docs/MIGRATION_EXAMPLES.md

Questions? Check the docs first - they're comprehensive!
```

### For Frontend Team

**Slack Message:**
```
🔄 Backend API Response Format Update

Backend now returns unified format for all endpoints:

{
  "data": { ... },
  "error": { "code": "...", "message": "..." } | null,
  "meta": { "request_id": "...", "timestamp": "..." }
}

✅ TypeScript types: /docs/API_TYPES_TYPESCRIPT.ts
✅ Error codes for specific handling (AUTH_INVALID_TOKEN → redirect to login)
✅ Request ID for bug tracking

📚 API Contracts: /API_CONTRACT_COMPLETE.md
```

---

## 🔍 Debugging with Request ID

### Frontend Error Report
```typescript
console.error(`[${error.meta.request_id}] ${error.error.code}: ${error.error.message}`);

// Send to Sentry
Sentry.captureException(error, {
  extra: {
    requestId: error.meta.request_id,
    errorCode: error.error.code,
  }
});
```

### Find in Backend Logs
```bash
# On Koyeb or local
grep "request_id=9f4c7a8b" server.log

# Output:
# INFO  Incoming request  request_id=9f4c7a8b method=GET path=/api/ingredients/123
# ERROR Ingredient not found  request_id=9f4c7a8b ingredient_id=123
```

### Correlate with Database
```sql
-- If you log to database
SELECT * FROM request_logs WHERE request_id = '9f4c7a8b-1234-5678-90ab-cdef12345678';
```

---

## 📞 Support

### Resources
1. **Philosophy & Why:** `docs/BACKEND_AS_SOURCE_OF_TRUTH.md`
2. **How to Migrate:** `docs/MIGRATION_EXAMPLES.md`
3. **Quick Setup:** `QUICKSTART_REQUEST_ID.md`
4. **Task Checklist:** `IMPLEMENTATION_CHECKLIST.md`
5. **API Reference:** `API_CONTRACT_COMPLETE.md`

### Questions?
- Check migration examples first (they cover 90% of cases)
- Look at similar handlers in codebase
- Refer to error codes in `internal/models/errors.go`

---

## 🏆 Best Practices

### DO:
✅ Use `models.SuccessResponse()` for all success responses  
✅ Use `models.ErrorResponse()` for all error responses  
✅ Include request_id in all logs: `middleware.GetRequestID(r.Context())`  
✅ Use specific error codes: `models.ErrorIngredientNotFound`  
✅ Provide helpful error details for users  

### DON'T:
❌ Mix response formats (always use unified)  
❌ Return raw errors: `http.Error(w, err.Error(), 500)`  
❌ Use generic error codes when specific ones exist  
❌ Forget to log with request_id  
❌ Skip error details (they help debugging)  

---

## 🔥 Priority Order

1. **Add Middleware** (5 min) - Foundation for everything
2. **Migrate Auth** (1 day) - All authenticated requests depend on this
3. **Migrate Admin Ingredients** (1 day) - Frontend admin panel depends on this
4. **Migrate Admin Recipes** (1 day) - AI feature (recently fixed)
5. **Everything else** (4 days) - Lower priority but important

**Start NOW:** Add RequestIDMiddleware to `cmd/server/main.go`

---

**Last Updated:** 2026-01-11  
**Status:** ✅ Ready to start  
**Next Action:** See `QUICKSTART_REQUEST_ID.md`  
**Priority:** 🔥 CRITICAL
