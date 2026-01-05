# 🔧 Fix: DELETE /api/admin/ingredients/{id} - 404 Error

## 🎯 Problem

Frontend calling `DELETE /api/admin/ingredients/{id}` was getting **404 Not Found**.

**Error:**
```
DELETE http://localhost:3000/api/admin/ingredients/f0124a11-bfd2-47c8-8406-8370adedca6b 404 (Not Found)
```

**Root cause:**
- `/api/admin/ingredients` routes didn't exist
- Only `/stock/{id}` (pro_chef) and `/catalog/ingredients` (read-only) existed
- No admin CRUD for the Ingredient catalog itself

---

## ✅ Solution

### 1. Added Admin Routes for Ingredient Catalog

**File:** `internal/modules/ingredients/module.go`

```go
// 🔧 ADMIN ROUTES - Управление справочником ингредиентов (ТОЛЬКО admin/super_admin)
r.Route("/admin/ingredients", func(r chi.Router) {
    r.Use(jwtMiddleware)
    r.Use(middleware.AdminMiddleware)

    r.Get("/", m.handlers.ListIngredients)        // Список ингредиентов
    r.Post("/", m.handlers.Create)                // Создать ингредиент
    r.Get("/{id}", m.handlers.GetOne)             // Детали ингредиента
    r.Put("/{id}", m.handlers.Update)             // Обновить ингредиент
    r.Delete("/{id}", m.handlers.DeleteFromCatalog) // Удалить ингредиент (из каталога)
    r.Get("/search", m.handlers.Search)           // Поиск ингредиентов
})
```

### 2. Created Repository Method for Catalog Deletion

**File:** `internal/database/ingredient_repository.go`

```go
// DeleteIngredient удаляет ингредиент из каталога (admin operation)
// Cascade: автоматически удалит связанные StockItem через foreign key constraints
func (r *IngredientRepository) DeleteIngredient(ingredientID string) error {
	result := DB.Delete(&models.Ingredient{}, "id = ?", ingredientID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("ingredient not found")
	}
	return nil
}
```

### 3. Added Service Method

**File:** `internal/modules/ingredients/service/service.go`

```go
// DeleteIngredientFromCatalog удаление ингредиента из каталога (admin operation)
// Cascade: автоматически удалит связанные StockItem
func (s *IngredientsService) DeleteIngredientFromCatalog(ingredientID string) error {
	return s.repo.DeleteIngredient(ingredientID)
}
```

### 4. Created Admin Handler

**File:** `internal/modules/ingredients/transport/http/handlers.go`

```go
// DeleteFromCatalog удаление ингредиента из каталога (admin operation)
func (h *IngredientsHandlers) DeleteFromCatalog(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.service.DeleteIngredientFromCatalog(id); err != nil {
		if err.Error() == "ingredient not found" {
			httpx.NotFound(w, "Ingredient not found")
			return
		}
		httpx.InternalError(w, "Failed to delete ingredient from catalog")
		return
	}

	httpx.Success(w, map[string]string{"message": "Ingredient deleted from catalog successfully"})
}
```

---

## 🔍 Separation of Concerns

Now we have THREE distinct ingredient endpoints:

### 1. `/catalog/ingredients` - Read-only catalog browsing
- **Who:** ALL authenticated users
- **Purpose:** Browse ingredient catalog
- **Operations:** GET list, GET search

### 2. `/stock` - Stock management
- **Who:** `pro_chef` only
- **Purpose:** Manage warehouse inventory (StockItem)
- **Operations:** CRUD on StockItem (quantity, batch, expiry)

### 3. `/admin/ingredients` - Catalog administration
- **Who:** `admin`, `super_admin` only
- **Purpose:** Manage the ingredient **catalog** itself
- **Operations:** CRUD on Ingredient (name, category, unit, etc.)
- **Delete behavior:** Cascade deletes associated StockItems

---

## 🎯 Key Differences

| Endpoint | Deletes | ID Type | Auth Level |
|----------|---------|---------|------------|
| `DELETE /stock/{id}` | StockItem + Ingredient | StockItem ID | pro_chef |
| `DELETE /admin/ingredients/{id}` | Ingredient (cascade StockItems) | Ingredient ID | admin/super_admin |

---

## ✅ Result

**Frontend can now:**
```typescript
DELETE /api/admin/ingredients/{ingredientId}
Authorization: Bearer {token}

// Response 200:
{
  "success": true,
  "data": {
    "message": "Ingredient deleted from catalog successfully"
  }
}

// Response 404:
{
  "success": false,
  "code": 404,
  "message": "Ingredient not found"
}
```

---

## 🔥 Test

```bash
# Login as admin
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"admin_password_123"}'

# Delete ingredient from catalog
curl -X DELETE http://localhost:8080/api/admin/ingredients/{id} \
  -H "Authorization: Bearer {token}"
```

---

## 📂 Modified Files

```
internal/database/ingredient_repository.go    ✅ Added DeleteIngredient()
internal/modules/ingredients/service/service.go    ✅ Added DeleteIngredientFromCatalog()
internal/modules/ingredients/transport/http/handlers.go    ✅ Added DeleteFromCatalog()
internal/modules/ingredients/module.go    ✅ Added /admin/ingredients routes
```

**Status:** ✅ Fixed - Frontend can now delete ingredients via admin panel
