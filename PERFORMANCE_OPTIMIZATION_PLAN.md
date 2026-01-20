# 🎯 Performance Optimization Plan - Fridge API

**Дата:** 20 января 2026  
**Статус:** Categories ✅ FIXED | Performance ⚠️ NEEDS OPTIMIZATION

---

## ✅ ЧТО РАБОТАЕТ ПРАВИЛЬНО

### 1️⃣ Категории — ИСПРАВЛЕНЫ ✅

**Доказательство из production logs (Koyeb):**

```
processing item {"ingredient_name":"Łosoś","ingredient_category":"fish"}
processing item {"ingredient_name":"Wołowina (rostbef)","ingredient_category":"meat"}
processing item {"ingredient_name":"Яица","ingredient_category":"egg"}
processing item {"ingredient_name":"Olej roślinny","ingredient_category":"condiment"}
processing item {"ingredient_name":"Соль","ingredient_category":"condiment"}
processing item {"ingredient_name":"Kefir","ingredient_category":"dairy"}
processing item {"ingredient_name":"Śmietana 18%","ingredient_category":"dairy"}
processing item {"ingredient_name":"Makaron ryżowy","ingredient_category":"grain"}
processing item {"ingredient_name":"Kasza gryczana","ingredient_category":"grain"}
```

**Результат:**
- ✅ `Ingredient.Category` читается корректно из БД
- ✅ `categoryKey` больше НЕ падает в `"other"`
- ✅ Все 9 продуктов имеют правильные категории (fish, meat, egg, dairy, condiment, grain)
- ✅ Каталог категорий работает как задумано
- ✅ Локализация через Accept-Language (pl/en/ru)

**Frontend TODO:**
- [ ] Hard refresh (Cmd+Shift+R) для очистки кэша
- [ ] Проверить: используется `item.categoryKey` не `item.category`
- [ ] Обновить TypeScript интерфейс если нужно

**Вывод:** КАТЕГОРИИ — ЗАКРЫТО. Отличная работа! 🎉

---

## ⚠️ ЧТО НУЖНО ОПТИМИЗИРОВАТЬ

### ❌ ПРОБЛЕМА №1 — N+1 SQL ЗАПРОСЫ НА ЦЕНУ (КРИТИЧНО 🔥)

#### 🔎 Что происходит сейчас

**Production logs показывают:**

```sql
-- Для КАЖДОГО продукта отдельный запрос:
SELECT * FROM "user_fridge_price_history"
WHERE user_fridge_item_id = '932c1f69-1454-44c8-9e54-b6062a5c0883'
ORDER BY created_at DESC
LIMIT 1

SELECT * FROM "user_fridge_price_history"
WHERE user_fridge_item_id = '51e71b0a-b97a-4bc3-bb70-41acf5f1c09c'
ORDER BY created_at DESC
LIMIT 1

-- ... еще 7 таких запросов
```

**Количество запросов:** 9 (по одному на каждый продукт)

#### 💀 Почему это плохо

| Кол-во продуктов | Кол-во SQL запросов | Время выполнения |
|------------------|---------------------|------------------|
| 10 | 11 запросов | ~300ms |
| 100 | **101 запрос** | ~3000ms (3 sec) |
| 1000 | **1001 запрос** | ~30 sec 💀 |

**Проблемы:**
- ⛔ Не масштабируется вообще
- ⛔ Под нагрузкой API ляжет
- ⛔ Каждый запрос: 17-35ms + network overhead
- ⛔ Классический N+1 problem

#### ✅ РЕШЕНИЕ: 2 ВАРИАНТА

---

### 🎯 ВАРИАНТ 1 (РЕКОМЕНДУЕМЫЙ) — Batch Query с DISTINCT ON

**Суть:** Один запрос вместо N запросов

#### SQL Query

```sql
-- PostgreSQL DISTINCT ON — получить последнюю цену для ВСЕХ items сразу
SELECT DISTINCT ON (uph.user_fridge_item_id)
  uph.user_fridge_item_id,
  uph.price_per_unit,
  uph.unit_for_price,
  uph.currency,
  uph.created_at
FROM user_fridge_price_history uph
WHERE uph.user_fridge_item_id IN (
  '932c1f69-1454-44c8-9e54-b6062a5c0883',
  '51e71b0a-b97a-4bc3-bb70-41acf5f1c09c',
  -- ... все ID продуктов
)
ORDER BY uph.user_fridge_item_id, uph.created_at DESC;
```

