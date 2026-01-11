# 🔄 Migration Examples: Old → New Response Format

This guide shows how to migrate existing handlers to the new unified response format.

---

## 📦 Import Required Packages

```go
import (
    "github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
    "github.com/dmitrijfomin/menu-fodifood/backend/internal/middleware"
)
```

---

## Example 1: Auth Login Handler

### ❌ OLD (Inconsistent Format)

```go
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
    var req LoginRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    user, token, err := h.service.Login(req.Email, req.Password)
    if err != nil {
        http.Error(w, "Invalid credentials", http.StatusUnauthorized)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "success": true,
        "data": map[string]interface{}{
            "token": token,
            "user":  user,
        },
        "message": "Login successful",
    })
}
```

### ✅ NEW (Unified Format)

```go
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
    requestID := middleware.GetRequestID(r.Context())
    
    var req LoginRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        logger.Log.Error("Failed to decode login request",
            zap.String("request_id", requestID),
            zap.Error(err),
        )
        models.BadRequestError(w, r,
            models.ErrorGeneralInvalidJSON,
            "Invalid request body",
            "The request body must be valid JSON with email and password fields",
        )
        return
    }

    user, token, err := h.service.Login(req.Email, req.Password)
    if err != nil {
        logger.Log.Warn("Login failed",
            zap.String("request_id", requestID),
            zap.String("email", req.Email),
            zap.Error(err),
        )
        models.UnauthorizedError(w, r,
            models.ErrorAuthInvalidCredentials,
            "Invalid credentials",
            "Please check your email and password and try again",
        )
        return
    }

    logger.Log.Info("User logged in successfully",
        zap.String("request_id", requestID),
        zap.String("user_id", user.ID),
    )

    models.SuccessResponse(w, r, map[string]interface{}{
        "token": token,
        "user":  user,
    })
}
```

### 📊 Response Comparison

**OLD Response:**
```json
{
  "success": true,
  "data": {
    "token": "eyJhbGci...",
    "user": { ... }
  },
  "message": "Login successful"
}
```

**NEW Response:**
```json
{
  "data": {
    "token": "eyJhbGci...",
    "user": { ... }
  },
  "error": null,
  "meta": {
    "request_id": "9f4c7a8b-1234-5678-90ab-cdef12345678",
    "timestamp": "2026-01-11T10:30:00Z",
    "version": "v1"
  }
}
```

---

## Example 2: Get Ingredient by ID

### ❌ OLD (Direct Encoding)

```go
func (h *IngredientHandler) GetByID(w http.ResponseWriter, r *http.Request) {
    id := chi.URLParam(r, "id")
    
    ingredient, err := h.service.GetByID(id)
    if err != nil {
        http.Error(w, "Ingredient not found", http.StatusNotFound)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(ingredient)
}
```

### ✅ NEW (Unified Format)

```go
func (h *IngredientHandler) GetByID(w http.ResponseWriter, r *http.Request) {
    requestID := middleware.GetRequestID(r.Context())
    id := chi.URLParam(r, "id")
    
    logger.Log.Info("Fetching ingredient",
        zap.String("request_id", requestID),
        zap.String("ingredient_id", id),
    )
    
    ingredient, err := h.service.GetByID(id)
    if err != nil {
        logger.Log.Error("Ingredient not found",
            zap.String("request_id", requestID),
            zap.String("ingredient_id", id),
            zap.Error(err),
        )
        models.NotFoundError(w, r,
            models.ErrorIngredientNotFound,
            "Ingredient not found",
            fmt.Sprintf("Ingredient with ID '%s' does not exist", id),
        )
        return
    }

    models.SuccessResponse(w, r, ingredient)
}
```

---

## Example 3: Create Recipe (with validation)

### ❌ OLD

```go
func (h *RecipeHandler) Create(w http.ResponseWriter, r *http.Request) {
    var req CreateRecipeRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid input", http.StatusBadRequest)
        return
    }

    recipe, err := h.service.Create(req)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(recipe)
}
```

### ✅ NEW

