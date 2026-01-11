# 🎯 Request ID - Advanced Usage Guide

> **Status:** ✅ All features implemented  
> **Date:** 2026-01-11

---

## 📋 Three Pillars of Request ID System

### 1. ✅ Request ID in Response Body (meta)
### 2. ✅ Automatic Request ID in Logs
### 3. ✅ Propagate Request ID to Downstream Services

---

## 1️⃣ Request ID in Response Body

### What It Does
Every response automatically includes `request_id` in the `meta` field:

```json
{
  "data": { "name": "Tomato" },
  "error": null,
  "meta": {
    "request_id": "9f4c7a8b-1234-5678-90ab-cdef12345678",
    "timestamp": "2026-01-11T10:30:00Z",
    "version": "v1"
  }
}
```

### Why It Matters
- **Frontend Logging:** Send request_id to Sentry/error tracking
- **User Support:** User reports error → support finds exact request in logs
- **End-to-End Tracing:** Track request from UI through API to database

### Frontend Integration (TypeScript)

```typescript
// Automatic error tracking
try {
  const response = await fetch('/api/ingredients/123');
  const json = await response.json();
  
  if (json.error) {
    // Send to Sentry with request_id
    Sentry.captureException(new Error(json.error.message), {
      extra: {
        requestId: json.meta.request_id,
        errorCode: json.error.code,
        details: json.error.details,
      },
      tags: {
        endpoint: '/api/ingredients/123',
      }
    });
  }
} catch (error) {
  console.error('Network error:', error);
}
```

### Support Team Use Case

**User:** "I got an error when searching for ingredients at 10:30 AM"

**Support:**
1. User provides request_id from error message (frontend shows it)
2. Search backend logs: `grep "request_id=9f4c7a8b" server.log`
3. See exact error, context, database queries, AI calls - everything

---

## 2️⃣ Automatic Request ID in Logs

### The Problem (Before)
Developers had to manually add request_id to every log:

```go
// ❌ TEDIOUS - Easy to forget
func (h *Handler) GetIngredient(w http.ResponseWriter, r *http.Request) {
    requestID := middleware.GetRequestID(r.Context())
    
    logger.Log.Info("Fetching ingredient",
        zap.String("request_id", requestID),  // Manual!
        zap.String("ingredient_id", id),
    )
    
    ingredient, err := h.service.GetByID(id)
    if err != nil {
        logger.Log.Error("Failed to fetch",
            zap.String("request_id", requestID),  // Manual again!
            zap.Error(err),
        )
        return
    }
    
    logger.Log.Info("Ingredient fetched",
        zap.String("request_id", requestID),  // Manual third time!
        zap.String("name", ingredient.Name),
    )
}
```

### The Solution (Now)
Use `logger.WithContext(ctx)` - request_id added automatically:

```go
// ✅ AUTOMATIC - Never forget
func (h *Handler) GetIngredient(w http.ResponseWriter, r *http.Request) {
    log := logger.WithContext(r.Context())  // Get logger with request_id
    
    log.Info("Fetching ingredient",
        zap.String("ingredient_id", id),
    )
    // Output: INFO Fetching ingredient request_id=9f4c7a8b ingredient_id=123
    
    ingredient, err := h.service.GetByID(id)
    if err != nil {
        log.Error("Failed to fetch", zap.Error(err))
        // Output: ERROR Failed to fetch request_id=9f4c7a8b error=not found
        return
    }
    
    log.Info("Ingredient fetched",
        zap.String("name", ingredient.Name),
    )
    // Output: INFO Ingredient fetched request_id=9f4c7a8b name=Tomato
}
```

### Pass Logger to Services

For even cleaner code, pass context-aware logger to services:

```go
// Handler
func (h *Handler) GetIngredient(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    log := logger.WithContext(ctx)
    
    ingredient, err := h.service.GetByID(ctx, log, id)
    if err != nil {
        log.Error("Service failed", zap.Error(err))
        models.NotFoundError(w, r, ...)
        return
    }
    
    models.SuccessResponse(w, r, ingredient)
}

// Service
func (s *Service) GetByID(ctx context.Context, log *zap.Logger, id string) (*Ingredient, error) {
    log.Info("Querying database", zap.String("id", id))
    // Output: INFO Querying database request_id=9f4c7a8b id=123
    
    ingredient, err := s.repo.FindByID(ctx, id)
    if err != nil {
        log.Error("Database query failed", zap.Error(err))
        // Output: ERROR Database query failed request_id=9f4c7a8b error=...
        return nil, err
    }
    
    log.Info("Ingredient found", zap.String("name", ingredient.Name))
    return ingredient, nil
}
```

### Benefits
- ✅ **Never forget** to log request_id
- ✅ **Consistent** across entire codebase
- ✅ **Easy to grep** all logs for one request
- ✅ **Clean code** - no manual request_id passing

