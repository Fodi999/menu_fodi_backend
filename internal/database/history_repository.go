package database

import (
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
)

type HistoryRepository struct {
	db *gorm.DB
}

func NewHistoryRepository(db *gorm.DB) *HistoryRepository {
	return &HistoryRepository{db: db}
}

// HistoryFilters for querying history events
type HistoryFilters struct {
	EventType  string     // Filter by event type
	SourceType string     // Filter by source type
	StartDate  *time.Time // From date
	EndDate    *time.Time // To date
	Limit      int        // Max results
}

// Create creates a new history event
func (r *HistoryRepository) Create(event *models.HistoryEvent) error {
	result := r.db.Create(event)
	if result.Error != nil {
		return fmt.Errorf("failed to create history event: %w", result.Error)
	}
	return nil
}

// CreateWithMetadata creates event with structured metadata
func (r *HistoryRepository) CreateWithMetadata(
	userID string,
	eventType models.HistoryEventType,
	sourceType models.HistorySourceType,
	sourceID *string,
	portions *int,
	metadata map[string]interface{},
) error {
	// Marshal metadata to JSON
	var metadataJSON datatypes.JSON
	if metadata != nil {
		jsonBytes, err := json.Marshal(metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata: %w", err)
		}
		metadataJSON = jsonBytes
	}

	event := &models.HistoryEvent{
		UserID:     userID,
		EventType:  eventType,
		SourceType: sourceType,
		SourceID:   sourceID,
		Portions:   portions,
		Metadata:   metadataJSON,
		CreatedAt:  time.Now(),
	}

	return r.Create(event)
}

// GetByUserID returns all history events for a user
func (r *HistoryRepository) GetByUserID(userID string, limit int) ([]models.HistoryEvent, error) {
	var events []models.HistoryEvent

	query := r.db.Where("user_id = ?", userID).Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&events).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get history events: %w", err)
	}

	return events, nil
}

// GetByFilters returns history events with filters
func (r *HistoryRepository) GetByFilters(userID string, filters HistoryFilters) ([]models.HistoryEvent, error) {
	var events []models.HistoryEvent

	query := r.db.Where("user_id = ?", userID)

	// Apply filters
	if filters.EventType != "" {
		query = query.Where("event_type = ?", filters.EventType)
	}

	if filters.SourceType != "" {
		query = query.Where("source_type = ?", filters.SourceType)
	}

	if filters.StartDate != nil {
		query = query.Where("created_at >= ?", filters.StartDate)
	}

	if filters.EndDate != nil {
		query = query.Where("created_at <= ?", filters.EndDate)
	}

	query = query.Order("created_at DESC")

	if filters.Limit > 0 {
		query = query.Limit(filters.Limit)
	} else {
		query = query.Limit(100) // Default limit
	}

	err := query.Find(&events).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get filtered history events: %w", err)
	}

	return events, nil
}

// GetStatsByUser returns analytics statistics for user's history
func (r *HistoryRepository) GetStatsByUser(userID string, startDate, endDate *time.Time) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Build base query
	query := r.db.Model(&models.HistoryEvent{}).Where("user_id = ?", userID)

	if startDate != nil {
		query = query.Where("created_at >= ?", startDate)
	}
	if endDate != nil {
		query = query.Where("created_at <= ?", endDate)
	}

	// Count by event type
	var eventCounts []struct {
		EventType string
		Count     int64
	}
	err := query.
		Select("event_type, COUNT(*) as count").
		Group("event_type").
		Scan(&eventCounts).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get event counts: %w", err)
	}

	// Convert to map
	eventTypeMap := make(map[string]int64)
	for _, ec := range eventCounts {
		eventTypeMap[ec.EventType] = ec.Count
	}
	stats["event_counts"] = eventTypeMap

	// Sum portions consumed
	var totalPortionsConsumed int64
	err = query.
		Where("event_type = ? AND portions IS NOT NULL", models.EventTypeConsume).
		Select("COALESCE(SUM(portions), 0)").
		Scan(&totalPortionsConsumed).Error
	if err != nil {
		return nil, fmt.Errorf("failed to sum consumed portions: %w", err)
	}
	stats["total_portions_consumed"] = totalPortionsConsumed

	// Sum portions cooked
	var totalPortionsCooked int64
	err = query.
		Where("event_type = ? AND portions IS NOT NULL", models.EventTypeCook).
		Select("COALESCE(SUM(portions), 0)").
		Scan(&totalPortionsCooked).Error
	if err != nil {
		return nil, fmt.Errorf("failed to sum cooked portions: %w", err)
	}
	stats["total_portions_cooked"] = totalPortionsCooked

	// Total events count
	var totalEvents int64
	query.Count(&totalEvents)
	stats["total_events"] = totalEvents

	return stats, nil
}

// GetRecentActivity returns recent history events for dashboard
func (r *HistoryRepository) GetRecentActivity(userID string, limit int) ([]models.HistoryEvent, error) {
	if limit <= 0 {
		limit = 10
	}

	var events []models.HistoryEvent
	err := r.db.
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Find(&events).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get recent activity: %w", err)
	}

	return events, nil
}
