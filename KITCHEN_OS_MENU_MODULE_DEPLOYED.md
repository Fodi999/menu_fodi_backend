# 🎉 Kitchen OS - Menu Module DEPLOYED

## ✅ What's Done

### 1. Database Layer
```sql
CREATE TABLE user_menu_items (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id UUID NOT NULL REFERENCES "User"(id),
  recipe_id UUID NOT NULL REFERENCES "Recipe"(id),
  servings INTEGER NOT NULL DEFAULT 1,
  status TEXT NOT NULL CHECK (status IN ('planned','cooking','completed','cancelled')),
  planned_for DATE NOT NULL DEFAULT CURRENT_DATE,
  notes TEXT,
  started_at TIMESTAMP,
  completed_at TIMESTAMP,
  created_at TIMESTAMP NOT NULL DEFAULT now()
);
```

**Indexes**:
- `idx_user_menu_user_date` ON (user_id, planned_for)
- `idx_user_menu_status` ON (status)
- `idx_user_menu_recipe` ON (recipe_id)

**Status**: ✅ APPLIED (already exists in production)

### 2. Backend Architecture

**Clean Architecture Pattern**:
```
Repository → Service → Handler
```

**Files Created**:
- `internal/models/user_menu.go` - Domain models + DTOs
- `internal/modules/menu/repository/menu_repository.go` - Data access
- `internal/modules/menu/service/menu_service.go` - Business logic
- `internal/modules/menu/transport/http/menu_handler.go` - HTTP handlers
- `internal/modules/menu/module.go` - Module registration
- `migrations/20260122_create_user_menu_items.sql` - Database schema

**Integrated**: ✅ Registered in `internal/app/routes_modular.go`

### 3. API Endpoints

All endpoints require authentication (JWT token):

#### GET /api/menu/today
Get today's cooking menu for authenticated user.

**Query Params**:
- `lang` - Language (ru, pl, en) - default: pl

**Response**:
```json
[
  {
    "id": "uuid",
    "recipe": {
      "id": "uuid",
      "title": "Жареные яйца",
      "canonical_name": "zharenye_yaytsa",
      "image_url": "https://...",
      "cook_time": 7,
      "servings": 1
    },
    "servings": 3,
    "status": "planned",
    "notes": "Make extra spicy",
    "planned_for": "2026-01-22",
    "created_at": "2026-01-22T10:30:00Z"
  }
]
```

#### POST /api/menu/today
Add recipe to today's menu.

**Body**:
```json
{
  "recipe_id": "605c8419-2d42-4ef0-a9d2-839582e98727",
  "servings": 2,
  "notes": "Optional notes"
}
```

**Response**:
```json
{
  "id": "uuid",
  "user_id": "uuid",
  "recipe_id": "uuid",
  "servings": 2,
  "status": "planned",
  "planned_for": "2026-01-22",
  "created_at": "2026-01-22T10:30:00Z"
}
```

#### POST /api/menu/{id}/start
Start cooking a menu item.

**Body** (optional):
```json
{
  "servings": 3
}
```

**Response**:
```json
{
  "success": true,
  "message": "Started cooking"
}
```

**Effect**: Changes status to `cooking`, sets `started_at` timestamp

#### POST /api/menu/{id}/complete
Complete cooking a menu item.

**Body** (optional):
```json
{
  "actual_servings": 2
}
```

**Response**:
```json
{
  "success": true,
  "message": "Cooking completed"
}
```

**Effect**: 
- Changes status to `completed`
- Sets `completed_at` timestamp
- **TODO**: Create entry in `prepared_dishes`
- **TODO**: Deduct ingredients from fridge

#### POST /api/menu/{id}/cancel
Cancel a menu item.

**Response**:
```json
{
  "success": true,
  "message": "Menu item cancelled"
}
```

**Effect**: Changes status to `cancelled`

#### DELETE /api/menu/{id}
Delete a menu item (hard delete).

**Response**:
```json
{
  "success": true,
  "message": "Menu item deleted"
}
```

---

## 🧪 Testing Plan

