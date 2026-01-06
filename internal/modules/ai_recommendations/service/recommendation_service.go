package service

import (
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/ai_recommendations/dto"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/ai_recommendations/types"
	"gorm.io/gorm"
)

// RecommendationService - детерминистический AI engine для рекомендаций
type RecommendationService struct {
	db *gorm.DB
}

// NewRecommendationService - конструктор
func NewRecommendationService(db *gorm.DB) *RecommendationService {
	return &RecommendationService{
		db: db,
	}
}

// GetRecommendations - главный метод: собирает все типы рекомендаций
func (s *RecommendationService) GetRecommendations(userID string) (*dto.RecommendationsData, error) {
	data := &dto.RecommendationsData{
		Urgent:   []dto.UrgentRecommendation{},
		Budget:   []dto.BudgetRecommendation{},
		Cook:     []dto.CookSuggestion{},
		Insights: []dto.WasteInsight{},
	}

	// TYPE 1 - URGENT (срочные действия)
	urgent, err := s.AnalyzeUrgent(userID)
	if err != nil {
		log.Printf("[AI] Error analyzing urgent: %v", err)
	} else {
		data.Urgent = urgent
	}

	// TYPE 2 - BUDGET (бюджетные предупреждения)
	budget, err := s.AnalyzeBudget(userID)
	if err != nil {
		log.Printf("[AI] Error analyzing budget: %v", err)
	} else {
		data.Budget = budget
	}

	// TYPE 3 - COOK SUGGESTIONS (что приготовить)
	cook, err := s.AnalyzeCookSuggestions(userID)
	if err != nil {
		log.Printf("[AI] Error analyzing cook suggestions: %v", err)
	} else {
		data.Cook = cook
	}

	// TYPE 4 - WASTE INSIGHTS (аналитика потерь)
	insights, err := s.AnalyzeWasteInsights(userID)
	if err != nil {
		log.Printf("[AI] Error analyzing waste insights: %v", err)
	} else {
		data.Insights = insights
	}

	return data, nil
}

// AnalyzeUrgent - TYPE 1: проверяет prepared_dishes и fridge на истечение срока
func (s *RecommendationService) AnalyzeUrgent(userID string) ([]dto.UrgentRecommendation, error) {
	recommendations := []dto.UrgentRecommendation{}
	now := time.Now()
	urgentThreshold := now.Add(72 * time.Hour) // 72 hours ahead

	// 1. Проверяем prepared_dishes (готовые блюда)
	type PreparedDishResult struct {
		ID            string
		RecipeID      string
		LocalName     string
		ExpiresAt     time.Time
		PortionsAvail int
	}

	var dishes []PreparedDishResult
	err := s.db.Raw(`
		SELECT pd.id, pd.recipe_id, 
		       COALESCE(r.local_name, r.canonical_name) as local_name,
		       pd.expires_at, pd.portions_available
		FROM prepared_dishes pd
		LEFT JOIN recipes r ON r.id::text = pd.recipe_id
		WHERE pd.user_id::text = ?
		  AND pd.expires_at IS NOT NULL
		  AND pd.expires_at <= ?
		  AND pd.portions_available > 0
		ORDER BY pd.expires_at ASC
	`, userID, urgentThreshold).Scan(&dishes).Error

	if err != nil {
		return nil, fmt.Errorf("failed to query prepared dishes: %w", err)
	}

	for _, dish := range dishes {
		hoursLeft := int(dish.ExpiresAt.Sub(now).Hours())
		if hoursLeft < 0 {
			hoursLeft = 0
		}

		message := fmt.Sprintf("Zjedz dziś %s — wygasa za %d godz", dish.LocalName, hoursLeft)
		if hoursLeft < 24 {
			message = fmt.Sprintf("⚠️ Zjedz DZIŚ %s — wygasa jutro!", dish.LocalName)
		}

		urgencyScore := types.CalculateUrgencyScore(dish.ExpiresAt)

		recommendations = append(recommendations, dto.UrgentRecommendation{
			Type:      string(types.TypePreparedExpiring),
			Message:   message,
			DishID:    &dish.ID,
			DishName:  &dish.LocalName,
			ExpiresAt: dish.ExpiresAt.Format(time.RFC3339),
			HoursLeft: hoursLeft,
			Score:     urgencyScore,
		})
	}

	// 2. Проверяем fridge (скоропортящиеся ингредиенты)
	type FridgeItemResult struct {
		ID         string
		Name       string
		ExpiryDate *time.Time
		Quantity   float64
		Unit       string
	}

	var fridgeItems []FridgeItemResult
	err = s.db.Raw(`
		SELECT fi.id, ic.name, fi.expiry_date, fi.quantity, fi.unit
		FROM user_fridge_items fi
		LEFT JOIN ingredients_catalog ic ON ic.id = fi.ingredient_id
		WHERE fi.user_id::text = ?
		  AND fi.expiry_date IS NOT NULL
		  AND fi.expiry_date <= ?
		  AND fi.quantity > 0
		ORDER BY fi.expiry_date ASC
	`, userID, urgentThreshold).Scan(&fridgeItems).Error

	if err != nil {
		return nil, fmt.Errorf("failed to query fridge items: %w", err)
	}

	for _, item := range fridgeItems {
		if item.ExpiryDate == nil {
			continue
		}

		hoursLeft := int(item.ExpiryDate.Sub(now).Hours())
		if hoursLeft < 0 {
			hoursLeft = 0
		}

		message := fmt.Sprintf("Użyj %s — wygasa za %d godz", item.Name, hoursLeft)
		if hoursLeft < 24 {
			message = fmt.Sprintf("⚠️ Użyj DZIŚ %s — wygasa jutro!", item.Name)
		}

		urgencyScore := types.CalculateUrgencyScore(*item.ExpiryDate)

		recommendations = append(recommendations, dto.UrgentRecommendation{
			Type:      string(types.TypeFridgeExpiring),
			Message:   message,
			ItemID:    &item.ID,
			ItemName:  &item.Name,
			ExpiresAt: item.ExpiryDate.Format(time.RFC3339),
			HoursLeft: hoursLeft,
			Score:     urgencyScore,
		})
	}

	// Сортируем по score (самые срочные — первые)
	sort.Slice(recommendations, func(i, j int) bool {
		return recommendations[i].Score > recommendations[j].Score
	})

	// Лимит: top 5
	if len(recommendations) > 5 {
		recommendations = recommendations[:5]
	}

	return recommendations, nil
}

