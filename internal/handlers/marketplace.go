package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/database"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
	"github.com/dmitrijfomin/menu-fodifood/backend/pkg/utils"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

// GetMarketRecipesHandler GET /api/market/recipes - все публичные рецепты
func GetMarketRecipesHandler(w http.ResponseWriter, r *http.Request) {
	// Параметры фильтрации
	category := r.URL.Query().Get("category")
	difficulty := r.URL.Query().Get("difficulty")
	maxPriceStr := r.URL.Query().Get("maxPrice")
	minRatingStr := r.URL.Query().Get("minRating")
	sortBy := r.URL.Query().Get("sortBy") // "popular", "newest", "rating", "price"
	
	query := database.DB.Where("is_public = ?", true)

	// Фильтры
	if category != "" {
		query = query.Where("category = ?", category)
	}
	if difficulty != "" {
		query = query.Where("difficulty = ?", difficulty)
	}
	if maxPriceStr != "" {
		maxPrice, _ := strconv.ParseFloat(maxPriceStr, 64)
		query = query.Where("price <= ?", maxPrice)
	}
	if minRatingStr != "" {
		minRating, _ := strconv.ParseFloat(minRatingStr, 64)
		query = query.Where("rating >= ?", minRating)
	}

	// Сортировка
	switch sortBy {
	case "popular":
		query = query.Order("purchases DESC")
	case "newest":
		query = query.Order("created_at DESC")
	case "rating":
		query = query.Order("rating DESC")
	case "price":
		query = query.Order("price ASC")
	default:
		query = query.Order("created_at DESC")
	}

	var recipes []models.PersonalRecipe
	if err := query.Limit(50).Find(&recipes).Error; err != nil {
		utils.RespondWithJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"status":  "error",
			"message": "Failed to fetch recipes",
		})
		return
	}

	// Добавляем информацию об авторе
	type RecipeWithAuthor struct {
		models.PersonalRecipe
		AuthorName   string  `json:"authorName"`
		AuthorLevel  int     `json:"authorLevel"`
		AuthorAvatar string  `json:"authorAvatar"`
		ReviewCount  int     `json:"reviewCount"`
		AvgReview    float64 `json:"avgReview"`
	}

	var enrichedRecipes []RecipeWithAuthor
	for _, recipe := range recipes {
		var profile models.UserProfile
		database.DB.Where("user_id = ?", recipe.UserID).First(&profile)

		// Подсчёт отзывов
		var reviewCount int64
		var avgReview float64
		database.DB.Model(&models.RecipeReview{}).Where("recipe_id = ?", recipe.ID).Count(&reviewCount)
		database.DB.Model(&models.RecipeReview{}).Where("recipe_id = ?", recipe.ID).Select("AVG(rating)").Row().Scan(&avgReview)

		enrichedRecipes = append(enrichedRecipes, RecipeWithAuthor{
			PersonalRecipe: recipe,
			AuthorName:     profile.Name,
			AuthorLevel:    profile.Level,
			AuthorAvatar:   profile.AvatarURL,
			ReviewCount:    int(reviewCount),
			AvgReview:      avgReview,
		})
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"status": "ok",
		"data": map[string]interface{}{
			"recipes": enrichedRecipes,
			"total":   len(enrichedRecipes),
		},
	})
}

