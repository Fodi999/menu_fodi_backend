-- +goose Up
CREATE TABLE IF NOT EXISTS weekly_budgets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id TEXT NOT NULL REFERENCES "User"(id) ON DELETE CASCADE,
    week_start DATE NOT NULL,  -- Monday of the week (ISO week standard)
    
    -- Budget tracking
    planned_budget DECIMAL(10, 2) NOT NULL DEFAULT 0.00,
    spent_budget DECIMAL(10, 2) NOT NULL DEFAULT 0.00,
    waste_cost DECIMAL(10, 2) NOT NULL DEFAULT 0.00,
    
    -- Metadata
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    -- Ensure one budget per user per week
    CONSTRAINT unique_user_week UNIQUE (user_id, week_start),
    
    -- Budget constraints
    CONSTRAINT positive_planned_budget CHECK (planned_budget >= 0),
    CONSTRAINT positive_spent_budget CHECK (spent_budget >= 0),
    CONSTRAINT positive_waste_cost CHECK (waste_cost >= 0)
);

-- Index for fetching user's budgets sorted by week
CREATE INDEX IF NOT EXISTS idx_weekly_budgets_user_week ON weekly_budgets(user_id, week_start DESC);

-- Index for current week queries
CREATE INDEX IF NOT EXISTS idx_weekly_budgets_current_week ON weekly_budgets(user_id, week_start) WHERE week_start >= CURRENT_DATE - INTERVAL '7 days';

COMMENT ON TABLE weekly_budgets IS 'Weekly food budget tracking for users';
COMMENT ON COLUMN weekly_budgets.week_start IS 'Monday of the week (ISO 8601)';
COMMENT ON COLUMN weekly_budgets.planned_budget IS 'User-set weekly food budget limit';
COMMENT ON COLUMN weekly_budgets.spent_budget IS 'Actual money spent on consumed food (from consume events)';
COMMENT ON COLUMN weekly_budgets.waste_cost IS 'Money wasted on expired/discarded food';

-- +goose Down
DROP TABLE IF EXISTS weekly_budgets;
