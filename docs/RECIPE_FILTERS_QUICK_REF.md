# Recipe Filters - Quick Reference

## ✅ Status: IMPLEMENTED

Все фильтры **УЖЕ работают** в endpoint `/api/recipes/match`.

---

## 📋 Available Filters

### 1. Country Filter
**Фильтр по стране**

```http
GET /api/recipes/match?country=Poland
GET /api/recipes/match?country=Italy
GET /api/recipes/match?country=France
```

**Supported countries** (из seed data):
- `Poland` - Pierogi, Bigos, Jajecznica
- `Italy` - Carbonara, Pizza Margherita
- `Greece` - Greek Salad

---

### 2. Category Filter
**Фильтр по категории блюда**

```http
GET /api/recipes/match?category=main
GET /api/recipes/match?category=salad
GET /api/recipes/match?category=soup
```

**Categories**:
- `main` - основные блюда
- `soup` - супы
- `salad` - салаты
- `appetizer` - закуски
- `side` - гарниры
- `dessert` - десерты
- `beverage` - напитки

---

### 3. Difficulty Filter
**Фильтр по сложности**

```http
GET /api/recipes/match?difficulty=easy
GET /api/recipes/match?difficulty=medium
GET /api/recipes/match?difficulty=hard
```

**Levels**:
- `easy` - простые рецепты (10-30 мин)
- `medium` - средней сложности (30-90 мин)
- `hard` - сложные рецепты (90+ мин)

---

### 4. Time Filter
**Максимальное время приготовления**

```http
GET /api/recipes/match?maxTime=30
GET /api/recipes/match?maxTime=60
GET /api/recipes/match?maxTime=120
```

**Examples**:
- `maxTime=30` - только быстрые рецепты (≤30 мин)
- `maxTime=60` - до часа
- No limit if omitted

---

### 5. Allergen Exclusion
**Исключить аллергены**

```http
GET /api/recipes/match?excludeAllergens=gluten
GET /api/recipes/match?excludeAllergens=gluten,lactose
GET /api/recipes/match?excludeAllergens=gluten,lactose,eggs
```

**Available allergens** (EU 14):
- `gluten` - пшеница, ячмень, рожь
- `lactose` - молочные продукты
- `eggs` - яйца
- `fish` - рыба
- `shellfish` - ракообразные
- `nuts` - орехи
- `peanuts` - арахис
- `soy` - соя
- `celery` - сельдерей
- `mustard` - горчица
- `sesame` - кунжут
- `sulfites` - сульфиты
- `lupin` - люпин
- `molluscs` - моллюски

---

### 6. Diet Tags
**Фильтр по диетам**

```http
GET /api/recipes/match?dietTags=vegetarian
GET /api/recipes/match?dietTags=vegan,gluten-free
GET /api/recipes/match?dietTags=keto,high-protein
```

**Available tags**:
- `vegetarian` - без мяса и рыбы
- `vegan` - без животных продуктов
- `gluten-free` - без глютена
- `dairy-free` - без молочных продуктов
- `keto` - низкоуглеводная, высокожировая
- `paleo` - без зерновых, бобовых, молочных
- `low-carb` - низкоуглеводная
- `high-protein` - высокобелковая
- `pescatarian` - рыба, но без мяса
- `halal` - по исламским законам
- `kosher` - по еврейским законам

---

### 7. Match Score Filter
**Минимальный score совпадения**

```http
GET /api/recipes/match?minScore=50
GET /api/recipes/match?minScore=70
GET /api/recipes/match?minScore=90
```

**Default**: `50` (50% совпадение)

**Use cases**:
- `minScore=90` - только почти готовые рецепты
- `minScore=50` - показать больше вариантов
- `minScore=0` - все рецепты (даже если ничего нет)

---

### 8. Limit Results
**Максимум результатов**

```http
GET /api/recipes/match?limit=5
GET /api/recipes/match?limit=10
GET /api/recipes/match?limit=20
```

**Default**: `10`

---

## 🎯 Combined Filters