// AnalyzeBudget - TYPE 2: анализ бюджета и прогноз расходов
func (s *RecommendationService) AnalyzeBudget(userID string) ([]dto.BudgetRecommendation, error) {
	recommendations := []dto.BudgetRecommendation{}
	now := time.Now()

	// Получаем текущую неделю бюджета
	type BudgetResult struct {
		WeekStart     time.Time
		PlannedBudget float64
		SpentBudget   float64
		WasteCost     float64
	}

	var budget BudgetResult
	err := s.db.Raw(`
		SELECT week_start, planned_budget, spent_budget, waste_cost
		FROM weekly_budgets
		WHERE user_id::text = ?
		  AND week_start = (
		      SELECT DATE_TRUNC('week', ?::date) + INTERVAL '0 days'
		  )
		LIMIT 1
	`, userID, now).Scan(&budget).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to query budget: %w", err)
	}

	// Если бюджет не установлен, не показываем рекомендации
	if budget.PlannedBudget == 0 {
		return recommendations, nil
	}

	remaining := budget.PlannedBudget - budget.SpentBudget
	spentPercentage := (budget.SpentBudget / budget.PlannedBudget) * 100

	// Вычисляем, сколько дней до конца недели
	weekStart := budget.WeekStart
	if weekStart.IsZero() {
		// Если budget не найден, считаем от понедельника текущей недели
		weekStart = now.AddDate(0, 0, -int(now.Weekday())+1)
	}
	weekEnd := weekStart.AddDate(0, 0, 7)
	daysLeft := int(weekEnd.Sub(now).Hours() / 24)
	if daysLeft < 0 {
		daysLeft = 0
	}

	// Средний расход в день (за прошедшие дни недели)
	daysPassed := 7 - daysLeft
	if daysPassed <= 0 {
		daysPassed = 1
	}
	avgDailySpend := budget.SpentBudget / float64(daysPassed)

	// Прогноз расхода до конца недели
	projectedSpend := budget.SpentBudget + (avgDailySpend * float64(daysLeft))

	// Определяем, укладываемся ли в бюджет
	isOnTrack := projectedSpend <= budget.PlannedBudget

	// Формируем рекомендацию
	var message string
	var recType types.RecommendationType

	if budget.SpentBudget > budget.PlannedBudget {
		recType = types.TypeBudgetExceeded
		message = fmt.Sprintf("❌ Przekroczyłeś budżet o %.2f PLN", budget.SpentBudget-budget.PlannedBudget)
	} else if spentPercentage >= 80 {
		recType = types.TypeBudgetWarning
		message = fmt.Sprintf("⚠️ Pozostało tylko %.2f PLN na ten tydzień (%d dni)", remaining, daysLeft)
	} else if spentPercentage >= 50 {
		recType = types.TypeBudgetWarning
		message = fmt.Sprintf("💡 Pozostało %.2f PLN — gotuj oszczędnie", remaining)
	} else {
		recType = types.TypeBudgetOK
		message = fmt.Sprintf("✅ Budżet OK — pozostało %.2f PLN", remaining)
	}

	budgetPressure := types.CalculateBudgetPressure(spentPercentage)

	recommendations = append(recommendations, dto.BudgetRecommendation{
		Type:           string(recType),
		Message:        message,
		Remaining:      remaining,
		DaysLeft:       daysLeft,
		AvgDailySpend:  avgDailySpend,
		ProjectedSpend: projectedSpend,
		IsOnTrack:      isOnTrack,
		Score:          budgetPressure,
	})

	return recommendations, nil
}

