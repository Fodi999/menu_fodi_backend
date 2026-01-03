# 📊 Recipe Database Schema

Generated: 2026-01-03

## 📈 Current Statistics

```
Total Recipes: 10
Countries: 3 (Poland, Italy, Greece)
Categories: 3 (main, salad, soup)
Saved Recipes: 4
Cook Logs: 7
Avg Ingredients per Recipe: 4.1
```

---

## 🗂️ Main Tables

### 1️⃣ **Recipe** (Core table)

```sql
Table: "Recipe"
Columns:
  - id                  uuid (PK, auto-generated)
  - canonicalName       varchar(255) UNIQUE NOT NULL  -- "Pierogi Ruskie"
  - localName           varchar(255) NOT NULL          -- "Pierogi ruskie"
  - country             varchar(100) NOT NULL          -- "Poland"
  - region              varchar(100)                   -- "Małopolska"
  - category            varchar(50) NOT NULL           -- enum: appetizer, soup, salad, main, side, dessert, beverage
  - difficulty          varchar(20) NOT NULL           -- enum: easy, medium, hard
  - timeMinutes         integer NOT NULL               -- 90
  - servings            integer NOT NULL DEFAULT 4     -- 4
  - steps               jsonb NOT NULL DEFAULT '[]'    -- [{step: 1, instruction: "..."}]
  - nutritionProfile    jsonb DEFAULT '{}'             -- {calories: 350, protein: 12, ...}
  - source              jsonb NOT NULL                 -- {type: "traditional", author: "..."}
  - description         text                           -- Full description
  - imageUrl            text                           -- URL to recipe image
  - createdAt           timestamp NOT NULL DEFAULT now()
  - updatedAt           timestamp NOT NULL DEFAULT now()

Indexes:
  - PRIMARY KEY (id)
  - UNIQUE (canonicalName)
  - idx_recipe_category (category)
  - idx_recipe_country (country)
  - idx_recipe_difficulty (difficulty)
  - idx_recipe_time (timeMinutes)

Check Constraints:
  - category IN ('appetizer', 'soup', 'salad', 'main', 'side', 'dessert', 'beverage')
  - difficulty IN ('easy', 'medium', 'hard')
```

**Example Data:**
```
ID: e8ab233b-46dc-4a55-8e87-1c0dd3656790
Canonical Name: Pierogi Ruskie
Local Name: Pierogi ruskie
Country: Poland
Category: main
Difficulty: medium
Time: 90 minutes
Servings: 4
```

---

### 2️⃣ **RecipeIngredient** (Recipe-Ingredient junction)

```sql
Table: "RecipeIngredient"
Columns:
  - id              uuid (PK, auto-generated)
  - recipeId        uuid NOT NULL (FK → Recipe.id ON DELETE CASCADE)
  - ingredientId    text NOT NULL (FK → Ingredient.id ON DELETE RESTRICT)
  - ingredientKey   varchar(255) NOT NULL      -- "potato" (для поиска/matching)
  - quantity        numeric(10,2) NOT NULL     -- 500.00
  - unit            varchar(50) NOT NULL       -- "g", "ml", "pcs"
  - optional        boolean DEFAULT false      -- true для опциональных ингредиентов
  - sortOrder       integer NOT NULL DEFAULT 0 -- Порядок отображения
  - groupName       varchar(50)                -- "Тесто", "Начинка", NULL
  - createdAt       timestamp NOT NULL DEFAULT now()

Indexes:
  - PRIMARY KEY (id)
  - UNIQUE (recipeId, ingredientId) -- один ингредиент только раз в рецепте
  - idx_recipe_ingredient_recipe (recipeId)
  - idx_recipe_ingredient_key (ingredientKey)
  - idx_recipe_ingredient_group (groupName)

Foreign Keys:
  - recipeId → Recipe(id) ON DELETE CASCADE
  - ingredientId → Ingredient(id) ON DELETE RESTRICT
```

**Example Data (Pierogi Ruskie):**
```
Ingredient      | Quantity | Unit | Optional | Group
----------------|----------|------|----------|------
Ziemniak        | 500.00   | g    | false    | NULL
Twaróg          | 250.00   | g    | false    | NULL
Cebula          | 200.00   | g    | false    | NULL
Mąka pszenna    | 400.00   | g    | false    | NULL
Jaja            | 1.00     | pcs  | false    | NULL
Masło           | 50.00    | g    | true     | NULL
```

