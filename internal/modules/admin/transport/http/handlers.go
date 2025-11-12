package http

import (
	"encoding/json"
	"net/http"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/middleware"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/admin/service"
	authservice "github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/auth/service"
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

// GetAdminDashboard возвращает aggregated dashboard data для админ панели
func (h *AdminHandlers) GetAdminDashboard(w http.ResponseWriter, r *http.Request) {
	// Извлекаем claims из контекста
	claims, ok := r.Context().Value(middleware.UserContextKey).(*authservice.Claims)
	if !ok || claims == nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Получаем профиль админа
	profile, err := h.service.GetAdminProfile(claims.UserID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch admin profile")
		return
	}

	// Получаем статистику
	stats, err := h.service.GetAdminStats()
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch stats")
		return
	}

	// Получаем последние заказы (graceful degradation если ошибка)
	var recentOrders interface{} = []interface{}{}
	if orders, err := h.service.GetRecentOrders(5); err == nil {
		recentOrders = orders
	}

	// Получаем информацию о токенах (graceful degradation если ошибка)
	var tokenStats interface{} = nil
	if ts, err := h.service.GetTokenBankStats(); err == nil {
		tokenStats = ts
	}

	// Формируем dashboard response
	dashboard := map[string]interface{}{
		"admin":        profile,
		"stats":        stats,
		"recentOrders": recentOrders,
		"tokenStats":   tokenStats,
	}

	utils.RespondWithJSON(w, http.StatusOK, dashboard)
}

// GetAdminProfile возвращает профиль текущего администратора с управляемыми ресурсами
func (h *AdminHandlers) GetAdminProfile(w http.ResponseWriter, r *http.Request) {
	// Извлекаем claims из контекста (устанавливается AuthMiddleware)
	claims, ok := r.Context().Value(middleware.UserContextKey).(*authservice.Claims)
	if !ok || claims == nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	profile, err := h.service.GetAdminProfile(claims.UserID)
	if err != nil {
		if err.Error() == "admin not found" {
			utils.RespondWithError(w, http.StatusNotFound, "Admin not found")
		} else {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch admin profile")
		}
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, profile)
}

// Token Bank Handlers

// GetAllTokenBanks возвращает все записи токин-банков пользователей
func (h *AdminHandlers) GetAllTokenBanks(w http.ResponseWriter, r *http.Request) {
	tokenBanks, err := h.service.GetAllTokenBanks()
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch token banks")
		return
	}
	utils.RespondWithJSON(w, http.StatusOK, tokenBanks)
}

// GetTokenBankStats возвращает статистику по токинам
func (h *AdminHandlers) GetTokenBankStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.service.GetTokenBankStats()
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch token bank stats")
		return
	}
	utils.RespondWithJSON(w, http.StatusOK, stats)
}

// GetUserTokenBank возвращает токин-банк конкретного пользователя
func (h *AdminHandlers) GetUserTokenBank(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")

	tokenBank, err := h.service.GetTokenBankByUserID(userID)
	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "Token bank not found")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, tokenBank)
}

// AllocateTokens выделяет токины пользователю
func (h *AdminHandlers) AllocateTokens(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID string `json:"user_id"`
		Amount int64  `json:"amount"`
		Reason string `json:"reason,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	if req.UserID == "" || req.Amount <= 0 {
		utils.RespondWithError(w, http.StatusBadRequest, "user_id and amount are required; amount must be positive")
		return
	}

	err := h.service.AllocateTokens(req.UserID, req.Amount)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to allocate tokens: "+err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Tokens allocated successfully",
		"user_id": req.UserID,
		"amount":  req.Amount,
	})
}

// RevokeTokens отзывает токины у пользователя
func (h *AdminHandlers) RevokeTokens(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID string `json:"user_id"`
		Amount int64  `json:"amount"`
		Reason string `json:"reason,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	if req.UserID == "" || req.Amount <= 0 {
		utils.RespondWithError(w, http.StatusBadRequest, "user_id and amount are required; amount must be positive")
		return
	}

	err := h.service.RevokeTokens(req.UserID, req.Amount)
	if err != nil {
		switch err.Error() {
		case "insufficient tokens":
			utils.RespondWithError(w, http.StatusBadRequest, "Insufficient tokens to revoke")
		case "token bank not found for user":
			utils.RespondWithError(w, http.StatusNotFound, "Token bank not found for user")
		default:
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to revoke tokens: "+err.Error())
		}
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Tokens revoked successfully",
		"user_id": req.UserID,
		"amount":  req.Amount,
	})
}

// SetTokenBalance устанавливает точное значение баланса токинов
func (h *AdminHandlers) SetTokenBalance(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID  string `json:"user_id"`
		Balance int64  `json:"balance"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	if req.UserID == "" || req.Balance < 0 {
		utils.RespondWithError(w, http.StatusBadRequest, "user_id is required and balance must be non-negative")
		return
	}

	err := h.service.SetTokenBalance(req.UserID, req.Balance)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to set token balance: "+err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Token balance set successfully",
		"user_id": req.UserID,
		"balance": req.Balance,
	})
}

