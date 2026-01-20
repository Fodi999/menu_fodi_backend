# 📦 Ingredient Categories API - Frontend Integration Guide

## 🎯 Цель

Перенести категории продуктов из хардкода на фронтенде в **справочник** с бэкенда.

### ❌ БЫЛО: Категории "выдумываются" на фронте

```typescript
// ❌ Плохо - хардкод на фронте
const CATEGORY_MAP_PL: Record<string, string> = {
  fish: "Ryby",
  meat: "Mięso",
  egg: "Jajka",
  // ...
};
```

**Проблемы:**
- ❌ Нет списка всех доступных категорий с бэкенда
- ❌ Нельзя гарантировать порядок отображения
- ❌ Нельзя сменить язык без передеплоя фронта
- ❌ Нельзя добавить новую категорию без обновления фронта

### ✅ СТАЛО: Backend явно отдаёт каталог категорий

```typescript
// ✅ Хорошо - категории из API
const categories = await fetchCategories('pl');
// [
//   {key: "all", label: "Wszystkie", icon: "🧊", sortOrder: 0},
//   {key: "fish", label: "Ryby", icon: "🐟", sortOrder: 1},
//   {key: "meat", label: "Mięso", icon: "🥩", sortOrder: 2},
//   ...
// ]
```

**Преимущества:**
- ✅ Backend явно отдаёт список категорий
- ✅ Гарантирован порядок сортировки (sortOrder)
- ✅ Смена языка = смена Accept-Language заголовка
- ✅ Новая категория = добавляем в БД, фронт сам подхватит

---

## 🔌 API Endpoint

### `GET /api/catalog/ingredient-categories`

**Авторизация:** Bearer Token (JWT)  
**Локализация:** Заголовок `Accept-Language: pl | en | ru`

#### Request Example

```bash
curl -X GET \
  'https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/catalog/ingredient-categories' \
  -H 'Authorization: Bearer YOUR_JWT_TOKEN' \
  -H 'Accept-Language: pl'
```

#### Response Example (Accept-Language: pl)

```json
{
  "success": true,
  "data": {
    "categories": [
      {
        "key": "all",
        "label": "Wszystko",
        "icon": "🧊",
        "sortOrder": 0
      },
      {
        "key": "fish",
        "label": "Ryby",
        "icon": "🐟",
        "sortOrder": 1
      },
      {
        "key": "meat",
        "label": "Mięso",
        "icon": "🥩",
        "sortOrder": 2
      },
      {
        "key": "egg",
        "label": "Jajka",
        "icon": "🥚",
        "sortOrder": 3
      },
      {
        "key": "dairy",
        "label": "Nabiał",
        "icon": "🥛",
        "sortOrder": 4
      },
      {
        "key": "vegetable",
        "label": "Warzywa",
        "icon": "🥕",
        "sortOrder": 5
      },
      {
        "key": "fruit",
        "label": "Owoce",
        "icon": "🍎",
        "sortOrder": 6
      },
      {
        "key": "grain",
        "label": "Zboża",
        "icon": "🌾",
        "sortOrder": 7
      },
      {
        "key": "condiment",
        "label": "Przyprawy",
        "icon": "🧂",
        "sortOrder": 8
      },
      {
        "key": "other",
        "label": "Inne",
        "icon": "📦",
        "sortOrder": 9
      }
    ]
  }
}
```

#### Response Example (Accept-Language: en)

```json
{
  "success": true,
  "data": {
    "categories": [
      {
        "key": "all",
        "label": "All",
        "icon": "🧊",
        "sortOrder": 0
      },
      {
        "key": "fish",
        "label": "Fish",
        "icon": "🐟",
        "sortOrder": 1
      },
      {
        "key": "meat",
        "label": "Meat",
        "icon": "🥩",
        "sortOrder": 2
      }
      // ...
    ]
  }
}
```

#### Response Example (Accept-Language: ru)

```json
{
  "success": true,
  "data": {
    "categories": [
      {
        "key": "all",
        "label": "Все",
        "icon": "🧊",
        "sortOrder": 0
      },
      {
        "key": "fish",
        "label": "Рыба",
        "icon": "🐟",
        "sortOrder": 1
      },
      {
        "key": "meat",
        "label": "Мясо",
        "icon": "🥩",
        "sortOrder": 2
      }
      // ...
    ]
  }
}
```

---

## 🔑 Response Schema

| Field        | Type   | Description                                          |
|-------------|--------|------------------------------------------------------|
| `key`       | string | Стабильный идентификатор категории (fish, meat, dairy) |
| `label`     | string | Локализованное название (зависит от Accept-Language) |
| `icon`      | string | Emoji-иконка (🐟, 🥩, 🥛)                              |
| `sortOrder` | int    | Порядок отображения (0 = первая, 9 = последняя)     |

### ⚠️ ВАЖНО: Backend делает локализацию

**Backend отдаёт ТОЛЬКО локализованный `label`**, а не все 3 языка.

