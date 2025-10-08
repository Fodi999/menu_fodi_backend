package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
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

	// Возвращаем полные StockItem объекты с вложенным Ingredient
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

	// Возвращаем полный StockItem с вложенным Ingredient
	utils.RespondWithJSON(w, http.StatusOK, stockItem)
}

// CreateIngredient создание нового ингредиента
func CreateIngredient(w http.ResponseWriter, r *http.Request) {
	var req models.CreateIngredientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	log.Printf("📥 CreateIngredient Request: %+v\n", req)

	// Проверяем обязательные поля
	if req.Name == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Name is required")
		return
	}
	if req.Unit == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Unit is required")
		return
	}

	// 🔍 Автоматическое определение единицы измерения по названию ингредиента
	if autoUnit := detectDefaultUnit(req.Name); autoUnit != "" {
		req.Unit = autoUnit
	}

	// Создаём сам ингредиент
	ingredient := &models.Ingredient{
		ID:        uuid.New().String(),
		Name:      req.Name,
		Unit:      req.Unit,
		CreatedAt: time.Now(),
	}

	// Генерируем уникальный номер партии
	batchNumber := generateBatchNumber(req.Name)

	// Создаём складскую запись
	stockItem := &models.StockItem{
		ID:           uuid.New().String(),
		IngredientID: ingredient.ID,
		Quantity:     req.Quantity,
		UpdatedAt:    time.Now(),
		BatchNumber:  &batchNumber,
	}

	// 💡 Правильная логика присвоения: создаём указатели на значения из req
	if req.BruttoWeight > 0 {
		val := req.BruttoWeight
		stockItem.BruttoWeight = &val
	}
	if req.NettoWeight > 0 {
		val := req.NettoWeight
		stockItem.NettoWeight = &val
	}
	if req.WastePercentage >= 0 {
		val := req.WastePercentage
		stockItem.WastePercentage = &val
	}
	if req.ExpiryDays > 0 {
		val := req.ExpiryDays
		stockItem.ExpiryDays = &val
	}
	if req.Supplier != "" {
		val := req.Supplier
		stockItem.Supplier = &val
	}
	if req.Category != "" {
		val := req.Category
		stockItem.Category = &val
	}
	if req.PriceBrutto > 0 {
		val := req.PriceBrutto
		stockItem.PriceBrutto = &val
	}
	if req.PriceNetto > 0 {
		val := req.PriceNetto
		stockItem.PriceNetto = &val
	}
	if req.PricePerUnit > 0 {
		val := req.PricePerUnit
		stockItem.PricePerUnit = &val
	}

	log.Printf("📦 StockItem before save (Batch: %s): %+v\n", batchNumber, stockItem)

	if err := ingredientRepo.CreateIngredient(ingredient, stockItem); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to create ingredient")
		return
	}

	// Создаем запись о начальном поступлении
	// Используем bruttoWeight или nettoWeight как количество для движения
	movementQuantity := req.Quantity
	if movementQuantity == 0 && req.BruttoWeight > 0 {
		movementQuantity = req.BruttoWeight
	}
	if movementQuantity == 0 && req.NettoWeight > 0 {
		movementQuantity = req.NettoWeight
	}

	if movementQuantity > 0 {
		movement := &models.StockMovement{
			ID:          uuid.New().String(),
			StockItemID: stockItem.ID,
			Type:        "addition",
			Quantity:    movementQuantity,
			PriceBrutto: stockItem.PriceBrutto,
			PriceNetto:  stockItem.PriceNetto,
			CreatedAt:   time.Now(),
		}
		note := "Начальное поступление"
		movement.Note = &note

		if err := database.DB.Create(movement).Error; err != nil {
			log.Printf("⚠️ Failed to create stock movement: %v", err)
		} else {
			log.Printf("✅ Created stock movement: %s for %.2f units", movement.ID, movementQuantity)
		}
	}

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
	} else if req.Name != "" {
		// 🔍 Автоопределение единицы измерения при изменении названия
		if autoUnit := detectDefaultUnit(req.Name); autoUnit != "" {
			stockItem.Ingredient.Unit = autoUnit
		}
	}

	// Обновляем складские данные
	stockItem.Quantity = req.Quantity
	if req.BruttoWeight > 0 {
		val := req.BruttoWeight
		stockItem.BruttoWeight = &val
	}
	if req.NettoWeight > 0 {
		val := req.NettoWeight
		stockItem.NettoWeight = &val
	}
	if req.WastePercentage >= 0 {
		val := req.WastePercentage
		stockItem.WastePercentage = &val
	}
	if req.ExpiryDays > 0 {
		val := req.ExpiryDays
		stockItem.ExpiryDays = &val
	}
	if req.Supplier != "" {
		val := req.Supplier
		stockItem.Supplier = &val
	}
	if req.Category != "" {
		val := req.Category
		stockItem.Category = &val
	}
	if req.PriceBrutto > 0 {
		val := req.PriceBrutto
		stockItem.PriceBrutto = &val
	}
	if req.PriceNetto > 0 {
		val := req.PriceNetto
		stockItem.PriceNetto = &val
	}
	if req.PricePerUnit > 0 {
		val := req.PricePerUnit
		stockItem.PricePerUnit = &val
	}
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

// GetStockMovements получение истории движений товара
func GetStockMovements(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	stockItemID := vars["id"]

	var movements []models.StockMovement
	result := database.DB.
		Where("\"stockItemId\" = ?", stockItemID).
		Order("\"createdAt\" DESC").
		Limit(20).
		Find(&movements)

	if result.Error != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch stock movements")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, movements)
}

// detectDefaultUnit возвращает дефолтную единицу измерения по названию ингредиента
func detectDefaultUnit(name string) string {
	nameLower := strings.ToLower(name)

	defaultUnits := map[string]string{
		"мук":    "kg",
		"сахар":  "kg",
		"рис":    "kg",
		"круп":   "kg",
		"соль":   "kg",
		"вода":   "l",
		"масло":  "ml",
		"молок":  "l",
		"яйц":    "pcs",
		"лосос":  "kg",
		"сёмг":   "kg",
		"тунец":  "kg",
		"креве":  "kg",
		"угор":   "kg",
		"сыр":    "kg",
		"соус":   "ml",
		"уксус":  "ml",
		"нори":   "pcs",
		"васаби": "kg",
		"имбир":  "kg",
		"авока":  "pcs",
		"огуре":  "pcs",
	}

	for key, unit := range defaultUnits {
		if strings.Contains(nameLower, key) {
			log.Printf("⚙️ Автоматически установлена единица '%s' для ингредиента '%s'", unit, name)
			return unit
		}
	}

	return "" // если не найдено — не меняем
}

// generateBatchNumber генерирует уникальный номер партии
func generateBatchNumber(ingredientName string) string {
	now := time.Now()

	// Берём первые 3 руны (символа) для поддержки UTF-8
	runes := []rune(ingredientName)
	prefix := ""
	if len(runes) >= 3 {
		prefix = string(runes[:3])
	} else {
		prefix = string(runes)
	}

	// Преобразуем в заглавные и убираем пробелы
	prefix = strings.ToUpper(strings.ReplaceAll(prefix, " ", ""))

	// Формат: КРЕ-20251006-020417 (3 буквы-дата-время)
	return fmt.Sprintf("%s-%s-%s",
		prefix,
		now.Format("20060102"),
		now.Format("150405"))
}