// PurchaseRecipeHandler POST /api/market/purchase - покупка рецепта
func PurchaseRecipeHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RecipeID string `json:"recipeId"`
		BuyerID  string `json:"buyerId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithJSON(w, http.StatusBadRequest, map[string]interface{}{
			"status":  "error",
			"message": "Invalid request body",
		})
		return
	}

	// Проверяем рецепт
	var recipe models.PersonalRecipe
	if err := database.DB.Where("id = ?", req.RecipeID).First(&recipe).Error; err != nil {
		utils.RespondWithJSON(w, http.StatusNotFound, map[string]interface{}{
			"status":  "error",
			"message": "Recipe not found",
		})
		return
	}

	// Проверка: нельзя купить свой рецепт
	if recipe.UserID.String() == req.BuyerID {
		utils.RespondWithJSON(w, http.StatusBadRequest, map[string]interface{}{
			"status":  "error",
			"message": "Cannot purchase your own recipe",
		})
		return
	}

	// Проверка: уже куплен?
	var existingPurchase models.RecipePurchase
	if err := database.DB.Where("recipe_id = ? AND buyer_id = ?", req.RecipeID, req.BuyerID).First(&existingPurchase).Error; err == nil {
		utils.RespondWithJSON(w, http.StatusBadRequest, map[string]interface{}{
			"status":  "error",
			"message": "You already own this recipe",
		})
		return
	}

	// Получаем профили покупателя и продавца
	var buyerProfile, sellerProfile models.UserProfile
	if err := database.DB.Where("user_id = ?", req.BuyerID).First(&buyerProfile).Error; err != nil {
		utils.RespondWithJSON(w, http.StatusNotFound, map[string]interface{}{
			"status":  "error",
			"message": "Buyer profile not found",
		})
		return
	}
	database.DB.Where("user_id = ?", recipe.UserID).First(&sellerProfile)

	// Проверка баланса
	if buyerProfile.WalletBalance < recipe.Price {
		utils.RespondWithJSON(w, http.StatusBadRequest, map[string]interface{}{
			"status":  "error",
			"message": "Insufficient ChefToken balance",
			"required": recipe.Price,
			"current":  buyerProfile.WalletBalance,
		})
		return
	}

	// Транзакция покупки
	tx := database.DB.Begin()

	// 1. Снимаем деньги с покупателя
	buyerProfile.WalletBalance -= recipe.Price
	if err := tx.Save(&buyerProfile).Error; err != nil {
		tx.Rollback()
		utils.RespondWithJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"status":  "error",
			"message": "Failed to deduct from buyer wallet",
		})
		return
	}

	// 2. Комиссия платформы 10%
	commission := recipe.Price * 0.10
	netAmount := recipe.Price - commission

	// 3. Начисляем продавцу 90%
	sellerProfile.WalletBalance += netAmount
	if err := tx.Save(&sellerProfile).Error; err != nil {
		tx.Rollback()
		utils.RespondWithJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"status":  "error",
			"message": "Failed to credit seller wallet",
		})
		return
	}

	// 4. Создаём запись покупки
	purchase := models.RecipePurchase{
		RecipeID:   req.RecipeID,
		BuyerID:    req.BuyerID,
		SellerID:   recipe.UserID.String(),
		Price:      recipe.Price,
		Commission: commission,
		NetAmount:  netAmount,
	}
	if err := tx.Create(&purchase).Error; err != nil {
		tx.Rollback()
		utils.RespondWithJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"status":  "error",
			"message": "Failed to create purchase record",
		})
		return
	}

	// 5. Обновляем счётчик покупок рецепта
	recipe.Purchases++
	if err := tx.Save(&recipe).Error; err != nil {
		tx.Rollback()
		utils.RespondWithJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"status":  "error",
			"message": "Failed to update recipe purchases",
		})
		return
	}

	// 6. Создаём транзакции в кошельке
	buyerTxID, _ := uuid.Parse(purchase.ID)
	buyerTx := models.WalletTransaction{
		UserID:      uuid.MustParse(req.BuyerID),
		Amount:      -recipe.Price,
		Type:        "purchase",
		Description: "Recipe purchase: " + recipe.Title,
		RelatedID:   buyerTxID,
	}
	sellerTx := models.WalletTransaction{
		UserID:      recipe.UserID,
		Amount:      netAmount,
		Type:        "sale",
		Description: "Recipe sale: " + recipe.Title,
		RelatedID:   buyerTxID,
	}
	tx.Create(&buyerTx)
	tx.Create(&sellerTx)

	tx.Commit()

	log.Printf("[MARKETPLACE] 💰 Recipe purchased: %s by %s for %.2f ChefToken", recipe.Title, buyerProfile.Name, recipe.Price)

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"status": "ok",
		"data": map[string]interface{}{
			"purchaseId":    purchase.ID,
			"recipe":        recipe.Title,
			"price":         recipe.Price,
			"commission":    commission,
			"sellerReceived": netAmount,
			"buyerBalance":  buyerProfile.WalletBalance,
		},
	})
}

// GetUserPurchasesHandler GET /api/user/{userId}/purchases - купленные рецепты
func GetUserPurchasesHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userId"]

	var purchases []models.RecipePurchase
	if err := database.DB.Where("buyer_id = ?", userID).Order("created_at DESC").Find(&purchases).Error; err != nil {
		utils.RespondWithJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"status":  "error",
			"message": "Failed to fetch purchases",
		})
		return
	}

	// Загружаем информацию о рецептах
	var enrichedPurchases []map[string]interface{}
	for _, purchase := range purchases {
		var recipe models.PersonalRecipe
		database.DB.Where("id = ?", purchase.RecipeID).First(&recipe)

		enrichedPurchases = append(enrichedPurchases, map[string]interface{}{
			"purchaseId":  purchase.ID,
			"recipeId":    recipe.ID,
			"title":       recipe.Title,
			"category":    recipe.Category,
			"imageUrl":    recipe.ImageURL,
			"price":       purchase.Price,
			"purchasedAt": purchase.CreatedAt,
		})
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"status": "ok",
		"data":   enrichedPurchases,
	})
}

// GetSellerStatsHandler GET /api/market/stats/{userId} - статистика продавца
func GetSellerStatsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sellerID := vars["userId"]

	var stats models.MarketStats
	stats.SellerID = sellerID

	// Общее количество продаж
	var totalSales int64
	database.DB.Model(&models.RecipePurchase{}).Where("seller_id = ?", sellerID).Count(&totalSales)
	stats.TotalSales = int(totalSales)

	// Общий доход (net amount)
	database.DB.Model(&models.RecipePurchase{}).Where("seller_id = ?", sellerID).Select("SUM(net_amount)").Row().Scan(&stats.TotalRevenue)

	// Средний рейтинг рецептов продавца
	database.DB.Table("PersonalRecipe").
		Where("user_id = ? AND rating > 0", sellerID).
		Select("AVG(rating)").
		Row().Scan(&stats.AverageRating)

	// Топ рецепт по продажам
	var topRecipe struct {
		RecipeID string
		Sales    int
	}
	database.DB.Model(&models.RecipePurchase{}).
		Select("recipe_id, COUNT(*) as sales").
		Where("seller_id = ?", sellerID).
		Group("recipe_id").
		Order("sales DESC").
		Limit(1).
		Scan(&topRecipe)

	if topRecipe.RecipeID != "" {
		var recipe models.PersonalRecipe
		database.DB.Where("id = ?", topRecipe.RecipeID).First(&recipe)
		stats.TopRecipe = recipe.Title
		stats.TopRecipeSales = topRecipe.Sales
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"status": "ok",
		"data":   stats,
	})
}

// GetLeaderboardHandler возвращает глобальный рейтинг поваров
func GetLeaderboardHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Query параметры
		sortBy := r.URL.Query().Get("sortBy")
		if sortBy == "" {
			sortBy = "xp" // по умолчанию сортируем по опыту
		}

		language := r.URL.Query().Get("language")
		limitStr := r.URL.Query().Get("limit")
		limit := 50 // по умолчанию топ-50
		if limitStr != "" {
			if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 100 {
				limit = parsedLimit
			}
		}

		// Базовый запрос - JOIN UserProfile + User
		type LeaderboardRow struct {
			UserID           string
			Name             string
			AvatarURL        string
			Level            int
			TotalXP          int
			Language         string
			TotalSales       int64
			TotalRevenue     float64
			AverageRating    float64
			RecipeCount      int64
			AchievementCount int64
		}

		var rows []LeaderboardRow

		// SQL запрос с агрегацией
		query := db.Table("\"UserProfile\"").
			Select(`
				"UserProfile".user_id,
				"User".name,
				"UserProfile".avatar_url,
				"UserProfile".level,
				"UserProfile".xp as total_xp,
				"UserProfile".language,
				COALESCE(sales.total_sales, 0) as total_sales,
				COALESCE(sales.total_revenue, 0) as total_revenue,
				COALESCE(recipes.avg_rating, 0) as average_rating,
				COALESCE(recipes.recipe_count, 0) as recipe_count,
				COALESCE(achievements.achievement_count, 0) as achievement_count
			`).
			Joins(`LEFT JOIN "User" ON "User".id::text = "UserProfile".user_id::text`).
			Joins(`
				LEFT JOIN (
					SELECT seller_id, COUNT(*) as total_sales, SUM(net_amount) as total_revenue
					FROM "RecipePurchase"
					GROUP BY seller_id
				) sales ON sales.seller_id::text = "UserProfile".user_id::text
			`).
			Joins(`
				LEFT JOIN (
					SELECT user_id, COUNT(*) as recipe_count, AVG(rating) as avg_rating
					FROM "PersonalRecipe"
					WHERE is_public = true AND rating > 0
					GROUP BY user_id
				) recipes ON recipes.user_id::text = "UserProfile".user_id::text
			`).
			Joins(`
				LEFT JOIN (
					SELECT user_id, COUNT(*) as achievement_count
					FROM "UserAchievement"
					GROUP BY user_id
				) achievements ON achievements.user_id::text = "UserProfile".user_id::text
			`)

		// Фильтр по языку
		if language != "" {
			query = query.Where("\"UserProfile\".language = ?", language)
		}

		// Сортировка
		switch sortBy {
		case "sales":
			query = query.Order("total_sales DESC, total_xp DESC")
		case "rating":
			query = query.Order("average_rating DESC, total_xp DESC")
		case "revenue":
			query = query.Order("total_revenue DESC, total_xp DESC")
		default: // "xp"
			query = query.Order("total_xp DESC, level DESC")
		}

		query = query.Limit(limit)

		if err := query.Find(&rows).Error; err != nil {
			utils.RespondWithJSON(w, http.StatusInternalServerError, map[string]interface{}{
				"status": "error",
				"error":  "Failed to fetch leaderboard",
			})
			return
		}

		// Формируем ответ с рангами
		leaders := make([]models.ChefLeaderboardEntry, len(rows))
		for i, row := range rows {
			leaders[i] = models.ChefLeaderboardEntry{
				Rank:             i + 1,
				UserID:           row.UserID,
				Name:             row.Name,
				AvatarURL:        row.AvatarURL,
				Level:            row.Level,
				TotalXP:          row.TotalXP,
				Language:         row.Language,
				TotalSales:       int(row.TotalSales),
				TotalRevenue:     row.TotalRevenue,
				AverageRating:    row.AverageRating,
				RecipeCount:      int(row.RecipeCount),
				AchievementCount: int(row.AchievementCount),
			}
		}

		response := models.LeaderboardResponse{
			Leaders: leaders,
			Total:   len(leaders),
			SortBy:  sortBy,
		}

		utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
			"status": "ok",
			"data":   response,
		})
	}
}