### Example 1: Quick Vegetarian Italian
```http
GET /api/recipes/match?country=Italy&dietTags=vegetarian&maxTime=30&minScore=60
```

**Returns**: Итальянские вегетарианские рецепты до 30 мин с 60%+ match.

---

### Example 2: No Gluten, No Lactose
```http
GET /api/recipes/match?excludeAllergens=gluten,lactose&minScore=70
```

**Returns**: Рецепты без глютена и лактозы с 70%+ match.

---

### Example 3: Easy Polish Recipes
```http
GET /api/recipes/match?country=Poland&difficulty=easy&maxTime=60
```

**Returns**: Простые польские рецепты до часа.

---

### Example 4: Keto High-Protein
```http
GET /api/recipes/match?dietTags=keto,high-protein&excludeAllergens=gluten&limit=5
```

**Returns**: Топ-5 кето высокобелковых рецептов без глютена.

---

## 🎨 Frontend Implementation

### Simple Filter UI

```typescript
// components/RecipeFilters.tsx
import { useState } from 'react';

const RecipeFilters = ({ onApply }) => {
  const [filters, setFilters] = useState({
    country: '',
    category: '',
    difficulty: '',
    maxTime: 0,
    excludeAllergens: [],
    dietTags: [],
    minScore: 50,
  });

  const applyFilters = () => {
    const params = new URLSearchParams();
    
    if (filters.country) params.append('country', filters.country);
    if (filters.category) params.append('category', filters.category);
    if (filters.difficulty) params.append('difficulty', filters.difficulty);
    if (filters.maxTime) params.append('maxTime', filters.maxTime.toString());
    if (filters.excludeAllergens.length) {
      params.append('excludeAllergens', filters.excludeAllergens.join(','));
    }
    if (filters.dietTags.length) {
      params.append('dietTags', filters.dietTags.join(','));
    }
    params.append('minScore', filters.minScore.toString());
    
    onApply(params);
  };

  return (
    <div className="filters">
      {/* Country */}
      <Select
        label="Kraj"
        value={filters.country}
        onChange={(e) => setFilters({ ...filters, country: e.target.value })}
      >
        <option value="">Wszystkie</option>
        <option value="Poland">Polska</option>
        <option value="Italy">Włochy</option>
        <option value="France">Francja</option>
        <option value="Greece">Grecja</option>
      </Select>

      {/* Difficulty */}
      <Select
        label="Trudność"
        value={filters.difficulty}
        onChange={(e) => setFilters({ ...filters, difficulty: e.target.value })}
      >
        <option value="">Wszystkie</option>
        <option value="easy">Łatwe</option>
        <option value="medium">Średnie</option>
        <option value="hard">Trudne</option>
      </Select>

      {/* Max Time */}
      <Input
        type="number"
        label="Maks. czas (min)"
        value={filters.maxTime || ''}
        onChange={(e) => setFilters({ ...filters, maxTime: parseInt(e.target.value) || 0 })}
        placeholder="Bez limitu"
      />

      {/* Allergens */}
      <MultiSelect
        label="Wyklucz alergeny"
        options={[
          { value: 'gluten', label: 'Gluten' },
          { value: 'lactose', label: 'Laktoza' },
          { value: 'eggs', label: 'Jajka' },
          { value: 'nuts', label: 'Orzechy' },
        ]}
        value={filters.excludeAllergens}
        onChange={(values) => setFilters({ ...filters, excludeAllergens: values })}
      />

      {/* Diet Tags */}
      <MultiSelect
        label="Diety"
        options={[
          { value: 'vegetarian', label: 'Wegetariańska' },
          { value: 'vegan', label: 'Wegańska' },
          { value: 'keto', label: 'Keto' },
          { value: 'gluten-free', label: 'Bez glutenu' },
        ]}
        value={filters.dietTags}
        onChange={(values) => setFilters({ ...filters, dietTags: values })}
      />

      {/* Min Score */}
      <Slider
        label="Min. dopasowanie"
        min={0}
        max={100}
        step={10}
        value={filters.minScore}
        onChange={(value) => setFilters({ ...filters, minScore: value })}
      />

      <Button onClick={applyFilters}>Zastosuj filtry</Button>
    </div>
  );
};
```