**Преимущества:**
- ✅ PostgreSQL делает это очень быстро (< 20ms для 100 items)
- ✅ Получаешь последнюю цену для ВСЕХ items сразу
- ✅ Один сетевой round-trip вместо N

#### Go Implementation

**Файл:** `internal/modules/fridge/service/fridge_service.go`

```go
// ✅ НОВЫЙ МЕТОД: Batch load последних цен для множества items
func (s *FridgeService) getLastPricesBatch(itemIDs []string) (map[string]*models.UserFridgePriceHistory, error) {
	if len(itemIDs) == 0 {
		return make(map[string]*models.UserFridgePriceHistory), nil
	}

	var prices []models.UserFridgePriceHistory
	
	// DISTINCT ON query - получает последнюю цену для каждого item_id
	err := s.db.Raw(`
		SELECT DISTINCT ON (user_fridge_item_id)
			id,
			user_fridge_item_id,
			price_per_unit,
			unit_for_price,
			currency,
			source,
			created_at
		FROM user_fridge_price_history
		WHERE user_fridge_item_id = ANY($1)
		ORDER BY user_fridge_item_id, created_at DESC
	`, pq.Array(itemIDs)).Scan(&prices).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to batch load prices: %w", err)
	}

	// Строим map для быстрого lookup
	priceMap := make(map[string]*models.UserFridgePriceHistory)
	for i := range prices {
		priceMap[prices[i].UserFridgeItemID] = &prices[i]
	}

	return priceMap, nil
}

// ✅ ОБНОВИТЬ: GetUserItemsV2 - использовать batch loading
func (s *FridgeService) GetUserItemsV2(userID string) ([]models.FridgeItemResponseV2, error) {
	// 1. Получаем все items пользователя (как сейчас)
	items, err := s.fridgeRepo.GetUserItems(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user items: %w", err)
	}

	if len(items) == 0 {
		return []models.FridgeItemResponseV2{}, nil
	}

	// 2. ✅ НОВОЕ: Собираем все item IDs
	itemIDs := make([]string, len(items))
	for i, item := range items {
		itemIDs[i] = item.ID
	}

	// 3. ✅ НОВОЕ: Загружаем все цены ОДНИМ запросом
	priceMap, err := s.getLastPricesBatch(itemIDs)
	if err != nil {
		logger.Error("failed to batch load prices", zap.Error(err))
		// Продолжаем без цен, не фейлим весь запрос
		priceMap = make(map[string]*models.UserFridgePriceHistory)
	}

	// 4. Формируем response (как сейчас, но берем цены из map)
	response := make([]models.FridgeItemResponseV2, len(items))
	for i, item := range items {
		// Берем цену из map вместо отдельного запроса
		lastPrice := priceMap[item.ID]

		var priceInfo *models.PriceInfo
		var computed *models.ComputedPrice

		if lastPrice != nil {
			priceInfo = &models.PriceInfo{
				Value: lastPrice.PricePerUnit,
				Per:   lastPrice.UnitForPrice,
			}

			// Вычисляем общую стоимость
			totalCost, unitPrice := s.computeTotalCost(
				item.Quantity,
				item.Unit,
				lastPrice.PricePerUnit,
				lastPrice.UnitForPrice,
			)

			computed = &models.ComputedPrice{
				UnitPrice: unitPrice,
				TotalCost: totalCost,
			}
		}

		response[i] = models.FridgeItemResponseV2{
			ID:          item.ID,
			Name:        item.Ingredient.GetLocalizedName("en"),
			CategoryKey: item.Ingredient.Category,
			Quantity:    item.Quantity,
			Unit:        item.Unit,
			DaysLeft:    s.calculateDaysLeft(item.ExpiresAt),
			Price:       priceInfo,
			Computed:    computed,
		}
	}

	return response, nil
}
```

**Изменения:**
1. ✅ Добавить метод `getLastPricesBatch()` - batch loading цен
2. ✅ Обновить `GetUserItemsV2()` - использовать batch вместо N запросов
3. ✅ Убрать старый метод `getLastPrice()` - он больше не нужен

**Результат:**
- 🚀 Вместо 9 запросов → **1 запрос**
- 🚀 Время выполнения: ~300ms → **~50ms**
- 🚀 Масштабируется на 1000+ продуктов без проблем

