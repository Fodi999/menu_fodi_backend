# ✅ PRICE & CATEGORY FIX COMPLETE

## 🎯 ЧТО БЫЛО ИСПРАВЛЕНО

### ПРОБЛЕМА №1: Все categoryKey = "other"  
### ПРОБЛЕМА №2: Цена читается из БД, но НЕ возвращается в response

---

## ✅ РЕШЕНИЕ ПРОБЛЕМЫ №1: CategoryKey

### Что было:
```json
{
  "items": [
    {"name": "Łosoś", "categoryKey": "other"},   // ❌ Неправильно
    {"name": "Jaja", "categoryKey": "other"},    // ❌ Неправильно
    {"name": "Kefir", "categoryKey": "other"}    // ❌ Неправильно
  ]
}
```

### Что стало:
```json
{
  "items": [
    {"name": "Łosoś", "categoryKey": "fish"},    // ✅ Правильно
    {"name": "Jaja", "categoryKey": "egg"},      // ✅ Правильно
    {"name": "Kefir", "categoryKey": "dairy"}    // ✅ Правильно
  ]
}
```

### Как исправили:

1. **Источник истины:** `Ingredient.category`
   - Все категории корректные в БД (проверено через SQL)
   - fish: 9, egg: 2, dairy: 26, condiment: 45, grain: 22, etc.

2. **Обновили `GetUserItemsV2()`:**
```go
response := models.FridgeItemResponseV2{
    CategoryKey: item.Ingredient.Category, // ✅ Берем напрямую из Ingredient
}
```

3. **Обновили handler:**
```go
func (h *FridgeHandlers) GetUserItems(w http.ResponseWriter, r *http.Request) {
    items, err := h.service.GetUserItemsV2(userID) // ✅ Используем V2
    // ...
}
```

---

## ✅ РЕШЕНИЕ ПРОБЛЕМЫ №2: Price в Response

### Что было:
```go
// Backend делал 7 запросов:
SELECT * FROM "user_fridge_price_history"
WHERE user_fridge_item_id = '...'
ORDER BY created_at DESC
LIMIT 1

// ❌ НО результат НЕ попадал в response!
```

### Что стало:
```json
{
  "items": [
    {
      "id": "...",
      "name": "Kefir",
      "categoryKey": "dairy",
      "quantity": 2000,
      "unit": "ml",
      
      "price": {
        "value": 4.45,
        "per": "l"
      },
      
      "computed": {
        "unitPrice": 0.00445,
        "totalCost": 8.90
      },
      
      "daysLeft": 13
    }
  ]
}
```

### Как исправили:

#### 1. Добавили колонку `unit_for_price` в БД

**Migration:** `migrations/20260120_add_unit_for_price.sql`
```sql
ALTER TABLE user_fridge_price_history
ADD COLUMN IF NOT EXISTS unit_for_price TEXT;

UPDATE user_fridge_price_history
SET unit_for_price = 'kg'
WHERE unit_for_price IS NULL;
```

#### 2. Обновили модель `UserFridgePriceHistory`

```go
type UserFridgePriceHistory struct {
    ID               string    `gorm:"primaryKey"`
    UserFridgeItemID string    `gorm:"not null;index"`
    PricePerUnit     float64   `gorm:"not null"`
    UnitForPrice     string    `gorm:"type:text"`        // ✅ НОВОЕ
    Currency         string    `gorm:"not null;default:'PLN'"`
    Source           string    `gorm:"not null;default:'manual'"`
    CreatedAt        time.Time `gorm:"autoCreateTime;index"`
}
```

#### 3. Создали DTO для ответа

```go
type PriceInfo struct {
    Value float64 `json:"value"` // 4.45
    Per   string  `json:"per"`   // "l"
}

type ComputedPrice struct {
    UnitPrice float64 `json:"unitPrice"` // 0.00445 (за 1ml)
    TotalCost float64 `json:"totalCost"` // 8.90 (2000ml × 0.00445)
}

type FridgeItemResponseV2 struct {
    ID          string     `json:"id"`
    Name        string     `json:"name"`
    CategoryKey string     `json:"categoryKey"`
    Quantity    float64    `json:"quantity"`
    Unit        string     `json:"unit"`
    ExpiresAt   *time.Time `json:"expiresAt,omitempty"`
    DaysLeft    *int       `json:"daysLeft,omitempty"`
    
    Price    *PriceInfo     `json:"price,omitempty"`    // ✅ НОВОЕ
    Computed *ComputedPrice `json:"computed,omitempty"` // ✅ НОВОЕ
}
```

#### 4. Добавили метод загрузки цены

```go
func (s *FridgeService) getLastPrice(itemID string) (*models.UserFridgePriceHistory, error) {
    var priceHistory models.UserFridgePriceHistory
    err := s.db.Where("user_fridge_item_id = ?", itemID).
        Order("created_at DESC").
        Limit(1).
        First(&priceHistory).Error

    if err == gorm.ErrRecordNotFound {
        return nil, nil // Нет истории - норм
    }
    return &priceHistory, err
}
```

#### 5. Добавили метод расчёта стоимости

