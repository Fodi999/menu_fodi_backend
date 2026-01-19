# ✅ РЕШЕНИЕ: Автозаполнение единиц измерения в форме "Добавить в холодильник"

## 🎯 Суть проблемы

Пользователь выбирает продукт из каталога → система должна **автоматически подставить правильную единицу измерения** (мл, г, шт) из базы данных. Пользователь вводит только **количество** и **цену**.

---

## ✅ Решение готово (backend уже поддерживает)

### 1. API Autocomplete возвращает `unit`:

**Запрос:**
```bash
GET /api/admin/ingredients/suggest?q=молок&limit=5
Header: Accept-Language: ru
```

**Ответ:**
```json
{
  "data": [
    {
      "id": "abc-123",
      "name": "Молоко 2%",
      "category": "dairy",
      "nutritionGroup": "dairy",
      "unit": "ml"           👈 ЕДИНИЦА ИЗМЕРЕНИЯ
    }
  ]
}
```

### 2. Frontend использует этот `unit`:

```typescript
// При выборе продукта из autocomplete
const handleSelect = (ingredient) => {
  setSelectedIngredient(ingredient);
  setUnit(ingredient.unit);  // 👈 Автоматически ml/g/pcs
}

// При отправке в /api/fridge/items
const payload = {
  ingredientId: selected.id,
  quantity: 500,
  unit: selected.unit,       // 👈 Из autocomplete
  expiresAt: "2026-01-27",
  priceTotal: 50
};
```

### 3. UI показывает единицу (read-only):

```tsx
<label>Количество ({selected?.unit})</label>
<input 
  type="number" 
  placeholder={`например, 500 ${selected?.unit}`}
/>

// ИЛИ
<span className="badge">{selected.unit}</span>
```

---

## 📊 Единицы измерения в базе

| Unit | Количество | Примеры продуктов |
|------|-----------|-------------------|
| `ml` | 58 | Молоко, масло, сок |
| `g` | 102 | Мука, сахар, мясо |
| `pcs` | 32 | Яйца, помидоры, огурцы |
| `tsp` | 12 | Специи, разрыхлитель |
| `tbsp` | 6 | Мед, томатная паста |
| `cup` | 3 | Крупы |
| `pinch` | 1 | Соль, перец |

---

## 🎨 UX рекомендация

### ✅ Правильно (единица read-only):
```
Продукт: Молоко 2%
Единица: ml • Категория: Молочные

Количество (ml)
[ 500          ] ml
```

### ❌ Неправильно (дать выбрать):
```
Количество
[ 500 ] [выбрать: мл/г/шт ▼]  ❌ НЕ ТАК!
```

**Почему?** У каждого продукта **фиксированная единица** в БД. Молоко всегда в `ml`, яйца в `pcs`.

---

## 📝 Итого

**Что делать:**
1. ✅ Получить `unit` из `/api/admin/ingredients/suggest`
2. ✅ Показать в UI (disabled или badge)
3. ✅ Отправить в `/api/fridge/items` без изменений
4. ❌ **НЕ давать пользователю выбирать единицу**

**Документация:** `FRONTEND_FRIDGE_UNITS_AUTO_FILL_GUIDE.md` (полная инструкция с кодом)

---

**Готово! Backend готов, осталось только использовать на фронтенде.** 🎉
