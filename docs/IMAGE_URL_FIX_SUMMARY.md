# ✅ Image URL Fix - Решение проблемы

## 🐛 Проблема

Фронтенд не получал `imageUrl` при запросе рецептов через admin endpoint `/api/admin/recipes`, хотя данные были в базе.

**Симптомы:**
```javascript
// Фронтенд лог
[useAdminRecipes] 🖼️ imageUrl: undefined
[useAdminRecipes] 🔑 All keys: ['id', 'canonicalName', 'title', ...] // НЕТ imageUrl
```

**База данных:**
```sql
SELECT imageUrl FROM "Recipe" WHERE id = '6b8628ef-...';
-- ✅ Есть: https://res.cloudinary.com/dwrn0ohbp/image/upload/...
```

---

## 🔍 Причина

### Два endpoint'а - разные обработчики

1. **Публичный API** (`GET /api/recipes`) - **работал ✅**
   - Handler: `internal/modules/recipes/transport/http/handler.go`
   - Метод: `ListRecipes()`
   - Включал `imageUrl` в ответ

2. **Admin API** (`GET /api/admin/recipes`) - **не работал ❌**
   - Handler: `internal/modules/admin/transport/http/handlers.go`
   - Метод: `GetAllRecipes()`
   - **НЕ включал `imageUrl` в DTO**

### Root cause

В `RecipeResponse` DTO отсутствовали поля:
```go
// internal/modules/admin/transport/http/handlers.go (СТАРЫЙ КОД)
type RecipeResponse struct {
    ID            string `json:"id"`
    CanonicalName string `json:"canonicalName"`
    Title         string `json:"title"`
    // ... другие поля
    // ❌ ImageUrl отсутствует
    // ❌ ImagePublicId отсутствует
}
```

---

## ✅ Решение

### Изменения в `internal/modules/admin/transport/http/handlers.go`

#### 1. Добавлены поля в `RecipeResponse` struct

```go
type RecipeResponse struct {
    ID            string  `json:"id"`
    CanonicalName string  `json:"canonicalName"`
    Title         string  `json:"title"`
    
    // ✅ ДОБАВЛЕНО
    ImageUrl      *string `json:"imageUrl,omitempty"`
    ImagePublicId *string `json:"imagePublicId,omitempty"`
    
    // ... остальные поля
}
```

#### 2. Обновлён mapper `ToRecipeResponse()`

```go
func ToRecipeResponse(recipe *models.RecipeCatalog) RecipeResponse {
    response := RecipeResponse{
        ID:            recipe.ID.String(),
        CanonicalName: recipe.CanonicalName,
        Title:         recipe.Title,
        
        // ✅ ДОБАВЛЕНО: маппинг image полей
        ImageUrl:      stringPtrOrNil(recipe.ImageUrl),
        ImagePublicId: stringPtrOrNil(recipe.ImagePublicId),
        
        // ... остальные поля
    }
    
    return response
}

// ✅ ДОБАВЛЕНА: helper функция
func stringPtrOrNil(s string) *string {
    if s == "" {
        return nil
    }
    return &s
}
```

---

## 🧪 Тестирование

### До исправления ❌
```bash
curl /api/admin/recipes | jq '.data[0].imageUrl'
# null (хотя в БД есть данные)
```

### После исправления ✅
```bash
curl /api/admin/recipes | jq '.data[0] | {id, title, imageUrl}'
# {
#   "id": "4aa22783-45cc-4fc4-8800-4340a5c93ce9",
#   "title": "яйца жареные на масле",
#   "imageUrl": "https://res.cloudinary.com/dwrn0ohbp/image/upload/v1768827688/recipes/recipe_4aa22783-45cc-4fc4-8800-4340a5c93ce9.webp"
# }
```

---

## 📊 Результат

### Оба endpoint'а теперь работают

| Endpoint | URL | Auth | ImageUrl |
|----------|-----|------|----------|
| Public Catalog | `GET /api/recipes` | ❌ None | ✅ Возвращает |
| Admin Catalog | `GET /api/admin/recipes` | ✅ JWT (admin) | ✅ Возвращает |

### Проверка в production

```bash
# 1. Публичный API
curl "https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/recipes" | \
  jq '.data.recipes[0].imageUrl'
# ✅ "https://res.cloudinary.com/dwrn0ohbp/..."

# 2. Admin API (с JWT)
TOKEN=$(curl -s -X POST \
  https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"admin_password_123"}' | \
  jq -r '.data.token')

curl "https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/admin/recipes" \
  -H "Authorization: Bearer $TOKEN" | \
  jq '.data[0].imageUrl'
# ✅ "https://res.cloudinary.com/dwrn0ohbp/..."
```

---

## 🎯 Commits

1. **43a1fa2** - "Add imageUrl fields to RecipeCatalog model"
   - Добавлены поля `ImageUrl` и `ImagePublicId` в модель `RecipeCatalog`

2. **7df45af** - "Include imageUrl in ListRecipes API response"
   - Добавлен `imageUrl` в ответ публичного API `/api/recipes`

3. **f8c9e1d** - "Add imageUrl to admin RecipeResponse DTO"
   - Добавлен `imageUrl` в ответ admin API `/api/admin/recipes` ✅

---

## 🔗 Связанные файлы

- `internal/models/recipe_catalog.go` - модель с image полями
- `internal/modules/recipes/transport/http/handler.go` - публичный API handler
- `internal/modules/admin/transport/http/handlers.go` - admin API handler (исправлен)
- `internal/modules/admin/transport/http/recipe_image.go` - upload/delete handlers
- `pkg/cloudinary/client.go` - Cloudinary client

---

## ✅ Статус

**Полностью работает на production!** 🎉

Теперь фронтенд может получать `imageUrl` от обоих endpoint'ов и отображать фотографии рецептов в каталоге.
