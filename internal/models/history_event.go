package models

import (
	"time"

	"gorm.io/datatypes"
)

// HistoryEventType represents the type of event
type HistoryEventType string

const (
	EventTypeCook         HistoryEventType = "cook"
	EventTypeConsume      HistoryEventType = "consume"
	EventTypeWaste        HistoryEventType = "waste"
	EventTypeManual       HistoryEventType = "manual"
	EventTypeFridgeAdd    HistoryEventType = "fridge_add"
	EventTypeFridgeRemove HistoryEventType = "fridge_remove"
	// EventTypeExpired is an alias for EventTypeWaste (automatic expiry cleanup)
	EventTypeExpired HistoryEventType = "waste" // Uses 'waste' enum value for expired items
)

// HistorySourceType represents what triggered the event
type HistorySourceType string

const (
	SourceTypePreparedDish HistorySourceType = "prepared_dish"
	SourceTypeRecipe       HistorySourceType = "recipe"
	SourceTypeFridge       HistorySourceType = "fridge"
	SourceTypeManual       HistorySourceType = "manual"
	// SourceTypeAuto is an alias for SourceTypeFridge (automatic system actions)
	SourceTypeAuto HistorySourceType = "fridge" // Uses 'fridge' enum value for auto actions
)

// HistoryEvent represents a user action in the system
type HistoryEvent struct {
	ID         string            `gorm:"column:id;type:uuid;primaryKey" json:"id"`
	UserID     string            `gorm:"column:user_id;type:text;not null" json:"user_id"`
	EventType  HistoryEventType  `gorm:"column:event_type;type:history_event_type;not null" json:"event_type"`
	SourceType HistorySourceType `gorm:"column:source_type;type:history_source_type;not null" json:"source_type"`
	SourceID   *string           `gorm:"column:source_id;type:text" json:"source_id,omitempty"`
	Portions   *int              `gorm:"column:portions" json:"portions,omitempty"`
	Metadata   datatypes.JSON    `gorm:"column:metadata;type:jsonb" json:"metadata,omitempty"`
	CreatedAt  time.Time         `gorm:"column:created_at;not null;default:NOW()" json:"created_at"`
}

// TableName specifies the database table name
func (HistoryEvent) TableName() string {
	return "history_events"
}

// Helper methods for event type checks
func (e *HistoryEvent) IsCook() bool {
	return e.EventType == EventTypeCook
}

func (e *HistoryEvent) IsConsume() bool {
	return e.EventType == EventTypeConsume
}

func (e *HistoryEvent) IsWaste() bool {
	return e.EventType == EventTypeWaste
}

func (e *HistoryEvent) IsFridgeAction() bool {
	return e.EventType == EventTypeFridgeAdd || e.EventType == EventTypeFridgeRemove
}

// Helper methods for source type checks
func (e *HistoryEvent) FromPreparedDish() bool {
	return e.SourceType == SourceTypePreparedDish
}

func (e *HistoryEvent) FromRecipe() bool {
	return e.SourceType == SourceTypeRecipe
}

// ExpiredItemMetadata represents metadata for expired fridge items
type ExpiredItemMetadata struct {
	IngredientID   string  `json:"ingredient_id"`
	IngredientName string  `json:"ingredient_name"`
	Quantity       float64 `json:"quantity"`
	Unit           string  `json:"unit"`
	Cost           float64 `json:"cost"`           // Total cost (quantity * pricePerUnit)
	PricePerUnit   float64 `json:"price_per_unit"` // Price per unit at time of expiry
	Currency       string  `json:"currency"`       // PLN, EUR, USD
	ExpiryDate     string  `json:"expiry_date"`    // ISO 8601 date
	ArrivedAt      string  `json:"arrived_at"`     // When item was added to fridge
	DaysInFridge   int     `json:"days_in_fridge"` // How long item was stored
	Reason         string  `json:"reason"`         // "expiry_date_passed"
	Context        string  `json:"context"`        // "fridge_cleanup"
}
