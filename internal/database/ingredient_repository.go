package database

import (
	"strings"
	"unicode"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
	"golang.org/x/text/unicode/norm"
)

// IngredientRepository репозиторий для работы с ингредиентами
type IngredientRepository struct{}

// normalizeSearchQuery removes diacritics and lowercases the query
// Works for Polish (ą→a, ł→l) and other languages
func normalizeSearchQuery(s string) string {
	// Step 1: Lowercase
	s = strings.ToLower(s)
	
	// Step 2: Only remove diacritics from Latin characters (ą→a, ę→e)
	// Preserve Cyrillic, Greek, and other non-Latin scripts
	result := ""
	for _, r := range s {
		// If character is Cyrillic (U+0400 to U+04FF) or other non-Latin, keep as-is
		if (r >= 0x0400 && r <= 0x04FF) || !unicode.Is(unicode.Latin, r) {
			result += string(r)
		} else {
			// For Latin characters, decompose and remove diacritics
			normalized := norm.NFD.String(string(r))
			for _, nr := range normalized {
				if !unicode.Is(unicode.Mn, nr) { // Keep only non-mark characters
					result += string(nr)
				}
			}
		}
	}
	
	return result
}

// FindAll возвращает все ингредиенты со складскими остатками
func (r *IngredientRepository) FindAll() ([]models.StockItem, error) {
	var stockItems []models.StockItem
	result := DB.
		Preload("Ingredient").
		Order(`"updatedAt" DESC`).
		Find(&stockItems)
	if result.Error != nil {
		return nil, result.Error
	}
	return stockItems, nil
}

// FindByID находит складской остаток по ID
func (r *IngredientRepository) FindByID(id string) (*models.StockItem, error) {
	var stockItem models.StockItem
	result := DB.Preload("Ingredient").First(&stockItem, "id = ?", id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &stockItem, nil
}

// CreateIngredient создает новый ингредиент и складской остаток
func (r *IngredientRepository) CreateIngredient(ingredient *models.Ingredient, stockItem *models.StockItem) error {
	// Создаем в транзакции
	tx := DB.Begin()

	// Создаем ингредиент
	if err := tx.Create(ingredient).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Создаем складской остаток
	stockItem.IngredientID = ingredient.ID
	if err := tx.Create(stockItem).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// UpdateStockItem обновляет складской остаток
func (r *IngredientRepository) UpdateStockItem(stockItem *models.StockItem) error {
	result := DB.Save(stockItem)
	return result.Error
}

// UpdateIngredient обновляет ингредиент
func (r *IngredientRepository) UpdateIngredient(ingredient *models.Ingredient) error {
	result := DB.Save(ingredient)
	return result.Error
}

// DeleteStockItem удаляет складской остаток и ингредиент
func (r *IngredientRepository) DeleteStockItem(id string) error {
	// Находим складской остаток
	var stockItem models.StockItem
	if err := DB.First(&stockItem, "id = ?", id).Error; err != nil {
		return err
	}

	// Удаляем в транзакции
	tx := DB.Begin()

	// Удаляем складской остаток
	if err := tx.Delete(&stockItem).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Удаляем ингредиент
	if err := tx.Delete(&models.Ingredient{}, "id = ?", stockItem.IngredientID).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// Search ищет ингредиенты по имени (для автокомплита)
// Используется ВСЕМИ пользователями: home_chef, pro_chef, admin
// Multilingual search: поддерживает PL/EN/RU языки
func (r *IngredientRepository) Search(query string) ([]models.Ingredient, error) {
	var ingredients []models.Ingredient

	// Нормализуем поисковый запрос в Go (без SQL unaccent)
	// Убирает диакритику (ą→a, ł→l, ę→e) и приводит к lowercase
	normalizedQuery := normalizeSearchQuery(query) + "%"

	// MULTILINGUAL QUERY: поиск по всем языковым колонкам
	// Поддерживает: Polish (name_pl), English (name_en), Russian (name_ru), legacy (name)
	// Примеры:
	//   "огур" → найдет "огурец" (name_ru)
	//   "cucum" → найдет "cucumber" (name_en)
	//   "ogór" → найдет "ogórek" (name_pl или name)
	result := DB.
		Where("LOWER(name) LIKE ? OR LOWER(name_pl) LIKE ? OR LOWER(name_en) LIKE ? OR LOWER(name_ru) LIKE ?",
			normalizedQuery, normalizedQuery, normalizedQuery, normalizedQuery).
		Order("name ASC").
		Limit(20).
		Find(&ingredients)

	if result.Error != nil {
		return nil, result.Error
	}

	return ingredients, nil
}

// GetAllIngredients возвращает все ингредиенты из каталога
func (r *IngredientRepository) GetAllIngredients() ([]models.Ingredient, error) {
	var ingredients []models.Ingredient
	result := DB.Order("name ASC").Find(&ingredients)
	if result.Error != nil {
		return nil, result.Error
	}
	return ingredients, nil
}

// List возвращает ингредиенты с фильтрацией по категории и поиском
// Используется для просмотра каталога с фильтрами
func (r *IngredientRepository) List(category, search string) ([]models.Ingredient, error) {
	var ingredients []models.Ingredient

	query := DB.Model(&models.Ingredient{})

	// Фильтр по категории
	if category != "" {
		query = query.Where("category = ?", category)
	}

	// Поиск по префиксу имени (регистронезависимый)
	if search != "" {
		searchPattern := strings.ToLower(search) + "%"
		query = query.Where("LOWER(name) LIKE ?", searchPattern)
	}

	result := query.
		Order("name ASC").
		Limit(250).
		Find(&ingredients)

	if result.Error != nil {
		return nil, result.Error
	}

	return ingredients, nil
}

// GetByCategory возвращает ингредиенты по категории
func (r *IngredientRepository) GetByCategory(category string) ([]models.Ingredient, error) {
	var ingredients []models.Ingredient

	result := DB.
		Where("category = ?", category).
		Order("name ASC").
		Limit(250).
		Find(&ingredients)

	if result.Error != nil {
		return nil, result.Error
	}

	return ingredients, nil
}

// GetIngredientByID возвращает ингредиент из каталога по ID
func (r *IngredientRepository) GetIngredientByID(id string) (*models.Ingredient, error) {
	var ingredient models.Ingredient
	result := DB.Where("id = ?", id).First(&ingredient)
	if result.Error != nil {
		return nil, result.Error
	}
	return &ingredient, nil
}