❌ **НЕ будет такого:**
```json
{
  "key": "fish",
  "label_pl": "Ryby",
  "label_en": "Fish",
  "label_ru": "Рыба"
}
```

✅ **БУДЕТ:**
```json
// Accept-Language: pl
{"key": "fish", "label": "Ryby"}

// Accept-Language: en
{"key": "fish", "label": "Fish"}

// Accept-Language: ru
{"key": "fish", "label": "Рыба"}
```

---

## 🎨 Frontend Implementation

### Step 1: Fetch Categories on App Load

```typescript
// services/categoryService.ts
interface Category {
  key: string;
  label: string;
  icon: string;
  sortOrder: number;
}

export async function fetchCategories(language: string): Promise<Category[]> {
  const response = await fetch(
    'https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/catalog/ingredient-categories',
    {
      headers: {
        'Authorization': `Bearer ${getToken()}`,
        'Accept-Language': language // 'pl', 'en', 'ru'
      }
    }
  );
  
  const result = await response.json();
  
  // ⚠️ ВАЖНО: Backend возвращает {success: true, data: {categories: [...]}}
  if (!result.success || !result.data?.categories) {
    throw new Error('Failed to fetch categories');
  }
  
  return result.data.categories;
}
```

### Step 2: Store in Context/State

```typescript
// context/CategoryContext.tsx
import React, { createContext, useContext, useEffect, useState } from 'react';
import { fetchCategories } from '../services/categoryService';

interface CategoryContextType {
  categories: Category[];
  loading: boolean;
}

const CategoryContext = createContext<CategoryContextType>({
  categories: [],
  loading: true
});

export function CategoryProvider({ children, language }: { children: React.ReactNode, language: string }) {
  const [categories, setCategories] = useState<Category[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchCategories(language)
      .then(setCategories)
      .finally(() => setLoading(false));
  }, [language]); // ✅ Перезагружаем при смене языка

  return (
    <CategoryContext.Provider value={{ categories, loading }}>
      {children}
    </CategoryContext.Provider>
  );
}

export const useCategories = () => useContext(CategoryContext);
```

### Step 3: Build Filter Buttons Dynamically

```typescript
// components/FridgeFilters.tsx
import { useCategories } from '../context/CategoryContext';

export function FridgeFilters({ activeCategory, onCategoryChange }) {
  const { categories, loading } = useCategories();

  if (loading) return <div>Loading categories...</div>;

  return (
    <div className="filter-buttons">
      {categories.map(category => (
        <button
          key={category.key}
          className={activeCategory === category.key ? 'active' : ''}
          onClick={() => onCategoryChange(category.key)}
        >
          <span className="icon">{category.icon}</span>
          <span className="label">{category.label}</span>
        </button>
      ))}
    </div>
  );
}
```

### Step 4: Filter Items by category

⚠️ **ВАЖНО:** Используйте `item.category`, а не локализованное название!

```typescript
// pages/FridgePage.tsx
const [activeCategory, setActiveCategory] = useState('all');
const [fridgeItems, setFridgeItems] = useState<FridgeItem[]>([]);

// ✅ Фильтруем по category (стабильный ключ)
const filteredItems = fridgeItems.filter(item => {
  if (activeCategory === 'all') return true;
  return item.category === activeCategory; // ✅ item.category (fish, meat, dairy)
});

// ❌ НЕ ДЕЛАЙТЕ ТАК:
// return item.categoryName === "Ryby"; // ПЛОХО - зависит от языка!
```

---

## 🔄 Изменения в /api/fridge/items

### ✅ ТЕКУЩИЙ формат (используется сейчас)

```json
{
  "success": true,
  "data": {
    "items": [
      {
        "id": 123,
        "name": "Łosoś",
        "category": "fish",  // ✅ Используйте это поле для фильтрации
        "daysLeft": 3,
        "icon": "🐟"
      },
      {
        "id": 124,
        "name": "Jaja",
        "category": "egg",
        "daysLeft": 5,
        "icon": "🥚"
      }
    ]
  }
}
```

### 📌 Как работает связь категорий с продуктами:

1. **Каждый продукт в ingredient таблице имеет поле `category`** (fish, meat, egg, dairy, etc.)
2. **При добавлении в fridge**, category автоматически копируется из ingredient
3. **Backend возвращает `category`** как стабильный ключ (НЕ переводится!)
4. **Frontend получает список категорий** с локализованными названиями через `/api/catalog/ingredient-categories`
5. **Фильтрация работает по `category` === `categoryKey`**

### 🎯 Пример фильтрации:

