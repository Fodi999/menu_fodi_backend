# Frontend Setup Guide for canonicalName

## 🎯 Что изменилось в API

### ✅ Новое поле в Recipe:
```json
{
  "id": "6b8628ef-ef1e-42eb-a166-924566bb9b7b",
  "canonicalName": "fried_salmon",  // ← НОВОЕ (English slug)
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

## 📦 Шаг 1: Обновить TypeScript типы

### Добавить в `types/recipe.ts`:

```typescript
export interface Recipe {
  id: string;
  canonicalName: string; // ← ДОБАВИТЬ это поле
  category: 'main' | 'salad' | 'soup' | 'dessert' | 'breakfast' | 'snack';
  country: string;
  difficulty: 'easy' | 'medium' | 'hard';
  timeMinutes: number;
  servings: number;
  
  titles?: {
    pl?: string;
    en?: string;
    ru?: string;
  };
  
  descriptions?: {
    pl?: string;
    en?: string;
    ru?: string;
  };
  
  // Optional fields
  ingredients?: RecipeIngredient[];
  createdAt?: string;
  updatedAt?: string;
}
```

---

## 🔧 Шаг 2: Обновить существующий код

### ❌ БЫЛО (старый способ):
```typescript
// Использование localName или title (нестабильно)
<RecipeCard key={recipe.id} />  // ID может меняться
<a href={`/recipes/${recipe.id}`}>Recipe</a>  // Не SEO-friendly
```

### ✅ СТАЛО (новый способ):
```typescript
// Используем canonicalName (стабильный, SEO-friendly)
<RecipeCard key={recipe.canonicalName} />  // Уникальный slug
<a href={`/recipes/${recipe.canonicalName}`}>Recipe</a>  // /recipes/fried_salmon
```

---

## 🎨 Шаг 3: Helper функции (utils/recipe.ts)

```typescript
/**
 * Получить название на текущем языке
 */
export function getRecipeTitle(
  recipe: Recipe, 
  lang: 'pl' | 'en' | 'ru' = 'en'
): string {
  return recipe.titles?.[lang] 
    || recipe.titles?.en 
    || formatCanonicalName(recipe.canonicalName);
}

/**
 * Форматировать canonicalName для UI
 * "fried_salmon" → "Fried Salmon"
 */
function formatCanonicalName(canonical: string): string {
  return canonical
    .split('_')
    .map(word => word.charAt(0).toUpperCase() + word.slice(1))
    .join(' ');
}

/**
 * Получить URL рецепта (SEO-friendly)
 */
export function getRecipeUrl(recipe: Recipe): string {
  return `/recipes/${recipe.canonicalName}`;
}
```

---

## 🔗 Шаг 4: Routing (Next.js)

### App Router: `app/recipes/[canonicalName]/page.tsx`

```typescript
interface RecipePageProps {
  params: {
    canonicalName: string; // ← URL: /recipes/fried_salmon
  };
}

export default async function RecipePage({ params }: RecipePageProps) {
  const recipe = await fetchRecipe(params.canonicalName);
  
  return (
    <div>
      <h1>{getRecipeTitle(recipe, 'pl')}</h1>
      <p>{recipe.timeMinutes} minut</p>
    </div>
  );
}
```

### SEO Metadata:
```typescript
export async function generateMetadata({ params }: RecipePageProps) {
  const recipe = await fetchRecipe(params.canonicalName);
  
  return {
    title: getRecipeTitle(recipe, 'pl'),
    description: recipe.descriptions?.pl,
    openGraph: {
      url: `https://menu-fodifood.vercel.app/recipes/${recipe.canonicalName}`,
    },
  };
}
```

---

## 🔍 Шаг 5: API Client

```typescript
const API_BASE = 'https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api';

/**
 * Получить все рецепты
 */
export async function getRecipes(): Promise<Recipe[]> {
  const res = await fetch(`${API_BASE}/recipes`);
  const data = await res.json();
  return data.data.recipes;
}

/**
 * Получить рецепт по canonicalName
 */
