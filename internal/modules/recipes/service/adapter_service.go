package service

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/recipes/dto"
	"gorm.io/gorm"
)

// RecipeAdapterService handles AI-powered recipe adaptation
type RecipeAdapterService struct {
	db         *gorm.DB
	groqClient GroqClient // Interface for AI client
}

// GroqClient interface for AI adaptation
type GroqClient interface {
	AdaptRecipe(prompt string, temperature float64) (string, error)
}

func NewRecipeAdapterService(db *gorm.DB, groqClient GroqClient) *RecipeAdapterService {
	return &RecipeAdapterService{
		db:         db,
		groqClient: groqClient,
	}
}

// AdaptRecipe adapts existing recipe to available ingredients using AI
func (s *RecipeAdapterService) AdaptRecipe(req dto.AdaptRecipeRequest) (*dto.AdaptedRecipeData, error) {
	// 1. Load original recipe from catalog
	var recipe models.RecipeCatalog
	err := s.db.
		Preload("Ingredients.Ingredient").
		Preload("Allergens").
		Preload("DietTags").
		Where("id = ?", req.RecipeID).
		First(&recipe).Error

	if err != nil {
		return nil, fmt.Errorf("recipe not found: %w", err)
	}

	// 2. Build AI prompt for adaptation
	prompt := s.buildAdaptationPrompt(recipe, req)

	// 3. Call AI for adaptation
	aiResponse, err := s.groqClient.AdaptRecipe(prompt, 0.7)
	if err != nil {
		return nil, fmt.Errorf("AI adaptation failed: %w", err)
	}

	// 4. Parse AI response
	adaptedData, err := s.parseAdaptationResponse(aiResponse, recipe, req)
	if err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	return adaptedData, nil
}

// buildAdaptationPrompt creates AI prompt for recipe adaptation
func (s *RecipeAdapterService) buildAdaptationPrompt(recipe models.RecipeCatalog, req dto.AdaptRecipeRequest) string {
	language := req.Language
	if language == "" {
		language = "pl"
	}

	// Build available ingredients list
	availableList := []string{}
	for _, ing := range req.FridgeSnapshot {
		availableList = append(availableList, fmt.Sprintf("- %s: %.0f %s", ing.Name, ing.Quantity, ing.Unit))
	}

	// Build missing ingredients list
	missingList := strings.Join(req.MissingIngredients, ", ")

	// Build original steps
	var steps []dto.RecipeStep
	json.Unmarshal(recipe.Steps, &steps)

	stepsText := ""
	for _, step := range steps {
		stepsText += fmt.Sprintf("%d. %s\n", step.Step, step.Instruction)
	}

	// Build required ingredients from recipe
	requiredList := []string{}
	for _, ing := range recipe.Ingredients {
		requiredList = append(requiredList, fmt.Sprintf("- %s: %.0f %s", ing.Ingredient.Name, ing.Quantity, ing.Unit))
	}

	prompt := fmt.Sprintf(`You are a professional chef adapting a recipe to available ingredients.

**CRITICAL RULES:**
1. DO NOT invent a new recipe
2. DO NOT change the core dish identity
3. ONLY adapt the existing recipe to available ingredients
4. Keep the original recipe name (with modifications if substitutions made)
5. Preserve cooking techniques when possible

**Original Recipe:**
Name: %s
Country: %s
Difficulty: %s
Time: %d minutes
Servings: %d

**Required Ingredients:**
%s

**Available in Fridge:**
%s

**Missing Ingredients:**
%s

**User Preferences:**
- Allow substitutions: %t
- Prefer expiring items: %t
- Simplify steps: %t
%s

**Your Task:**
Adapt this recipe to work with available ingredients. You can:
- ✅ Replace missing ingredients with similar available ones
- ✅ Reduce portions if not enough ingredients
- ✅ Simplify steps if requested
- ✅ Skip optional ingredients
- ✅ Adjust cooking times if needed

- ❌ DO NOT change the dish type (pasta stays pasta, soup stays soup)
- ❌ DO NOT invent completely different recipe
- ❌ DO NOT add ingredients not in fridge

**Output Format (JSON ONLY):**
{
  "adaptedName": "Recipe name (add suffix like 'z kurczakiem' if substituted)",
  "adaptedServings": number,
  "adaptedIngredients": [
    {
      "originalName": "string",
      "substitutedWith": "string or null",
      "quantity": number,
      "unit": "string",
      "isAvailable": boolean,
      "reason": "why substituted (if applicable)"
    }
  ],
  "adaptedSteps": [
    {
      "step": 1,
      "instruction": "Adapted instruction in %s"
    }
  ],
  "adaptations": [
    {
      "type": "substitution|portion_reduced|step_simplified|ingredient_removed",
      "description": "What changed in %s",
      "impact": "minor|moderate|major"
    }
  ],
  "canCookNow": boolean,
  "difficultyChange": "easier|same|harder",
  "timeChange": number (in minutes, can be negative)
}

Return ONLY valid JSON, no markdown, no explanation.`,
		recipe.CanonicalName,
		recipe.Country,
		recipe.Difficulty,
		recipe.TimeMinutes,
		recipe.Servings,
		strings.Join(requiredList, "\n"),
		strings.Join(availableList, "\n"),
		missingList,
		req.UserPreferences != nil && req.UserPreferences.AllowSubstitutions,
		req.UserPreferences != nil && req.UserPreferences.PreferExpiring,
		req.UserPreferences != nil && req.UserPreferences.SimplifySteps,
		s.buildPreferencesText(req.UserPreferences),
		language,
		language,
	)

	return prompt
}

