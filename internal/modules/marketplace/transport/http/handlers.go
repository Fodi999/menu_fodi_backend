package http

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/middleware"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/marketplace/dto"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/marketplace/repo"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/marketplace/service"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/platform/httpx"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/platform/logger"
)

// MarketplaceHandlers contains marketplace HTTP handlers
type MarketplaceHandlers struct {
	service           service.MarketplaceService
	cloudinaryService *service.CloudinaryService
}

// NewMarketplaceHandlers creates new marketplace handlers
func NewMarketplaceHandlers(marketplace service.MarketplaceService) *MarketplaceHandlers {
	return &MarketplaceHandlers{
		service:           marketplace,
		cloudinaryService: service.NewCloudinaryService(),
	}
}


// GetMarketRecipes godoc
// @Summary Get marketplace recipes
// @Description Get all public recipes from marketplace with filters
// @Tags marketplace
// @Produce json
// @Param category query string false "Category filter"
// @Param difficulty query string false "Difficulty filter"
// @Param maxPrice query number false "Max price filter"
// @Param minRating query number false "Min rating filter"
// @Param sortBy query string false "Sort by: popular, newest, rating, price"
// @Param limit query int false "Limit results (max 100)"
// @Success 200 {object} dto.MarketplaceResponse
// @Failure 500 {object} httpx.ErrorResponse
// @Router /api/marketplace/recipes [get]
func (h *MarketplaceHandlers) GetMarketRecipes(w http.ResponseWriter, r *http.Request) {
	filters := dto.MarketplaceFilters{
		Category:   r.URL.Query().Get("category"),
		Difficulty: r.URL.Query().Get("difficulty"),
		SortBy:     r.URL.Query().Get("sortBy"),
	}

	if maxPriceStr := r.URL.Query().Get("maxPrice"); maxPriceStr != "" {
		if maxPrice, err := strconv.ParseFloat(maxPriceStr, 64); err == nil {
			filters.MaxPrice = maxPrice
		}
	}

	if minRatingStr := r.URL.Query().Get("minRating"); minRatingStr != "" {
		if minRating, err := strconv.ParseFloat(minRatingStr, 64); err == nil {
			filters.MinRating = minRating
		}
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			filters.Limit = limit
		}
	}

	response, err := h.service.GetMarketRecipes(filters)
	if err != nil {
		logger.Error("failed to get market recipes", zap.Error(err))
		httpx.InternalError(w, "failed to fetch recipes")
		return
	}

	httpx.Success(w, response)
}

// PurchaseRecipe godoc
// @Summary Purchase recipe
// @Description Purchase a recipe from marketplace
// @Tags marketplace
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.PurchaseRequest true "Purchase request"
// @Success 200 {object} dto.PurchaseResponse
// @Failure 400 {object} httpx.ErrorResponse
// @Failure 401 {object} httpx.ErrorResponse
// @Failure 404 {object} httpx.ErrorResponse
// @Failure 500 {object} httpx.ErrorResponse
// @Router /api/marketplace/purchase [post]
func (h *MarketplaceHandlers) PurchaseRecipe(w http.ResponseWriter, r *http.Request) {
	userIDPtr := middleware.GetUserID(r)
	if userIDPtr == nil {
		logger.Error("user ID not found in context")
		httpx.Unauthorized(w, "unauthorized")
		return
	}
	userID := *userIDPtr

	var req dto.PurchaseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("failed to decode request", zap.Error(err))
		httpx.BadRequest(w, "invalid request body")
		return
	}

	// Set buyer ID from JWT
	req.BuyerID = userID.String()

	response, err := h.service.PurchaseRecipe(req)
	if err != nil {
		switch err {
		case repo.ErrRecipeNotFound:
			httpx.NotFound(w, "recipe not found")
		case repo.ErrCannotBuyOwnRecipe:
			httpx.BadRequest(w, "cannot purchase your own recipe")
		case repo.ErrAlreadyPurchased:
			httpx.BadRequest(w, "you already own this recipe")
		case repo.ErrInsufficientFunds:
			httpx.BadRequest(w, "insufficient ChefToken balance")
		default:
			logger.Error("failed to purchase recipe", zap.Error(err), zap.String("user_id", userID.String()))
			httpx.InternalError(w, "failed to process purchase")
		}
		return
	}

	logger.Info("recipe purchased",
		zap.String("user_id", userID.String()),
		zap.String("recipe_id", req.RecipeID),
		zap.Float64("price", response.Price))

	httpx.Success(w, response)
}

// GetUserPurchases godoc
// @Summary Get user purchases
// @Description Get all recipes purchased by user
// @Tags marketplace
// @Security BearerAuth
// @Produce json
// @Success 200 {array} dto.UserPurchase
// @Failure 401 {object} httpx.ErrorResponse
// @Failure 500 {object} httpx.ErrorResponse
// @Router /api/marketplace/purchases [get]
func (h *MarketplaceHandlers) GetUserPurchases(w http.ResponseWriter, r *http.Request) {
	userIDPtr := middleware.GetUserID(r)
	if userIDPtr == nil {
		logger.Error("user ID not found in context")
		httpx.Unauthorized(w, "unauthorized")
		return
	}
	userID := *userIDPtr

	purchases, err := h.service.GetUserPurchases(userID)
	if err != nil {
		logger.Error("failed to get purchases", zap.Error(err), zap.String("user_id", userID.String()))
		httpx.InternalError(w, "failed to fetch purchases")
		return
	}

	httpx.Success(w, purchases)
}