---

### 3️⃣ **RecipeCookLog** (Cooking history with economics)

```sql
Table: "RecipeCookLog"
Columns:
  - id                  uuid (PK, auto-generated)
  - userId              text NOT NULL (FK → User.id ON DELETE CASCADE)
  - recipeId            uuid NOT NULL (FK → Recipe.id ON DELETE CASCADE)
  - servingsMultiplier  numeric(10,2) NOT NULL DEFAULT 1.0  -- 1.5 = готовим на 6 порций
  - cookedAt            timestamptz NOT NULL DEFAULT now()
  - usedValue           numeric(10,2) NOT NULL DEFAULT 0    -- Ценность использованных продуктов (PLN)
  - wasteRiskSaved      numeric(10,2) NOT NULL DEFAULT 0    -- Спасено от выброса (PLN)
  - totalRecipeCost     numeric(10,2) NOT NULL DEFAULT 0    -- Общая стоимость рецепта (PLN)
  - idempotencyKey      varchar(255) UNIQUE                 -- Предотвращает дубликаты
  - createdAt           timestamptz NOT NULL DEFAULT now()
  - updatedAt           timestamptz NOT NULL DEFAULT now()

Indexes:
  - PRIMARY KEY (id)
  - UNIQUE (userId, recipeId, cookedAt) -- один рецепт раз в момент времени
  - UNIQUE (idempotencyKey)
  - idx_cook_log_user (userId)
  - idx_cook_log_recipe (recipeId)
  - idx_cook_log_cooked_at (cookedAt DESC)

Foreign Keys:
  - userId → User(id) ON DELETE CASCADE
  - recipeId → Recipe(id) ON DELETE CASCADE
```

**Example Data:**
```
Cooked At: 2025-12-21 23:26:18
Recipe: Greek Salad
Servings Multiplier: 1.00
Total Recipe Cost: 4.95 PLN
Used Value: 4.95 PLN
Waste Risk Saved: 1.40 PLN
```

---

### 4️⃣ **RecipeCookIngredient** (Деталировка использованных ингредиентов)

```sql
Table: "RecipeCookIngredient"
Columns:
  - id              uuid (PK)
  - cookLogId       uuid NOT NULL (FK → RecipeCookLog.id ON DELETE CASCADE)
  - ingredientId    text NOT NULL (FK → Ingredient.id ON DELETE RESTRICT)
  - quantityUsed    numeric(10,2) NOT NULL -- Сколько использовано из холодильника
  - unit            varchar(50) NOT NULL
  - sourceType      varchar(50) NOT NULL   -- 'fridge' или 'purchased'
  - fridgeItemId    uuid                   -- FK → user_fridge_items.id (если из холодильника)
  - costPerUnit     numeric(10,2)          -- Цена за единицу (PLN)
  - totalCost       numeric(10,2)          -- Общая стоимость = quantityUsed * costPerUnit
  - wasteRisk       numeric(10,2)          -- Риск выброса (если близко к expire)
  - createdAt       timestamptz NOT NULL DEFAULT now()

Foreign Keys:
  - cookLogId → RecipeCookLog(id) ON DELETE CASCADE
  - ingredientId → Ingredient(id) ON DELETE RESTRICT
  - fridgeItemId → user_fridge_items(id) ON DELETE SET NULL
```

---

### 5️⃣ **RecipeDietTag** (Диетические теги)

```sql
Table: "RecipeDietTag"
Columns:
  - id          uuid (PK)
  - recipeId    uuid NOT NULL (FK → Recipe.id ON DELETE CASCADE)
  - dietTagId   uuid NOT NULL (FK → DietTag.id ON DELETE CASCADE)

Indexes:
  - PRIMARY KEY (id)
  - UNIQUE (recipeId, dietTagId)
  - idx_recipe_diet_tag_recipe (recipeId)
  - idx_recipe_diet_tag_tag (dietTagId)

Foreign Keys:
  - recipeId → Recipe(id) ON DELETE CASCADE
  - dietTagId → DietTag(id) ON DELETE CASCADE
```

