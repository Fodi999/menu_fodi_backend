-- Migration: Create Tasks tables
-- Description: Creates tables for task system (tasks, user_tasks, task_categories)
-- Date: 2025-12-11

-- ============================================
-- 1. Create task_categories table
-- ============================================
CREATE TABLE IF NOT EXISTS task_categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    icon VARCHAR(100),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- ============================================
-- 2. Create tasks table
-- ============================================
CREATE TABLE IF NOT EXISTS tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(255) NOT NULL,
    description TEXT,
    reward BIGINT NOT NULL DEFAULT 0,
    is_active BOOLEAN DEFAULT true,
    category_id UUID,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Foreign key to task_categories
    CONSTRAINT fk_task_category 
        FOREIGN KEY (category_id) 
        REFERENCES task_categories(id) 
        ON DELETE SET NULL
);

-- ============================================
-- 3. Create user_tasks table (user progress)
-- ============================================
CREATE TABLE IF NOT EXISTS user_tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id TEXT NOT NULL,
    task_id UUID NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    progress INT DEFAULT 0,
    completed_at TIMESTAMP,
    reward_claimed BOOLEAN DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Foreign keys
    CONSTRAINT fk_user_task_user 
        FOREIGN KEY (user_id) 
        REFERENCES "User"(id) 
        ON DELETE CASCADE,
    
    CONSTRAINT fk_user_task_task 
        FOREIGN KEY (task_id) 
        REFERENCES tasks(id) 
        ON DELETE CASCADE,
    
    -- Unique constraint: один пользователь может иметь только одну запись для одного задания
    CONSTRAINT unique_user_task UNIQUE (user_id, task_id),
    
    -- Check constraint: статус должен быть валидным
    CONSTRAINT check_status 
        CHECK (status IN ('pending', 'in_progress', 'completed', 'failed')),
    
    -- Check constraint: прогресс от 0 до 100
    CONSTRAINT check_progress 
        CHECK (progress >= 0 AND progress <= 100)
);

-- ============================================
-- 4. Create indexes for better performance
-- ============================================

-- Index for user_tasks by user_id (frequently queried)
CREATE INDEX IF NOT EXISTS idx_user_tasks_user_id ON user_tasks(user_id);

-- Index for user_tasks by task_id
CREATE INDEX IF NOT EXISTS idx_user_tasks_task_id ON user_tasks(task_id);

-- Index for user_tasks by status (filter active tasks)
CREATE INDEX IF NOT EXISTS idx_user_tasks_status ON user_tasks(status);

-- Index for tasks by is_active (filter active tasks)
CREATE INDEX IF NOT EXISTS idx_tasks_is_active ON tasks(is_active);

-- Index for tasks by category_id
CREATE INDEX IF NOT EXISTS idx_tasks_category_id ON tasks(category_id);

-- Composite index for user_id + status (common query pattern)
CREATE INDEX IF NOT EXISTS idx_user_tasks_user_status ON user_tasks(user_id, status);

-- ============================================
-- 5. Seed initial task categories
-- ============================================
INSERT INTO task_categories (name, description, icon) VALUES
    ('daily', 'Ежедневные задания', '📅'),
    ('cooking', 'Кулинарные задания', '👨‍🍳'),
    ('social', 'Социальные задания', '👥'),
    ('learning', 'Обучающие задания', '📚'),
    ('achievement', 'Достижения', '🏆'),
    ('special', 'Специальные задания', '⭐')
ON CONFLICT (name) DO NOTHING;

-- ============================================
-- 6. Seed initial tasks
-- ============================================
INSERT INTO tasks (title, description, reward, is_active, category_id) VALUES
    (
        'Первый рецепт',
        'Создайте свой первый рецепт в приложении',
        50,
        true,
        (SELECT id FROM task_categories WHERE name = 'cooking' LIMIT 1)
    ),
    (
        'Ежедневный вход',
        'Заходите в приложение каждый день',
        10,
        true,
        (SELECT id FROM task_categories WHERE name = 'daily' LIMIT 1)
    ),
    (
        'Поделиться рецептом',
        'Поделитесь вашим рецептом с друзьями',
        25,
        true,
        (SELECT id FROM task_categories WHERE name = 'social' LIMIT 1)
    ),
    (
        'Изучить основы',
        'Пройдите вводный курс по кулинарии',
        100,
        true,
        (SELECT id FROM task_categories WHERE name = 'learning' LIMIT 1)
    ),
    (
        'Мастер-шеф',
        'Создайте 10 рецептов',
        200,
        true,
        (SELECT id FROM task_categories WHERE name = 'achievement' LIMIT 1)
    ),
    (
        'Добавить 5 ингредиентов в холодильник',
        'Заполните свой виртуальный холодильник',
        30,
        true,
        (SELECT id FROM task_categories WHERE name = 'daily' LIMIT 1)
    ),
    (
        'Приготовить рецепт дня',
        'Приготовьте специальный рецепт дня',
        75,
        true,
        (SELECT id FROM task_categories WHERE name = 'special' LIMIT 1)
    ),
    (
        'Пригласить друга',
        'Пригласите друга в приложение',
        150,
        true,
        (SELECT id FROM task_categories WHERE name = 'social' LIMIT 1)
    )
ON CONFLICT DO NOTHING;

-- ============================================
-- 7. Create trigger for updated_at auto-update
-- ============================================

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger for tasks table
DROP TRIGGER IF EXISTS update_tasks_updated_at ON tasks;
CREATE TRIGGER update_tasks_updated_at
    BEFORE UPDATE ON tasks
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Trigger for user_tasks table
DROP TRIGGER IF EXISTS update_user_tasks_updated_at ON user_tasks;
CREATE TRIGGER update_user_tasks_updated_at
    BEFORE UPDATE ON user_tasks
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- ============================================
-- Verification queries
-- ============================================

-- Verify task_categories
-- SELECT * FROM task_categories;

-- Verify tasks
-- SELECT t.id, t.title, t.reward, tc.name as category FROM tasks t
-- LEFT JOIN task_categories tc ON t.category_id = tc.id;

-- Verify user_tasks structure
-- SELECT column_name, data_type, is_nullable, column_default 
-- FROM information_schema.columns 
-- WHERE table_name = 'user_tasks';
