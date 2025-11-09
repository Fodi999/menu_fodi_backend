package http

import (
	"encoding/json"
	"net/http"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/database"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
	"github.com/dmitrijfomin/menu-fodifood/backend/pkg/utils"
	"github.com/go-chi/chi/v5"
)

type AdminHandlers struct{}

func NewAdminHandlers() *AdminHandlers {
	return &AdminHandlers{}
}

func (h *AdminHandlers) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	db := database.GetDB()
	var users []models.User
	if err := db.Find(&users).Error; err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch users")
		return
	}
	utils.RespondWithJSON(w, http.StatusOK, users)
}

func (h *AdminHandlers) UpdateUser(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")
	db := database.GetDB()

	var user models.User
	if err := db.First(&user, "id = ?", userID).Error; err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "User not found")
		return
	}

	var req struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Email != "" {
		user.Email = req.Email
	}

	if err := db.Save(&user).Error; err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to update user")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, user)
}

func (h *AdminHandlers) DeleteUser(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")
	db := database.GetDB()

	if err := db.Delete(&models.User{}, "id = ?", userID).Error; err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "User not found")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "User deleted successfully"})
}

func (h *AdminHandlers) UpdateUserRole(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID string `json:"user_id"`
		Role   string `json:"role"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	db := database.GetDB()
	if err := db.Model(&models.User{}).Where("id = ?", req.UserID).Update("role", req.Role).Error; err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to update role")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Role updated successfully"})
}

func (h *AdminHandlers) GetAllOrders(w http.ResponseWriter, r *http.Request) {
	db := database.GetDB()
	var orders []models.Order
	if err := db.Order("created_at DESC").Find(&orders).Error; err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch orders")
		return
	}
	utils.RespondWithJSON(w, http.StatusOK, orders)
}

func (h *AdminHandlers) GetRecentOrders(w http.ResponseWriter, r *http.Request) {
	db := database.GetDB()
	var orders []models.Order
	if err := db.Order("created_at DESC").Limit(10).Find(&orders).Error; err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch recent orders")
		return
	}
	utils.RespondWithJSON(w, http.StatusOK, orders)
}

func (h *AdminHandlers) UpdateOrderStatus(w http.ResponseWriter, r *http.Request) {
	orderID := chi.URLParam(r, "id")
	db := database.GetDB()

	var req struct {
		Status string `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	var order models.Order
	if err := db.Model(&order).Where("id = ?", orderID).Update("status", req.Status).Error; err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to update order status")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Order status updated"})
}

func (h *AdminHandlers) GetAdminStats(w http.ResponseWriter, r *http.Request) {
	db := database.GetDB()

	var userCount, orderCount int64
	db.Model(&models.User{}).Count(&userCount)
	db.Model(&models.Order{}).Count(&orderCount)

	stats := map[string]interface{}{
		"totalUsers":  userCount,
		"totalOrders": orderCount,
	}

	utils.RespondWithJSON(w, http.StatusOK, stats)
}