// AnalyzeCookSuggestions - TYPE 3: что приготовить (matching рецептов с fridge)
func (s *RecommendationService) AnalyzeCookSuggestions(userID string) ([]dto.CookSuggestion, error) {
	suggestions := []dto.CookSuggestion{}

	// Получаем saved_recipes пользователя
	type SavedRecipeResult struct {
		RecipeID      string
		CanonicalName string
		LocalName     string
		Category      string
	}

	var savedRecipes []SavedRecipeResult
	err := s.db.Raw(`
		SELECT r.id as recipe_id, r.canonical_name, r.local_name, r.category
		FROM saved_recipes sr
		JOIN recipes r ON r.id = sr.recipe_id
		WHERE sr.user_id::text = ?
		LIMIT 20
	`, userID).Scan(&savedRecipes).Error

	if err != nil {
		return nil, fmt.Errorf("failed to query saved recipes: %w", err)
	}

	// Для каждого рецепта считаем ingredient match
	for _, recipe := range savedRecipes {
		// TODO: Implement ingredient matching logic
		// Для MVP можно упростить и просто показать топ-3 рецепта из saved

		// Получаем preference score из истории (сколько раз готовили)
		var cookCount int
		s.db.Raw(`
			SELECT COUNT(*)
			FROM history_events
			WHERE user_id::text = ?
			  AND event_type = 'cook'
			  AND source_id = ?
		`, userID, recipe.RecipeID).Scan(&cookCount)

		preferenceScore := types.CalculatePreferenceScore(cookCount)

		// Для MVP: простая эвристика
		ingredientMatch := 80.0  // Заглушка (в полной версии — реальный расчёт)
		freshnessUrgency := 15.0 // Заглушка

		scoreParams := types.ScoreParams{
			PreferenceMatch:  preferenceScore,
			IngredientMatch:  ingredientMatch,
			FreshnessUrgency: freshnessUrgency,
		}

		totalScore := types.CalculateScore(scoreParams)

		recipeName := recipe.LocalName
		if recipeName == "" {
			recipeName = recipe.CanonicalName
		}

		suggestions = append(suggestions, dto.CookSuggestion{
			Type:             string(types.TypeCookSuggestion),
			RecipeID:         recipe.RecipeID,
			RecipeName:       recipeName,
			Category:         recipe.Category,
			Reason:           "Masz zapisany ten przepis",
			IngredientMatch:  ingredientMatch,
			PreferenceScore:  preferenceScore,
			FreshnessUrgency: freshnessUrgency,
			MissingCount:     0,
			Score:            totalScore,
		})
	}

	// Сортируем по score
	sort.Slice(suggestions, func(i, j int) bool {
		return suggestions[i].Score > suggestions[j].Score
	})

	// Лимит: top 5
	if len(suggestions) > 5 {
		suggestions = suggestions[:5]
	}

	return suggestions, nil
}

