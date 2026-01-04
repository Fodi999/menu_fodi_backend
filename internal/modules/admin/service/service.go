package service

import (
	"errors"
	"math"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/database"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
	"gorm.io/gorm"
)

// GetUsersParams параметры для фильтрации пользователей
type GetUsersParams struct {
	Page   int
	Limit  int
	Role   *string
	Status *string
	Search *string
}

// UserListResponse ответ со списком пользователей и метаданными
type UserListResponse struct {
	Users []models.User  `json:"users"`
	Meta  PaginationMeta `json:"meta"`
}

// PaginationMeta метаданные пагинации
type PaginationMeta struct {
	Total      int `json:"total"`
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	TotalPages int `json:"totalPages"`
}

// AdminService интерфейс для бизнес-логики администратора
type AdminService interface {
	// Users
	GetAllUsers() ([]models.User, error)
	GetUsersWithFilters(params GetUsersParams) (*UserListResponse, error)
	GetUsersStats() (map[string]interface{}, error)
	UpdateUser(userID string, name, email string) (*models.User, error)
	DeleteUser(userID string) error
	UpdateUserRole(userID, role string) error

	// Orders
	GetAllOrders() ([]models.Order, error)
	GetRecentOrders(limit int) ([]models.Order, error)
	UpdateOrderStatus(orderID, status string) error

	// Statistics
	GetAdminStats() (map[string]interface{}, error)

	// Admin Profile
	GetAdminProfile(adminID string) (map[string]interface{}, error)

	// Token Bank
	GetAllTokenBanks() ([]models.TokenBank, error)
	GetTokenBankByUserID(userID string) (*models.TokenBank, error)
	AllocateTokens(userID string, amount int64) error
	RevokeTokens(userID string, amount int64) error
	SetTokenBalance(userID string, balance int64) error
	GetTokenBankStats() (*models.TokenBankStats, error)

	// Treasury
	GetTreasuryInfo() (*models.TokenBank, error)
	GetTreasuryBalance() (int64, error)
	AllocateFromTreasury(userID string, amount int64) error
	AllocateWelcomeBonus(userID string) error
	AllocateQuestReward(userID string, questID string, rewardAmount int64) error
	AllocateAchievementReward(userID string, achievementID string, rewardAmount int64) error

	// Token Spending (for AI, marketplace, etc.)
	SpendTokens(userID string, amount int64) error
	CheckUserBalance(userID string, requiredAmount int64) (bool, error)

	// Token Transactions History
	GetAllTransactions(limit, offset int) ([]models.TokenTransaction, error)
	GetUserTransactions(userID string, limit, offset int) ([]models.TokenTransaction, error)
	GetTransactionsByType(txType string, limit, offset int) ([]models.TokenTransaction, error)
	GetTransactionStats() (map[string]interface{}, error)

	// Ingredients Catalog
	GetAllIngredients() ([]models.Ingredient, error)
	GetIngredientsStats() (map[string]interface{}, error)

	// Ingredient Catalog Management
	BulkImportIngredients(ingredients []struct {
		Name                 string
		Unit                 string
		Category             string
		DefaultShelfLifeDays *int
		DefaultPricePerUnit  *float64
	}) (int, error)

	// Recipes Catalog
	GetAllRecipes() ([]models.RecipeCatalog, error)
	GetRecipesStats() (map[string]interface{}, error)
}

// adminService реализация интерфейса AdminService
type adminService struct {
	db *gorm.DB
}

// NewAdminService создаёт новый экземпляр сервиса администратора
func NewAdminService() AdminService {
	return &adminService{
		db: database.GetDB(),
	}
}

