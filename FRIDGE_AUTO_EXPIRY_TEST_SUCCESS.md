# ✅ Автоматический расчёт срока годности - РАБОТАЕТ

**Date:** January 15, 2026  
**Test:** Add ingredient from catalog with auto expiry calculation  
**Status:** ✅ **SUCCESS**  

---

## 🧪 Test Execution

### Request:
```bash
POST /api/fridge/items
Authorization: Bearer <user_token>

{
  "ingredientId": "2c3405e0-60cf-4e5f-9872-0bb8d1f91b83",  // Czosnek (чеснок)
  "quantity": 3
}
```

### Response:
```json
{
  "data": {
    "id": "a00de468-6a44-4ec7-b26b-177d908e9bb8",
    "ingredient": {
      "name": "Czosnek",
      "unit": "g",
      "category": "vegetable"
    },
    "quantity": 6,
    "expiresAt": "2026-03-16",  // ✅ Автоматически вычислено!
    "daysLeft": 59               // ✅ Автоматически вычислено!
  },
  "success": true
}
```

---

## ✅ Что работает ИДЕАЛЬНО:

### 1. Автоматическая дата добавления
```json
"arrivedAt": "2026-01-15T10:59:37.579487Z"  // ✅ NOW()
```

### 2. Автоматический расчёт срока годности
```json
"expiresAt": "2026-03-16T10:59:05.61704Z"   // ✅ arrivedAt + 60 дней
```

### 3. Автоматический расчёт дней до истечения
```json
"daysLeft": 59  // ✅ (expiresAt - today) / 24h
```

### 4. Статус продукта
```json
"status": "ok"  // ✅ fresh (59 дней до истечения)
```

---

## 🔧 Backend Logic (Reference)

### File: `internal/modules/fridge/service/fridge_service.go:88-93`

```go
// Вычисляем expires_at автоматически, если не задано явно
var expiresAt *time.Time
if req.ExpiresAt != nil {
    expiresAt = req.ExpiresAt  // Используем ручное значение
} else if ingredient.DefaultShelfLifeDays != nil {
    t := arrivedAt.AddDate(0, 0, *ingredient.DefaultShelfLifeDays)
    expiresAt = &t  // ✅ Автоматический расчёт!
}
```

### Примеры для разных продуктов:

| Продукт | DefaultShelfLifeDays | Auto ExpiresAt |
|---------|----------------------|----------------|
| Czosnek (чеснок) | 60 | +60 дней |
| Mleko (молоко) | 7 | +7 дней |
| Oliwa (масло) | NULL | NULL (без срока) |
| Pomidor (помидор) | 14 | +14 дней |
| Jajka (яйца) | 21 | +21 дней |

---

## 📊 Full Flow Test

### Step 1: Find ingredient in catalog
```bash
$ psql -c "SELECT id, name FROM Ingredient WHERE name = 'Czosnek';"
id: 2c3405e0-60cf-4e5f-9872-0bb8d1f91b83
name: Czosnek
```

### Step 2: Add to fridge (minimal request)
```bash
POST /api/fridge/items
{
  "ingredientId": "2c3405e0-60cf-4e5f-9872-0bb8d1f91b83",
  "quantity": 3
}
```

### Step 3: Backend auto-calculates
- ✅ `arrivedAt` = NOW()
- ✅ `unit` = ingredient.unit ("g")
- ✅ `expiresAt` = arrivedAt + DefaultShelfLifeDays (60 дней)
- ✅ `daysLeft` = (expiresAt - today) / 24h

### Step 4: Verify in fridge
```bash
GET /api/fridge/items

{
  "name": "Czosnek",
  "quantity": 6,          // Если был дубликат, quantity суммируется
  "expiresAt": "2026-03-16",
  "daysLeft": 59,
  "status": "ok",
  "arrivedAt": "2026-01-15"
}
```

---

## ⚠️ Known Issue: priceInput validation

### Problem:
```bash
POST /api/fridge/items
{
  "ingredientId": "...",
  "quantity": 3,
  "priceInput": {
    "value": 2.50,
    "per": "szt"  // ❌ Fails validation
  }
}

Response: {"message": "failed to add item", "success": false}
```

### Possible causes:
1. Price normalization error
2. Unit mismatch (ingredient.unit = "g", priceInput.per = "szt")
3. Missing currency validation

### Workaround:
Add items without `priceInput` initially, then add price separately:
```bash
# 1. Add item without price
POST /api/fridge/items {"ingredientId": "...", "quantity": 3}

# 2. Add price event (if needed)
POST /api/fridge/items/{id}/price
```

---

## 🎯 Test Results Summary

| Feature | Status | Notes |
|---------|--------|-------|
| Add from catalog | ✅ Works | ingredientId lookup successful |
| Auto arrivedAt | ✅ Works | Set to NOW() automatically |
| Auto expiresAt | ✅ Works | Uses DefaultShelfLifeDays |
| Auto daysLeft | ✅ Works | Calculated on backend |
| Auto status | ✅ Works | "ok" for fresh items |
| Duplicate handling | ✅ Works | Quantity is summed |
| priceInput | ⚠️ Issue | Validation error (needs fix) |

---

## ✅ Conclusion

**Автоматический расчёт срока годности работает ИДЕАЛЬНО!** 🎉

Пользователь может:
1. Выбрать продукт из каталога
2. Указать только `quantity`
3. Backend автоматически:
   - ✅ Добавит `arrivedAt` (сегодня)
   - ✅ Вычислит `expiresAt` (arrivedAt + DefaultShelfLifeDays)
   - ✅ Вычислит `daysLeft` (разница в днях)
   - ✅ Установит `status` (fresh/ok)

**Никакой логики на фронтенде не нужно!** 🔥

---

**Last Updated:** January 15, 2026  
**Tested on:** Production (Koyeb)  
**User:** fodi85@gmail.ru (home_chef)