---

### 🎯 ВАРИАНТ 2 (БЫСТРЫЙ FIX) — Использовать current_price

**Суть:** Убрать history из GET вообще, использовать денормализованное поле

#### Преимущества

- ✅ Уже реализовано! (`current_price_per_unit` уже обновляется)
- ✅ Нет дополнительных SQL запросов
- ✅ Работает из коробки

#### Логика

**GET /api/fridge/items** возвращает:
```sql
SELECT 
  ufi.*,
  i.name_en, i.name_pl, i.name_ru, i.category
FROM user_fridge_items ufi
LEFT JOIN ingredients i ON i.id = ufi.ingredient_id
WHERE ufi.user_id = $1
```

**Цена уже в `ufi.current_price_per_unit`** — не нужны дополнительные запросы!

#### Go Implementation

```go
// ✅ ВАРИАНТ 2: Использовать current_price из user_fridge_items
func (s *FridgeService) GetUserItemsV2(userID string) ([]models.FridgeItemResponseV2, error) {
	items, err := s.fridgeRepo.GetUserItems(userID) // уже загружает current_price_per_unit
	if err != nil {
		return nil, fmt.Errorf("failed to get user items: %w", err)
	}

	response := make([]models.FridgeItemResponseV2, len(items))
	for i, item := range items {
		var priceInfo *models.PriceInfo
		var computed *models.ComputedPrice

		// ✅ Используем current_price вместо history
		if item.CurrentPricePerUnit != nil && *item.CurrentPricePerUnit > 0 {
			priceInfo = &models.PriceInfo{
				Value: *item.CurrentPricePerUnit,
				Per:   item.Unit, // или храните current_price_unit если нужно
			}

			totalCost, unitPrice := s.computeTotalCost(
				item.Quantity,
				item.Unit,
				*item.CurrentPricePerUnit,
				item.Unit,
			)

			computed = &models.ComputedPrice{
				UnitPrice: unitPrice,
				TotalCost: totalCost,
			}
		}

		response[i] = models.FridgeItemResponseV2{
			ID:          item.ID,
			Name:        item.Ingredient.GetLocalizedName("en"),
			CategoryKey: item.Ingredient.Category,
			Quantity:    item.Quantity,
			Unit:        item.Unit,
			DaysLeft:    s.calculateDaysLeft(item.ExpiresAt),
			Price:       priceInfo,
			Computed:    computed,
		}
	}

	return response, nil
}
```

#### История цен → Отдельный endpoint

```go
// GET /api/fridge/items/:id/price-history
func (h *FridgeHandlers) GetPriceHistory(w http.ResponseWriter, r *http.Request) {
	itemID := chi.URLParam(r, "id")
	
	history, err := h.service.GetPriceHistory(itemID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get price history")
		return
	}
	
	respondSuccess(w, map[string]interface{}{
		"history": history,
	})
}
```

**Когда использовать:**
- ✅ Для графиков изменения цены
- ✅ Для аналитики
- ✅ Для детальной страницы продукта

**Когда НЕ использовать:**
- ❌ В списке холодильника (GET /api/fridge/items)

---

### 🤔 Какой вариант выбрать?

| Критерий | Вариант 1 (Batch) | Вариант 2 (Current) |
|----------|-------------------|---------------------|
| **Скорость** | ⚡ Очень быстро (1 запрос) | ⚡⚡ Мгновенно (0 запросов) |
| **Сложность** | 🛠️ Средняя (новый метод) | ✅ Легко (уже работает) |
| **Точность цены** | 📊 Всегда актуальная | 📊 Кэшированная (может быть устаревшей) |
| **Масштабируемость** | 🚀 Отлично | 🚀 Идеально |

**Рекомендация:**
- 🎯 **ВАРИАНТ 2** для MVP (быстро, просто, работает)
- 🎯 **ВАРИАНТ 1** если нужна точность (всегда свежая цена из history)

---

## ❌ ПРОБЛЕМА №2 — SLOW SQL при SELECT по ID

### 🔎 Что происходит

```sql
-- SLOW SQL >= 200ms
SELECT * FROM "user_fridge_items"
WHERE id = '932c1f69-1454-44c8-9e54-b6062a5c0883'
LIMIT 1
-- Execution time: 200ms+ 💀
```

**Проблема:**
- ⛔ SELECT по `id` должен быть < 5ms
- ⛔ 200ms - это красный флаг

