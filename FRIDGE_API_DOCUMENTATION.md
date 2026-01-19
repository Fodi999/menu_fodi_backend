# 🧊 Fridge & Notifications API Documentation

## 📡 Fridge Endpoints

### GET /api/fridge/items
Получить все продукты пользователя в холодильнике.

**Auth**: Required (JWT)

**Response**:
```json
{
  "data": [
    {
      "id": "uuid",
      "userId": "uuid",
      "ingredientId": "uuid",
      "quantity": 1200,
      "unit": "g",
      "expiresAt": "2026-01-16T00:00:00Z",
      "status": "fresh",
      "daysLeft": 1,
      "priceTotal": 14.76,
      "createdAt": "2026-01-15T08:00:00Z",
      "ingredient": {
        "id": "uuid",
        "name": "Карась",
        "namePl": "Karaś",
        "category": "fish"
      }
    }
  ]
}
```

**Backend автоматически**:
- Вычисляет `daysLeft` и `status`
- Обновляет expired items
- Создает уведомления (если нужно)

---

### POST /api/fridge/items
Добавить продукт в холодильник.

**Auth**: Required (JWT)

**Request**:
```json
{
  "ingredientId": "uuid",
  "quantity": 1200,
  "unit": "g",
  "expiresAt": "2026-01-16T00:00:00Z",
  "priceTotal": 14.76
}
```

**Response**:
```json
{
  "data": {
    "id": "uuid",
    "status": "fresh",
    "daysLeft": 1,
    ...
  }
}
```

---

### PATCH /api/fridge/items/:id
Обновить продукт (количество, дату истечения, цену).

**Auth**: Required (JWT)

**Request**:
```json
{
  "quantity": 800,
  "expiresAt": "2026-01-17T00:00:00Z",
  "priceTotal": 10.00
}
```

**Response**:
```json
{
  "data": {
    "id": "uuid",
    "quantity": 800,
    "status": "fresh",
    "daysLeft": 2,
    ...
  }
}
```

---

### DELETE /api/fridge/items/:id
Удалить продукт из холодильника (физическое удаление).

**Auth**: Required (JWT)

**Response**:
```json
{
  "message": "Item deleted successfully"
}
```

---

### POST /api/fridge/items/:id/discard
Выбросить продукт (мягкое удаление, меняет status на `discarded`).

**Auth**: Required (JWT)

**Response**:
```json
{
  "message": "Item discarded successfully"
}
```

**Используется для**:
- Трекинга потерь
- Аналитики выброшенной еды
- Статистики трат

---

## 🔔 Notifications Endpoints

### GET /api/notifications
Получить уведомления пользователя.

**Auth**: Required (JWT)

**Query params**:
- `unreadOnly` (boolean) - только непрочитанные

**Response**:
```json
{
  "data": [
    {
      "id": "uuid",
      "userId": "uuid",
      "type": "ai",
      "level": "warning",
      "title": "Используй карася сегодня",
      "message": "AI рекомендует приготовить карася, иначе ты потеряешь 14.76 PLN",
      "meta": "{\"fridgeItemId\":\"uuid\",\"daysLeft\":1}",
      "readAt": null,
      "createdAt": "2026-01-15T08:00:00Z"
    },
    {
      "id": "uuid",
      "type": "fridge",
      "level": "critical",
      "title": "Продукт просрочен",
      "message": "Молоко просрочено. Потеря: 5.40 PLN",
      "meta": "{\"fridgeItemId\":\"uuid\",\"daysLeft\":-2}",
      "readAt": "2026-01-15T09:00:00Z",
      "createdAt": "2026-01-14T08:00:00Z"
    }
  ]
}
```

**Типы уведомлений**:
- `system` - системные
- `order` - заказы
- `user` - от других пользователей
- `fridge` - холодильник (системные)
- `ai` - AI рекомендации
- `backup` - бекапы

**Уровни**:
- `info` - информация (2-3 дня)
- `warning` - предупреждение (1 день)
- `critical` - критично (сегодня или просрочено)

---