// GetAllUsers возвращает всех пользователей
func (s *adminService) GetAllUsers() ([]models.User, error) {
	var users []models.User
	if err := s.db.Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

// GetUsersWithFilters возвращает пользователей с фильтрацией и пагинацией
func (s *adminService) GetUsersWithFilters(params GetUsersParams) (*UserListResponse, error) {
	// Базовый запрос
	query := s.db.Model(&models.User{})

	// Применяем фильтры
	if params.Role != nil && *params.Role != "" {
		query = query.Where("role = ?", *params.Role)
	}

	if params.Status != nil && *params.Status != "" {
		query = query.Where("status = ?", *params.Status)
	}

	if params.Search != nil && *params.Search != "" {
		searchPattern := "%" + *params.Search + "%"
		query = query.Where("email ILIKE ? OR name ILIKE ?", searchPattern, searchPattern)
	}

	// Считаем total с учётом фильтров
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	// Применяем пагинацию
	offset := (params.Page - 1) * params.Limit
	query = query.Limit(params.Limit).Offset(offset)

	// Сортировка по дате создания (новые первыми)
	query = query.Order("\"createdAt\" DESC")

	// Выполняем запрос
	var users []models.User
	if err := query.Find(&users).Error; err != nil {
		return nil, err
	}

	// Вычисляем количество страниц
	totalPages := int(math.Ceil(float64(total) / float64(params.Limit)))

	return &UserListResponse{
		Users: users,
		Meta: PaginationMeta{
			Total:      int(total),
			Page:       params.Page,
			Limit:      params.Limit,
			TotalPages: totalPages,
		},
	}, nil
}

// GetUsersStats возвращает статистику по пользователям
func (s *adminService) GetUsersStats() (map[string]interface{}, error) {
	type Stats struct {
		Total       int64 `gorm:"column:total"`
		ActiveToday int64 `gorm:"column:active_today"`
		Blocked     int64 `gorm:"column:blocked"`
	}

	var stats Stats

	// Query with FILTER clause for efficient counting
	// active_today: last_login today (since 00:00)
	// blocked: status = 'blocked'
	err := s.db.Raw(`
		SELECT
			COUNT(*)                                        AS total,
			COUNT(*) FILTER (
				WHERE last_login >= DATE_TRUNC('day', NOW())
			)                                               AS active_today,
			COUNT(*) FILTER (
				WHERE status = 'blocked'
			)                                               AS blocked
		FROM "User"
	`).Scan(&stats).Error

	if err != nil {
		return nil, err
	}

	// Note: Premium users logic depends on your business model
	// Could be based on subscription, special role, or UserProfile field
	// For now, returning 0 as placeholder
	return map[string]interface{}{
		"total":        stats.Total,
		"active_today": stats.ActiveToday,
		"blocked":      stats.Blocked,
		"premium":      0, // TODO: implement premium logic when business model is defined
	}, nil
}

// UpdateUser обновляет данные пользователя (имя и email)
func (s *adminService) UpdateUser(userID string, name, email string) (*models.User, error) {
	var user models.User
	if err := s.db.First(&user, "id = ?", userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	// Обновляем только если поля не пусты
	if name != "" {
		user.Name = name
	}
	if email != "" {
		user.Email = email
	}

	if err := s.db.Save(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

// DeleteUser удаляет пользователя
func (s *adminService) DeleteUser(userID string) error {
	result := s.db.Delete(&models.User{}, "id = ?", userID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("user not found")
	}
	return nil
}

// UpdateUserRole изменяет роль пользователя
func (s *adminService) UpdateUserRole(userID, role string) error {
	// Валидация роли (все доступные роли)
	validRoles := map[string]bool{
		models.RoleHomeChef:  true,
		models.RoleProChef:   true,
		models.RoleAdmin:     true,
		models.RoleSuperAdmin: true,
	}
	if !validRoles[role] {
		return errors.New("invalid role: must be one of home_chef, pro_chef, admin, super_admin")
	}

	result := s.db.Model(&models.User{}).Where("id = ?", userID).Update("role", role)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("user not found")
	}
	return nil
}

// GetAllOrders возвращает все заказы, отсортированные по дате создания (новые в начале)
func (s *adminService) GetAllOrders() ([]models.Order, error) {
	var orders []models.Order
	if err := s.db.Order("created_at DESC").Find(&orders).Error; err != nil {
		return nil, err
	}
	return orders, nil
}

// GetRecentOrders возвращает последние N заказов
func (s *adminService) GetRecentOrders(limit int) ([]models.Order, error) {
	if limit <= 0 {
		limit = 10 // дефолтное значение
	}
	var orders []models.Order
	if err := s.db.Order("created_at DESC").Limit(limit).Find(&orders).Error; err != nil {
		return nil, err
	}
	return orders, nil
}

// UpdateOrderStatus изменяет статус заказа
func (s *adminService) UpdateOrderStatus(orderID, status string) error {
	result := s.db.Model(&models.Order{}).Where("id = ?", orderID).Update("status", status)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("order not found")
	}
	return nil
}

// GetAdminStats возвращает статистику: количество пользователей и заказов
func (s *adminService) GetAdminStats() (map[string]interface{}, error) {
	var userCount, orderCount int64

	if err := s.db.Model(&models.User{}).Count(&userCount).Error; err != nil {
		return nil, err
	}

	if err := s.db.Model(&models.Order{}).Count(&orderCount).Error; err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"totalUsers":  userCount,
		"totalOrders": orderCount,
	}, nil
}

// GetAdminProfile возвращает профиль администратора с информацией о пользователе
func (s *adminService) GetAdminProfile(adminID string) (map[string]interface{}, error) {
	var user models.User
	if err := s.db.First(&user, "id = ?", adminID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("admin not found")
		}
		return nil, err
	}

	// Подсчитаем количество управляемых пользователей (всех пользователей в системе)
	var userCount int64
	if err := s.db.Model(&models.User{}).Count(&userCount).Error; err != nil {
		return nil, err
	}

	// Подсчитаем количество управляемых заказов
	var orderCount int64
	if err := s.db.Model(&models.Order{}).Count(&orderCount).Error; err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"id":            user.ID,
		"name":          user.Name,
		"email":         user.Email,
		"role":          user.Role,
		"createdAt":     user.CreatedAt,
		"managedUsers":  userCount,
		"managedOrders": orderCount,
		"totalStats":    map[string]interface{}{"users": userCount, "orders": orderCount},
	}, nil
}

