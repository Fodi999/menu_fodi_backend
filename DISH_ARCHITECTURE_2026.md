# 🍽️ Dish (Карточка блюда) — Архитектура 2026

**Дата:** 2026-01-26  
**Статус:** 📋 Спецификация  
**Принцип:** Копируем Recipe pipeline → адаптируем под Dish

---

## 🎯 Ключевой принцип

### ❗ НЕ придумываем новую архитектуру
### ✅ Копируем pipeline рецепта → адаптируем под Dish

**У тебя уже есть:**
- ✅ `CreateRecipeWithAI`
- ✅ `draft` / `approve` статусы
- ✅ Admin редактирует
- ✅ History лог
- ✅ Status lifecycle

**Это золото. Мы его не выбрасываем.**

---

## 🧩 Recipe vs Dish — Сравнение

| Параметр         | Recipe                      | Dish                           |
|------------------|-----------------------------|--------------------------------|
| **Назначение**   | Технология приготовления    | Продажа готового блюда         |
| **AI**           | Генерирует шаги             | Генерирует цену + описание     |
| **Статус**       | draft → approved            | draft → approved → published   |
| **Редактирование** | Да                        | Да                             |
| **Автор**        | user / admin                | admin                          |
| **Видимость**    | Каталог рецептов            | Marketplace (меню)             |
| **Цена**         | ❌ Нет                      | ✅ Да (cost, price, margin)    |
| **Маржа**        | ❌ Нет                      | ✅ Да (%)                      |
| **Холодильник**  | Проверка `canCookNow`       | Проверка `isAvailable`         |

---

## 🏗️ 1. Модель Dish (БД)

### Структура таблицы `dishes`

```sql
CREATE TABLE dishes (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    recipe_id       UUID NOT NULL REFERENCES recipe_catalog(id) ON DELETE CASCADE,
    
    -- Контент
    title           VARCHAR(255) NOT NULL,
    description     TEXT,
    image_url       TEXT,
    
    -- Финансы
    cost            DECIMAL(10,2) NOT NULL,  -- Себестоимость (из холодильника)
    price           DECIMAL(10,2) NOT NULL,  -- Цена продажи
    margin          DECIMAL(5,2) NOT NULL,   -- Маржа в %
    
    -- Статус
    status          VARCHAR(20) NOT NULL DEFAULT 'draft',  -- draft | approved | published
    is_available    BOOLEAN NOT NULL DEFAULT true,         -- Зависит от холодильника
    
    -- Метаданные
    created_by      UUID NOT NULL REFERENCES "User"(id),
    approved_by     UUID REFERENCES "User"(id),
    approved_at     TIMESTAMP,
    
    created_at      TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP NOT NULL DEFAULT NOW(),
    
    -- Индексы
    CONSTRAINT dishes_status_check CHECK (status IN ('draft', 'approved', 'published'))
);

CREATE INDEX idx_dishes_recipe_id ON dishes(recipe_id);
CREATE INDEX idx_dishes_status ON dishes(status);
CREATE INDEX idx_dishes_is_available ON dishes(is_available);
CREATE INDEX idx_dishes_created_by ON dishes(created_by);
```

### Go Model

```go
// internal/models/dish.go

package models

import (
    "time"
    "github.com/google/uuid"
)

// DishStatus represents the lifecycle state of a dish
type DishStatus string

const (
    DishStatusDraft     DishStatus = "draft"      // AI generated, needs review
    DishStatusApproved  DishStatus = "approved"   // Admin reviewed & approved
    DishStatusPublished DishStatus = "published"  // Available for customers
)

// Dish represents a commercial dish card for marketplace
type Dish struct {
    ID          uuid.UUID  `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
    RecipeID    uuid.UUID  `gorm:"type:uuid;not null" json:"recipeId"`
    
    // Content
    Title       string     `gorm:"type:varchar(255);not null" json:"title"`
    Description string     `gorm:"type:text" json:"description"`
    ImageURL    string     `gorm:"type:text" json:"imageUrl"`
    
    // Finance
    Cost        float64    `gorm:"type:decimal(10,2);not null" json:"cost"`        // Cost to make
    Price       float64    `gorm:"type:decimal(10,2);not null" json:"price"`       // Selling price
    Margin      float64    `gorm:"type:decimal(5,2);not null" json:"margin"`       // Margin %
    
    // Status
    Status      DishStatus `gorm:"type:varchar(20);not null;default:'draft'" json:"status"`
    IsAvailable bool       `gorm:"not null;default:true" json:"isAvailable"`
    
    // Metadata
    CreatedBy   uuid.UUID  `gorm:"type:uuid;not null" json:"createdBy"`
    ApprovedBy  *uuid.UUID `gorm:"type:uuid" json:"approvedBy,omitempty"`
    ApprovedAt  *time.Time `json:"approvedAt,omitempty"`
    
    CreatedAt   time.Time  `gorm:"not null;default:NOW()" json:"createdAt"`
    UpdatedAt   time.Time  `gorm:"not null;default:NOW()" json:"updatedAt"`
    
    // Relations
    Recipe      RecipeCatalog `gorm:"foreignKey:RecipeID" json:"recipe,omitempty"`
    Creator     User          `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
    Approver    *User         `gorm:"foreignKey:ApprovedBy" json:"approver,omitempty"`
}

func (Dish) TableName() string {
    return "dishes"
}
```

