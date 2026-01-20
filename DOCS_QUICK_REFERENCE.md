# 📚 Backend Documentation - Quick Reference

## 🎯 Категории продуктов (Categories)

**Проблема на frontend:** `TypeError: Cannot read properties of undefined (reading 'sort')`

**Причина:** Backend возвращает `{success: true, data: {categories: [...]}}`, а frontend ожидает `{categories: [...]}`

**Решение:** См. **CATEGORY_FILTERING_QUICK_START.md** (314 строк)

**API Endpoint:**
```bash
GET /api/catalog/ingredient-categories
Headers:
  - Authorization: Bearer JWT_TOKEN
  - Accept-Language: pl | en | ru
```

**Response:**
```json
{
  "success": true,
  "data": {
    "categories": [
      {"key": "fish", "label": "Ryby", "icon": "🐟", "sortOrder": 1}
    ]
  }
}
```

**Frontend fix:**
```typescript
const result = await response.json();
return result.data.categories; // ✅ Правильный путь
```

---

## 💰 Цены в холодильнике (Fridge Prices)

**Проблема:** Backend загружает price_history, но НЕ возвращает во frontend

**Архитектура:** Цена = Агрегат `user_fridge_item` + `last price history`

**Решение:** См. **FRIDGE_PRICE_RESPONSE_ARCHITECTURE.md** (полная документация)

**Что нужно добавить:**

1. Обновить DTO:
```go
type FridgeItemResponse struct {
    // ... existing fields
    Price    *PriceInfo     `json:"price,omitempty"`
    Computed *ComputedPrice `json:"computed,omitempty"`
}
```

2. Загрузить последнюю цену:
```go
func (s *FridgeService) getLastPrice(itemID uuid.UUID) (*models.UserFridgePriceHistory, error) {
    var priceHistory models.UserFridgePriceHistory
    err := s.db.Where("user_fridge_item_id = ?", itemID).
        Order("created_at DESC").
        Limit(1).
        First(&priceHistory).Error
    return &priceHistory, err
}
```

3. Вычислить стоимость:
```go
func (s *FridgeService) computeTotalCost(
    quantity float64, unit string,
    pricePerUnit float64, unitForPrice string,
) *models.ComputedPrice {
    // Normalize to base units (g, ml, pcs)
    // Calculate: totalCost = quantity × pricePerUnit
}
```

**Ожидаемый ответ API:**
```json
{
  "id": "...",
  "name": "Kasza gryczana",
  "quantity": 500,
  "unit": "g",
  "price": {
    "value": 6.3,
    "per": "kg"
  },
  "computed": {
    "unitPrice": 0.0063,
    "totalCost": 3.15
  }
}
```

---

## 📖 Полная документация

| Файл | Описание | Строк |
|------|----------|-------|
| **CATEGORY_FILTERING_QUICK_START.md** | Быстрый фикс для frontend (структура ответа категорий) | 314 |
| **INGREDIENT_CATEGORIES_API_GUIDE.md** | Полная документация API категорий с примерами | ~400 |
| **FRIDGE_PRICE_RESPONSE_ARCHITECTURE.md** | Архитектура добавления цен в ответ /api/fridge/items | ~500 |

---

## 🔥 Приоритеты реализации

### HIGH (сейчас):
1. ✅ API категорий работает (deployed)
2. ⚠️ Frontend fix: использовать `result.data.categories`
3. ⚠️ Добавить `price` и `computed` в fridge response

### MEDIUM (скоро):
1. Изменить `category` → `categoryKey` в /api/fridge/items
2. Удалить хардкод категорий на frontend

### LOW (потом):
1. Удалить временную поддержку g/ml в normalizePrice()
2. Добавить статистику общей стоимости холодильника

---

## 🧪 Тестирование

### Categories API:
```bash
# Polish
curl -H "Accept-Language: pl" -H "Authorization: Bearer TOKEN" \
  https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/catalog/ingredient-categories

# Expected: {"success":true,"data":{"categories":[...]}}
```

### Fridge Items (после реализации):
```bash
curl -H "Authorization: Bearer TOKEN" \
  https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/fridge/items

# Expected: items with "price" and "computed" fields
```

---

## 📞 Questions?

- **Categories:** CATEGORY_FILTERING_QUICK_START.md
- **Prices:** FRIDGE_PRICE_RESPONSE_ARCHITECTURE.md
- **API Examples:** INGREDIENT_CATEGORIES_API_GUIDE.md
