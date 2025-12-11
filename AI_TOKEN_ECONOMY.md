# 💰 AI Token Economy System

## Обзор

Система токен-экономики для AI-запросов с динамическим ценообразованием и прозрачным отслеживанием расходов.

---

## 1. Механика расчёта стоимости AI-запросов

### Вариант 1: Фиксированная стоимость (Простой)
```go
const AI_REQUEST_COST = 1

func CalculateCost(text string) int64 {
    return AI_REQUEST_COST
}
```
- ✅ **Преимущества**: Просто, предсказуемо
- ❌ **Недостатки**: Не учитывает сложность запроса

---

### Вариант 2: По длине запроса (Рекомендуется)
```go
func CalculateCostByLength(text string) int64 {
    length := len(text)
    
    if length < 50 {
        return 1  // Короткий вопрос
    }
    if length < 200 {
        return 3  // Средний запрос
    }
    if length < 500 {
        return 5  // Длинный запрос
    }
    return 10  // Очень длинный текст
}
```

**Примеры:**
- "Что такое рецепт?" → 1 токен
- "Как приготовить пасту карбонара с пошаговыми инструкциями?" → 3 токена
- "Сгенерируй полный план питания на неделю для семьи из 4 человек..." → 10 токенов

---

### Вариант 3: По уровню сложности (Премиум)
```go
type AIComplexity string

const (
    Basic    AIComplexity = "basic"    // Простые вопросы
    Pro      AIComplexity = "pro"      // Генерация контента
    Advanced AIComplexity = "advanced" // Сложные задачи
)

func CalculateCostByComplexity(complexity AIComplexity) int64 {
    switch complexity {
    case Basic:
        return 1  // Простой вопрос-ответ
    case Pro:
        return 5  // Генерация рецепта/плана
    case Advanced:
        return 10 // Персонализированный анализ
    default:
        return 1
    }
}
```

**Пользователь выбирает уровень:**
```json
{
  "message": "Сгенерируй рецепт",
  "complexity": "pro"
}
```

---

### Вариант 4: Комбинированный (Оптимальный)
```go
func CalculateCost(text string, complexity AIComplexity) int64 {
    // Базовая стоимость по длине
    baseCost := int64(1)
    length := len(text)
    
    if length > 500 {
        baseCost = 5
    } else if length > 200 {
        baseCost = 3
    }
    
    // Множитель по сложности
    multiplier := int64(1)
    switch complexity {
    case "basic":
        multiplier = 1
    case "pro":
        multiplier = 2
    case "advanced":
        multiplier = 3
    }
    
    return baseCost * multiplier
}
```

**Таблица стоимости:**
| Длина \ Сложность | Basic | Pro | Advanced |
|-------------------|-------|-----|----------|
| < 50 символов     | 1     | 2   | 3        |
| 50-200            | 3     | 6   | 9        |
| 200-500           | 5     | 10  | 15       |
| > 500             | 5     | 10  | 15       |

---

## 2. Прайс-лист операций

```go
var TokenPricing = map[string]int64{
    // AI Operations
    "ai_simple_question":     1,
    "ai_recipe_generation":   5,
    "ai_meal_plan_week":      10,
    "ai_nutrition_analysis":  3,
    "ai_cooking_assistant":   2,
    "ai_image_recognition":   7,
    
    // Marketplace
    "premium_recipe_unlock":  25,
    "chef_course_access":     50,
    "pro_subscription_month": 100,
    
    // Features
    "export_recipe_pdf":      2,
    "share_private_recipe":   1,
    "calendar_sync":          5,
}
```

---

## 3. Пользовательский опыт (UX)

### 3.1 Перед AI-запросом
```
┌──────────────────────────────────────┐
│  🤖 AI Cooking Assistant             │
├──────────────────────────────────────┤
│  💰 Ваш баланс: 97 токенов          │
│  💸 Стоимость ответа: ~3 токена     │
│                                      │
│  [Введите ваш вопрос...]            │
│                                      │
│  ○ Basic (1x)  ◉ Pro (2x)  ○ Adv (3x)│
│                                      │
│  [ Отправить запрос ]               │
└──────────────────────────────────────┘
```

### 3.2 После успешного запроса
```
┌──────────────────────────────────────┐
│  ✅ Запрос обработан                │
│  💰 Списано: 3 токена               │
│  💵 Новый баланс: 94 токена         │
│                                      │
│  🤖 Ответ:                          │
│  Вот пошаговый рецепт карбонары...  │
└──────────────────────────────────────┘
```