---

### Usage in Page

```typescript
// pages/recipes/index.tsx
import { useRecipeMatches } from '@/hooks/useRecipeMatches';
import { RecipeFilters } from '@/components/RecipeFilters';

const RecipesPage = () => {
  const [params, setParams] = useState(new URLSearchParams());
  const { recipes, loading } = useRecipeMatches(params.toString());

  return (
    <div>
      <RecipeFilters onApply={setParams} />
      
      {loading ? (
        <Spinner />
      ) : (
        <RecipeGrid recipes={recipes} />
      )}
    </div>
  );
};
```

---

## 📊 Filter Performance

**Query optimization** already implemented:

```go
// Indexed fields for fast filtering
CREATE INDEX idx_recipe_country ON "Recipe"(country);
CREATE INDEX idx_recipe_category ON "Recipe"(category);
CREATE INDEX idx_recipe_difficulty ON "Recipe"(difficulty);
CREATE INDEX idx_recipe_time ON "Recipe"("timeMinutes");
```

**Expected performance**:
- Simple filter (country, difficulty): **~10ms**
- Complex filter (allergens + diet tags): **~20-30ms**
- Full match calculation: **~50-150ms**

---

## 🎯 Filter Priority

### Phase 1: MVP (Now)
- ✅ Country
- ✅ Difficulty
- ✅ Max Time
- ✅ Exclude Allergens
- ✅ Diet Tags
- ✅ Min Score

### Phase 2: Advanced (Later)
- ⏳ Cuisine type (Italian, French, Asian)
- ⏳ Preparation method (baked, fried, grilled)
- ⏳ Ingredient count (< 5 ingredients)
- ⏳ Cost range (cheap, moderate, expensive)
- ⏳ Seasonality (winter, summer recipes)

---

## 🧪 Testing Examples

### Test 1: Vegetarian only
```bash
curl -H "Authorization: Bearer $TOKEN" \
  "https://api.fodifood.com/api/recipes/match?dietTags=vegetarian"
```

**Expected**: Pierogi, Pizza Margherita, Jajecznica, Greek Salad

---

### Test 2: No gluten, no lactose
```bash
curl -H "Authorization: Bearer $TOKEN" \
  "https://api.fodifood.com/api/recipes/match?excludeAllergens=gluten,lactose"
```

**Expected**: Greek Salad, Jajecznica (if no milk)

---

### Test 3: Quick recipes (<30 min)
```bash
curl -H "Authorization: Bearer $TOKEN" \
  "https://api.fodifood.com/api/recipes/match?maxTime=30"
```

**Expected**: Carbonara, Jajecznica

---

### Test 4: Polish easy recipes
```bash
curl -H "Authorization: Bearer $TOKEN" \
  "https://api.fodifood.com/api/recipes/match?country=Poland&difficulty=easy"
```

**Expected**: Jajecznica

---

## ✅ Implementation Status

### Backend ✅
- [x] RecipeFilters struct
- [x] loadRecipesWithFilters() with all filters
- [x] Query parameter parsing
- [x] Database indexes
- [x] Allergen exclusion logic
- [x] Diet tag filtering

### Frontend ⏳
- [ ] RecipeFilters component
- [ ] Multi-select for allergens
- [ ] Multi-select for diet tags
- [ ] Slider for min score
- [ ] Apply filters button
- [ ] Clear filters button
- [ ] Save filter presets (user preferences)

---

## 🚀 Next Steps

1. **Frontend**: Создать RecipeFilters component
2. **UX**: Add filter chips (показывать активные фильтры)
3. **Analytics**: Track which filters used most
4. **Presets**: "Vegetarian", "Keto", "Quick meals" buttons
5. **User Preferences**: Save favorite filters in profile

---

## 📚 Related Docs

- `API_CONTRACT_RECIPE_MATCH.md` - Full API contract
- `RECIPE_CATALOG_QUICK_REF.md` - Catalog overview
- `migrations/035_create_recipe_catalog.sql` - Database schema
