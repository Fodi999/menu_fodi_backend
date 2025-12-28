# 🧹 Автоматическая Очистка Просроченных Продуктов

## 📋 Обзор

Система автоматически удаляет просроченные продукты из холодильника и переносит их в корзину отходов с полным трекингом потерь.

## 🔄 Принцип Работы

### Триггеры Автоматической Очистки

Проверка просроченных продуктов происходит в следующих случаях:

1. **При получении списка холодильника** (`GET /api/fridge/items`)
   - Перед возвратом списка автоматически вызывается `cleanupExpiredItems()`
   - Пользователь не видит просроченные продукты в списке
   - Если очистка не удалась, запрос всё равно успешен (fail-safe)

2. **При попытке обновить количество** (`PATCH /api/fridge/items/{id}`)
   - Если продукт просрочен, вместо обновления он удаляется
   - Возвращается 404 (продукт больше не существует)
   - Создаётся событие в истории потерь

## 🔍 Алгоритм Очистки

```go
func cleanupExpiredItems(userID string) error {
    // 1. Найти все просроченные продукты
    WHERE user_id = ? AND expires_at IS NOT NULL AND expires_at < NOW()
    
    // 2. Для каждого просроченного продукта:
    for item in expiredItems {
        // 2.1 Рассчитать стоимость потери
        cost = quantity × current_price_per_unit
        
        // 2.2 Вычислить дни в холодильнике
        daysInFridge = NOW() - arrived_at
        
        // 2.3 Создать событие в истории
        INSERT INTO history_events (
            event_type: "expired",
            source_type: "auto",
            metadata: {
                ingredient_id, name, quantity, unit,
                cost, price_per_unit, currency,
                expiry_date, arrived_at, days_in_fridge,
                reason: "expiry_date_passed",
                context: "auto_cleanup_on_list"
            }
        )
        
        // 2.4 Удалить из холодильника
        DELETE FROM user_fridge_items WHERE id = item.id
    }
}
```

## 📊 Метаданные События

Каждое событие потери содержит полную информацию:

```json
{
  "ingredient_id": "uuid",
  "ingredient_name": "Молоко",
  "quantity": 1.0,
  "unit": "л",
  "cost": 4.5,
  "price_per_unit": 0.0045,
  "currency": "PLN",
  "expiry_date": "2024-12-25T00:00:00Z",
  "arrived_at": "2024-12-20T10:30:00Z",
  "days_in_fridge": 5,
  "reason": "expiry_date_passed",
  "context": "auto_cleanup_on_list"
}
```

## 🎯 API Endpoints

### GET /api/history/losses?days=30

Получить статистику потерь за период (включая просроченные продукты):

**Ответ:**
```json
{
  "events": [
    {
      "id": "uuid",
      "name": "Молоко",
      "quantity": 1.0,
      "unit": "л",
      "loss": 4.5,
      "reason": "expiry_date_passed",
      "addedDate": "2024-12-20T10:30:00Z",
      "expiryDate": "2024-12-25T00:00:00Z",
      "daysInFridge": 5
    }
  ],
  "summary": {
    "totalProducts": 10,
    "totalValue": 45.0,
    "avgValue": 4.5,
    "currency": "PLN"
  }
}
```

## 🔧 Ручная Очистка (Опционально)

Для массовой очистки всех пользователей можно использовать команду:

```bash
# В Koyeb добавить переменную DATABASE_URL и запустить:
./bin/cleanup_expired

# Локально:
DATABASE_URL="postgresql://..." go run cmd/cleanup_expired/main.go
```

**Вывод:**
```
✅ Connected to database

🧹 Starting expired items cleanup...
==================================================

🗑️  Found 3 expired items

[1/3] Processing: Молоко (1.00 л)
  ✅ Removed | Cost: 4.50 PLN | Expired: 2024-12-25

[2/3] Processing: Сыр (200.00 g)
  ✅ Removed | Cost: 8.20 PLN | Expired: 2024-12-23

[3/3] Processing: Йогурт (150.00 ml)
  ✅ Removed | Cost: 2.30 PLN | Expired: 2024-12-24

==================================================
📊 Cleanup Summary:
   ✅ Successfully processed: 3 items
   ⚠️  Failed: 0 items
   💰 Total loss: 15.00 PLN

🎉 Cleanup completed!
```

## 📈 Преимущества

1. **Автоматизация** - Пользователь не думает об очистке
2. **Полный Трекинг** - Все потери фиксируются в истории
3. **Аналитика** - Статистика помогает сократить food waste
4. **Безопасность** - Fail-safe подход (не блокирует основные запросы)
5. **Точность** - Рассчитывается реальная стоимость потерь

## 🚀 Развитие

### Текущая Версия (MVP)
- ✅ Автоматическое удаление при запросе списка
- ✅ Проверка при обновлении количества
- ✅ Полное логирование событий
- ✅ Расчёт стоимости потерь

### Будущие Улучшения
- 🔜 Push-уведомления за день до истечения срока
- 🔜 Email-отчёты о потерях за месяц
- 🔜 AI-рекомендации по сокращению waste
- 🔜 Прогноз потерь на основе истории
- 🔜 Интеграция с "умными" холодильниками

## 🔗 Связанные Документы

- [Архитектура системы](./SYSTEM_ARCHITECTURE_SUMMARY.md)
- [AI Insights Roadmap](./AI_FRIDGE_INSIGHTS_ROADMAP.md)
- [История событий](./RECIPE_FRIDGE_INTEGRATION.md)