### 3.3 Недостаточно токенов
```
┌──────────────────────────────────────┐
│  ⚠️ Недостаточно токенов             │
│                                      │
│  Требуется: 5 токенов               │
│  У вас: 2 токена                    │
│                                      │
│  💡 Как получить токены:            │
│  • Выполните ежедневное задание     │
│  • Пригласите друга (+50 токенов)   │
│  • Разместите рецепт (+10 токенов)  │
│                                      │
│  [ Перейти к заданиям ]             │
└──────────────────────────────────────┘
```

---

## 4. API Спецификация

### POST `/api/ai/chat`

**Request:**
```json
{
  "message": "Как приготовить пасту карбонара?",
  "complexity": "pro"
}
```

**Response (Success):**
```json
{
  "success": true,
  "cost": 5,
  "tokens_spent": 5,
  "balance_before": 100,
  "balance_after": 95,
  "answer": "Для приготовления классической пасты карбонара вам понадобится...",
  "timestamp": "2025-12-11T12:34:56Z"
}
```

**Response (Insufficient Tokens):**
```json
{
  "success": false,
  "error": "not enough tokens to process AI request",
  "required": 5,
  "available": 2,
  "shortfall": 3,
  "suggestions": [
    {
      "action": "complete_daily_task",
      "reward": 10,
      "description": "Выполните ежедневное задание"
    },
    {
      "action": "invite_friend",
      "reward": 50,
      "description": "Пригласите друга"
    }
  ]
}
```

---

## 5. История транзакций (Token Transactions)

### Таблица `token_transactions`
```sql
CREATE TABLE token_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id TEXT NOT NULL REFERENCES "User"(id),
    amount BIGINT NOT NULL,
    type VARCHAR(50) NOT NULL, -- 'earn' или 'spend'
    reason VARCHAR(100) NOT NULL,
    description TEXT,
    balance_before BIGINT NOT NULL,
    balance_after BIGINT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_transactions_user_id ON token_transactions(user_id);
CREATE INDEX idx_transactions_type ON token_transactions(type);
CREATE INDEX idx_transactions_created_at ON token_transactions(created_at DESC);
```

### Примеры записей:
```sql
-- AI Request
INSERT INTO token_transactions (user_id, amount, type, reason, description, balance_before, balance_after)
VALUES ('user-123', -5, 'spend', 'ai_request', 'AI Recipe Generation (Pro)', 100, 95);

-- Task Reward
INSERT INTO token_transactions (user_id, amount, type, reason, description, balance_before, balance_after)
VALUES ('user-123', 50, 'earn', 'task_completion', 'Completed: First Recipe', 95, 145);

-- Welcome Bonus
INSERT INTO token_transactions (user_id, amount, type, reason, description, balance_before, balance_after)
VALUES ('user-123', 100, 'earn', 'welcome_bonus', 'New user welcome bonus', 0, 100);
```

---

## 6. Реализация в коде

### 6.1 AI Service с расчётом стоимости
```go
// internal/modules/ai/service/service.go

type AIService interface {
    Chat(userID, message string, complexity string) (*ChatResponse, error)
    CalculateCost(message string, complexity string) int64
}

func (s *aiService) CalculateCost(message string, complexity string) int64 {
    baseCost := int64(1)
    length := len(message)
    
    if length > 500 {
        baseCost = 5
    } else if length > 200 {
        baseCost = 3
    }
    
    multiplier := int64(1)
    switch complexity {
    case "pro":
        multiplier = 2
    case "advanced":
        multiplier = 3
    default:
        multiplier = 1
    }
    
    return baseCost * multiplier
}
```

### 6.2 AI Handler с проверкой токенов
```go
// internal/modules/ai/transport/http/handlers.go

func (h *AIHandler) Chat(w http.ResponseWriter, r *http.Request) {
    // 1. Извлекаем userID из контекста (JWT)
    claims, ok := r.Context().Value(middleware.UserContextKey).(*auth.Claims)
    if !ok {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }
    userID := claims.UserID
    
    // 2. Парсим запрос
    var req ChatRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }
    
    // 3. Рассчитываем стоимость
    cost := h.aiService.CalculateCost(req.Message, req.Complexity)
    
    // 4. Получаем текущий баланс
    balanceBefore, err := h.tokenBankService.GetUserBalance(userID)
    if err != nil {
        http.Error(w, "Failed to get balance", http.StatusInternalServerError)
        return
    }
    
    // 5. Проверяем и списываем токены
    if err := h.tokenBankService.SpendTokensForAIRequest(userID, cost); err != nil {
        if err.Error() == "not enough tokens to process AI request" {
            json.NewEncoder(w).Encode(map[string]interface{}{
                "success": false,
                "error": "insufficient_tokens",
                "required": cost,
                "available": balanceBefore,
                "message": "Недостаточно токенов для выполнения запроса",
            })
            return
        }
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    // 6. Генерируем AI ответ
    answer, err := h.aiService.GenerateResponse(req.Message)
    if err != nil {
        // Возвращаем токены обратно при ошибке AI
        h.tokenBankService.EarnTokens(userID, cost, "ai_error_refund")
        http.Error(w, "AI service error", http.StatusInternalServerError)
        return
    }
    
    // 7. Возвращаем успешный ответ
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(ChatResponse{
        Success:       true,
        Cost:          cost,
        BalanceBefore: balanceBefore,
        BalanceAfter:  balanceBefore - cost,
        Answer:        answer,
        Timestamp:     time.Now(),
    })
}
```

