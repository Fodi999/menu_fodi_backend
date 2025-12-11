package database

import (
	"errors"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
	"gorm.io/gorm"
)

// TreasuryUserID константа для ID системного казначейства
const TreasuryUserID = "TREASURY"

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
// и автоматически выдаёт приветственный бонус 100 токенов из Treasury
func (r *TokenBankRepository) InitializeTokenBankForUser(userID string) error {
	// Создаём токин-банк с нулевым балансом
	tokenBank := &models.TokenBank{
		UserID:         userID,
		Balance:        0,
		TotalAllocated: 0,
		TotalUsed:      0,
	}
	
	if err := r.Create(tokenBank); err != nil {
		return err
	}

	// Автоматически выдаём приветственный бонус 100 токенов из Treasury
	welcomeBonus := int64(100)
	if err := r.AllocateFromTreasury(userID, welcomeBonus); err != nil {
		// Логируем ошибку, но не блокируем создание пользователя
		// Токин-банк создан, но бонус не выдан - администратор может выдать вручную
		return nil
	}

	return nil
}

// Delete удаляет запись токин-банка
func (r *TokenBankRepository) Delete(id string) error {
	result := DB.Delete(&models.TokenBank{}, "id = ?", id)
	return result.Error
}

// GetTreasuryBalance возвращает баланс казначейства
func (r *TokenBankRepository) GetTreasuryBalance() (int64, error) {
	var treasury models.TokenBank
	result := DB.Where("user_id = ?", TreasuryUserID).First(&treasury)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return 0, errors.New("treasury not found")
		}
		return 0, result.Error
	}
	return treasury.Balance, nil
}

// AllocateFromTreasury выделяет токены из казначейства пользователю
func (r *TokenBankRepository) AllocateFromTreasury(userID string, amount int64) error {
	if amount <= 0 {
		return errors.New("amount must be positive")
	}

	// Начинаем транзакцию
	return DB.Transaction(func(tx *gorm.DB) error {
		// 1. Проверяем баланс казначейства
		var treasury models.TokenBank
		if err := tx.Where("user_id = ?", TreasuryUserID).First(&treasury).Error; err != nil {
			return errors.New("treasury not found")
		}

		if treasury.Balance < amount {
			return errors.New("insufficient treasury balance")
		}

		// 2. Уменьшаем баланс казначейства и увеличиваем total_used
		if err := tx.Model(&models.TokenBank{}).
			Where("user_id = ?", TreasuryUserID).
			Updates(map[string]interface{}{
				"balance":    gorm.Expr("balance - ?", amount),
				"total_used": gorm.Expr("total_used + ?", amount),
			}).Error; err != nil {
			return err
		}

		// 3. Увеличиваем баланс пользователя
		result := tx.Model(&models.TokenBank{}).
			Where("user_id = ?", userID).
			Updates(map[string]interface{}{
				"balance":         gorm.Expr("balance + ?", amount),
				"total_allocated": gorm.Expr("total_allocated + ?", amount),
			})

		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return errors.New("user token bank not found")
		}

		return nil
	})
}

// GetTreasuryInfo возвращает полную информацию о казначействе
func (r *TokenBankRepository) GetTreasuryInfo() (*models.TokenBank, error) {
	return r.FindByUserID(TreasuryUserID)
}

// AllocateWelcomeBonus автоматически выделяет приветственный бонус новому пользователю из казначейства
func (r *TokenBankRepository) AllocateWelcomeBonus(userID string, bonusAmount int64) error {
	if bonusAmount <= 0 {
		bonusAmount = 100 // Дефолтный приветственный бонус
	}
	return r.AllocateFromTreasury(userID, bonusAmount)
}

// AllocateQuestReward выделяет награду за выполнение квеста из казначейства
func (r *TokenBankRepository) AllocateQuestReward(userID string, questID string, rewardAmount int64) error {
	if rewardAmount <= 0 {
		return errors.New("reward amount must be positive")
	}
	return r.AllocateFromTreasury(userID, rewardAmount)
}

// AllocateAchievementReward выделяет награду за достижение из казначейства
func (r *TokenBankRepository) AllocateAchievementReward(userID string, achievementID string, rewardAmount int64) error {
	if rewardAmount <= 0 {
		return errors.New("reward amount must be positive")
	}
	return r.AllocateFromTreasury(userID, rewardAmount)
}

// CheckTreasuryBalance проверяет, достаточно ли токенов в казначействе
func (r *TokenBankRepository) CheckTreasuryBalance(requiredAmount int64) (bool, error) {
	balance, err := r.GetTreasuryBalance()
	if err != nil {
		return false, err
	}
	return balance >= requiredAmount, nil
}

// SpendTokens списывает токены у пользователя и возвращает их в казначейство
// Используется для оплаты AI-запросов, покупок в маркетплейсе и других расходов
// Это противоположность AllocateFromTreasury - замыкает цикл токен-экономики
func (r *TokenBankRepository) SpendTokens(userID string, amount int64) error {
	if amount <= 0 {
		return errors.New("amount must be positive")
	}

	// Атомарная транзакция для безопасного списания токенов
	return DB.Transaction(func(tx *gorm.DB) error {
		// 1. Получаем токен-банк пользователя
		var userBank models.TokenBank
		if err := tx.Where("user_id = ?", userID).First(&userBank).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("user token bank not found")
			}
			return err
		}

		// 2. Проверяем, достаточно ли токенов у пользователя
		if userBank.Balance < amount {
			return errors.New("insufficient tokens")
		}

		// 3. Списываем токены у пользователя
		if err := tx.Model(&models.TokenBank{}).
			Where("user_id = ?", userID).
			Updates(map[string]interface{}{
				"balance":    gorm.Expr("balance - ?", amount),
				"total_used": gorm.Expr("total_used + ?", amount),
			}).Error; err != nil {
			return err
		}

		// 4. Возвращаем токены в казначейство (замыкаем цикл)
		var treasury models.TokenBank
		if err := tx.Where("user_id = ?", TreasuryUserID).First(&treasury).Error; err != nil {
			return errors.New("treasury not found")
		}

		// 5. Увеличиваем баланс казначейства (токены возвращаются в систему)
		if err := tx.Model(&models.TokenBank{}).
			Where("user_id = ?", TreasuryUserID).
			Updates(map[string]interface{}{
				"balance": gorm.Expr("balance + ?", amount),
				// total_used казначейства уменьшается, так как токены вернулись
				"total_used": gorm.Expr("total_used - ?", amount),
			}).Error; err != nil {
			return err
		}

		return nil
	})
}

// CheckUserBalance проверяет, достаточно ли токенов у пользователя для расхода
func (r *TokenBankRepository) CheckUserBalance(userID string, requiredAmount int64) (bool, error) {
	userBank, err := r.FindByUserID(userID)
	if err != nil {
		return false, err
	}
	return userBank.Balance >= requiredAmount, nil
}
