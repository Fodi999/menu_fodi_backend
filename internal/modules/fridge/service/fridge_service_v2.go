package service

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
	notificationService "github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/notifications/service"
	"gorm.io/gorm"
)

// FridgeServiceV2 сервис для работы с холодильником (v2 с поддержкой expiry)
type FridgeServiceV2 interface {
	// Items
	GetItems(userID string) ([]models.FridgeItem, error)
	AddItem(userID string, req AddFridgeItemRequest) (*models.FridgeItem, error)
	UpdateItem(itemID string, userID string, req UpdateFridgeItemRequest) (*models.FridgeItem, error)
	DeleteItem(itemID string, userID string) error
	DiscardItem(itemID string, userID string) error

	// Auto-checks
	CheckAndNotifyExpiring(userID string) error
}

type fridgeServiceV2 struct {
	db                  *gorm.DB
	notificationService notificationService.NotificationService
}

func NewFridgeServiceV2(db *gorm.DB) FridgeServiceV2 {
	return &fridgeServiceV2{
		db:                  db,
		notificationService: notificationService.NewNotificationService(db),
	}
}

// PriceInput структура для цены с единицей измерения (альтернативный формат)
type PriceInput struct {
	Value float64 `json:"value"` // Цена
	Per   string  `json:"per"`   // Единица измерения (g, ml, pcs)
}

// AddFridgeItemRequest запрос на добавление продукта
type AddFridgeItemRequest struct {
	IngredientID string      `json:"ingredientId" binding:"required"`
	Quantity     float64     `json:"quantity" binding:"required,gt=0"`
	Unit         string      `json:"unit" binding:"required"`
	ExpiresAt    *time.Time  `json:"expiresAt"`
	PriceTotal   float64     `json:"priceTotal"`       // Формат 1: прямая цена (legacy)
	PriceInput   *PriceInput `json:"priceInput"`       // Формат 2: цена с единицей (новый)
}

// GetPriceTotal возвращает итоговую цену, поддерживая оба формата
func (r *AddFridgeItemRequest) GetPriceTotal() float64 {
	// Приоритет: PriceInput > PriceTotal
	if r.PriceInput != nil && r.PriceInput.Value > 0 {
		// Если PriceInput.Per совпадает с Unit, используем value напрямую
		// Иначе это цена за единицу, умножаем на количество
		if r.PriceInput.Per == r.Unit {
			return r.PriceInput.Value
		}
		// Цена указана за единицу (например, "78.44 PLN за 1g")
		// Пересчитываем на общее количество
		return r.PriceInput.Value * r.Quantity
	}
	// Fallback на старый формат
	return r.PriceTotal
}

// UpdateFridgeItemRequest запрос на обновление продукта
type UpdateFridgeItemRequest struct {
	Quantity   *float64   `json:"quantity" binding:"omitempty,gt=0"`
	ExpiresAt  *time.Time `json:"expiresAt"`
	PriceTotal *float64   `json:"priceTotal"`
}

// GetItems получить все продукты пользователя с автопроверкой
func (s *fridgeServiceV2) GetItems(userID string) ([]models.FridgeItem, error) {
	var items []models.FridgeItem

	err := s.db.Preload("Ingredient").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&items).Error

	if err != nil {
		return nil, fmt.Errorf("failed to fetch fridge items: %w", err)
	}

	// Обновляем статус и daysLeft для каждого продукта
	freshItems := make([]models.FridgeItem, 0, len(items))
	for i := range items {
		result := EvaluateFridgeItem(&items[i])
		items[i].Status = result.Status
		items[i].DaysLeft = result.DaysLeft

		// Сохраняем изменения в БД если статус изменился
		if items[i].Status == models.FridgeItemStatusExpired {
			s.db.Model(&items[i]).Updates(map[string]interface{}{
				"status":    result.Status,
				"days_left": result.DaysLeft,
			})
			// ❌ НЕ добавляем expired продукты в результат
			continue
		}

		// ✅ Добавляем только fresh/ok продукты
		freshItems = append(freshItems, items[i])
	}

	// Проверяем и создаем уведомления (1 раз в день)
	go s.CheckAndNotifyExpiring(userID)

	return freshItems, nil
}