// buildPreferencesText formats additional preferences
func (s *RecipeAdapterService) buildPreferencesText(prefs *dto.AdaptationPreferences) string {
	if prefs == nil {
		return ""
	}

	text := ""

	if prefs.ReduceServings != nil {
		text += fmt.Sprintf("- Reduce servings to: %d\n", *prefs.ReduceServings)
	}

	if len(prefs.AvoidAllergens) > 0 {
		text += fmt.Sprintf("- Avoid allergens: %s\n", strings.Join(prefs.AvoidAllergens, ", "))
	}

	return text
}

// parseAdaptationResponse parses AI JSON response
func (s *RecipeAdapterService) parseAdaptationResponse(
	aiResponse string,
	originalRecipe models.RecipeCatalog,
	req dto.AdaptRecipeRequest,
) (*dto.AdaptedRecipeData, error) {
	// Clean response (remove markdown if present)
	cleaned := strings.TrimSpace(aiResponse)
	cleaned = strings.TrimPrefix(cleaned, "```json")
	cleaned = strings.TrimPrefix(cleaned, "```")
	cleaned = strings.TrimSuffix(cleaned, "```")
	cleaned = strings.TrimSpace(cleaned)

	// Parse JSON
	var aiData struct {
		AdaptedName        string                  `json:"adaptedName"`
		AdaptedServings    int                     `json:"adaptedServings"`
		AdaptedIngredients []dto.AdaptedIngredient `json:"adaptedIngredients"`
		AdaptedSteps       []dto.RecipeStep        `json:"adaptedSteps"`
		Adaptations        []dto.Adaptation        `json:"adaptations"`
		CanCookNow         bool                    `json:"canCookNow"`
		DifficultyChange   string                  `json:"difficultyChange"`
		TimeChange         int                     `json:"timeChange"`
	}

	err := json.Unmarshal([]byte(cleaned), &aiData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse AI JSON: %w, response: %s", err, cleaned)
	}

	// Build response
	return &dto.AdaptedRecipeData{
		OriginalRecipeID:   originalRecipe.ID.String(),
		OriginalName:       originalRecipe.CanonicalName,
		OriginalServings:   originalRecipe.Servings,
		AdaptedName:        aiData.AdaptedName,
		AdaptedServings:    aiData.AdaptedServings,
		AdaptedSteps:       aiData.AdaptedSteps,
		AdaptedIngredients: aiData.AdaptedIngredients,
		Adaptations:        aiData.Adaptations,
		CanCookNow:         aiData.CanCookNow,
		DifficultyChange:   aiData.DifficultyChange,
		TimeChange:         aiData.TimeChange,
		AdaptedAt:          time.Now(),
	}, nil
}

// ValidateAdaptation ensures AI didn't go off-track
func (s *RecipeAdapterService) ValidateAdaptation(
	original models.RecipeCatalog,
	adapted *dto.AdaptedRecipeData,
) error {
	// Check 1: Name similarity (должно содержать оригинальное название или его часть)
	originalLower := strings.ToLower(original.CanonicalName)
	adaptedLower := strings.ToLower(adapted.AdaptedName)

	// Extract main dish name (e.g., "Carbonara" from "Spaghetti Carbonara")
	words := strings.Split(originalLower, " ")
	foundMatch := false
	for _, word := range words {
		if len(word) > 4 && strings.Contains(adaptedLower, word) {
			foundMatch = true
			break
		}
	}

	if !foundMatch {
		return fmt.Errorf("adapted name '%s' too different from original '%s'", adapted.AdaptedName, original.CanonicalName)
	}

	// Check 2: Category consistency (pasta stays pasta, soup stays soup)
	// This would require category detection, skip for now

	// Check 3: Reasonable portion reduction
	if adapted.AdaptedServings < 1 || adapted.AdaptedServings > original.Servings*2 {
		return fmt.Errorf("invalid servings: %d (original: %d)", adapted.AdaptedServings, original.Servings)
	}

	// Check 4: Has at least some steps
	if len(adapted.AdaptedSteps) == 0 {
		return fmt.Errorf("no cooking steps in adapted recipe")
	}

	return nil
}
