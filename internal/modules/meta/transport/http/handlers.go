package http

import (
	"net/http"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/meta/service"
	"github.com/dmitrijfomin/menu-fodifood/backend/pkg/utils"
)

type MetaHandlers struct {
	service service.MetaService
}

func NewMetaHandlers(service service.MetaService) *MetaHandlers {
	return &MetaHandlers{service: service}
}

// GetCountries handles GET /api/meta/countries
func (h *MetaHandlers) GetCountries(w http.ResponseWriter, r *http.Request) {
	countries, err := h.service.GetCountries()
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch countries")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"data": countries,
		"meta": map[string]interface{}{
			"total":     len(countries),
			"cacheable": true,
		},
	})
}

// GetCuisines handles GET /api/meta/cuisines
func (h *MetaHandlers) GetCuisines(w http.ResponseWriter, r *http.Request) {
	cuisines, err := h.service.GetCuisines()
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch cuisines")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"data": cuisines,
		"meta": map[string]interface{}{
			"total":     len(cuisines),
			"cacheable": true,
		},
	})
}

// GetCategories handles GET /api/meta/categories
func (h *MetaHandlers) GetCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := h.service.GetCategories()
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch categories")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"data": categories,
		"meta": map[string]interface{}{
			"total":     len(categories),
			"cacheable": true,
		},
	})
}

// GetDifficulties handles GET /api/meta/difficulties
func (h *MetaHandlers) GetDifficulties(w http.ResponseWriter, r *http.Request) {
	difficulties, err := h.service.GetDifficulties()
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch difficulties")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"data": difficulties,
		"meta": map[string]interface{}{
			"total":     len(difficulties),
			"cacheable": true,
		},
	})
}
