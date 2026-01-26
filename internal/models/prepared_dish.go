package models

import (
	"time"
)

// PreparedDish represents a dish prepared by user from a recipe
// Category is NOT stored here - accessed via JOIN to Recipe table (single source of truth)
type PreparedDish struct {
	ID                string     `gorm:"column:id;type:uuid;primaryKey" json:"id"`
	UserID            string     `gorm:"column:user_id;type:text;not null" json:"user_id"`
	RecipeID          string     `gorm:"column:recipe_id;type:uuid;not null" json:"recipe_id"`
	PortionsAvailable int        `gorm:"column:portions_available;not null;default:0" json:"portions_available"`
	PortionsInitial   int        `gorm:"column:portions_initial;not null" json:"portions_initial"`
	PreparedAt        time.Time  `gorm:"column:prepared_at;not null;default:NOW()" json:"prepared_at"`
	ExpiresAt         *time.Time `gorm:"column:expires_at" json:"expires_at,omitempty"`
	Source            string     `gorm:"column:source;not null;default:'cook'" json:"source"` // 'cook' or 'manual'
	CostPerPortion    *float64   `gorm:"column:cost_per_portion;type:decimal(10,2)" json:"cost_per_portion,omitempty"`
	TotalCost         *float64   `gorm:"column:total_cost;type:decimal(10,2)" json:"total_cost,omitempty"`
	CreatedAt         time.Time  `gorm:"column:created_at;not null;default:NOW()" json:"created_at"`
	UpdatedAt         time.Time  `gorm:"column:updated_at;not null;default:NOW()" json:"updated_at"`

	// Relation to Recipe (loaded via JOIN for category access)
	Recipe *RecipeCatalog `gorm:"-" json:"recipe,omitempty"`
}

// TableName specifies the database table name
func (PreparedDish) TableName() string {
	return "prepared_dishes"
}

// IsAvailable checks if dish has portions left
func (d *PreparedDish) IsAvailable() bool {
	return d.PortionsAvailable > 0
}

// IsExpired checks if dish is past expiration date
func (d *PreparedDish) IsExpired() bool {
	if d.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*d.ExpiresAt)
}

// ConsumedPortions calculates how many portions were consumed
func (d *PreparedDish) ConsumedPortions() int {
	return d.PortionsInitial - d.PortionsAvailable
}

// PreparedDishStatus represents the status of a prepared dish
type PreparedDishStatus string

const (
	PreparedDishStatusAvailable PreparedDishStatus = "available"
	PreparedDishStatusFinished  PreparedDishStatus = "finished"
	PreparedDishStatusExpired   PreparedDishStatus = "expired"
)

// GetStatus returns the computed status of the dish
func (d *PreparedDish) GetStatus() PreparedDishStatus {
	// Check expired first
	if d.IsExpired() {
		return PreparedDishStatusExpired
	}

	// Check if finished (no portions left)
	if d.PortionsAvailable == 0 {
		return PreparedDishStatusFinished
	}

	// Otherwise available
	return PreparedDishStatusAvailable
}
