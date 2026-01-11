package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/middleware"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/admin/service"
	authservice "github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/auth/service"
	"github.com/dmitrijfomin/menu-fodifood/backend/pkg/utils"
	"github.com/go-chi/chi/v5"
)

// IngredientResponse - DTO для ответа API (camelCase для frontend)
type IngredientResponse struct {
	ID              string `json:"id"`
	NamePl          string `json:"namePl"`
	NameEn          string `json:"nameEn"`
	NameRu          string `json:"nameRu"`
	Category        string `json:"category"`        // Culinary category (UI)
	NutritionGroup  string `json:"nutritionGroup"`  // Nutritional grouping (analytics)
	Unit            string `json:"unit"`
	NormalizedValue string `json:"normalizedValue"`
	AutoTranslated  bool   `json:"autoTranslated"`
}

// ToIngredientResponse - mapper из модели в DTO
func ToIngredientResponse(i *models.Ingredient) IngredientResponse {
	resp := IngredientResponse{
		ID:             i.ID,
		Category:       i.Category,
		NutritionGroup: i.NutritionGroup,
		Unit:           i.Unit,
		AutoTranslated: i.AutoTranslated,
	}

	// Безопасное разыменование указателей
	if i.NamePL != nil {
		resp.NamePl = *i.NamePL
	}
	if i.NameEN != nil {
		resp.NameEn = *i.NameEN
	}
	if i.NameRU != nil {
		resp.NameRu = *i.NameRU
	}
	if i.NormalizedValue != nil {
		resp.NormalizedValue = *i.NormalizedValue
	}

	return resp
}

// RecipeResponse - DTO для рецепта (camelCase для frontend)
type RecipeResponse struct {
	ID                 string      `json:"id"`
	CanonicalName      string      `json:"canonicalName"`
	Title              string      `json:"title"`
	NamePl             string      `json:"namePl"`
	NameEn             string      `json:"nameEn"`
	NameRu             string      `json:"nameRu"`
	DescriptionPl      string      `json:"descriptionPl"`
	DescriptionEn      string      `json:"descriptionEn"`
	DescriptionRu      string      `json:"descriptionRu"`
	Country            string      `json:"country"`
	Region             string      `json:"region"`
	Category           string      `json:"category"`
	Difficulty         string      `json:"difficulty"`
	TimeMinutes        int         `json:"timeMinutes"`
	Servings           int         `json:"servings"`
	PortionWeightGrams int         `json:"portionWeightGrams"`
	StepsPl            interface{} `json:"stepsPl"`
	StepsEn            interface{} `json:"stepsEn"`
	StepsRu            interface{} `json:"stepsRu"`
	NutritionProfile   interface{} `json:"nutritionProfile"`
	Source             interface{} `json:"source"`
	CreatedAt          string      `json:"createdAt"`
	UpdatedAt          string      `json:"updatedAt"`
}

// ToRecipeResponse - mapper из модели в DTO
func ToRecipeResponse(r *models.RecipeCatalog) RecipeResponse {
	resp := RecipeResponse{
		ID:            r.ID.String(),
		CanonicalName: r.CanonicalName,
		Title:         r.Title,
		Country:       r.Country,
		Category:      r.Category,
		Difficulty:    r.Difficulty,
		TimeMinutes:   r.TimeMinutes,
		Servings:      r.Servings,
		CreatedAt:     r.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:     r.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	// Безопасное разыменование указателей
	if r.NamePl != nil {
		resp.NamePl = *r.NamePl
	}
	if r.NameEn != nil {
		resp.NameEn = *r.NameEn
	}
	if r.NameRu != nil {
		resp.NameRu = *r.NameRu
	}
	if r.DescriptionPl != nil {
		resp.DescriptionPl = *r.DescriptionPl
	}
	if r.DescriptionEn != nil {
		resp.DescriptionEn = *r.DescriptionEn
	}
	if r.DescriptionRu != nil {
		resp.DescriptionRu = *r.DescriptionRu
	}
	if r.Region != nil {
		resp.Region = *r.Region
	}
	if r.PortionWeightGrams != nil {
		resp.PortionWeightGrams = *r.PortionWeightGrams
	}

	// JSONB поля
	if len(r.StepsPl) > 0 {
		json.Unmarshal(r.StepsPl, &resp.StepsPl)
	}
	if len(r.StepsEn) > 0 {
		json.Unmarshal(r.StepsEn, &resp.StepsEn)
	}
	if len(r.StepsRu) > 0 {
		json.Unmarshal(r.StepsRu, &resp.StepsRu)
	}
	if len(r.NutritionProfile) > 0 {
		json.Unmarshal(r.NutritionProfile, &resp.NutritionProfile)
	}
	if len(r.Source) > 0 {
		json.Unmarshal(r.Source, &resp.Source)
	}

	return resp
}

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
	// Парсим query параметры
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 1000 {
		limit = 20
	}

	// Получаем фильтры
	role := r.URL.Query().Get("role")
	status := r.URL.Query().Get("status")
	search := r.URL.Query().Get("search")

	// Создаём параметры
	params := service.GetUsersParams{
		Page:  page,
		Limit: limit,
	}

	if role != "" {
		params.Role = &role
	}
	if status != "" {
		params.Status = &status
	}
	if search != "" {
		params.Search = &search
	}

	// Получаем пользователей с фильтрами
	response, err := h.service.GetUsersWithFilters(params)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch users")
		return
	}

	// Возвращаем ответ
	utils.RespondWithJSON(w, http.StatusOK, response)
}

