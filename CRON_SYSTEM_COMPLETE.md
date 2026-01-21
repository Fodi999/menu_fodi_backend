# ⏰ CRON SYSTEM - AUTOMATIC EXPIRY NOTIFICATIONS

## 🎯 STATUS: ✅ PRODUCTION READY

Backend автоматически проверяет истекающие продукты и создаёт уведомления **каждый день в 08:00**.

---

## 📋 ЧТО СДЕЛАНО

### 1. ✅ CRON Scheduler (Internal Go-based)

**Файл:** `internal/cron/fridge_expiry_checker.go`

**Технология:** `github.com/robfig/cron/v3`

**Особенности:**
- ⏰ Запускается каждый день в **08:00 Europe/Warsaw**
- 🌍 Правильный timezone (не UTC!)
- 🔄 Graceful shutdown (останавливается при SIGTERM)
- 🧪 Поддержка ручного запуска для тестирования

**Архитектура:**
```go
// Инициализация в internal/app/server.go
cronChecker := cron.NewFridgeExpiryChecker(database.DB)
cronChecker.Start()
logger.Info("⏰ CRON jobs initialized - Daily fridge expiry checks at 08:00 Europe/Warsaw")

// Graceful shutdown
defer cronChecker.Stop()
```

---

## 🔄 КАК ЭТО РАБОТАЕТ

### Daily Flow (08:00):

```
08:00 Europe/Warsaw
    ↓
CRON triggers checkAllUsers()
    ↓
Query: SELECT DISTINCT user_id FROM user_fridge_items WHERE expires_at IS NOT NULL
    ↓
For each user:
    CheckAndNotifyExpiringItems(db, userID)
        ↓
        Group items by severity:
        - daysLeft ≤ 0 → CRITICAL ⛔
        - daysLeft = 1 → WARNING ⚠️
        - daysLeft = 2 → INFO ℹ️
        ↓
        Create notifications via NotificationService
        (auto-deduplication via unique_key)
        ↓
        User sees notifications in app
```

---

## 🧪 ТЕСТИРОВАНИЕ

### Вариант 1: Ручной запуск (не ждать 08:00)

```bash
# Компилируем утилиту
go build -o bin/test_cron ./cmd/test_cron

# Запускаем проверку СЕЙЧАС
./bin/test_cron
```

**Или используй готовый скрипт:**
```bash
./test_cron_now.sh
```

**Что происходит:**
1. Подключается к БД
2. Находит всех пользователей с продуктами
3. Запускает CheckAndNotifyExpiringItems() для каждого
4. Показывает статистику (success, errors)

**Проверка результатов:**
```bash
# 1. Посмотреть уведомления пользователя
curl -H "Authorization: Bearer $TOKEN" \
  https://your-api.com/api/notifications

# 2. Проверить badge count
curl -H "Authorization: Bearer $TOKEN" \
  https://your-api.com/api/notifications/unread-count
```

### Вариант 2: Проверка автоматического запуска

```bash
# Запускаем сервер
./bin/server

# Логи покажут:
# 🕐 CRON: Fridge expiry checker started (daily at 08:00 Europe/Warsaw)

# Ждём 08:00 или проверяем логи на следующий день:
# 🔍 [2026-01-22 08:00:00] Starting daily fridge expiry check...
# 📊 Found 15 users with fridge items
# ✅ Daily check completed in 1.2s: 15 users processed, 0 errors
```

---

## 📊 PRODUCTION CHECKLIST

### ✅ Реализовано:

- [x] CRON scheduler инициализируется при старте сервера
- [x] Правильный timezone (Europe/Warsaw, не UTC)
- [x] Graceful shutdown (останавливается при SIGTERM/SIGINT)
- [x] Обработка всех пользователей с продуктами
- [x] Использует новую архитектуру (NotificationService)
- [x] Автоматическая деduplication (unique_key)
- [x] Логирование статистики (processed users, errors, duration)
- [x] Тестовая утилита для ручного запуска

### 📋 Рекомендуется добавить (опционально):

- [ ] Метрики в Prometheus
  ```go
  // internal/platform/metrics/cron.go
  cronJobDuration := prometheus.NewHistogram(...)
  cronNotificationsCreated := prometheus.NewCounterVec(...)
  ```

- [ ] Алерты в Slack/Email при ошибках
  ```go
  if errorCount > 0 {
      notifyOps("CRON failed for %d users", errorCount)
  }
  ```

- [ ] Dashboard (Grafana)
  - Количество обработанных пользователей
  - Количество созданных уведомлений (critical/warning/info)
  - Performance (duration, errors)