// GetAllTokenBanks возвращает все записи токин-банков
func (s *adminService) GetAllTokenBanks() ([]models.TokenBank, error) {
	repo := &database.TokenBankRepository{}
	tokenBanks, err := repo.FindAll()
	if err != nil {
		return nil, err
	}
	// Если нет данных, возвращаем пустой массив вместо nil
	if len(tokenBanks) == 0 {
		return []models.TokenBank{}, nil
	}
	return tokenBanks, nil
}

// GetTokenBankByUserID возвращает токин-банк пользователя
func (s *adminService) GetTokenBankByUserID(userID string) (*models.TokenBank, error) {
	repo := &database.TokenBankRepository{}
	return repo.FindByUserID(userID)
}

// AllocateTokens выделяет токины пользователю
func (s *adminService) AllocateTokens(userID string, amount int64) error {
	repo := &database.TokenBankRepository{}
	return repo.AllocateTokens(userID, amount)
}

// RevokeTokens отзывает токины у пользователя
func (s *adminService) RevokeTokens(userID string, amount int64) error {
	repo := &database.TokenBankRepository{}
	return repo.RevokeTokens(userID, amount)
}

// SetTokenBalance устанавливает точное значение баланса токинов
func (s *adminService) SetTokenBalance(userID string, balance int64) error {
	repo := &database.TokenBankRepository{}
	return repo.SetBalance(userID, balance)
}

// GetTokenBankStats возвращает статистику по токинам
func (s *adminService) GetTokenBankStats() (*models.TokenBankStats, error) {
	repo := &database.TokenBankRepository{}
	stats, err := repo.GetTokenBankStats()
	if err != nil {
		// Если ошибка, возвращаем пустую статистику вместо ошибки
		return &models.TokenBankStats{
			TotalTokensAllocated:  0,
			TotalTokensUsed:       0,
			TotalUsersWithTokens:  0,
			AverageBalancePerUser: 0,
		}, nil
	}
	return stats, nil
}

