-- ============================================================================
-- NOTIFICATIONS V2 - ПРАВИЛЬНАЯ АРХИТЕКТУРА
-- ============================================================================
-- Цель: чистая система уведомлений только для expiry tracking
-- Исключить: добавлен/удалён продукт (это activity logs, не notifications)
--
-- Date: 2026-01-21
-- ============================================================================

-- 1️⃣ Добавляем УНИКАЛЬНОСТЬ чтобы не было дублей
-- Одно и то же событие не может быть создано дважды в один день
ALTER TABLE notifications 
  ADD COLUMN IF NOT EXISTS unique_key VARCHAR(255);

-- Индекс для быстрого поиска по unique_key
CREATE INDEX IF NOT EXISTS idx_notifications_unique_key 
  ON notifications(unique_key) 
  WHERE status = 'active' AND unique_key IS NOT NULL;

-- 2️⃣ Добавляем статус для отслеживания актуальности
ALTER TABLE notifications 
  ADD COLUMN IF NOT EXISTS status VARCHAR(20) DEFAULT 'active' 
  CHECK (status IN ('active', 'resolved', 'expired'));

-- Индекс для быстрого поиска активных уведомлений
CREATE INDEX IF NOT EXISTS idx_notifications_active 
  ON notifications(user_id, status, created_at DESC) 
  WHERE status = 'active' AND read_at IS NULL;

-- 3️⃣ Улучшаем индекс для группировки по level
CREATE INDEX IF NOT EXISTS idx_notifications_by_level 
  ON notifications(user_id, level, status, created_at DESC) 
  WHERE status = 'active';

-- ============================================================================
-- БИЗНЕС-ПРАВИЛА (документация)
-- ============================================================================
-- 
-- ✅ Создавать уведомления ТОЛЬКО для:
--    - daysLeft <= 0 → CRITICAL (срочно использовать)
--    - daysLeft = 1  → WARNING (скоро испортится)
--    - daysLeft = 2  → INFO (только summary)
--
-- ❌ НЕ создавать для:
--    - Добавлен продукт (это activity log)
--    - Удалён продукт (это activity log)
--    - Изменена цена (это history event)
--    - daysLeft >= 3 (слишком рано)
--
-- 🔒 Уникальность:
--    unique_key = user_id + type + level + date + fridge_item_id
--    Гарантирует: одно уведомление на продукт в день
--
-- 📦 Meta структура для fridge notifications:
--    {
--      "fridgeItemId": "uuid",
--      "ingredientId": "uuid",
--      "ingredientName": "Łosoś",
--      "daysLeft": -1,
--      "expiresAt": "2026-01-20T00:00:00Z",
--      "quantity": 500,
--      "unit": "g"
--    }
--
-- ============================================================================

COMMENT ON COLUMN notifications.unique_key IS 'Hash для предотвращения дублей';
COMMENT ON COLUMN notifications.status IS 'active=текущее, resolved=решено, expired=устарело';