// GetSellerStats godoc
// @Summary Get seller statistics
// @Description Get statistics for a seller (total sales, revenue, ratings)
// @Tags marketplace
// @Produce json
// @Param userId path string true "Seller User ID"
// @Success 200 {object} dto.SellerStats
// @Failure 400 {object} httpx.ErrorResponse
// @Failure 500 {object} httpx.ErrorResponse
// @Router /api/marketplace/stats/{userId} [get]
func (h *MarketplaceHandlers) GetSellerStats(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "userId")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		httpx.BadRequest(w, "invalid user ID")
		return
	}

	stats, err := h.service.GetSellerStats(userID)
	if err != nil {
		logger.Error("failed to get seller stats", zap.Error(err), zap.String("seller_id", userID.String()))
		httpx.InternalError(w, "failed to fetch stats")
		return
	}

	httpx.Success(w, stats)
}

// GetLeaderboard godoc
// @Summary Get global leaderboard
// @Description Get global chef leaderboard with various sorting options
// @Tags marketplace
// @Produce json
// @Param sortBy query string false "Sort by: xp, sales, rating, revenue"
// @Param language query string false "Filter by language"
// @Param limit query int false "Limit results (default 50, max 100)"
// @Success 200 {object} dto.LeaderboardResponse
// @Failure 500 {object} httpx.ErrorResponse
// @Router /api/marketplace/leaderboard [get]
func (h *MarketplaceHandlers) GetLeaderboard(w http.ResponseWriter, r *http.Request) {
	sortBy := r.URL.Query().Get("sortBy")
	language := r.URL.Query().Get("language")

	limit := 50
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil {
			limit = parsedLimit
		}
	}

	response, err := h.service.GetLeaderboard(sortBy, language, limit)
	if err != nil {
		logger.Error("failed to get leaderboard", zap.Error(err))
		httpx.InternalError(w, "failed to fetch leaderboard")
		return
	}

	httpx.Success(w, response)
}

// UploadImage godoc
// @Summary Upload image
// @Description Upload image file and get Cloudinary URL
// @Tags marketplace
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param image formData file true "Image file to upload"
// @Success 200 {object} dto.UploadImageResponse
// @Failure 400 {object} httpx.ErrorResponse
// @Failure 401 {object} httpx.ErrorResponse
// @Failure 500 {object} httpx.ErrorResponse
// @Router /api/upload/image [post]
func (h *MarketplaceHandlers) UploadImage(w http.ResponseWriter, r *http.Request) {
	// Check authentication
	userIDPtr := middleware.GetUserID(r)
	if userIDPtr == nil {
		logger.Error("user ID not found in context")
		httpx.Unauthorized(w, "unauthorized")
		return
	}

	// Parse multipart form (max 10MB)
	if err := r.ParseMultipartForm(10 * 1024 * 1024); err != nil {
		logger.Error("failed to parse multipart form", zap.Error(err))
		httpx.BadRequest(w, "failed to parse form data")
		return
	}

	// Get image file from form
	file, handler, err := r.FormFile("image")
	if err != nil {
		logger.Error("failed to get image file", zap.Error(err))
		httpx.BadRequest(w, "image file is required")
		return
	}
	defer file.Close()

	// Validate file size (max 10MB)
	if handler.Size > 10*1024*1024 {
		httpx.BadRequest(w, "file size exceeds 10MB limit")
		return
	}

	// Read file data
	fileData, err := io.ReadAll(file)
	if err != nil {
		logger.Error("failed to read file data", zap.Error(err))
		httpx.InternalError(w, "failed to read file")
		return
	}

	// Validate file type (only images)
	contentType := handler.Header.Get("Content-Type")
	validTypes := map[string]bool{
		"image/jpeg":       true,
		"image/jpg":        true,
		"image/png":        true,
		"image/webp":       true,
		"image/gif":        true,
		"image/svg+xml":    true,
	}

	if !validTypes[contentType] {
		httpx.BadRequest(w, "invalid file type - only JPEG, PNG, WebP, GIF, and SVG are allowed")
		return
	}

	// Upload to Cloudinary
	uploadResp, err := h.cloudinaryService.UploadImage(fileData, handler.Filename)
	if err != nil {
		logger.Error("failed to upload image to cloudinary",
			zap.Error(err),
			zap.String("user_id", userIDPtr.String()),
			zap.String("filename", handler.Filename))
		httpx.InternalError(w, "failed to upload image")
		return
	}

	logger.Info("image uploaded successfully",
		zap.String("user_id", userIDPtr.String()),
		zap.String("filename", handler.Filename),
		zap.String("cloudinary_url", uploadResp.SecureURL))

	response := dto.UploadImageResponse{
		Success:   true,
		URL:       uploadResp.URL,
		SecureURL: uploadResp.SecureURL,
		PublicID:  uploadResp.PublicID,
		Message:   "Image uploaded successfully",
	}

	httpx.Success(w, response)
}