// GetTreasuryInfo возвращает информацию о казначействе
func (s *adminService) GetTreasuryInfo() (*models.TokenBank, error) {
	repo := &database.TokenBankRepository{}
	return repo.GetTreasuryInfo()
}

// GetTreasuryBalance возвращает баланс казначейства
func (s *adminService) GetTreasuryBalance() (int64, error) {
	repo := &database.TokenBankRepository{}
	treasury, err := repo.GetTreasury()
	if err != nil {
		return 0, err
	}
	return treasury.Balance, nil
}

// AllocateFromTreasury выделяет токены из казначейства пользователю
func (s *adminService) AllocateFromTreasury(userID string, amount int64) error {
	repo := &database.TokenBankRepository{}
	return repo.AllocateFromTreasury(userID, amount)
}

// AllocateWelcomeBonus выделяет приветственный бонус новому пользователю (по умолчанию 100 токенов)
func (s *adminService) AllocateWelcomeBonus(userID string) error {
	repo := &database.TokenBankRepository{}
	return repo.AllocateWelcomeBonus(userID, 100)
}

// AllocateQuestReward выделяет награду за выполнение квеста
func (s *adminService) AllocateQuestReward(userID string, questID string, rewardAmount int64) error {
	repo := &database.TokenBankRepository{}
	return repo.AllocateQuestReward(userID, questID, rewardAmount)
}

// AllocateAchievementReward выделяет награду за достижение
func (s *adminService) AllocateAchievementReward(userID string, achievementID string, rewardAmount int64) error {
	repo := &database.TokenBankRepository{}
	return repo.AllocateAchievementReward(userID, achievementID, rewardAmount)
}

// SpendTokens списывает токены у пользователя и возвращает их в казначейство
// Используется для оплаты AI-запросов, покупок в маркетплейсе и других расходов
func (s *adminService) SpendTokens(userID string, amount int64) error {
	repo := &database.TokenBankRepository{}
	return repo.SpendTokens(userID, amount)
}

// CheckUserBalance проверяет, достаточно ли токенов у пользователя
func (s *adminService) CheckUserBalance(userID string, requiredAmount int64) (bool, error) {
	repo := &database.TokenBankRepository{}
	return repo.CheckUserBalance(userID, requiredAmount)
}

// ============================================
// Token Transactions History
// ============================================

// GetAllTransactions получает все транзакции с пагинацией
func (s *adminService) GetAllTransactions(limit, offset int) ([]models.TokenTransaction, error) {
	repo := &database.TokenTransactionRepository{}
	return repo.GetAllTransactions(limit, offset)
}

// GetUserTransactions получает транзакции конкретного пользователя
func (s *adminService) GetUserTransactions(userID string, limit, offset int) ([]models.TokenTransaction, error) {
	repo := &database.TokenTransactionRepository{}
	return repo.GetUserTransactions(userID, limit, offset)
}

// GetTransactionsByType получает транзакции по типу
func (s *adminService) GetTransactionsByType(txType string, limit, offset int) ([]models.TokenTransaction, error) {
	repo := &database.TokenTransactionRepository{}
	return repo.GetTransactionsByType(txType, limit, offset)
}

// GetTransactionStats получает статистику транзакций
func (s *adminService) GetTransactionStats() (map[string]interface{}, error) {
	repo := &database.TokenTransactionRepository{}
	return repo.GetTransactionStats()
}