### 🐛 Причина (99%)

❌ Нет индекса на `id` (или `id` не является PRIMARY KEY)

### ✅ РЕШЕНИЕ

#### Проверить текущую схему

```sql
-- Проверяем, есть ли PRIMARY KEY
SELECT
  tc.table_name,
  kcu.column_name,
  tc.constraint_type
FROM information_schema.table_constraints tc
JOIN information_schema.key_column_usage kcu
  ON tc.constraint_name = kcu.constraint_name
WHERE tc.table_name = 'user_fridge_items'
  AND tc.constraint_type = 'PRIMARY KEY';
```

#### Вариант 1: Добавить PRIMARY KEY (если его нет)

```sql
-- Если PRIMARY KEY не установлен
ALTER TABLE user_fridge_items
ADD PRIMARY KEY (id);
```

#### Вариант 2: Добавить индекс (если PK уже на другом поле)

```sql
-- Если PK есть, но на другом поле
CREATE INDEX CONCURRENTLY idx_user_fridge_items_id
ON user_fridge_items(id);
```

**Результат:**
- 🚀 SELECT по ID: 200ms → **< 1ms**

---

## ❌ ПРОБЛЕМА №3 — ЛИШНИЙ SELECT перед INSERT

### 🔎 Что происходит

```sql
-- ЛИШНИЙ SELECT:
SELECT * FROM "user_fridge_items" WHERE id = '...'

-- Потом INSERT:
INSERT INTO user_fridge_price_history ...

-- Потом UPDATE:
UPDATE user_fridge_items SET current_price_per_unit = ...
```

**Вопрос:** ЗАЧЕМ этот SELECT?

### 🐛 Причина

- ❌ GORM делает лишний SELECT перед UPDATE/INSERT
- ❌ Если ты уже знаешь `item_id`, этот SELECT не нужен

### ✅ РЕШЕНИЕ

#### НЕПРАВИЛЬНО (текущая реализация):

```go
// ❌ GORM загружает весь item перед обновлением
var item models.UserFridgeItem
s.db.First(&item, "id = ?", itemID) // ЛИШНИЙ SELECT!

item.CurrentPricePerUnit = &price
s.db.Save(&item) // UPDATE
```

#### ПРАВИЛЬНО:

```go
// ✅ UPDATE напрямую без SELECT
s.db.Model(&models.UserFridgeItem{}).
	Where("id = ?", itemID).
	Updates(map[string]interface{}{
		"current_price_per_unit": price,
		"price_updated_at":       time.Now(),
	})
```

**Результат:**
- 🚀 Убираем 1 лишний SELECT на каждое обновление цены
- 🚀 POST /api/fridge/items работает быстрее

---

## 🧠 АРХИТЕКТУРНО ПРАВИЛЬНАЯ МОДЕЛЬ (РЕКОМЕНДУЮ)

### API Response Structure

```json
GET /api/fridge/items
{
  "success": true,
  "data": {
    "items": [
      {
        "id": "932c1f69-1454-44c8-9e54-b6062a5c0883",
        "name": "Śmietana 18%",
        "categoryKey": "dairy",
        "quantity": 1200,
        "unit": "ml",
        "daysLeft": 6,
        "arrivedAt": "2026-01-14T12:00:00Z",
        "expiresAt": "2026-01-26T12:00:00Z",
        "price": {
          "value": 7.45,
          "per": "l",
          "currency": "PLN"
        },
        "computed": {
          "unitPrice": 0.007,
          "totalCost": 8.94
        }
      }
    ]
  }
}
```

### История цен — Отдельный endpoint

```json
GET /api/fridge/items/:id/price-history
{
  "success": true,
  "data": {
    "history": [
      {
        "id": "...",
        "pricePerUnit": 7.45,
        "unitForPrice": "l",
        "currency": "PLN",
        "source": "manual",
        "createdAt": "2026-01-20T10:00:00Z"
      },
      {
        "id": "...",
        "pricePerUnit": 6.90,
        "unitForPrice": "l",
        "currency": "PLN",
        "source": "manual",
        "createdAt": "2026-01-13T15:30:00Z"
      }
    ],
    "analysis": {
      "trend": "up",
      "percentChange": 7.97,
      "lastPrice": 7.45,
      "previousPrice": 6.90
    }
  }
}
```

