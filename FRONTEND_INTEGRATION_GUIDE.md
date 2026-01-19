# Frontend Integration Guide - canonicalName

## 🎯 Что изменилось в API

### ✅ Новое поле в Recipe:
```json
{
  "id": "6b8628ef-...",
  "canonicalName": "fried_salmon",  // ← NEW: English slug
  "category": "main",
  "difficulty": "easy",
  "timeMinutes": 11,
  "servings": 1,
  "titles": {
    "pl": "Łosoś smażony",
    "en": "Fried Salmon",
    "ru": "Жареный лосось"
  }
}
```

---

## 📝 Шаг 1: Обновить TypeScript типы

Скопируйте файл `FRONTEND_RECIPE_TYPES.ts` в ваш фронтенд проект:

```bash
# Из backend/docs/ скопировать в frontend/src/types/
cp docs/FRONTEND_RECIPE_TYPES.ts ../frontend/src/types/recipe.ts
```

Или вручную добавьте поле:

```typescript
export interface Recipe {
  id: string;
  canonicalName: string; // ← ДОБАВИТЬ ЭТО ПОЛЕ
  category: string;
  // ... остальные поля
}
```

---

## 📝 Шаг 2: Обновить API запросы (если нужно)

### Next.js App Router

```typescript
// app/recipes/page.tsx
import { Recipe } from '@/types/recipe';

async function getRecipes(): Promise<Recipe[]> {
  const res = await fetch('https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/recipes');
  const json = await res.json();
  return json.data.recipes; // canonicalName уже включён
}

export default async function RecipesPage() {
  const recipes = await getRecipes();
  
  return (
    <div>
      {recipes.map(recipe => (
        // ✅ Используйте canonicalName как key
        <RecipeCard key={recipe.canonicalName} recipe={recipe} />
      ))}
    </div>
  );
}
```

### React (SPA)

```typescript
// hooks/useRecipes.ts
import { useEffect, useState } from 'react';
import { Recipe } from '@/types/recipe';

export function useRecipes() {
  const [recipes, setRecipes] = useState<Recipe[]>([]);
  const [loading, setLoading] = useState(true);
  
  useEffect(() => {
    fetch('https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/recipes')
      .then(res => res.json())
      .then(json => {
        setRecipes(json.data.recipes); // canonicalName уже включён
        setLoading(false);
      });
  }, []);
  
  return { recipes, loading };
}
```

---

## 📝 Шаг 3: Использовать canonicalName для URLs (SEO)

### Next.js Dynamic Routes

```typescript
// app/recipes/[canonicalName]/page.tsx
import { Recipe, getRecipeTitle } from '@/types/recipe';

async function getRecipe(canonicalName: string): Promise<Recipe> {
  const res = await fetch(
    `https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/recipes/${canonicalName}`
  );
  const json = await res.json();
  return json.data.recipe;
}

export default async function RecipePage({ 
  params 
}: { 
  params: { canonicalName: string } 
}) {
  const recipe = await getRecipe(params.canonicalName);
  const title = getRecipeTitle(recipe, 'pl'); // или 'en', 'ru'
  
  return (
    <article>
      <h1>{title}</h1>
      {/* ... */}
    </article>
  );
}

// ✅ SEO-friendly URLs:
// /recipes/fried_salmon
// /recipes/scrambled_eggs
// /recipes/pasta_carbonara
```

### React Router

```typescript
// Router.tsx
import { BrowserRouter, Routes, Route } from 'react-router-dom';

<BrowserRouter>
  <Routes>
    <Route path="/recipes" element={<RecipesList />} />
    <Route path="/recipes/:canonicalName" element={<RecipeDetail />} />
  </Routes>
</BrowserRouter>

// RecipeDetail.tsx
import { useParams } from 'react-router-dom';

export function RecipeDetail() {
  const { canonicalName } = useParams();
  
  // Fetch recipe by canonicalName
  const recipe = useRecipe(canonicalName);
  
  return <div>{/* ... */}</div>;
}
```

---

## 📝 Шаг 4: Отображать название на текущем языке

```typescript
// components/RecipeCard.tsx
import { Recipe, getRecipeTitle } from '@/types/recipe';
import { useLanguage } from '@/hooks/useLanguage';

export function RecipeCard({ recipe }: { recipe: Recipe }) {
  const { currentLang } = useLanguage(); // 'pl' | 'en' | 'ru'
  const title = getRecipeTitle(recipe, currentLang);
  
  return (
    <div>
      <h3>{title}</h3>
      {/* canonicalName для внутренней логики, не показываем */}
      <Link href={`/recipes/${recipe.canonicalName}`}>
        View Recipe
      </Link>
    </div>
  );
}
```

---