---

## 🤖 2. AI Сервис генерации блюд

### Файл: `internal/modules/admin/service/dish_ai.go`

```go
package service

import (
    "context"
    "encoding/json"
    "fmt"
    "github.com/google/uuid"
    "github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
)

// ===========================
// DTO для генерации блюда
// ===========================

type GenerateDishRequest struct {
    RecipeID     string  `json:"recipeId" binding:"required"`
    TargetMargin float64 `json:"targetMargin" binding:"required,min=0,max=100"`
    Language     string  `json:"language"` // en, pl, ru
}

type GenerateDishResponse struct {
    DishID      uuid.UUID `json:"dishId"`
    Title       string    `json:"title"`
    Description string    `json:"description"`
    Cost        float64   `json:"cost"`
    Price       float64   `json:"price"`
    Margin      float64   `json:"margin"`
    Status      string    `json:"status"`
}

// ===========================
// AI Service Method
// ===========================

// GenerateDishWithAI создаёт карточку блюда через AI (аналог CreateRecipeWithAI)
func (s *adminService) GenerateDishWithAI(req GenerateDishRequest, adminID string) (*models.Dish, error) {
    // 1️⃣ Загружаем рецепт из каталога
    recipe, err := s.loadRecipeByID(req.RecipeID)
    if err != nil {
        return nil, fmt.Errorf("recipe not found: %w", err)
    }
    
    // 2️⃣ Проверяем, можно ли приготовить блюдо (canCookNow)
    canCook, missingIngredients, err := s.checkRecipeAvailability(recipe.ID)
    if err != nil {
        return nil, fmt.Errorf("failed to check recipe availability: %w", err)
    }
    
    if !canCook {
        return nil, fmt.Errorf("cannot create dish: missing ingredients: %v", missingIngredients)
    }
    
    // 3️⃣ Рассчитываем себестоимость блюда
    cost, err := s.calculateDishCost(recipe)
    if err != nil {
        return nil, fmt.Errorf("failed to calculate cost: %w", err)
    }
    
    // 4️⃣ Рассчитываем цену на основе маржи
    price := s.calculatePrice(cost, req.TargetMargin)
    
    // 5️⃣ Формируем prompt для AI (генерация описания и названия)
    aiPrompt := s.buildDishPrompt(recipe, cost, price, req.TargetMargin, req.Language)
    
    // 6️⃣ Вызываем AI для генерации контента
    aiResponse, err := s.generateDishContentViaAI(aiPrompt)
    if err != nil {
        return nil, fmt.Errorf("AI generation failed: %w", err)
    }
    
    // 7️⃣ Сохраняем блюдо как draft
    dish := &models.Dish{
        ID:          uuid.New(),
        RecipeID:    uuid.MustParse(req.RecipeID),
        Title:       aiResponse.Title,
        Description: aiResponse.Description,
        ImageURL:    recipe.ImageURL, // Используем изображение рецепта
        Cost:        cost,
        Price:       price,
        Margin:      req.TargetMargin,
        Status:      models.DishStatusDraft,
        IsAvailable: true,
        CreatedBy:   uuid.MustParse(adminID),
    }
    
    if err := s.db.Create(dish).Error; err != nil {
        return nil, fmt.Errorf("failed to save dish: %w", err)
    }
    
    // 8️⃣ Логируем событие в history
    s.logDishEvent(adminID, dish.ID.String(), "dish_created", map[string]interface{}{
        "recipe_id": req.RecipeID,
        "cost":      cost,
        "price":     price,
        "margin":    req.TargetMargin,
    })
    
    logger.Info("Dish created via AI",
        zap.String("dish_id", dish.ID.String()),
        zap.String("recipe_id", req.RecipeID),
        zap.Float64("cost", cost),
        zap.Float64("price", price),
        zap.Float64("margin", req.TargetMargin),
    )
    
    return dish, nil
}

// ===========================
// Вспомогательные методы
// ===========================

// calculateDishCost рассчитывает себестоимость блюда на основе ингредиентов
func (s *adminService) calculateDishCost(recipe *models.RecipeCatalog) (float64, error) {
    totalCost := 0.0
    
    for _, ingredient := range recipe.Ingredients {
        // Пропускаем optional ингредиенты
        if ingredient.Optional {
            continue
        }
        
        // Загружаем актуальную цену из БД
        var catalogIngredient models.Ingredient
        if err := s.db.First(&catalogIngredient, "id = ?", ingredient.IngredientID).Error; err != nil {
            return 0, fmt.Errorf("ingredient not found: %s", ingredient.IngredientID)
        }
        
        pricePerUnit := 0.0
        if catalogIngredient.DefaultPricePerUnit != nil {
            pricePerUnit = *catalogIngredient.DefaultPricePerUnit
        }
        
        ingredientCost := ingredient.Quantity * pricePerUnit
        totalCost += ingredientCost
    }
    
    // Округляем до 2 знаков
    return math.Round(totalCost*100) / 100, nil
}

// calculatePrice рассчитывает цену на основе себестоимости и маржи
func (s *adminService) calculatePrice(cost float64, marginPercent float64) float64 {
    // Формула: Price = Cost / (1 - Margin/100)
    // Пример: Cost = 10, Margin = 60% → Price = 10 / (1 - 0.6) = 25
    price := cost / (1 - marginPercent/100)
    return math.Round(price*100) / 100
}

// buildDishPrompt формирует prompt для AI
func (s *adminService) buildDishPrompt(
    recipe *models.RecipeCatalog,
    cost, price, margin float64,
    language string,
) string {
    return fmt.Sprintf(`