func (h *AdminHandlers) GetUsersStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.service.GetUsersStats()
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch user stats")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, stats)
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

// Treasury Handlers

// GetTreasuryInfo возвращает информацию о казначействе (treasury)
func (h *AdminHandlers) GetTreasuryInfo(w http.ResponseWriter, r *http.Request) {
	treasury, err := h.service.GetTreasuryInfo()
	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "Treasury not found")
		return
	}

	// Добавляем вычисляемые поля
	response := map[string]interface{}{
		"id":              treasury.ID,
		"user_id":         treasury.UserID,
		"balance":         treasury.Balance,
		"total_allocated": treasury.TotalAllocated,
		"total_used":      treasury.TotalUsed,
		"total_supply":    treasury.TotalAllocated, // Начальный supply
		"distributed":     treasury.TotalUsed,
		"remaining":       treasury.Balance,
		"created_at":      treasury.CreatedAt,
		"updated_at":      treasury.UpdatedAt,
	}

	utils.RespondWithJSON(w, http.StatusOK, response)
}

// AllocateFromTreasury выделяет токены из казначейства пользователю
func (h *AdminHandlers) AllocateFromTreasury(w http.ResponseWriter, r *http.Request) {
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

	// Не позволяем выделять токены казначейству из казначейства
	if req.UserID == "TREASURY" {
		utils.RespondWithError(w, http.StatusBadRequest, "Cannot allocate tokens to treasury from itself")
		return
	}

	err := h.service.AllocateFromTreasury(req.UserID, req.Amount)
	if err != nil {
		switch err.Error() {
		case "insufficient treasury balance":
			utils.RespondWithError(w, http.StatusBadRequest, "Insufficient treasury balance")
		case "user token bank not found":
			utils.RespondWithError(w, http.StatusNotFound, "User token bank not found")
		case "treasury not found":
			utils.RespondWithError(w, http.StatusInternalServerError, "Treasury not initialized")
		default:
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to allocate from treasury: "+err.Error())
		}
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Tokens allocated from treasury successfully",
		"user_id": req.UserID,
		"amount":  req.Amount,
		"source":  "TREASURY",
	})
}

// GetTreasuryBalance возвращает только баланс казначейства (упрощённый endpoint)
func (h *AdminHandlers) GetTreasuryBalance(w http.ResponseWriter, r *http.Request) {
	balance, err := h.service.GetTreasuryBalance()
	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "Treasury not found")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]int64{
		"balance": balance,
	})
}

// GetTreasuryStats возвращает полную статистику Treasury
func (h *AdminHandlers) GetTreasuryStats(w http.ResponseWriter, r *http.Request) {
	treasury, err := h.service.GetTreasuryInfo()
	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "Treasury not found")
		return
	}

	stats := map[string]int64{
		"totalIssued": treasury.TotalAllocated, // Всего выпущено
		"circulating": treasury.TotalUsed,      // В обращении (использовано)
		"locked":      0,                       // Заблокировано (пока 0)
		"available":   treasury.Balance,        // Доступно
		"balance":     treasury.Balance,        // Текущий баланс
	}

	utils.RespondWithJSON(w, http.StatusOK, stats)
}

