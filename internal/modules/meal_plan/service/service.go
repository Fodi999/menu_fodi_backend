package service

import (
	"fmt"
	"log"
	"strings"

	ai_core "github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/ai_core"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/database"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/meal_plan/dto"
	nutritionservice "github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/nutrition/service"
)

// MealPlanService handles business logic for meal plan generation
type MealPlanService struct {
	nutritionService *nutritionservice.NutritionService
	aiClient         *ai_core.GroqClient
}

// NewMealPlanService creates a new service instance
func NewMealPlanService() *MealPlanService {
	return &MealPlanService{
		nutritionService: nutritionservice.NewNutritionService(),
		aiClient:         ai_core.NewGroqClient(),
	}
}

// GenerateMealPlan generates a meal plan using AI
func (s *MealPlanService) GenerateMealPlan(req *dto.MealPlanRequest, userID string) (*dto.MealPlanResponse, error) {
	// Validate inputs
	if req.Language == "" {
		req.Language = "ua"
	}
	if req.TargetCalories == 0 {
		req.TargetCalories = 2000
	}
	if req.Days == 0 || req.Days > 14 {
		req.Days = 7
	}

	// Get fridge items if useFridge is true
	var fridgeItems []models.UserFridge
	if req.UseFridge && userID != "" {
		db := database.GetDB()
		if err := db.Where("user_id = ? AND available = ?", userID, true).Find(&fridgeItems).Error; err != nil {
			log.Printf("Error fetching fridge items: %v", err)
		}
	}

	// Build AI prompt
	prompt := s.buildMealPlanPrompt(req, fridgeItems)

	// Call Groq AI
	messages := []ai_core.GroqMessage{
		{Role: "system", Content: "You are a professional nutritionist and meal planning expert."},
		{Role: "user", Content: prompt},
	}

	aiResult, err := s.aiClient.Chat(messages, 0.7, 2000)
	if err != nil {
		log.Printf("Error generating meal plan: %v", err)
		return nil, fmt.Errorf("failed to generate meal plan: %w", err)
	}

	if len(aiResult.Choices) == 0 {
		return nil, fmt.Errorf("no meal plan generated")
	}

	aiResponse := aiResult.Choices[0].Message.Content

	// Parse AI response into meal plan
	plan := s.parseMealPlanResponse(aiResponse, req.Days, req.Language)

	// Calculate nutrition for each meal
	totalCalories := 0.0
	for i, day := range plan {
		// Estimate calories for each meal
		breakfastCal := s.estimateMealCalories(day.Breakfast)
		lunchCal := s.estimateMealCalories(day.Lunch)
		dinnerCal := s.estimateMealCalories(day.Dinner)
		snackCal := 0.0
		if day.Snack != "" {
			snackCal = s.estimateMealCalories(day.Snack)
		}

		dayTotal := breakfastCal + lunchCal + dinnerCal + snackCal
		plan[i].TotalCalories = dayTotal
		totalCalories += dayTotal
	}

	avgPerDay := totalCalories / float64(len(plan))

	return &dto.MealPlanResponse{
		Plan:          plan,
		TotalCalories: totalCalories,
		AvgPerDay:     avgPerDay,
		Success:       true,
	}, nil
}