You are a professional restaurant menu writer.

Task: Create a compelling dish card for a restaurant menu.

Recipe Information:
- Name: %s
- Description: %s
- Cooking Time: %d minutes
- Difficulty: %s
- Category: %s

Pricing:
- Cost: %.2f PLN
- Price: %.2f PLN
- Margin: %.0f%%

Requirements:
1. Create an attractive dish title (short, appetizing)
2. Write a compelling description (2-3 sentences, highlight key ingredients and cooking method)
3. Language: %s
4. Format: JSON

Output format:
{
  "title": "Dish name",
  "description": "Compelling description that makes customers want to order"
}
`, 
        recipe.Title,
        recipe.Description,
        recipe.TimeMinutes,
        recipe.Difficulty,
        recipe.Category,
        cost,
        price,
        margin,
        language,
    )
}

// generateDishContentViaAI вызывает AI для генерации контента
func (s *adminService) generateDishContentViaAI(prompt string) (*DishAIResponse, error) {
    // TODO: Вызвать Groq/OpenAI API
    // Аналогично generateRecipeViaAI
    
    // Временная заглушка для примера
    return &DishAIResponse{
        Title:       "AI Generated Title",
        Description: "AI Generated Description",
    }, nil
}

type DishAIResponse struct {
    Title       string `json:"title"`
    Description string `json:"description"`
}
```

---

## 🔄 3. Lifecycle статусов

### Диаграмма состояний

```
┌─────────┐
│  draft  │ ← AI генерирует блюдо
└────┬────┘
     │ admin редактирует (цена, описание, название)
     ↓
┌──────────┐
│ approved │ ← Admin утверждает
└────┬─────┘
     │ публикация
     ↓
┌───────────┐
│ published │ ← Доступно покупателям
└───────────┘
```

### Переходы статусов

