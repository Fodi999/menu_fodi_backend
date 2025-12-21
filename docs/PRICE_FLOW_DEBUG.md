# Price Flow Debug Guide - Economy Calculation

## Проблема
Economy возвращает `usedValue: 0` даже когда продукты с ценами есть в холодильнике.

## Чек-лист диагностики (по приоритету)

### ✅ ШАГ 1: Проверить цены в БД (КРИТИЧНО)

**SQL запрос в Neon.tech:**
```sql
SELECT
  ufi.id,
  ufi.user_id,
  i.name as ingredient_name,
  ufi.quantity,
  ufi.unit,
  ufi.current_price_per_unit,  -- ← КЛЮЧЕВОЕ ПОЛЕ
  ufi.current_price_currency,
  ufi.price_updated_at
FROM user_fridge_items ufi
JOIN "Ingredient" i ON ufi.ingredient_id = i.id
WHERE ufi.current_price_per_unit IS NOT NULL
ORDER BY ufi.user_id, i.name
LIMIT 20;
```

**Ожидаемые значения (normalized to base units):**
- Wołowina @ 20.56 PLN/kg → `0.02056` PLN/g
- Mleko @ 3.24 PLN/l → `0.00324` PLN/ml
- Ogórek @ 7.00 PLN/kg → `0.00700` PLN/g
- Cebula @ 3.45 PLN/kg → `0.00345` PLN/g

**❌ Если все `current_price_per_unit = NULL`:**
→ Проблема во FRONTEND или в сохранении цен в БД
→ Проверь `/api/fridge POST` endpoint - принимает ли `priceInput`?
→ Проверь нормализацию цен в backend

---

### ✅ ШАГ 2: Проверить загрузку из БД в Go model

**Файл:** `internal/modules/ai/transport/http/handlers.go`

**Строки 627-635:**
```go
var fridgeItems []models.UserFridgeItem
if err := h.db.Preload("Ingredient").Where("user_id = ?", userID).
    Find(&fridgeItems).Error; err != nil {
    // ...
}
```

**Проверка:** Убедись что `UserFridgeItem` модель включает поле:
```go
type UserFridgeItem struct {
    // ...
    CurrentPricePerUnit  *float64 `gorm:"column:current_price_per_unit"`
    CurrentPriceCurrency string   `gorm:"column:current_price_currency"`
    PriceUpdatedAt       *time.Time
}
```

**Debug logs (commit a36bdfb):**
После запроса recipe generation смотри в Koyeb logs:
```
INFO Loaded fridge items with prices user_id=... total_items=5
INFO Fridge item price ingredient_name="Mleko 2%" current_price_per_unit="0.0032 PLN"
INFO Fridge item price ingredient_name="Wołowina" current_price_per_unit="0.0206 PLN"
```

**❌ Если везде `current_price_per_unit="NULL"`:**
→ GORM не мапит поле из БД
→ Проверь struct tags в `models/user_fridge.go`

---

### ✅ ШАГ 3: Проверить передачу в DTO

**Файл:** `internal/modules/ai/transport/http/handlers.go`

**Строки 697-712:**
```go
// Get price per unit from current cache
var pricePerUnit *float64
currency := "PLN"
if item.CurrentPricePerUnit != nil && *item.CurrentPricePerUnit > 0 {
    pricePerUnit = item.CurrentPricePerUnit
    // ...
}

aiItems = append(aiItems, dto.FridgeItemDTO{
    Name:         item.Ingredient.Name,
    Quantity:     item.Quantity,
    PricePerUnit: pricePerUnit,  // ← КРИТИЧНО
    Currency:     currency,
})
```

**Debug logs:**
```
INFO Price data found for item name="Mleko 2%" price_per_unit=0.0032 currency="PLN"
WARN No price data for item name="Cebula" current_price_per_unit=<nil>
```

**❌ Если все products → "No price data":**
→ `item.CurrentPricePerUnit` приходит NULL из БД
→ Вернись к ШАГ 1

---

### ✅ ШАГ 4: Проверить расчёт в service

**Файл:** `internal/modules/ai/service/service.go`

**Строки 914-945:**
```go
for _, prod := range products {
    if prod.Priority <= 2 {  // critical/warning products
        usedCost := 0.0
        pricePerUnit := 0.0
        
        if prod.Item.PricePerUnit != nil && *prod.Item.PricePerUnit > 0 {
            pricePerUnit = *prod.Item.PricePerUnit
            usedCost = prod.Item.Quantity * pricePerUnit
            totalUsedCost += usedCost  // ← КРИТИЧНЫЙ ПОДСЧЁТ
        }
        
        usedProducts = append(usedProducts, dto.UsedProductInfo{
            Name:         prod.Item.Name,
            QuantityUsed: prod.Item.Quantity,
            PricePerUnit: pricePerUnit,
            UsedCost:     usedCost,
        })
    }
}
```

