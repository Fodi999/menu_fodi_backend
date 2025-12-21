# 📊 Backend Economy Calculation - Implementation Status

## ✅ РЕАЛИЗОВАНО

Backend **УЖЕ СЧИТАЕТ** экономику правильно:

```go
// service.go, строка 924-925
usedCost = quantity × pricePerUnit
totalUsedCost += usedCost   // Σ(quantityUsed × pricePerUnit)

// service.go, строка 954
savedMoney = totalUsedCost - estimatedExtraCost
```

## 📤 ФОРМАТ ОТВЕТА

```json
POST /api/ai/create-recipe-from-fridge

Response:
{
  "success": true,
  "data": {
    "recipe": {
      "economy": {
        "usedValue": 18.42,          // ✅ Σ(quantityUsed × pricePerUnit)
        "estimatedExtraCost": 0,     // ✅ Pantry cost from AI
        "savedMoney": 18.42,         // ✅ usedValue - estimatedExtraCost
        "currency": "PLN"
      }
    },
    "usedProducts": [
      {
        "name": "Wołowina",
        "quantityUsed": 400,
        "unit": "g",
        "pricePerUnit": 0.05,        // PLN/g (50 PLN/kg)
        "usedCost": 20.0,            // 400g × 0.05 = 20 PLN
        "currency": "PLN"
      }
    ]
  }
}
```

## ⚠️ ПРОБЛЕМА

**Если economy возвращает 0:**
```json
"economy": {
  "usedValue": 0,
  "estimatedExtraCost": 0,
  "savedMoney": 0,
  "currency": "PLN"
}
```

**Причина:** У продуктов в БД **НЕТ** `current_price_per_unit`

## 🔧 КАК ДОБАВИТЬ ЦЕНЫ

### Способ 1: Через SQL (Neon.tech)

```sql
-- Example: Wołowina = 50 PLN/kg = 0.05 PLN/g
UPDATE user_fridge_items 
SET current_price_per_unit = 0.05,
    current_price_currency = 'PLN'
WHERE ingredient_id = (
    SELECT id FROM "Ingredient" 
    WHERE LOWER(name) = LOWER('Wołowina')
)
AND user_id = '924a278d-bdb6-4ad5-9167-d9e9495f6dab';

-- Cebula = 3 PLN/kg = 0.003 PLN/g
UPDATE user_fridge_items 
SET current_price_per_unit = 0.003,
    current_price_currency = 'PLN'
WHERE ingredient_id = (
    SELECT id FROM "Ingredient" 
    WHERE LOWER(name) = LOWER('Cebula')
)
AND user_id = '924a278d-bdb6-4ad5-9167-d9e9495f6dab';

-- Mleko = 4 PLN/L = 0.004 PLN/ml
UPDATE user_fridge_items 
SET current_price_per_unit = 0.004,
    current_price_currency = 'PLN'
WHERE ingredient_id = (
    SELECT id FROM "Ingredient" 
    WHERE LOWER(name) LIKE '%mleko%'
)
AND user_id = '924a278d-bdb6-4ad5-9167-d9e9495f6dab';
```

### Способ 2: Через Frontend UI

**TODO:** Добавить поле "Cena" в форму добавления/редактирования продукта:

```tsx
<input 
  type="number" 
  placeholder="Cena za kg/l (PLN)"
  onChange={e => {
    const pricePerKg = parseFloat(e.target.value);
    const pricePerUnit = unit === 'g' ? pricePerKg / 1000 : pricePerKg / 1000;
    setCurrentPricePerUnit(pricePerUnit);
  }}
/>
```

## 🧪 ТЕСТ С РЕАЛЬНЫМИ ЦЕНАМИ

После добавления цен:

```bash
curl -X POST https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/ai/create-recipe-from-fridge \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"language": "pl"}'
```

**Ожидаемый результат:**
```json
{
  "economy": {
    "usedValue": 23.0,        // 20 (Wołowina) + 0.6 (Cebula) + 2.4 (Mleko)
    "estimatedExtraCost": 0,  // AI: olej+sól=0 (basic pantry)
    "savedMoney": 23.0,       // 23 - 0 = 23 PLN saved
    "currency": "PLN"
  }
}
```

## 📋 CHECKLIST

- [x] Backend logic implemented (usedValue, savedMoney)
- [x] Handles missing prices gracefully (returns 0)
- [x] Returns economy structure always
- [x] Per-product cost breakdown in usedProducts
- [ ] Add prices to test products in DB
- [ ] Frontend UI for price input
- [ ] Verify economy calculation with real data
- [ ] Display economy on recipe page

## 🎯 NEXT STEPS

1. **Добавь цены тестовым продуктам** (через SQL или UI)
2. **Сгенерируй рецепт** → проверь что economy != 0
3. **Отобрази на фронтенде**:
   ```tsx
   <div className="economy">
     <p>💰 Wartość produktów: {recipe.economy.usedValue} PLN</p>
     <p>🛒 Koszt pantry: {recipe.economy.estimatedExtraCost} PLN</p>
     <p>✅ Zaoszczędzono: {recipe.economy.savedMoney} PLN</p>
   </div>
   ```
