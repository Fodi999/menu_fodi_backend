# 🏗️ Fridge Price Response - Правильная Архитектура

## 🎯 Проблема

**СЕЙЧАС:** Backend загружает историю цен из БД, но НЕ возвращает её во frontend:

```go
// ✅ Загружаем из БД (УЖЕ ЕСТЬ)
SELECT * FROM "user_fridge_price_history"
WHERE user_fridge_item_id = $1
ORDER BY created_at DESC
LIMIT 1

// ❌ НО не мапим в response DTO!
```

**Frontend не получает:**
- Цену за единицу (например, 6.30 zł/kg)
- Единицу измерения цены (kg, l, pcs)
- Общую стоимость продукта в холодильнике

---

## ✅ ПРАВИЛЬНАЯ АРХИТЕКТУРА

### Концепция: Цена = Агрегат user_fridge_item + last price history

```
┌─────────────────────┐
│ user_fridge_items   │
├─────────────────────┤
│ id                  │ ←─────┐
│ name                │       │
│ quantity            │       │ JOIN
│ unit                │       │
│ expires_at          │       │
└─────────────────────┘       │
                              │
┌──────────────────────────┐  │
│ user_fridge_price_history│  │
├──────────────────────────┤  │
│ user_fridge_item_id      │──┘
│ price_per_unit          │ ← ПОСЛЕДНЯЯ цена
│ unit_for_price          │ ← kg, l, pcs
│ created_at              │
└──────────────────────────┘
         │
         │ MAP TO
         ▼
┌──────────────────────────┐
│ Frontend Response DTO    │
├──────────────────────────┤
│ {                        │
│   id, name, quantity...  │
│   price: {               │
│     value: 6.3,          │
│     per: "kg"            │
│   },                     │
│   computed: {            │
│     unitPrice: 0.0063,   │
│     totalCost: 0.0126    │
│   }                      │
│ }                        │
└──────────────────────────┘
```

---

## 🔧 ЧТО НУЖНО ИЗМЕНИТЬ

### 1️⃣ Обновить структуру DTO

**File:** `internal/models/fridge_item.go`

```go
// FridgeItemResponse - DTO для GET /api/fridge/items
type FridgeItemResponse struct {
    ID          string    `json:"id"`
    Name        string    `json:"name"`
    CategoryKey string    `json:"categoryKey"` // ✅ Изменили с "category" на "categoryKey"
    Quantity    float64   `json:"quantity"`
    Unit        string    `json:"unit"`
    ExpiresAt   time.Time `json:"expiresAt"`
    DaysLeft    int       `json:"daysLeft"`
    
    // ✅ НОВОЕ: Информация о цене
    Price       *PriceInfo       `json:"price,omitempty"`    // Цена за единицу
    Computed    *ComputedPrice   `json:"computed,omitempty"` // Расчёты
}

// PriceInfo - цена за единицу (из user_fridge_price_history)
type PriceInfo struct {
    Value float64 `json:"value"` // 6.3 (цена)
    Per   string  `json:"per"`   // "kg", "l", "pcs"
}

// ComputedPrice - вычисленные значения
type ComputedPrice struct {
    UnitPrice float64 `json:"unitPrice"` // Цена за 1 грамм/мл (для внутренних расчётов)
    TotalCost float64 `json:"totalCost"` // Общая стоимость quantity в холодильнике
}
```

---

### 2️⃣ Загрузить последнюю цену в сервисе

**File:** `internal/modules/fridge/service/fridge_service.go`

#### Вариант A: В методе GetItems() (рекомендуется)

```go
func (s *FridgeService) GetItems(userID string) ([]models.FridgeItemResponse, error) {
    // 1. Загружаем items
    var items []models.UserFridgeItem
    if err := s.db.Where("user_id = ?", userID).
        Preload("Ingredient").
        Find(&items).Error; err != nil {
        return nil, err
    }

    // 2. Мапим в response DTO
    response := make([]models.FridgeItemResponse, len(items))
    for i, item := range items {
        response[i] = models.FridgeItemResponse{
            ID:          item.ID.String(),
            Name:        item.Name,
            CategoryKey: item.Ingredient.Category, // ✅ categoryKey вместо category
            Quantity:    item.Quantity,
            Unit:        item.Unit,
            ExpiresAt:   item.ExpiresAt,
            DaysLeft:    calculateDaysLeft(item.ExpiresAt),
        }

        // ✅ НОВОЕ: Загружаем последнюю цену
        lastPrice, err := s.getLastPrice(item.ID)
        if err == nil && lastPrice != nil {
            response[i].Price = &models.PriceInfo{
                Value: lastPrice.PricePerUnit,
                Per:   lastPrice.UnitForPrice,
            }

            // Вычисляем стоимость
            response[i].Computed = s.computeTotalCost(
                item.Quantity,
                item.Unit,
                lastPrice.PricePerUnit,
                lastPrice.UnitForPrice,
            )
        }
    }

    return response, nil
}
```

#### Добавить вспомогательный метод:

