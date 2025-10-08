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

// CreateOrderRequest структура запроса для создания заказа
type CreateOrderRequest struct {
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	Address string `json:"address"`
	Comment string `json:"comment"`
	Items   []struct {
		ProductID string  `json:"productId"`
		Quantity  int     `json:"quantity"`
		Price     float64 `json:"price"`
	} `json:"items"`
}

// CreateOrder создание нового заказа
func CreateOrder(w http.ResponseWriter, r *http.Request) {
	var req CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Валидация с очисткой пробелов
	if strings.TrimSpace(req.Name) == "" ||
		strings.TrimSpace(req.Phone) == "" ||
		strings.TrimSpace(req.Address) == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Name, phone and address are required")
		return
	}

	if len(req.Items) == 0 {
		utils.RespondWithError(w, http.StatusBadRequest, "Order must contain at least one item")
		return
	}

	// Получаем ID пользователя из контекста (если авторизован)
	var userID *string
	if uid, ok := r.Context().Value("userID").(string); ok && uid != "" {
		userID = &uid
	}
	// Если userID == nil, это гостевой заказ

	// Рассчитываем общую сумму с округлением
	var total float64
	for _, item := range req.Items {
		total += item.Price * float64(item.Quantity)
	}
	// Округляем до 2 знаков после запятой
	total = math.Round(total*100) / 100

	// Создаём заказ
	orderID := uuid.New().String()
	order := models.Order{
		ID:        orderID,
		UserID:    userID,
		Name:      strings.TrimSpace(req.Name),
		Status:    "pending",
		Total:     total,
		Address:   strings.TrimSpace(req.Address),
		Phone:     strings.TrimSpace(req.Phone),
		Comment:   strings.TrimSpace(req.Comment),
		CreatedAt: time.Now(),
	}

	// Используем транзакцию для атомарности операции
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Сохраняем заказ
	if err := tx.Create(&order).Error; err != nil {
		tx.Rollback()
		log.Printf("[ORDER] ❌ Error creating order: %v", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to create order")
		return
	}

	// Сохраняем позиции заказа
	for _, item := range req.Items {
		orderItem := models.OrderItem{
			ID:        uuid.New().String(),
			OrderID:   orderID,
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     item.Price,
		}

		if err := tx.Create(&orderItem).Error; err != nil {
			tx.Rollback()
			log.Printf("[ORDER] ❌ Error creating order item: %v", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to create order item")
			return
		}
	}

	// Коммитим транзакцию
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		log.Printf("[ORDER] ❌ Error committing transaction: %v", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to commit order")
		return
	}

	log.Printf("[ORDER] ✅ Created ID=%s, Total=%.2f, Status=%s, Items=%d",
		orderID, total, order.Status, len(req.Items))

	// Отправляем WebSocket уведомление о новом заказе с расширенной информацией
	BroadcastOrderNotification("new_order", map[string]interface{}{
		"orderId":    orderID,
		"total":      total,
		"status":     order.Status,
		"name":       req.Name,
		"phone":      req.Phone,
		"address":    req.Address,
		"itemsCount": len(req.Items),
		"createdAt":  order.CreatedAt,
	})

	// Формируем ответ с redirectTo для удобного перехода на страницу заказа
	response := map[string]interface{}{
		"message": "Order created successfully",
		"orderId": orderID,
		"total":   total,
		"status":  order.Status,
	}

	// Если пользователь авторизован, добавляем redirectTo
	if userID != nil && *userID != "" {
		response["redirectTo"] = "/orders/" + orderID
	}

	utils.RespondWithJSON(w, http.StatusCreated, response)
}

// GetAllOrders получение всех заказов (для админа)
func GetAllOrders(w http.ResponseWriter, r *http.Request) {
	type OrderResponse struct {
		ID        string    `json:"id"`
		UserID    string    `json:"userId"`
		Status    string    `json:"status"`
		Total     float64   `json:"total"`
		Address   string    `json:"address"`
		Phone     string    `json:"phone"`
		Comment   string    `json:"comment"`
		CreatedAt time.Time `json:"createdAt"`
		User      struct {
			ID    string `json:"id"`
			Name  string `json:"name"`
			Email string `json:"email"`
		} `json:"user"`
		Items []struct {
			ID       string  `json:"id"`
			Quantity int     `json:"quantity"`
			Price    float64 `json:"price"`
			Product  struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"product"`
		} `json:"items"`
	}

	var orders []OrderResponse

	// Получаем заказы с JOIN к пользователям и позициям
	rows, err := database.DB.Raw(`
		SELECT 
			o.id as order_id,
			o.user_id as user_id,
			o.status,
			o.total,
			o.address,
			o.phone,
			o.comment,
			o.created_at as created_at,
			u.id as u_id,
			u.name as u_name,
			u.email as u_email,
			oi.id as item_id,
			oi.quantity,
			oi.price,
			p.id as p_id,
			p.name as p_name
		FROM "Order" o
		LEFT JOIN "User" u ON o.user_id = u.id
		LEFT JOIN "OrderItem" oi ON o.id = oi.order_id
		LEFT JOIN "Product" p ON oi.product_id = p.id
		ORDER BY o.created_at DESC
	`).Rows()

	if err != nil {
		log.Printf("[ORDER] ❌ Error fetching orders: %v", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch orders")
		return
	}
	defer rows.Close()

	ordersMap := make(map[string]*OrderResponse)

	for rows.Next() {
		var orderID, status, address, phone, comment string
		var userID *string // Nullable для гостевых заказов
		var total float64
		var createdAt time.Time
		var uID, uName, uEmail *string
		var itemID, pID, pName *string
		var quantity *int
		var price *float64

		if err := rows.Scan(
			&orderID, &userID, &status, &total, &address, &phone, &comment, &createdAt,
			&uID, &uName, &uEmail,
			&itemID, &quantity, &price,
			&pID, &pName,
		); err != nil {
			log.Printf("[ORDER] ❌ Error scanning order row: %v", err)
			continue
		}

		// Если заказ ещё не добавлен в map
		if _, exists := ordersMap[orderID]; !exists {
			userIDStr := ""
			if userID != nil {
				userIDStr = *userID
			}

			order := &OrderResponse{
				ID:        orderID,
				UserID:    userIDStr,
				Status:    status,
				Total:     total,
				Address:   address,
				Phone:     phone,
				Comment:   comment,
				CreatedAt: createdAt,
				Items: []struct {
					ID       string  `json:"id"`
					Quantity int     `json:"quantity"`
					Price    float64 `json:"price"`
					Product  struct {
						ID   string `json:"id"`
						Name string `json:"name"`
					} `json:"product"`
				}{},
			}

			if uID != nil {
				order.User.ID = *uID
			}
			if uName != nil {
				order.User.Name = *uName
			}
			if uEmail != nil {
				order.User.Email = *uEmail
			}

			ordersMap[orderID] = order
		}

		// Добавляем позицию заказа
		if itemID != nil && pID != nil {
			item := struct {
				ID       string  `json:"id"`
				Quantity int     `json:"quantity"`
				Price    float64 `json:"price"`
				Product  struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"product"`
			}{
				ID:       *itemID,
				Quantity: *quantity,
				Price:    *price,
			}
			item.Product.ID = *pID
			item.Product.Name = *pName

			ordersMap[orderID].Items = append(ordersMap[orderID].Items, item)
		}
	}

	// Преобразуем map в массив
	for _, order := range ordersMap {
		orders = append(orders, *order)
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"orders": orders,
	})
}

