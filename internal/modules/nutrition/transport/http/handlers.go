package http

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/database"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/nutrition/service"
	"github.com/dmitrijfomin/menu-fodifood/backend/pkg/utils"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// NutritionHandlers handles HTTP requests for nutrition calculations
type NutritionHandlers struct {
	nutritionService *service.NutritionService
}

// NewNutritionHandlers creates a new handlers instance
func NewNutritionHandlers(svc *service.NutritionService) *NutritionHandlers {
	return &NutritionHandlers{
		nutritionService: svc,
	}
}

// parseIngredientString extracts name, quantity, and unit from ingredient string
// Examples: "egg 2 pcs", "milk 100 ml", "Яйця 2 шт"
func parseIngredientString(input string) service.Ingredient {
	input = strings.TrimSpace(input)

	// Regular expression to match: "name quantity unit" or "name quantity"
	// Matches numbers (int/float) followed by optional unit
	re := regexp.MustCompile(`^(.+?)\s+(\d+(?:\.\d+)?)\s*(.*)$`)
	matches := re.FindStringSubmatch(input)

	if len(matches) < 3 {
		// No quantity found, return just the name
		return service.Ingredient{
			Name:     input,
			Quantity: 0,
			Unit:     "",
		}
	}

	name := strings.TrimSpace(matches[1])
	quantity, _ := strconv.ParseFloat(matches[2], 64)
	unit := strings.TrimSpace(matches[3])

	return service.Ingredient{
		Name:     name,
		Quantity: quantity,
		Unit:     unit,
	}
}

// GetRecipeNutrition calculates and returns nutrition facts for a recipe
// GET /api/nutrition/recipe/{id}
func (h *NutritionHandlers) GetRecipeNutrition(w http.ResponseWriter, r *http.Request) {
	recipeIDStr := chi.URLParam(r, "id")

	recipeID, err := uuid.Parse(recipeIDStr)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid recipe ID")
		return
	}

	db := database.GetDB()

	// Get recipe from database
	var recipe struct {
		ID          uuid.UUID `json:"id"`
		Ingredients string    `json:"ingredients"`
		Servings    int       `json:"servings"`
	}

	if err := db.Table("ai_generated_recipes").
		Select("id, ingredients, servings").
		Where("id = ?", recipeID).
		First(&recipe).Error; err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "Recipe not found")
		return
	}

	// Parse ingredients from JSON string
	var ingredientsList []string
	if err := json.Unmarshal([]byte(recipe.Ingredients), &ingredientsList); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Invalid ingredients format")
		return
	}

	// Convert and parse to service.Ingredient format
	ingredients := make([]service.Ingredient, 0, len(ingredientsList))
	for _, item := range ingredientsList {
		parsed := parseIngredientString(item)
		ingredients = append(ingredients, parsed)
	}

	// Calculate nutrition
	servings := recipe.Servings
	if servings == 0 {
		servings = 2 // default
	}

	nutrition := h.nutritionService.CalculateRecipeNutrition(ingredients, servings)

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"recipeId":  recipe.ID,
		"nutrition": nutrition,
		"success":   true,
	})
}

// CalculateCustomNutrition calculates nutrition for custom ingredients
// POST /api/nutrition/calculate
func (h *NutritionHandlers) CalculateCustomNutrition(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Ingredients []string `json:"ingredients"`
		Servings    int      `json:"servings"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Servings == 0 {
		req.Servings = 2
	}

	// Parse and convert to service.Ingredient format
	ingredients := make([]service.Ingredient, 0, len(req.Ingredients))
	for _, item := range req.Ingredients {
		parsed := parseIngredientString(item)
		ingredients = append(ingredients, parsed)
	}

	nutrition := h.nutritionService.CalculateRecipeNutrition(ingredients, req.Servings)

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"nutrition": nutrition,
		"success":   true,
	})
}