```go
// getLastPrice - загружает последнюю цену для item
func (s *FridgeService) getLastPrice(itemID uuid.UUID) (*models.UserFridgePriceHistory, error) {
    var priceHistory models.UserFridgePriceHistory
    err := s.db.Where("user_fridge_item_id = ?", itemID).
        Order("created_at DESC").
        Limit(1).
        First(&priceHistory).Error
    
    if err == gorm.ErrRecordNotFound {
        return nil, nil // Нет истории цен — это нормально
    }
    if err != nil {
        return nil, err
    }
    
    return &priceHistory, nil
}
```

#### Добавить метод расчёта стоимости:

```go
// computeTotalCost - вычисляет общую стоимость продукта в холодильнике
func (s *FridgeService) computeTotalCost(
    quantity float64,
    unit string,
    pricePerUnit float64,
    unitForPrice string,
) *models.ComputedPrice {
    
    // Нормализуем quantity к базовым единицам (g, ml, pcs)
    var quantityInBaseUnits float64
    switch unit {
    case "kg":
        quantityInBaseUnits = quantity * 1000 // kg → g
    case "g":
        quantityInBaseUnits = quantity
    case "l":
        quantityInBaseUnits = quantity * 1000 // l → ml
    case "ml":
        quantityInBaseUnits = quantity
    case "pcs":
        quantityInBaseUnits = quantity
    default:
        return nil // Неизвестная единица
    }
    
    // Нормализуем цену к базовым единицам
    var pricePerBaseUnit float64
    switch unitForPrice {
    case "kg":
        pricePerBaseUnit = pricePerUnit / 1000 // zł/kg → zł/g
    case "g":
        pricePerBaseUnit = pricePerUnit
    case "l":
        pricePerBaseUnit = pricePerUnit / 1000 // zł/l → zł/ml
    case "ml":
        pricePerBaseUnit = pricePerUnit
    case "pcs":
        pricePerBaseUnit = pricePerUnit // zł/pcs
    default:
        return nil
    }
    
    // Вычисляем общую стоимость
    totalCost := quantityInBaseUnits * pricePerBaseUnit
    
    return &models.ComputedPrice{
        UnitPrice: pricePerBaseUnit,  // Цена за 1g/ml/pcs
        TotalCost: totalCost,          // Общая стоимость
    }
}
```

---

### 3️⃣ Пример правильного ответа API

#### Request:

```bash
GET /api/fridge/items
Authorization: Bearer JWT_TOKEN
```

#### Response:

```json
{
  "success": true,
  "data": {
    "items": [
      {
        "id": "3f6f73fb-1234-5678-9abc-def012345678",
        "name": "Kasza gryczana",
        "categoryKey": "grain",
        "quantity": 500,
        "unit": "g",
        "expiresAt": "2027-01-20T00:00:00Z",
        "daysLeft": 364,
        
        "price": {
          "value": 6.3,
          "per": "kg"
        },
        
        "computed": {
          "unitPrice": 0.0063,
          "totalCost": 3.15
        }
      },
      {
        "id": "9a8b7c6d-4321-8765-dcba-fedcba098765",
        "name": "Łosoś",
        "categoryKey": "fish",
        "quantity": 300,
        "unit": "g",
        "expiresAt": "2026-01-23T00:00:00Z",
        "daysLeft": 3,
        
        "price": {
          "value": 45.0,
          "per": "kg"
        },
        
        "computed": {
          "unitPrice": 0.045,
          "totalCost": 13.5
        }
      },
      {
        "id": "5e4d3c2b-9876-5432-fedc-ba9876543210",
        "name": "Jaja",
        "categoryKey": "egg",
        "quantity": 10,
        "unit": "pcs",
        "expiresAt": "2026-01-25T00:00:00Z",
        "daysLeft": 5,
        
        "price": {
          "value": 1.2,
          "per": "pcs"
        },
        
        "computed": {
          "unitPrice": 1.2,
          "totalCost": 12.0
        }
      },
      {
        "id": "1a2b3c4d-5678-90ab-cdef-0123456789ab",
        "name": "Sól",
        "categoryKey": "condiment",
        "quantity": 1000,
        "unit": "g",
        "expiresAt": "2027-01-20T00:00:00Z",
        "daysLeft": 365,
        
        "price": null,
        "computed": null
      }
    ]
  }
}
```

#### Объяснение примера:

1. **Kasza gryczana:**
   - Цена: 6.30 zł/kg
   - В холодильнике: 500g
   - UnitPrice: 0.0063 zł/g (6.30 / 1000)
   - TotalCost: 3.15 zł (500g × 0.0063)

2. **Łosoś:**
   - Цена: 45.00 zł/kg
   - В холодильнике: 300g
   - UnitPrice: 0.045 zł/g
   - TotalCost: 13.50 zł (300g × 0.045)

3. **Jaja:**
   - Цена: 1.20 zł/pcs
   - В холодильнике: 10 pcs
   - UnitPrice: 1.20 zł/pcs
   - TotalCost: 12.00 zł (10 × 1.20)

4. **Sól:**
   - Цена НЕ указана (price: null)
   - computed: null

---

## 🎨 Frontend Integration

### Отображение цены

