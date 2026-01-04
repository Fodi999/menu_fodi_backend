# 📊 Как проверить активность пользователей

## 🚀 Быстрая проверка (рекомендуется)

```bash
./check_activity.sh
```

Покажет:
- ✅ Общая статистика (всего, активные сегодня/неделя/месяц)
- ✅ Разбивка по ролям (home_chef, admin, investor)
- ✅ Риск оттока (Churn Risk) - пользователи неактивные 14+ дней
- ✅ ТОП-10 самых активных пользователей

---

## 📈 Текущая статистика (4 января 2026)

### Общая картина
```
📈 Total Users:          54
🟢 Active Today:          0  (0%)
🟡 Active This Week:      0  (0%)
🟠 Active This Month:    14  (25.9%)
🔴 Inactive 30+ days:    40  (74.1%)
⛔ Blocked:               0  (0%)
```

### По ролям
```
Role         Total  Active 30d  Active %
-----------  -----  ----------  --------
home_chef      49        13      26.5%
admin           4         1      25.0%
investor        1         0       0.0%
```

### ⚠️ Churn Risk
- **54 пользователя** (100%) не заходили больше 14 дней
- Это нормально для тестового окружения
- При запуске продакшена нужно отслеживать этот показатель

---

## 🔍 Детальный анализ

### Полный SQL анализ (10 разных отчётов)
```bash
psql "$DATABASE_URL" -f sql/check_user_activity.sql
```

### Конкретные пользователи

#### ТОП-10 активных
```sql
SELECT name, email, last_login 
FROM "User" 
ORDER BY last_login DESC 
LIMIT 10;
```

#### Неактивные 7+ дней
```sql
SELECT name, email, last_login, NOW() - last_login AS inactive_for
FROM "User"
WHERE last_login < NOW() - INTERVAL '7 days'
ORDER BY last_login ASC;
```

#### По дате последнего входа
```sql
SELECT 
  DATE(last_login) AS date,
  COUNT(*) AS logins
FROM "User"
WHERE last_login >= NOW() - INTERVAL '30 days'
GROUP BY DATE(last_login)
ORDER BY date DESC;
```

---

## 🎯 Индикаторы активности

### 🟢 Активный пользователь
- **Заходил за последние 7 дней**
- Показатель здоровья продукта: **>40%** от всех пользователей

### 🟡 Умеренно активный
- **Заходил за последние 30 дней**
- Можно привлечь email-напоминаниями или push-уведомлениями

### 🔴 Риск оттока (Churn Risk)
- **Не заходил 14+ дней**
- Нужны меры удержания (реактивация, спецпредложения)

### ⛔ Заблокирован
- **status = 'blocked'**
- Не может войти в систему

---

## 📊 API для фронтенда

### GET /api/admin/users/stats

```json
{
  "total": 54,
  "active_today": 0,
  "blocked": 0,
  "premium": 0
}
```

### GET /api/admin/users

```json
{
  "users": [
    {
      "id": "...",
      "name": "John Doe",
      "email": "john@example.com",
      "role": "home_chef",
      "status": "active",
      "lastLogin": "2025-12-21T13:06:21.026Z",
      "createdAt": "2025-11-15T10:20:30.000Z"
    }
  ],
  "meta": {
    "total": 54,
    "page": 1,
    "limit": 20,
    "totalPages": 3
  }
}
```

---

## 🎨 Отображение в UI

### Статус активности
```typescript
function getActivityStatus(lastLogin: string | null): string {
  if (!lastLogin) return '❌ Ніколи';
  
  const now = new Date();
  const login = new Date(lastLogin);
  const hours = (now.getTime() - login.getTime()) / (1000 * 60 * 60);
  
  if (hours < 1) return '🟢 Щойно';
  if (hours < 24) return '🟢 Сьогодні';
  if (hours < 168) return '🟡 Цього тижня';  // 7 days
  if (hours < 720) return '🟠 Цього місяця';  // 30 days
  return '🔴 Давно';
}
```

