package service

import (
	"fmt"
	"log"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/database"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
	tokenservice "github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/wallet/service"
	"github.com/google/uuid"
)

// SubscriptionService - сервис для работы с подписками и инвестициями
type SubscriptionService struct {
	tokenService *tokenservice.TokenService
}

// NewSubscriptionService создает новый экземпляр SubscriptionService
func NewSubscriptionService() *SubscriptionService {
	return &SubscriptionService{
		tokenService: tokenservice.NewTokenService(),
	}
}

// Subscribe - пользователь инвестирует в бизнес (покупает токены)
func (s *SubscriptionService) Subscribe(userID, businessID string, tokensAmount int64) (*models.BusinessSubscription, *models.Transaction, error) {
	db := database.GetDB()

	// Проверка существования бизнеса
	var business models.Business
	if err := db.First(&business, "id = ?", businessID).Error; err != nil {
		return nil, nil, fmt.Errorf("business not found: %w", err)
	}

	// Проверка существования токена бизнеса
	var businessToken models.BusinessToken
	if err := db.Where("business_id = ?", businessID).First(&businessToken).Error; err != nil {
		return nil, nil, fmt.Errorf("business token not found: %w", err)
	}

	// Проверка доступности токенов
	if businessToken.TotalSupply < tokensAmount {
		return nil, nil, fmt.Errorf("insufficient token supply: available %d, requested %d", businessToken.TotalSupply, tokensAmount)
	}

	// Рассчитываем стоимость инвестиции
	investmentAmount := businessToken.Price * float64(tokensAmount)

	// Начало транзакции БД
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Проверка существующей подписки
	var subscription models.BusinessSubscription
	err := tx.Where("user_id = ? AND business_id = ?", userID, businessID).First(&subscription).Error

	if err != nil {
		// Создаем новую подписку
		subscription = models.BusinessSubscription{
			ID:          uuid.New().String(),
			UserID:      userID,
			BusinessID:  businessID,
			TokensOwned: tokensAmount,
			Invested:    investmentAmount,
		}

		if err := tx.Create(&subscription).Error; err != nil {
			tx.Rollback()
			return nil, nil, fmt.Errorf("failed to create subscription: %w", err)
		}
	} else {
		// Обновляем существующую подписку
		subscription.TokensOwned += tokensAmount
		subscription.Invested += investmentAmount

		if err := tx.Save(&subscription).Error; err != nil {
			tx.Rollback()
			return nil, nil, fmt.Errorf("failed to update subscription: %w", err)
		}
	}

	// Создаем транзакцию покупки
	transaction := models.Transaction{
		ID:         uuid.New().String(),
		BusinessID: businessID,
		FromUser:   userID,
		ToUser:     business.OwnerID, // Владелец бизнеса получает деньги
		Tokens:     tokensAmount,
		Amount:     investmentAmount,
		TxType:     "buy",
	}

	if err := tx.Create(&transaction).Error; err != nil {
		tx.Rollback()
		return nil, nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	// Уменьшаем supply токенов (пользователь "покупает" их из доступного supply)
	businessToken.TotalSupply -= tokensAmount
	if err := tx.Save(&businessToken).Error; err != nil {
		tx.Rollback()
		return nil, nil, fmt.Errorf("failed to update token supply: %w", err)
	}

	// Пересчитываем цену токена после покупки (больше инвесторов = выше цена)
	newPrice := s.tokenService.CalculateTokenPrice(businessID, businessToken.TotalSupply)
	businessToken.Price = newPrice
	if err := tx.Save(&businessToken).Error; err != nil {
		tx.Rollback()
		return nil, nil, fmt.Errorf("failed to update token price: %w", err)
	}

	// Коммит транзакции
	if err := tx.Commit().Error; err != nil {
		return nil, nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("[SUBSCRIPTION] ✅ User %s invested $%.2f in business %s (%s), bought %d tokens at $%.2f/token",
		userID, investmentAmount, business.Name, businessID, tokensAmount, businessToken.Price)

	return &subscription, &transaction, nil
}

// Unsubscribe - пользователь выходит из инвестиции (продает токены обратно)
func (s *SubscriptionService) Unsubscribe(userID, businessID string) error {
	db := database.GetDB()

	// Проверка существования подписки
	var subscription models.BusinessSubscription
	if err := db.Where("user_id = ? AND business_id = ?", userID, businessID).First(&subscription).Error; err != nil {
		return fmt.Errorf("subscription not found: %w", err)
	}

	// Проверка существования токена бизнеса
	var businessToken models.BusinessToken
	if err := db.Where("business_id = ?", businessID).First(&businessToken).Error; err != nil {
		return fmt.Errorf("business token not found: %w", err)
	}

	// Получаем информацию о бизнесе
	var business models.Business
	if err := db.First(&business, "id = ?", businessID).Error; err != nil {
		return fmt.Errorf("business not found: %w", err)
	}

	tokensToReturn := subscription.TokensOwned
	refundAmount := businessToken.Price * float64(tokensToReturn)

	// Начало транзакции БД
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Создаем транзакцию продажи
	transaction := models.Transaction{
		ID:         uuid.New().String(),
		BusinessID: businessID,
		FromUser:   business.OwnerID, // Владелец "выкупает" токены обратно
		ToUser:     userID,
		Tokens:     tokensToReturn,
		Amount:     refundAmount,
		TxType:     "sell",
	}

	if err := tx.Create(&transaction).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create transaction: %w", err)
	}

	// Возвращаем токены в supply
	businessToken.TotalSupply += tokensToReturn
	if err := tx.Save(&businessToken).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update token supply: %w", err)
	}

	// Пересчитываем цену токена после продажи
	newPrice := s.tokenService.CalculateTokenPrice(businessID, businessToken.TotalSupply)
	businessToken.Price = newPrice
	if err := tx.Save(&businessToken).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update token price: %w", err)
	}

	// Удаляем подписку
	if err := tx.Delete(&subscription).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete subscription: %w", err)
	}

	// Коммит транзакции
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("[SUBSCRIPTION] ✅ User %s unsubscribed from business %s (%s), sold %d tokens at $%.2f/token for $%.2f refund",
		userID, business.Name, businessID, tokensToReturn, businessToken.Price, refundAmount)

	return nil
}

