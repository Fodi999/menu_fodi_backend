package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

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
	ingredients, err := h.service.GetAllIngredients()
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch ingredients")
		return
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

// GetAllRecipes возвращает весь каталог рецептов для админ-панели
func (h *AdminHandlers) GetAllRecipes(w http.ResponseWriter, r *http.Request) {
	recipes, err := h.service.GetAllRecipes()
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch recipes")
		return
	}

	// Формат совместимый с фронтендом (data + meta)
	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"data": recipes,
		"meta": map[string]interface{}{
			"total": len(recipes),
			"count": len(recipes),
		},
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
// Создаёт новый ингредиент с автоматическим переводом через Groq AI
func (h *AdminHandlers) CreateIngredient(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("🎯 CreateIngredient handler called\n")
	
	var req struct {
		InputName string `json:"inputName"`
		InputLang string `json:"inputLang"`
		Category  string `json:"category"`
		Unit      string `json:"unit"`
	}

	// Декодируем запрос
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		fmt.Printf("❌ Failed to decode body: %v\n", err)
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Логируем полученные данные
	fmt.Printf("📦 Received: inputName='%s', inputLang='%s', category='%s', unit='%s'\n", 
		req.InputName, req.InputLang, req.Category, req.Unit)

	// Валидация
	if req.InputName == "" {
		fmt.Printf("❌ inputName is empty\n")
		utils.RespondWithError(w, http.StatusBadRequest, "inputName is required")
		return
	}
	if req.InputLang == "" || (req.InputLang != "pl" && req.InputLang != "en" && req.InputLang != "ru") {
		utils.RespondWithError(w, http.StatusBadRequest, "inputLang must be pl, en, or ru")
		return
	}
	if req.Category == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "category is required")
		return
	}
	if req.Unit == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "unit is required")
		return
	}

	// Получаем userID из контекста (опционально, для логирования)
	userID := middleware.GetUserID(r)
	var userIDStr string
	if userID != nil {
		userIDStr = userID.String()
	}

	// 🤖 Вызываем сервис для создания ингредиента С ПЕРЕВОДОМ ЧЕРЕЗ AI
	ingredient, err := h.service.CreateIngredientSimple(req.InputName, req.InputLang, req.Category, req.Unit, userIDStr)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to create ingredient: %v", err))
		return
	}

	// Формируем ответ
	utils.RespondWithJSON(w, http.StatusCreated, map[string]interface{}{
		"success": true,
		"message": "Ingredient created and translated successfully",
		"data": map[string]interface{}{
			"id":             ingredient.ID,
			"namePl":         ingredient.NamePL,
			"nameEn":         ingredient.NameEN,
			"nameRu":         ingredient.NameRU,
			"category":       ingredient.Category,
			"unit":           ingredient.Unit,
			"autoTranslated": ingredient.AutoTranslated,
		},
	})
}
