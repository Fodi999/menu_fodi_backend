-- =====================================================
-- МИГРАЦИЯ 028: Исправление ролей пользователей
-- =====================================================
-- Обновляем всех пользователей с role='user' на 'home_chef'
-- =====================================================

-- 1. Добавляем значения в enum если их нет (безопасно)
DO $$
BEGIN
    -- Проверяем и добавляем home_chef
    IF NOT EXISTS (SELECT 1 FROM pg_enum WHERE enumlabel = 'home_chef' AND enumtypid = 'Role'::regtype) THEN
        ALTER TYPE "Role" ADD VALUE 'home_chef';
    END IF;
    
    -- Проверяем и добавляем pro_chef
    IF NOT EXISTS (SELECT 1 FROM pg_enum WHERE enumlabel = 'pro_chef' AND enumtypid = 'Role'::regtype) THEN
        ALTER TYPE "Role" ADD VALUE 'pro_chef';
    END IF;
END $$;

-- 2. Обновляем всех пользователей с некорректной ролью 'user' на 'home_chef'
UPDATE "User"
SET role = 'home_chef'
WHERE role = 'user';

-- 3. Проверка результата
SELECT 
    role,
    COUNT(*) as user_count
FROM "User"
GROUP BY role
ORDER BY role;

-- ✅ Ожидаемый результат:
-- role       | user_count
-- -----------|------------
-- admin      | X
-- home_chef  | Y (все бывшие 'user' + новые)
-- pro_chef   | Z

COMMENT ON TABLE "User" IS 'Users table - default role is home_chef for MVP';
