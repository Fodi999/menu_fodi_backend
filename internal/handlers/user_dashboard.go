package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/database"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
	"github.com/dmitrijfomin/menu-fodifood/backend/pkg/utils"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// GetUserProfile возвращает профиль ученика
func GetUserProfile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userId"]

	var profile models.UserProfile
	if err := database.DB.Where("user_id = ?", userID).First(&profile).Error; err != nil {
		// Если профиль не найден, создаём новый
		if err.Error() == "record not found" {
			var user models.User
			if err := database.DB.Where("id = ?", userID).First(&user).Error; err != nil {
				utils.RespondError(w, http.StatusNotFound, "User not found", err.Error())
				return
			}

			parsedUserID, err := uuid.Parse(user.ID)
			if err != nil {
				utils.RespondError(w, http.StatusBadRequest, "Invalid user ID format", err.Error())
				return
			}

			profile = models.UserProfile{
				UserID:           parsedUserID,
				Name:             user.Name,
				Email:            user.Email,
				Level:            1,
				Stars:            0,
				Role:             "student",
				Language:         "pl",
				XP:               0,
				CompletedCourses: 0,
				WalletBalance:    0.00,
			}

			if err := database.DB.Create(&profile).Error; err != nil {
				utils.RespondError(w, http.StatusInternalServerError, "Failed to create profile", err.Error())
				return
			}
		} else {
			utils.RespondError(w, http.StatusInternalServerError, "Database error", err.Error())
			return
		}
	}

	utils.RespondJSON(w, http.StatusOK, profile)
}

