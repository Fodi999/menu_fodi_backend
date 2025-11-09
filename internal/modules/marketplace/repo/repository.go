package repo

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/marketplace/dto"
)

var (
	ErrRecipeNotFound     = errors.New("recipe not found")
	ErrAlreadyPurchased   = errors.New("recipe already purchased")
	ErrCannotBuyOwnRecipe = errors.New("cannot purchase your own recipe")
	ErrInsufficientFunds  = errors.New("insufficient ChefToken balance")
)

// MarketplaceRepository handles marketplace data operations
type MarketplaceRepository interface {
	// Recipe operations
	GetMarketRecipes(filters dto.MarketplaceFilters) ([]dto.RecipeWithAuthor, error)
	GetRecipeByID(recipeID uuid.UUID) (*models.PersonalRecipe, error)
	IncrementPurchases(recipeID uuid.UUID) error

	// Purchase operations
	CheckPurchaseExists(recipeID, buyerID uuid.UUID) (bool, error)
	CreatePurchase(purchase *models.RecipePurchase) error
	GetUserPurchases(userID uuid.UUID) ([]dto.UserPurchase, error)

	// Profile operations
	GetUserProfile(userID uuid.UUID) (*models.UserProfile, error)
	UpdateWalletBalance(userID uuid.UUID, amount float64) error

	// Stats operations
	GetSellerStats(sellerID uuid.UUID) (*dto.SellerStats, error)
	GetLeaderboard(sortBy, language string, limit int) ([]dto.LeaderboardEntry, error)

	// Transaction operations
	CreateWalletTransaction(tx *models.WalletTransaction) error
}

type marketplaceRepository struct {
	db *gorm.DB
}

// NewMarketplaceRepository creates new marketplace repository
func NewMarketplaceRepository(db *gorm.DB) MarketplaceRepository {
	return &marketplaceRepository{db: db}
}

func (r *marketplaceRepository) GetMarketRecipes(filters dto.MarketplaceFilters) ([]dto.RecipeWithAuthor, error) {
	query := r.db.Where("is_public = ?", true)

	// Apply filters
	if filters.Category != "" {
		query = query.Where("category = ?", filters.Category)
	}
	if filters.Difficulty != "" {
		query = query.Where("difficulty = ?", filters.Difficulty)
	}
	if filters.MaxPrice > 0 {
		query = query.Where("price <= ?", filters.MaxPrice)
	}
	if filters.MinRating > 0 {
		query = query.Where("rating >= ?", filters.MinRating)
	}

	// Sorting
	switch filters.SortBy {
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

	// Limit
	limit := 50
	if filters.Limit > 0 && filters.Limit <= 100 {
		limit = filters.Limit
	}
	query = query.Limit(limit)

	var recipes []models.PersonalRecipe
	if err := query.Find(&recipes).Error; err != nil {
		return nil, err
	}

	// Enrich with author info
	var enrichedRecipes []dto.RecipeWithAuthor
	for _, recipe := range recipes {
		var profile models.UserProfile
		r.db.Where("user_id = ?", recipe.UserID).First(&profile)

		// Get review stats
		var reviewCount int64
		var avgReview float64
		r.db.Model(&models.RecipeReview{}).Where("recipe_id = ?", recipe.ID).Count(&reviewCount)
		r.db.Model(&models.RecipeReview{}).Where("recipe_id = ?", recipe.ID).Select("AVG(rating)").Row().Scan(&avgReview)

		enrichedRecipes = append(enrichedRecipes, dto.RecipeWithAuthor{
			ID:           recipe.ID,
			UserID:       recipe.UserID,
			Title:        recipe.Title,
			Description:  recipe.Description,
			Category:     recipe.Category,
			Difficulty:   recipe.Difficulty,
			Price:        recipe.Price,
			Rating:       recipe.Rating,
			Purchases:    recipe.Purchases,
			ImageURL:     recipe.ImageURL,
			IsPublic:     recipe.IsPublic,
			CreatedAt:    recipe.CreatedAt,
			AuthorName:   profile.Name,
			AuthorLevel:  profile.Level,
			AuthorAvatar: profile.AvatarURL,
			ReviewCount:  int(reviewCount),
			AvgReview:    avgReview,
		})
	}

	return enrichedRecipes, nil
}

func (r *marketplaceRepository) GetRecipeByID(recipeID uuid.UUID) (*models.PersonalRecipe, error) {
	var recipe models.PersonalRecipe
	err := r.db.Where("id = ?", recipeID).First(&recipe).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRecipeNotFound
		}
		return nil, err
	}
	return &recipe, nil
}

func (r *marketplaceRepository) IncrementPurchases(recipeID uuid.UUID) error {
	return r.db.Model(&models.PersonalRecipe{}).
		Where("id = ?", recipeID).
		UpdateColumn("purchases", gorm.Expr("purchases + ?", 1)).Error
}

