# 🧊 Система предотвращения потерь в холодильнике

## 📋 Обзор

**Цель**: Backend сам определяет просроченные/истекающие продукты и создает умные уведомления с помощью AI.

## 🎯 Ключевые принципы

1. ✅ Backend вычисляет статус продукта (НЕ фронтенд)
2. ✅ `expired` ≠ `discarded` (просрочен ≠ выброшен)
3. ✅ Уведомления создаются CRON или при GET `/api/fridge/items`
4. ✅ AI формирует текст, НЕ логику
5. ✅ Уведомления НЕ дублируются (проверка по дате)
6. ✅ Ничего не удаляется физически

---

## 📊 Модели данных

### FridgeItem

```go
type FridgeItem struct {
    ID           string           // UUID
    UserID       string           // Владелец
    IngredientID string           // Ссылка на каталог
    Quantity     float64          // Количество
    Unit         string           // g, ml, pcs
    ExpiresAt    *time.Time       // Срок годности (nullable)
    Status       FridgeItemStatus // fresh, expired, discarded
    DaysLeft     *int             // Вычисляется на backend
    PriceTotal   float64          // Для расчета потерь
    CreatedAt    time.Time
    UpdatedAt    time.Time
}
```

**Статусы:**
- `fresh` - продукт свежий
- `expired` - срок истек (автоматически)
- `discarded` - выброшен пользователем

### Notification

```go
type Notification struct {
    ID        string            // UUID
    UserID    string            // Кому
    Type      NotificationType  // system, order, user, fridge, ai, backup
    Level     NotificationLevel // info, warning, critical
    Title     string            // Заголовок
    Message   string            // Текст (может быть от AI)
    Meta      *string           // JSON: {fridgeItemId, daysLeft}
    ReadAt    *time.Time        // Null = не прочитано
    CreatedAt time.Time
}
```

**Типы уведомлений:**
- `fridge` - системные уведомления о холодильнике
- `ai` - AI-рекомендации

**Уровни:**
- `info` - 2-3 дня до истечения
- `warning` - 1 день до истечения  
- `critical` - истекает сегодня или просрочено

---

## 🔄 Логика определения статуса

### Функция `EvaluateFridgeItem`

```go
func EvaluateFridgeItem(item *FridgeItem) FridgeEvaluationResult {
    if item.ExpiresAt == nil {
        return { Status: "fresh", DaysLeft: nil }
    }

    today := startOfDay(now)
    expiry := startOfDay(item.ExpiresAt)
    
    daysLeft := differenceInDays(expiry, today)

    if daysLeft < 0 {
        return { Status: "expired", DaysLeft: daysLeft }
    }

    return { Status: "fresh", DaysLeft: daysLeft }
}
```

### Правило уровней уведомлений

| `daysLeft` | `level`    | Смысл                       |
|------------|------------|----------------------------|
| 3-2        | `info`     | Напоминание                |
| 1          | `warning`  | Срочно использовать        |
| 0          | `critical` | Истекает сегодня           |
| < 0        | `critical` | Просрочено (потеря денег)  |

---

## ⚙️ Когда создавать уведомления

### ❌ НЕправильно:
- При каждом запросе `/api/fridge/items`
- При каждом обновлении продукта

### ✅ Правильно:
1. **CRON** (1 раз в день, утром) - идеально
2. **При первом GET** `/api/fridge/items` за день (fallback)

---

## 🤖 Интеграция AI

AI **НЕ нужен** для вычислений.  
AI нужен для **генерации текста**.

### Пример prompt

```
User has ingredient "Карась", 1200 g
Expires in 1 day
Total value: 14.76 PLN

Generate a short warning message encouraging to cook it today.
Keep it under 100 characters, urgent tone.
```

### AI возвращает:

```
Приготовь карася сегодня — завтра он станет непригодным, 
и ты потеряешь 14.76 PLN.
```

Backend сохраняет это как `message` в уведомлении.

---

## 📱 Примеры уведомлений

### 1. Warning (1 день)

```
type: "ai"
level: "warning"
title: "Используй карася сегодня"
message: "AI рекомендует приготовить карася, иначе ты потеряешь 14.76 PLN"
meta: {"fridgeItemId": "uuid", "daysLeft": 1}
```

### 2. Critical (просрочено)