// StreamTreasury предоставляет SSE stream для real-time обновлений баланса Treasury
func (h *AdminHandlers) StreamTreasury(w http.ResponseWriter, r *http.Request) {
	// Устанавливаем SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Отправляем текущий баланс сразу
	balance, err := h.service.GetTreasuryBalance()
	if err != nil {
		http.Error(w, "Failed to get treasury balance", http.StatusInternalServerError)
		return
	}

	initialData := map[string]interface{}{
		"balance": balance,
		"type":    "initial",
	}
	dataJSON, _ := json.Marshal(initialData)
	w.Write([]byte("data: "))
	w.Write(dataJSON)
	w.Write([]byte("\n\n"))
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}

	// TODO: Интеграция с WebSocket EventBus для real-time обновлений
	// Пока держим соединение открытым
	<-r.Context().Done()
}

// ============================================
// Token Transactions History Handlers
// ============================================

// GetAllTransactions возвращает все транзакции токенов
func (h *AdminHandlers) GetAllTransactions(w http.ResponseWriter, r *http.Request) {
	// Параметры пагинации
	limit := 50
	offset := 0

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		fmt.Sscanf(limitStr, "%d", &limit)
	}
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		fmt.Sscanf(offsetStr, "%d", &offset)
	}

	transactions, err := h.service.GetAllTransactions(limit, offset)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch transactions")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, transactions)
}

// GetUserTransactions возвращает транзакции конкретного пользователя
func (h *AdminHandlers) GetUserTransactions(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")

	limit := 50
	offset := 0

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		fmt.Sscanf(limitStr, "%d", &limit)
	}
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		fmt.Sscanf(offsetStr, "%d", &offset)
	}

	transactions, err := h.service.GetUserTransactions(userID, limit, offset)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch user transactions")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, transactions)
}

// GetTransactionsByType возвращает транзакции по типу с фильтрами
func (h *AdminHandlers) GetTransactionsByType(w http.ResponseWriter, r *http.Request) {
	txType := r.URL.Query().Get("type")

	limit := 50
	offset := 0

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		fmt.Sscanf(limitStr, "%d", &limit)
	}
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		fmt.Sscanf(offsetStr, "%d", &offset)
	}

	var transactions []interface{}
	var err error

	if txType != "" {
		txs, e := h.service.GetTransactionsByType(txType, limit, offset)
		err = e
		for _, tx := range txs {
			transactions = append(transactions, tx)
		}
	} else {
		txs, e := h.service.GetAllTransactions(limit, offset)
		err = e
		for _, tx := range txs {
			transactions = append(transactions, tx)
		}
	}

	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch transactions")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, transactions)
}

// GetTransactionStats возвращает статистику транзакций
func (h *AdminHandlers) GetTransactionStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.service.GetTransactionStats()
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch transaction stats")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, stats)
}