---

## 3️⃣ Propagate Request ID to Downstream Services

### The Problem
When your handler calls external services (AI, payment gateways, etc.), you lose tracing:

```
Frontend → Backend → [AI Service]
   ↓          ↓           ↓
request_id  request_id   ??? (lost)
```

You can't correlate slow AI responses or errors with your backend logs.

### The Solution
Propagate request_id to all downstream calls.

---

### Example 1: Groq AI Recipe Generation

**Before (No Tracing):**
```go
func (s *AIService) GenerateRecipe(ctx context.Context, req Request) (*Recipe, error) {
    body := buildGroqRequest(req)
    
    httpReq, _ := http.NewRequest("POST", "https://api.groq.com/v1/chat", body)
    httpReq.Header.Set("Authorization", "Bearer "+s.apiKey)
    // ❌ No request_id - can't correlate with logs
    
    resp, err := s.client.Do(httpReq)
    if err != nil {
        // If AI times out, we don't know which backend request it was
        return nil, err
    }
    
    return parseRecipe(resp)
}
```

**After (Full Tracing):**
```go
import "github.com/dmitrijfomin/menu-fodifood/backend/internal/httputil"

func (s *AIService) GenerateRecipe(ctx context.Context, req Request) (*Recipe, error) {
    log := logger.WithContext(ctx)
    log.Info("Starting AI recipe generation",
        zap.String("title", req.Title),
        zap.String("language", req.Language),
    )
    
    body := buildGroqRequest(req)
    
    httpReq, _ := http.NewRequest("POST", "https://api.groq.com/v1/chat", body)
    httpReq.Header.Set("Authorization", "Bearer "+s.apiKey)
    
    // ✅ Propagate request_id to Groq
    httputil.AddRequestIDToHeaders(ctx, httpReq.Header)
    // Now Groq logs will include X-Request-ID: 9f4c7a8b...
    
    log.Info("Calling Groq API", zap.String("model", "llama-3.3-70b"))
    
    resp, err := s.client.Do(httpReq)
    if err != nil {
        log.Error("Groq API failed",
            zap.Error(err),
            zap.Duration("elapsed", time.Since(start)),
        )
        // Now in Groq dashboard: search by X-Request-ID to find slow request
        return nil, err
    }
    
    log.Info("Groq API success",
        zap.Int("status", resp.StatusCode),
        zap.Duration("elapsed", time.Since(start)),
    )
    
    return parseRecipe(resp)
}
```

---

### Example 2: Internal Service Calls

If you have microservices or separate Go services:

```go
func (s *Service) CallInternalAPI(ctx context.Context, data Data) (*Result, error) {
    log := logger.WithContext(ctx)
    
    // Create request
    req, _ := http.NewRequest("POST", "http://internal-service/process", body)
    
    // Propagate request_id
    httputil.AddRequestIDToHeaders(ctx, req.Header)
    
    log.Info("Calling internal service")
    
    resp, err := s.client.Do(req)
    if err != nil {
        log.Error("Internal service failed", zap.Error(err))
        return nil, err
    }
    
    return parseResult(resp)
}
```

**Now the internal service sees:**
```
// In internal service logs:
INFO Request received request_id=9f4c7a8b path=/process
```

**You can correlate:**
```bash
# Main service logs
grep "request_id=9f4c7a8b" main-service.log

# Internal service logs  
grep "request_id=9f4c7a8b" internal-service.log

# See the complete flow!
```

---

### Example 3: Database Queries (Advanced)

Some databases support query comments with correlation IDs:

```go
func (r *Repository) FindIngredient(ctx context.Context, id string) (*Ingredient, error) {
    requestID := logger.GetRequestID(ctx)
    
    // Add request_id as SQL comment (PostgreSQL)
    query := fmt.Sprintf(`
        /* request_id: %s */
        SELECT * FROM ingredients WHERE id = $1
    `, requestID)
    
    var ingredient Ingredient
    err := r.db.QueryRowContext(ctx, query, id).Scan(&ingredient)
    
    return &ingredient, err
}
```

**Now in PostgreSQL logs:**
```sql
/* request_id: 9f4c7a8b-1234-5678 */
SELECT * FROM ingredients WHERE id = '123'
```

You can correlate slow queries with specific API requests!

---

## 🎯 Complete Debugging Flow

### Scenario: User reports "AI recipe generation is slow"

#### Step 1: Get Request ID from Frontend
```typescript
// Frontend captures request_id on error/slow response
const response = await fetch('/api/admin/recipes/preview-ai', { ... });
const json = await response.json();

if (json.data) {
  console.log(`Recipe generated! (${json.meta.request_id})`);
} else {
  // Error - show request_id to user
  alert(`Error: ${json.error.message}\nRequest ID: ${json.meta.request_id}`);
}
```