func (r *marketplaceRepository) CheckPurchaseExists(recipeID, buyerID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&models.RecipePurchase{}).
		Where("recipe_id = ? AND buyer_id = ?", recipeID.String(), buyerID.String()).
		Count(&count).Error
	return count > 0, err
}

func (r *marketplaceRepository) CreatePurchase(purchase *models.RecipePurchase) error {
	return r.db.Create(purchase).Error
}

func (r *marketplaceRepository) GetUserPurchases(userID uuid.UUID) ([]dto.UserPurchase, error) {
	var purchases []models.RecipePurchase
	if err := r.db.Where("buyer_id = ?", userID.String()).
		Order("created_at DESC").
		Find(&purchases).Error; err != nil {
		return nil, err
	}

	var enrichedPurchases []dto.UserPurchase
	for _, purchase := range purchases {
		var recipe models.PersonalRecipe
		r.db.Where("id = ?", purchase.RecipeID).First(&recipe)

		enrichedPurchases = append(enrichedPurchases, dto.UserPurchase{
			PurchaseID:  purchase.ID,
			RecipeID:    recipe.ID,
			Title:       recipe.Title,
			Category:    recipe.Category,
			ImageURL:    recipe.ImageURL,
			Price:       purchase.Price,
			PurchasedAt: purchase.CreatedAt,
		})
	}

	return enrichedPurchases, nil
}

func (r *marketplaceRepository) GetUserProfile(userID uuid.UUID) (*models.UserProfile, error) {
	var profile models.UserProfile
	err := r.db.Where("user_id = ?", userID).First(&profile).Error
	if err != nil {
		return nil, err
	}
	return &profile, nil
}

func (r *marketplaceRepository) UpdateWalletBalance(userID uuid.UUID, amount float64) error {
	return r.db.Model(&models.UserProfile{}).
		Where("user_id = ?", userID).
		UpdateColumn("wallet_balance", gorm.Expr("wallet_balance + ?", amount)).Error
}

func (r *marketplaceRepository) GetSellerStats(sellerID uuid.UUID) (*dto.SellerStats, error) {
	stats := &dto.SellerStats{
		SellerID: sellerID.String(),
	}

	// Total sales count
	var totalSales int64
	r.db.Model(&models.RecipePurchase{}).
		Where("seller_id = ?", sellerID.String()).
		Count(&totalSales)
	stats.TotalSales = int(totalSales)

	// Total revenue (net amount)
	r.db.Model(&models.RecipePurchase{}).
		Where("seller_id = ?", sellerID.String()).
		Select("SUM(net_amount)").
		Row().Scan(&stats.TotalRevenue)

	// Average rating
	r.db.Table("\"PersonalRecipe\"").
		Where("user_id = ? AND rating > 0", sellerID).
		Select("AVG(rating)").
		Row().Scan(&stats.AverageRating)

	// Top recipe by sales
	var topRecipe struct {
		RecipeID string
		Sales    int
	}
	r.db.Model(&models.RecipePurchase{}).
		Select("recipe_id, COUNT(*) as sales").
		Where("seller_id = ?", sellerID.String()).
		Group("recipe_id").
		Order("sales DESC").
		Limit(1).
		Scan(&topRecipe)

	if topRecipe.RecipeID != "" {
		var recipe models.PersonalRecipe
		r.db.Where("id = ?", topRecipe.RecipeID).First(&recipe)
		stats.TopRecipe = recipe.Title
		stats.TopRecipeSales = topRecipe.Sales
	}

	return stats, nil
}

func (r *marketplaceRepository) GetLeaderboard(sortBy, language string, limit int) ([]dto.LeaderboardEntry, error) {
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

	query := r.db.Table("\"UserProfile\"").
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

	if language != "" {
		query = query.Where("\"UserProfile\".language = ?", language)
	}

	switch sortBy {
	case "sales":
		query = query.Order("total_sales DESC, total_xp DESC")
	case "rating":
		query = query.Order("average_rating DESC, total_xp DESC")
	case "revenue":
		query = query.Order("total_revenue DESC, total_xp DESC")
	default:
		query = query.Order("total_xp DESC, level DESC")
	}

	if limit <= 0 || limit > 100 {
		limit = 50
	}
	query = query.Limit(limit)

	if err := query.Find(&rows).Error; err != nil {
		return nil, err
	}

	leaders := make([]dto.LeaderboardEntry, len(rows))
	for i, row := range rows {
		leaders[i] = dto.LeaderboardEntry{
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

	return leaders, nil
}

func (r *marketplaceRepository) CreateWalletTransaction(tx *models.WalletTransaction) error {
	return r.db.Create(tx).Error
}