// AnalyzeWasteInsights - TYPE 4: аналитика паттернов потерь
func (s *RecommendationService) AnalyzeWasteInsights(userID string) ([]dto.WasteInsight, error) {
	insights := []dto.WasteInsight{}

	// 1. Общая статистика waste за последние 30 дней
	type WasteStatsResult struct {
		TotalWaste  float64
		TotalCooked int
		TotalWasted int
		WasteRate   float64
	}

	var stats WasteStatsResult
	err := s.db.Raw(`
		WITH waste_events AS (
			SELECT 
				COALESCE((metadata->>'waste_cost')::float, 0) as waste_cost
			FROM history_events
			WHERE user_id::text = ?
			  AND event_type = 'waste'
			  AND created_at >= NOW() - INTERVAL '30 days'
		),
		cook_events AS (
			SELECT COUNT(*) as total_cooked
			FROM history_events
			WHERE user_id::text = ?
			  AND event_type = 'cook'
			  AND created_at >= NOW() - INTERVAL '30 days'
		),
		waste_count AS (
			SELECT COUNT(*) as total_wasted
			FROM history_events
			WHERE user_id::text = ?
			  AND event_type = 'waste'
			  AND created_at >= NOW() - INTERVAL '30 days'
		)
		SELECT 
			COALESCE(SUM(w.waste_cost), 0) as total_waste,
			COALESCE(c.total_cooked, 0) as total_cooked,
			COALESCE(wc.total_wasted, 0) as total_wasted,
			CASE 
				WHEN c.total_cooked > 0 THEN (wc.total_wasted::float / c.total_cooked * 100)
				ELSE 0
			END as waste_rate
		FROM waste_events w
		CROSS JOIN cook_events c
		CROSS JOIN waste_count wc
	`, userID, userID, userID).Scan(&stats).Error

	if err != nil {
		return nil, fmt.Errorf("failed to query waste stats: %w", err)
	}

	// Если есть значимый waste, добавляем инсайт
	if stats.TotalWasted > 0 {
		message := fmt.Sprintf("Zmarnowałeś %d dań (%.0f%%) w ostatnim miesiącu",
			stats.TotalWasted, stats.WasteRate)

		suggestion := "Gotuj mniejsze porcje lub częściej spożywaj resztki"
		if stats.WasteRate >= 30 {
			suggestion = "⚠️ Wysoki poziom marnotrawstwa — rozważ mniejsze porcje"
		}

		wasteRisk := types.CalculateWasteRisk(stats.WasteRate)
		score := 50.0 + wasteRisk // Higher waste = higher insight importance

		insights = append(insights, dto.WasteInsight{
			Type:       string(types.TypeWasteInsight),
			Message:    message,
			WasteRate:  stats.WasteRate,
			TotalWaste: stats.TotalWaste,
			Period:     "last_month",
			Suggestion: suggestion,
			Score:      score,
		})
	}

	// 2. Waste по категориям (какие блюда чаще выбрасывают)
	type CategoryWasteResult struct {
		Category   string
		WasteCount int
		TotalCost  float64
	}

	var categoryWaste []CategoryWasteResult
	err = s.db.Raw(`
		SELECT 
			COALESCE(r.category, 'unknown') as category,
			COUNT(*) as waste_count,
			COALESCE(SUM((he.metadata->>'waste_cost')::float), 0) as total_cost
		FROM history_events he
		LEFT JOIN prepared_dishes pd ON pd.id::text = he.source_id
		LEFT JOIN recipes r ON r.id::text = pd.recipe_id
		WHERE he.user_id::text = ?
		  AND he.event_type = 'waste'
		  AND he.created_at >= NOW() - INTERVAL '30 days'
		GROUP BY r.category
		ORDER BY waste_count DESC
		LIMIT 3
	`, userID).Scan(&categoryWaste).Error

	if err != nil {
		log.Printf("[AI] Error querying category waste: %v", err)
	}

	for _, cw := range categoryWaste {
		if cw.WasteCount > 0 {
			message := fmt.Sprintf("Najczęściej marnujesz dania z kategorii '%s' (%d razy)",
				cw.Category, cw.WasteCount)

			insights = append(insights, dto.WasteInsight{
				Type:       string(types.TypeCategoryInsight),
				Message:    message,
				Category:   &cw.Category,
				WasteRate:  0, // TODO: Calculate category-specific rate
				TotalWaste: cw.TotalCost,
				Period:     "last_month",
				Suggestion: fmt.Sprintf("Rozważ mniejsze porcje dla kategorii '%s'", cw.Category),
				Score:      40.0,
			})
		}
	}

	// Сортируем по score
	sort.Slice(insights, func(i, j int) bool {
		return insights[i].Score > insights[j].Score
	})

	// Лимит: top 3
	if len(insights) > 3 {
		insights = insights[:3]
	}

	return insights, nil
}