---

## 🚀 DEPLOYMENT

### Koyeb / Cloud Platform (автоматический запуск)

**Ничего делать не нужно!** 🎉

CRON запускается автоматически при старте контейнера:

```dockerfile
# Dockerfile
FROM golang:1.21 AS builder
WORKDIR /app
COPY . .
RUN go build -o server ./cmd/server

FROM alpine:latest
COPY --from=builder /app/server /server
CMD ["/server"]
# ✅ CRON стартует внутри процесса, не нужен отдельный контейнер
```

**При деплое:**
1. Контейнер стартует → `server` запускается
2. `internal/app/server.go` инициализирует CRON
3. Scheduler добавляет задачу "0 8 * * *" (каждый день в 08:00)
4. CRON работает в фоне, не блокирует HTTP server

**Graceful restart:**
- Platform отправляет SIGTERM
- Server останавливает CRON (`cronChecker.Stop()`)
- Завершает текущие запросы
- Shutdown завершается корректно

---

## 🐛 TROUBLESHOOTING

### Проблема: Уведомления не создаются

**Диагностика:**
```bash
# 1. Проверь есть ли продукты с expires_at
psql $DATABASE_URL -c "
SELECT user_id, COUNT(*) 
FROM user_fridge_items 
WHERE expires_at IS NOT NULL 
GROUP BY user_id;
"

# 2. Запусти ручной тест
./bin/test_cron

# 3. Проверь логи сервера
# Должна быть строка:
# 🕐 CRON: Fridge expiry checker started (daily at 08:00 Europe/Warsaw)
```

### Проблема: Неправильное время запуска

**Проверка timezone:**
```bash
# В логах должно быть:
# 🕐 CRON: Fridge expiry checker started (daily at 08:00 Europe/Warsaw)

# Если видишь UTC - проблема с загрузкой timezone
# Решение: установи tzdata в контейнер
```

```dockerfile
# Dockerfile - добавить tzdata
FROM alpine:latest
RUN apk add --no-cache tzdata
ENV TZ=Europe/Warsaw
```

### Проблема: CRON останавливается при ошибке

**Защита добавлена:**
```go
// internal/cron/fridge_expiry_checker.go
for _, userID := range userIDs {
    if err := CheckAndNotifyExpiringItems(db, userID); err != nil {
        fmt.Printf("❌ Failed to check user %s: %v\n", userID, err)
        errorCount++
        // ✅ Продолжаем обработку остальных пользователей!
    } else {
        successCount++
    }
}
```

---

## 📈 NEXT LEVEL ENHANCEMENTS

### 1. Dynamic Schedule (per user timezone)

```go
// Учитывать timezone пользователя
// Пример: польский пользователь → 08:00 Warsaw
//         русский пользователь → 08:00 Moscow
```

### 2. Smart Timing (ML-based)

```go
// Анализировать когда пользователь обычно открывает приложение
// Отправлять уведомления перед активностью
```

### 3. Batch Optimization

```go
// Вместо N запросов для N пользователей
// Сделать 1 массовый запрос:
items := db.Find("expires_at <= NOW() + INTERVAL '2 days'")
// Group by user, create notifications in batch
```

---

## ✅ РЕЗУЛЬТАТ

### Backend Status: 🎉 100% COMPLETE

**Что работает автоматически:**
1. ✅ Проверка истекающих продуктов каждый день в 08:00
2. ✅ Создание уведомлений с правильной группировкой (critical/warning/info)
3. ✅ Деduplication (не создаём дубли)
4. ✅ Graceful shutdown (корректная остановка при деплое)
5. ✅ Тестирование без ожидания 08:00

**Архитектурные принципы соблюдены:**
- Notifications = только для expiry (не для add/delete)
- GET endpoints не мутируют состояние
- Badge count = critical + warning (INFO excluded)
- Unique_key предотвращает спам
- CRON работает в отдельной горутине (не блокирует API)

---

## 📚 RELATED DOCS

- [NOTIFICATIONS_ARCHITECTURE.md](./NOTIFICATIONS_ARCHITECTURE.md) - полная архитектура
- [NOTIFICATIONS_QUICK_START.md](./NOTIFICATIONS_QUICK_START.md) - quick reference
- [test_cron_now.sh](./test_cron_now.sh) - скрипт для тестирования

---

**Last Updated:** 21 января 2026
**Status:** Production Ready ✅
**Next Step:** Frontend integration (см. NOTIFICATIONS_QUICK_START.md)