// UpdateUserProfile обновляет профиль ученика
func UpdateUserProfile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userId"]

	var input struct {
		Name      string `json:"name"`
		AvatarURL string `json:"avatarUrl"`
		Language  string `json:"language"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.RespondError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	var profile models.UserProfile
	if err := database.DB.Where("user_id = ?", userID).First(&profile).Error; err != nil {
		utils.RespondError(w, http.StatusNotFound, "Profile not found", err.Error())
		return
	}

	// Обновляем только переданные поля
	updates := make(map[string]interface{})
	if input.Name != "" {
		updates["name"] = input.Name
	}
	if input.AvatarURL != "" {
		updates["avatar_url"] = input.AvatarURL
	}
	if input.Language != "" {
		updates["language"] = input.Language
	}

	if err := database.DB.Model(&profile).Updates(updates).Error; err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to update profile", err.Error())
		return
	}

	utils.RespondJSON(w, http.StatusOK, profile)
}

// GetUserProgress возвращает прогресс по курсам
func GetUserProgress(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userId"]

	var progress []models.UserProgress
	if err := database.DB.Where("user_id = ?", userID).Order("last_accessed_at DESC").Find(&progress).Error; err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Database error", err.Error())
		return
	}

	utils.RespondJSON(w, http.StatusOK, progress)
}

// GetUserCertificates возвращает сертификаты ученика
func GetUserCertificates(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userId"]

	var certificates []models.Certificate
	if err := database.DB.Where("user_id = ?", userID).Order("issued_at DESC").Find(&certificates).Error; err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Database error", err.Error())
		return
	}

	utils.RespondJSON(w, http.StatusOK, certificates)
}

// GetUserRecipes возвращает личные рецепты ученика
func GetUserRecipes(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userId"]

	var recipes []models.PersonalRecipe
	if err := database.DB.Where("user_id = ?", userID).Order("created_at DESC").Find(&recipes).Error; err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Database error", err.Error())
		return
	}

	utils.RespondJSON(w, http.StatusOK, recipes)
}

// CreateUserRecipe создаёт новый личный рецепт
func CreateUserRecipe(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userId"]

	var input models.PersonalRecipe
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.RespondError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Валидация
	if input.Title == "" || len(input.Ingredients) == 0 || len(input.Steps) == 0 {
		utils.RespondError(w, http.StatusBadRequest, "Missing required fields", "title, ingredients, and steps are required")
		return
	}

	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, "Invalid user ID", err.Error())
		return
	}

	input.UserID = parsedUserID
	input.CreatedAt = time.Now()
	input.UpdatedAt = time.Now()

	if err := database.DB.Create(&input).Error; err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to create recipe", err.Error())
		return
	}

	utils.RespondJSON(w, http.StatusCreated, input)
}

// DeleteUserRecipe удаляет личный рецепт
func DeleteUserRecipe(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	recipeID := vars["recipeId"]

	if err := database.DB.Where("id = ?", recipeID).Delete(&models.PersonalRecipe{}).Error; err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to delete recipe", err.Error())
		return
	}

	utils.RespondJSON(w, http.StatusOK, map[string]string{"message": "Recipe deleted successfully"})
}

// GetUserWallet возвращает баланс ChefToken и историю транзакций
func GetUserWallet(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userId"]

	// Получаем профиль для баланса
	var profile models.UserProfile
	if err := database.DB.Where("user_id = ?", userID).First(&profile).Error; err != nil {
		// Якщо профіль не знайдено, створюємо його автоматично
		userUUID, parseErr := uuid.Parse(userID)
		if parseErr != nil {
			utils.RespondError(w, http.StatusBadRequest, "Invalid user ID", parseErr.Error())
			return
		}

		// Отримуємо базову інформацію про користувача
		var user models.User
		if userErr := database.DB.Where("id = ?", userID).First(&user).Error; userErr != nil {
			utils.RespondError(w, http.StatusNotFound, "User not found", userErr.Error())
			return
		}

		// Створюємо новий профіль з початковим балансом 0
		profile = models.UserProfile{
			UserID:        userUUID,
			Name:          user.Name,
			Email:         user.Email,
			WalletBalance: 0.00,
			Level:         1,
			Role:          "student",
			Language:      "en",
		}

		if createErr := database.DB.Create(&profile).Error; createErr != nil {
			utils.RespondError(w, http.StatusInternalServerError, "Failed to create profile", createErr.Error())
			return
		}
	}

	// Получаем последние транзакции
	var transactions []models.WalletTransaction
	database.DB.Where("user_id = ?", userID).Order("created_at DESC").Limit(50).Find(&transactions)

	response := map[string]interface{}{
		"balance":      profile.WalletBalance,
		"transactions": transactions,
	}

	utils.RespondJSON(w, http.StatusOK, response)
}

// GetMarketPurchases возвращает купленные рецепты
func GetMarketPurchases(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userId"]

	var purchases []models.MarketPurchase
	if err := database.DB.Where("buyer_id = ?", userID).Order("created_at DESC").Find(&purchases).Error; err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Database error", err.Error())
		return
	}

	// Получаем детали рецептов
	recipeIDs := make([]uuid.UUID, len(purchases))
	for i, p := range purchases {
		recipeIDs[i] = p.RecipeID
	}

	var recipes []models.PersonalRecipe
	if len(recipeIDs) > 0 {
		database.DB.Where("id IN ?", recipeIDs).Find(&recipes)
	}

	response := map[string]interface{}{
		"purchases": purchases,
		"recipes":   recipes,
	}

	utils.RespondJSON(w, http.StatusOK, response)
}

// GetUserDashboard возвращает полную панель управления ученика
func GetUserDashboard(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userId"]

	// Получаем профиль
	var profile models.UserProfile
	if err := database.DB.Where("user_id = ?", userID).First(&profile).Error; err != nil {
		utils.RespondError(w, http.StatusNotFound, "Profile not found", err.Error())
		return
	}

	// Рассчитываем прогресс до следующего уровня
	nextLevelXP := profile.Level * 500 // каждый уровень = 500 XP
	progressToNextLevel := 0.0
	if nextLevelXP > 0 {
		progressToNextLevel = (float64(profile.XP) / float64(nextLevelXP)) * 100
		if progressToNextLevel > 100 {
			progressToNextLevel = 100
		}
	}

	// Получаем общую статистику курсов
	var totalCourses int64
	database.DB.Model(&models.Course{}).Where("is_published = ?", true).Count(&totalCourses)

	// Получаем прогресс по курсам
	var courseProgress []models.UserProgress
	database.DB.Where("user_id = ?", userID).Order("last_accessed_at DESC").Limit(5).Find(&courseProgress)

	// Получаем последние активности (тесты)
	var recentQuizzes []models.UserQuiz
	database.DB.Where("user_id = ?", userID).Order("completed_at DESC").Limit(5).Find(&recentQuizzes)

	// Получаем информацию о курсах для активностей
	courseIDs := make([]uuid.UUID, 0)
	for _, q := range recentQuizzes {
		courseIDs = append(courseIDs, q.CourseID)
	}
	var courses []models.Course
	if len(courseIDs) > 0 {
		database.DB.Where("id IN ?", courseIDs).Find(&courses)
	}

	// Формируем активности
	recentActivity := make([]map[string]interface{}, 0)
	for _, quiz := range recentQuizzes {
		var courseName string
		for _, c := range courses {
			if c.ID == quiz.CourseID {
				courseName = c.Title
				break
			}
		}
		recentActivity = append(recentActivity, map[string]interface{}{
			"type":      "quiz_completed",
			"course":    courseName,
			"stars":     quiz.StarsEarned,
			"score":     quiz.Score,
			"timestamp": quiz.CompletedAt,
		})
	}

	// Рекомендации курсов (на основе языка и уровня)
	var recommendations []models.Course
	database.DB.Where("is_published = ? AND language = ? AND level <= ?", true, profile.Language, profile.Level+2).
		Order("level ASC").
		Limit(3).
		Find(&recommendations)

	// Формируем рекомендации с процентом совпадения
	recommendationsList := make([]map[string]interface{}, 0)
	for _, rec := range recommendations {
		matchPercentage := 100
		if rec.Level > profile.Level {
			matchPercentage = 85
		}
		recommendationsList = append(recommendationsList, map[string]interface{}{
			"courseId":    rec.ID,
			"title":       rec.Title,
			"description": rec.Description,
			"level":       rec.Level,
			"match":       matchPercentage,
			"imageUrl":    rec.ImageURL,
		})
	}

	// Получаем последние транзакции
	var recentTransactions []models.WalletTransaction
	database.DB.Where("user_id = ?", userID).Order("created_at DESC").Limit(5).Find(&recentTransactions)

	// 🍱 Kitchen Simulation: активные рецепты (последние 3)
	var activeRecipes []models.PersonalRecipe
	database.DB.Where("user_id = ?", userID).Order("updated_at DESC").Limit(3).Find(&activeRecipes)

	activeRecipesList := make([]map[string]interface{}, 0)
	for _, recipe := range activeRecipes {
		progress := 100 // завершён
		if recipe.Rating == 0 {
			progress = 75 // не оценён AI
		}
		activeRecipesList = append(activeRecipesList, map[string]interface{}{
			"id":              recipe.ID,
			"title":           recipe.Title,
			"progress":        progress,
			"ingredientsUsed": len(recipe.Ingredients),
			"difficulty":      recipe.Difficulty,
			"imageUrl":        recipe.ImageURL,
		})
	}

	// 🧠 AI Mentor Feed: генерируем случайные советы на основе языка
	mentorTips := []string{}
	switch profile.Language {
	case "pl":
		mentorTips = []string{
			"Spróbuj użyć noża Yanagiba dla idealnych plasterków sashimi.",
			"Ryż sushi powinien mieć temperaturę pokojową - nie za gorący!",
			"Pamiętaj: świeża ryba to podstawa doskonałego sushi.",
		}
	case "ua":
		mentorTips = []string{
			"Спробуйте використовувати ніж Yanagiba для ідеальних скибочок сашімі.",
			"Рис для суші повинен бути кімнатної температури - не занадто гарячий!",
			"Пам'ятайте: свіжа риба - основа ідеального суші.",
		}
	default:
		mentorTips = []string{
			"Try using a Yanagiba knife for perfect sashimi slices.",
			"Sushi rice should be room temperature - not too hot!",
			"Remember: fresh fish is the foundation of excellent sushi.",
		}
	}

	// 🛍️ Marketplace: предложения купить рецепты других (топ 3 по рейтингу)
	var marketSuggestions []models.PersonalRecipe
	database.DB.Where("is_public = ? AND user_id != ?", true, userID).
		Order("rating DESC, purchases DESC").
		Limit(3).
		Find(&marketSuggestions)

	marketList := make([]map[string]interface{}, 0)
	for _, recipe := range marketSuggestions {
		// Получаем имя шефа
		var chefProfile models.UserProfile
		chefName := "Unknown Chef"
		if err := database.DB.Where("user_id = ?", recipe.UserID).First(&chefProfile).Error; err == nil {
			chefName = chefProfile.Name
		}

		marketList = append(marketList, map[string]interface{}{
			"recipeId":  recipe.ID,
			"title":     recipe.Title,
			"price":     recipe.Price,
			"chef":      chefName,
			"rating":    recipe.Rating,
			"purchases": recipe.Purchases,
			"imageUrl":  recipe.ImageURL,
			"category":  recipe.Category,
		})
	}

	// 🎓 Achievements: получаем badges пользователя
	var userAchievements []models.UserAchievement
	database.DB.Where("user_id = ?", userID).Find(&userAchievements)

	achievementIDs := make([]uuid.UUID, 0)
	for _, ua := range userAchievements {
		achievementIDs = append(achievementIDs, ua.AchievementID)
	}

	var achievements []models.Achievement
	badges := make([]map[string]interface{}, 0)
	if len(achievementIDs) > 0 {
		database.DB.Where("id IN ?", achievementIDs).Find(&achievements)
		for _, ach := range achievements {
			badges = append(badges, map[string]interface{}{
				"code":        ach.Code,
				"title":       ach.Title,
				"description": ach.Description,
				"iconUrl":     ach.IconURL,
				"category":    ach.Category,
			})
		}
	}

	// 🧾 Certificates: получаем сертификаты
	var certificates []models.Certificate
	database.DB.Where("user_id = ?", userID).Order("issued_at DESC").Find(&certificates)

	certificateList := make([]map[string]interface{}, 0)
	for _, cert := range certificates {
		certificateList = append(certificateList, map[string]interface{}{
			"id":         cert.ID,
			"courseName": cert.CourseName,
			"level":      cert.Level,
			"stars":      cert.Stars,
			"pdfUrl":     cert.PDFURL,
			"issuedAt":   cert.IssuedAt,
		})
	}

	// ⚙️ AI Stats: определяем ранг шефа на основе уровня и звёзд
	chefRank := "Novice Cook"
	switch profile.Language {
	case "pl":
		if profile.Level >= 8 {
			chefRank = "Mistrz Sushi"
		} else if profile.Level >= 5 {
			chefRank = "Profesjonalny Chef"
		} else if profile.Level >= 3 {
			chefRank = "Zaawansowany Uczeń"
		} else if profile.Level >= 1 {
			chefRank = "Początkujący Kucharz"
		}
	case "ua":
		if profile.Level >= 8 {
			chefRank = "Майстер Суші"
		} else if profile.Level >= 5 {
			chefRank = "Професійний Шеф"
		} else if profile.Level >= 3 {
			chefRank = "Досвідчений Учень"
		} else if profile.Level >= 1 {
			chefRank = "Початківець Кухар"
		}
	default:
		if profile.Level >= 8 {
			chefRank = "Sushi Master"
		} else if profile.Level >= 5 {
			chefRank = "Professional Chef"
		} else if profile.Level >= 3 {
			chefRank = "Advanced Student"
		} else if profile.Level >= 1 {
			chefRank = "Novice Cook"
		}
	}

	// Итоговый response
	response := map[string]interface{}{
		"profile": map[string]interface{}{
			"name":                profile.Name,
			"email":               profile.Email,
			"avatarUrl":           profile.AvatarURL,
			"level":               profile.Level,
			"xp":                  profile.XP,
			"nextLevelXP":         nextLevelXP,
			"progressToNextLevel": progressToNextLevel,
			"stars":               profile.Stars,
			"role":                profile.Role,
			"language":            profile.Language,
			"chefRank":            chefRank, // ⚙️ AI Stats
		},
		"stats": map[string]interface{}{
			"completedCourses": profile.CompletedCourses,
			"totalCourses":     totalCourses,
			"stars":            profile.Stars,
			"walletBalance":    profile.WalletBalance,
			"totalRecipes":     0, // добавим позже
		},
		"courseProgress":     courseProgress,
		"recentActivity":     recentActivity,
		"recommendations":    recommendationsList,
		"recentTransactions": recentTransactions,
		// 🍱 Kitchen Simulation
		"activeRecipes": activeRecipesList,
		// 🧠 AI Mentor Feed
		"mentorTips": mentorTips,
		// 🛍️ Marketplace
		"marketSuggestions": marketList,
		// 🎓 Achievements
		"badges": badges,
		// 🧾 Certificates
		"certificates": certificateList,
	}

	// Получаем количество рецептов
	var recipeCount int64
	database.DB.Model(&models.PersonalRecipe{}).Where("user_id = ?", userID).Count(&recipeCount)
	response["stats"].(map[string]interface{})["totalRecipes"] = recipeCount

	utils.RespondJSON(w, http.StatusOK, response)
}
