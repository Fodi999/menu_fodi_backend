# Admin Recipe Management API - Professional Draft/Publish Workflow

## 🎯 Архитектурная философия

**Один Recipe — разные режимы работы**

```
Recipe
├── status (draft / published / archived)
├── source (manual / ai / import)
└── content completeness
```

**НЕ делаем:**
- ❌ `admin_recipes` table
- ❌ `catalog_recipes` table  
- ❌ Separate models

**Делаем:**
- ✅ ONE `Recipe` table
- ✅ Status-based lifecycle
- ✅ Separate modules по назначению

---

## 📦 Модули

### 1. `recipes_admin` (NEW) - Admin CRUD
**Назначение:** Управление рецептами админами (draft/publish workflow)  
**Endpoints:** `/api/admin/recipes/*`  
**Функции:**
- Create draft
- Update draft (PATCH)
- Publish (full validation)
- Archive

### 2. `recipes` (Existing) - User Operations
**Назначение:** Matching, cooking, AI, user recipes  
**Endpoints:** `/api/recipes/*`  
**Функции:**
- Match with fridge
- Cook recipes
- AI adaptation
- User-generated recipes

---

## 🔄 Recipe Lifecycle

```
┌─────────┐  Create   ┌─────────┐  Publish   ┌───────────┐
│  DRAFT  │ ────────> │  DRAFT  │ ────────>  │ PUBLISHED │
└─────────┘   (POST)  └─────────┘   (POST)   └───────────┘
                │                                    │
                │                                    │
                │ Edit (PATCH)                       │ Archive
                │                                    ▼
                │                              ┌───────────┐
                └──────────────────────────────│ ARCHIVED  │
                                               └───────────┘
                                                     │
                                                     │ Republish
                                                     │
                                               Back to PUBLISHED
```

---

## 📚 API Reference

### 1. Create Draft Recipe

**Endpoint:** `POST /api/admin/recipes`

**Validation:** ✅ MINIMAL (только required fields)

**Request:**
```json
{
  "title": "Pierogi ruskie",
  "canonicalName": "Pierogi Ruskie",
  "description": "Traditional Polish dumplings",
  "imageUrl": "https://example.com/image.jpg",
  "country": "PL",
  "category": "main",
  "difficulty": "medium",
  "timeMinutes": 45,
  "servings": 4
}
```

**Response:** `201 Created`
```json
{
  "status": "success",
  "data": {
    "id": "uuid-here",
    "title": "Pierogi ruskie",
    "canonicalName": "Pierogi Ruskie",
    "status": "draft",
    "authorId": "admin-uuid",
    "createdAt": "2026-01-05T12:00:00Z"
  }
}
```

**Backend автоматически устанавливает:**
- ✅ `source = {"type":"manual"}`
- ✅ `status = "draft"`
- ✅ `authorId` из JWT token

---

### 2. Update Draft Recipe (PATCH)

**Endpoint:** `PATCH /api/admin/recipes/{id}`

**Validation:** ✅ NONE (можно обновлять что угодно)

**Constraints:**
- ⚠️ Только если `status = "draft"`
- ❌ Published recipes нельзя редактировать (только archive)

**Request (partial updates):**
```json
{
  "description": "Updated description",
  "timeMinutes": 50,
  "ingredients": [
    {
      "ingredientId": "potato-uuid",
      "quantity": 500,
      "unit": "g",
      "optional": false
    }
  ],
  "steps": [
    {
      "order": 1,
      "description": "Peel and boil potatoes",
      "duration": 20
    },
    {
      "order": 2,
      "description": "Prepare dough",
      "duration": 15
    }
  ]
}
```

**Response:** `200 OK`
```json
{
  "status": "success",
  "data": {
    "id": "uuid-here",
    "title": "Pierogi ruskie",
    "status": "draft",
    "updatedAt": "2026-01-05T12:30:00Z"
  }
}
```

---

### 3. Publish Recipe

**Endpoint:** `POST /api/admin/recipes/{id}/publish`

**Validation:** ✅ FULL (строгие требования)

**Required:**
- ✅ `ingredients.length >= 1`
- ✅ `steps.length >= 1`
- ✅ `title.length >= 3`
- ✅ Steps в sequential order (1, 2, 3...)

