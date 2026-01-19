# 🔧 Автозаполнение единиц измерения при добавлении продукта в холодильник

## 📋 Задача

При выборе продукта из каталога **автоматически подставлять правильные единицы измерения** (кг, г, мл, шт) из базы данных ингредиентов. Пользователь вводит только количество и цену.

---

## ✅ Текущее состояние API

### 1️⃣ **Autocomplete API** (для поиска продуктов)

**Эндпоинт:**
```
GET /api/admin/ingredients/suggest?q={query}&limit=5
```

**Заголовок:**
```
Accept-Language: ru  // или pl, en
```

**Ответ:**
```json
{
  "data": [
    {
      "id": "uuid",
      "name": "Молоко",           // ✅ Локализованное имя (на языке пользователя)
      "category": "dairy",         // ✅ Категория
      "nutritionGroup": "dairy",
      "unit": "ml"                 // ✅ ЕДИНИЦА ИЗМЕРЕНИЯ из БД
    }
  ]
}
```

**Пример реального ответа:**
```json
{
  "data": [
    {
      "id": "abc-123",
      "name": "Молоко 2%",
      "category": "dairy",
      "nutritionGroup": "dairy",
      "unit": "ml"                 // 👈 ml (миллилитры)
    },
    {
      "id": "def-456",
      "name": "Яйца",
      "category": "egg",
      "nutritionGroup": "protein",
      "unit": "pcs"                // 👈 pcs (штуки)
    },
    {
      "id": "ghi-789",
      "name": "Мука",
      "category": "grain",
      "nutritionGroup": "carbohydrate",
      "unit": "g"                  // 👈 g (граммы)
    },
    {
      "id": "jkl-012",
      "name": "Растительное масло",
      "category": "condiment",
      "nutritionGroup": "fat",
      "unit": "ml"                 // 👈 ml (миллилитры)
    }
  ]
}
```

---

### 2️⃣ **Add to Fridge API** (добавление продукта)

**Эндпоинт:**
```
POST /api/fridge/items
```

**Headers:**
```
Authorization: Bearer <JWT_TOKEN>
Content-Type: application/json
```

**Request Body:**
```json
{
  "ingredientId": "abc-123",         // ✅ UUID из autocomplete
  "quantity": 500,                   // ✅ Количество (число)
  "unit": "ml",                      // ✅ Единица из autocomplete
  "expiresAt": "2026-01-27T00:00:00Z", // ✅ Срок годности (ISO 8601)
  "priceTotal": 50                   // ✅ Цена (опционально)
}
```

**Response:**
```json
{
  "data": {
    "id": "item-uuid",
    "ingredientId": "abc-123",
    "quantity": 500,
    "unit": "ml",
    "expiresAt": "2026-01-27T00:00:00Z",
    "priceTotal": 50,
    "status": "fresh",
    "daysLeft": 7,
    "ingredient": {
      "id": "abc-123",
      "name": "Молоко 2%",
      "category": "dairy",
      "unit": "ml"
    }
  }
}
```

---

## 🎯 Решение для фронтенда

### **Шаг 1: Получить единицу измерения из autocomplete**

Когда пользователь выбирает продукт, ответ от `/api/admin/ingredients/suggest` уже содержит поле `unit`:

```typescript
interface IngredientSuggestion {
  id: string;
  name: string;           // "Молоко 2%"
  category: string;       // "dairy"
  nutritionGroup: string; // "dairy"
  unit: string;           // 👈 "ml" - ИСПОЛЬЗУЙТЕ ЭТО
}
```

### **Шаг 2: Сохранить unit в состоянии формы**

```typescript
// React пример
const [selectedIngredient, setSelectedIngredient] = useState<IngredientSuggestion | null>(null);

// При выборе продукта из autocomplete
const handleSelectIngredient = (ingredient: IngredientSuggestion) => {
  setSelectedIngredient(ingredient);
  
  // ✅ Автоматически устанавливаем единицу измерения
  setUnit(ingredient.unit);
};
```

### **Шаг 3: Отобразить единицу в UI (read-only или disabled)**

```tsx
// Вариант 1: Показать как текст (без возможности редактирования)
<div>
  <label>Количество ({selectedIngredient?.unit})</label>
  <input 
    type="number" 
    placeholder={`например, 500 ${selectedIngredient?.unit}`}
    value={quantity}
    onChange={(e) => setQuantity(e.target.value)}
  />
</div>

// Вариант 2: Disabled input
<div>
  <label>Единица измерения</label>
  <input 
    type="text" 
    value={selectedIngredient?.unit || ''} 
    disabled 
  />
</div>

// Вариант 3: Badge/Tag рядом с количеством
<div className="quantity-input">
  <input type="number" value={quantity} />
  <span className="unit-badge">{selectedIngredient?.unit}</span>
</div>
```