```go
// ApproveDish утверждает блюдо (draft → approved)
func (s *adminService) ApproveDish(dishID, adminID string) error {
    var dish models.Dish
    if err := s.db.First(&dish, "id = ?", dishID).Error; err != nil {
        return fmt.Errorf("dish not found: %w", err)
    }
    
    if dish.Status != models.DishStatusDraft {
        return fmt.Errorf("only draft dishes can be approved")
    }
    
    approvedAt := time.Now()
    approverID := uuid.MustParse(adminID)
    
    updates := map[string]interface{}{
        "status":      models.DishStatusApproved,
        "approved_by": approverID,
        "approved_at": approvedAt,
        "updated_at":  time.Now(),
    }
    
    if err := s.db.Model(&dish).Updates(updates).Error; err != nil {
        return fmt.Errorf("failed to approve dish: %w", err)
    }
    
    // Логируем в history
    s.logDishEvent(adminID, dishID, "dish_approved", map[string]interface{}{
        "old_status": models.DishStatusDraft,
        "new_status": models.DishStatusApproved,
    })
    
    return nil
}

// PublishDish публикует блюдо (approved → published)
func (s *adminService) PublishDish(dishID, adminID string) error {
    var dish models.Dish
    if err := s.db.First(&dish, "id = ?", dishID).Error; err != nil {
        return fmt.Errorf("dish not found: %w", err)
    }
    
    if dish.Status != models.DishStatusApproved {
        return fmt.Errorf("only approved dishes can be published")
    }
    
    // Проверяем доступность ингредиентов перед публикацией
    canCook, _, err := s.checkRecipeAvailability(dish.RecipeID)
    if err != nil {
        return fmt.Errorf("failed to check availability: %w", err)
    }
    
    if !canCook {
        return fmt.Errorf("cannot publish: ingredients are not available")
    }
    
    updates := map[string]interface{}{
        "status":       models.DishStatusPublished,
        "is_available": true,
        "updated_at":   time.Now(),
    }
    
    if err := s.db.Model(&dish).Updates(updates).Error; err != nil {
        return fmt.Errorf("failed to publish dish: %w", err)
    }
    
    // Логируем в history
    s.logDishEvent(adminID, dishID, "dish_published", map[string]interface{}{
        "old_status": models.DishStatusApproved,
        "new_status": models.DishStatusPublished,
    })
    
    return nil
}
```

---

## 🌐 4. API Endpoints

### 4.1 Генерация блюда

```
POST /api/admin/dishes/generate-from-recipe
Authorization: Bearer <super_admin_token>
Content-Type: application/json

{
  "recipeId": "uuid",
  "targetMargin": 65,
  "language": "en"
}

Response 201 Created:
{
  "dishId": "uuid",
  "title": "AI Generated Title",
  "description": "Compelling description...",
  "cost": 15.50,
  "price": 44.29,
  "margin": 65.0,
  "status": "draft"
}
```

### 4.2 Редактирование блюда

```
PATCH /api/admin/dishes/{id}
Authorization: Bearer <admin_token>
Content-Type: application/json

{
  "title": "Updated Title",
  "description": "Updated description",
  "price": 45.00,
  "margin": 70.0
}

Response 200 OK:
{
  "message": "Dish updated successfully",
  "dish": { ... }
}
```

### 4.3 Утверждение блюда

```
POST /api/admin/dishes/{id}/approve
Authorization: Bearer <admin_token>

Response 200 OK:
{
  "message": "Dish approved successfully",
  "status": "approved"
}
```

### 4.4 Публикация блюда

```
POST /api/admin/dishes/{id}/publish
Authorization: Bearer <super_admin_token>

Response 200 OK:
{
  "message": "Dish published successfully",
  "status": "published",
  "isAvailable": true
}
```

### 4.5 Список блюд (админка)

```
GET /api/admin/dishes?status=draft&limit=20
Authorization: Bearer <admin_token>

Response 200 OK:
{
  "data": [
    {
      "id": "uuid",
      "recipeId": "uuid",
      "title": "...",
      "description": "...",
      "cost": 15.50,
      "price": 44.29,
      "margin": 65.0,
      "status": "draft",
      "isAvailable": true,
      "createdAt": "2026-01-26T...",
      "recipe": { ... }
    }
  ],
  "count": 5
}
```

### 4.6 Marketplace (публичные блюда)

```
GET /api/marketplace/dishes?category=Main&available=true
(Optional auth)

Response 200 OK:
{
  "data": [
    {
      "id": "uuid",
      "title": "...",
      "description": "...",
      "price": 44.29,
      "image": "...",
      "cookingTime": 45,
      "difficulty": "medium",
      "isAvailable": true
    }
  ]
}
```

---

## 🧠 5. Очень важный момент: Dish vs Холодильник

### ❌ НЕПРАВИЛЬНО

```go
// НЕ делаем прямую зависимость от холодильника
func GetDishPrice(dishID) {
    fridge := GetFridge()
    cost := CalculateCost(fridge) // ❌ Пересчитываем каждый раз
    return cost * (1 + margin)
}
```

