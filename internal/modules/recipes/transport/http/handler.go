package http

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/recipes/dto"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/recipes/service"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"gorm.io/datatypes"
)

type RecipeHandler struct {
	matchService   *service.RecipeMatchService
	adapterService *service.RecipeAdapterService
	cookService    *service.RecipeCookService
	logger         *zap.Logger
}

func NewRecipeHandler(
	matchService *service.RecipeMatchService,
	adapterService *service.RecipeAdapterService,
	cookService *service.RecipeCookService,
	logger *zap.Logger,
) *RecipeHandler {
	return &RecipeHandler{
		matchService:   matchService,
		adapterService: adapterService,
		cookService:    cookService,
		logger:         logger,
	}
}

// MatchRecipes finds recipes matching user's fridge
// GET /api/recipes/match?country=Poland&maxTime=60&excludeAllergens=gluten,lactose&minScore=50
func (h *RecipeHandler) MatchRecipes(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by auth middleware)
	userID, ok := r.Context().Value("userID").(string)
	if !ok || userID == "" {
		// DEV MODE: Allow testing without auth by using test user ID from query
		testUserID := r.URL.Query().Get("testUserID")
		if testUserID != "" {
			h.logger.Warn("⚠️ DEV MODE: Using test userID from query parameter")
			userID = testUserID
		} else {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
	}

	// Parse query parameters
	filters := service.RecipeFilters{
		Country:          r.URL.Query().Get("country"),
		Category:         r.URL.Query().Get("category"),
		Difficulty:       r.URL.Query().Get("difficulty"),
		MaxTime:          parseIntQuery(r, "maxTime", 0),
		ExcludeAllergens: parseArrayQuery(r, "excludeAllergens"),
		IncludeDietTags:  parseArrayQuery(r, "dietTags"),
		MinScore:         parseFloatQuery(r, "minScore", 0.0),
		OnlyCookable:     parseBoolQuery(r, "onlyCookable", false),
		Limit:            parseIntQuery(r, "limit", 10),
	}

	h.logger.Info("Matching recipes with fridge",
		zap.String("userId", userID),
		zap.Any("filters", filters),
	)

	// Find matching recipes
	matches, err := h.matchService.MatchRecipesWithFridge(userID, filters)
	if err != nil {
		h.logger.Error("Failed to match recipes", zap.Error(err))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(dto.RecipeMatchResponse{
			Success: false,
			Error:   "Failed to find recipes",
		})
		return
	}

	h.logger.Info("Found matching recipes",
		zap.String("userId", userID),
		zap.Int("count", len(matches)),
	)

	// Convert to DTO format
	recipeItems := make([]dto.RecipeMatchItem, len(matches))
	for i, match := range matches {
		recipeItems[i] = convertToRecipeMatchItem(match)
	}

	// Return results with standard contract
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dto.RecipeMatchResponse{
		Success: true,
		Data: &dto.RecipeMatchData{
			Recipes: recipeItems,
			Count:   len(recipeItems),
		},
	})
}

// GetRecipeByID returns full recipe details
// GET /api/recipes/:id
func (h *RecipeHandler) GetRecipeByID(w http.ResponseWriter, r *http.Request) {
	recipeID := chi.URLParam(r, "id")
	if recipeID == "" {
		http.Error(w, "Recipe ID required", http.StatusBadRequest)
		return
	}

	// TODO: Implement recipe detail loading
	h.logger.Info("Getting recipe by ID", zap.String("recipeId", recipeID))

	http.Error(w, "Not implemented yet", http.StatusNotImplemented)
}

// ListRecipes returns filtered recipe catalog
// GET /api/recipes?country=Poland&category=main&difficulty=easy&maxTime=30
func (h *RecipeHandler) ListRecipes(w http.ResponseWriter, r *http.Request) {
	filters := service.RecipeFilters{
		Country:          r.URL.Query().Get("country"),
		Category:         r.URL.Query().Get("category"),
		Difficulty:       r.URL.Query().Get("difficulty"),
		MaxTime:          parseIntQuery(r, "maxTime", 0),
		ExcludeAllergens: parseArrayQuery(r, "excludeAllergens"),
		IncludeDietTags:  parseArrayQuery(r, "dietTags"),
		Limit:            parseIntQuery(r, "limit", 20),
	}

	h.logger.Info("Listing recipes", zap.Any("filters", filters))

	// TODO: Implement recipe listing
	http.Error(w, "Not implemented yet", http.StatusNotImplemented)
}

