package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/dmitrijfomin/menu-fodifood/backend/pkg/utils"
)

// Временное хранилище ингредиентов (замените на БД)
type Ingredient struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	Unit         string  `json:"unit"` // g, ml, pcs
	CurrentStock float64 `json:"currentStock"`
	MinStock     float64 `json:"minStock"`
	MaxStock     float64 `json:"maxStock"`
	CostPerUnit  float64 `json:"costPerUnit"`
	Supplier     *string `json:"supplier,omitempty"`
	CreatedAt    string  `json:"createdAt"`
	UpdatedAt    string  `json:"updatedAt"`
}

var ingredients = []Ingredient{}

// GetAllIngredients получение всех ингредиентов
func GetAllIngredients(w http.ResponseWriter, r *http.Request) {
	utils.RespondWithJSON(w, http.StatusOK, ingredients)
}

// CreateIngredient создание нового ингредиента
func CreateIngredient(w http.ResponseWriter, r *http.Request) {
	var req Ingredient
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	// TODO: Генерация ID и сохранение в БД
	req.ID = "ingredient-" + req.Name
	req.CreatedAt = "2025-10-05T21:00:00Z"
	req.UpdatedAt = "2025-10-05T21:00:00Z"

	ingredients = append(ingredients, req)

	utils.RespondWithJSON(w, http.StatusCreated, req)
}

// UpdateIngredient обновление ингредиента
func UpdateIngredient(w http.ResponseWriter, r *http.Request) {
	// TODO: Реализовать обновление через БД
	utils.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Updated"})
}

// DeleteIngredient удаление ингредиента
func DeleteIngredient(w http.ResponseWriter, r *http.Request) {
	// TODO: Реализовать удаление через БД
	utils.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Deleted"})
}
