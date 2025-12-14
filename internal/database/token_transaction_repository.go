package database

import (
	"time"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
)

// TokenTransactionRepository управляет операциями с историей токенов
type TokenTransactionRepository struct{}

// LogTransaction записывает транзакцию в историю
func (r *TokenTransactionRepository) LogTransaction(tx *models.TokenTransaction) error {
	return DB.Create(tx).Error
}

// LogTreasuryAllocation записывает выделение токенов из Treasury
func (r *TokenTransactionRepository) LogTreasuryAllocation(toUserID string, amount int64, txType, description string, metadata map[string]interface{}) error {
	tx := &models.TokenTransaction{
		FromUserID:  nil, // NULL = Treasury
		ToUserID:    &toUserID,
		Amount:      amount,
		Type:        txType,
		Description: description,
		Metadata:    metadata,
	}
	return r.LogTransaction(tx)
}

// LogTreasurySpending записывает трату токенов (возврат в Treasury или burn)
func (r *TokenTransactionRepository) LogTreasurySpending(fromUserID string, amount int64, txType, description string, metadata map[string]interface{}) error {
	tx := &models.TokenTransaction{
		FromUserID:  &fromUserID,
		ToUserID:    nil, // NULL = Treasury/Burn
		Amount:      amount,
		Type:        txType,
		Description: description,
		Metadata:    metadata,
	}
	return r.LogTransaction(tx)
}

// LogUserTransfer записывает перевод между пользователями
func (r *TokenTransactionRepository) LogUserTransfer(fromUserID, toUserID string, amount int64, description string) error {
	tx := &models.TokenTransaction{
		FromUserID:  &fromUserID,
		ToUserID:    &toUserID,
		Amount:      amount,
		Type:        models.TransactionTypeUserTransfer,
		Description: description,
	}
	return r.LogTransaction(tx)
}

// GetAllTransactions получает все транзакции с пагинацией
func (r *TokenTransactionRepository) GetAllTransactions(limit, offset int) ([]models.TokenTransaction, error) {
	var transactions []models.TokenTransaction
	err := DB.
		Preload("FromUser").
		Preload("ToUser").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&transactions).Error
	return transactions, err
}

// GetUserTransactions получает транзакции конкретного пользователя
func (r *TokenTransactionRepository) GetUserTransactions(userID string, limit, offset int) ([]models.TokenTransaction, error) {
	var transactions []models.TokenTransaction
	err := DB.
		Preload("FromUser").
		Preload("ToUser").
		Where("from_user_id = ? OR to_user_id = ?", userID, userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&transactions).Error
	return transactions, err
}

// GetTransactionsByType получает транзакции по типу
func (r *TokenTransactionRepository) GetTransactionsByType(txType string, limit, offset int) ([]models.TokenTransaction, error) {
	var transactions []models.TokenTransaction
	err := DB.
		Preload("FromUser").
		Preload("ToUser").
		Where("type = ?", txType).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&transactions).Error
	return transactions, err
}

// GetTransactionsByPeriod получает транзакции за период
func (r *TokenTransactionRepository) GetTransactionsByPeriod(startDate, endDate time.Time, limit, offset int) ([]models.TokenTransaction, error) {
	var transactions []models.TokenTransaction
	err := DB.
		Preload("FromUser").
		Preload("ToUser").
		Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&transactions).Error
	return transactions, err
}

// GetTransactionsWithFilters получает транзакции с множественными фильтрами
func (r *TokenTransactionRepository) GetTransactionsWithFilters(
	userID *string,
	txType *string,
	startDate *time.Time,
	endDate *time.Time,
	limit, offset int,
) ([]models.TokenTransaction, error) {
	query := DB.Preload("FromUser").Preload("ToUser")

	if userID != nil {
		query = query.Where("from_user_id = ? OR to_user_id = ?", *userID, *userID)
	}

	if txType != nil {
		query = query.Where("type = ?", *txType)
	}

	if startDate != nil && endDate != nil {
		query = query.Where("created_at BETWEEN ? AND ?", *startDate, *endDate)
	}

	var transactions []models.TokenTransaction
	err := query.
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&transactions).Error

	return transactions, err
}

// GetTreasuryAllocations получает все выделения из Treasury
func (r *TokenTransactionRepository) GetTreasuryAllocations(limit, offset int) ([]models.TokenTransaction, error) {
	var transactions []models.TokenTransaction
	err := DB.
		Preload("ToUser").
		Where("from_user_id IS NULL AND to_user_id IS NOT NULL").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&transactions).Error
	return transactions, err
}

// GetTreasurySpending получает все траты в Treasury
func (r *TokenTransactionRepository) GetTreasurySpending(limit, offset int) ([]models.TokenTransaction, error) {
	var transactions []models.TokenTransaction
	err := DB.
		Preload("FromUser").
		Where("from_user_id IS NOT NULL AND to_user_id IS NULL").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&transactions).Error
	return transactions, err
}

// GetTransactionStats получает статистику транзакций
func (r *TokenTransactionRepository) GetTransactionStats() (map[string]interface{}, error) {
	var stats struct {
		TotalTransactions int64
		TotalAllocated    int64
		TotalSpent        int64
		UniqueUsers       int64
	}

	// Общее количество транзакций
	DB.Model(&models.TokenTransaction{}).Count(&stats.TotalTransactions)

	// Всего выделено из Treasury
	DB.Model(&models.TokenTransaction{}).
		Where("from_user_id IS NULL AND to_user_id IS NOT NULL").
		Select("COALESCE(SUM(amount), 0)").
		Scan(&stats.TotalAllocated)

	// Всего потрачено
	DB.Model(&models.TokenTransaction{}).
		Where("from_user_id IS NOT NULL AND to_user_id IS NULL").
		Select("COALESCE(SUM(amount), 0)").
		Scan(&stats.TotalSpent)

	// Уникальные пользователи
	DB.Model(&models.TokenTransaction{}).
		Where("to_user_id IS NOT NULL").
		Distinct("to_user_id").
		Count(&stats.UniqueUsers)

	return map[string]interface{}{
		"total_transactions": stats.TotalTransactions,
		"total_allocated":    stats.TotalAllocated,
		"total_spent":        stats.TotalSpent,
		"unique_users":       stats.UniqueUsers,
	}, nil
}