```typescript
// 1. Загрузили категории с бэкенда
const categories = [
  {key: "fish", label: "Ryby", icon: "🐟"},  // label зависит от Accept-Language
  {key: "egg", label: "Jajka", icon: "🥚"}
];

// 2. Загрузили продукты из холодильника
const fridgeItems = [
  {id: 123, name: "Łosoś", category: "fish"},  // category = ключ из БД
  {id: 124, name: "Jaja", category: "egg"}
];

// 3. Фильтруем по category
const filtered = fridgeItems.filter(item => 
  activeCategory === 'all' || item.category === activeCategory
);
// Если activeCategory === "fish", получим [{id: 123, name: "Łosoś"}]

// 4. Отображаем кнопки фильтров
categories.map(cat => (
  <button onClick={() => setActiveCategory(cat.key)}>
    {cat.icon} {cat.label}  // 🐟 Ryby (на польском)
  </button>
))
```

**⚠️ ВАЖНО:**  
- Поле `category` в `/api/fridge/items` - это **стабильный ключ** (fish, meat, egg)
- Оно **НЕ переводится** на бэкенде
- Локализацию делаете вы на фронте, сопоставляя `item.category` с `category.key`

---

## 🔄 Планируемые изменения (optional)

В будущем возможно переименование поля для ясности:

### ❌ СТАРЫЙ формат (deprecated)

```json
{
  "items": [
    {
      "id": 123,
      "name": "Łosoś",
      "category": "fish",  // ❌ Это поле будет удалено
      "daysLeft": 3
    }
  ]
}
```

### ✅ НОВЫЙ формат (planned - optional)

```json
{
  "items": [
    {
      "id": 123,
      "name": "Łosoś",
      "categoryKey": "fish",  // ✅ Переименовано для ясности (optional)
      "daysLeft": 3
    }
  ]
}
```

**⚠️ ВРЕМЕННО:**  
Поле `category` работает и будет работать.  
Переименование в `categoryKey` - это **опциональное улучшение** для большей ясности.

**Вывод:** Используйте `item.category` для фильтрации уже сейчас! 🎯

---

## 🧪 Testing

### Test 1: Fetch Categories (Polish)

```bash
curl -H "Accept-Language: pl" \
     -H "Authorization: Bearer YOUR_TOKEN" \
     https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/catalog/ingredient-categories
```

**Expected:** All labels in Polish (Ryby, Mięso, Jajka...)

### Test 2: Fetch Categories (English)

```bash
curl -H "Accept-Language: en" \
     -H "Authorization: Bearer YOUR_TOKEN" \
     https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/catalog/ingredient-categories
```

**Expected:** All labels in English (Fish, Meat, Eggs...)

### Test 3: Fetch Categories (Russian)

```bash
curl -H "Accept-Language: ru" \
     -H "Authorization: Bearer YOUR_TOKEN" \
     https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/catalog/ingredient-categories
```

**Expected:** All labels in Russian (Рыба, Мясо, Яйца...)

### Test 4: Verify sortOrder

```bash
curl -H "Accept-Language: pl" \
     -H "Authorization: Bearer YOUR_TOKEN" \
     https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/catalog/ingredient-categories | jq '.categories[].sortOrder'
```

**Expected:** `0, 1, 2, 3, 4, 5, 6, 7, 8, 9` (ascending order)

---

## 📋 Migration Checklist for Frontend

- [ ] Create `services/categoryService.ts` with `fetchCategories()`
- [ ] Create `CategoryContext` for state management
- [ ] Wrap app in `<CategoryProvider language={currentLanguage}>`
- [ ] Update `FridgeFilters` to use `useCategories()` hook
- [ ] Change filtering logic to use `item.category` matching `categoryKey`
- [ ] Remove hardcoded `CATEGORY_MAP_PL` / `CATEGORY_MAP_EN` / `CATEGORY_MAP_RU`
- [ ] Test category switching with different languages
- [ ] Test "All" filter (item.category === "all" or show all items)
- [ ] Test each specific category filter (item.category === "fish", "meat", etc.)

---

## 🎯 Key Points

1. **Categories are reference data**, not hardcoded enum
2. **Backend does localization** based on `Accept-Language` header
3. **Frontend receives ONLY localized label**, not all 3 languages
4. **Use `item.category` for filtering**, matching with `category.key` from API
5. **Icons are emoji** for universal support
6. **sortOrder controls display order**, managed in database
7. **Refetch categories when language changes** (useEffect dependency)
8. **Backend returns `{success: true, data: {categories: [...]}}`** - standard format

---

## 🔮 Future: Adding New Category

**Backend:**
```sql
INSERT INTO ingredient_categories (key, icon, sort_order, label_pl, label_en, label_ru)
VALUES ('seafood', '🦐', 10, 'Owoce morza', 'Seafood', 'Морепродукты');
```

**Frontend:**  
**НИЧЕГО НЕ НУЖНО ДЕЛАТЬ!** 🎉  
Фронт автоматически подхватит новую категорию при следующей загрузке.

---

## 📞 Contact

Questions? Ask @Fodi999

**Deployment:** Koyeb автоматически деплоит при push в `main`  
**Database:** PostgreSQL Neon (production)  
**API Base:** `https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app`