**Examples:**
- Vegetarian
- Vegan
- Gluten-free
- Keto
- Low-carb

---

### 6️⃣ **RecipeAllergen** (Аллергены)

```sql
Table: "RecipeAllergen"
Columns:
  - id          uuid (PK)
  - recipeId    uuid NOT NULL (FK → Recipe.id ON DELETE CASCADE)
  - allergenId  uuid NOT NULL (FK → Allergen.id ON DELETE CASCADE)

Indexes:
  - PRIMARY KEY (id)
  - UNIQUE (recipeId, allergenId)
  - idx_recipe_allergen_recipe (recipeId)
  - idx_recipe_allergen_allergen (allergenId)

Foreign Keys:
  - recipeId → Recipe(id) ON DELETE CASCADE
  - allergenId → Allergen(id) ON DELETE CASCADE
```

**Examples:**
- Gluten (wheat, rye)
- Dairy (milk, cheese)
- Eggs
- Nuts
- Shellfish

---

### 7️⃣ **user_saved_recipes** (Сохраненные рецепты пользователей)

```sql
Table: "user_saved_recipes"
Columns:
  - id          uuid (PK)
  - user_id     text NOT NULL (FK → User.id ON DELETE CASCADE)
  - recipe_id   uuid NOT NULL (FK → Recipe.id ON DELETE CASCADE)
  - saved_at    timestamptz NOT NULL DEFAULT now()
  - notes       text                 -- Личные заметки пользователя

Indexes:
  - PRIMARY KEY (id)
  - UNIQUE (user_id, recipe_id)
  - idx_saved_recipes_user (user_id)
  - idx_saved_recipes_recipe (recipe_id)
  - idx_saved_recipes_saved_at (saved_at DESC)

Foreign Keys:
  - user_id → User(id) ON DELETE CASCADE
  - recipe_id → Recipe(id) ON DELETE CASCADE
```

**Current Stats:**
- Total Saved Recipes: 4

---

### 8️⃣ **user_recipe_sessions** (Сессии рекомендаций)

```sql
Table: "user_recipe_sessions"
Columns:
  - id                uuid (PK)
  - user_id           text NOT NULL (FK → User.id ON DELETE CASCADE)
  - last_recipe_id    uuid (FK → Recipe.id ON DELETE SET NULL)
  - excluded_recipes  jsonb DEFAULT '[]'  -- [uuid, uuid, ...] - исключенные рецепты
  - session_start     timestamptz NOT NULL DEFAULT now()
  - last_activity     timestamptz NOT NULL DEFAULT now()

Indexes:
  - PRIMARY KEY (id)
  - idx_recipe_sessions_user (user_id)
  - idx_recipe_sessions_last_activity (last_activity DESC)

Foreign Keys:
  - user_id → User(id) ON DELETE CASCADE
  - last_recipe_id → Recipe(id) ON DELETE SET NULL
```

**Purpose:** Отслеживает историю рекомендаций, чтобы не показывать один и тот же рецепт дважды подряд.

---

## 🔗 Entity Relationships

```
Recipe (1) ──────────────── (N) RecipeIngredient ──────── (1) Ingredient
   │                                                             │
   │                                                             │
   ├─────────── (N) RecipeDietTag ──────── (1) DietTag         │
   │                                                             │
   ├─────────── (N) RecipeAllergen ───────(1) Allergen         │
   │                                                             │
   ├─────────── (N) RecipeCookLog ──────── (1) User            │
   │                  │                                          │
   │                  └──────── (N) RecipeCookIngredient ───────┘
   │                                  │
   │                                  └──────── (1) user_fridge_items
   │
   ├─────────── (N) user_saved_recipes ──────── (1) User
   │
   └─────────── (N) user_recipe_sessions ─────── (1) User
```

---

## 📊 Data Distribution

### By Category:
```
main     → 8 recipes (80%)
salad    → 1 recipe  (10%)
soup     → 1 recipe  (10%)
```

### By Country:
```
Poland   → 7 recipes (70%)
Italy    → 2 recipes (20%)
Greece   → 1 recipe  (10%)
```

### By Difficulty:
```
easy     → majority
medium   → some complex recipes
hard     → advanced recipes
```

---