**Warnings (не блокируют):**
- ⚠️ Missing description
- ⚠️ Missing nutrition
- ⚠️ Missing translations

**Request:**
```json
{
  "ingredients": [
    {
      "ingredientId": "potato-uuid",
      "quantity": 500,
      "unit": "g",
      "optional": false,
      "notes": "Peeled"
    },
    {
      "ingredientId": "flour-uuid",
      "quantity": 300,
      "unit": "g",
      "optional": false
    }
  ],
  "steps": [
    {
      "order": 1,
      "description": "Peel and boil potatoes until soft, about 20 minutes",
      "duration": 20,
      "temperature": 100
    },
    {
      "order": 2,
      "description": "Prepare dough by mixing flour with water",
      "duration": 15
    },
    {
      "order": 3,
      "description": "Form dumplings and cook in boiling water",
      "duration": 10,
      "temperature": 100
    }
  ],
  "force": false
}
```

**Response:** `200 OK`
```json
{
  "status": "success",
  "data": {
    "id": "uuid-here",
    "title": "Pierogi ruskie",
    "status": "published",
    "publishedAt": "2026-01-05T13:00:00Z",
    "ingredientsCount": 2,
    "stepsCount": 3,
    "warnings": [
      "Missing nutrition information"
    ]
  }
}
```

**Errors:**
```json
{
  "error": "at least 1 ingredient required for publishing",
  "status": 400
}
```

```json
{
  "error": "at least 1 step required for publishing",
  "status": 400
}
```

```json
{
  "error": "can only publish draft or archived recipes",
  "status": 403
}
```

---

### 4. Archive Recipe

**Endpoint:** `POST /api/admin/recipes/{id}/archive`

**Purpose:** Hide recipe from public (can be republished later)

**Request:** Empty body `{}`

**Response:** `200 OK`
```json
{
  "status": "success",
  "message": "Recipe archived successfully"
}
```

---

### 5. Get Draft Recipes

**Endpoint:** `GET /api/admin/recipes/drafts`

**Response:** `200 OK`
```json
{
  "status": "success",
  "data": [
    {
      "id": "uuid-1",
      "title": "Draft Recipe 1",
      "status": "draft",
      "updatedAt": "2026-01-05T12:00:00Z"
    },
    {
      "id": "uuid-2",
      "title": "Draft Recipe 2",
      "status": "draft",
      "updatedAt": "2026-01-05T11:30:00Z"
    }
  ],
  "meta": {
    "total": 2
  }
}
```

---

## 🔒 Security

**Authentication:** All endpoints require JWT token

**Role Check:** Admin/Super Admin only (can add middleware)

**Author Check:** Draft recipes belong to creator

---

## 📊 Database Schema

### Status Field

```sql
ALTER TABLE "Recipe" 
ADD COLUMN status VARCHAR(20) NOT NULL DEFAULT 'draft'
CHECK (status IN ('draft', 'published', 'archived'));

CREATE INDEX idx_recipe_status ON "Recipe"(status);
```

### Recipe Statuses

| Status | Editable | Public | Can Publish | Can Archive |
|--------|----------|--------|-------------|-------------|
| `draft` | ✅ Yes | ❌ No | ✅ Yes | ✅ Yes |
| `published` | ❌ No | ✅ Yes | ❌ No | ✅ Yes |
| `archived` | ❌ No | ❌ No | ✅ Yes (republish) | ❌ No |

---

## 🧪 Testing Workflow

### Complete Flow Test

```bash
# 1. Login as admin
TOKEN=$(curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"admin_password_123"}' \
  | jq -r '.token')

# 2. Create draft
DRAFT=$(curl -X POST http://localhost:8080/api/admin/recipes \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Test Recipe",
    "country": "PL",
    "category": "main",
    "difficulty": "easy",
    "timeMinutes": 30,
    "servings": 4
  }')

RECIPE_ID=$(echo $DRAFT | jq -r '.data.id')
echo "Created draft: $RECIPE_ID"

# 3. Update draft (add ingredients/steps)
curl -X PATCH http://localhost:8080/api/admin/recipes/$RECIPE_ID \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "description": "Updated description",
    "ingredients": [
      {
        "ingredientId": "some-uuid",
        "quantity": 500,
        "unit": "g"
      }
    ],
    "steps": [
      {
        "order": 1,
        "description": "First step"
      }
    ]
  }'

# 4. Publish
curl -X POST http://localhost:8080/api/admin/recipes/$RECIPE_ID/publish \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "ingredients": [{"ingredientId":"uuid","quantity":500,"unit":"g"}],
    "steps": [{"order":1,"description":"Cook it"}]
  }'

# 5. Try to edit published (should fail)
curl -X PATCH http://localhost:8080/api/admin/recipes/$RECIPE_ID \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"Changed"}' # ❌ Error: can only update draft recipes

# 6. Archive
curl -X POST http://localhost:8080/api/admin/recipes/$RECIPE_ID/archive \
  -H "Authorization: Bearer $TOKEN"
```

