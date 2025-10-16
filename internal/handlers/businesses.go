package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/database"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/services"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

var businessTokenService = services.NewTokenService()

// CreateBusinessRequest запрос на создание бизнеса
type CreateBusinessRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Category    string `json:"category"`
	City        string `json:"city"`
	OwnerID     string `json:"owner_id"`
}

// 📋 GET /api/businesses
func GetBusinesses(w http.ResponseWriter, r *http.Request) {
	db := database.GetDB()
	var businesses []models.Business

	if err := db.Order("created_at DESC").Find(&businesses).Error; err != nil {
		http.Error(w, "Failed to fetch businesses", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(businesses)
}

// ➕ POST /api/businesses/create
func CreateBusiness(w http.ResponseWriter, r *http.Request) {
	db := database.GetDB()

	var input CreateBusinessRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if input.Name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}

	business := models.Business{
		ID:          uuid.New().String(),
		Name:        input.Name,
		Description: input.Description,
		Category:    input.Category,
		City:        input.City,
		OwnerID:     input.OwnerID,
		IsActive:    true,
	}

	if err := db.Create(&business).Error; err != nil {
		log.Printf("[BUSINESS] ❌ Error creating business: %v", err)
		http.Error(w, "Failed to create business", http.StatusInternalServerError)
		return
	}

	log.Printf("[BUSINESS] ✅ Created ID=%s, Name=%s, Owner=%s", business.ID, business.Name, business.OwnerID)

	// Автоматическое создание первоначального токена
	token, err := businessTokenService.MintInitialToken(business.ID)
	if err != nil {
		log.Printf("[BUSINESS] ⚠️ Warning: Failed to create initial token: %v", err)
		// Не прерываем процесс, бизнес уже создан
	} else {
		log.Printf("[BUSINESS] 🪙 Initial token created: Symbol=%s, Price=$%.2f", token.Symbol, token.Price)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":  "✅ Business created successfully",
		"business": business,
		"token":    token,
	})
}

// 🔍 GET /api/businesses/{id}
func GetBusinessByID(w http.ResponseWriter, r *http.Request) {
	db := database.GetDB()
	vars := mux.Vars(r)
	businessID := vars["id"]

	var business models.Business
	if err := db.First(&business, "id = ?", businessID).Error; err != nil {
		log.Printf("[BUSINESS] ❌ Business not found: ID=%s", businessID)
		http.Error(w, "Business not found", http.StatusNotFound)
		return
	}

	log.Printf("[BUSINESS] ✅ Retrieved ID=%s, Name=%s", business.ID, business.Name)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(business)
}

// ✏️ PUT /api/businesses/{id}
func UpdateBusiness(w http.ResponseWriter, r *http.Request) {
	db := database.GetDB()
	vars := mux.Vars(r)
	businessID := vars["id"]

	// Проверка существования бизнеса
	var business models.Business
	if err := db.First(&business, "id = ?", businessID).Error; err != nil {
		log.Printf("[BUSINESS] ❌ Business not found for update: ID=%s", businessID)
		http.Error(w, "Business not found", http.StatusNotFound)
		return
	}

	// Парсинг обновлений
	var input struct {
		Name        *string `json:"name"`
		Description *string `json:"description"`
		Category    *string `json:"category"`
		City        *string `json:"city"`
		IsActive    *bool   `json:"isActive"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Обновление полей (только если переданы)
	updates := make(map[string]interface{})
	
	if input.Name != nil {
		if *input.Name == "" {
			http.Error(w, "Name cannot be empty", http.StatusBadRequest)
			return
		}
		updates["name"] = *input.Name
	}
	if input.Description != nil {
		updates["description"] = *input.Description
	}
	if input.Category != nil {
		updates["category"] = *input.Category
	}
	if input.City != nil {
		updates["city"] = *input.City
	}
	if input.IsActive != nil {
		updates["is_active"] = *input.IsActive
	}

	// Применение обновлений
	if err := db.Model(&business).Updates(updates).Error; err != nil {
		log.Printf("[BUSINESS] ❌ Error updating business: %v", err)
		http.Error(w, "Failed to update business", http.StatusInternalServerError)
		return
	}

	// Получение обновленных данных
	db.First(&business, "id = ?", businessID)

	log.Printf("[BUSINESS] ✅ Updated ID=%s, Name=%s", business.ID, business.Name)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":  "✅ Business updated successfully",
		"business": business,
	})
}

// 🗑️ DELETE /api/businesses/{id}
func DeleteBusiness(w http.ResponseWriter, r *http.Request) {
	db := database.GetDB()
	vars := mux.Vars(r)
	businessID := vars["id"]

	// Проверка существования бизнеса
	var business models.Business
	if err := db.First(&business, "id = ?", businessID).Error; err != nil {
		log.Printf("[BUSINESS] ❌ Business not found for deletion: ID=%s", businessID)
		http.Error(w, "Business not found", http.StatusNotFound)
		return
	}

	// Мягкое удаление (деактивация вместо полного удаления)
	// Для полного удаления можно использовать: db.Delete(&business)
	if err := db.Model(&business).Update("is_active", false).Error; err != nil {
		log.Printf("[BUSINESS] ❌ Error deactivating business: %v", err)
		http.Error(w, "Failed to delete business", http.StatusInternalServerError)
		return
	}

	log.Printf("[BUSINESS] ✅ Deactivated (soft delete) ID=%s, Name=%s", business.ID, business.Name)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "✅ Business deactivated successfully",
		"id":      businessID,
	})
}

// 🗑️ DELETE /api/businesses/{id}/permanent (опциональный - для полного удаления)
func PermanentDeleteBusiness(w http.ResponseWriter, r *http.Request) {
	db := database.GetDB()
	vars := mux.Vars(r)
	businessID := vars["id"]

	// Проверка существования бизнеса
	var business models.Business
	if err := db.First(&business, "id = ?", businessID).Error; err != nil {
		log.Printf("[BUSINESS] ❌ Business not found for permanent deletion: ID=%s", businessID)
		http.Error(w, "Business not found", http.StatusNotFound)
		return
	}

	// Полное удаление из базы данных
	if err := db.Delete(&business).Error; err != nil {
		log.Printf("[BUSINESS] ❌ Error permanently deleting business: %v", err)
		http.Error(w, "Failed to permanently delete business", http.StatusInternalServerError)
		return
	}

	log.Printf("[BUSINESS] ✅ Permanently deleted ID=%s, Name=%s", business.ID, business.Name)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "✅ Business permanently deleted",
		"id":      businessID,
	})
}