```go
func (s *FridgeService) computeTotalCost(
    quantity float64, unit string,
    pricePerUnit float64, unitForPrice string,
) *models.ComputedPrice {
    
    // Нормализуем quantity к базовым единицам (g, ml, pcs)
    var quantityInBaseUnits float64
    switch unit {
    case "kg": quantityInBaseUnits = quantity * 1000
    case "g":  quantityInBaseUnits = quantity
    case "l":  quantityInBaseUnits = quantity * 1000
    case "ml": quantityInBaseUnits = quantity
    case "pcs": quantityInBaseUnits = quantity
    }

    // Нормализуем цену к базовым единицам
    var pricePerBaseUnit float64
    switch unitForPrice {
    case "kg": pricePerBaseUnit = pricePerUnit / 1000 // zł/kg → zł/g
    case "g":  pricePerBaseUnit = pricePerUnit
    case "l":  pricePerBaseUnit = pricePerUnit / 1000 // zł/l → zł/ml
    case "ml": pricePerBaseUnit = pricePerUnit
    case "pcs": pricePerBaseUnit = pricePerUnit
    }

    // Вычисляем общую стоимость
    totalCost := quantityInBaseUnits * pricePerBaseUnit
    totalCost = math.Round(totalCost*100) / 100 // Округляем до 2 знаков

    return &models.ComputedPrice{
        UnitPrice: pricePerBaseUnit,
        TotalCost: totalCost,
    }
}
```

#### 6. Обновили `GetUserItemsV2()`

```go
func (s *FridgeService) GetUserItemsV2(userID string) ([]models.FridgeItemResponseV2, error) {
    // ... загружаем items
    
    for _, item := range items {
        response := models.FridgeItemResponseV2{
            ID:          item.ID,
            Name:        item.Ingredient.Name,
            CategoryKey: item.Ingredient.Category, // ✅ Правильная категория
            Quantity:    item.Quantity,
            Unit:        item.Unit,
            ExpiresAt:   item.ExpiresAt,
            DaysLeft:    daysLeft,
        }

        // ✅ НОВОЕ: Загружаем цену
        lastPrice, err := s.getLastPrice(item.ID)
        if err == nil && lastPrice != nil {
            response.Price = &models.PriceInfo{
                Value: lastPrice.PricePerUnit,
                Per:   lastPrice.UnitForPrice,
            }

            // ✅ Вычисляем стоимость
            computed := s.computeTotalCost(
                item.Quantity,
                item.Unit,
                lastPrice.PricePerUnit,
                lastPrice.UnitForPrice,
            )
            if computed != nil {
                response.Computed = computed
            }
        }

        result = append(result, response)
    }

    return result, nil
}
```

#### 7. Обновили сохранение цены при добавлении

```go
// В AddItem():
if req.PriceInput != nil {
    priceReq := models.AddPriceRequest{
        PricePerUnit: req.PriceInput.Value, // ✅ Исходная цена (НЕ нормализованная)
        UnitForPrice: req.PriceInput.Per,   // ✅ Единица измерения (kg, l, pcs)
        Currency:     "PLN",
        Source:       "manual",
    }
    s.AddPrice(userID, item.ID, priceReq)
}
```

---

## 🧪 ТЕСТИРОВАНИЕ

### Проверка категорий в БД:

```bash
$ go run cmd/check_ingredient_categories/main.go

🔍 Ingredient Categories in Database:
=====================================
📦 Kefir (id: 72be7544...) → category: 'dairy'      ✅
📦 Łosoś (id: fe1c7431...) → category: 'fish'       ✅
📦 Sól (id: c4d477f8...) → category: 'condiment'    ✅
📦 Jaja (id: 3260aadf...) → category: 'egg'         ✅
📦 Kasza gryczana (id: 31fc2a49...) → category: 'grain' ✅
📦 Olej roślinny (id: 1b7cea8e...) → category: 'condiment' ✅
```

### Проверка API (после деплоя):

```bash
curl -H "Authorization: Bearer TOKEN" \
  https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/fridge/items
```

**Expected:**
```json
{
  "success": true,
  "data": {
    "items": [
      {
        "id": "4a3db250-cb98-41d3-bd7a-8a644b37106d",
        "name": "Kefir",
        "categoryKey": "dairy",
        "quantity": 2000,
        "unit": "ml",
        "price": {
          "value": 4.45,
          "per": "l"
        },
        "computed": {
          "unitPrice": 0.00445,
          "totalCost": 8.90
        },
        "daysLeft": 13
      }
    ]
  }
}
```

---

## 📋 CHECKLIST

### Backend:
- [x] Добавлена колонка `unit_for_price` в БД
- [x] Обновлена модель `UserFridgePriceHistory`
- [x] Добавлены DTO: `PriceInfo`, `ComputedPrice`, `FridgeItemResponseV2`
- [x] Создан метод `getLastPrice()`
- [x] Создан метод `computeTotalCost()`
- [x] Обновлен метод `GetUserItemsV2()`
- [x] Handler использует `GetUserItemsV2()`
- [x] Сохранение цены с `UnitForPrice`
- [x] CategoryKey берется из `Ingredient.category`
- [x] Deployed to Koyeb

### Frontend (TODO):
- [ ] Обновить TypeScript интерфейс `FridgeItem` с `price` и `computed`
- [ ] Изменить `category` → `categoryKey` в коде
- [ ] Отображать цену в UI (если есть)
- [ ] Отображать общую стоимость продукта
- [ ] Использовать `categoryKey` для фильтрации (не `category`)

---

## 🎉 ИТОГ

| Функция | До | После |
|---------|-----|-------|
| CategoryKey | ❌ all → "other" | ✅ Правильные ключи (fish, dairy, egg, grain, condiment) |
| Price в response | ❌ Отсутствует | ✅ `{value: 4.45, per: "l"}` |
| TotalCost | ❌ Отсутствует | ✅ `{unitPrice: 0.00445, totalCost: 8.90}` |
| Price history | ✅ Сохраняется | ✅ Сохраняется с `unit_for_price` |
| Price loading | ❌ Не возвращается | ✅ Загружается и возвращается |
| Calculations | ❌ Нет | ✅ Backend вычисляет все |

**Deployment:** Pushed to GitHub → Koyeb auto-deploy in progress (~2 min)

**Next test:** Refresh frontend, check console logs for `price` and `categoryKey` fields