### PATCH /api/notifications/:id/read
Пометить уведомление как прочитанное.

**Auth**: Required (JWT)

**Response**:
```json
{
  "message": "Notification marked as read"
}
```

---

### POST /api/notifications/read-all
Пометить все уведомления как прочитанные.

**Auth**: Required (JWT)

**Response**:
```json
{
  "message": "All notifications marked as read"
}
```

---

### GET /api/notifications/unread-count
Получить количество непрочитанных уведомлений.

**Auth**: Required (JWT)

**Response**:
```json
{
  "count": 5
}
```

---

## 🕐 CRON задача

### Ежедневная проверка
**Время**: каждый день в 08:00 UTC (09:00 Варшава, 11:00 Москва)

**Что делает**:
1. Находит всех пользователей с продуктами в холодильнике
2. Для каждого пользователя:
   - Вычисляет `daysLeft` для всех fresh items
   - Обновляет status expired items
   - Создает уведомления (с защитой от дублей)
3. Логирует результаты

**Можно запустить вручную** (для теста):
```go
cronChecker.RunNow()
```

---

## 🤖 AI генерация текстов

### Персонализированные сообщения

AI генерирует тексты в зависимости от срочности:

**Просрочено** (< 0 дней):
```
"Niestety, karaś się zepsuł. Strata: 14.76 PLN 😞"
```

**Истекает сегодня** (0 дней):
```
"Ugotuj karasia dzisiaj! Nie trać 14.76 PLN 🐟"
```

**Завтра** (1 день):
```
"Masz karasia (1200g). Użyj jutro, bo stracisz 14.76 PLN"
```

**Скоро** (2-3 дня):
```
"Karaś wygasa za 2 dni. Nie zapomnij użyć!"
```

### Преимущества AI:
- ✅ Персонализация под продукт
- ✅ Естественный язык
- ✅ Мотивирующий тон
- ✅ Учет ценности продукта

---

## 📊 Примеры использования

### Frontend: получить продукты и показать уведомления

```typescript
// 1. Получить продукты холодильника
const fridgeItems = await fetch('/api/fridge/items', {
  headers: { Authorization: `Bearer ${token}` }
}).then(r => r.json())

// Backend автоматически вычисляет daysLeft и создает уведомления

// 2. Получить уведомления
const notifications = await fetch('/api/notifications?unreadOnly=true', {
  headers: { Authorization: `Bearer ${token}` }
}).then(r => r.json())

// 3. Показать badge с количеством
const { count } = await fetch('/api/notifications/unread-count', {
  headers: { Authorization: `Bearer ${token}` }
}).then(r => r.json())

console.log(`У вас ${count} непрочитанных уведомлений`)
```

### Frontend: выбросить продукт

```typescript
await fetch(`/api/fridge/items/${itemId}/discard`, {
  method: 'POST',
  headers: { Authorization: `Bearer ${token}` }
})

// Backend изменит status на 'discarded' (для статистики)
```

---

## ✅ Проверка работы

### Тест 1: Добавить продукт с истечением завтра
```bash
curl -X POST http://localhost:8080/api/fridge/items \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "ingredientId": "uuid",
    "quantity": 1000,
    "unit": "g",
    "expiresAt": "2026-01-16T00:00:00Z",
    "priceTotal": 12.50
  }'
```

### Тест 2: Получить уведомления
```bash
curl http://localhost:8080/api/notifications \
  -H "Authorization: Bearer $TOKEN"
```

### Тест 3: Запустить CRON вручную
```bash
# В коде сервера
cronChecker.RunNow()
```

---

## 🎯 Итог

**Это полноценная система предотвращения потерь:**

1. ✅ Backend сам вычисляет срок годности
2. ✅ Автоматические уведомления с AI-текстами
3. ✅ Единая система уведомлений (fridge + orders + system)
4. ✅ Ежедневная CRON проверка
5. ✅ Трекинг потерь через `discarded` статус
6. ✅ Персонализированные AI-сообщения

**Результат**: меньше выброшенной еды = больше сэкономленных денег 💰
