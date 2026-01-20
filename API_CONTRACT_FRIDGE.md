# 📋 API Contract: Fridge Items

**Endpoint:** `GET /api/fridge/items`  
**Authorization:** Bearer JWT  
**Localization:** Accept-Language header (pl | en | ru)

---

## ✅ ПРАВИЛЬНЫЙ КОНТРАКТ

### Request

```http
GET /api/fridge/items HTTP/1.1
Authorization: Bearer <JWT_TOKEN>
Accept-Language: ru
```

### Response

```json
{
  "success": true,
  "data": {
    "items": [
      {
        "id": "932c1f69-1454-44c8-9e54-b6062a5c0883",
        "name": "Łosoś",
        "ingredient": {
          "id": "fe1c7431-b1b7-4d36-94bf-74276481983e",
          "name": "Łosoś",
          "namePl": "Łosoś",
          "nameEn": "Salmon",
          "nameRu": "Лосось",
          "unit": "g"
        },
        "categoryKey": "fish",
        "quantity": 2000,
        "unit": "g",
        "expiresAt": "2026-01-25T12:00:00Z",
        "daysLeft": 4,
        "currentPrice": {
          "value": 12.34,
          "per": "kg",
          "currency": "PLN",
          "updatedAt": "2026-01-20T10:30:00Z"
        }
      }
    ]
  }
}
```

---

## 📊 Field Descriptions

### `id` (string, required)
UUID продукта в холодильнике

### `ingredient` (object, required)
Информация об ингредиенте из каталога:
- `id` - UUID ингредиента в каталоге
- `name` - Legacy поле (польское название или fallback)
- `namePl` - Польское название (может быть null)
- `nameEn` - Английское название (может быть null)
- `nameRu` - Русское название (может быть null)
- `unit` - Единица измерения (g, ml, pcs)

**ВАЖНО:**
- ✅ Backend отдаёт **ВСЕ переводы**
- ✅ Frontend **сам выбирает** нужный язык
- ✅ Fallback логика на фронте: `nameRu || namePl || name`

**Frontend пример:**
```typescript
const getIngredientName = (ingredient: IngredientInfo, lang: string) => {
  switch(lang) {
    case 'ru': return ingredient.nameRu || ingredient.namePl || ingredient.name;
    case 'en': return ingredient.nameEn || ingredient.namePl || ingredient.name;
    case 'pl': return ingredient.namePl || ingredient.name;
    default: return ingredient.name;
  }
}
```

### `categoryKey` (string, required)
**Stable key категории** - НЕ зависит от языка!

Возможные значения:
- `fish` - Рыба
- `meat` - Мясо
- `egg` - Яйца
- `dairy` - Молочные продукты
- `vegetable` - Овощи
- `fruit` - Фрукты
- `grain` - Крупы
- `condiment` - Специи/масла
- `other` - Прочее (fallback)

**ВАЖНО:**
- ✅ Всегда стабильный ключ (fish, meat, etc.)
- ❌ НИКОГДА не возвращается label ("Рыба", "Ryby", "Fish")
- ✅ Используется для фильтрации на фронте

### `quantity` (number, required)
Количество продукта

### `unit` (string, required)
Единица измерения количества (g, ml, pcs)

### `expiresAt` (string, optional)
ISO 8601 дата истечения срока годности

### `daysLeft` (number, optional)
Количество дней до истечения (вычисляется на backend)

### `currentPrice` (object, optional)
Текущая цена продукта:
- `value` - Числовое значение цены
- `per` - За какую единицу измерения (kg, l, pcs)
- `currency` - Валюта (PLN, EUR, USD)
- `updatedAt` - Когда обновлена цена

---

## 🎯 Ключевые правила контракта

### 1. CategoryKey - ТОЛЬКО из БД

```go
// ✅ ПРАВИЛЬНО:
categoryKey := ingredient.Category
if categoryKey == "" {
    categoryKey = "other"
}

// ❌ НЕПРАВИЛЬНО:
categoryKey := mapCategoryByName(ingredient.Name)      // НЕТ!
categoryKey := translateCategory(ingredient.Category)  // НЕТ!
```

**Источник правды:** поле `category` в таблице `Ingredient`

### 2. Локализация - ТОЛЬКО для имён

```
Accept-Language: ru → ingredient.name = "Лосось"
Accept-Language: pl → ingredient.name = "Łosoś"
Accept-Language: en → ingredient.name = "Salmon"
```