**Проблемы:**
- Холодильник постоянно меняется
- Цена блюда будет скакать
- Невозможно зафиксировать коммерческое предложение

### ✅ ПРАВИЛЬНО

```go
// Dish хранит СНИМОК себестоимости на момент создания
type Dish struct {
    Cost        float64  // Снимок на момент создания
    Price       float64  // Зафиксированная цена
    IsAvailable bool     // Флаг доступности (обновляется фоном)
}

// Фоновая задача проверяет доступность
func UpdateDishesAvailability() {
    dishes := GetPublishedDishes()
    
    for _, dish := range dishes {
        canCook := CheckRecipeAvailability(dish.RecipeID)
        
        if dish.IsAvailable != canCook {
            // Обновляем флаг
            UpdateDishAvailability(dish.ID, canCook)
            
            // Логируем изменение
            if !canCook {
                NotifyAdmin("Dish unavailable", dish.ID)
            }
        }
    }
}
```

**Преимущества:**
- ✅ Стабильная цена
- ✅ Асинхронная проверка доступности
- ✅ Dish = коммерческая сущность (не зависит от холодильника напрямую)

---

## 📊 6. История изменений (History Log)

### События для Dish

```go
// История событий блюда
type DishHistoryEvent struct {
    ID         uuid.UUID              `json:"id"`
    DishID     string                 `json:"dishId"`
    EventType  string                 `json:"eventType"`
    SourceType string                 `json:"sourceType"` // admin
    SourceID   *string                `json:"sourceId"`
    Metadata   map[string]interface{} `json:"metadata"`
    CreatedAt  time.Time              `json:"createdAt"`
}

// Типы событий
const (
    DishEventCreated          = "dish_created"
    DishEventUpdated          = "dish_updated"
    DishEventApproved         = "dish_approved"
    DishEventPublished        = "dish_published"
    DishEventUnpublished      = "dish_unpublished"
    DishEventPriceChanged     = "dish_price_changed"
    DishEventAvailabilityChanged = "dish_availability_changed"
)
```

### Пример логирования

```go
s.logDishEvent(adminID, dishID, "dish_price_changed", map[string]interface{}{
    "old_price": 44.29,
    "new_price": 49.99,
    "old_margin": 65.0,
    "new_margin": 70.0,
    "reason": "price_adjustment",
})
```

---

## 🎨 7. Frontend Integration

### 7.1 Админка — Кнопка генерации блюда

```typescript
// В карточке рецепта
interface Recipe {
  id: string;
  title: string;
  canCookNow: boolean; // ← Ключевой флаг
}

// UI
{recipe.canCookNow && (
  <Button onClick={() => openDishModal(recipe)}>
    🍽️ Создать карточку блюда
  </Button>
)}
```

### 7.2 Модалка генерации блюда

```typescript
interface GenerateDishModal {
  recipe: Recipe;
  cost: number;        // Рассчитывается автоматически
  margin: number;      // Слайдер 0-100%
  price: number;       // Рассчитывается: cost / (1 - margin/100)
}

// UI
<Modal>
  <h3>Создать карточку блюда</h3>
  
  <RecipePreview recipe={recipe} />
  
  <CostDisplay cost={cost} />
  
  <MarginSlider
    value={margin}
    onChange={(value) => setMargin(value)}
    min={0}
    max={100}
  />
  
  <PricePreview 
    cost={cost} 
    margin={margin}
    price={calculatePrice(cost, margin)}
  />
  
  <Button onClick={generateDish}>
    Сгенерировать блюдо
  </Button>
</Modal>
```

### 7.3 Редактирование блюда

```typescript
interface EditDishForm {
  title: string;
  description: string;
  price: number;
  margin: number;
}

// После генерации → редактирование
function DishEditor({ dish }: { dish: Dish }) {
  return (
    <Form>
      <Input
        label="Название"
        value={dish.title}
        onChange={(title) => updateDish({ title })}
      />
      
      <Textarea
        label="Описание"
        value={dish.description}
        onChange={(description) => updateDish({ description })}
      />
      
      <NumberInput
        label="Цена"
        value={dish.price}
        onChange={(price) => updateDish({ price })}
      />
      
      <MarginSlider
        value={dish.margin}
        onChange={(margin) => updateDish({ margin })}
      />
      
      <Button onClick={approveDish}>
        ✅ Утвердить блюдо
      </Button>
    </Form>
  );
}
```

### 7.4 Marketplace (клиентская часть)

