package models

import (
	"time"
)

// WeeklyBudget represents user's weekly food budget tracking
type WeeklyBudget struct {
	ID            string    `gorm:"column:id;type:uuid;primaryKey" json:"id"`
	UserID        string    `gorm:"column:user_id;type:text;not null" json:"user_id"`
	WeekStart     time.Time `gorm:"column:week_start;type:date;not null" json:"week_start"` // Monday of the week
	PlannedBudget float64   `gorm:"column:planned_budget;type:decimal(10,2);not null;default:0" json:"planned_budget"`
	SpentBudget   float64   `gorm:"column:spent_budget;type:decimal(10,2);not null;default:0" json:"spent_budget"`
	WasteCost     float64   `gorm:"column:waste_cost;type:decimal(10,2);not null;default:0" json:"waste_cost"`
	CreatedAt     time.Time `gorm:"column:created_at;not null;default:NOW()" json:"created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at;not null;default:NOW()" json:"updated_at"`
}

// TableName specifies the database table name
func (WeeklyBudget) TableName() string {
	return "weekly_budgets"
}

// GetRemainingBudget calculates how much budget is left
func (b *WeeklyBudget) GetRemainingBudget() float64 {
	return b.PlannedBudget - b.SpentBudget
}

// GetWastePercentage calculates waste as percentage of planned budget
func (b *WeeklyBudget) GetWastePercentage() float64 {
	if b.PlannedBudget == 0 {
		return 0
	}
	return (b.WasteCost / b.PlannedBudget) * 100
}

// GetSpentPercentage calculates spent as percentage of planned budget
func (b *WeeklyBudget) GetSpentPercentage() float64 {
	if b.PlannedBudget == 0 {
		return 0
	}
	return (b.SpentBudget / b.PlannedBudget) * 100
}

// IsOverBudget checks if user exceeded their planned budget
func (b *WeeklyBudget) IsOverBudget() bool {
	return b.SpentBudget > b.PlannedBudget
}

// GetSavedMoney calculates money saved (remaining budget)
func (b *WeeklyBudget) GetSavedMoney() float64 {
	remaining := b.GetRemainingBudget()
	if remaining < 0 {
		return 0 // No savings if over budget
	}
	return remaining
}

// GetTotalCost calculates total cost (spent + waste)
func (b *WeeklyBudget) GetTotalCost() float64 {
	return b.SpentBudget + b.WasteCost
}

// GetEfficiencyScore calculates budget efficiency (0-100)
// 100 = perfect (no waste, under budget)
// Lower = worse (high waste or over budget)
func (b *WeeklyBudget) GetEfficiencyScore() float64 {
	if b.PlannedBudget == 0 {
		return 100
	}

	// Penalty for going over budget
	budgetPenalty := 0.0
	if b.IsOverBudget() {
		overAmount := b.SpentBudget - b.PlannedBudget
		budgetPenalty = (overAmount / b.PlannedBudget) * 50 // Max 50% penalty
	}

	// Penalty for waste
	wastePenalty := b.GetWastePercentage() / 2 // Max 50% penalty

	score := 100 - budgetPenalty - wastePenalty
	if score < 0 {
		return 0
	}
	return score
}

// GetWeekNumber returns ISO week number
func (b *WeeklyBudget) GetWeekNumber() int {
	_, week := b.WeekStart.ISOWeek()
	return week
}
