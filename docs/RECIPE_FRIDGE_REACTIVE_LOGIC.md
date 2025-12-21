# Recipe ↔ Fridge Reactive Logic

## 🎯 Цель

**Рецепт НЕ статичный** - он реагирует на изменения холодильника в реальном времени.

## 🔄 Flow: User добавляет ингредиент

### Старая логика (плохо):
```
1. User добавляет ингредиент в холодильник
2. User нажимает "Stwórz przepis"
3. AI генерирует НОВЫЙ рецепт
4. Economy пересчитывается
```

❌ **Проблемы**:
- Каждый раз новый рецепт
- AI тратит токены
- Нет consistency
- Медленно (3-5 сек)

### Новая логика (правильно):
```
1. User добавляет ингредиент в холодильник
2. Frontend автоматически вызывает GET /api/recipes/match
3. Backend пересчитывает score для существующих рецептов
4. UI обновляется:
   - missing → used
   - score ↑
   - canCookNow может стать true
```

✅ **Преимущества**:
- Instant update (50-150ms)
- Нет AI токенов
- Consistency
- UX: видно прогресс

---

## 📡 WebSocket vs Polling vs Manual Refresh

### Option 1: Manual Refresh (MVP)
**Что**: User нажимает "Odśwież" после добавления ингредиента.

```typescript
// Frontend
const refreshRecipes = async () => {
  const matches = await recipeApi.match({ minScore: 50 });
  setRecipes(matches);
};

// After adding ingredient
await fridgeApi.addIngredient({ name: "Jajko", quantity: 4 });
await refreshRecipes(); // Manual call
```

**Pros**:
- Простая реализация
- Нет polling overhead

**Cons**:
- User должен помнить обновить
- Не automatic

---

### Option 2: Auto Refresh on Fridge Change (Recommended)
**Что**: Frontend автоматически вызывает `/recipes/match` после каждого изменения холодильника.

```typescript
// Frontend hook
const useFridgeSync = () => {
  const { data: recipes, mutate } = useSWR('/api/recipes/match');
  
  const addIngredient = async (ingredient) => {
    await fridgeApi.addIngredient(ingredient);
    await mutate(); // Auto refresh recipes
  };
  
  const removeIngredient = async (id) => {
    await fridgeApi.removeIngredient(id);
    await mutate(); // Auto refresh recipes
  };
  
  return { recipes, addIngredient, removeIngredient };
};
```

**Pros**:
- Automatic
- No WebSocket complexity
- Works с SWR/React Query caching

**Cons**:
- Extra API call на каждое изменение

---

### Option 3: WebSocket Real-time (Future)
**Что**: Backend pushes updates через WebSocket.

```typescript
// Backend WebSocket
ws.on('fridge:updated', (userID) => {
  const matches = await matchService.MatchRecipesWithFridge(userID);
  ws.emit(`user:${userID}:recipes:updated`, matches);
});

// Frontend
ws.on('recipes:updated', (matches) => {
  setRecipes(matches);
});
```

**Pros**:
- Real-time
- No polling

**Cons**:
- Complex setup
- Overkill для MVP

---

## 🎨 Frontend Implementation

### 1. SWR with Auto Refresh (Recommended)

```typescript
// hooks/useRecipeMatches.ts
import useSWR from 'swr';

export const useRecipeMatches = (filters?: RecipeFilters) => {
  const params = new URLSearchParams(filters);
  
  const { data, error, mutate } = useSWR(
    `/api/recipes/match?${params}`,
    fetcher,
    {
      revalidateOnFocus: false,
      revalidateOnReconnect: false,
      dedupingInterval: 5000, // Don't refetch within 5 sec
    }
  );
  
  const refresh = () => mutate();
  
  return {
    recipes: data?.data?.recipes || [],
    count: data?.data?.count || 0,
    loading: !data && !error,
    error,
    refresh,
  };
};
```

### 2. Fridge Context with Auto Sync

```typescript
// context/FridgeContext.tsx
import { createContext, useContext } from 'react';
import { useRecipeMatches } from '@/hooks/useRecipeMatches';

interface FridgeContextValue {
  items: FridgeItem[];
  recipes: RecipeMatchItem[];
  addIngredient: (item: FridgeItem) => Promise<void>;
  removeIngredient: (id: string) => Promise<void>;
  refreshRecipes: () => void;
}

const FridgeContext = createContext<FridgeContextValue>(null);

export const FridgeProvider: React.FC = ({ children }) => {
  const [items, setItems] = useState<FridgeItem[]>([]);
  const { recipes, refresh: refreshRecipes } = useRecipeMatches();
  
  const addIngredient = async (item: FridgeItem) => {
    await fridgeApi.addIngredient(item);
    
    // Refresh fridge
    const updated = await fridgeApi.getAll();
    setItems(updated);
    
    // Auto refresh recipes
    refreshRecipes();
    
    toast.success(`Dodano ${item.name}`);
  };
  
  const removeIngredient = async (id: string) => {
    await fridgeApi.removeIngredient(id);
    
    setItems(items.filter(i => i.id !== id));
    refreshRecipes(); // Auto refresh
    
    toast.success('Usunięto składnik');
  };
  
  return (
    <FridgeContext.Provider value={{
      items,
      recipes,
      addIngredient,
      removeIngredient,
      refreshRecipes,
    }}>
      {children}
    </FridgeContext.Provider>
  );
};

export const useFridge = () => useContext(FridgeContext);
```