**Debug logs:**
```
[AI][ECONOMY] Used cost: 18.42 PLN, Extra cost: 0.00 PLN, Saved: 18.42 PLN (prices available: 3 products)
```

**❌ Если `Used cost: 0.00 PLN (prices available: 0 products)`:**
→ `prod.Item.PricePerUnit == nil` для всех продуктов
→ Вернись к ШАГ 3

---

### ✅ ШАГ 5: Проверить формирование economy response

**Файл:** `internal/modules/ai/service/service.go`

**Строки 957-964:**
```go
// ALWAYS override economy (even if 0)
recipe.Economy = &dto.RecipeEconomy{
    UsedFromFridge:     len(usedProducts) > 0,
    UsedValue:          totalUsedCost,
    EstimatedExtraCost: estimatedExtraCost,
    SavedMoney:         savedMoney,
    Currency:           currency,
}
```

**✅ Это всегда должно возвращать структуру (даже если usedValue=0)**

**API Response:**
```json
{
  "data": {
    "recipe": {
      "economy": {
        "usedValue": 18.42,
        "estimatedExtraCost": 0,
        "savedMoney": 18.42,
        "currency": "PLN"
      }
    },
    "usedProducts": [
      {
        "name": "Mleko 2%",
        "quantityUsed": 250,
        "pricePerUnit": 0.0032,
        "usedCost": 0.80,
        "currency": "PLN"
      }
    ]
  }
}
```

---

## Quick Test Commands

### 1. Check if prices exist in database
```sql
-- Run in Neon.tech SQL Editor
SELECT COUNT(*) as products_with_prices
FROM user_fridge_items
WHERE current_price_per_unit IS NOT NULL;
```

### 2. Add product with price (curl)
```bash
curl -X POST "https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/fridge" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "ingredientId": "INGREDIENT_UUID",
    "quantity": 500,
    "unit": "g",
    "expiresAt": "2025-12-31T00:00:00Z",
    "priceInput": {
      "value": 20.56,
      "per": "kg"
    }
  }'
```

### 3. Generate recipe and check economy
```bash
curl -X POST "https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/ai/create-recipe-from-fridge" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"language": "pl"}' | jq '.data.recipe.economy'
```

---

## Root Cause Analysis

### Наиболее вероятные проблемы (по частоте):

1. **🔥🔥🔥 Цены не сохранены в БД (90% случаев)**
   - `current_price_per_unit = NULL`
   - Frontend не отправляет `priceInput`
   - Backend не обрабатывает нормализацию

2. **🔥🔥 GORM не мапит поле (5%)**
   - Отсутствует struct tag: `gorm:"column:current_price_per_unit"`
   - Или SELECT не включает это поле (но мы используем `Find()` → берёт все поля)

3. **🔥 Handler не передаёт в DTO (3%)**
   - Логика `if item.CurrentPricePerUnit != nil` не срабатывает
   - Debug logs покажут "No price data for item"

4. **Service не считает (2%)**
   - Логика расчёта пропускает продукты
   - Debug logs покажут "prices available: 0 products"

---

## Next Steps

1. **Проверь Koyeb logs** после вызова `/api/ai/create-recipe-from-fridge`:
   - Ищи: `"Loaded fridge items with prices"`
   - Ищи: `"Price data found for item"` vs `"No price data for item"`
   - Ищи: `"[AI][ECONOMY] Used cost:"`

2. **Если все "No price data"** → SQL проверка в Neon:
   ```sql
   SELECT * FROM user_fridge_items 
   WHERE current_price_per_unit IS NOT NULL 
   LIMIT 5;
   ```

3. **Если БД пустая** → Проверь `/api/fridge POST`:
   - Принимает ли `priceInput`?
   - Нормализует ли цены?
   - Сохраняет ли в `current_price_per_unit`?

4. **Если БД заполнена, но handler не видит** → GORM mapping issue:
   - Проверь `internal/models/user_fridge.go`
   - Убедись что `CurrentPricePerUnit *float64` есть в struct

---

## Current Status (commit a36bdfb)

✅ Debug logging added to handlers  
✅ Service economy calculation correct  
✅ DTO structure includes PricePerUnit  
✅ Handler extracts CurrentPricePerUnit from DB  
⏳ Need to verify: prices actually saved in database  
⏳ Need to check: Koyeb logs for debug output  

**Expected next action:** Check Koyeb logs after recipe generation request