// AddItem добавить продукт в холодильник
func (s *fridgeServiceV2) AddItem(userID string, req AddFridgeItemRequest) (*models.FridgeItem, error) {
	// Проверяем существование ингредиента
	var ingredient models.Ingredient
	if err := s.db.First(&ingredient, "id = ?", req.IngredientID).Error; err != nil {
		return nil, fmt.Errorf("ingredient not found: %w", err)
	}

	// Создаем новый item
	item := models.FridgeItem{
		UserID:       userID,
		IngredientID: req.IngredientID,
		Quantity:     req.Quantity,
		Unit:         req.Unit,
		ExpiresAt:    req.ExpiresAt,
		PriceTotal:   req.GetPriceTotal(), // Поддержка обоих форматов
		Status:       models.FridgeItemStatusFresh,
	}

	// Вычисляем начальный статус
	result := EvaluateFridgeItem(&item)
	item.Status = result.Status
	item.DaysLeft = result.DaysLeft

	if err := s.db.Create(&item).Error; err != nil {
		return nil, fmt.Errorf("failed to create fridge item: %w", err)
	}

	// Загружаем связанный ингредиент
	s.db.Preload("Ingredient").First(&item, "id = ?", item.ID)

	// Создаем уведомление если нужно
	if result.DaysLeft != nil {
		go CreateExpiryNotification(s.db, &item, *result.DaysLeft)
	}

	return &item, nil
}

// UpdateItem обновить продукт
func (s *fridgeServiceV2) UpdateItem(itemID string, userID string, req UpdateFridgeItemRequest) (*models.FridgeItem, error) {
	var item models.FridgeItem

	if err := s.db.Where("id = ? AND user_id = ?", itemID, userID).First(&item).Error; err != nil {
		return nil, fmt.Errorf("fridge item not found: %w", err)
	}

	// Обновляем поля
	updates := make(map[string]interface{})
	if req.Quantity != nil {
		updates["quantity"] = *req.Quantity
	}
	if req.ExpiresAt != nil {
		updates["expires_at"] = *req.ExpiresAt
	}
	if req.PriceTotal != nil {
		updates["price_total"] = *req.PriceTotal
	}

	if len(updates) > 0 {
		if err := s.db.Model(&item).Updates(updates).Error; err != nil {
			return nil, fmt.Errorf("failed to update fridge item: %w", err)
		}
	}

	// Пересчитываем статус
	s.db.First(&item, "id = ?", itemID)
	result := EvaluateFridgeItem(&item)
	s.db.Model(&item).Updates(map[string]interface{}{
		"status":    result.Status,
		"days_left": result.DaysLeft,
	})

	// Загружаем обновленные данные
	s.db.Preload("Ingredient").First(&item, "id = ?", itemID)

	return &item, nil
}

// DeleteItem удалить продукт из холодильника
func (s *fridgeServiceV2) DeleteItem(itemID string, userID string) error {
	// 1. ОБЯЗАТЕЛЬНО получить данные ПЕРЕД удалением
	var item models.FridgeItem
	if err := s.db.Where("id = ? AND user_id = ?", itemID, userID).
		Preload("Ingredient").
		First(&item).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("fridge item not found")
		}
		return fmt.Errorf("failed to get fridge item: %w", err)
	}

	// 2. Удаляем продукт
	result := s.db.Where("id = ? AND user_id = ?", itemID, userID).Delete(&models.FridgeItem{})

	if result.Error != nil {
		return fmt.Errorf("failed to delete fridge item: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("fridge item not found")
	}

	// 3. Создаём уведомление ПОСЛЕ успешного удаления
	s.createItemDeletedNotification(userID, &item)

	return nil
}

// DiscardItem выбросить продукт (мягкое удаление)
func (s *fridgeServiceV2) DiscardItem(itemID string, userID string) error {
	var item models.FridgeItem

	// 1. Получаем продукт с данными ингредиента
	if err := s.db.Where("id = ? AND user_id = ?", itemID, userID).
		Preload("Ingredient").
		First(&item).Error; err != nil {
		return fmt.Errorf("fridge item not found: %w", err)
	}

	// 2. Меняем статус на discarded
	if err := s.db.Model(&item).Update("status", models.FridgeItemStatusDiscarded).Error; err != nil {
		return fmt.Errorf("failed to discard item: %w", err)
	}

	fmt.Printf("🗑️  Item %s discarded (loss: %.2f PLN)\n", item.ID, item.PriceTotal)

	// 3. Создаём уведомление ПОСЛЕ успешного выброса
	s.createItemDiscardedNotification(userID, &item)

	return nil
}

