package handlers

import (
	"encoding/json"
	"log"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/database"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
	"github.com/dmitrijfomin/menu-fodifood/backend/pkg/utils"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// normalizeFloat округляет число до указанного количества знаков после запятой
func normalizeFloat(value float64, decimals int) float64 {
	mult := math.Pow(10, float64(decimals))
	return math.Round(value*mult) / mult
}

// convertToBaseUnit конвертирует значение в базовую единицу измерения
// Граммы → кг, Миллилитры → литры
func convertToBaseUnit(value float64, unit string) float64 {
	switch strings.ToLower(unit) {
	case "g":
		return value / 1000
	case "ml":
		return value / 1000
	default:
		return value
	}
}

// calculateCostPerUnit рассчитывает себестоимость за единицу полуфабриката
// Учитывает конвертацию единиц измерения и защищает от деления на ноль
func calculateCostPerUnit(ingredients []models.SemiFinishedIngredientInput, outputQty float64) float64 {
	var totalCost float64
	for _, ing := range ingredients {
		qty := convertToBaseUnit(ing.Quantity, ing.Unit)
		totalCost += qty * ing.PricePerUnit
	}
	if outputQty == 0 {
		return 0
	}
	return normalizeFloat(totalCost/outputQty, 2)
}

// GetSemiFinished возвращает список всех полуфабрикатов
func GetSemiFinished(w http.ResponseWriter, r *http.Request) {
	var semiFinished []models.SemiFinished

	// Не загружаем связанные ингредиенты для списка, только для GetByID
	if err := database.DB.Omit("Ingredients").Order("created_at DESC").Find(&semiFinished).Error; err != nil {
		log.Printf("Error fetching semi-finished: %v", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch semi-finished")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, semiFinished)
}

// GetSemiFinishedByID возвращает полуфабрикат по ID с ингредиентами
func GetSemiFinishedByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var sf models.SemiFinished
	if err := database.DB.First(&sf, "id = ?", id).Error; err != nil {
		if err.Error() == "record not found" {
			utils.RespondWithError(w, http.StatusNotFound, "Semi-finished not found")
			return
		}
		log.Printf("Error fetching semi-finished: %v", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch semi-finished")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, sf)
}

// CreateSemiFinished создаёт новый полуфабрикат
func CreateSemiFinished(w http.ResponseWriter, r *http.Request) {
	var req models.CreateSemiFinishedRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Валидация
	if req.Name == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Name is required")
		return
	}
	if req.OutputQuantity <= 0 {
		utils.RespondWithError(w, http.StatusBadRequest, "Output quantity must be positive")
		return
	}
	if req.OutputUnit == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Output unit is required")
		return
	}
	if len(req.Ingredients) == 0 {
		utils.RespondWithError(w, http.StatusBadRequest, "At least one ingredient is required")
		return
	}

	// Валидация ингредиентов
	for i, ing := range req.Ingredients {
		if ing.IngredientID == "" {
			utils.RespondWithError(w, http.StatusBadRequest, "Ingredient ID is required for all ingredients")
			return
		}
		if ing.IngredientName == "" {
			utils.RespondWithError(w, http.StatusBadRequest, "Ingredient name is required for all ingredients")
			return
		}
		if ing.Unit == "" {
			utils.RespondWithError(w, http.StatusBadRequest, "Unit is required for all ingredients")
			return
		}
		if ing.Quantity <= 0 {
			utils.RespondWithError(w, http.StatusBadRequest, "Ingredient quantity must be positive")
			return
		}

		// Проверяем существование ингредиента в БД
		var existingIngredient models.Ingredient
		if err := database.DB.First(&existingIngredient, "id = ?", ing.IngredientID).Error; err != nil {
			log.Printf("Ingredient #%d not found: ID=%s, Error=%v", i+1, ing.IngredientID, err)
			utils.RespondWithError(w, http.StatusBadRequest, "Ingredient with ID '"+ing.IngredientID+"' does not exist")
			return
		}
	}

	// Проверка на дубликаты названий
	var exists int64
	database.DB.Model(&models.SemiFinished{}).
		Where("LOWER(name) = LOWER(?)", req.Name).
		Count(&exists)
	if exists > 0 {
		utils.RespondWithError(w, http.StatusConflict, "Semi-finished with this name already exists")
		return
	}

	// Нормализация числовых значений
	for i := range req.Ingredients {
		req.Ingredients[i].Quantity = normalizeFloat(req.Ingredients[i].Quantity, 3)
		req.Ingredients[i].PricePerUnit = normalizeFloat(req.Ingredients[i].PricePerUnit, 2)
		req.Ingredients[i].TotalPrice = normalizeFloat(req.Ingredients[i].TotalPrice, 2)
	}
	req.OutputQuantity = normalizeFloat(req.OutputQuantity, 3)

	// Рассчитываем себестоимость за единицу с учётом конвертации единиц
	costPerUnit := calculateCostPerUnit(req.Ingredients, req.OutputQuantity)
	totalCost := normalizeFloat(costPerUnit*req.OutputQuantity, 2)

	// Создаём полуфабрикат
	id := uuid.New().String()
	now := time.Now()

	var description *string
	if req.Description != "" {
		description = &req.Description
	}

	sf := models.SemiFinished{
		ID:             id,
		Name:           req.Name,
		Description:    description,
		OutputQuantity: req.OutputQuantity,
		OutputUnit:     req.OutputUnit,
		CostPerUnit:    costPerUnit,
		TotalCost:      totalCost,
		Category:       req.Category,
		IsVisible:      true,
		IsArchived:     false,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	// Начинаем транзакцию
	tx := database.DB.Begin()
	if tx.Error != nil {
		log.Printf("Error starting transaction: %v", tx.Error)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to create semi-finished")
		return
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Создаём полуфабрикат
	if err := tx.Create(&sf).Error; err != nil {
		tx.Rollback()
		log.Printf("Error creating semi-finished: %v", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to create semi-finished")
		return
	}

	// Добавляем ингредиенты
	for _, ing := range req.Ingredients {
		log.Printf("Adding ingredient: ID=%s, Name=%s, Qty=%.3f %s, Price=%.2f",
			ing.IngredientID, ing.IngredientName, ing.Quantity, ing.Unit, ing.PricePerUnit)

		ingredient := models.SemiFinishedIngredient{
			ID:             uuid.New().String(),
			SemiFinishedID: id,
			IngredientID:   ing.IngredientID,
			IngredientName: ing.IngredientName,
			Quantity:       ing.Quantity,
			Unit:           ing.Unit,
			PricePerUnit:   ing.PricePerUnit,
			TotalPrice:     ing.TotalPrice,
		}
		if err := tx.Create(&ingredient).Error; err != nil {
			tx.Rollback()
			log.Printf("Error adding ingredient to semi-finished: %v", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to add ingredients")
			return
		}
	}

	// Коммитим транзакцию
	if err := tx.Commit().Error; err != nil {
		log.Printf("Error committing transaction: %v", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to save semi-finished")
		return
	}

	log.Printf("📦 Semi-finished '%s' создан: %.3f %s, себестоимость = %.2f ₽/ед, всего = %.2f ₽ (ID: %s)",
		req.Name, req.OutputQuantity, req.OutputUnit, sf.CostPerUnit, sf.TotalCost, id)
	utils.RespondWithJSON(w, http.StatusCreated, sf)
}

// UpdateSemiFinished обновляет полуфабрикат
func UpdateSemiFinished(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var req models.UpdateSemiFinishedRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Проверяем существование полуфабриката
	var sf models.SemiFinished
	if err := database.DB.First(&sf, "id = ?", id).Error; err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "Semi-finished not found")
		return
	}

	// Валидация и обновление полей
	if req.Name != nil && *req.Name != "" {
		// Проверка на дубликаты названий (кроме текущего)
		var exists int64
		database.DB.Model(&models.SemiFinished{}).
			Where("LOWER(name) = LOWER(?) AND id != ?", *req.Name, id).
			Count(&exists)
		if exists > 0 {
			utils.RespondWithError(w, http.StatusConflict, "Semi-finished with this name already exists")
			return
		}
		sf.Name = *req.Name
	}

	if req.Description != nil {
		sf.Description = req.Description
	}

	if req.Category != nil {
		sf.Category = *req.Category
	}

	if req.OutputUnit != nil {
		sf.OutputUnit = *req.OutputUnit
	}

	// Обработка ингредиентов и расчёт себестоимости
	if len(req.Ingredients) > 0 {
		// Нормализация числовых значений
		for i := range req.Ingredients {
			req.Ingredients[i].Quantity = normalizeFloat(req.Ingredients[i].Quantity, 3)
			req.Ingredients[i].PricePerUnit = normalizeFloat(req.Ingredients[i].PricePerUnit, 2)
			req.Ingredients[i].TotalPrice = normalizeFloat(req.Ingredients[i].TotalPrice, 2)
		}
	}

	// Если передано новое количество продукции или новые ингредиенты, пересчитываем себестоимость
	outputQty := sf.OutputQuantity
	if req.OutputQuantity != nil && *req.OutputQuantity > 0 {
		outputQty = normalizeFloat(*req.OutputQuantity, 3)
		sf.OutputQuantity = outputQty
	}

	// Пересчитываем себестоимость если переданы новые ингредиенты
	if len(req.Ingredients) > 0 {
		costPerUnit := calculateCostPerUnit(req.Ingredients, outputQty)
		sf.CostPerUnit = costPerUnit
		sf.TotalCost = normalizeFloat(costPerUnit*outputQty, 2)
	} else if req.OutputQuantity != nil {
		// Если только количество изменилось, пересчитываем общую стоимость
		sf.TotalCost = normalizeFloat(sf.CostPerUnit*outputQty, 2)
	}
	sf.UpdatedAt = time.Now()

	// Начинаем транзакцию
	tx := database.DB.Begin()
	if tx.Error != nil {
		log.Printf("Error starting transaction: %v", tx.Error)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to update semi-finished")
		return
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Обновляем полуфабрикат
	if err := tx.Save(&sf).Error; err != nil {
		tx.Rollback()
		log.Printf("Error updating semi-finished: %v", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to update semi-finished")
		return
	}

	// Удаляем старые ингредиенты и добавляем новые только если они переданы
	if len(req.Ingredients) > 0 {
		// Удаляем старые ингредиенты
		if err := tx.Where("semi_finished_id = ?", id).Delete(&models.SemiFinishedIngredient{}).Error; err != nil {
			tx.Rollback()
			log.Printf("Error deleting old ingredients: %v", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to update ingredients")
			return
		}

		// Добавляем новые ингредиенты
		for _, ing := range req.Ingredients {
			ingredient := models.SemiFinishedIngredient{
				ID:             uuid.New().String(),
				SemiFinishedID: id,
				IngredientID:   ing.IngredientID,
				IngredientName: ing.IngredientName,
				Quantity:       ing.Quantity,
				Unit:           ing.Unit,
				PricePerUnit:   ing.PricePerUnit,
				TotalPrice:     ing.TotalPrice,
			}
			if err := tx.Create(&ingredient).Error; err != nil {
				tx.Rollback()
				log.Printf("Error adding ingredient: %v", err)
				utils.RespondWithError(w, http.StatusInternalServerError, "Failed to update ingredients")
				return
			}
		}
	}

	// Коммитим транзакцию
	if err := tx.Commit().Error; err != nil {
		log.Printf("Error committing transaction: %v", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to save changes")
		return
	}

	log.Printf("📦 Semi-finished '%s' обновлён: %.3f %s, себестоимость = %.2f ₽/ед, всего = %.2f ₽ (ID: %s)",
		sf.Name, sf.OutputQuantity, sf.OutputUnit, sf.CostPerUnit, sf.TotalCost, id)
	utils.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Semi-finished updated successfully"})
}

// DeleteSemiFinished удаляет полуфабрикат
func DeleteSemiFinished(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Начинаем транзакцию
	tx := database.DB.Begin()
	if tx.Error != nil {
		log.Printf("Error starting transaction: %v", tx.Error)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to delete semi-finished")
		return
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Удаляем ингредиенты
	if err := tx.Where("semi_finished_id = ?", id).Delete(&models.SemiFinishedIngredient{}).Error; err != nil {
		tx.Rollback()
		log.Printf("Error deleting ingredients: %v", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to delete semi-finished")
		return
	}

	// Удаляем полуфабрикат
	result := tx.Delete(&models.SemiFinished{}, "id = ?", id)
	if result.Error != nil {
		tx.Rollback()
		log.Printf("Error deleting semi-finished: %v", result.Error)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to delete semi-finished")
		return
	}

	if result.RowsAffected == 0 {
		tx.Rollback()
		utils.RespondWithError(w, http.StatusNotFound, "Semi-finished not found")
		return
	}

	// Коммитим транзакцию
	if err := tx.Commit().Error; err != nil {
		log.Printf("Error committing transaction: %v", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to delete semi-finished")
		return
	}

	log.Printf("✅ Semi-finished deleted: ID %s", id)
	utils.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Semi-finished deleted successfully"})
}
