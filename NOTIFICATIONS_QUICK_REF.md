# 📢 Уведомления - Краткая инструкция

## 🎯 Как работает

### 1. CRON (автоматически каждый день в 08:00 UTC)

```
08:00 UTC - CRON запускается
↓
Проверяет все продукты у всех пользователей
↓
Находит продукты с истекающим сроком
↓
Создаёт уведомления в БД
```

### 2. Типы уведомлений

| Days Left | Level | Пример |
|-----------|-------|--------|
| < 0 | 🔴 CRITICAL | "Mleko просрочен. Потеря: 6.50 PLN" |
| = 0 | 🔴 CRITICAL | "Jajka нужно использовать сегодня!" |
| = 1 | 🟡 WARNING | "Pomidor истекает завтра - успей приготовить!" |
| 2-3 | 🔵 INFO | "Ogórek истечет через 2 дня" |

### 3. AI подсказки (Groq API)

**Для истекающих продуктов:**
```
"Zrób szybką sałatkę caprese lub pyszny sos pomidorowy! 🍝"
```

**Для просроченных:**
```
"Mleko przeterminowało się 24 dni temu. Strata: 6.50 PLN. 
Sprawdź lodówkę częściej! 🥛"
```

## 🔔 API

```bash
# Количество непрочитанных
GET /api/notifications/unread-count
→ {"count": 7}

# Все уведомления
GET /api/notifications
→ {"data": [...]}

# Отметить прочитанным
PATCH /api/notifications/{id}/read

# Отметить все
POST /api/notifications/read-all
```

## ⏰ Следующий запуск

**Завтра:** 2026-01-16 08:00 UTC  
**Ожидается:** 7 уведомлений о просроченных продуктах  
**Потери:** ~80 PLN  

---

**Полная документация:** `NOTIFICATION_SYSTEM_GUIDE.md`