## 📝 Шаг 5: Генерация списка рецептов (если нужно)

### Static Site Generation (SSG)

```typescript
// Next.js
export async function generateStaticParams() {
  const res = await fetch('https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/recipes');
  const json = await res.json();
  
  return json.data.recipes.map((recipe: Recipe) => ({
    canonicalName: recipe.canonicalName, // ✅ Уникальный путь
  }));
}

// Generates:
// /recipes/fried_salmon
// /recipes/scrambled_eggs
// /recipes/pasta_carbonara
// ... (12 страниц)
```

---

## ⚠️ ВАЖНО: Что НЕ делать

### ❌ НЕ генерируйте canonicalName на фронтенде

```typescript
// ❌ НЕПРАВИЛЬНО
const canonicalName = recipe.title.toLowerCase().replace(/ /g, '_');

// ✅ ПРАВИЛЬНО
const canonicalName = recipe.canonicalName; // Уже есть от backend
```

### ❌ НЕ используйте id для URLs

```typescript
// ❌ НЕПРАВИЛЬНО (плохо для SEO)
<Link href={`/recipes/${recipe.id}`}>

// ✅ ПРАВИЛЬНО (SEO-friendly)
<Link href={`/recipes/${recipe.canonicalName}`}>
```

### ❌ НЕ полагайтесь на localName или title

```typescript
// ❌ DEPRECATED (будет удалено)
recipe.localName
recipe.title

// ✅ ПРАВИЛЬНО
recipe.canonicalName  // для URLs, keys
recipe.titles[lang]   // для отображения
```

---

## 🎯 Best Practices

### 1. Используйте canonicalName для React keys

```typescript
// ✅ Уникальный и стабильный key
{recipes.map(recipe => (
  <RecipeCard key={recipe.canonicalName} recipe={recipe} />
))}
```

### 2. Используйте titles[lang] для отображения

```typescript
// ✅ Показываем локализованное название
const displayName = recipe.titles?.pl || recipe.canonicalName;
```

### 3. Используйте canonicalName для навигации

```typescript
// ✅ SEO-friendly URL
const url = `/recipes/${recipe.canonicalName}`;
```

---

## 📊 Производительность API

После оптимизации (18 января 2026):

```
GET /api/recipes                    0.064 ms  ✅
GET /api/recipes?category=main      0.060 ms  ✅
GET /api/recipes/:canonicalName     0.041 ms  ✅
GET /api/recipes/stats              0.036 ms  ✅
```

**Индексы добавлены:**
- `idx_recipe_created_at_desc` — для сортировки
- `idx_recipe_category` — для фильтрации
- `idx_recipe_category_created_at` — для комбинированных запросов
- `Recipe_canonicalName_unique` — для поиска по slug (UNIQUE)

---

## 🔧 Примеры использования

### Список рецептов

```typescript
function RecipesList() {
  const { recipes } = useRecipes();
  const lang = useLanguage();
  
  return (
    <ul>
      {recipes.map(recipe => (
        <li key={recipe.canonicalName}>
          <Link href={`/recipes/${recipe.canonicalName}`}>
            {getRecipeTitle(recipe, lang)}
          </Link>
        </li>
      ))}
    </ul>
  );
}
```

### Детали рецепта

```typescript
function RecipeDetail({ canonicalName }: { canonicalName: string }) {
  const recipe = useRecipe(canonicalName);
  const lang = useLanguage();
  
  if (!recipe) return <div>Loading...</div>;
  
  return (
    <article>
      <h1>{getRecipeTitle(recipe, lang)}</h1>
      <p>{getRecipeDescription(recipe, lang)}</p>
      
      <meta property="og:url" content={`/recipes/${recipe.canonicalName}`} />
      <meta property="og:title" content={getRecipeTitle(recipe, 'en')} />
    </article>
  );
}
```

### Фильтрация по категории

```typescript
function RecipesByCategory({ category }: { category: string }) {
  const url = `https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/recipes?category=${category}`;
  const { recipes } = useFetch<Recipe[]>(url);
  
  return (
    <div>
      <h2>{category} Recipes</h2>
      {recipes.map(recipe => (
        <RecipeCard key={recipe.canonicalName} recipe={recipe} />
      ))}
    </div>
  );
}
```

---

## ✅ Итого

После интеграции у вас будет:

1. ✅ SEO-friendly URLs (`/recipes/fried_salmon`)
2. ✅ Уникальные React keys (`canonicalName`)
3. ✅ Локализованные названия (`titles[lang]`)
4. ✅ Стабильные идентификаторы (не меняются при переводе)
5. ✅ Готовность к масштабированию (индексы, оптимизация)

**Backend генерирует canonical names, Frontend использует их!**