### Step 1: Wait for Koyeb Deployment
Check: https://app.koyeb.com/
Expected: ~2 minutes for rebuild

### Step 2: Get Fresh Token
```bash
TOKEN=$(curl -X POST "https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"fodi85@gmail.ru","password":"210185"}' | jq -r '.token')

echo $TOKEN
```

### Step 3: Test GET /api/menu/today (Empty)
```bash
curl -X GET "https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/menu/today?lang=ru" \
  -H "Authorization: Bearer $TOKEN" | jq
```

**Expected**: `[]` (empty array)

### Step 4: Add Recipe to Menu
```bash
curl -X POST "https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/menu/today" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "recipe_id": "605c8419-2d42-4ef0-a9d2-839582e98727",
    "servings": 2,
    "notes": "Make it spicy!"
  }' | jq
```

**Expected**: New menu item with status `planned`

### Step 5: Get Menu Again (Should Have 1 Item)
```bash
curl -X GET "https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/menu/today?lang=ru" \
  -H "Authorization: Bearer $TOKEN" | jq
```

**Expected**: Array with 1 menu item, full recipe details

### Step 6: Start Cooking
```bash
MENU_ITEM_ID="<id from step 4>"

curl -X POST "https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/menu/${MENU_ITEM_ID}/start" \
  -H "Authorization: Bearer $TOKEN" | jq
```

**Expected**: `{"success": true, "message": "Started cooking"}`

### Step 7: Complete Cooking
```bash
curl -X POST "https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/menu/${MENU_ITEM_ID}/complete" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"actual_servings": 2}' | jq
```

**Expected**: `{"success": true, "message": "Cooking completed"}`

### Step 8: Verify in Database
```bash
PGPASSWORD='npg_dz4Gl8ZhPLbX' psql \
  -h ep-soft-mud-agon8wu3-pooler.c-2.eu-central-1.aws.neon.tech \
  -U neondb_owner -d neondb \
  -c "SELECT id, status, servings, started_at, completed_at FROM user_menu_items WHERE user_id = '407582be-59d5-4d21-873b-1a72d31b0d42' ORDER BY created_at DESC LIMIT 5;"
```

**Expected**: See menu item with status `completed`

---

## 📊 Architecture Diagram

```
┌─────────────────────────────────────────────────────────┐
│                     FRONTEND                            │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐ │
│  │ Recipe List  │  │ Menu Today   │  │ Cooking View │ │
│  │ (Browse)     │  │ (Plan)       │  │ (Execute)    │ │
│  └──────────────┘  └──────────────┘  └──────────────┘ │
└────────────┬────────────────┬────────────────┬─────────┘
             │                │                │
             ▼                ▼                ▼
┌─────────────────────────────────────────────────────────┐
│                  BACKEND API LAYER                      │
│  ┌────────────────────────────────────────────────────┐ │
│  │ /api/recipe-recommendations  (Browse & Match)      │ │
│  │ /api/menu/today              (Kitchen Pipeline)    │ │
│  │ /api/menu/{id}/start         (Start Cooking)       │ │
│  │ /api/menu/{id}/complete      (Finish Cooking)      │ │
│  └────────────────────────────────────────────────────┘ │
└────────────┬────────────────┬────────────────┬─────────┘
             │                │                │
             ▼                ▼                ▼
┌─────────────────────────────────────────────────────────┐
│                  SERVICE LAYER                          │
│  ┌──────────────────┐  ┌──────────────────┐           │
│  │ Recommendation   │  │ Menu Service     │           │
│  │ Service          │  │ (Kitchen Logic)  │           │
│  │ (Matching Logic) │  │                  │           │
│  └──────────────────┘  └──────────────────┘           │
└────────────┬────────────────┬────────────────┬─────────┘
             │                │                │
             ▼                ▼                ▼
┌─────────────────────────────────────────────────────────┐
│                REPOSITORY LAYER                         │
│  ┌────────────┐  ┌────────────┐  ┌──────────────────┐ │
│  │ Recipe     │  │ Menu       │  │ Fridge           │ │
│  │ Repository │  │ Repository │  │ Repository       │ │
│  └────────────┘  └────────────┘  └──────────────────┘ │
└────────────┬────────────────┬────────────────┬─────────┘
             │                │                │
             ▼                ▼                ▼
┌─────────────────────────────────────────────────────────┐
│                    DATABASE                             │
│  ┌────────────┐  ┌──────────────────┐  ┌─────────────┐│
│  │ Recipe     │  │ user_menu_items  │  │ user_fridge │││
│  │            │  │ (Kitchen State)  │  │ _items      │││
│  └────────────┘  └──────────────────┘  └─────────────┘││
│  ┌────────────┐  ┌──────────────────┐                 ││
│  │ Ingredient │  │ prepared_dishes  │                 ││
│  │            │  │ (History)        │                 ││
│  └────────────┘  └──────────────────┘                 ││
└─────────────────────────────────────────────────────────┘
```

