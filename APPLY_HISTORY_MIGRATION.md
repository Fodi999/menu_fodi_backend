# 🔧 Применение миграции history_events на Koyeb

## ⚠️ КРИТИЧНО: Используй ЭТОТ SQL, не упрощённый!

Код использует **PostgreSQL ENUM типы**, которые должны точно совпадать с GORM моделью.

## 📋 SQL для выполнения в Koyeb PostgreSQL Console

```sql
-- 1. Создать ENUM типы
CREATE TYPE history_event_type AS ENUM (
    'cook', 
    'consume', 
    'waste',       -- 👈 Используется для expired items
    'manual', 
    'fridge_add', 
    'fridge_remove'
);

CREATE TYPE history_source_type AS ENUM (
    'prepared_dish', 
    'recipe', 
    'fridge',      -- 👈 Используется для auto actions
    'manual'
);

-- 2. Создать таблицу
CREATE TABLE IF NOT EXISTS history_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id TEXT NOT NULL REFERENCES "User"(id) ON DELETE CASCADE,
    event_type history_event_type NOT NULL,
    source_type history_source_type NOT NULL,
    source_id TEXT,
    
    -- Event details
    portions INT,
    metadata JSONB,
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 3. Создать индексы для производительности
CREATE INDEX IF NOT EXISTS idx_history_events_user_created 
    ON history_events(user_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_history_events_type 
    ON history_events(user_id, event_type);

CREATE INDEX IF NOT EXISTS idx_history_events_source 
    ON history_events(source_type, source_id);

CREATE INDEX IF NOT EXISTS idx_history_events_metadata 
    ON history_events USING GIN (metadata);
```

## 🎯 Как применить

### Вариант 1: Через Koyeb Dashboard
1. Открой: https://app.koyeb.com
2. Выбери свой сервис → Database
3. Найди PostgreSQL instance
4. Открой SQL Console / Query Editor
5. Скопируй весь SQL выше
6. Execute

### Вариант 2: Через psql CLI
```bash
# Получи DATABASE_URL из Koyeb Environment Variables
psql "postgresql://user:pass@host:5432/dbname" -f migrations/049_create_history_events.sql
```

## ✅ Проверка после применения

```sql
-- Должны существовать:
SELECT typname FROM pg_type WHERE typname IN ('history_event_type', 'history_source_type');

-- Должна вернуть 0 строк (пока нет данных):
SELECT COUNT(*) FROM history_events;
```

## 🧪 Проверка API после миграции

```bash
# Должен вернуть 401 (требуется auth):
curl https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/history/losses?days=30

# С токеном должен вернуть 200 с пустыми данными:
curl -H "Authorization: Bearer YOUR_TOKEN" \
  https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/history/losses?days=30
```

Ожидаемый ответ:
```json
{
  "success": true,
  "data": {
    "events": [],
    "summary": {
      "totalItems": 0,
      "totalCost": 0,
      "currency": "PLN"
    }
  }
}
```

## 🚨 Важно для будущего

Эта миграция должна была применяться автоматически при деплое. 

**TODO:** Настроить автоматические миграции через:
- Goose в Dockerfile
- Или отдельный init container
- Или CI/CD pipeline step

**Без этого каждый новый домен (tokens, ai_logs, etc) будет требовать ручного вмешательства.**

## 📚 Source

Миграция из: `migrations/049_create_history_events.sql`
Модель GORM: `internal/models/history_event.go`
