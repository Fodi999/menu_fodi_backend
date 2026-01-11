# 🚀 Quick Start: Enable Request ID Tracking

**Time:** 5 minutes  
**Impact:** Immediate improvement in debugging

---

## Step 1: Add Middleware to Router

**File:** `cmd/server/main.go`

**Find this section:**
```go
r := chi.NewRouter()
r.Use(middleware.Logger)
r.Use(middleware.CORS)
```

**Change to:**
```go
r := chi.NewRouter()

// IMPORTANT: RequestIDMiddleware must be FIRST
// This ensures all logs include request_id
r.Use(middleware.RequestIDMiddleware)

r.Use(middleware.Logger)
r.Use(middleware.CORS)
```

---

## Step 2: Verify It Works

### Test 1: Without Request ID
```bash
curl http://localhost:8080/health -v
```

**Expected Response Header:**
```
X-Request-ID: 9f4c7a8b-1234-5678-90ab-cdef12345678
```

### Test 2: With Custom Request ID
```bash
curl -H "X-Request-ID: my-test-123" http://localhost:8080/health -v
```

**Expected Response Header:**
```
X-Request-ID: my-test-123
```

---

## Step 3: Check Logs

**Run server:**
```bash
./bin/server
```

**Expected in logs:**
```
INFO    Incoming request    method=GET path=/health request_id=9f4c7a8b-1234...
```

---

## ✅ Done!

Request ID tracking is now enabled. Every request gets:
1. ✅ **Unique identifier** - UUID generated automatically
2. ✅ **Logged automatically** - All requests logged with request_id
3. ✅ **Returned in response header** - X-Request-ID header
4. ✅ **Returned in response body** - meta.request_id in JSON
5. ✅ **Available in context** - For handlers via logger.GetRequestID(ctx)

---

## 🚀 Advanced Features (Already Enabled)

### 1. Request ID in Response Body ✅
Every response includes request_id in meta:
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

**Benefits:**
- Frontend can log request_id to Sentry
- Support team can trace user issues
- End-to-end correlation (UI → API → DB)

---

### 2. Automatic Request ID in Logs ✅
Use `logger.WithContext(ctx)` to automatically include request_id:

**Old Way (manual):**
```go
logger.Log.Info("Processing request",
    zap.String("request_id", middleware.GetRequestID(r.Context())),
    zap.String("user_id", userID),
)
```

**New Way (automatic):**
```go
logger.WithContext(r.Context()).Info("Processing request",
    zap.String("user_id", userID),
)
// Output: INFO Processing request request_id=9f4c7a8b user_id=123
```

**Benefits:**
- Developers don't need to remember request_id
- Always consistent in logs
- Easy to grep/search

---

### 3. Propagate Request ID to Downstream Services ✅
When calling AI services or external APIs, pass request_id along:

**Example 1: Groq AI Call**
```go
import "github.com/dmitrijfomin/menu-fodifood/backend/internal/httputil"

func (s *AIService) GenerateRecipe(ctx context.Context, req Request) (*Recipe, error) {
    // Create HTTP request
    httpReq, _ := http.NewRequest("POST", "https://api.groq.com/v1/chat", body)
    
    // Automatically add X-Request-ID header
    httputil.AddRequestIDToHeaders(ctx, httpReq.Header)
    
    // Make request
    resp, err := s.client.Do(httpReq)
    // ...
}
```

**Example 2: Manual Propagation**
```go
requestID := httputil.PropagateRequestID(ctx)
req.Header.Set("X-Request-ID", requestID)
req.Header.Set("X-Correlation-ID", requestID) // Some services use this
```

**Benefits:**
- Track request across multiple services
- Debug AI timeouts or weird responses
- Full distributed tracing

---

## 📊 Complete Request Flow

```
Frontend                Backend                  AI Service
   |                       |                         |
   |-- POST /recipe ------>|                         |
   |   X-Request-ID:       |                         |
   |   abc-123            |                         |
   |                       |                         |
   |                       |-- Log: Processing       |
   |                       |   request_id=abc-123    |
   |                       |                         |
   |                       |-- POST /chat ---------->|
   |                       |   X-Request-ID:         |
   |                       |   abc-123               |
   |                       |                         |
   |                       |<-- Response ------------|
   |                       |                         |
   |<-- Response ----------|                         |
   |   X-Request-ID:       |                         |
   |   abc-123             |                         |
   |   {                   |                         |
   |     "data": {...},    |                         |
   |     "meta": {         |                         |
   |       "request_id":   |                         |
   |       "abc-123"       |                         |
   |     }                 |                         |
   |   }                   |                         |
```

Now you can trace the ENTIRE flow with one ID!

---

## ✅ Verification Checklist

Use this to confirm everything works:

- [ ] ✅ X-Request-ID in response header
- [ ] ✅ request_id in response body (meta.request_id)
- [ ] ✅ request_id in all logs
- [ ] ✅ Custom request_id not overwritten
- [ ] ✅ logger.WithContext(ctx) works
- [ ] ✅ httputil.AddRequestIDToHeaders works

**All should be ✅** - Test with:
```bash
# 1. Check header
curl -v http://localhost:8080/health

# 2. Check body
curl http://localhost:8080/health | jq '.meta.request_id'

# 3. Check logs
./bin/server # Look for: request_id=...

# 4. Custom ID preserved
curl -H "X-Request-ID: my-test-123" http://localhost:8080/health | jq '.meta.request_id'
# Should return: "my-test-123"
```

---

## 🎯 Next Priority: Migrate Admin Ingredients

**Why this endpoint first:**
- ✅ Most used by frontend (autocomplete)
- ✅ Previously caused 500 errors
- ✅ Perfect to test request_id in real conditions

**File to migrate:**
- `internal/modules/admin/transport/http/ingredients.go`
- Start with `SuggestIngredients()` handler

**See:** `docs/MIGRATION_EXAMPLES.md` - Example 2

---

**Next:** Start migrating handlers to use `models.SuccessResponse()` and `models.ErrorResponse()`

See: `docs/MIGRATION_EXAMPLES.md`
