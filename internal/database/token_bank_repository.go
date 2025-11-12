package database

import (
	"errors"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
	"gorm.io/gorm"
)

// TokenBankRepository репозиторий для работы с банком токинов
type TokenBankRepository struct{}

// FindByUserID находит запись токин-банка по ID пользователя
func (r *TokenBankRepository) FindByUserID(userID string) (*models.TokenBank, error) {
	var tokenBank models.TokenBank
	result := DB.Where("user_id = ?", userID).First(&tokenBank)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("token bank not found for user")
		}
		return nil, result.Error
	}
	return &tokenBank, nil
}

// FindByID находит запись токин-банка по ID
func (r *TokenBankRepository) FindByID(id string) (*models.TokenBank, error) {
	var tokenBank models.TokenBank
	result := DB.First(&tokenBank, "id = ?", id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &tokenBank, nil
}

// FindAll возвращает все записи токин-банка с информацией о пользователях
func (r *TokenBankRepository) FindAll() ([]models.TokenBank, error) {
	var tokenBanks []models.TokenBank
	result := DB.Find(&tokenBanks)
	if result.Error != nil {
		return nil, result.Error
	}
	return tokenBanks, nil
}

// Create создает новую запись токин-банка
func (r *TokenBankRepository) Create(tokenBank *models.TokenBank) error {
	result := DB.Create(tokenBank)
	return result.Error
}

// Update обновляет запись токин-банка
func (r *TokenBankRepository) Update(tokenBank *models.TokenBank) error {
	result := DB.Save(tokenBank)
	return result.Error
}

// AllocateTokens выделяет токины пользователю (увеличивает balance и total_allocated)
func (r *TokenBankRepository) AllocateTokens(userID string, amount int64) error {
	if amount <= 0 {
		return errors.New("amount must be positive")
	}

	result := DB.Model(&models.TokenBank{}).
		Where("user_id = ?", userID).
		Updates(map[string]interface{}{
			"balance":         gorm.Expr("balance + ?", amount),
			"total_allocated": gorm.Expr("total_allocated + ?", amount),
		})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("token bank not found for user")
	}
	return nil
}

// RevokeTokens отзывает токины у пользователя (уменьшает balance)
func (r *TokenBankRepository) RevokeTokens(userID string, amount int64) error {
	if amount <= 0 {
		return errors.New("amount must be positive")
	}

	// Сначала проверяем баланс
	tb, err := r.FindByUserID(userID)
	if err != nil {
		return err
	}

	if tb.Balance < amount {
		return errors.New("insufficient tokens")
	}

	result := DB.Model(&models.TokenBank{}).
		Where("user_id = ?", userID).
		Update("balance", gorm.Expr("balance - ?", amount))

	if result.Error != nil {
		return result.Error
	}
	return nil
}

// UseTokens использует токины (уменьшает balance и увеличивает total_used)
func (r *TokenBankRepository) UseTokens(userID string, amount int64) error {
	if amount <= 0 {
		return errors.New("amount must be positive")
	}

	// Сначала проверяем баланс
	tb, err := r.FindByUserID(userID)
	if err != nil {
		return err
	}

	if tb.Balance < amount {
		return errors.New("insufficient tokens")
	}

	result := DB.Model(&models.TokenBank{}).
		Where("user_id = ?", userID).
		Updates(map[string]interface{}{
			"balance":    gorm.Expr("balance - ?", amount),
			"total_used": gorm.Expr("total_used + ?", amount),
		})

	if result.Error != nil {
		return result.Error
	}
	return nil
}

// SetBalance устанавливает точное значение баланса
func (r *TokenBankRepository) SetBalance(userID string, balance int64) error {
	if balance < 0 {
		return errors.New("balance cannot be negative")
	}

	result := DB.Model(&models.TokenBank{}).
		Where("user_id = ?", userID).
		Update("balance", balance)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("token bank not found for user")
	}
	return nil
}

// GetTokenBankStats возвращает статистику по токинам
func (r *TokenBankRepository) GetTokenBankStats() (*models.TokenBankStats, error) {
	var stats models.TokenBankStats

	// Получаем сумму всех выделенных токинов
	result := DB.Model(&models.TokenBank{}).
		Select("COALESCE(SUM(total_allocated), 0) as total_tokens_allocated",
			"COALESCE(SUM(total_used), 0) as total_tokens_used",
			"COUNT(DISTINCT user_id) as total_users_with_tokens").
		Scan(&stats)

	if result.Error != nil {
		return nil, result.Error
	}

	// Вычисляем средний баланс
	if stats.TotalUsersWithTokens > 0 {
		var totalBalance int64
		result := DB.Model(&models.TokenBank{}).
			Select("COALESCE(SUM(balance), 0)").
			Row().
			Scan(&totalBalance)
		if result != nil {
			return nil, result
		}
		stats.AverageBalancePerUser = float64(totalBalance) / float64(stats.TotalUsersWithTokens)
	}

	return &stats, nil
}

// InitializeTokenBankForUser инициализирует токин-банк для нового пользователя
func (r *TokenBankRepository) InitializeTokenBankForUser(userID string) error {
	tokenBank := &models.TokenBank{
		UserID:         userID,
		Balance:        0,
		TotalAllocated: 0,
		TotalUsed:      0,
	}
	return r.Create(tokenBank)
}

// Delete удаляет запись токин-банка
func (r *TokenBankRepository) Delete(id string) error {
	result := DB.Delete(&models.TokenBank{}, "id = ?", id)
	return result.Error
}
