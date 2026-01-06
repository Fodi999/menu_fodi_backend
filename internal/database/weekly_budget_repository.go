package database

import (
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
)

type WeeklyBudgetRepository struct {
	db *gorm.DB
}

func NewWeeklyBudgetRepository(db *gorm.DB) *WeeklyBudgetRepository {
	return &WeeklyBudgetRepository{db: db}
}

// GetMondayOfWeek returns the Monday of the week for a given date
func GetMondayOfWeek(t time.Time) time.Time {
	// Get weekday (0 = Sunday, 1 = Monday, ...)
	weekday := int(t.Weekday())

	// Calculate days to subtract to get to Monday
	daysToSubtract := weekday - 1
	if weekday == 0 { // Sunday
		daysToSubtract = 6
	}

	monday := t.AddDate(0, 0, -daysToSubtract)
	// Return date at midnight
	return time.Date(monday.Year(), monday.Month(), monday.Day(), 0, 0, 0, 0, monday.Location())
}

// GetOrCreateCurrentWeek gets or creates the budget for current week
func (r *WeeklyBudgetRepository) GetOrCreateCurrentWeek(userID string) (*models.WeeklyBudget, error) {
	weekStart := GetMondayOfWeek(time.Now())

	return r.GetOrCreateForWeek(userID, weekStart)
}

// GetOrCreateForWeek gets or creates budget for specific week
func (r *WeeklyBudgetRepository) GetOrCreateForWeek(userID string, weekStart time.Time) (*models.WeeklyBudget, error) {
	weekStart = GetMondayOfWeek(weekStart) // Normalize to Monday

	var budget models.WeeklyBudget

	err := r.db.Where("user_id = ? AND week_start = ?", userID, weekStart).First(&budget).Error

	if err == gorm.ErrRecordNotFound {
		// Create new budget
		budget = models.WeeklyBudget{
			UserID:        userID,
			WeekStart:     weekStart,
			PlannedBudget: 0,
			SpentBudget:   0,
			WasteCost:     0,
		}

		err = r.db.Create(&budget).Error
		if err != nil {
			return nil, fmt.Errorf("failed to create weekly budget: %w", err)
		}

		return &budget, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get weekly budget: %w", err)
	}

	return &budget, nil
}

// UpdateSpentBudget increments spent_budget (called on consume)
func (r *WeeklyBudgetRepository) UpdateSpentBudget(userID string, amount float64) error {
	budget, err := r.GetOrCreateCurrentWeek(userID)
	if err != nil {
		return err
	}

	return r.db.Model(&models.WeeklyBudget{}).
		Where("id = ?", budget.ID).
		Updates(map[string]interface{}{
			"spent_budget": gorm.Expr("spent_budget + ?", amount),
			"updated_at":   time.Now(),
		}).Error
}

// UpdateWasteCost increments waste_cost (called on waste)
func (r *WeeklyBudgetRepository) UpdateWasteCost(userID string, amount float64) error {
	budget, err := r.GetOrCreateCurrentWeek(userID)
	if err != nil {
		return err
	}

	return r.db.Model(&models.WeeklyBudget{}).
		Where("id = ?", budget.ID).
		Updates(map[string]interface{}{
			"waste_cost": gorm.Expr("waste_cost + ?", amount),
			"updated_at": time.Now(),
		}).Error
}

// SetPlannedBudget sets the planned budget for current week
func (r *WeeklyBudgetRepository) SetPlannedBudget(userID string, amount float64) error {
	budget, err := r.GetOrCreateCurrentWeek(userID)
	if err != nil {
		return err
	}

	return r.db.Model(&models.WeeklyBudget{}).
		Where("id = ?", budget.ID).
		Updates(map[string]interface{}{
			"planned_budget": amount,
			"updated_at":     time.Now(),
		}).Error
}

// GetByUserAndWeek gets budget for specific week
func (r *WeeklyBudgetRepository) GetByUserAndWeek(userID string, weekStart time.Time) (*models.WeeklyBudget, error) {
	weekStart = GetMondayOfWeek(weekStart)

	var budget models.WeeklyBudget
	err := r.db.Where("user_id = ? AND week_start = ?", userID, weekStart).First(&budget).Error

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get weekly budget: %w", err)
	}

	return &budget, nil
}

// GetRecentWeeks gets budgets for last N weeks
func (r *WeeklyBudgetRepository) GetRecentWeeks(userID string, weeks int) ([]models.WeeklyBudget, error) {
	if weeks <= 0 {
		weeks = 4
	}

	var budgets []models.WeeklyBudget

	err := r.db.
		Where("user_id = ?", userID).
		Order("week_start DESC").
		Limit(weeks).
		Find(&budgets).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get recent budgets: %w", err)
	}

	return budgets, nil
}

// GetWeeklyStats calculates aggregate stats across weeks
func (r *WeeklyBudgetRepository) GetWeeklyStats(userID string) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Get current week
	currentWeek, err := r.GetOrCreateCurrentWeek(userID)
	if err != nil {
		return nil, err
	}

	stats["current_week"] = currentWeek

	// Get last 4 weeks for trends
	recentBudgets, err := r.GetRecentWeeks(userID, 4)
	if err != nil {
		return nil, err
	}

	stats["recent_weeks"] = recentBudgets

	// Calculate averages
	if len(recentBudgets) > 0 {
		totalSpent := 0.0
		totalWaste := 0.0
		totalPlanned := 0.0

		for _, b := range recentBudgets {
			totalSpent += b.SpentBudget
			totalWaste += b.WasteCost
			totalPlanned += b.PlannedBudget
		}

		count := float64(len(recentBudgets))
		stats["avg_spent"] = totalSpent / count
		stats["avg_waste"] = totalWaste / count
		stats["avg_planned"] = totalPlanned / count
		stats["avg_efficiency"] = currentWeek.GetEfficiencyScore() // Current week efficiency
	}

	return stats, nil
}
