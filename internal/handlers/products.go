package handlers

import (
	"encoding/json"
	"log"
	"math"
	"net/http"
	"strings"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/database"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// GetAllProducts получить все продукты (для админки)
func GetAllProducts(w http.ResponseWriter, r *http.Request) {
	var products []models.Product

	if err := database.DB.Order(`"createdAt" DESC`).Find(&products).Error; err != nil {
		http.Error(w, "Failed to fetch products", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(products)
}

// GetPublicProducts получить только видимые продукты (для главной страницы)
func GetPublicProducts(w http.ResponseWriter, r *http.Request) {
	var products []models.Product

	// Фильтруем только видимые продукты
	if err := database.DB.Where(`"isVisible" = ?`, true).Order(`"createdAt" DESC`).Find(&products).Error; err != nil {
		http.Error(w, "Failed to fetch products", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(products)
}

// GetProduct получить один продукт по ID
func GetProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	productID := vars["id"]

	var product models.Product
	if err := database.DB.
		Preload("Ingredients").
		Preload("SemiFinished").
		Where("id = ?", productID).
		First(&product).Error; err != nil {
		http.Error(w, "Product not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(product)
}

// CreateProduct создать новый продукт
func CreateProduct(w http.ResponseWriter, r *http.Request) {
	var req models.CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Валидация
	if req.Name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}
	if len(req.Name) > 100 {
		http.Error(w, "Name too long (max 100 characters)", http.StatusBadRequest)
		return
	}
	if req.Price < 0 {
		http.Error(w, "Price must be positive", http.StatusBadRequest)
		return
	}

	// Нормализация цены
	req.Price = normalizeProductFloat(req.Price, 2)

	// Автоопределение категории, если не указана
	if req.Category == "" {
		detectedCategory := detectCategory(req.Name)
		if detectedCategory != "" {
			req.Category = detectedCategory
			log.Printf("⚙️ Автоматически установлена категория '%s' для продукта '%s'", detectedCategory, req.Name)
		} else {
			http.Error(w, "Category is required", http.StatusBadRequest)
			return
		}
	}

	// Проверка на дубликаты
	var exists int64
	database.DB.Model(&models.Product{}).
		Where("LOWER(name) = LOWER(?) AND LOWER(category) = LOWER(?)", req.Name, req.Category).
		Count(&exists)
	if exists > 0 {
		http.Error(w, "Product with this name already exists in this category", http.StatusConflict)
		return
	}

	// Создание продукта
	productID := uuid.New().String()
	product := models.Product{
		ID:          productID,
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		ImageURL:    req.ImageURL,
		Weight:      req.Weight,
		Category:    req.Category,
		IsVisible:   req.IsVisible,
	}

	// Начинаем транзакцию
	tx := database.DB.Begin()
	if tx.Error != nil {
		log.Printf("Error starting transaction: %v", tx.Error)
		http.Error(w, "Failed to create product", http.StatusInternalServerError)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Создаём продукт
	if err := tx.Create(&product).Error; err != nil {
		tx.Rollback()
		log.Printf("❌ Failed to create product: %v", err)
		http.Error(w, "Failed to create product", http.StatusInternalServerError)
		return
	}

	// Добавляем ингредиенты, если есть
	for _, ing := range req.Ingredients {
		ingredient := models.ProductIngredient{
			ID:             uuid.New().String(),
			ProductID:      productID,
			IngredientID:   ing.IngredientID,
			IngredientName: ing.IngredientName,
			Quantity:       normalizeProductFloat(ing.Quantity, 3),
			Unit:           ing.Unit,
			PricePerUnit:   normalizeProductFloat(ing.PricePerUnit, 2),
			TotalPrice:     normalizeProductFloat(ing.TotalPrice, 2),
		}
		if err := tx.Create(&ingredient).Error; err != nil {
			tx.Rollback()
			log.Printf("Error adding ingredient: %v", err)
			http.Error(w, "Failed to add ingredients", http.StatusInternalServerError)
			return
		}
	}

	// Добавляем полуфабрикаты, если есть
	for _, sf := range req.SemiFinished {
		semiFinished := models.ProductSemiFinished{
			ID:               uuid.New().String(),
			ProductID:        productID,
			SemiFinishedID:   sf.SemiFinishedID,
			SemiFinishedName: sf.SemiFinishedName,
			Quantity:         normalizeProductFloat(sf.Quantity, 3),
			Unit:             sf.Unit,
			CostPerUnit:      normalizeProductFloat(sf.CostPerUnit, 2),
			TotalCost:        normalizeProductFloat(sf.TotalCost, 2),
		}
		if err := tx.Create(&semiFinished).Error; err != nil {
			tx.Rollback()
			log.Printf("Error adding semi-finished: %v", err)
			http.Error(w, "Failed to add semi-finished products", http.StatusInternalServerError)
			return
		}
	}

	// Коммитим транзакцию
	if err := tx.Commit().Error; err != nil {
		log.Printf("Error committing transaction: %v", err)
		http.Error(w, "Failed to save product", http.StatusInternalServerError)
		return
	}

	log.Printf("✅ Product created: %s (%.2f ₽, %s) with %d ingredients and %d semi-finished",
		product.Name, product.Price, product.Category, len(req.Ingredients), len(req.SemiFinished))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(product)
}

// UpdateProduct обновить продукт
func UpdateProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	productID := vars["id"]

	var req models.UpdateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Валидация
	if req.Name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}
	if len(req.Name) > 100 {
		http.Error(w, "Name too long (max 100 characters)", http.StatusBadRequest)
		return
	}
	if req.Price < 0 {
		http.Error(w, "Price must be positive", http.StatusBadRequest)
		return
	}

	// Нормализация цены
	req.Price = normalizeProductFloat(req.Price, 2)

	// Проверка существования
	var product models.Product
	if err := database.DB.Where("id = ?", productID).First(&product).Error; err != nil {
		http.Error(w, "Product not found", http.StatusNotFound)
		return
	}

	// Обновление полей
	product.Name = req.Name
	product.Description = req.Description
	product.Price = req.Price
	product.ImageURL = req.ImageURL
	product.Weight = req.Weight
	product.Category = req.Category

	// Обновление isVisible, если передано
	if req.IsVisible != nil {
		product.IsVisible = *req.IsVisible
	}

	if err := database.DB.Save(&product).Error; err != nil {
		log.Printf("❌ Failed to update product: %v", err)
		http.Error(w, "Failed to update product", http.StatusInternalServerError)
		return
	}

	log.Printf("✅ Product updated: %s (%.2f ₽, %s)", product.Name, product.Price, product.Category)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(product)
}

// DeleteProduct удалить продукт
func DeleteProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	productID := vars["id"]

	// Проверка существования
	var product models.Product
	if err := database.DB.Where("id = ?", productID).First(&product).Error; err != nil {
		http.Error(w, "Product not found", http.StatusNotFound)
		return
	}

	// Удаление
	if err := database.DB.Delete(&product).Error; err != nil {
		http.Error(w, "Failed to delete product", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Product deleted successfully"})
}

// normalizeProductFloat округляет число до указанного количества знаков
func normalizeProductFloat(value float64, decimals int) float64 {
	mult := math.Pow(10, float64(decimals))
	return math.Round(value*mult) / mult
}

// detectCategory автоматически определяет категорию по названию продукта
func detectCategory(name string) string {
	nameLower := strings.ToLower(name)

	if strings.Contains(nameLower, "ролл") {
		return "Роллы"
	} else if strings.Contains(nameLower, "суши") || strings.Contains(nameLower, "нигири") {
		return "Суши"
	} else if strings.Contains(nameLower, "салат") {
		return "Салаты"
	} else if strings.Contains(nameLower, "напиток") || strings.Contains(nameLower, "сок") || strings.Contains(nameLower, "чай") {
		return "Напитки"
	} else if strings.Contains(nameLower, "суп") || strings.Contains(nameLower, "мисо") {
		return "Супы"
	}

	return ""
}