```go
func (h *RecipeHandler) Create(w http.ResponseWriter, r *http.Request) {
    requestID := middleware.GetRequestID(r.Context())
    
    var req CreateRecipeRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        logger.Log.Error("Failed to decode recipe request",
            zap.String("request_id", requestID),
            zap.Error(err),
        )
        models.BadRequestError(w, r,
            models.ErrorGeneralInvalidJSON,
            "Invalid request body",
            "The request body must be valid JSON",
        )
        return
    }

    // Validate input
    if err := req.Validate(); err != nil {
        logger.Log.Warn("Recipe validation failed",
            zap.String("request_id", requestID),
            zap.Error(err),
        )
        models.BadRequestError(w, r,
            models.ErrorRecipeValidationFailed,
            "Recipe validation failed",
            err.Error(),
        )
        return
    }

    recipe, err := h.service.Create(req)
    if err != nil {
        logger.Log.Error("Failed to create recipe",
            zap.String("request_id", requestID),
            zap.Error(err),
        )
        models.InternalServerError(w, r,
            models.ErrorGeneralDatabaseError,
            "Failed to create recipe",
            "An internal error occurred while creating the recipe",
        )
        return
    }

    logger.Log.Info("Recipe created successfully",
        zap.String("request_id", requestID),
        zap.String("recipe_id", recipe.ID),
    )

    models.CreatedResponse(w, r, recipe)
}
```

---

## Example 4: List Ingredients (Paginated)

### ❌ OLD

```go
func (h *IngredientHandler) List(w http.ResponseWriter, r *http.Request) {
    page, _ := strconv.Atoi(r.URL.Query().Get("page"))
    limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
    
    if page < 1 {
        page = 1
    }
    if limit < 1 {
        limit = 50
    }

    ingredients, total, err := h.service.List(page, limit)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "data":  ingredients,
        "total": total,
        "page":  page,
        "limit": limit,
    })
}
```

### ✅ NEW

```go
func (h *IngredientHandler) List(w http.ResponseWriter, r *http.Request) {
    requestID := middleware.GetRequestID(r.Context())
    
    page, _ := strconv.Atoi(r.URL.Query().Get("page"))
    limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
    
    if page < 1 {
        page = 1
    }
    if limit < 1 || limit > 100 {
        limit = 50
    }

    logger.Log.Info("Listing ingredients",
        zap.String("request_id", requestID),
        zap.Int("page", page),
        zap.Int("limit", limit),
    )

    ingredients, total, err := h.service.List(page, limit)
    if err != nil {
        logger.Log.Error("Failed to list ingredients",
            zap.String("request_id", requestID),
            zap.Error(err),
        )
        models.InternalServerError(w, r,
            models.ErrorGeneralDatabaseError,
            "Failed to fetch ingredients",
            "An internal error occurred while fetching ingredients",
        )
        return
    }

    models.PaginatedResponse(w, r, ingredients, page, limit, total)
}
```

---

## Example 5: Fridge - Add Item with Price

### ❌ OLD

```go
func (h *FridgeHandler) AddItem(w http.ResponseWriter, r *http.Request) {
    userID := r.Context().Value("user_id").(string)
    
    var req AddItemRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid input", http.StatusBadRequest)
        return
    }

    item, err := h.service.AddItem(userID, req)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(map[string]interface{}{
        "success": true,
        "data":    item,
    })
}
```

### ✅ NEW

```go
func (h *FridgeHandler) AddItem(w http.ResponseWriter, r *http.Request) {
    requestID := middleware.GetRequestID(r.Context())
    userID := r.Context().Value("user_id").(string)
    
    var req AddItemRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        logger.Log.Error("Failed to decode fridge item request",
            zap.String("request_id", requestID),
            zap.String("user_id", userID),
            zap.Error(err),
        )
        models.BadRequestError(w, r,
            models.ErrorGeneralInvalidJSON,
            "Invalid request body",
            "The request body must be valid JSON",
        )
        return
    }

    // Validate input
    if req.IngredientID == "" {
        models.BadRequestError(w, r,
            models.ErrorFridgeInvalidInput,
            "Missing ingredient ID",
            "The ingredientId field is required",
        )
        return
    }

    if req.Quantity <= 0 {
        models.BadRequestError(w, r,
            models.ErrorFridgeInvalidInput,
            "Invalid quantity",
            "Quantity must be greater than zero",
        )
        return
    }

    item, err := h.service.AddItem(userID, req)
    if err != nil {
        logger.Log.Error("Failed to add fridge item",
            zap.String("request_id", requestID),
            zap.String("user_id", userID),
            zap.Error(err),
        )
        models.InternalServerError(w, r,
            models.ErrorGeneralDatabaseError,
            "Failed to add item to fridge",
            "An internal error occurred while adding the item",
        )
        return
    }

    logger.Log.Info("Fridge item added successfully",
        zap.String("request_id", requestID),
        zap.String("user_id", userID),
        zap.String("item_id", item.ID),
    )

    models.CreatedResponse(w, r, item)
}
```

