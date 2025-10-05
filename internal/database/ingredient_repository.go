package database

import (
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
)

// IngredientRepository репозиторий для работы с ингредиентами
type IngredientRepository struct{}

// FindAll возвращает все ингредиенты со складскими остатками
func (r *IngredientRepository) FindAll() ([]models.StockItem, error) {
	var stockItems []models.StockItem
	result := DB.Preload("Ingredient").Find(&stockItems)
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