// GetUserSubscriptions - получить все подписки пользователя
func (s *SubscriptionService) GetUserSubscriptions(userID string) ([]models.BusinessSubscription, error) {
	db := database.GetDB()

	var subscriptions []models.BusinessSubscription
	if err := db.Where("user_id = ?", userID).Preload("Business").Find(&subscriptions).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch subscriptions: %w", err)
	}

	log.Printf("[SUBSCRIPTION] 📋 Fetched %d subscriptions for user %s", len(subscriptions), userID)

	return subscriptions, nil
}

// GetBusinessSubscribers - получить всех инвесторов бизнеса
func (s *SubscriptionService) GetBusinessSubscribers(businessID string) ([]models.BusinessSubscription, error) {
	db := database.GetDB()

	var subscriptions []models.BusinessSubscription
	if err := db.Where("business_id = ?", businessID).Preload("User").Find(&subscriptions).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch subscribers: %w", err)
	}

	log.Printf("[SUBSCRIPTION] 📋 Fetched %d subscribers for business %s", len(subscriptions), businessID)

	return subscriptions, nil
}

// GetSubscriptionStats - получить статистику подписки
func (s *SubscriptionService) GetSubscriptionStats(userID, businessID string) (*models.BusinessSubscription, error) {
	db := database.GetDB()

	var subscription models.BusinessSubscription
	if err := db.Where("user_id = ? AND business_id = ?", userID, businessID).
		Preload("Business").
		First(&subscription).Error; err != nil {
		return nil, fmt.Errorf("subscription not found: %w", err)
	}

	// Получаем текущую цену токена для расчета текущей стоимости
	var businessToken models.BusinessToken
	if err := db.Where("business_id = ?", businessID).First(&businessToken).Error; err == nil {
		currentValue := businessToken.Price * float64(subscription.TokensOwned)
		log.Printf("[SUBSCRIPTION] 📊 Subscription stats: User=%s, Business=%s, Tokens=%d, Invested=$%.2f, CurrentValue=$%.2f, Profit=$%.2f",
			userID, businessID, subscription.TokensOwned, subscription.Invested, currentValue, currentValue-subscription.Invested)
	}

	return &subscription, nil
}