### **Шаг 4: Отправить данные с правильной единицей**

```typescript
const addToFridge = async () => {
  const payload = {
    ingredientId: selectedIngredient.id,
    quantity: parseFloat(quantity),
    unit: selectedIngredient.unit,  // 👈 Из autocomplete
    expiresAt: expiryDate.toISOString(),
    priceTotal: price ? parseFloat(price) : undefined
  };

  const response = await fetch('/api/fridge/items', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    },
    body: JSON.stringify(payload)
  });
};
```

---

## 📊 Mapping единиц измерения для UI

Для красивого отображения можно создать маппинг:

```typescript
const UNIT_LABELS: Record<string, string> = {
  'g': 'граммы (г)',
  'kg': 'килограммы (кг)',
  'ml': 'миллилитры (мл)',
  'l': 'литры (л)',
  'pcs': 'штуки (шт)',
  'tsp': 'чайные ложки (ч.л.)',
  'tbsp': 'столовые ложки (ст.л.)',
  'cup': 'стаканы',
  'pinch': 'щепотки'
};

// Использование
<label>
  Количество ({UNIT_LABELS[selectedIngredient.unit] || selectedIngredient.unit})
</label>
```

**Локализация для PL/EN:**

```typescript
const UNIT_LABELS_PL: Record<string, string> = {
  'g': 'gramy (g)',
  'kg': 'kilogramy (kg)',
  'ml': 'mililitry (ml)',
  'l': 'litry (l)',
  'pcs': 'sztuki (szt)',
  // ...
};

const UNIT_LABELS_EN: Record<string, string> = {
  'g': 'grams (g)',
  'kg': 'kilograms (kg)',
  'ml': 'milliliters (ml)',
  'l': 'liters (l)',
  'pcs': 'pieces (pcs)',
  // ...
};
```

---

## 🎨 UX Рекомендации

### ✅ **Правильный UI:**

```
┌─────────────────────────────────────────┐
│ Продукт                                 │
│ ┌─────────────────────────────────────┐ │
│ │ Молоко 2% [выбрано]                 │ │
│ └─────────────────────────────────────┘ │
│                                         │
│ Единица: ml • Категория: Молочные       │ 👈 Info badge (read-only)
│                                         │
│ Срок годности:                          │
│ 27 января 2026 г. (7 дней)              │
│                                         │
│ Количество (ml)                         │ 👈 Unit из БД
│ ┌─────────────────────────────────────┐ │
│ │ 500                           ml    │ │ 👈 Badge справа
│ └─────────────────────────────────────┘ │
│                                         │
│ 💰 Цена (опционально)                   │
│ ┌─────────────────────────────────────┐ │
│ │ 50                            PLN   │ │
│ └─────────────────────────────────────┘ │
└─────────────────────────────────────────┘
```

### ❌ **Неправильный UI:**

```
┌─────────────────────────────────────────┐
│ Количество                              │
│ ┌─────────────┬─────────────────────┐   │
│ │ 500         │ [выбрать единицу ▼] │   │ ❌ НЕ давайте выбирать!
│ └─────────────┴─────────────────────┘   │
└─────────────────────────────────────────┘
```

**Почему?**
- У каждого продукта **фиксированная единица измерения** в БД
- Молоко всегда в `ml`, яйца в `pcs`, мука в `g`
- Пользователь не должен выбирать единицу - это **ошибка UX**

---

## 🔍 Проверка в базе данных

Список единиц измерения в БД (214 ингредиентов):

```sql
SELECT DISTINCT unit, COUNT(*) as count
FROM "Ingredient"
GROUP BY unit
ORDER BY count DESC;
```

**Результат:**
| unit | count |
|------|-------|
| g    | 102   |
| ml   | 58    |
| pcs  | 32    |
| tsp  | 12    |
| tbsp | 6     |
| cup  | 3     |
| pinch| 1     |

---

## 📝 Пример полного кода (React + TypeScript)