// buildMealPlanPrompt creates the AI prompt for meal plan generation
func (s *MealPlanService) buildMealPlanPrompt(req *dto.MealPlanRequest, fridgeItems []models.UserFridge) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Create a %d-day meal plan with approximately %d calories per day.\n\n", req.Days, req.TargetCalories))

	sb.WriteString("IMPORTANT FORMAT RULES:\n")
	sb.WriteString("- Use EXACTLY this format for each day:\n")
	sb.WriteString("Day 1:\n")
	sb.WriteString("Breakfast: [Simple meal name in Ukrainian]\n")
	sb.WriteString("Lunch: [Simple meal name in Ukrainian]\n")
	sb.WriteString("Dinner: [Simple meal name in Ukrainian]\n")
	sb.WriteString("Snack: [Optional snack name]\n\n")

	sb.WriteString("Requirements:\n")
	sb.WriteString("- Each meal should have a SHORT, clear Ukrainian name (3-5 words max)\n")
	sb.WriteString("- Distribute calories: breakfast 25%, lunch 35%, dinner 30%, snack 10%\n")
	sb.WriteString("- NO REPEATS: every meal should be different\n")
	sb.WriteString("- Balanced: proteins, carbs, healthy fats, vegetables\n")
	sb.WriteString("- Popular Ukrainian/European dishes\n\n")

	if len(fridgeItems) > 0 {
		sb.WriteString("Available ingredients in fridge:\n")
		for _, item := range fridgeItems {
			sb.WriteString(fmt.Sprintf("- %s (%.1f %s)\n", item.Product, item.Quantity, item.Unit))
		}
		sb.WriteString("\nTry to use these ingredients when creating meals.\n\n")
	}

	sb.WriteString("Example format:\n")
	sb.WriteString("Day 1:\n")
	sb.WriteString("Breakfast: Омлет з помідорами\n")
	sb.WriteString("Lunch: Борщ український\n")
	sb.WriteString("Dinner: Курка на грилі з овочами\n")
	sb.WriteString("Snack: Йогурт з фруктами\n\n")

	sb.WriteString("Now create the meal plan following this EXACT format:\n")

	return sb.String()
}

// parseMealPlanResponse parses AI response into structured meal plan
func (s *MealPlanService) parseMealPlanResponse(aiResponse string, days int, language string) []dto.DayMeal {
	plan := make([]dto.DayMeal, 0, days)

	// Split by days
	lines := strings.Split(aiResponse, "\n")

	dayNames := s.getDayNames(language)
	currentDay := dto.DayMeal{}
	dayIndex := 0

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Check if new day starts
		if strings.HasPrefix(strings.ToLower(line), "day ") ||
			strings.Contains(strings.ToLower(line), "день ") ||
			s.isDayName(line, dayNames) {
			if currentDay.Day != "" {
				plan = append(plan, currentDay)
			}
			if dayIndex < len(dayNames) {
				currentDay = dto.DayMeal{Day: dayNames[dayIndex]}
				dayIndex++
			}
			continue
		}

		// Parse meal types
		if strings.Contains(strings.ToLower(line), "breakfast") ||
			strings.Contains(strings.ToLower(line), "сніданок") ||
			strings.Contains(strings.ToLower(line), "завтрак") {
			parts := strings.Split(line, ":")
			if len(parts) > 1 {
				currentDay.Breakfast = strings.TrimSpace(parts[1])
			}
		} else if strings.Contains(strings.ToLower(line), "lunch") ||
			strings.Contains(strings.ToLower(line), "обід") ||
			strings.Contains(strings.ToLower(line), "обед") {
			parts := strings.Split(line, ":")
			if len(parts) > 1 {
				currentDay.Lunch = strings.TrimSpace(parts[1])
			}
		} else if strings.Contains(strings.ToLower(line), "dinner") ||
			strings.Contains(strings.ToLower(line), "вечеря") ||
			strings.Contains(strings.ToLower(line), "ужин") {
			parts := strings.Split(line, ":")
			if len(parts) > 1 {
				currentDay.Dinner = strings.TrimSpace(parts[1])
			}
		} else if strings.Contains(strings.ToLower(line), "snack") ||
			strings.Contains(strings.ToLower(line), "перекус") {
			parts := strings.Split(line, ":")
			if len(parts) > 1 {
				currentDay.Snack = strings.TrimSpace(parts[1])
			}
		}
	}

	// Add last day
	if currentDay.Day != "" {
		plan = append(plan, currentDay)
	}

	// Fill with generic days if not enough
	for len(plan) < days {
		dayIndex := len(plan)
		if dayIndex < len(dayNames) {
			plan = append(plan, dto.DayMeal{
				Day:       dayNames[dayIndex],
				Breakfast: "Вівсянка з фруктами",
				Lunch:     "Курка з овочами",
				Dinner:    "Риба на грилі з салатом",
			})
		}
	}

	return plan[:days]
}

