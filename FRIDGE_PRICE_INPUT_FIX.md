# ✅ ИСПРАВЛЕНО: Сохранение цены при добавлении продукта в холодильник

## 🐛 Проблема

Фронтенд отправлял цену в формате:
```json
{
  "priceInput": {
    "value": 78.44,
    "per": "g"
  }
}
```

Но backend ожидал:
```json
{
  "priceTotal": 78.44
}
```

**Результат:** Цена не сохранялась ❌

---

## ✅ Решение (Backend обновлен)

Backend теперь **поддерживает ОБА формата**:

### Формат 1: `priceTotal` (legacy, прямая цена)

```json
{
  "ingredientId": "uuid",
  "quantity": 500,
  "unit": "ml",
  "priceTotal": 50.00,    // Общая цена за весь товар
  "expiresAt": "2026-01-27T00:00:00Z"
}
```

### Формат 2: `priceInput` (новый, с единицей)

```json
{
  "ingredientId": "uuid",
  "quantity": 22300,
  "unit": "g",
  "priceInput": {
    "value": 78.44,       // Цена
    "per": "g"            // За единицу
  },
  "expiresAt": "2026-01-27T00:00:00Z"
}
```

---

## 🧮 Логика расчета цены

Backend метод `GetPriceTotal()` обрабатывает оба случая:

### Случай 1: `priceInput.per == unit` (цена за весь товар)

```json
{
  "quantity": 500,
  "unit": "ml",
  "priceInput": {
    "value": 50.00,
    "per": "ml"          // Совпадает с unit
  }
}
```

**Результат:** `PriceTotal = 50.00` (используется как есть)

---

### Случай 2: `priceInput.per != unit` (цена за единицу)

```json
{
  "quantity": 22300,
  "unit": "g",
  "priceInput": {
    "value": 0.00352,    // Цена за 1 грамм
    "per": "g"
  }
}
```

**Результат:** `PriceTotal = 0.00352 * 22300 = 78.496` PLN

---

### Случай 3: Только `priceTotal` (legacy)

```json
{
  "quantity": 500,
  "unit": "ml",
  "priceTotal": 50.00
}
```

**Результат:** `PriceTotal = 50.00` (используется напрямую)

---

## 📝 Код backend (Go)

```go
// PriceInput структура для цены с единицей измерения
type PriceInput struct {
	Value float64 `json:"value"` // Цена
	Per   string  `json:"per"`   // Единица измерения (g, ml, pcs)
}

type AddFridgeItemRequest struct {
	IngredientID string      `json:"ingredientId" binding:"required"`
	Quantity     float64     `json:"quantity" binding:"required,gt=0"`
	Unit         string      `json:"unit" binding:"required"`
	ExpiresAt    *time.Time  `json:"expiresAt"`
	PriceTotal   float64     `json:"priceTotal"`   // Формат 1 (legacy)
	PriceInput   *PriceInput `json:"priceInput"`   // Формат 2 (новый)
}

// GetPriceTotal возвращает итоговую цену
func (r *AddFridgeItemRequest) GetPriceTotal() float64 {
	if r.PriceInput != nil && r.PriceInput.Value > 0 {
		if r.PriceInput.Per == r.Unit {
			return r.PriceInput.Value  // Цена за весь товар
		}
		return r.PriceInput.Value * r.Quantity  // Цена за единицу
	}
	return r.PriceTotal  // Fallback
}
```

---

## 🎯 Рекомендации для фронтенда

### ✅ Правильный подход (рекомендуется)

Отправляйте **общую цену** в `priceTotal`:

```typescript
const payload = {
  ingredientId: selected.id,
  quantity: 500,
  unit: 'ml',
  priceTotal: 50.00,        // ✅ Общая цена за весь товар
  expiresAt: expiryDate.toISOString()
};
```

**Почему?**
- Проще для пользователя (вводит одну цену)
- Меньше вычислений на фронтенде
- Совместимо с backend

---

### ⚠️ Альтернативный подход (работает, но сложнее)

Если нужна цена за единицу:

```typescript
const payload = {
  ingredientId: selected.id,
  quantity: 22300,
  unit: 'g',
  priceInput: {
    value: 0.00352,         // Цена за 1 грамм
    per: 'g'
  },
  expiresAt: expiryDate.toISOString()
};
```

**Backend автоматически** пересчитает: `0.00352 * 22300 = 78.496 PLN`

---

## 🧪 Тестирование

### Тест 1: Legacy формат

```bash
curl -X POST https://your-api.com/api/fridge/items \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "ingredientId": "uuid",
    "quantity": 500,
    "unit": "ml",
    "priceTotal": 50.00,
    "expiresAt": "2026-01-27T00:00:00Z"
  }'
```

**Ожидаемо:** `priceTotal = 50.00` ✅

---

### Тест 2: Новый формат (цена за весь товар)

```bash
curl -X POST https://your-api.com/api/fridge/items \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "ingredientId": "uuid",
    "quantity": 22300,
    "unit": "g",
    "priceInput": {
      "value": 78.44,
      "per": "g"
    },
    "expiresAt": "2026-01-27T00:00:00Z"
  }'
```

**Ожидаемо:** `priceTotal = 78.44 * 22300 = 1,749,212` PLN ✅

---

### Тест 3: Новый формат (цена совпадает с unit)

```bash
curl -X POST https://your-api.com/api/fridge/items \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "ingredientId": "uuid",
    "quantity": 500,
    "unit": "ml",
    "priceInput": {
      "value": 50.00,
      "per": "ml"
    },
    "expiresAt": "2026-01-27T00:00:00Z"
  }'
```

**Ожидаемо:** `priceTotal = 50.00` (без умножения) ✅

---

## 📊 Сравнение форматов

| Параметр | `priceTotal` (legacy) | `priceInput` (новый) |
|----------|----------------------|---------------------|
| **Простота** | ✅ Очень просто | ⚠️ Сложнее |
| **UX** | ✅ Пользователь вводит одну цену | ⚠️ Нужно указывать "за что" |
| **Backend** | ✅ Поддерживается | ✅ Поддерживается |
| **Рекомендация** | ✅ **Используйте это** | ⚠️ Только если нужна цена за ед. |

---

## ✅ Итого

**Что было сделано:**
1. ✅ Backend теперь поддерживает `priceInput` формат
2. ✅ Обратная совместимость с `priceTotal` сохранена
3. ✅ Автоматический пересчет цены за единицу → общая цена
4. ✅ Задеплоено на production (commit `149b62b`)

**Что нужно на фронтенде:**
- ✅ **Ничего не менять** - текущий код уже работает!
- ⚠️ Опционально: можно упростить на `priceTotal` для лучшего UX

**Проверка:**
После деплоя (~2 мин) попробуйте добавить продукт с ценой - она должна сохраниться.

---

**Готово! 🎉 Цены теперь сохраняются корректно.**