---

## 7. Способы получения токенов

### 7.1 Ежедневные задания
- Войти в приложение: **+10 токенов**
- Добавить 5 ингредиентов в холодильник: **+30 токенов**
- Приготовить рецепт дня: **+75 токенов**

### 7.2 Социальные действия
- Поделиться рецептом: **+25 токенов**
- Пригласить друга: **+150 токенов**
- Оставить отзыв на рецепт: **+5 токенов**

### 7.3 Достижения
- Создать первый рецепт: **+50 токенов**
- Мастер-шеф (10 рецептов): **+200 токенов**
- Изучить основы кулинарии: **+100 токенов**

### 7.4 Покупка (опционально)
- Пакет "Старт": 500 токенов за $2.99
- Пакет "Про": 1500 токенов за $7.99
- Пакет "Шеф": 5000 токенов за $19.99

---

## 8. Dashboard для пользователя

```
┌─────────────────────────────────────────────┐
│  💰 Мой токен-банк                         │
├─────────────────────────────────────────────┤
│                                             │
│  Текущий баланс: 237 токенов               │
│  ━━━━━━━━━━━━━━━━━━━━━━━ 24% от макс.     │
│                                             │
│  📊 Статистика:                            │
│  • Всего получено: 500 токенов            │
│  • Потрачено: 263 токена                  │
│  • AI запросов: 47                        │
│                                             │
│  📈 Последние транзакции:                  │
│  ─────────────────────────────────────────  │
│  -5  AI: генерация рецепта      2 мин назад│
│  +50 Задание: пригласи друга    1 час назад│
│  -3  AI: простой вопрос         3 часа    │
│  +100 Welcome bonus             Вчера     │
│                                             │
│  [ Получить токены ]  [ История ]          │
└─────────────────────────────────────────────┘
```

---

## 9. Метрики для аналитики

### Backend логирование
```go
log.Printf("[TOKEN_SPEND] user=%s, amount=%d, reason=%s, balance_after=%d", 
    userID, cost, "ai_request", newBalance)
```

### Метрики для мониторинга
- Средняя стоимость AI-запроса
- Токенов потрачено в день
- Конверсия: пользователи без токенов → выполнение заданий
- Топ-5 способов получения токенов

---

## 10. Рекомендации

### ✅ Лучшие практики
1. **Прозрачность**: Всегда показывайте стоимость ДО запроса
2. **Предупреждения**: Уведомляйте при балансе < 10 токенов
3. **Возврат**: Возвращайте токены при ошибках AI
4. **Логирование**: Записывайте все транзакции для аудита

### 💡 Советы по ценообразованию
- Начните с низких цен (1-5 токенов)
- Давайте щедрые бонусы за регистрацию (100 токенов)
- Делайте ежедневные задания доступными (10-30 токенов)
- Премиум-функции должны стоить дороже (50-100 токенов)

---

## Итоговая токен-экономика

```
                     TREASURY (1,000,000,000)
                            ↓
        ┌──────────────────┴──────────────────┐
        ↓                                      ↓
   ALLOCATE (выдача)                     SPEND (возврат)
        ↓                                      ↑
    ┌───┴────┬───────┬─────────┐              │
    ↓        ↓       ↓         ↓              │
  Welcome  Tasks  Achievements  Friends       │
  Bonus                                        │
    │        │       │         │               │
    └───┬────┴───┬───┴────┬────┘               │
        ↓        ↓        ↓                    │
      USER BALANCE (активный)                  │
            │                                  │
            └──────────────────────────────────┘
                   AI, Marketplace, Features
```

**Замкнутый цикл:** Токены выдаются из Treasury → пользователь тратит → возвращаются в Treasury

**Fixed Supply:** Общее количество всегда = 1 млрд токенов