```typescript
// Список доступных блюд
GET /api/marketplace/dishes?category=Main&available=true

interface DishCard {
  id: string;
  title: string;
  description: string;
  price: number;
  image: string;
  cookingTime: number;
  isAvailable: boolean;
}

// UI
{dishes.map(dish => (
  <DishCard key={dish.id}>
    <Image src={dish.image} />
    <Title>{dish.title}</Title>
    <Description>{dish.description}</Description>
    <Price>{dish.price} PLN</Price>
    
    {dish.isAvailable ? (
      <Button>Заказать</Button>
    ) : (
      <Badge>Недоступно</Badge>
    )}
  </DishCard>
))}
```

---

## 🔥 8. Преимущества этой архитектуры

### ✅ Не ломаем существующую архитектуру
- Используем проверенный Recipe pipeline
- Копируем логику генерации, статусов, редактирования
- Минимум новых абстракций

### ✅ AI используется там, где есть деньги
- Генерация привлекательного описания
- Автоматический расчёт цены
- Оптимизация маржи

### ✅ Админ всегда контролирует результат
- AI генерирует draft
- Админ редактирует и утверждает
- Публикация только после проверки

### ✅ Один рецепт → много блюд
- Можно создать несколько вариантов с разной ценой
- A/B тестирование цен
- Сезонные предложения

### ✅ Можно автоматически управлять доступностью
- Фоновая задача проверяет холодильник
- Автоматически выключает блюда при нехватке ингредиентов
- Уведомляет админа

---

## 📋 9. План реализации

### Этап 1: База данных (1 день)
- [ ] Создать миграцию для таблицы `dishes`
- [ ] Создать модель `Dish` в Go
- [ ] Добавить индексы и constraints

### Этап 2: AI Сервис (2-3 дня)
- [ ] Создать `dish_ai.go` (копия `recipe_ai.go`)
- [ ] Реализовать `GenerateDishWithAI`
- [ ] Реализовать расчёт себестоимости
- [ ] Реализовать AI промпт для описания
- [ ] Интеграция с Groq API

### Этап 3: CRUD операции (1 день)
- [ ] Создать `ApproveDish`
- [ ] Создать `PublishDish`
- [ ] Создать `UpdateDish`
- [ ] Создать `GetDishes` (список)
- [ ] Создать `GetDishByID`

### Этап 4: API Endpoints (1 день)
- [ ] `POST /api/admin/dishes/generate-from-recipe`
- [ ] `PATCH /api/admin/dishes/{id}`
- [ ] `POST /api/admin/dishes/{id}/approve`
- [ ] `POST /api/admin/dishes/{id}/publish`
- [ ] `GET /api/admin/dishes`
- [ ] `GET /api/marketplace/dishes` (публичный)

### Этап 5: History & Logging (0.5 дня)
- [ ] Добавить события `dish_*` в history
- [ ] Логирование создания, изменения, утверждения

### Этап 6: Фоновая задача доступности (1 день)
- [ ] CRON job для проверки `isAvailable`
- [ ] Обновление флага при изменении холодильника
- [ ] Уведомления админа

### Этап 7: Frontend (3-4 дня)
- [ ] Кнопка "Создать карточку блюда" в рецепте
- [ ] Модалка генерации блюда
- [ ] Форма редактирования блюда
- [ ] Список блюд в админке
- [ ] Marketplace (клиентская часть)

### Этап 8: Тестирование (2 дня)
- [ ] Unit тесты для сервиса
- [ ] Integration тесты для API
- [ ] E2E тесты для frontend

---

## ✅ Итог

### Что реализуем

1. ✅ **Модель Dish** — аналог Recipe + финансы
2. ✅ **AI генерация** — копия `recipe_ai.go`
3. ✅ **Статусы** — draft → approved → published
4. ✅ **Редактирование** — как в рецептах
5. ✅ **History** — полный аудит
6. ✅ **Availability** — фоновая проверка

### Принципы

- 🏗️ **Не изобретаем велосипед** — копируем Recipe pipeline
- 💰 **AI + деньги** — генерация описания и цены
- 👨‍💼 **Админ контролирует** — всегда проверка перед публикацией
- 📊 **Снимок стоимости** — не зависим от холодильника напрямую
- 🔄 **Асинхронная доступность** — фоновая задача

---

**Статус:** 📋 Готово к реализации  
**Архитектура:** ✅ Проверена (копия Recipe pipeline)  
**Готовность:** Backend спецификация 100%