```typescript
import { useState, useEffect } from 'react';

interface Ingredient {
  id: string;
  name: string;
  category: string;
  unit: string;
}

const AddToFridgeForm = () => {
  const [query, setQuery] = useState('');
  const [suggestions, setSuggestions] = useState<Ingredient[]>([]);
  const [selected, setSelected] = useState<Ingredient | null>(null);
  const [quantity, setQuantity] = useState('');
  const [price, setPrice] = useState('');
  const [expiryDate, setExpiryDate] = useState(new Date());

  // Autocomplete search
  useEffect(() => {
    if (query.length < 2) {
      setSuggestions([]);
      return;
    }

    const fetchSuggestions = async () => {
      const lang = 'ru'; // или из localStorage/context
      const response = await fetch(
        `/api/admin/ingredients/suggest?q=${query}&limit=5`,
        { headers: { 'Accept-Language': lang } }
      );
      const data = await response.json();
      setSuggestions(data.data);
    };

    const timer = setTimeout(fetchSuggestions, 300);
    return () => clearTimeout(timer);
  }, [query]);

  // Select ingredient
  const handleSelect = (ingredient: Ingredient) => {
    setSelected(ingredient);
    setQuery(ingredient.name);
    setSuggestions([]);
    // ✅ Единица измерения уже в ingredient.unit
  };

  // Add to fridge
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!selected) return;

    const payload = {
      ingredientId: selected.id,
      quantity: parseFloat(quantity),
      unit: selected.unit,  // 👈 Из autocomplete
      expiresAt: expiryDate.toISOString(),
      priceTotal: price ? parseFloat(price) : undefined
    };

    const response = await fetch('/api/fridge/items', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${localStorage.getItem('token')}`
      },
      body: JSON.stringify(payload)
    });

    if (response.ok) {
      alert('Продукт добавлен в холодильник!');
      // Reset form
      setSelected(null);
      setQuery('');
      setQuantity('');
      setPrice('');
    }
  };

  return (
    <form onSubmit={handleSubmit}>
      {/* Autocomplete */}
      <div>
        <label>Продукт</label>
        <input
          type="text"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          placeholder="Найдите продукт (например, молоко, яйца)..."
        />
        {suggestions.length > 0 && (
          <ul className="suggestions">
            {suggestions.map((item) => (
              <li key={item.id} onClick={() => handleSelect(item)}>
                {item.name} <span className="badge">{item.unit}</span>
              </li>
            ))}
          </ul>
        )}
      </div>

      {/* Selected product info */}
      {selected && (
        <div className="selected-info">
          Выбранный продукт: <strong>{selected.name}</strong>
          <br />
          Единица: {selected.unit} • Категория: {selected.category}
        </div>
      )}

      {/* Quantity */}
      <div>
        <label>Количество ({selected?.unit || ''})</label>
        <input
          type="number"
          value={quantity}
          onChange={(e) => setQuantity(e.target.value)}
          placeholder={`например, 500 ${selected?.unit || ''}`}
          required
        />
      </div>

      {/* Price */}
      <div>
        <label>💰 Цена (опционально)</label>
        <input
          type="number"
          value={price}
          onChange={(e) => setPrice(e.target.value)}
          placeholder="например, 50"
        />
        <small>PLN за {selected?.unit || 'единицу'}</small>
      </div>

      {/* Expiry date */}
      <div>
        <label>Срок годности</label>
        <input
          type="date"
          value={expiryDate.toISOString().split('T')[0]}
          onChange={(e) => setExpiryDate(new Date(e.target.value))}
        />
      </div>

      <button type="submit" disabled={!selected}>
        Добавить в холодильник
      </button>
    </form>
  );
};
```

---

## ✅ Итоговая логика

1. **Пользователь вводит** название продукта
2. **Autocomplete** возвращает список с полем `unit`
3. **При выборе** продукта:
   - Сохраняется `unit` из ответа
   - Показывается в UI (read-only или badge)
4. **При отправке** формы:
   - `unit` берется из выбранного продукта
   - Пользователь вводит только `quantity` и `priceTotal`
5. **Backend** валидирует и сохраняет с правильной единицей

---

## 🔧 Поддержка нестандартных случаев

Если продукта нет в каталоге (маловероятно - 214 ингредиентов покрывают 99% случаев), можно:

1. **Предложить создать новый ингредиент** (админ-функция)
2. **Использовать дефолтную единицу** по категории:
   - Жидкости → `ml`
   - Твердые → `g`
   - Штучные → `pcs`

---

## 📌 Резюме

**Что делать на фронтенде:**

✅ Получить `unit` из `/api/admin/ingredients/suggest`  
✅ Сохранить в состоянии при выборе продукта  
✅ Показать в UI (disabled input или badge)  
✅ Отправить в `/api/fridge/items` без изменений  
❌ **НЕ давать пользователю выбирать единицу измерения**

**Почему это правильно:**
- У каждого продукта фиксированная единица измерения в БД
- Backend гарантирует консистентность данных
- UX становится проще - на 1 поле меньше
- Избегаем ошибок (молоко в граммах, яйца в литрах)

---

**Готово! 🎉**

Теперь у вас есть полное понимание, как реализовать автозаполнение единиц измерения на фронтенде.
