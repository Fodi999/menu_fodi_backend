# User Status and Activity Tracking

## 📊 Три статуса пользователей (не путать!)

### 1️⃣ Account Status (Статус аккаунта) - БАЗА

**Назначение**: Административный статус, управляется админом

**Поле в БД**: `status VARCHAR(20) NOT NULL DEFAULT 'active'`

**Возможные значения**:
- `active` - нормальный пользователь (может логиниться и работать)
- `blocked` - заблокирован админом (НЕ может логиниться)
- `pending` - не подтверждён / ограниченный доступ (опционально)

**Логика**:
- При `status = 'blocked'` → Login возвращает `ErrAccountBlocked`
- При `status = 'active'` → всё работает нормально
- При `status = 'pending'` → можно добавить ограничения (TODO)

### 2️⃣ User Activity (Активность) - АНАЛИТИКА

**Назначение**: Отслеживание последнего входа пользователя

**Поле в БД**: `last_login TIMESTAMP`

**Обновление**: 
- Автоматически при каждом успешном логине
- В `AuthService.Login()`: вызывается `UpdateLastLogin(userID, NOW())`

**Использование**:
```sql
-- ✅ Активные сегодня (с 00:00 текущего дня) - ПРАВИЛЬНО
WHERE last_login >= DATE_TRUNC('day', NOW())

-- ❌ За последние 24 часа (плавающая цифра) - НЕ рекомендуется
WHERE last_login >= NOW() - INTERVAL '24 hours'

-- ✅ Альтернатива (PostgreSQL)
WHERE DATE(last_login) = CURRENT_DATE
```

**Почему `DATE_TRUNC('day', NOW())` лучше?**
- ✅ Цифры стабильны в течение дня (не "плавают")
- ✅ Понятно: "сегодня" = с 00:00 до текущего момента
- ✅ Админ видит: кто заходил **сегодня**, а не "за последние 24ч"
- ✅ Соответствует бизнес-логике дневных отчётов

### 3️⃣ Premium Status - БИЗНЕС-МОДЕЛЬ

**Статус**: TODO (зависит от бизнес-модели)

**Возможные варианты**:
- Поле `is_premium BOOLEAN` в User
- Отдельная таблица Subscriptions
- Проверка роли (например, `role = 'pro_chef'`)

---

## 🔧 Реализация

### Database Schema (Migration 059)

```sql
-- 1. Account status
ALTER TABLE "User"
ADD COLUMN status VARCHAR(20) NOT NULL DEFAULT 'active';

CREATE INDEX idx_user_status ON "User"(status);

ALTER TABLE "User"
ADD CONSTRAINT check_user_status CHECK (status IN ('active', 'blocked', 'pending'));

-- 2. Activity tracking
ALTER TABLE "User"
ADD COLUMN last_login TIMESTAMP;

CREATE INDEX idx_user_last_login ON "User"(last_login);

-- Initialize for existing users
UPDATE "User"
SET last_login = created_at
WHERE last_login IS NULL;
```

### Model (internal/models/user.go)

```go
const (
    UserStatusActive  = "active"
    UserStatusBlocked = "blocked"
    UserStatusPending = "pending"
)

type User struct {
    ID        string       `gorm:"primaryKey" json:"id"`
    Email     string       `gorm:"unique" json:"email"`
    Name      string       `json:"name"`
    Password  string       `gorm:"column:password" json:"-"`
    Role      string       `gorm:"default:home_chef" json:"role"`
    Status    string       `gorm:"default:active" json:"status"`
    LastLogin *time.Time   `gorm:"column:last_login" json:"lastLogin,omitempty"`
    CreatedAt time.Time    `gorm:"autoCreateTime" json:"createdAt"`
}
```

### Auth Logic (Login)

```go
func (s *AuthService) Login(req dto.LoginRequest) (*dto.AuthResponse, error) {
    // 1. Find user
    user, err := s.repo.FindByEmail(req.Email)
    if err != nil {
        return nil, ErrInvalidCredentials
    }

    // 2. Check if blocked
    if user.Status == models.UserStatusBlocked {
        return nil, ErrAccountBlocked
    }

    // 3. Verify password
    if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
        return nil, ErrInvalidCredentials
    }

    // 4. Update last_login
    now := time.Now()
    s.repo.UpdateLastLogin(user.ID, now)

    // 5. Generate token
    token, err := GenerateToken(user.ID, user.Email, user.Role)
    // ...
}
```

---

## 📈 Admin Statistics API

### Endpoint: `GET /api/admin/users/stats`

**Response**:
```json
{
  "total": 54,
  "active_today": 12,
  "blocked": 2,
  "premium": 0
}
```

**SQL Query (промышленный стандарт)**:
```sql
SELECT
  COUNT(*)                         AS total,
  COUNT(*) FILTER (
    WHERE last_login >= DATE_TRUNC('day', NOW())
  )                                AS active_today,
  COUNT(*) FILTER (
    WHERE status = 'blocked'
  )                                AS blocked
FROM "User";
```

