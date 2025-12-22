package types

import "time"

// RecommendationType - типы рекомендаций
type RecommendationType string

const (
	// TYPE 1 - URGENT
	TypePreparedExpiring RecommendationType = "prepared_expiring"
	TypeFridgeExpiring   RecommendationType = "fridge_expiring"
	
	// TYPE 2 - BUDGET
	TypeBudgetWarning  RecommendationType = "budget_warning"
	TypeBudgetOK       RecommendationType = "budget_ok"
	TypeBudgetExceeded RecommendationType = "budget_exceeded"
	
	// TYPE 3 - COOK SUGGESTION
	TypeCookSuggestion RecommendationType = "cook_suggestion"
	
	// TYPE 4 - WASTE INSIGHTS
	TypeWasteInsight    RecommendationType = "waste_insight"
	TypePortionInsight  RecommendationType = "portion_insight"
	TypeCategoryInsight RecommendationType = "category_insight"
)

// Scoring weights (для определения приоритета рекомендаций)
const (
	// Urgency weights (TYPE 1)
	UrgencyWeightExpiresSoon    = 100.0 // < 24 hours
	UrgencyWeightExpiresWarning = 70.0  // 24-48 hours
	UrgencyWeightExpiresLater   = 40.0  // 48-72 hours
	
	// Budget weights (TYPE 2)
	BudgetPressureHigh   = 90.0  // > 80% spent
	BudgetPressureMedium = 60.0  // 50-80% spent
	BudgetPressureLow    = 30.0  // < 50% spent
	
	// Cook suggestion weights (TYPE 3)
	PreferenceWeightHigh   = 40.0 // Часто готовят
	PreferenceWeightMedium = 25.0 // Иногда готовят
	PreferenceWeightLow    = 10.0 // Редко готовят
	
	// Waste weights (TYPE 4)
	WasteRiskHigh   = -30.0 // Часто выбрасывают (penalty)
	WasteRiskMedium = -15.0 // Иногда выбрасывают
	WasteRiskLow    = 0.0   // Редко выбрасывают
)

// ScoreParams - параметры для расчёта score
type ScoreParams struct {
	UrgencyWeight    float64
	BudgetPressure   float64
	PreferenceMatch  float64
	WasteRisk        float64
	IngredientMatch  float64
	FreshnessUrgency float64
}

// CalculateScore - вычисляет итоговый score рекомендации
// Formula: score = urgency + budget_pressure + preference_match + ingredient_match + freshness_urgency - waste_risk
func CalculateScore(params ScoreParams) float64 {
	score := params.UrgencyWeight +
		params.BudgetPressure +
		params.PreferenceMatch +
		params.IngredientMatch +
		params.FreshnessUrgency -
		params.WasteRisk
	
	// Clamp to 0-100
	if score < 0 {
		return 0
	}
	if score > 100 {
		return 100
	}
	return score
}

// CalculateUrgencyScore - вычисляет urgency score на основе времени до истечения
func CalculateUrgencyScore(expiresAt time.Time) float64 {
	now := time.Now()
	hoursLeft := expiresAt.Sub(now).Hours()
	
	if hoursLeft < 0 {
		return 0 // Уже истёк
	}
	if hoursLeft < 24 {
		return UrgencyWeightExpiresSoon
	}
	if hoursLeft < 48 {
		return UrgencyWeightExpiresWarning
	}
	if hoursLeft < 72 {
		return UrgencyWeightExpiresLater
	}
	return 0 // Не срочно
}

// CalculateBudgetPressure - вычисляет budget pressure на основе % использования бюджета
func CalculateBudgetPressure(spentPercentage float64) float64 {
	if spentPercentage >= 80 {
		return BudgetPressureHigh
	}
	if spentPercentage >= 50 {
		return BudgetPressureMedium
	}
	return BudgetPressureLow
}

// CalculatePreferenceScore - вычисляет preference score на основе частоты приготовления
func CalculatePreferenceScore(cookCount int) float64 {
	if cookCount >= 5 {
		return PreferenceWeightHigh
	}
	if cookCount >= 2 {
		return PreferenceWeightMedium
	}
	if cookCount >= 1 {
		return PreferenceWeightLow
	}
	return 0
}

// CalculateWasteRisk - вычисляет waste risk penalty на основе % потерь
func CalculateWasteRisk(wastePercentage float64) float64 {
	if wastePercentage >= 50 {
		return WasteRiskHigh // Penalty
	}
	if wastePercentage >= 25 {
		return WasteRiskMedium // Penalty
	}
	return WasteRiskLow // No penalty
}

// CalculateIngredientMatchScore - вычисляет ingredient match score
func CalculateIngredientMatchScore(availableCount, totalCount int) float64 {
	if totalCount == 0 {
		return 0
	}
	matchPercentage := float64(availableCount) / float64(totalCount) * 100
	return matchPercentage * 0.4 // Weight: 40% of max score
}

// CalculateFreshnessUrgency - вычисляет freshness urgency на основе скоропортящихся ингредиентов
func CalculateFreshnessUrgency(expiringIngredientsCount int) float64 {
	// Чем больше скоропортящихся ингредиентов использует рецепт, тем выше urgency
	if expiringIngredientsCount >= 3 {
		return 30.0
	}
	if expiringIngredientsCount >= 2 {
		return 20.0
	}
	if expiringIngredientsCount >= 1 {
		return 10.0
	}
	return 0
}