// ImportIngredients handles bulk import of ingredient catalog items
func (h *AdminHandlers) ImportIngredients(w http.ResponseWriter, r *http.Request) {
	var payload []struct {
		Name                 string   `json:"name"`
		Unit                 string   `json:"unit"`
		Category             string   `json:"category"`
		DefaultShelfLifeDays *int     `json:"defaultShelfLifeDays,omitempty"`
		DefaultPricePerUnit  *float64 `json:"defaultPricePerUnit,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	if len(payload) == 0 {
		utils.RespondWithError(w, http.StatusBadRequest, "Empty ingredients array")
		return
	}

	// Convert payload to service format
	req := make([]struct {
		Name                 string
		Unit                 string
		Category             string
		DefaultShelfLifeDays *int
		DefaultPricePerUnit  *float64
	}, len(payload))

	for i, item := range payload {
		req[i] = struct {
			Name                 string
			Unit                 string
			Category             string
			DefaultShelfLifeDays *int
			DefaultPricePerUnit  *float64
		}{
			Name:                 item.Name,
			Unit:                 item.Unit,
			Category:             item.Category,
			DefaultShelfLifeDays: item.DefaultShelfLifeDays,
			DefaultPricePerUnit:  item.DefaultPricePerUnit,
		}
	}

	imported, err := h.service.BulkImportIngredients(req)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Import failed: %v", err))
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success":  true,
		"imported": imported,
		"total":    len(req),
	})
}

// GetAllIngredients возвращает полный каталог ингредиентов для админ-панели
func (h *AdminHandlers) GetAllIngredients(w http.ResponseWriter, r *http.Request) {
	// Получаем параметр поиска из query
	searchQuery := r.URL.Query().Get("search")
	
	ingredients, err := h.service.GetAllIngredients()
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch ingredients")
		return
	}

	// 🔍 Фильтруем по поисковому запросу если он указан
	if searchQuery != "" {
		searchQuery = strings.ToLower(searchQuery)
		var filtered []models.Ingredient
		for _, ing := range ingredients {
			// Ищем в name, namePl, nameEn, nameRu
			if strings.Contains(strings.ToLower(ing.Name), searchQuery) ||
				(ing.NamePL != nil && strings.Contains(strings.ToLower(*ing.NamePL), searchQuery)) ||
				(ing.NameEN != nil && strings.Contains(strings.ToLower(*ing.NameEN), searchQuery)) ||
				(ing.NameRU != nil && strings.Contains(strings.ToLower(*ing.NameRU), searchQuery)) {
				filtered = append(filtered, ing)
			}
		}
		ingredients = filtered
	}

	// Формат совместимый с фронтендом (data + meta)
	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"data": ingredients,
		"meta": map[string]interface{}{
			"total": len(ingredients),
			"count": len(ingredients),
		},
	})
}

// GetIngredientsStats возвращает статистику по ингредиентам
func (h *AdminHandlers) GetIngredientsStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.service.GetIngredientsStats()
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch ingredients stats")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, stats)
}

// GetAllRecipes возвращает каталог рецептов с фильтрацией и пагинацией
func (h *AdminHandlers) GetAllRecipes(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	
	// Парсим фильтры из query параметров
	filter := service.ParseRecipeFilter(r)

	// Логируем фильтры для отладки и аналитики
	fmt.Printf("📊 Recipes filter: page=%d, limit=%d, sort=%s, category=%v, difficulty=%v, timeLte=%v, ingredientIds=%v\n",
		filter.Page, filter.Limit, filter.Sort, filter.Category, filter.Difficulty, filter.TimeLte, filter.IngredientIDs)

	// Получаем отфильтрованные рецепты
	recipes, total, err := h.service.GetFilteredRecipes(filter)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch recipes")
		return
	}

	// Преобразуем в DTO для фронтенда
	recipeResponses := make([]RecipeResponse, len(recipes))
	for i, recipe := range recipes {
		recipeResponses[i] = ToRecipeResponse(&recipe)
	}

	// Observability: логируем время выполнения
	elapsed := time.Since(startTime)
	if elapsed > 300*time.Millisecond {
		fmt.Printf("⚠️  SLOW QUERY: Recipe catalog took %v (filters: category=%v, difficulty=%v, ingredients=%d)\n",
			elapsed, filter.Category, filter.Difficulty, len(filter.IngredientIDs))
	} else {
		fmt.Printf("✅ Recipe catalog query: %v\n", elapsed)
	}

	// Формат совместимый с фронтендом (data + meta + pagination)
	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"data": recipeResponses,
		"meta": map[string]interface{}{
			"page":  filter.Page,
			"limit": filter.Limit,
			"total": total,
			"count": len(recipeResponses),
		},
	})
}

// GetRecipeFilterMetadata возвращает метаданные для UI фильтров
func (h *AdminHandlers) GetRecipeFilterMetadata(w http.ResponseWriter, r *http.Request) {
	meta, err := h.service.GetRecipeFilterMetadata()
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch filter metadata")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    meta,
	})
}

// GetRecipesStats возвращает статистику по рецептам
func (h *AdminHandlers) GetRecipesStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.service.GetRecipesStats()
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch recipes stats")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, stats)
}

// CreateIngredient создает новый ингредиент в каталоге
// POST /api/admin/ingredients
// 🧠 ПОЛНАЯ AI-КЛАССИФИКАЦИЯ - принимает только inputName
func (h *AdminHandlers) CreateIngredient(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("🎯 CreateIngredient handler called (AI Classification Mode)\n")
	
	var req struct {
		InputName string `json:"inputName"`
	}

	// Декодируем запрос
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		fmt.Printf("❌ Failed to decode body: %v\n", err)
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Валидация
	if strings.TrimSpace(req.InputName) == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "inputName is required")
		return
	}

	// Получаем userID из контекста
	userID := middleware.GetUserID(r)
	var userIDStr string
	if userID != nil {
		userIDStr = userID.String()
	}

	fmt.Printf("📦 Creating ingredient from: '%s'\n", req.InputName)

	// � ПОЛНАЯ AI-КЛАССИФИКАЦИЯ - AI определяет всё: язык, переводы, категорию, единицы
	ingredient, err := h.service.CreateIngredientWithAI(req.InputName, userIDStr)
	if err != nil {
		// Проверка на дубликат
		if strings.Contains(err.Error(), "INGREDIENT_ALREADY_EXISTS") {
			utils.RespondWithError(w, http.StatusConflict, fmt.Sprintf("Ingredient already exists: %v", err))
			return
		}
		utils.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to create ingredient: %v", err))
		return
	}

	// ✅ Формируем ответ через mapper (безопасное разыменование указателей)
	utils.RespondWithJSON(w, http.StatusCreated, map[string]interface{}{
		"success": true,
		"message": "Ingredient created via AI classification",
		"data":    ToIngredientResponse(ingredient),
	})
}

// SuggestIngredients - быстрый поиск для autocomplete (без AI, с локализацией)
// GET /api/admin/ingredients/suggest?q=абр&limit=5
// Поддерживает заголовок Accept-Language: pl/en/ru
func (h *AdminHandlers) SuggestIngredients(w http.ResponseWriter, r *http.Request) {
	// 🛡️ Защита от panic
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("🚨 PANIC in SuggestIngredients handler: %v\n", r)
			utils.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		}
	}()

	query := r.URL.Query().Get("q")
	limitStr := r.URL.Query().Get("limit")

	// Получаем язык из Accept-Language заголовка
	acceptLang := r.Header.Get("Accept-Language")
	lang := normalizeLang(acceptLang)

	fmt.Printf("📥 Request: GET /suggest?q=%s&limit=%s (Accept-Language: %s → %s)\n", 
		query, limitStr, acceptLang, lang)

	// Default limit
	limit := 5
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 20 {
			limit = parsedLimit
		}
	}

	// Дополнительная валидация query
	query = strings.TrimSpace(query)
	if len(query) > 100 {
		fmt.Printf("⚠️ Query too long (%d chars), truncating to 100\n", len(query))
		query = query[:100]
	}

	fmt.Printf("🔍 SuggestIngredients: query='%s', limit=%d, lang='%s'\n", query, limit, lang)

	// Получаем подсказки из service с указанием языка
	suggestions, err := h.service.SuggestIngredients(query, limit, lang)
	if err != nil {
		fmt.Printf("❌ Suggest failed: %v\n", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch suggestions")
		return
	}

	fmt.Printf("✅ Returning %d suggestions (lang=%s)\n", len(suggestions), lang)
	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"data": suggestions,
	})
}

// normalizeLang нормализует Accept-Language заголовок в один из трех языков
// Примеры: "pl-PL" → "pl", "en-US" → "en", "ru-RU" → "ru"
func normalizeLang(acceptLang string) string {
	acceptLang = strings.ToLower(strings.TrimSpace(acceptLang))
	
	// Парсим первый язык из списка (например: "pl-PL,en;q=0.9" → "pl")
	if idx := strings.Index(acceptLang, ","); idx > 0 {
		acceptLang = acceptLang[:idx]
	}
	
	// Проверяем префикс языка
	switch {
	case strings.HasPrefix(acceptLang, "pl"):
		return "pl"
	case strings.HasPrefix(acceptLang, "ru"):
		return "ru"
	case strings.HasPrefix(acceptLang, "en"):
		return "en"
	default:
		// По умолчанию английский
		return "en"
	}
}

// IngredientHint - AI подсказка при конфликте
// POST /api/admin/ingredients/hint
func (h *AdminHandlers) IngredientHint(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Input    string   `json:"input"`
		Existing []string `json:"existing"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Валидация
	if strings.TrimSpace(req.Input) == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "input is required")
		return
	}

	fmt.Printf("💡 IngredientHint: input='%s', existing=%v\n", req.Input, req.Existing)

	// Получаем AI подсказку
	hint, err := h.service.GenerateIngredientHint(req.Input, req.Existing)
	if err != nil {
		fmt.Printf("❌ AI hint failed: %v\n", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "AI hint generation failed")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"suggestion": hint,
	})
}