### Относительное время
```typescript
function formatRelativeTime(lastLogin: string): string {
  const now = new Date();
  const login = new Date(lastLogin);
  const diff = now.getTime() - login.getTime();
  
  const minutes = Math.floor(diff / (1000 * 60));
  const hours = Math.floor(diff / (1000 * 60 * 60));
  const days = Math.floor(diff / (1000 * 60 * 60 * 24));
  
  if (minutes < 1) return 'Щойно';
  if (minutes < 60) return `${minutes} хв тому`;
  if (hours < 24) return `${hours} год тому`;
  if (days === 1) return 'Вчора';
  if (days < 7) return `${days} дн тому`;
  if (days < 30) return `${Math.floor(days / 7)} тиж тому`;
  
  return login.toLocaleDateString('uk-UA');
}
```

---

## 🔄 Автоматическое обновление

### При логине
```go
// internal/modules/auth/service/service.go
func (s *AuthService) Login(req dto.LoginRequest) (*dto.AuthResponse, error) {
    // ... проверки ...
    
    // ✅ Автоматически обновляет last_login
    now := time.Now()
    s.repo.UpdateLastLogin(user.ID, now)
    
    // ... генерация токена ...
}
```

### В базе данных
```sql
-- last_login обновляется при каждом успешном входе
UPDATE "User"
SET last_login = NOW()
WHERE id = $1;
```

---

## 📈 Метрики для отслеживания

### 1. DAU (Daily Active Users)
```sql
SELECT COUNT(DISTINCT user_id) 
FROM "User"
WHERE last_login >= CURRENT_DATE;
```

### 2. WAU (Weekly Active Users)
```sql
SELECT COUNT(DISTINCT user_id)
FROM "User" 
WHERE last_login >= NOW() - INTERVAL '7 days';
```

### 3. MAU (Monthly Active Users)
```sql
SELECT COUNT(DISTINCT user_id)
FROM "User"
WHERE last_login >= NOW() - INTERVAL '30 days';
```

### 4. Retention Rate (7 дней)
```sql
SELECT 
  ROUND(
    COUNT(*) FILTER (WHERE last_login >= NOW() - INTERVAL '7 days') * 100.0 / 
    COUNT(*), 
    2
  ) AS retention_7d
FROM "User"
WHERE "createdAt" < NOW() - INTERVAL '7 days';
```

### 5. Churn Rate (30 дней)
```sql
SELECT 
  ROUND(
    COUNT(*) FILTER (WHERE last_login < NOW() - INTERVAL '30 days' OR last_login IS NULL) * 100.0 / 
    COUNT(*), 
    2
  ) AS churn_30d
FROM "User";
```

---

## 🎯 Рекомендации

### Для активации пользователей
1. **Email-уведомления** для неактивных 7+ дней
2. **Push-уведомления** о новых рецептах
3. **Персональные рекомендации** на основе холодильника
4. **Gamification** (achievements, streaks)

### Для удержания (retention)
1. **Weekly digest** с популярными рецептами
2. **Напоминания** об истекающих продуктах
3. **Социальные функции** (друзья, sharing)
4. **Премиум-функции** для активных пользователей

### Для снижения оттока (churn)
1. **Win-back campaigns** для неактивных 14+ дней
2. **Опросы** - почему не пользуются приложением
3. **Специальные предложения** для возвращения
4. **Упрощение** сложных функций

---

## 🔗 Связанные документы

- [USER_STATUS_AND_ACTIVITY.md](./USER_STATUS_AND_ACTIVITY.md) - Полная документация
- [sql/check_user_activity.sql](../sql/check_user_activity.sql) - 10 SQL отчётов
- [check_activity.sh](../check_activity.sh) - Быстрый скрипт проверки

---

## ⚡ Быстрые команды

```bash
# Проверить активность
./check_activity.sh

# Полный анализ
psql "$DATABASE_URL" -f sql/check_user_activity.sql

# Только статистика
psql "$DATABASE_URL" -c "SELECT COUNT(*) FILTER (WHERE last_login >= NOW() - INTERVAL '24 hours') AS active_today, COUNT(*) AS total FROM \"User\";"

# ТОП-5 активных
psql "$DATABASE_URL" -c "SELECT name, email, last_login FROM \"User\" ORDER BY last_login DESC LIMIT 5;"

# Неактивные 30+ дней
psql "$DATABASE_URL" -c "SELECT COUNT(*) FROM \"User\" WHERE last_login < NOW() - INTERVAL '30 days';"
```
