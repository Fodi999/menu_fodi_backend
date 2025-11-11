package http

import (
	"encoding/json"
	"net/http"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/admin/service"
	"github.com/dmitrijfomin/menu-fodifood/backend/pkg/utils"
	"github.com/go-chi/chi/v5"
)

type AdminHandlers struct {
	service service.AdminService
	policy  service.AdminPolicy
}

func NewAdminHandlers(svc service.AdminService, pol service.AdminPolicy) *AdminHandlers {
	return &AdminHandlers{
		service: svc,
		policy:  pol,
	}
}

func (h *AdminHandlers) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.service.GetAllUsers()
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch users")
		return
	}
	utils.RespondWithJSON(w, http.StatusOK, users)
}

func (h *AdminHandlers) UpdateUser(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")

	var req struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	user, err := h.service.UpdateUser(userID, req.Name, req.Email)
	if err != nil {
		if err.Error() == "user not found" {
			utils.RespondWithError(w, http.StatusNotFound, "User not found")
		} else {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to update user")
		}
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, user)
}

func (h *AdminHandlers) DeleteUser(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")

	err := h.service.DeleteUser(userID)
	if err != nil {
		if err.Error() == "user not found" {
			utils.RespondWithError(w, http.StatusNotFound, "User not found")
		} else {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to delete user")
		}
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

	err := h.service.UpdateUserRole(req.UserID, req.Role)
	if err != nil {
		switch err.Error() {
		case "user not found":
			utils.RespondWithError(w, http.StatusNotFound, "User not found")
		case "invalid role":
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid role")
		default:
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to update role")
		}
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Role updated successfully"})
}

func (h *AdminHandlers) GetAllOrders(w http.ResponseWriter, r *http.Request) {
	orders, err := h.service.GetAllOrders()
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch orders")
		return
	}
	utils.RespondWithJSON(w, http.StatusOK, orders)
}

func (h *AdminHandlers) GetRecentOrders(w http.ResponseWriter, r *http.Request) {
	orders, err := h.service.GetRecentOrders(10)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch recent orders")
		return
	}
	utils.RespondWithJSON(w, http.StatusOK, orders)
}

func (h *AdminHandlers) UpdateOrderStatus(w http.ResponseWriter, r *http.Request) {
	orderID := chi.URLParam(r, "id")

	var req struct {
		Status string `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	err := h.service.UpdateOrderStatus(orderID, req.Status)
	if err != nil {
		if err.Error() == "order not found" {
			utils.RespondWithError(w, http.StatusNotFound, "Order not found")
		} else {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to update order status")
		}
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Order status updated"})
}

func (h *AdminHandlers) GetAdminStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.service.GetAdminStats()
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch stats")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, stats)
}
