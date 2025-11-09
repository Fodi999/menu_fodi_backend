package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/middleware"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/wallet/dto"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/wallet/service"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/platform/httpx"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/platform/logger"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// WalletHandler handles HTTP requests for wallet operations
type WalletHandler struct {
	service *service.WalletService
}

// NewWalletHandler creates a new wallet handler
func NewWalletHandler(svc *service.WalletService) *WalletHandler {
	return &WalletHandler{
		service: svc,
	}
}

// GetBalance handles GET /api/wallet/balance
func (h *WalletHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	userIDPtr := middleware.GetUserID(r)
	if userIDPtr == nil {
		httpx.Unauthorized(w, "Unauthorized")
		return
	}
	userID := *userIDPtr

	balance, err := h.service.GetBalance(userID)
	if err != nil {
		logger.Error("Failed to get balance", zap.Error(err), zap.String("userId", userID.String()))
		httpx.InternalError(w, "Failed to get balance")
		return
	}

	httpx.Success(w, balance)
}

// PurchaseTokens handles POST /api/wallet/purchase
func (h *WalletHandler) PurchaseTokens(w http.ResponseWriter, r *http.Request) {
	userIDPtr := middleware.GetUserID(r)
	if userIDPtr == nil {
		httpx.Unauthorized(w, "Unauthorized")
		return
	}
	userID := *userIDPtr

	var req dto.PurchaseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.BadRequest(w, "Invalid request body")
		return
	}

	transaction, err := h.service.PurchaseTokens(userID, req)
	if err != nil {
		logger.Error("Failed to purchase tokens", zap.Error(err), zap.String("userId", userID.String()))
		
		if err == service.ErrInvalidAmount {
			httpx.BadRequest(w, err.Error())
			return
		}
		
		httpx.InternalError(w, "Failed to purchase tokens")
		return
	}

	// Add success message
	response := map[string]interface{}{
		"success":     true,
		"message":     "Tokens purchased successfully",
		"transaction": transaction,
	}

	httpx.Success(w, response)
}

// SpendTokens handles POST /api/wallet/spend
func (h *WalletHandler) SpendTokens(w http.ResponseWriter, r *http.Request) {
	userIDPtr := middleware.GetUserID(r)
	if userIDPtr == nil {
		httpx.Unauthorized(w, "Unauthorized")
		return
	}
	userID := *userIDPtr

	var req dto.SpendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.BadRequest(w, "Invalid request body")
		return
	}

	transaction, err := h.service.SpendTokens(userID, req)
	if err != nil {
		logger.Error("Failed to spend tokens", zap.Error(err), zap.String("userId", userID.String()))
		
		if err == service.ErrInsufficientBalance {
			httpx.BadRequest(w, "Insufficient balance")
			return
		}
		
		if err == service.ErrInvalidAmount {
			httpx.BadRequest(w, err.Error())
			return
		}
		
		httpx.InternalError(w, "Failed to spend tokens")
		return
	}

	response := map[string]interface{}{
		"success":     true,
		"message":     "Tokens spent successfully",
		"transaction": transaction,
	}

	httpx.Success(w, response)
}

// GetWalletInfo handles GET /api/user/{userId}/wallet
func (h *WalletHandler) GetWalletInfo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userIDStr := vars["userId"]

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		httpx.BadRequest(w, "Invalid user ID")
		return
	}

	balance, err := h.service.GetBalance(userID)
	if err != nil {
		logger.Error("Failed to get wallet info", zap.Error(err), zap.String("userId", userID.String()))
		httpx.InternalError(w, "Failed to get wallet info")
		return
	}

	httpx.Success(w, balance)
}

// GetTransactionHistory handles GET /api/wallet/transactions?limit=50
func (h *WalletHandler) GetTransactionHistory(w http.ResponseWriter, r *http.Request) {
	userIDPtr := middleware.GetUserID(r)
	if userIDPtr == nil {
		httpx.Unauthorized(w, "Unauthorized")
		return
	}
	userID := *userIDPtr

	// Parse limit from query params
	limitStr := r.URL.Query().Get("limit")
	limit := 50
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	transactions, err := h.service.GetTransactionHistory(userID, limit)
	if err != nil {
		logger.Error("Failed to get transaction history", zap.Error(err), zap.String("userId", userID.String()))
		httpx.InternalError(w, "Failed to get transactions")
		return
	}

	response := map[string]interface{}{
		"success":      true,
		"transactions": transactions,
		"count":        len(transactions),
	}

	httpx.Success(w, response)
}

// GrantWelcomeTokens handles POST /api/user/{userId}/wallet/grant-welcome
func (h *WalletHandler) GrantWelcomeTokens(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userIDStr := vars["userId"]

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		httpx.BadRequest(w, "Invalid user ID")
		return
	}

	// Grant 100 welcome tokens
	if err := h.service.GrantWelcomeTokens(userID, 100); err != nil {
		logger.Error("Failed to grant welcome tokens", zap.Error(err), zap.String("userId", userID.String()))
		httpx.InternalError(w, "Failed to grant tokens")
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Welcome tokens granted successfully",
		"amount":  100,
	}

	httpx.Success(w, response)
}