### 3. Recipe Card with Live Updates

```typescript
// components/RecipeCard.tsx
import { useFridge } from '@/context/FridgeContext';

const RecipeCard: React.FC<{ recipe: RecipeMatchItem }> = ({ recipe }) => {
  const { addIngredient } = useFridge();
  
  const addMissingIngredients = async () => {
    for (const ing of recipe.missingIngredients) {
      await addIngredient({
        ingredientId: ing.ingredientId,
        name: ing.name,
        quantity: ing.quantity,
        unit: ing.unit,
      });
    }
    
    // Recipes will auto-refresh via FridgeContext
    toast.success('Dodano wszystkie brakujące składniki');
  };
  
  return (
    <Card>
      <CardHeader>
        <Title>{recipe.localName}</Title>
        <Badge color={recipe.canCookNow ? 'green' : 'yellow'}>
          {recipe.score.toFixed(0)}% Match
        </Badge>
      </CardHeader>
      
      <CardBody>
        {/* Progress bar that updates live */}
        <ProgressBar value={recipe.score} max={100} />
        
        {recipe.missingIngredients.length > 0 && (
          <Box>
            <Text>Brakuje {recipe.missingIngredients.length} składników:</Text>
            <List>
              {recipe.missingIngredients.map(ing => (
                <ListItem key={ing.ingredientId}>
                  {ing.name} - {ing.quantity} {ing.unit}
                </ListItem>
              ))}
            </List>
          </Box>
        )}
      </CardBody>
      
      <CardFooter>
        {recipe.canCookNow ? (
          <Button color="green">🍳 Gotuj teraz</Button>
        ) : (
          <Button color="yellow" onClick={addMissingIngredients}>
            ➕ Dodaj brakujące składniki
          </Button>
        )}
      </CardFooter>
    </Card>
  );
};
```

---

## 🔄 Lifecycle Example

### Scenario: User buduje Carbonara krok po kroku

**Initial state** (холодильник пуст):
```json
GET /api/recipes/match

{
  "recipes": [
    {
      "canonicalName": "Spaghetti Carbonara",
      "score": 0,
      "coverage": 0,
      "canCookNow": false,
      "usedIngredients": [],
      "missingIngredients": [
        { "name": "Makaron spaghetti", "quantity": 400, "unit": "g" },
        { "name": "Jajko", "quantity": 4, "unit": "pcs" },
        { "name": "Boczek", "quantity": 150, "unit": "g" },
        { "name": "Parmezan", "quantity": 100, "unit": "g" }
      ]
    }
  ]
}
```

---

**Step 1**: User добавляет 500g makaronu
```typescript
await addIngredient({ name: "Makaron spaghetti", quantity: 500, unit: "g" });
// Auto triggers: GET /api/recipes/match
```

**Response**:
```json
{
  "recipes": [
    {
      "canonicalName": "Spaghetti Carbonara",
      "score": 25,  // ⬆️ was 0
      "coverage": 0.25,  // ⬆️ was 0
      "canCookNow": false,
      "usedIngredients": [
        { "name": "Makaron spaghetti", "quantity": 400, "unit": "g", "available": 500 }
      ],
      "missingIngredients": [
        { "name": "Jajko", "quantity": 4, "unit": "pcs" },
        { "name": "Boczek", "quantity": 150, "unit": "g" },
        { "name": "Parmezan", "quantity": 100, "unit": "g" }
      ]
    }
  ]
}
```

**UI Update**: Progress bar 0% → 25%

---

**Step 2**: User добавляет 6 jajek
```typescript
await addIngredient({ name: "Jajko", quantity: 6, unit: "pcs" });
```

**Response**:
```json
{
  "recipes": [
    {
      "score": 50,  // ⬆️ was 25
      "coverage": 0.5,
      "canCookNow": false,
      "usedIngredients": [
        { "name": "Makaron spaghetti", "quantity": 400, "unit": "g", "available": 500 },
        { "name": "Jajko", "quantity": 4, "unit": "pcs", "available": 6 }
      ],
      "missingIngredients": [
        { "name": "Boczek", "quantity": 150, "unit": "g" },
        { "name": "Parmezan", "quantity": 100, "unit": "g" }
      ]
    }
  ]
}
```

**UI Update**: Progress bar 25% → 50%

---

**Step 3**: User добавляет 200g boczku
```typescript
await addIngredient({ name: "Boczek", quantity: 200, unit: "g" });
```