// AdaptRecipe adapts existing recipe to available ingredients using AI
// POST /api/recipes/:id/adapt
func (h *RecipeHandler) AdaptRecipe(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, ok := r.Context().Value("userID").(string)
	if !ok || userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	recipeID := chi.URLParam(r, "id")
	if recipeID == "" {
		http.Error(w, "Recipe ID required", http.StatusBadRequest)
		return
	}

	// Parse request body
	var req dto.AdaptRecipeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Set recipeID from URL
	req.RecipeID = recipeID

	h.logger.Info("Adapting recipe",
		zap.String("userId", userID),
		zap.String("recipeId", recipeID),
		zap.Int("fridgeItems", len(req.FridgeSnapshot)),
		zap.Int("missingItems", len(req.MissingIngredients)),
	)

	// Call adapter service
	adaptedRecipe, err := h.adapterService.AdaptRecipe(req)
	if err != nil {
		h.logger.Error("Failed to adapt recipe", zap.Error(err))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(dto.AdaptRecipeResponse{
			Success: false,
			Error:   "Failed to adapt recipe",
			Message: err.Error(),
		})
		return
	}

	h.logger.Info("Recipe adapted successfully",
		zap.String("userId", userID),
		zap.String("recipeId", recipeID),
		zap.String("adaptedName", adaptedRecipe.AdaptedName),
		zap.Int("adaptations", len(adaptedRecipe.Adaptations)),
	)

	// Return adapted recipe
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dto.AdaptRecipeResponse{
		Success: true,
		Data:    adaptedRecipe,
	})
}

// CookRecipe deducts ingredients from fridge and logs cooking event
// POST /api/recipes/:id/cook
func (h *RecipeHandler) CookRecipe(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, ok := r.Context().Value("userID").(string)
	if !ok || userID == "" {
		// DEV MODE: Allow testing without auth
		testUserID := r.URL.Query().Get("testUserID")
		if testUserID != "" {
			h.logger.Warn("⚠️ DEV MODE: Using test userID from query parameter")
			userID = testUserID
		} else {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
	}

	recipeID := chi.URLParam(r, "id")
	if recipeID == "" {
		http.Error(w, "Recipe ID required", http.StatusBadRequest)
		return
	}

	// Parse request body
	var req dto.CookRecipeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Set recipeID from URL
	req.RecipeID = recipeID

	// Default servings multiplier
	if req.ServingsMultiplier <= 0 {
		req.ServingsMultiplier = 1.0
	}

	h.logger.Info("Cooking recipe",
		zap.String("userId", userID),
		zap.String("recipeId", recipeID),
		zap.Float64("servingsMultiplier", req.ServingsMultiplier),
		zap.String("idempotencyKey", req.IdempotencyKey),
	)

	// Prepare idempotency key
	var idempotencyKeyPtr *string
	if req.IdempotencyKey != "" {
		idempotencyKeyPtr = &req.IdempotencyKey
	}

	// Cook recipe
	cookData, err := h.cookService.CookRecipe(
		userID,
		recipeID,
		req.ServingsMultiplier,
		idempotencyKeyPtr,
	)
	if err != nil {
		h.logger.Error("Failed to cook recipe", zap.Error(err))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(dto.CookRecipeResponse{
			Success: false,
			Error:   "Failed to cook recipe",
			Message: err.Error(),
		})
		return
	}

	h.logger.Info("Recipe cooked successfully",
		zap.String("userId", userID),
		zap.String("recipeId", recipeID),
		zap.String("cookLogId", cookData.CookLogID),
		zap.Float64("usedValue", cookData.UsedValue),
		zap.Float64("wasteRiskSaved", cookData.WasteRiskSaved),
		zap.Int("remainingItems", cookData.RemainingItems),
	)

	// Return cook result
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dto.CookRecipeResponse{
		Success: true,
		Data:    cookData,
	})
}

// Helper functions

func parseIntQuery(r *http.Request, key string, defaultValue int) int {
	val := r.URL.Query().Get(key)
	if val == "" {
		return defaultValue
	}
	parsed, err := strconv.Atoi(val)
	if err != nil {
		return defaultValue
	}
	return parsed
}

func parseFloatQuery(r *http.Request, key string, defaultValue float64) float64 {
	val := r.URL.Query().Get(key)
	if val == "" {
		return defaultValue
	}
	parsed, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return defaultValue
	}
	return parsed
}

func parseBoolQuery(r *http.Request, key string, defaultValue bool) bool {
	val := r.URL.Query().Get(key)
	if val == "" {
		return defaultValue
	}
	parsed, err := strconv.ParseBool(val)
	if err != nil {
		return defaultValue
	}
	return parsed
}