```
type: "fridge"
level: "critical"
title: "Продукт просрочен"
message: "Карась просрочен. Потеря: 14.76 PLN"
meta: {"fridgeItemId": "uuid", "daysLeft": -2}
```

### 3. Info (2-3 дня)

```
type: "ai"
level: "info"
title: "Скоро истечет срок"
message: "Карась истечет через 2 дня. Не забудь использовать!"
meta: {"fridgeItemId": "uuid", "daysLeft": 2}
```

---

## 🔧 API Endpoints

### GET /api/fridge/items

**Ответ:**
```json
{
  "data": [
    {
      "id": "uuid",
      "ingredientId": "uuid",
      "quantity": 1200,
      "unit": "g",
      "expiresAt": "2026-01-16T00:00:00Z",
      "status": "fresh",
      "daysLeft": 1,
      "priceTotal": 14.76
    }
  ]
}
```

**Backend автоматически:**
1. Вычисляет `daysLeft`
2. Обновляет `status` если просрочен
3. Создает уведомления (если нужно и не было сегодня)

### GET /api/notifications

**Ответ:**
```json
{
  "data": [
    {
      "id": "uuid",
      "type": "ai",
      "level": "warning",
      "title": "Используй карася сегодня",
      "message": "AI рекомендует приготовить...",
      "readAt": null,
      "createdAt": "2026-01-15T08:00:00Z"
    }
  ]
}
```

---

## 🔄 CRON задача (рекомендуется)

### Ежедневная проверка (утром)

```go
// Запускается каждый день в 08:00
func DailyFridgeCheck() {
    users := getAllActiveUsers()
    
    for _, user := range users {
        CheckAndNotifyExpiringItems(db, user.ID)
    }
}
```

### Алгоритм `CheckAndNotifyExpiringItems`:

1. Получить все `fresh` items с `expires_at`
2. Для каждого:
   - Вычислить `daysLeft`
   - Если `expired` → обновить статус
   - Если нужно уведомление → проверить дубли → создать
3. Логировать результаты

---

## 🗑️ Что делать с просроченными продуктами

### Автоматически:
- `status` меняется на `expired`
- Создается critical уведомление

### Пользователь может:
1. **Подтвердить утилизацию** → `status = discarded`
2. **Исправить дату** → если ошибся при вводе
3. **Игнорировать** → продукт остается `expired`

### ❌ НЕ удалять физически из БД
Причины:
- История потерь
- Аналитика
- Статистика трат

---

## 📊 Аналитика потерь (будущее)

### Возможные метрики:

```sql
-- Общие потери за месяц
SELECT SUM(price_total) 
FROM fridge_items 
WHERE status = 'discarded' 
  AND updated_at >= '2026-01-01';

-- Топ-5 самых часто выбрасываемых продуктов
SELECT ingredient_id, COUNT(*) 
FROM fridge_items 
WHERE status = 'discarded'
GROUP BY ingredient_id 
ORDER BY COUNT(*) DESC 
LIMIT 5;
```

---

## ✅ Чек-лист реализации

- [x] Модель `FridgeItem` с полями `status`, `daysLeft`, `priceTotal`
- [x] Модель `Notification` с типами `fridge`, `ai`
- [x] Функция `EvaluateFridgeItem` (вычисление статуса)
- [x] Функция `CreateExpiryNotification` (создание уведомлений)
- [x] Функция `CheckAndNotifyExpiringItems` (проверка всех продуктов)
- [x] Проверка дубликатов уведомлений (по дате)
- [ ] CRON задача (ежедневная проверка)
- [ ] AI-генерация текстов уведомлений
- [ ] Endpoint GET `/api/fridge/items` с авто-проверкой
- [ ] Endpoint GET `/api/notifications`
- [ ] Endpoint PATCH `/api/fridge/items/:id` (изменить дату/количество)
- [ ] Endpoint POST `/api/fridge/items/:id/discard` (выбросить)

---

## 🎯 Итог

**Это НЕ просто холодильник.**  
**Это система предотвращения потерь.**

Backend знает:
- Что у тебя есть
- Когда это испортится
- Сколько ты потеряешь

AI помогает:
- Напомнить вовремя
- Мотивировать готовить
- Предложить рецепты

**Результат**: меньше выброшенной еды, больше сэкономленных денег.