User copies request_id: `9f4c7a8b-1234-5678-90ab-cdef12345678`

#### Step 2: Search Backend Logs
```bash
grep "request_id=9f4c7a8b" server.log
```

**Output:**
```
INFO  Incoming request         request_id=9f4c7a8b method=POST path=/api/admin/recipes/preview-ai
INFO  Starting AI generation   request_id=9f4c7a8b title="Grilled Salmon" language=en
INFO  Calling Groq API         request_id=9f4c7a8b model=llama-3.3-70b
WARN  Groq API slow response   request_id=9f4c7a8b elapsed=28.5s  ← Found it!
INFO  AI generation complete   request_id=9f4c7a8b steps_count=5
```

#### Step 3: Check Groq Dashboard
Search Groq logs by `X-Request-ID: 9f4c7a8b-1234-5678-90ab-cdef12345678`

**Find:**
- Model: llama-3.3-70b-versatile
- Tokens: 2500 input, 1200 output
- Latency: 28.5s (HIGH - this model usually takes 5-10s)
- Reason: High load on Groq infrastructure at that time

#### Step 4: Root Cause Found
- ✅ Not a backend bug
- ✅ Not a database issue
- ✅ Groq had high latency (external service)
- ✅ Can switch to faster model or add timeout

**Total debugging time:** 5 minutes (vs hours without request_id)

---

## 📊 Helper Functions Reference

### 1. Get Request ID from Context
```go
import "github.com/dmitrijfomin/menu-fodifood/backend/internal/platform/logger"

requestID := logger.GetRequestID(ctx)
// Returns: "9f4c7a8b..." or "unknown"
```

### 2. Get Logger with Request ID
```go
log := logger.WithContext(ctx)
log.Info("message") // Automatically includes request_id
```

### 3. Add Request ID to HTTP Headers
```go
import "github.com/dmitrijfomin/menu-fodifood/backend/internal/httputil"

httputil.AddRequestIDToHeaders(ctx, req.Header)
// Adds: X-Request-ID: 9f4c7a8b...
```

### 4. Get Request ID for Manual Use
```go
requestID := httputil.PropagateRequestID(ctx)
req.Header.Set("X-Request-ID", requestID)
req.Header.Set("X-Correlation-ID", requestID)  // Some services prefer this
```

---

## ✅ Best Practices

### DO:
✅ Use `logger.WithContext(ctx)` in all handlers  
✅ Propagate request_id to all external services  
✅ Pass context through entire call stack  
✅ Log request_id when debugging complex flows  
✅ Show request_id in frontend error messages  

### DON'T:
❌ Manually add request_id to logs (use WithContext)  
❌ Create new contexts (pass existing ctx)  
❌ Forget to propagate to external APIs  
❌ Log sensitive data with request_id  
❌ Rely only on X-Request-ID header (use meta.request_id in body too)  

---

## 🎓 Migration Checklist

Use this when migrating handlers:

```go
// ❌ BEFORE
func (h *Handler) SomeHandler(w http.ResponseWriter, r *http.Request) {
    logger.Log.Info("Processing request")
    
    result, err := h.service.DoSomething(param)
    if err != nil {
        logger.Log.Error("Failed", zap.Error(err))
        http.Error(w, "error", 500)
        return
    }
    
    json.NewEncoder(w).Encode(result)
}

// ✅ AFTER
func (h *Handler) SomeHandler(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    log := logger.WithContext(ctx)  // 1. Get logger with request_id
    
    log.Info("Processing request")  // 2. Logs include request_id automatically
    
    result, err := h.service.DoSomething(ctx, param)  // 3. Pass context
    if err != nil {
        log.Error("Failed", zap.Error(err))
        models.InternalServerError(w, r,  // 4. Unified response
            models.ErrorGeneralInternalError,
            "Failed to process",
            err.Error(),
        )
        return
    }
    
    models.SuccessResponse(w, r, result)  // 5. meta.request_id in response
}
```

---

## 🔥 Quick Wins

### Win 1: One-Line Logger Upgrade
```go
// Old: logger.Log.Info(...)
// New: logger.WithContext(ctx).Info(...)
// Result: request_id automatically in all logs
```

### Win 2: One-Line External API Tracing
```go
// Old: Make HTTP request
// New: httputil.AddRequestIDToHeaders(ctx, req.Header)
// Result: Full tracing to external services
```

### Win 3: Frontend Gets Request ID
```go
// No code change needed!
// All responses now have meta.request_id
// Frontend can send to Sentry/error tracking
```

---

**Status:** ✅ All features implemented and tested  
**Next:** Migrate handlers to use these features  
**Priority:** 🔥 Start with Admin Ingredients (SuggestIngredients)