## 🎯 Key Features

### 1. **Recipe Matching**
- Match recipes to user's fridge ingredients
- Calculate match percentage
- Identify missing ingredients
- Sort by best match

### 2. **Economics Tracking**
- Track total recipe cost (`totalRecipeCost`)
- Monitor ingredient usage value (`usedValue`)
- Calculate waste prevention (`wasteRiskSaved`)
- Historical cost analysis

### 3. **Flexible Ingredient Groups**
```sql
groupName examples:
- "Тесто" (Dough)
- "Начинка" (Filling)
- "Соус" (Sauce)
- NULL (ungrouped)
```

### 4. **Cooking Sessions**
- Idempotency via `idempotencyKey`
- Prevent duplicate cook logs
- Track serving multipliers
- Detailed ingredient usage

### 5. **User Preferences**
- Save favorite recipes
- Add personal notes
- Exclude unwanted recipes
- Session-based recommendations

---

## 🔍 Common Queries

### Get Recipe with Ingredients:
```sql
SELECT 
    r."canonicalName",
    r.difficulty,
    r."timeMinutes",
    json_agg(
        json_build_object(
            'ingredient', i.name,
            'quantity', ri.quantity,
            'unit', ri.unit,
            'optional', ri.optional
        ) ORDER BY ri."sortOrder"
    ) as ingredients
FROM "Recipe" r
JOIN "RecipeIngredient" ri ON r.id = ri."recipeId"
JOIN "Ingredient" i ON ri."ingredientId" = i.id
WHERE r.id = 'e8ab233b-46dc-4a55-8e87-1c0dd3656790'
GROUP BY r.id;
```

### Recipe Match Score:
```sql
-- Calculates how many ingredients user has in fridge
SELECT 
    r.id,
    r."canonicalName",
    COUNT(DISTINCT ri."ingredientId") as total_ingredients,
    COUNT(DISTINCT CASE 
        WHEN ufi.ingredient_id IS NOT NULL THEN ri."ingredientId"
    END) as matched_ingredients,
    ROUND(
        COUNT(DISTINCT CASE WHEN ufi.ingredient_id IS NOT NULL THEN ri."ingredientId" END)::numeric 
        / COUNT(DISTINCT ri."ingredientId") * 100, 
        1
    ) as match_percentage
FROM "Recipe" r
JOIN "RecipeIngredient" ri ON r.id = ri."recipeId"
LEFT JOIN user_fridge_items ufi ON ri."ingredientId" = ufi.ingredient_id 
    AND ufi.user_id = '407582be-59d5-4d21-873b-1a72d31b0d42'
GROUP BY r.id, r."canonicalName"
HAVING COUNT(DISTINCT ri."ingredientId") > 0
ORDER BY match_percentage DESC;
```

### User Cooking Statistics:
```sql
SELECT 
    u.name,
    COUNT(*) as total_cooks,
    SUM(rcl."totalRecipeCost") as total_spent,
    SUM(rcl."wasteRiskSaved") as total_waste_saved,
    COUNT(DISTINCT rcl."recipeId") as unique_recipes_cooked
FROM "RecipeCookLog" rcl
JOIN "User" u ON rcl."userId" = u.id
WHERE rcl."userId" = '407582be-59d5-4d21-873b-1a72d31b0d42'
GROUP BY u.id, u.name;
```

---

## 📝 Notes

- **Multilingual Support**: Ready for `canonicalName` (EN) + `localName` (native language)
- **JSONB Fields**: Flexible schema for `steps`, `nutritionProfile`, `source`
- **Cascade Deletes**: Deleting recipe removes all related data automatically
- **Unique Constraints**: Prevent duplicate recipe-ingredient, recipe-tag, etc.
- **Timestamps**: All tables track `createdAt` and `updatedAt`
- **Economics**: Full cost tracking from fridge to cooked meal

---

## 🚀 Future Enhancements

- [ ] Add recipe ratings table
- [ ] Add recipe comments/reviews
- [ ] Add recipe images storage
- [ ] Add recipe video URLs
- [ ] Add nutritional analysis per serving
- [ ] Add recipe variations (e.g., "Pierogi Ruskie (vegan)")
- [ ] Add meal planning table
- [ ] Add shopping list generation