---

## 🎨 Frontend Integration

### React Hook Example

```typescript
// hooks/useAdminRecipes.ts
export const useCreateDraft = () => {
  const createDraft = async (data: CreateRecipeInput) => {
    const response = await api.post('/api/admin/recipes', data);
    return response.data.data; // { id, title, status: 'draft' }
  };
  
  return { createDraft };
};

export const useUpdateDraft = () => {
  const updateDraft = async (id: string, data: Partial<UpdateRecipeInput>) => {
    const response = await api.patch(`/api/admin/recipes/${id}`, data);
    return response.data.data;
  };
  
  return { updateDraft };
};

export const usePublishRecipe = () => {
  const publish = async (id: string, data: PublishRecipeInput) => {
    const response = await api.post(`/api/admin/recipes/${id}/publish`, data);
    return response.data.data;
  };
  
  return { publish };
};
```

### Form Flow

```tsx
// 1. Create Draft Form (minimal fields)
<RecipeDraftForm 
  onSubmit={async (data) => {
    const draft = await createDraft(data);
    navigate(`/admin/recipes/${draft.id}/edit`);
  }}
/>

// 2. Edit Draft Form (full editor)
<RecipeEditForm 
  recipeId={id}
  onSave={async (updates) => {
    await updateDraft(id, updates); // PATCH - partial updates
  }}
/>

// 3. Publish Button
<PublishButton 
  onClick={async () => {
    const result = await publish(id, {
      ingredients: [...],
      steps: [...]
    });
    
    if (result.warnings.length > 0) {
      showWarnings(result.warnings);
    }
    
    navigate(`/recipes/${id}`);
  }}
/>
```

---

## ✅ Success Criteria

### Draft Creation
- ✅ Minimal validation (only basic fields)
- ✅ No ingredients required
- ✅ No steps required
- ✅ Backend sets `source = manual`
- ✅ Backend sets `status = draft`

### Draft Editing
- ✅ PATCH endpoint (partial updates)
- ✅ Can add/edit ingredients
- ✅ Can add/edit steps
- ✅ Can edit any field
- ✅ Only works if `status = draft`

### Publishing
- ✅ Full validation
- ✅ Ingredients >= 1 required
- ✅ Steps >= 1 required
- ✅ Description check (warning)
- ✅ Sequential steps order validation
- ✅ Changes status to `published`
- ✅ Returns warnings (non-blocking issues)

---

## 🚀 Industry Standards

This pattern follows:
- ✅ **Strapi** - Draft/Publish CMS
- ✅ **Sanity** - Content staging
- ✅ **Directus** - Revision control
- ✅ **WordPress** - Post drafts

**Key principles:**
1. Separate concerns (admin vs user operations)
2. Progressive validation (lenient → strict)
3. Status-based lifecycle
4. Non-destructive workflow
5. Clear separation of modules

---

## 📝 Summary

**Module Structure:**
```
internal/modules/
├── recipes/          # User operations (matching, cooking, AI)
└── recipes_admin/    # Admin CRUD (draft/publish workflow)
    ├── dto/         # Create, Update, Publish DTOs
    ├── service/     # Business logic
    └── transport/   # HTTP handlers
```

**Endpoints:**
- `POST /api/admin/recipes` - Create draft (minimal)
- `PATCH /api/admin/recipes/{id}` - Update draft (anything)
- `POST /api/admin/recipes/{id}/publish` - Publish (strict validation)
- `POST /api/admin/recipes/{id}/archive` - Archive
- `GET /api/admin/recipes/drafts` - List drafts

**Status:** 🎉 Production-ready, professional architecture