**НЕ локализуются:**
- ❌ `categoryKey` (всегда "fish", не "рыба")
- ❌ `unit` (всегда "g", не "грамм")
- ❌ `currency` (всегда "PLN", не "злотый")

### 3. CurrentPrice - всегда включать если есть

```go
// ✅ Используем денормализованные поля из user_fridge_items:
if item.CurrentPricePerUnit != nil {
    response.CurrentPrice = &CurrentPriceInfo{
        Value:     *item.CurrentPricePerUnit,
        Per:       item.Unit,
        Currency:  item.CurrentPriceCurrency,
        UpdatedAt: item.PriceUpdatedAt,
    }
}
```

**Не делаем N+1 queries к price_history!**

### 4. Ingredient - структурированный объект

```json
// ✅ ПРАВИЛЬНО (новый контракт):
{
  "ingredient": {
    "id": "...",
    "name": "Лосось",
    "unit": "g"
  },
  "categoryKey": "fish",
  "quantity": 2000,
  "unit": "g"
}

// ❌ СТАРЫЙ формат (deprecated):
{
  "name": "Лосось",
  "categoryKey": "fish",
  "quantity": 2000,
  "unit": "g"
}
```

---

## 🧪 Тестирование контракта

### Test 1: Получить items с русской локализацией

```bash
curl -H "Authorization: Bearer <TOKEN>" \
     -H "Accept-Language: ru" \
     https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/fridge/items
```

**Проверить:**
- ✅ `ingredient.name` на русском
- ✅ `categoryKey` = "fish" (не "рыба")
- ✅ `currentPrice` есть (если была установлена)

### Test 2: Получить items с польской локализацией

```bash
curl -H "Authorization: Bearer <TOKEN>" \
     -H "Accept-Language: pl" \
     https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/fridge/items
```

**Проверить:**
- ✅ `ingredient.name` на польском
- ✅ `categoryKey` = "fish" (тот же ключ!)
- ✅ Структура идентична

### Test 3: Fallback при отсутствии языка

```bash
curl -H "Authorization: Bearer <TOKEN>" \
     -H "Accept-Language: *" \
     https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/fridge/items
```

**Ожидается:**
- ✅ Fallback на польский (pl)
- ✅ `categoryKey` всё ещё правильный

---

## 📚 Related Endpoints

### Categories catalog

```
GET /api/catalog/ingredient-categories
Accept-Language: ru

Response:
{
  "success": true,
  "data": {
    "categories": [
      {
        "key": "fish",      // ✅ То же что categoryKey!
        "label": "Рыба",    // Локализованное название для UI
        "icon": "🐟",
        "sortOrder": 1
      }
    ]
  }
}
```

**Связь:**
```
item.categoryKey === category.key  // ✅ Для фильтрации
category.label                      // ✅ Для отображения
```

---

## ⚠️ Частые ошибки

### ❌ ОШИБКА 1: Использование label для фильтрации

```typescript
// ❌ НЕПРАВИЛЬНО:
if (item.categoryKey === "Рыба") { ... }

// ✅ ПРАВИЛЬНО:
if (item.categoryKey === "fish") { ... }
```

### ❌ ОШИБКА 2: Чтение неправильного поля

```typescript
// ❌ НЕПРАВИЛЬНО:
const categoryKey = item.category || 'other'

// ✅ ПРАВИЛЬНО:
const categoryKey = item.categoryKey || 'other'
```

### ❌ ОШИБКА 3: Игнорирование Accept-Language

```typescript
// ❌ НЕПРАВИЛЬНО:
fetch('/api/fridge/items')  // нет локализации!

// ✅ ПРАВИЛЬНО:
fetch('/api/fridge/items', {
  headers: {
    'Accept-Language': userLanguage  // 'pl', 'en', 'ru'
  }
})
```

---

## 🎉 Summary

**Backend гарантирует:**
- ✅ `categoryKey` всегда stable key (fish, meat, etc.)
- ✅ `ingredient.name` локализовано по Accept-Language
- ✅ `currentPrice` включена если есть
- ✅ Структура данных консистентна

**Frontend должен:**
- ✅ Отправлять правильный Accept-Language header
- ✅ Фильтровать по `item.categoryKey`
- ✅ Отображать `category.label` из catalog
- ✅ Не трансформировать `categoryKey`

**Результат:**
- 🚀 Нет багов с категориями
- 🌍 Правильная локализация
- 💰 Цены всегда видны
- ⚡ Быстрая работа (без N+1 queries)
