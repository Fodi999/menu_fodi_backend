# canonicalName Quick Reference Card

## 🎯 TL;DR (для спешащих)

### Backend (Go):
```go
// Используй shared utility
import "github.com/.../pkg/utils"

canonicalName := utils.GenerateCanonicalName(title)
// "Яичница" → "scrambled_eggs"
```

### Frontend (TypeScript):
```typescript
interface Recipe {
  canonicalName: string;  // ← ДОБАВИТЬ
  titles?: { pl?: string; en?: string; ru?: string; };
}

// React key
<Card key={recipe.canonicalName} />

// URL
<Link href={`/recipes/${recipe.canonicalName}`} />

// Next.js routing
// app/recipes/[canonicalName]/page.tsx
```

---

## 📋 Быстрые ответы

### ❓ Зачем canonicalName?

**1 причина:** SEO-friendly URLs
- ❌ `/recipes/6b8628ef-ef1e-42eb-a166-924566bb9b7b`
- ✅ `/recipes/fried_salmon`

**2 причина:** Уникальность без UUID
- Стабильный идентификатор
- Не зависит от БД ID
- Работает на любом языке UI

**3 причина:** Backend-driven architecture
- Frontend НЕ решает как называть
- Backend - единственный источник истины
- AI и Admin используют одну логику

---

## 🔧 Что менять на фронтенде

### 1. Типы (5 минут)
```diff
interface Recipe {
  id: string;
+ canonicalName: string;
  category: string;
  // ... rest
}
```

### 2. React keys (2 минуты)
```diff
- key={recipe.id}
+ key={recipe.canonicalName}
```

### 3. URLs (5 минут)
```diff
- href={`/recipes/${recipe.id}`}
+ href={`/recipes/${recipe.canonicalName}`}
```

### 4. Routing (10 минут)
```diff
- app/recipes/[id]/page.tsx
+ app/recipes/[canonicalName]/page.tsx
```

**Итого: ~20 минут работы**

---

## 📊 Что уже работает (не трогать)

✅ Backend генерирует `canonicalName` автоматически
✅ БД нормализована (12 уникальных рецептов)
✅ API возвращает `canonicalName` в каждом рецепте
✅ Constraints в БД (UNIQUE + NOT NULL)
✅ Admin создаёт рецепты с правильными English slugs
✅ AI recommendations используют `canonicalName`

**Ничего настраивать на backend не нужно!**

---

## 🎨 Helper функции (copy-paste ready)

```typescript
// utils/recipe.ts

export function getRecipeTitle(recipe: Recipe, lang = 'en'): string {
  return recipe.titles?.[lang] || recipe.titles?.en || formatCanonical(recipe.canonicalName);
}

export function getRecipeUrl(recipe: Recipe): string {
  return `/recipes/${recipe.canonicalName}`;
}

function formatCanonical(slug: string): string {
  return slug.split('_').map(w => w[0].toUpperCase() + w.slice(1)).join(' ');
}
```

---

## 🚨 Частые ошибки

### ❌ НЕ делать так:
```typescript
// 1. НЕ генерировать canonicalName на фронте
const canonical = recipe.title.toLowerCase().replace(/ /g, '_'); // ❌

// 2. НЕ отправлять canonicalName в POST запросах
fetch('/api/recipes', {
  body: JSON.stringify({ canonicalName: 'my_recipe' }) // ❌
});

// 3. НЕ использовать localName или title для URLs
href={`/recipes/${recipe.localName}`} // ❌ (deprecated)
```

### ✅ ПРАВИЛЬНО:
```typescript
// 1. Backend генерирует canonicalName
const recipe = await fetch('/api/recipes/fried_salmon'); // ✅

// 2. Frontend только читает canonicalName
<Link href={`/recipes/${recipe.canonicalName}`} /> // ✅

// 3. Используй helper для отображения
{getRecipeTitle(recipe, currentLang)} // ✅
```

---

## 📞 API Endpoints

### Получить все рецепты:
```
GET /api/recipes
→ { data: { recipes: [{ canonicalName: "fried_salmon", ... }] }}
```

### Получить рецепт по canonical:
```
GET /api/recipes/fried_salmon
→ { data: { recipe: { canonicalName: "fried_salmon", ... }}}
```

### Создать рецепт (Admin):
```
POST /api/admin/recipes/create-ai
Authorization: Bearer <admin_token>
Body: { title: "Яичница", language: "ru", ... }
→ Backend сгенерирует canonicalName: "scrambled_eggs"
```

---

## 🎯 Миграция checklist

```markdown
Frontend Migration Checklist:

Step 1: Types
- [ ] Add `canonicalName: string` to Recipe interface

Step 2: Components
- [ ] Replace `key={recipe.id}` with `key={recipe.canonicalName}`
- [ ] Replace URLs: `/recipes/${id}` → `/recipes/${canonicalName}`
- [ ] Add helpers: getRecipeTitle(), getRecipeUrl()

Step 3: Routing
- [ ] Rename: [id] → [canonicalName] in route folders
- [ ] Update page component to use params.canonicalName
- [ ] Update SEO metadata to use canonicalName in URLs

Step 4: Features
- [ ] Update LocalStorage (favorites use canonicalName)
- [ ] Update Analytics (track by canonicalName)
- [ ] Update Search (search by canonicalName + titles)

Step 5: Testing
- [ ] Test URLs: /recipes/fried_salmon works
- [ ] Test SEO: og:url uses canonicalName
- [ ] Test favorites: save/load by canonicalName
- [ ] Test all languages: titles display correctly

Time estimate: 30-60 minutes
```

---

## 🔥 Production Example

```typescript
// Real working code (copy-paste ready)

import type { Recipe } from '@/types/recipe';

export default async function RecipePage({ 
  params 
}: { 
  params: { canonicalName: string } 
}) {
  // Fetch from API
  const res = await fetch(
    `https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/recipes/${params.canonicalName}`
  );
  const data = await res.json();
  const recipe: Recipe = data.data.recipe;
  
  // Display
  return (
    <article>
      <h1>{recipe.titles?.pl || recipe.canonicalName}</h1>
      <p>⏱ {recipe.timeMinutes} minut</p>
      <p>🍽 {recipe.servings} porcji</p>
    </article>
  );
}

// SEO
export async function generateMetadata({ params }: any) {
  const res = await fetch(`${API}/recipes/${params.canonicalName}`);
  const { data } = await res.json();
  
  return {
    title: data.recipe.titles?.pl,
    openGraph: {
      url: `https://menu-fodifood.vercel.app/recipes/${params.canonicalName}`,
    },
  };
}
```

---

## ✅ Done!

**Backend:** ✅ Working (nothing to change)  
**Frontend:** ⏳ 30 minutes to migrate  
**Result:** 🚀 SEO-friendly URLs + Production SaaS level