func parseArrayQuery(r *http.Request, key string) []string {
	val := r.URL.Query().Get(key)
	if val == "" {
		return []string{}
	}
	parts := strings.Split(val, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// convertToRecipeMatchItem converts service.RecipeMatch to dto.RecipeMatchItem
func convertToRecipeMatchItem(match service.RecipeMatch) dto.RecipeMatchItem {
	// Convert used ingredients
	usedIngredients := make([]dto.IngredientMatch, len(match.MatchedIngredients))
	for i, ing := range match.MatchedIngredients {
		usedIngredients[i] = dto.IngredientMatch{
			IngredientID:   ing.IngredientID,
			Name:           ing.Name,
			Quantity:       ing.Required,
			Unit:           ing.Unit,
			Available:      ing.Available,
			IsExpiringSoon: ing.IsExpiringSoon,
		}
	}

	// Convert missing ingredients
	missingIngredients := make([]dto.IngredientMatch, len(match.MissingIngredients))
	for i, ing := range match.MissingIngredients {
		missingIngredients[i] = dto.IngredientMatch{
			IngredientID:  ing.IngredientID,
			Name:          ing.Name,
			Quantity:      ing.Required,
			Unit:          ing.Unit,
			Optional:      ing.Optional,
			EstimatedCost: ing.EstimatedCost,
		}
	}

	// Extract allergen names
	allergens := make([]string, len(match.Recipe.Allergens))
	for i, allergen := range match.Recipe.Allergens {
		allergens[i] = allergen.Name
	}

	// Extract diet tag names
	dietTags := make([]string, len(match.Recipe.DietTags))
	for i, tag := range match.Recipe.DietTags {
		dietTags[i] = tag.Name
	}

	// Calculate coverage: usedRequired / totalRequired (excluding optional)
	coverage := 0.0
	usedRequired := 0
	totalRequired := 0
	
	// Count only required (non-optional) ingredients
	for _, ing := range match.MatchedIngredients {
		if !ing.Optional {
			usedRequired++
			totalRequired++
		}
	}
	for _, ing := range match.MissingIngredients {
		if !ing.Optional {
			totalRequired++
		}
	}
	
	if totalRequired > 0 {
		coverage = float64(usedRequired) / float64(totalRequired)
		coverage = roundToTwoDecimals(coverage)
	}

	return dto.RecipeMatchItem{
		RecipeID:           match.Recipe.ID.String(),
		CanonicalName:      match.Recipe.CanonicalName,
		LocalName:          match.Recipe.LocalName,
		Country:            match.Recipe.Country,
		Category:           match.Recipe.Category,
		Difficulty:         match.Recipe.Difficulty,
		TimeMinutes:        match.Recipe.TimeMinutes,
		Servings:           match.Recipe.Servings,
		Score:              match.MatchScore,
		Coverage:           coverage,
		UsedIngredients:    usedIngredients,
		MissingIngredients: missingIngredients,
		CanCookNow:         match.CanMakeNow,
		CostToComplete:     match.CostToComplete,
		UsedValue:          match.UsedValue,
		SavedMoney:         match.SavedMoney,
		TotalRecipeCost:    match.TotalRecipeCost,
		WasteRiskSaved:     match.WasteRiskSaved,
		HasExpiringItems:   match.HasExpiringItems,
		ExpiringItemsCount: match.ExpiringItemsCount,
		Allergens:          allergens,
		DietTags:           dietTags,
	}
}

// roundToTwoDecimals rounds a float to 2 decimal places
func roundToTwoDecimals(value float64) float64 {
	return math.Round(value*100) / 100
}

// GetRecommendation возвращает ОДИН лучший рецепт для замены AI-генерации
// POST /api/recipes/recommendations
func (h *RecipeHandler) GetRecommendation(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by auth middleware)
	userID, ok := r.Context().Value("userID").(string)
	if !ok || userID == "" {
		// DEV MODE: Allow testing without auth
		testUserID := r.URL.Query().Get("testUserID")
		if testUserID != "" {
			h.logger.Warn("⚠️ DEV MODE: Using test userID from query parameter")
			userID = testUserID
		} else {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
	}

	// Parse request body
	var req dto.RecommendationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(dto.RecommendationResponse{
			Success: false,
			Error:   "Invalid request format",
		})
		return
	}

	// Validate mode
	if req.Mode != "fridge" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(dto.RecommendationResponse{
			Success: false,
			Error:   "Only 'fridge' mode is supported",
		})
		return
	}

	// Set default limit
	if req.Limit <= 0 {
		req.Limit = 5
	}

	h.logger.Info("Getting recipe recommendation",
		zap.String("userId", userID),
		zap.String("mode", req.Mode),
		zap.Int("limit", req.Limit),
		zap.Int("excludeCount", len(req.ExcludeRecipeIds)),
	)

	// Get best recipe match (excluding already shown recipes)
	bestMatch, err := h.matchService.GetBestRecommendation(userID, req.Limit, req.ExcludeRecipeIds)
	if err != nil {
		h.logger.Error("Failed to get recommendation", zap.Error(err))
		w.Header().Set("Content-Type", "application/json")
		
		// Return 200 with friendly message instead of 500
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(dto.RecommendationResponse{
			Success: false,
			Error:   "No recipes available",
			Message: "We couldn't find any recipes matching your fridge. Try adding more ingredients!",
		})
		return
	}

	h.logger.Info("Found best recipe",
		zap.String("userId", userID),
		zap.String("recipeId", bestMatch.Recipe.ID.String()),
		zap.String("recipeName", bestMatch.Recipe.LocalName),
		zap.Float64("score", bestMatch.MatchScore),
		zap.Bool("canCookNow", bestMatch.CanMakeNow),
	)

	// Convert to recommendation response format (UI-compatible)
	response := convertToRecommendationResponse(bestMatch)

	// Return result
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// convertToRecommendationResponse преобразует RecipeMatch в формат совместимый с текущим UI
func convertToRecommendationResponse(match *service.RecipeMatch) dto.RecommendationResponse {
	// Recipe info
	recipeInfo := dto.RecipeInfo{
		ID:            match.Recipe.ID.String(),
		CanonicalName: match.Recipe.CanonicalName,
		LocalName:     match.Recipe.LocalName,
		Country:       match.Recipe.Country,
		Category:      match.Recipe.Category,
		Difficulty:    match.Recipe.Difficulty,
		TimeMinutes:   match.Recipe.TimeMinutes,
		Servings:      match.Recipe.Servings,
	}

	// Convert Steps from JSON to []string (support both formats)
	recipeInfo.Steps = parseStepsFromJSON(match.Recipe.Steps)

	// Allergens
	allergens := make([]string, len(match.Recipe.Allergens))
	for i, allergen := range match.Recipe.Allergens {
		allergens[i] = allergen.Name
	}
	recipeInfo.Allergens = allergens

	// Diet tags
	dietTags := make([]string, len(match.Recipe.DietTags))
	for i, tag := range match.Recipe.DietTags {
		dietTags[i] = tag.Name
	}
	recipeInfo.DietTags = dietTags

	// Missing ingredients (required only, no optional)
	missingRequired := []dto.MissingIngredient{}
	for _, missing := range match.MissingIngredients {
		if !missing.Optional {
			missingRequired = append(missingRequired, dto.MissingIngredient{
				IngredientID:  missing.IngredientID,
				Name:          missing.Name,
				Quantity:      missing.Required,
				Unit:          missing.Unit,
				EstimatedCost: missing.EstimatedCost,
			})
		}
	}

	// Used ingredients
	usedIngredients := make([]dto.UsedIngredient, len(match.MatchedIngredients))
	for i, matched := range match.MatchedIngredients {
		usedIngredients[i] = dto.UsedIngredient{
			IngredientID:   matched.IngredientID,
			Name:           matched.Name,
			Quantity:       matched.Required,
			Unit:           matched.Unit,
			Available:      matched.Available,
			IsExpiringSoon: matched.IsExpiringSoon,
		}
	}

	// Match info
	matchInfo := dto.MatchInfo{
		CanCookNow:      match.CanMakeNow,
		MissingRequired: missingRequired,
		UsedIngredients: usedIngredients,
	}

	// Economy info
	economyInfo := dto.EconomyInfo{
		UsedFromFridge: match.UsedValue,
		Saved:          match.SavedMoney,
	}

	return dto.RecommendationResponse{
		Success: true,
		Data: &dto.RecommendationData{
			Recipe:  recipeInfo,
			Match:   matchInfo,
		Economy: economyInfo,
	},
}
}

// parseStepsFromJSON parses steps from JSON supporting both formats:
// 1. Simple string array: ["1. Step one", "2. Step two", ...]
// 2. Object array: [{"step": 1, "instruction": "Step one"}, ...]
func parseStepsFromJSON(stepsJSON datatypes.JSON) []string {
	if len(stepsJSON) == 0 {
		return []string{}
	}

	// Try parsing as simple string array first (most common)
	var stringSteps []string
	if err := json.Unmarshal(stepsJSON, &stringSteps); err == nil {
		return stringSteps
	}

	// Try parsing as object array with {step, instruction} structure
	var objectSteps []struct {
		Step        int    `json:"step"`
		Instruction string `json:"instruction"`
	}
	if err := json.Unmarshal(stepsJSON, &objectSteps); err == nil {
		result := make([]string, len(objectSteps))
		for i, obj := range objectSteps {
			// Format as "1. instruction text"
			result[i] = fmt.Sprintf("%d. %s", obj.Step, obj.Instruction)
		}
		return result
	}

	// Fallback to empty array if both formats fail
	return []string{}
}
