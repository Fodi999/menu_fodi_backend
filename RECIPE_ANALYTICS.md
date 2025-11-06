# 📊 Recipe Analytics & Metrics System

Extended recipe system with comprehensive nutrition data, cost analytics, and ChefTokens rewards.

## 🗄️ Database Schema

### Recipe Table - Extended Fields

```sql
-- Nutrition Metrics (Брутто/Нетто/Калории)
gross_weight  INTEGER           -- Total gross weight of raw ingredients (grams)
net_weight    INTEGER           -- Net weight after processing (grams)
calories      INTEGER           -- Total calories (kcal)
protein       NUMERIC(10,2)     -- Protein content (grams)
fats          NUMERIC(10,2)     -- Fat content (grams)
carbs         NUMERIC(10,2)     -- Carbohydrate content (grams)
yield         INTEGER           -- Final cooked dish weight (grams)

-- Economics
cost          NUMERIC(10,2)     -- Total recipe cost (PLN)

-- ChefTokens Gamification
tokens_reward INTEGER DEFAULT 10 -- Tokens awarded for creating recipe
views_count   INTEGER DEFAULT 0  -- Number of views
tokens_earned INTEGER DEFAULT 0  -- Tokens earned from views (1 per 10 views)

-- Engagement
likes_count   INTEGER DEFAULT 0  -- Number of likes (future feature)
```

### Indexes

```sql
CREATE INDEX idx_recipe_calories ON "Recipe"(calories);
CREATE INDEX idx_recipe_tokens_earned ON "Recipe"(tokens_earned DESC);
```

## 📡 Extended API Endpoints

### 1. Create Recipe with Metrics

**POST** `/api/recipes`

```bash
curl -X POST http://localhost:8080/api/recipes \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "title": "Філадельфія рол",
    "description": "Свіжий рол з лососем",
    "imageUrl": "https://cdn.example.com/philadelphia.jpg",
    "grossWeight": 580,
    "netWeight": 520,
    "calories": 920,
    "protein": 42.5,
    "fats": 35.0,
    "carbs": 98.0,
    "yield": 480,
    "cost": 52.00,
    "tokensReward": 25
  }'
```

**Response:**
```json
{
  "status": "success",
  "data": {
    "id": "e041199f-74b9-404a-abd5-e590b3227d18",
    "title": "Філадельфія рол",
    "description": "Свіжий рол з лососем",
    "imageUrl": "https://cdn.example.com/philadelphia.jpg",
    "authorId": "ac3c9c55-54b3-4fd8-8e47-ad634ba675c2",
    "author": {
      "id": "ac3c9c55-54b3-4fd8-8e47-ad634ba675c2",
      "name": "Test Chef",
      "email": "testchef@example.com"
    },
    "grossWeight": 580,
    "netWeight": 520,
    "calories": 920,
    "protein": 42.5,
    "fats": 35.0,
    "carbs": 98.0,
    "yield": 480,
    "cost": 52.00,
    "tokensReward": 25,
    "viewsCount": 0,
    "tokensEarned": 0,
    "createdAt": "2025-11-05T13:05:15.123Z",
    "updatedAt": "2025-11-05T13:05:15.123Z"
  }
}
```

### 2. Track Recipe Views

**POST** `/api/recipes/{id}/view`

Increment view count and award ChefTokens (1 token per 10 views).

```bash
curl -X POST http://localhost:8080/api/recipes/e041199f-74b9-404a-abd5-e590b3227d18/view
```

**Response:**
```json
{
  "status": "success",
  "data": {
    "viewsCount": 15,
    "tokensEarned": 1
  }
}
```

**View Tracking Logic:**
- Each call increments `viewsCount` by 1
- When `viewsCount % 10 == 0`, `tokensEarned` increases by 1
- No authentication required (public endpoint)
- Can be called from frontend when user views recipe details

### 3. Get All Recipes (with metrics)

**GET** `/api/posts`

```bash
curl http://localhost:8080/api/posts
```

Returns all recipes sorted by `created_at DESC` with full nutrition metrics.

### 4. Get User's Recipes

**GET** `/api/users/{userId}/posts`

```bash
curl http://localhost:8080/api/users/ac3c9c55-54b3-4fd8-8e47-ad634ba675c2/posts
```

Returns user's recipes with all metrics (for profile page).

## 📊 Metric Definitions

### Weight Metrics (Брутто/Нетто/Выход)

```
Gross Weight (Брутто) → Net Weight (Нетто) → Yield (Выход)
     580g                    520g                480g
   (raw with              (cleaned,            (cooked,
    waste)                trimmed)        water evaporated)
```

**Example: Making Philadelphia Roll**

