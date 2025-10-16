-- 🧱 1. Таблица бизнесов
CREATE TABLE IF NOT EXISTS "Business" (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  owner_id TEXT REFERENCES "User"(id) ON DELETE CASCADE,
  name TEXT NOT NULL,
  description TEXT,
  category TEXT,
  city TEXT,
  is_active BOOLEAN DEFAULT true,
  created_at TIMESTAMP DEFAULT now(),
  updated_at TIMESTAMP DEFAULT now()
);

-- Индексы для быстрого поиска
CREATE INDEX IF NOT EXISTS idx_business_owner ON "Business"(owner_id);
CREATE INDEX IF NOT EXISTS idx_business_category ON "Business"(category);
CREATE INDEX IF NOT EXISTS idx_business_city ON "Business"(city);
CREATE INDEX IF NOT EXISTS idx_business_active ON "Business"(is_active);

-- 💰 2. Таблица токенов
CREATE TABLE IF NOT EXISTS "BusinessToken" (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  business_id UUID REFERENCES "Business"(id) ON DELETE CASCADE,
  symbol TEXT NOT NULL,
  total_supply BIGINT DEFAULT 1,
  price NUMERIC(10,2) DEFAULT 19.00,
  created_at TIMESTAMP DEFAULT now()
);

-- Индексы
CREATE INDEX IF NOT EXISTS idx_token_business ON "BusinessToken"(business_id);
CREATE INDEX IF NOT EXISTS idx_token_symbol ON "BusinessToken"(symbol);

-- 👥 3. Подписки / доли
CREATE TABLE IF NOT EXISTS "BusinessSubscription" (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id TEXT REFERENCES "User"(id) ON DELETE CASCADE,
  business_id UUID REFERENCES "Business"(id) ON DELETE CASCADE,
  tokens_owned BIGINT DEFAULT 1,
  invested NUMERIC(10,2) DEFAULT 19.00,
  created_at TIMESTAMP DEFAULT now()
);

-- Индексы
CREATE INDEX IF NOT EXISTS idx_subscription_user ON "BusinessSubscription"(user_id);
CREATE INDEX IF NOT EXISTS idx_subscription_business ON "BusinessSubscription"(business_id);
-- Уникальность: один пользователь может иметь только одну подписку на бизнес
CREATE UNIQUE INDEX IF NOT EXISTS idx_subscription_unique ON "BusinessSubscription"(user_id, business_id);

-- 🔄 4. Транзакции по токенам
CREATE TABLE IF NOT EXISTS "Transaction" (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  business_id UUID REFERENCES "Business"(id) ON DELETE CASCADE,
  from_user TEXT,
  to_user TEXT,
  tokens BIGINT,
  amount NUMERIC(10,2),
  tx_type TEXT, -- "buy", "sell", "burn", "transfer"
  created_at TIMESTAMP DEFAULT now()
);

-- Индексы
CREATE INDEX IF NOT EXISTS idx_transaction_business ON "Transaction"(business_id);
CREATE INDEX IF NOT EXISTS idx_transaction_from ON "Transaction"(from_user);
CREATE INDEX IF NOT EXISTS idx_transaction_to ON "Transaction"(to_user);
CREATE INDEX IF NOT EXISTS idx_transaction_type ON "Transaction"(tx_type);
CREATE INDEX IF NOT EXISTS idx_transaction_created ON "Transaction"(created_at DESC);