**Принцип:**
- ✅ История цен НЕ в списке (не нужна в 99% случаев)
- ✅ Отдельный endpoint для аналитики/графиков
- ✅ Загружается только когда нужно

---

## 📋 ЧЕКЛИСТ ИСПРАВЛЕНИЙ

### 🔥 ОБЯЗАТЕЛЬНО (Блокирует production)

- [ ] **Убрать N+1 price queries**
  - [ ] Вариант 1: Реализовать `getLastPricesBatch()` с DISTINCT ON
  - [ ] Вариант 2: Использовать `current_price_per_unit` из items
  - [ ] Обновить `GetUserItemsV2()` использовать batch loading
  - [ ] Убрать старый метод `getLastPrice()` (цикл по items)
  - [ ] Тестировать: 10, 100, 1000 продуктов

- [ ] **Добавить индекс/PK на `user_fridge_items.id`**
  - [ ] Проверить текущую схему (есть ли PK?)
  - [ ] Добавить `PRIMARY KEY (id)` или индекс
  - [ ] Проверить: SELECT по ID должен быть < 5ms

- [ ] **Использовать `current_price_*` в GET**
  - [ ] Убрать загрузку history в GetUserItemsV2
  - [ ] Использовать денормализованные поля из items
  - [ ] История → отдельный endpoint `/price-history`

### 👍 ЖЕЛАТЕЛЬНО (Оптимизация)

- [ ] **Убрать лишний SELECT перед INSERT**
  - [ ] Заменить `db.First() + db.Save()` на `db.Updates()`
  - [ ] Убрать загрузку item перед обновлением цены
  - [ ] Тестировать: POST /api/fridge/items должен быть быстрее

- [ ] **Разделить current price и price history**
  - [ ] GET /api/fridge/items → current_price
  - [ ] GET /api/fridge/items/:id/price-history → полная история
  - [ ] Добавить endpoint для аналитики цен

- [ ] **Убрать debug logging**
  - [ ] Удалить `logger.Info()` из GetUserItemsV2 loop
  - [ ] Оставить только Error/Warn логи
  - [ ] Commit: "refactor: remove debug logging after performance fix"

### 🎯 БОНУС (Nice to have)

- [ ] **Добавить кэширование категорий**
  - [ ] Redis для GET /api/catalog/ingredient-categories
  - [ ] TTL: 1 час (категории редко меняются)

- [ ] **Добавить pagination для GET /api/fridge/items**
  - [ ] Query params: `?limit=50&offset=0`
  - [ ] Для пользователей с 100+ продуктами

---

## 🟢 ХОРОШАЯ НОВОСТЬ

### Ты уже решил самые сложные проблемы! 🎉

✅ **Категории** — работают идеально (fish, meat, egg, dairy, condiment, grain)  
✅ **Локализация** — Accept-Language (pl/en/ru)  
✅ **Модель данных** — правильная архитектура (event sourcing для цен)  
✅ **История цен** — реализовано (user_fridge_price_history)  
✅ **Notification pipeline** — работает (system notifications)  
✅ **Price computation** — правильная нормализация (kg→g, l→ml)

### Осталось отполировать перформанс

Это **финальный этап перед production**:

1. 🔥 Убрать N+1 queries (критично)
2. 🔥 Добавить индексы (обязательно)
3. 👍 Оптимизировать SQL (желательно)

**Время на исправление:** 1-2 часа  
**Результат:** Production-ready API 🚀

---

## 📚 Связанные документы

- `PRICE_AND_CATEGORY_FIX_COMPLETE.md` - История исправления categoryKey
- `FRIDGE_API_DOCUMENTATION.md` - API документация
- `CANONICAL_INGREDIENTS_GUIDE.md` - Архитектура каталога

---

## 🎯 Next Steps

1. **Выбрать вариант оптимизации цены:**
   - Вариант 1 (Batch) - точность
   - Вариант 2 (Current) - скорость

2. **Реализовать исправления:**
   - Обновить `GetUserItemsV2()`
   - Добавить индекс на `id`
   - Убрать лишние SELECT

3. **Тестировать:**
   - GET /api/fridge/items с 10, 100, 1000 продуктами
   - Проверить время выполнения (должно быть < 100ms)

4. **Deploy и мониторинг:**
   - Проверить production logs
   - Убедиться: N+1 исчез
   - Убрать debug logging

**Готово к production!** 🚀