---

## Example 6: AI Recipe Preview (Long-running operation)

### ✅ NEW (Best Practice)

```go
func (h *RecipeHandler) PreviewAI(w http.ResponseWriter, r *http.Request) {
    requestID := middleware.GetRequestID(r.Context())
    
    var req AIRecipeRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        logger.Log.Error("Failed to decode AI recipe request",
            zap.String("request_id", requestID),
            zap.Error(err),
        )
        models.BadRequestError(w, r,
            models.ErrorGeneralInvalidJSON,
            "Invalid request body",
            "The request body must be valid JSON",
        )
        return
    }

    // Validate AI request
    if err := req.Validate(); err != nil {
        models.BadRequestError(w, r,
            models.ErrorRecipeValidationFailed,
            "Recipe validation failed",
            err.Error(),
        )
        return
    }

    logger.Log.Info("Starting AI recipe generation",
        zap.String("request_id", requestID),
        zap.String("title", req.Title),
        zap.String("language", req.Language),
    )

    // Call AI service (may take 5-30 seconds)
    recipe, err := h.aiService.GenerateRecipe(r.Context(), req)
    if err != nil {
        logger.Log.Error("AI recipe generation failed",
            zap.String("request_id", requestID),
            zap.Error(err),
        )
        models.InternalServerError(w, r,
            models.ErrorRecipeAIGenerationFailed,
            "Failed to generate recipe",
            fmt.Sprintf("AI generation failed: %v", err),
        )
        return
    }

    logger.Log.Info("AI recipe generated successfully",
        zap.String("request_id", requestID),
        zap.String("recipe_id", recipe.ID),
        zap.Int("steps_count", len(recipe.Steps)),
    )

    models.SuccessResponse(w, r, recipe)
}
```

---

## 🎯 Key Improvements

### 1. **Consistent Format**
- Every success: `{ data, error: null, meta }`
- Every error: `{ data: null, error, meta }`

### 2. **Request ID Tracking**
- Every log includes `request_id`
- Frontend can track request through entire flow
- Easy debugging: Sentry → logs → database

### 3. **Structured Logging**
- Use `zap.String()`, `zap.Error()`, etc.
- Easy to parse and search
- Automatic correlation with request_id

### 4. **Error Codes**
- Machine-readable: `INGREDIENT_NOT_FOUND`
- Consistent across API
- Easy to document and handle

### 5. **Better Error Messages**
- `code`: What went wrong (for machines)
- `message`: User-friendly message
- `details`: Additional context for debugging

---

## 📊 Frontend Integration

### TypeScript Interface

```typescript
interface APIResponse<T> {
  data: T | null;
  error: {
    code: string;
    message: string;
    details?: string;
  } | null;
  meta: {
    request_id: string;
    timestamp: string;
    version?: string;
  };
}

// Usage
const response = await fetch('/api/ingredients/123');
const json: APIResponse<Ingredient> = await response.json();

if (json.error) {
  console.error(`[${json.meta.request_id}] ${json.error.code}: ${json.error.message}`);
  // Handle specific error codes
  switch (json.error.code) {
    case 'INGREDIENT_NOT_FOUND':
      // Show "ingredient not found" UI
      break;
    case 'AUTH_INVALID_TOKEN':
      // Redirect to login
      break;
  }
} else {
  // Use json.data safely
  console.log(json.data);
}
```

---

## ✅ Migration Checklist

- [ ] Create `internal/models/response.go`
- [ ] Create `internal/models/errors.go`
- [ ] Create `internal/middleware/request_id.go`
- [ ] Add RequestIDMiddleware to router in `cmd/server/main.go`
- [ ] Migrate auth handlers (highest priority)
- [ ] Migrate admin/ingredients handlers
- [ ] Migrate admin/recipes handlers
- [ ] Migrate fridge handlers
- [ ] Migrate user handlers
- [ ] Update tests to expect new format
- [ ] Update frontend to parse new format
- [ ] Update API documentation

---

**Status:** ✅ Ready to implement  
**Priority:** 🔥 HIGH  
**Estimated time:** 3-5 days