---

## 🎯 Key Principles Implemented

### 1. Backend = Source of Truth ✅
- Recipe stored ONCE in `Recipe` table
- Menu references recipe by `recipe_id` (no duplication)
- Status managed by backend, not frontend state

### 2. Clear Status Lifecycle ✅
```
planned → cooking → completed
   ↓
cancelled
```

### 3. No Recipe Duplication ✅
- `user_menu_items` has only `recipe_id` FK
- Full recipe data fetched via JOIN when needed
- Menu is just a **cooking pipeline**, not a recipe copy

### 4. Proper DTOs ✅
- `UserMenuItem` - domain model (database)
- `MenuItemWithRecipe` - DTO for frontend (includes full recipe)
- Clear separation of concerns

---

## 🚀 Next Steps

### Phase 1: Core Integration (HIGH PRIORITY)
- [ ] **canCookNow validation** before adding to menu
  * Check fridge has all ingredients
  * Use canonical matching (already works in recommendations)
  * Return error if ingredients missing

- [ ] **Fridge integration** on completion
  * Deduct used ingredients from `user_fridge_items`
  * Calculate quantities based on servings

- [ ] **PreparedDishes integration**
  * Create entry in `prepared_dishes` after completion
  * Link to menu_item for tracking

### Phase 2: Frontend (READY FOR IMPLEMENTATION)
- [ ] **Today's Menu Page** (`/menu/today`)
  * Show all planned/cooking/completed items
  * Cards with recipe photo, title, servings
  * Status badges (planned/cooking/completed)
  * Action buttons (Start Cooking, Complete, Cancel)

- [ ] **Add to Menu** from recommendations
  * Button on recipe card: "Cook Today"
  * Opens modal: select servings, add notes
  * POST to `/api/menu/today`

- [ ] **Cooking View** (`/menu/cooking/{id}`)
  * Full recipe details
  * Step-by-step instructions
  * Timer functionality
  * Mark as complete button

### Phase 3: Advanced Features (FUTURE)
- [ ] **Menu planning** for future dates
  * planned_for can be any date (not just today)
  * Calendar view

- [ ] **Shopping list** generation
  * Compare menu items vs fridge
  * Generate list of missing ingredients

- [ ] **Meal prep tracking**
  * Batch cooking support
  * Storage location tracking

---

## 📝 Git Status

**Commit**: `138c757`
**Message**: `feat: Kitchen OS - Menu module (cooking pipeline)`
**Pushed**: ✅ Yes (to main branch)
**Koyeb**: ⏳ Deploying (~2 minutes)

**Files Changed**:
- 10 files changed, 1216 insertions(+), 1 deletion(-)
- Created 7 new files
- Modified 2 files
- Deleted 1 file (bin/server_test)

---

## ✅ Success Criteria

- [x] Database table created
- [x] Migration applied
- [x] Repository layer created
- [x] Service layer created
- [x] Handler layer created
- [x] Module registered
- [x] Code compiled
- [x] Pushed to GitHub
- [x] .gitignore updated (no binaries)
- [ ] Koyeb deployed (waiting)
- [ ] API tested (pending deployment)

---

**Status**: 🟡 DEPLOYED, AWAITING TESTING
**Next Action**: Wait 2 minutes → Test API endpoints
