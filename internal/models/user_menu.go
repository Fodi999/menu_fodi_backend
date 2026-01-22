package models

import (
	"time"

	"github.com/google/uuid"
)

// ============================================================================
// Kitchen Pipeline: User Menu Items
// Purpose: Track recipes user wants to cook TODAY (single source of truth)
// ============================================================================

// MenuItemStatus - статусы в процессе готовки
type MenuItemStatus string

const (
	MenuItemPlanned   MenuItemStatus = "planned"   // Добавлено в меню (хочу приготовить)
	MenuItemCooking   MenuItemStatus = "cooking"   // Готовим прямо сейчас
	MenuItemCompleted MenuItemStatus = "completed" // Приготовлено
	MenuItemCancelled MenuItemStatus = "cancelled" // Отменено
)

// UserMenuItem - рецепт в меню пользователя (на сегодня)
type UserMenuItem struct {
	ID       uuid.UUID      `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	UserID   string         `gorm:"type:text;not null" json:"user_id"` // TEXT to match "User".id
	RecipeID uuid.UUID      `gorm:"type:uuid;not null;column:recipe_id" json:"recipe_id"`
	Servings int            `gorm:"not null;default:1" json:"servings"`
	Status   MenuItemStatus `gorm:"type:text;not null;default:'planned'" json:"status"`
	
	// Dates
	PlannedFor        time.Time  `gorm:"type:date;not null;default:CURRENT_DATE" json:"planned_for"`
	CreatedAt         time.Time  `gorm:"autoCreateTime" json:"created_at"`
	StartedCookingAt  *time.Time `gorm:"type:timestamp" json:"started_cooking_at,omitempty"`
	CompletedAt       *time.Time `gorm:"type:timestamp" json:"completed_at,omitempty"`
	
	// Optional metadata
	Notes *string `gorm:"type:text" json:"notes,omitempty"`
	
	// Relations (не хранятся в БД, загружаются через Preload)
	Recipe *RecipeCatalog `gorm:"foreignKey:RecipeID" json:"recipe,omitempty"`
}

// TableName - имя таблицы для GORM
func (UserMenuItem) TableName() string {
	return "user_menu_items"
}

// ============================================================================
// DTO для API responses
// ============================================================================

// MenuItemResponse - ответ API для одного пункта меню
type MenuItemResponse struct {
	ID        string         `json:"id"`
	Servings  int            `json:"servings"`
	Status    MenuItemStatus `json:"status"`
	PlannedFor string        `json:"planned_for"` // YYYY-MM-DD format
	CreatedAt  string         `json:"created_at"`
	Notes      *string        `json:"notes,omitempty"`
	
	// Recipe details (full object, not just ID)
	Recipe RecipeBasicInfo `json:"recipe"`
	
	// Timestamps
	StartedCookingAt *string `json:"started_cooking_at,omitempty"`
	CompletedAt      *string `json:"completed_at,omitempty"`
}

// RecipeBasicInfo - базовая информация о рецепте для меню
type RecipeBasicInfo struct {
	ID            string  `json:"id"`
	Title         string  `json:"title"`
	CanonicalName string  `json:"canonical_name"`
	ImageURL      *string `json:"image_url"`
	CookTime      int     `json:"cook_time"` // minutes
	Servings      int     `json:"servings"`  // default servings
}

// ============================================================================
// Request DTOs
// ============================================================================

// AddToMenuRequest - запрос на добавление рецепта в меню
type AddToMenuRequest struct {
	RecipeID string `json:"recipe_id" binding:"required"` // UUID or canonical_name
	Servings int    `json:"servings,omitempty"`          // Default: 1
	Notes    string `json:"notes,omitempty"`
}

// UpdateMenuItemRequest - обновление параметров
type UpdateMenuItemRequest struct {
	Servings *int    `json:"servings,omitempty"`
	Notes    *string `json:"notes,omitempty"`
}

// StartCookingRequest - начать готовить
type StartCookingRequest struct {
	Servings int `json:"servings,omitempty"` // Can adjust before cooking
}

// CompleteCookingRequest - завершить готовку
type CompleteCookingRequest struct {
	ActualServings int `json:"actual_servings,omitempty"` // What was actually cooked
}

// ============================================================================
// API Contract Examples
// ============================================================================
//
// POST /api/menu/today
// {
//   "recipe_id": "605c8419-2d42-4ef0-a9d2-839582e98727",
//   "servings": 2,
//   "notes": "Extra spicy"
// }
//
// GET /api/menu/today
// [
//   {
//     "id": "...",
//     "servings": 2,
//     "status": "planned",
//     "recipe": {
//       "id": "...",
//       "title": "Жареные яйца",
//       "cook_time": 7
//     }
//   }
// ]
//
// POST /api/menu/{id}/start
// {
//   "servings": 3
// }
//
// POST /api/menu/{id}/complete
// {
//   "actual_servings": 2
// }
//
// ============================================================================