// getDayNames returns localized day names
func (s *MealPlanService) getDayNames(language string) []string {
	switch language {
	case "ua":
		return []string{"Понеділок", "Вівторок", "Середа", "Четвер", "П'ятниця", "Субота", "Неділя",
			"Понеділок (тиждень 2)", "Вівторок (тиждень 2)", "Середа (тиждень 2)", "Четвер (тиждень 2)",
			"П'ятниця (тиждень 2)", "Субота (тиждень 2)", "Неділя (тиждень 2)"}
	case "en":
		return []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday",
			"Monday (week 2)", "Tuesday (week 2)", "Wednesday (week 2)", "Thursday (week 2)",
			"Friday (week 2)", "Saturday (week 2)", "Sunday (week 2)"}
	case "ru":
		return []string{"Понедельник", "Вторник", "Среда", "Четверг", "Пятница", "Суббота", "Воскресенье",
			"Понедельник (неделя 2)", "Вторник (неделя 2)", "Среда (неделя 2)", "Четверг (неделя 2)",
			"Пятница (неделя 2)", "Суббота (неделя 2)", "Воскресенье (неделя 2)"}
	case "pl":
		return []string{"Poniedziałek", "Wtorek", "Środa", "Czwartek", "Piątek", "Sobota", "Niedziela",
			"Poniedziałek (tydzień 2)", "Wtorek (tydzień 2)", "Środa (tydzień 2)", "Czwartek (tydzień 2)",
			"Piątek (tydzień 2)", "Sobota (tydzień 2)", "Niedziela (tydzień 2)"}
	default:
		return []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"}
	}
}

// isDayName checks if line is a day name
func (s *MealPlanService) isDayName(line string, dayNames []string) bool {
	lower := strings.ToLower(line)
	for _, day := range dayNames {
		if strings.Contains(lower, strings.ToLower(day)) {
			return true
		}
	}
	return false
}

// estimateMealCalories estimates calories for a meal by name
func (s *MealPlanService) estimateMealCalories(mealName string) float64 {
	mealName = strings.ToLower(mealName)

	// Simple calorie estimation based on meal types
	calorieMap := map[string]float64{
		// Breakfast
		"омлет":        300,
		"каша":         250,
		"вівсянка":     280,
		"сирники":      350,
		"млинці":       400,
		"яєчня":        250,
		"мюслі":        300,
		"тості":        200,

		// Lunch
		"борщ":         400,
		"суп":          300,
		"курка":        450,
		"м'ясо":        500,
		"риба":         350,
		"стейк":        600,
		"котлета":      450,
		"плов":         550,
		"паста":        500,
		"макарони":     480,
		"рис":          400,

		// Dinner
		"салат":        200,
		"овочі":        150,
		"гриль":        400,
		"запіканка":    450,
		"піца":         650,
		"бургер":       700,

		// Snacks
		"фрукти":       100,
		"йогурт":       150,
		"горіхи":       200,
		"сир":          120,
		"батончик":    180,
	}

	// Check for keywords in meal name
	totalCal := 0.0
	matches := 0
	for keyword, cal := range calorieMap {
		if strings.Contains(mealName, keyword) {
			totalCal += cal
			matches++
		}
	}

	if matches > 0 {
		return totalCal / float64(matches)
	}

	// Default estimates by meal type inference
	if strings.Contains(mealName, "breakfast") || strings.Contains(mealName, "сніданок") {
		return 350
	} else if strings.Contains(mealName, "lunch") || strings.Contains(mealName, "обід") {
		return 500
	} else if strings.Contains(mealName, "dinner") || strings.Contains(mealName, "вечеря") {
		return 450
	} else if strings.Contains(mealName, "snack") || strings.Contains(mealName, "перекус") {
		return 150
	}

	return 400 // Default
}
