# ✅ Фильтрация expired продуктов - ИСПРАВЛЕНО

**Date:** January 15, 2026  
**Issue:** Expired продукты возвращаются в GET /api/fridge/items  
**Status:** ✅ **FIXED**  

---

## 🐛 Проблема

### Было:
```bash
GET /api/fridge/items

{
  "data": {
    "items": [
      {
        "name": "Milk",
        "daysLeft": -24,   // ❌ Просрочен 24 дня назад!
        "status": "expired"
      }
    ]
  }
}
```

**Проблема:** Expired продукты попадали в основной список холодильника

---

## ✅ Исправление

### 1. Фильтрация в V1 Service

**File:** `internal/modules/fridge/service/fridge_service.go`

#### GetUserFridgeItems (основной список):
```go
for _, item := range items {
    daysLeft := s.calculateDaysLeft(item.ExpiresAt)
    status := models.GetFridgeItemStatus(daysLeft)

    // ❌ НЕ отдаём expired продукты в основной список
    if status == "expired" {
        continue
    }

    response := models.FridgeItemListResponse{
        DaysLeft: daysLeft,
        Status:   status,
    }
    result = append(result, response)
}
```

#### GetExpiringSoon (истекающие продукты):
```go
for _, item := range items {
    daysLeft := s.calculateDaysLeft(item.ExpiresAt)
    status := models.GetFridgeItemStatus(daysLeft)

    // ❌ НЕ отдаём expired продукты
    if status == "expired" {
        continue
    }

    result = append(result, response)
}
```

---

### 2. Фильтрация в V2 Service

**File:** `internal/modules/fridge/service/fridge_service_v2.go`

```go
func (s *fridgeServiceV2) GetItems(userID string) ([]models.FridgeItem, error) {
    var items []models.FridgeItem
    
    err := s.db.Where("user_id = ?", userID).Find(&items).Error

    freshItems := make([]models.FridgeItem, 0, len(items))
    for i := range items {
        result := EvaluateFridgeItem(&items[i])
        items[i].Status = result.Status

        // Обновляем БД если просрочен
        if items[i].Status == models.FridgeItemStatusExpired {
            s.db.Model(&items[i]).Updates(map[string]interface{}{
                "status": result.Status,
            })
            // ❌ НЕ добавляем expired в результат
            continue
        }

        // ✅ Добавляем только fresh/ok
        freshItems = append(freshItems, items[i])
    }

    return freshItems, nil
}
```

---

## 🔧 Логика статусов

### File: `internal/models/user_fridge.go`

```go
func GetFridgeItemStatus(daysLeft *int) string {
    if daysLeft == nil {
        return "fresh"  // Нет срока годности
    }
    if *daysLeft < 0 {
        return "expired"  // ❌ Просрочен
    }
    if *daysLeft <= 2 {
        return "warning"  // ⚠️ Истекает
    }
    return "ok"  // ✅ Свежий
}
```

---

## 📊 Ожидаемое поведение

### После исправления:

#### GET /api/fridge/items (основной холодильник):
```json
{
  "data": {
    "items": [
      {
        "name": "Oliwa",
        "daysLeft": null,
        "status": "fresh"      // ✅ Без срока годности
      },
      {
        "name": "Czosnek",
        "daysLeft": 59,
        "status": "ok"         // ✅ Свежий
      },
      {
        "name": "Pomidor",
        "daysLeft": 1,
        "status": "warning"    // ⚠️ Истекает завтра
      }
      // ❌ Expired продукты НЕ показываются
    ]
  }
}
```

#### GET /api/fridge/items/expiring (истекающие):
```json
{
  "data": [
    {
      "name": "Pomidor",
      "daysLeft": 1,
      "status": "warning"
    }
    // ✅ Только warning/ok, БЕЗ expired
  ]
}
```

#### GET /api/notifications (уведомления):
```json
{
  "data": [
    {
      "type": "item_expired",
      "itemName": "Milk",
      "message": "Milk просрочилось 24 дня назад"
    }
  ]
}
```

**Разделение:**
- ✅ **Основной холодильник:** fresh, ok, warning (БЕЗ expired)
- ✅ **Уведомления:** expired продукты видны здесь
- ✅ **БД:** expired продукты сохраняются для статистики

---

## 🧪 Тестирование

### Test 1: Expired не попадает в список
```bash
# 1. Добавить продукт с истекшей датой
curl -X POST https://.../api/fridge/items \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "ingredientId": "...",
    "quantity": 1,
    "expiresAt": "2024-01-01"  // Прошлый год
  }'

# 2. Проверить основной список
curl https://.../api/fridge/items \
  -H "Authorization: Bearer $TOKEN" | jq '.data.items[] | select(.status == "expired")'

# Ожидается: пустой результат (expired продукты скрыты)
```

### Test 2: Warning продукты видны
```bash
curl https://.../api/fridge/items \
  -H "Authorization: Bearer $TOKEN" | jq '.data.items[] | select(.status == "warning")'

# Ожидается: продукты с daysLeft <= 2
```

### Test 3: Fresh продукты без срока
```bash
curl https://.../api/fridge/items \
  -H "Authorization: Bearer $TOKEN" | jq '.data.items[] | select(.daysLeft == null)'

# Ожидается: продукты с status = "fresh", daysLeft = null
```

---

## 📈 Статусы продуктов

| DaysLeft | Status | Показывать в UI? | Где видно? |
|----------|--------|------------------|------------|
| `null` | `fresh` | ✅ Да | Основной холодильник |
| `> 2` | `ok` | ✅ Да | Основной холодильник |
| `<= 2` | `warning` | ✅ Да | Основной холодильник + предупреждение |
| `< 0` | `expired` | ❌ Нет | Только в уведомлениях |

---

## 🔐 Правила

### ✅ ПРАВИЛЬНО:

1. **Основной список холодильника:**
   - Показываем: `fresh`, `ok`, `warning`
   - Скрываем: `expired`

2. **Уведомления:**
   - Показываем: `warning` (истекает), `expired` (просрочено)

3. **База данных:**
   - Сохраняем все статусы для статистики
   - Обновляем status автоматически при каждом GET

4. **CRON задача:**
   - Проверяет все продукты 1 раз в день
   - Создаёт уведомления для expired
   - НЕ удаляет из БД (для истории потерь)

### ❌ НЕПРАВИЛЬНО:

```go
// ❌ Не делать так:
if daysLeft < 0 {
    return response  // Вернёт expired в основной список
}

// ✅ Правильно:
if status == "expired" {
    continue  // Пропустить expired продукты
}
```

---

## 🎯 Результат

| Было | Стало |
|------|-------|
| ❌ Expired в основном списке | ✅ Expired только в уведомлениях |
| ❌ daysLeft: -24 в UI | ✅ Продукт скрыт из списка |
| ❌ "Осталось -24 дня" | ✅ Уведомление "Просрочено 24 дня назад" |

---

**Last Updated:** January 15, 2026  
**Changes:** 3 files modified (V1, V2, documentation)  
**Status:** ✅ Ready for deployment
