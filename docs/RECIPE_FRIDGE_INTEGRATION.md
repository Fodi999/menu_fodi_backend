# Recipe-Fridge Integration Quick Guide

## 🎯 Problem

Recipe должен **реагировать** на изменения холодильника в реальном времени.

## ✅ Solution (Already Implemented!)

Backend **УЖЕ** поддерживает реактивность - каждый вызов `/api/recipes/match` использует актуальное состояние холодильника.

**Нет кеша, нет stale data** - всегда fresh!

---

## 🚀 Frontend Integration (3 Steps)

### Step 1: Install SWR

```bash
npm install swr
```

### Step 2: Create Hook

```typescript
// hooks/useRecipeMatches.ts
import useSWR from 'swr';

const fetcher = (url: string) => 
  fetch(url, {
    headers: { 'Authorization': `Bearer ${getToken()}` }
  }).then(r => r.json());

export const useRecipeMatches = (filters = {}) => {
  const params = new URLSearchParams(filters);
  const { data, error, mutate } = useSWR(
    `/api/recipes/match?${params}`,
    fetcher
  );

  return {
    recipes: data?.data?.recipes || [],
    loading: !data && !error,
    refresh: mutate,
  };
};
```

### Step 3: Use in Component

```typescript
// pages/recipes/index.tsx
import { useRecipeMatches } from '@/hooks/useRecipeMatches';
import { useFridge } from '@/hooks/useFridge';

export default function RecipesPage() {
  const { recipes, refresh } = useRecipeMatches({ minScore: 50 });
  const { addIngredient } = useFridge();

  const handleAddIngredient = async (item) => {
    await addIngredient(item);
    await refresh(); // ← Auto refresh recipes
  };

  return (
    <div>
      <h1>Recipes ({recipes.length})</h1>
      {recipes.map(recipe => (
        <RecipeCard 
          key={recipe.recipeId} 
          recipe={recipe}
          onAddMissing={handleAddIngredient}
        />
      ))}
    </div>
  );
}
```

---

## 📊 Live Update Example

### User Journey: Building Carbonara

```
Initial: GET /api/recipes/match
→ Carbonara: score=0, canCookNow=false

User adds: Makaron spaghetti (500g)
→ Frontend: await refresh()
→ Carbonara: score=25, canCookNow=false

User adds: Jajko (6 pcs)
→ Frontend: await refresh()
→ Carbonara: score=50, canCookNow=false

User adds: Boczek (200g)
→ Frontend: await refresh()
→ Carbonara: score=75, canCookNow=false

User adds: Parmezan (150g)
→ Frontend: await refresh()
→ Carbonara: score=100, canCookNow=true ✅
```

---

## 🎨 UI Components

### Progress Animation

```typescript
const RecipeCard = ({ recipe }) => {
  const prevScore = useRef(recipe.score);
  
  useEffect(() => {
    if (recipe.score > prevScore.current) {
      // Animate progress bar
      playProgressAnimation(prevScore.current, recipe.score);
    }
    prevScore.current = recipe.score;
  }, [recipe.score]);

  return (
    <Card>
      <ProgressBar 
        value={recipe.score} 
        max={100}
        animated={true}
      />
      
      {recipe.score === 100 && (
        <Confetti /> // 🎉 Celebrate completion!
      )}
    </Card>
  );
};
```

### Add Missing Button

```typescript
const AddMissingButton = ({ recipe, onAdd }) => {
  const handleClick = async () => {
    for (const ing of recipe.missingIngredients) {
      await onAdd({
        name: ing.name,
        quantity: ing.quantity,
        unit: ing.unit,
      });
    }
    toast.success('Dodano wszystkie składniki!');
  };

  return (
    <Button onClick={handleClick}>
      ➕ Dodaj {recipe.missingIngredients.length} składników
    </Button>
  );
};
```

---

## ⚡ Performance Tips

### 1. Debounce Multiple Adds

```typescript
const useDebouncedRefresh = () => {
  const { refresh } = useRecipeMatches();
  const timeoutRef = useRef(null);

  const debouncedRefresh = () => {
    if (timeoutRef.current) {
      clearTimeout(timeoutRef.current);
    }
    timeoutRef.current = setTimeout(() => {
      refresh();
    }, 500); // Wait 500ms after last change
  };

  return debouncedRefresh;
};
```

### 2. Batch Add Endpoint (Future)

```typescript
// POST /api/fridge/batch
await fetch('/api/fridge/batch', {
  method: 'POST',
  body: JSON.stringify({
    items: [
      { name: "Makaron", quantity: 500, unit: "g" },
      { name: "Jajko", quantity: 6, unit: "pcs" },
      { name: "Boczek", quantity: 200, unit: "g" },
    ]
  })
});

// Single refresh instead of 3
await refresh();
```

---

## 🔍 Debug Mode

```typescript
// Enable debug logs
const { recipes, refresh } = useRecipeMatches({ minScore: 50 });

useEffect(() => {
  console.log('🔄 Recipes updated:', {
    count: recipes.length,
    topScore: recipes[0]?.score,
    canCookNow: recipes.filter(r => r.canCookNow).length,
  });
}, [recipes]);
```

---

## ✅ Checklist

### Backend (Done ✅)
- [x] `/api/recipes/match` loads current fridge
- [x] No caching (always fresh)
- [x] Score calculation
- [x] canCookNow logic
- [x] Performance <150ms

### Frontend (TODO)
- [ ] Install SWR
- [ ] Create `useRecipeMatches` hook
- [ ] Call `refresh()` after fridge changes
- [ ] Add progress animation
- [ ] Add "Add missing ingredients" button
- [ ] Test full user journey

---

## 🎯 Success Metrics

- **API latency**: <100ms
- **User completion rate**: >60% (0% → 100%)
- **Catalog hit rate**: >70%
- **Time to cook**: <5 min from first ingredient

---

## 📚 Related Docs

- `API_CONTRACT_RECIPE_MATCH.md` - Full API contract
- `RECIPE_FRIDGE_REACTIVE_LOGIC.md` - Detailed logic explanation
- `RECIPE_CATALOG_QUICK_REF.md` - Catalog overview