| Ingredient | Gross (g) | Net (g) | Notes |
|------------|-----------|---------|-------|
| Rice | 300 | 300 | No waste |
| Salmon | 180 | 150 | 30g skin/bones removed |
| Cucumber | 100 | 90 | 10g peel/seeds removed |
| Cream cheese | 80 | 80 | No waste |
| Nori | 20 | 20 | No waste |
| **Total** | **580** | **520** | |
| **Yield** | | | **480** (water loss) |

### Nutrition Metrics

- **Calories** - Total energy content in kcal for entire dish
- **Protein** - Essential for muscle building (grams)
- **Fats** - Energy dense, essential fatty acids (grams)
- **Carbs** - Primary energy source (grams)

**Calculation:**
```
Per Portion = Total / Portions

Example (4 portions):
- Calories: 920 kcal / 4 = 230 kcal per portion
- Protein: 42.5g / 4 = 10.6g per portion
- Fats: 35.0g / 4 = 8.8g per portion
- Carbs: 98.0g / 4 = 24.5g per portion
```

### Cost Metrics (PLN)

```
Cost = Σ (ingredient price × amount)

Example:
- Rice (300g): 1.50 PLN
- Salmon (180g): 25.00 PLN
- Cucumber (100g): 2.00 PLN
- Cream cheese (80g): 4.50 PLN
- Nori (20g): 8.00 PLN
- Other: 11.00 PLN
------------------------
Total: 52.00 PLN
```

**Cost Per Portion:** 52.00 PLN / 4 = 13.00 PLN

## 🏆 ChefTokens Reward System

### Token Sources

1. **Creation Reward** (`tokensReward`)
   - One-time reward when recipe is published
   - Based on complexity:
     - 10 tokens: Simple (< 30 min, < 5 ingredients)
     - 20-30 tokens: Medium (30-60 min, 5-10 ingredients)
     - 40-50 tokens: Complex (60+ min, 10+ ingredients, advanced techniques)

2. **View Rewards** (`tokensEarned`)
   - Ongoing passive income
   - 1 token per 10 views
   - Unlimited earning potential

### Token Economy

```
Total Tokens = tokensReward + tokensEarned

Example Recipe Journey:
Day 1:  Created recipe → +25 tokens (tokensReward)
Day 2:  50 views       → +5 tokens (tokensEarned)
Week 1: 200 views      → +20 tokens
Month 1: 1,000 views   → +100 tokens
------------------------
Total after 1 month: 150 tokens
```

### Leaderboard Query

Get top recipe creators by tokens earned:
```sql
SELECT 
  r.title,
  r.author_id,
  u.name,
  r.tokens_reward,
  r.tokens_earned,
  (r.tokens_reward + r.tokens_earned) as total_tokens
FROM "Recipe" r
JOIN "User" u ON r.author_id = u.id
ORDER BY (r.tokens_reward + r.tokens_earned) DESC
LIMIT 10;
```

## 📈 Analytics Queries

### Popular Recipes by Views
```sql
SELECT title, views_count, tokens_earned
FROM "Recipe"
ORDER BY views_count DESC
LIMIT 20;
```

### Highest Calorie Recipes
```sql
SELECT title, calories, calories/4 as calories_per_portion
FROM "Recipe"
WHERE calories IS NOT NULL
ORDER BY calories DESC
LIMIT 10;
```

### Most Profitable Recipes (Views → Tokens)
```sql
SELECT 
  title,
  views_count,
  tokens_earned,
  ROUND(views_count::NUMERIC / NULLIF(tokens_earned, 0), 2) as views_per_token
FROM "Recipe"
WHERE tokens_earned > 0
ORDER BY tokens_earned DESC
LIMIT 10;
```

### High Protein Recipes
```sql
SELECT title, protein, protein/4 as protein_per_portion
FROM "Recipe"
WHERE protein IS NOT NULL
ORDER BY protein DESC
LIMIT 10;
```

### Budget-Friendly Recipes
```sql
SELECT title, cost, cost/4 as cost_per_portion
FROM "Recipe"
WHERE cost IS NOT NULL
ORDER BY cost ASC
LIMIT 10;
```

## 🎨 Frontend Integration

### Display Recipe Card

```jsx
function RecipeCard({ recipe }) {
  return (
    <div className="recipe-card">
      <img src={recipe.imageUrl} alt={recipe.title} />
      <h3>{recipe.title}</h3>
      <p>{recipe.description}</p>
      
      {/* Nutrition Facts */}
      <div className="nutrition">
        <span>🔥 {recipe.calories} kcal</span>
        <span>🥩 {recipe.protein}g protein</span>
        <span>🥑 {recipe.fats}g fats</span>
        <span>🍞 {recipe.carbs}g carbs</span>
      </div>
      
      {/* Cost & Tokens */}
      <div className="economics">
        <span>💰 {recipe.cost} PLN</span>
        <span>🏆 {recipe.tokensReward} tokens</span>
      </div>
      
      {/* Engagement */}
      <div className="stats">
        <span>👁️ {recipe.viewsCount} views</span>
        <span>💎 {recipe.tokensEarned} earned</span>
      </div>
    </div>
  );
}
```