// BulkImportIngredients импортирует ингредиенты массово
func (s *adminService) BulkImportIngredients(ingredients []struct {
	Name                 string
	Unit                 string
	Category             string
	DefaultShelfLifeDays *int
	DefaultPricePerUnit  *float64
}) (int, error) {
	imported := 0

	for _, ing := range ingredients {
		// Проверка обязательных полей
		if ing.Name == "" || ing.Unit == "" || ing.Category == "" {
			continue // Пропускаем невалидные записи
		}

		// Создаём модель
		ingredient := models.Ingredient{
			Name:                 ing.Name,
			Unit:                 ing.Unit,
			Category:             ing.Category,
			DefaultShelfLifeDays: ing.DefaultShelfLifeDays,
			DefaultPricePerUnit:  ing.DefaultPricePerUnit,
		}

		// Upsert: если существует с таким именем - обновляем, иначе создаём
		result := s.db.Where("name = ?", ing.Name).
			Assign(ingredient).
			FirstOrCreate(&ingredient)

		if result.Error != nil {
			continue // Логируем, но продолжаем
		}

		imported++
	}

	return imported, nil
}

// GetAllIngredients возвращает весь каталог ингредиентов
func (s *adminService) GetAllIngredients() ([]models.Ingredient, error) {
	var ingredients []models.Ingredient
	
	if err := s.db.Order("category ASC, name ASC").Find(&ingredients).Error; err != nil {
		return nil, err
	}
	
	return ingredients, nil
}

// GetIngredientsStats возвращает статистику по ингредиентам
func (s *adminService) GetIngredientsStats() (map[string]interface{}, error) {
	var total int64
	if err := s.db.Model(&models.Ingredient{}).Count(&total).Error; err != nil {
		return nil, err
	}

	// Статистика по категориям
	var categoryStats []struct {
		Category string `json:"category"`
		Count    int64  `json:"count"`
	}
	
	if err := s.db.Model(&models.Ingredient{}).
		Select("category, COUNT(*) as count").
		Group("category").
		Order("count DESC").
		Scan(&categoryStats).Error; err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"total":      total,
		"categories": categoryStats,
	}, nil
}

// GetAllRecipes возвращает все рецепты из каталога с ингредиентами
func (s *adminService) GetAllRecipes() ([]models.RecipeCatalog, error) {
	var recipes []models.RecipeCatalog
	
	// Загружаем рецепты с ингредиентами и связанными данными
	if err := s.db.
		Preload("Ingredients.Ingredient"). // Загружаем ингредиенты с их полной информацией
		Preload("Allergens").              // Загружаем аллергены
		Preload("DietTags").               // Загружаем диетические теги
		Order("category ASC, \"canonicalName\" ASC").
		Find(&recipes).Error; err != nil {
		return nil, err
	}
	
	return recipes, nil
}

// GetRecipesStats возвращает статистику по рецептам
func (s *adminService) GetRecipesStats() (map[string]interface{}, error) {
	var total int64
	if err := s.db.Model(&models.RecipeCatalog{}).Count(&total).Error; err != nil {
		return nil, err
	}

	// Статистика по категориям
	var categoryStats []struct {
		Category string `json:"category"`
		Count    int64  `json:"count"`
	}
	
	if err := s.db.Model(&models.RecipeCatalog{}).
		Select("category, COUNT(*) as count").
		Group("category").
		Order("count DESC").
		Scan(&categoryStats).Error; err != nil {
		return nil, err
	}

	// Статистика по сложности
	var difficultyStats []struct {
		Difficulty string `json:"difficulty"`
		Count      int64  `json:"count"`
	}
	
	if err := s.db.Model(&models.RecipeCatalog{}).
		Select("difficulty, COUNT(*) as count").
		Group("difficulty").
		Order("count DESC").
		Scan(&difficultyStats).Error; err != nil {
		return nil, err
	}

	// Статистика по странам
	var countryStats []struct {
		Country string `json:"country"`
		Count   int64  `json:"count"`
	}
	
	if err := s.db.Model(&models.RecipeCatalog{}).
		Select("country, COUNT(*) as count").
		Group("country").
		Order("count DESC").
		Scan(&countryStats).Error; err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"total":       total,
		"categories":  categoryStats,
		"difficulty":  difficultyStats,
		"countries":   countryStats,
	}, nil
}