// GetUserOrders получение заказов текущего пользователя
func GetUserOrders(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	type OrderResponse struct {
		ID        string    `json:"id"`
		Status    string    `json:"status"`
		Total     float64   `json:"total"`
		Address   string    `json:"address"`
		Phone     string    `json:"phone"`
		Comment   string    `json:"comment"`
		CreatedAt time.Time `json:"createdAt"`
	}

	var orders []OrderResponse

	if err := database.DB.Table("Order").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&orders).Error; err != nil {
		log.Printf("[ORDER] ❌ Error fetching user orders: %v", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch orders")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"orders": orders,
	})
}

// UpdateOrderStatus обновление статуса заказа (только для админа)
func UpdateOrderStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderID := vars["id"]

	var req struct {
		Status string `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Валидация статуса
	validStatuses := map[string]bool{
		"pending":   true,
		"confirmed": true,
		"preparing": true,
		"delivered": true,
		"cancelled": true,
	}

	if !validStatuses[req.Status] {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid status")
		return
	}

	// Обновляем статус и updatedAt
	now := time.Now()
	if err := database.DB.Table("Order").
		Where("id = ?", orderID).
		Updates(map[string]interface{}{
			"status":     req.Status,
			"updated_at": now,
		}).Error; err != nil {
		log.Printf("[ORDER] ❌ Error updating order status: %v", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to update order status")
		return
	}

	log.Printf("[ORDER] 🟢 Updated status: ID=%s, Status=%s", orderID, req.Status)

	// Отправляем WebSocket уведомление об обновлении статуса
	BroadcastOrderNotification("order_updated", map[string]interface{}{
		"orderId":   orderID,
		"status":    req.Status,
		"updatedAt": now,
	})

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Order status updated successfully",
		"status":  req.Status,
	})
}
