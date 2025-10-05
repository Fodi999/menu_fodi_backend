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

var ingredientRepo = &database.IngredientRepository{}

// GetAllIngredients получение всех ингредиентов со складскими остатками
func GetAllIngredients(w http.ResponseWriter, r *http.Request) {
	stockItems, err := ingredientRepo.FindAll()
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch ingredients")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, stockItems)
}

// GetIngredient получение одного ингредиента
func GetIngredient(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	stockItem, err := ingredientRepo.FindByID(id)
	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "Ingredient not found")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, stockItem)
}

// CreateIngredient создание нового ингредиента
func CreateIngredient(w http.ResponseWriter, r *http.Request) {
	var req models.CreateIngredientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	// Валидация
	if req.Name == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Name is required")
		return
	}
	if req.Unit == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Unit is required")
		return
	}

	// Создаем ингредиент
	ingredient := &models.Ingredient{
		ID:        uuid.New().String(),
		Name:      req.Name,
		Unit:      req.Unit,
		CreatedAt: time.Now(),
	}

	// Создаем складской остаток
	stockItem := &models.StockItem{
		ID:           uuid.New().String(),
		IngredientID: ingredient.ID,
		Quantity:     req.Quantity,
		UpdatedAt:    time.Now(),
	}

	// Сохраняем в БД
	if err := ingredientRepo.CreateIngredient(ingredient, stockItem); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to create ingredient")
		return
	}

	// Загружаем полные данные для ответа
	stockItem.Ingredient = ingredient

	utils.RespondWithJSON(w, http.StatusCreated, stockItem)
}

// UpdateIngredient обновление ингредиента
func UpdateIngredient(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var req models.UpdateIngredientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	// Находим складской остаток
	stockItem, err := ingredientRepo.FindByID(id)
	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "Ingredient not found")
		return
	}

	// Обновляем ингредиент
	if req.Name != "" {
		stockItem.Ingredient.Name = req.Name
	}
	if req.Unit != "" {
		stockItem.Ingredient.Unit = req.Unit
	}

	// Обновляем количество
	stockItem.Quantity = req.Quantity
	stockItem.UpdatedAt = time.Now()

	// Сохраняем в БД
	if err := ingredientRepo.UpdateIngredient(stockItem.Ingredient); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to update ingredient")
		return
	}

	if err := ingredientRepo.UpdateStockItem(stockItem); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to update stock")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, stockItem)
}

// DeleteIngredient удаление ингредиента
func DeleteIngredient(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if err := ingredientRepo.DeleteStockItem(id); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to delete ingredient")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Ingredient deleted successfully"})
}