```typescript
interface FridgeItem {
  id: string;
  name: string;
  categoryKey: string;
  quantity: number;
  unit: string;
  expiresAt: string;
  daysLeft: number;
  
  price?: {
    value: number;  // 6.3
    per: string;    // "kg", "l", "pcs"
  };
  
  computed?: {
    unitPrice: number;  // 0.0063 (за 1g/ml/pcs)
    totalCost: number;  // 3.15 (общая стоимость)
  };
}
```

### UI пример:

```tsx
function FridgeItemCard({ item }: { item: FridgeItem }) {
  return (
    <div className="fridge-item">
      <h3>{item.name}</h3>
      <p>Quantity: {item.quantity} {item.unit}</p>
      <p>Expires in: {item.daysLeft} days</p>
      
      {item.price && (
        <div className="price-info">
          <p>Price: {item.price.value} zł/{item.price.per}</p>
        </div>
      )}
      
      {item.computed && (
        <div className="computed-cost">
          <p><strong>Total cost: {item.computed.totalCost.toFixed(2)} zł</strong></p>
        </div>
      )}
    </div>
  );
}
```

### Результат на экране:

```
┌─────────────────────────┐
│ 🌾 Kasza gryczana       │
│ Quantity: 500 g         │
│ Expires in: 364 days    │
│ Price: 6.30 zł/kg       │
│ Total cost: 3.15 zł     │
└─────────────────────────┘

┌─────────────────────────┐
│ 🐟 Łosoś                │
│ Quantity: 300 g         │
│ Expires in: 3 days ⚠️   │
│ Price: 45.00 zł/kg      │
│ Total cost: 13.50 zł    │
└─────────────────────────┘

┌─────────────────────────┐
│ 🥚 Jaja                 │
│ Quantity: 10 pcs        │
│ Expires in: 5 days      │
│ Price: 1.20 zł/pcs      │
│ Total cost: 12.00 zł    │
└─────────────────────────┘

┌─────────────────────────┐
│ 🧂 Sól                  │
│ Quantity: 1000 g        │
│ Expires in: 365 days    │
│ (No price data)         │
└─────────────────────────┘
```

---

## 📋 Implementation Checklist

### Backend:

- [ ] Обновить `FridgeItemResponse` struct (добавить `Price` и `Computed`)
- [ ] Изменить `category` → `categoryKey` в response
- [ ] Добавить метод `getLastPrice(itemID)` в `fridge_service.go`
- [ ] Добавить метод `computeTotalCost()` для расчёта стоимости
- [ ] Обновить `GetItems()` - загружать последнюю цену и мапить в DTO
- [ ] Протестировать: items с ценой, items без цены
- [ ] Протестировать: разные единицы (kg/g, l/ml, pcs)

### Frontend:

- [ ] Обновить TypeScript интерфейс `FridgeItem` (добавить `price?` и `computed?`)
- [ ] Изменить `category` → `categoryKey` в коде
- [ ] Обновить UI для отображения цены (если есть)
- [ ] Обновить UI для отображения общей стоимости
- [ ] Добавить визуальную индикацию для items без цены
- [ ] Протестировать отображение разных валют и единиц

---

## 🎯 Преимущества правильной архитектуры

✅ **Единый источник истины:** Цена хранится в `user_fridge_price_history`  
✅ **История цен:** Можем отследить изменение цен со временем  
✅ **Гибкость:** Цена может быть в разных единицах (kg, l, pcs)  
✅ **Автоматические расчёты:** Backend вычисляет общую стоимость  
✅ **Опциональность:** Items без цены тоже поддерживаются (price: null)  
✅ **Масштабируемость:** Легко добавить analytics (средняя цена, динамика)

---

## 🚀 Дальнейшие улучшения (future)

1. **Статистика холодильника:**
   ```json
   {
     "totalValue": 28.65,  // Общая стоимость всех продуктов
     "expiringValue": 13.50  // Стоимость продуктов с daysLeft < 3
   }
   ```

2. **История изменения цен:**
   ```
   GET /api/fridge/items/{id}/price-history
   ```

3. **Средняя цена по категориям:**
   ```json
   {
     "fish": {"avgPrice": 42.5, "unit": "kg"},
     "dairy": {"avgPrice": 3.2, "unit": "l"}
   }
   ```

4. **Уведомление о росте цен:**
   ```
   "Łosoś подорожал на 15% (было 39 zł/kg, стало 45 zł/kg)"
   ```

---

## 📞 Summary

### БЫЛО (неправильно):
- ❌ Backend загружает price_history, но НЕ возвращает во frontend
- ❌ Frontend не знает о ценах продуктов
- ❌ Невозможно рассчитать стоимость холодильника

### СТАЛО (правильно):
- ✅ Backend загружает последнюю цену из `user_fridge_price_history`
- ✅ Backend мапит цену в response DTO (`price`, `computed`)
- ✅ Frontend получает цену и может отобразить стоимость
- ✅ Архитектура поддерживает items без цены (optional fields)
- ✅ Цена за единицу + общая стоимость вычисляются автоматически

**Следующий шаг:** Реализовать код в `fridge_service.go` и обновить DTO в `fridge_item.go`