**Response**:
```json
{
  "recipes": [
    {
      "score": 75,  // ⬆️ was 50
      "coverage": 0.75,
      "canCookNow": false,
      "usedIngredients": [
        { "name": "Makaron spaghetti", "quantity": 400, "unit": "g", "available": 500 },
        { "name": "Jajko", "quantity": 4, "unit": "pcs", "available": 6 },
        { "name": "Boczek", "quantity": 150, "unit": "g", "available": 200 }
      ],
      "missingIngredients": [
        { "name": "Parmezan", "quantity": 100, "unit": "g" }
      ]
    }
  ]
}
```

**UI Update**: Progress bar 50% → 75%, badge "🟡 1 składnik brakuje"

---

**Step 4**: User добавляет 150g parmezanu
```typescript
await addIngredient({ name: "Parmezan", quantity: 150, unit: "g" });
```

**Response**:
```json
{
  "recipes": [
    {
      "score": 100,  // ⬆️ was 75
      "coverage": 1.0,
      "canCookNow": true,  // ✅ Changed!
      "usedIngredients": [
        { "name": "Makaron spaghetti", "quantity": 400, "unit": "g", "available": 500 },
        { "name": "Jajko", "quantity": 4, "unit": "pcs", "available": 6 },
        { "name": "Boczek", "quantity": 150, "unit": "g", "available": 200 },
        { "name": "Parmezan", "quantity": 100, "unit": "g", "available": 150 }
      ],
      "missingIngredients": []
    }
  ]
}
```

**UI Update**: 
- Progress bar 75% → 100% ✅
- Badge "🟢 Gotowe do gotowania!"
- Button changes: "➕ Dodaj składniki" → "🍳 Gotuj teraz"

---

## 🎯 Backend: Already Working!

**Current implementation** уже поддерживает это:

```go
// internal/modules/recipes/service/match_service.go

func (s *RecipeMatchService) MatchRecipesWithFridge(userID string, filters RecipeFilters) ([]RecipeMatch, error) {
    // 1. Load CURRENT fridge state
    fridgeItems, _ := s.loadFridgeWithPrices(userID)
    
    // 2. Load recipes from catalog
    recipes, _ := s.loadRecipesWithFilters(filters)
    
    // 3. Calculate match score for CURRENT fridge
    for _, recipe := range recipes {
        match := s.calculateRecipeMatch(recipe, fridgeMap)
        matches = append(matches, match)
    }
    
    return matches, nil
}
```

✅ **Каждый вызов** `/api/recipes/match` использует **актуальное** состояние холодильника.

Нет кеширования, нет stale data - всегда fresh!

---

## 📊 Performance Considerations

### Query Optimization
```go
// Efficient fridge loading
func (s *RecipeMatchService) loadFridgeWithPrices(userID string) ([]FridgeItem, error) {
    var items []models.UserFridgeItem
    
    // Single query with JOIN
    err := s.db.
        Preload("Ingredient").  // Eager load to avoid N+1
        Where("user_id = ?", userID).
        Find(&items).Error
    
    // Fast map lookup O(1)
    fridgeMap := make(map[string]*FridgeItem)
    for _, item := range items {
        key := normalizeIngredientName(item.Ingredient.Name)
        fridgeMap[key] = &FridgeItem{...}
    }
    
    return items, nil
}
```

**Timing**:
- Load fridge: ~10-20ms
- Load recipes: ~20-30ms
- Calculate matches: ~10-20ms
- **Total: ~50-150ms** ✅

---

## ✅ Implementation Checklist

### Backend (Already Done ✅)
- [x] `/api/recipes/match` loads current fridge state
- [x] Score calculation based on current ingredients
- [x] `canCookNow` updates automatically
- [x] `missingIngredients` vs `usedIngredients` split
- [x] No caching (always fresh data)

### Frontend (TODO)
- [ ] Create `useRecipeMatches` hook with SWR
- [ ] Create `FridgeContext` with auto-refresh
- [ ] Call `refreshRecipes()` after every fridge change
- [ ] Update UI: progress bars, badges, buttons
- [ ] Add "Dodaj brakujące składniki" button
- [ ] Show live updates (0% → 25% → 50% → 100%)

---

## 🚀 Next Steps

1. **Frontend integration**:
   ```bash
   # Install SWR
   npm install swr
   
   # Create hooks/useRecipeMatches.ts
   # Create context/FridgeContext.tsx
   # Update RecipeCard component
   ```

2. **Test flow**:
   - Start with empty fridge
   - Add ingredients one by one
   - Watch score increase in real-time
   - See `canCookNow` change from false → true

3. **UX improvements**:
   - Show progress animation on ingredient add
   - Confetti when `canCookNow` becomes true
   - Badge "🔥 {N} składników brakuje"

---

## 📈 Metrics to Track

- **Catalog hit rate**: % times user found recipe without AI
- **Average completion rate**: How many users go from 0% → 100%
- **Time to cook**: How long from first ingredient to `canCookNow`
- **API latency**: `/api/recipes/match` response time

**Target**: 80% catalog hit rate, <100ms API latency