### Track View on Recipe Detail Page

```javascript
async function trackRecipeView(recipeId) {
  try {
    const response = await fetch(`/api/recipes/${recipeId}/view`, {
      method: 'POST'
    });
    
    const { data } = await response.json();
    console.log(`Views: ${data.viewsCount}, Tokens: ${data.tokensEarned}`);
    
    // Update UI
    updateViewCount(data.viewsCount);
    
    // Show confetti if token earned
    if (data.viewsCount % 10 === 0) {
      showTokenReward(data.tokensEarned);
    }
  } catch (error) {
    console.error('Failed to track view:', error);
  }
}

// Call when user opens recipe detail page
useEffect(() => {
  trackRecipeView(recipe.id);
}, [recipe.id]);
```

### Filter Recipes by Nutrition

```javascript
// Get low-calorie recipes (< 500 kcal per portion)
const lowCalRecipes = recipes.filter(r => 
  r.calories && (r.calories / 4) < 500
);

// Get high-protein recipes (> 20g per portion)
const highProteinRecipes = recipes.filter(r =>
  r.protein && (r.protein / 4) > 20
);

// Get budget recipes (< 5 PLN per portion)
const budgetRecipes = recipes.filter(r =>
  r.cost && (r.cost / 4) < 5
);
```

## 🔄 Migration Path

### Migration 005 (Initial Metrics)
```sql
ALTER TABLE "Recipe"
ADD COLUMN gross_weight INTEGER,
ADD COLUMN net_weight INTEGER,
ADD COLUMN calories INTEGER,
ADD COLUMN yield INTEGER,
ADD COLUMN cost NUMERIC(10,2),
ADD COLUMN tokens_reward INTEGER DEFAULT 10,
ADD COLUMN views_count INTEGER DEFAULT 0;
```

### Migration 006 (Missing Nutrition)
```sql
ALTER TABLE "Recipe"
ADD COLUMN IF NOT EXISTS protein NUMERIC(10,2),
ADD COLUMN IF NOT EXISTS fats NUMERIC(10,2),
ADD COLUMN IF NOT EXISTS carbs NUMERIC(10,2),
ADD COLUMN IF NOT EXISTS tokens_earned INTEGER DEFAULT 0;
```

## 📱 Mobile App Features

### Recipe Detail Screen
- **Nutrition Facts Label** (FDA-style)
- **Cost Breakdown** (ingredient prices)
- **Chef Earnings** (tokens from this recipe)
- **Serving Size Calculator** (adjust portions, recalculate metrics)

### User Profile
- **Total Tokens Earned** (sum of all recipes)
- **Most Popular Recipe** (by views)
- **Total Views Across All Recipes**
- **Average Recipe Rating**

### Gamification
- **Achievements**
  - "First Recipe" - Create first recipe (+10 bonus tokens)
  - "Popular Chef" - Reach 1,000 total views (+50 tokens)
  - "Master Chef" - Create recipe with 50+ tokens reward
  - "Nutrition Expert" - Add full nutrition data to 10 recipes

## 🚀 Future Features

- [ ] Recipe versioning (track metric changes over time)
- [ ] User-submitted corrections (crowdsource better nutrition data)
- [ ] Recipe comparison tool (compare 2 recipes side-by-side)
- [ ] Meal planning (calculate daily nutrition from multiple recipes)
- [ ] Shopping list with cost estimation
- [ ] Carbon footprint calculation
- [ ] Allergen warnings
- [ ] Dietary labels (vegan, keto, paleo, etc.)

## 📊 Sample Data

See complete recipe with all metrics:
```bash
psql $DATABASE_URL -c "
SELECT 
  title,
  gross_weight,
  net_weight,
  calories,
  protein,
  fats,
  carbs,
  yield,
  cost,
  tokens_reward,
  views_count,
  tokens_earned
FROM \"Recipe\"
WHERE id='e041199f-74b9-404a-abd5-e590b3227d18';
"
```

## 💡 Best Practices

1. **Always provide all metrics** - Users love complete data
2. **Use AI generator** - Ensures realistic, consistent nutrition data
3. **Track views from day 1** - Passive token earning for creators
4. **Show per-portion metrics** - More useful than total
5. **Validate costs** - Keep prices updated for accuracy
6. **Encourage completeness** - Reward recipes with full data

---

**Part of FodiFoo Recipe Platform** - Bringing transparency and rewards to recipe sharing.