export async function getRecipe(canonicalName: string): Promise<Recipe> {
  const res = await fetch(`${API_BASE}/recipes/${canonicalName}`);
  const data = await res.json();
  return data.data.recipe;
}
```

---

## 📊 Шаг 6: Примеры использования

### 1. Список рецептов:
```typescript
function RecipeList({ recipes, lang }: Props) {
  return (
    <div className="grid">
      {recipes.map(recipe => (
        <RecipeCard
          key={recipe.canonicalName}  // ← Уникальный key
          recipe={recipe}
          lang={lang}
        />
      ))}
    </div>
  );
}
```

### 2. Карточка рецепта:
```typescript
function RecipeCard({ recipe, lang }: Props) {
  const title = getRecipeTitle(recipe, lang);
  const url = getRecipeUrl(recipe);
  
  return (
    <a href={url} className="recipe-card">
      <h3>{title}</h3>
      <p>{recipe.timeMinutes} min</p>
    </a>
  );
}
```

### 3. Favorites (LocalStorage):
```typescript
function addToFavorites(recipe: Recipe) {
  const favorites = JSON.parse(localStorage.getItem('favorites') || '[]');
  
  if (!favorites.includes(recipe.canonicalName)) {
    favorites.push(recipe.canonicalName);  // ← Храним только canonical
    localStorage.setItem('favorites', JSON.stringify(favorites));
  }
}

function getFavoriteRecipes(allRecipes: Recipe[]): Recipe[] {
  const canonicals = JSON.parse(localStorage.getItem('favorites') || '[]');
  return allRecipes.filter(r => canonicals.includes(r.canonicalName));
}
```

### 4. Search:
```typescript
function searchRecipes(recipes: Recipe[], query: string, lang: string) {
  return recipes.filter(recipe => {
    const title = getRecipeTitle(recipe, lang);
    return title.toLowerCase().includes(query.toLowerCase());
  });
}
```

---

## 🎯 Шаг 7: Миграция старого кода

### Найти и заменить:

1. **React keys:**
   ```typescript
   // БЫЛО:
   key={recipe.id}
   
   // СТАЛО:
   key={recipe.canonicalName}
   ```

2. **URLs:**
   ```typescript
   // БЫЛО:
   href={`/recipes/${recipe.id}`}
   
   // СТАЛО:
   href={`/recipes/${recipe.canonicalName}`}
   ```

3. **LocalStorage/Favorites:**
   ```typescript
   // БЫЛО:
   favorites.push(recipe.id)
   
   // СТАЛО:
   favorites.push(recipe.canonicalName)
   ```

4. **Analytics:**
   ```typescript
   // БЫЛО:
   trackEvent('view_recipe', { recipe_id: recipe.id })
   
   // СТАЛО:
   trackEvent('view_recipe', { recipe_id: recipe.canonicalName })
   ```

---

## ✅ Проверочный чеклист

- [ ] Обновлен тип `Recipe` (добавлено `canonicalName: string`)
- [ ] Созданы helper функции (`getRecipeTitle`, `getRecipeUrl`)
- [ ] Обновлены React keys (`key={recipe.canonicalName}`)
- [ ] Обновлены URLs (`/recipes/${canonicalName}`)
- [ ] Обновлен routing (Next.js `[canonicalName]/page.tsx`)
- [ ] Обновлен SEO metadata (используется `canonicalName`)
- [ ] Обновлен LocalStorage (favorites хранят `canonicalName`)
- [ ] Обновлена аналитика (треки используют `canonicalName`)
- [ ] Протестированы старые ссылки (redirect old URLs если нужно)

---

## 🚀 Деплой

После обновления фронтенда:

1. **Протестировать локально:**
   ```bash
   npm run dev
   # Проверить /recipes/fried_salmon
   ```

2. **Проверить API:**
   ```bash
   curl https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/recipes/fried_salmon
   ```

3. **Deploy:**
   ```bash
   git add .
   git commit -m "feat: use canonicalName for recipes (SEO-friendly URLs)"
   git push
   ```

---

## 📝 Дополнительно

### Redirect старых URL (опционально):

Если были `/recipes/[id]`, добавить redirect:

```typescript
// middleware.ts или getServerSideProps
if (isUUID(params.slug)) {
  // Old URL with UUID
  const recipe = await getRecipeById(params.slug);
  return {
    redirect: {
      destination: `/recipes/${recipe.canonicalName}`,
      permanent: true,
    },
  };
}
```

---

## 🎉 Готово!

Теперь фронтенд использует `canonicalName` для:
- ✅ SEO-friendly URLs (`/recipes/fried_salmon`)
- ✅ Стабильные React keys
- ✅ Уникальные идентификаторы
- ✅ Мультиязычность (через `titles`)
- ✅ Analytics & Tracking
- ✅ LocalStorage (favorites)

**Backend генерирует canonical names → Frontend использует их.**

Production SaaS уровень! 🚀