// CheckAndNotifyExpiring проверяет продукты и создает уведомления
func (s *fridgeServiceV2) CheckAndNotifyExpiring(userID string) error {
	return CheckAndNotifyExpiringItems(s.db, userID)
}

// SetDB устанавливает подключение к БД (для инжекции)
func (s *fridgeServiceV2) SetDB(db *gorm.DB) {
	s.db = db
}

// createItemDeletedNotification создаёт уведомление об удалении продукта
func (s *fridgeServiceV2) createItemDeletedNotification(userID string, item *models.FridgeItem) {
	if item.Ingredient == nil {
		fmt.Printf("⚠️  Cannot create delete notification: ingredient data missing (item_id=%s, user_id=%s)\n",
			item.ID, userID)
		return
	}

	// Получаем польское название если доступно
	ingredientName := item.Ingredient.Name
	if item.Ingredient.NamePL != nil && *item.Ingredient.NamePL != "" {
		ingredientName = *item.Ingredient.NamePL
	}

	// Формат: "Czosnek удалён из холодильника (3.5 g)"
	message := fmt.Sprintf("%s удалён из холодильника (%.1f %s)",
		ingredientName,
		item.Quantity,
		item.Unit,
	)

	// Meta данные для уведомления
	meta := map[string]interface{}{
		"fridgeItemId": item.ID,
		"ingredientId": item.IngredientID,
		"quantity":     item.Quantity,
		"unit":         item.Unit,
		"action":       "deleted",
	}

	metaBytes, err := json.Marshal(meta)
	if err != nil {
		fmt.Printf("⚠️  Failed to marshal notification meta (item_id=%s, error=%v)\n", item.ID, err)
		return
	}
	metaStr := string(metaBytes)

	// Создаём уведомление
	notification := &models.Notification{
		UserID:  userID,
		Type:    models.NotificationTypeFridge,
		Level:   models.NotificationLevelInfo,
		Title:   "Продукт удалён из холодильника",
		Message: message,
		Meta:    &metaStr,
	}

	// Не блокируем удаление при ошибке создания уведомления
	if err := s.notificationService.Create(notification); err != nil {
		fmt.Printf("⚠️  Failed to create delete notification (user_id=%s, item_id=%s, error=%v)\n",
			userID, item.ID, err)
	}
}

// createItemDiscardedNotification создаёт уведомление о выбросе продукта
func (s *fridgeServiceV2) createItemDiscardedNotification(userID string, item *models.FridgeItem) {
	if item.Ingredient == nil {
		fmt.Printf("⚠️  Cannot create discard notification: ingredient data missing (item_id=%s, user_id=%s)\n",
			item.ID, userID)
		return
	}

	// Получаем польское название если доступно
	ingredientName := item.Ingredient.Name
	if item.Ingredient.NamePL != nil && *item.Ingredient.NamePL != "" {
		ingredientName = *item.Ingredient.NamePL
	}

	// Определяем level и title в зависимости от стоимости
	level := models.NotificationLevelWarning
	title := "Продукт выброшен"

	if item.PriceTotal > 0 {
		level = models.NotificationLevelCritical
		title = "Потеря продукта"
	}

	// Формат: "Czosnek выброшен. Потеря: 5.50 PLN"
	message := fmt.Sprintf("%s выброшен. Потеря: %.2f PLN",
		ingredientName,
		item.PriceTotal,
	)

	// Meta данные для уведомления
	meta := map[string]interface{}{
		"fridgeItemId": item.ID,
		"ingredientId": item.IngredientID,
		"quantity":     item.Quantity,
		"unit":         item.Unit,
		"action":       "discarded",
		"loss":         item.PriceTotal,
	}

	metaBytes, err := json.Marshal(meta)
	if err != nil {
		fmt.Printf("⚠️  Failed to marshal notification meta (item_id=%s, error=%v)\n", item.ID, err)
		return
	}
	metaStr := string(metaBytes)

	// Создаём уведомление
	notification := &models.Notification{
		UserID:  userID,
		Type:    models.NotificationTypeFridge,
		Level:   level,
		Title:   title,
		Message: message,
		Meta:    &metaStr,
	}

	// Не блокируем выброс при ошибке создания уведомления
	if err := s.notificationService.Create(notification); err != nil {
		fmt.Printf("⚠️  Failed to create discard notification (user_id=%s, item_id=%s, error=%v)\n",
			userID, item.ID, err)
	}
}
