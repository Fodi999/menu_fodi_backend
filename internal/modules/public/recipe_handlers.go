package public

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/admin/service"
	"github.com/dmitrijfomin/menu-fodifood/backend/pkg/utils"
	"github.com/go-chi/chi/v5"
)

// PublicRecipeHandlers - публичные endpoints для SEO
type PublicRecipeHandlers struct {
	service service.AdminService
}

func NewPublicRecipeHandlers(svc service.AdminService) *PublicRecipeHandlers {
	return &PublicRecipeHandlers{
		service: svc,
	}
}

// GetPublicRecipes - публичный каталог рецептов
// GET /api/public/recipes
func (h *PublicRecipeHandlers) GetPublicRecipes(w http.ResponseWriter, r *http.Request) {
	// Парсим фильтры
	filter := service.ParseRecipeFilter(r)

	// Получаем рецепты
	recipes, total, err := h.service.GetFilteredRecipes(filter)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch recipes")
		return
	}

	// Преобразуем в публичный DTO (без internal полей)
	publicRecipes := make([]PublicRecipeResponse, len(recipes))
	for i, recipe := range recipes {
		publicRecipes[i] = ToPublicRecipeResponse(&recipe)
	}

	// Cache headers для SEO (set BEFORE response)
	w.Header().Set("Cache-Control", "public, max-age=120") // 2 минуты кеш

	// SEO-friendly response
	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"data": publicRecipes,
		"meta": map[string]interface{}{
			"page":  filter.Page,
			"limit": filter.Limit,
			"total": total,
			"count": len(publicRecipes),
		},
	})
}

// GetRecipeBySlug - получить рецепт по canonical name (SEO URL)
// GET /api/public/recipes/{slug}
func (h *PublicRecipeHandlers) GetRecipeBySlug(w http.ResponseWriter, r *http.Request) {
	slugParam := chi.URLParam(r, "slug")

	if slugParam == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Recipe slug is required")
		return
	}

	// 🌐 CRITICAL FOR SEO: Decode URL-encoded UTF-8 characters
	// Supports both:
	// - /recipes/%D0%BB%D0%BE%D1%81%D0%BE%D1%81%D1%8C... (URL-encoded, Google crawler)
	// - /recipes/лосось_на_сковороде_с_травами (human-readable, copy-paste)
	decodedSlug, err := url.PathUnescape(slugParam)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid recipe slug encoding")
		return
	}

	// Получаем рецепт по canonicalName
	recipe, err := h.service.GetRecipeByCanonicalName(decodedSlug)
	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "Recipe not found")
		return
	}

	// Преобразуем в публичный DTO
	publicRecipe := ToPublicRecipeResponse(recipe)

	// Кеширование для SEO (set BEFORE response)
	w.Header().Set("Cache-Control", "public, max-age=300") // 5 минут кеш

	// SEO-friendly response
	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    publicRecipe,
	})
}

// PublicRecipeResponse - публичный DTO (без internal полей)
type PublicRecipeResponse struct {
	ID                 string      `json:"id"`
	CanonicalName      string      `json:"canonicalName"`
	Title              string      `json:"title"`
	NamePl             string      `json:"namePl"`
	NameEn             string      `json:"nameEn"`
	NameRu             string      `json:"nameRu"`
	DescriptionPl      string      `json:"descriptionPl"`
	DescriptionEn      string      `json:"descriptionEn"`
	DescriptionRu      string      `json:"descriptionRu"`
	Country            string      `json:"country"`
	Region             string      `json:"region"`
	Category           string      `json:"category"`
	Difficulty         string      `json:"difficulty"`
	TimeMinutes        int         `json:"timeMinutes"`
	Servings           int         `json:"servings"`
	PortionWeightGrams int         `json:"portionWeightGrams"`
	StepsPl            interface{} `json:"stepsPl"`
	StepsEn            interface{} `json:"stepsEn"`
	StepsRu            interface{} `json:"stepsRu"`
	NutritionProfile   interface{} `json:"nutritionProfile"`
	CreatedAt          string      `json:"createdAt"`
	// НЕ включаем: source, authorId, updatedAt (internal info)
}

// ToPublicRecipeResponse - конвертер в публичный DTO
func ToPublicRecipeResponse(r *models.RecipeCatalog) PublicRecipeResponse {
	resp := PublicRecipeResponse{
		ID:            r.ID.String(),
		CanonicalName: r.CanonicalName,
		Title:         r.Title,
		Country:       r.Country,
		Category:      r.Category,
		Difficulty:    r.Difficulty,
		TimeMinutes:   r.TimeMinutes,
		Servings:      r.Servings,
		CreatedAt:     r.CreatedAt.Format("2006-01-02"),
	}

	// Мультиязычные поля
	if r.NamePl != nil {
		resp.NamePl = *r.NamePl
	}
	if r.NameEn != nil {
		resp.NameEn = *r.NameEn
	}
	if r.NameRu != nil {
		resp.NameRu = *r.NameRu
	}
	if r.DescriptionPl != nil {
		resp.DescriptionPl = *r.DescriptionPl
	}
	if r.DescriptionEn != nil {
		resp.DescriptionEn = *r.DescriptionEn
	}
	if r.DescriptionRu != nil {
		resp.DescriptionRu = *r.DescriptionRu
	}
	if r.Region != nil {
		resp.Region = *r.Region
	}
	if r.PortionWeightGrams != nil {
		resp.PortionWeightGrams = *r.PortionWeightGrams
	}

	// JSONB поля
	if len(r.StepsPl) > 0 {
		var steps interface{}
		if err := json.Unmarshal(r.StepsPl, &steps); err == nil {
			resp.StepsPl = steps
		}
	}
	if len(r.StepsEn) > 0 {
		var steps interface{}
		if err := json.Unmarshal(r.StepsEn, &steps); err == nil {
			resp.StepsEn = steps
		}
	}
	if len(r.StepsRu) > 0 {
		var steps interface{}
		if err := json.Unmarshal(r.StepsRu, &steps); err == nil {
			resp.StepsRu = steps
		}
	}
	if len(r.NutritionProfile) > 0 {
		var nutrition interface{}
		if err := json.Unmarshal(r.NutritionProfile, &nutrition); err == nil {
			resp.NutritionProfile = nutrition
		}
	}

	return resp
}