**Преимущества**:
- ✅ Один запрос вместо нескольких
- ✅ Эффективная работа с индексами
- ✅ FILTER clause - стандарт SQL (PostgreSQL 9.4+)
- ✅ Используется в Stripe, Notion, GitHub

---

## 🎨 Frontend Integration

### User List Table

| Колонка | Источник | Пример |
|---------|----------|--------|
| Name | `user.name` | "Дмитрий" |
| Email | `user.email` | "admin@example.com" |
| Role | `user.role` | "admin" / "home_chef" |
| **Status** | `user.status` | 🟢 Active / 🔴 Blocked |
| **Activity** | `user.lastLogin` | "2 години тому" / "Вчора" / "14 днів тому" |

### Relative Time для Activity

```typescript
function formatRelativeTime(lastLogin: string | null): string {
  if (!lastLogin) return 'Ніколи';
  
  const now = new Date();
  const login = new Date(lastLogin);
  const diff = now.getTime() - login.getTime();
  
  const hours = Math.floor(diff / (1000 * 60 * 60));
  const days = Math.floor(diff / (1000 * 60 * 60 * 24));
  
  if (hours < 1) return 'Щойно';
  if (hours < 24) return `${hours} ${pluralize(hours, 'година', 'години', 'годин')} тому`;
  if (days === 1) return 'Вчора';
  if (days < 7) return `${days} ${pluralize(days, 'день', 'дні', 'днів')} тому`;
  if (days < 30) return `${Math.floor(days / 7)} ${pluralize(Math.floor(days / 7), 'тиждень', 'тижні', 'тижнів')} тому`;
  
  return login.toLocaleDateString('uk-UA');
}
```

---

## ⚙️ Configuration

### Environment Variables

Не требуется дополнительных переменных - всё работает на существующем подключении к БД.

### Migration

```bash
# Apply migration (автоматически при деплое на Koyeb)
./bin/server migrate up

# Rollback (если нужно)
./bin/server migrate down
```

---

## 🔒 Security

### Blocked Users

- **Login**: Возвращает `401 Unauthorized` с сообщением "account is blocked"
- **JWT**: Существующие токены продолжают работать (для инвалидации нужно добавить token blacklist)
- **API**: Можно добавить middleware для проверки статуса на каждом запросе

### TODO: Token Invalidation

```go
// Option 1: Check status in auth middleware
func (m *AuthMiddleware) Handle(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        userID := GetUserIDFromContext(r.Context())
        
        user, err := userRepo.FindByID(userID)
        if err != nil || user.Status == models.UserStatusBlocked {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }
        
        next.ServeHTTP(w, r)
    })
}
```

---

## 📊 Analytics Queries

### User Activity Report

```sql
-- Пользователи по активности за последние 30 дней
SELECT
  DATE(last_login) AS date,
  COUNT(*) AS active_users
FROM "User"
WHERE last_login >= NOW() - INTERVAL '30 days'
GROUP BY DATE(last_login)
ORDER BY date DESC;
```

### Inactive Users (Churn Risk)

```sql
-- Пользователи, которые не заходили больше 7 дней
SELECT
  id,
  name,
  email,
  last_login,
  NOW() - last_login AS inactive_duration
FROM "User"
WHERE last_login < NOW() - INTERVAL '7 days'
  AND status = 'active'
ORDER BY last_login ASC;
```

### Status Distribution

```sql
-- Распределение по статусам
SELECT
  status,
  COUNT(*) AS count,
  ROUND(COUNT(*) * 100.0 / SUM(COUNT(*)) OVER (), 2) AS percentage
FROM "User"
GROUP BY status;
```

---

## 🚀 Future Enhancements

### 1. Activity Levels

```sql
ALTER TABLE "User"
ADD COLUMN activity_level VARCHAR(20) DEFAULT 'normal';

-- Categories: inactive, low, normal, high, power_user
-- Based on login frequency, actions performed, etc.
```

### 2. Login History

```sql
CREATE TABLE "LoginHistory" (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id TEXT NOT NULL REFERENCES "User"(id),
  login_at TIMESTAMP NOT NULL DEFAULT NOW(),
  ip_address VARCHAR(45),
  user_agent TEXT,
  success BOOLEAN NOT NULL
);

CREATE INDEX idx_login_history_user ON "LoginHistory"(user_id, login_at DESC);
```

### 3. Premium Status

```sql
-- Option 1: Simple boolean
ALTER TABLE "User"
ADD COLUMN is_premium BOOLEAN DEFAULT FALSE;

-- Option 2: Subscription table
CREATE TABLE "Subscription" (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id TEXT NOT NULL REFERENCES "User"(id),
  plan VARCHAR(50) NOT NULL, -- 'free', 'basic', 'premium', 'pro'
  starts_at TIMESTAMP NOT NULL,
  expires_at TIMESTAMP,
  is_active BOOLEAN NOT NULL DEFAULT TRUE
);
```

---

## 📚 References

- PostgreSQL FILTER clause: https://www.postgresql.org/docs/current/sql-expressions.html#SYNTAX-AGGREGATES
- Industry standard: Stripe Dashboard, Notion Analytics, GitHub Insights
- Migration file: `/migrations/059_add_user_status_and_activity.sql`
