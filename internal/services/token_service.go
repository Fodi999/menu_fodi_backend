package services

import (
	"fmt"
	"log"
	"math"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/database"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
	"github.com/google/uuid"
)

// TokenService - сервис для работы с токенами бизнеса
type TokenService struct{}

// NewTokenService создает новый экземпляр TokenService
func NewTokenService() *TokenService {
	return &TokenService{}
}

// MintInitialToken создает первоначальный токен при регистрации бизнеса
func (s *TokenService) MintInitialToken(businessID string) (*models.BusinessToken, error) {
	db := database.GetDB()

	// Проверка существования бизнеса
	var business models.Business
	if err := db.First(&business, "id = ?", businessID).Error; err != nil {
		return nil, fmt.Errorf("business not found: %w", err)
	}

	// Проверка, не создан ли уже токен
	var existingToken models.BusinessToken
	if err := db.Where("business_id = ?", businessID).First(&existingToken).Error; err == nil {
		return &existingToken, nil // Токен уже существует
	}

	// Создание символа токена (3 первые буквы названия + рандом)
	symbol := s.generateTokenSymbol(business.Name)

	// Создание первоначального токена
	token := models.BusinessToken{
		ID:          uuid.New().String(),
		BusinessID:  businessID,
		Symbol:      symbol,
		TotalSupply: 1,
		Price:       19.0, // Начальная цена $19
	}

	if err := db.Create(&token).Error; err != nil {
		return nil, fmt.Errorf("failed to create token: %w", err)
	}

	log.Printf("[TOKEN] ✅ Initial token minted: Business=%s, Symbol=%s, Supply=1, Price=$19", businessID, symbol)
	return &token, nil
}

// CreateToken создает новый токен вручную с заданными параметрами
func (s *TokenService) CreateToken(businessID, symbol string, initialPrice float64, totalSupply int64) (*models.BusinessToken, error) {
	db := database.GetDB()

	// Проверка существования бизнеса
	var business models.Business
	if err := db.First(&business, "id = ?", businessID).Error; err != nil {
		return nil, fmt.Errorf("business not found: %w", err)
	}

	// Проверка, не создан ли уже токен
	var existingToken models.BusinessToken
	if err := db.Where("business_id = ?", businessID).First(&existingToken).Error; err == nil {
		return nil, fmt.Errorf("token already exists for this business")
	}

	// Создание нового токена
	token := models.BusinessToken{
		ID:          uuid.New().String(),
		BusinessID:  businessID,
		Symbol:      symbol,
		TotalSupply: totalSupply,
		Price:       initialPrice,
	}

	if err := db.Create(&token).Error; err != nil {
		return nil, fmt.Errorf("failed to create token: %w", err)
	}

	// Загружаем Business для возврата полных данных
	if err := db.Preload("Business").First(&token, "id = ?", token.ID).Error; err != nil {
		log.Printf("[TOKEN] ⚠️ Token created but failed to load business: %v", err)
	}

	log.Printf("[TOKEN] 🆕 Manual token created: Business=%s, Symbol=%s, Supply=%d, Price=$%.2f",
		businessID, symbol, totalSupply, initialPrice)
	return &token, nil
}

// MintTokens создает новые токены (увеличение supply)
func (s *TokenService) MintTokens(businessID string, amount int64, reason string) (*models.BusinessToken, error) {
	db := database.GetDB()

	if amount <= 0 {
		return nil, fmt.Errorf("amount must be positive")
	}

	// Получение текущего токена
	var token models.BusinessToken
	if err := db.Where("business_id = ?", businessID).First(&token).Error; err != nil {
		return nil, fmt.Errorf("token not found for business: %w", err)
	}

	// Обновление supply
	oldSupply := token.TotalSupply
	token.TotalSupply += amount

	// Пересчет цены на основе активности бизнеса
	newPrice := s.calculateTokenPrice(businessID, token.TotalSupply)
	token.Price = newPrice

	// Сохранение изменений
	if err := db.Save(&token).Error; err != nil {
		return nil, fmt.Errorf("failed to update token: %w", err)
	}

	log.Printf("[TOKEN] ✅ Minted: Business=%s, Amount=%d, Supply: %d→%d, Price: $%.2f→$%.2f, Reason=%s",
		businessID, amount, oldSupply, token.TotalSupply, 19.0, newPrice, reason)

	return &token, nil
}

// BurnTokens уничтожает токены (уменьшение supply)
func (s *TokenService) BurnTokens(businessID string, amount int64, reason string) (*models.BusinessToken, error) {
	db := database.GetDB()

	if amount <= 0 {
		return nil, fmt.Errorf("amount must be positive")
	}

	// Получение текущего токена
	var token models.BusinessToken
	if err := db.Where("business_id = ?", businessID).First(&token).Error; err != nil {
		return nil, fmt.Errorf("token not found for business: %w", err)
	}

	// Проверка достаточности токенов
	if token.TotalSupply < amount {
		return nil, fmt.Errorf("insufficient token supply: have %d, want to burn %d", token.TotalSupply, amount)
	}

	// Обновление supply
	oldSupply := token.TotalSupply
	token.TotalSupply -= amount

	// Пересчет цены
	newPrice := s.calculateTokenPrice(businessID, token.TotalSupply)
	token.Price = newPrice

	// Сохранение изменений
	if err := db.Save(&token).Error; err != nil {
		return nil, fmt.Errorf("failed to update token: %w", err)
	}

	log.Printf("[TOKEN] 🔥 Burned: Business=%s, Amount=%d, Supply: %d→%d, Price: $%.2f→$%.2f, Reason=%s",
		businessID, amount, oldSupply, token.TotalSupply, 19.0, newPrice, reason)

	return &token, nil
}

// GetBusinessToken получает информацию о токене бизнеса
func (s *TokenService) GetBusinessToken(businessID string) (*models.BusinessToken, error) {
	db := database.GetDB()

	var token models.BusinessToken
	if err := db.Where("business_id = ?", businessID).Preload("Business").First(&token).Error; err != nil {
		return nil, fmt.Errorf("token not found: %w", err)
	}

	return &token, nil
}

// calculateTokenPrice рассчитывает цену токена на основе активности бизнеса
func (s *TokenService) calculateTokenPrice(businessID string, supply int64) float64 {
	db := database.GetDB()

	// Базовая цена
	basePrice := 19.0

	// 1. Получение количества подписок (инвесторов)
	var subscriptionCount int64
	db.Model(&models.BusinessSubscription{}).Where("business_id = ?", businessID).Count(&subscriptionCount)

	// 2. Получение общего объема инвестиций
	var totalInvested float64
	db.Model(&models.BusinessSubscription{}).
		Where("business_id = ?", businessID).
		Select("COALESCE(SUM(invested), 0)").
		Scan(&totalInvested)

	// 3. Получение количества транзакций
	var transactionCount int64
	db.Model(&models.Transaction{}).Where("business_id = ?", businessID).Count(&transactionCount)

	// Формула расчета цены:
	// Price = BasePrice * (1 + supply_multiplier + investor_multiplier + transaction_multiplier)
	
	// Множитель от supply (каждые 10 токенов добавляют 5%)
	supplyMultiplier := float64(supply) / 10.0 * 0.05

	// Множитель от инвесторов (каждый инвестор добавляет 2%)
	investorMultiplier := float64(subscriptionCount) * 0.02

	// Множитель от объема инвестиций (каждые $100 добавляют 1%)
	investmentMultiplier := (totalInvested / 100.0) * 0.01

	// Множитель от транзакций (каждые 5 транзакций добавляют 1%)
	transactionMultiplier := float64(transactionCount) / 5.0 * 0.01

	// Итоговая цена
	totalMultiplier := 1.0 + supplyMultiplier + investorMultiplier + investmentMultiplier + transactionMultiplier
	newPrice := basePrice * totalMultiplier

	// Ограничение роста цены (максимум 10x от базовой)
	maxPrice := basePrice * 10.0
	if newPrice > maxPrice {
		newPrice = maxPrice
	}

	// Округление до 2 знаков
	newPrice = math.Round(newPrice*100) / 100

	log.Printf("[TOKEN] 📊 Price calculation: Business=%s, Supply=%d, Investors=%d, Invested=$%.2f, Txs=%d → Price=$%.2f (multiplier=%.2f)",
		businessID, supply, subscriptionCount, totalInvested, transactionCount, newPrice, totalMultiplier)

	return newPrice
}

// generateTokenSymbol генерирует символ токена из названия бизнеса
func (s *TokenService) generateTokenSymbol(businessName string) string {
	// Берем первые 3 символа (или меньше) и переводим в верхний регистр
	runes := []rune(businessName)
	symbolLength := 3
	if len(runes) < symbolLength {
		symbolLength = len(runes)
	}

	symbol := string(runes[:symbolLength])
	
	// Добавляем "T" (Token) в конец
	return symbol + "T"
}

// RecalculatePrice пересчитывает цену токена на основе текущей активности
func (s *TokenService) RecalculatePrice(businessID string) (*models.BusinessToken, error) {
	db := database.GetDB()

	var token models.BusinessToken
	if err := db.Where("business_id = ?", businessID).First(&token).Error; err != nil {
		return nil, fmt.Errorf("token not found: %w", err)
	}

	oldPrice := token.Price
	newPrice := s.calculateTokenPrice(businessID, token.TotalSupply)
	token.Price = newPrice

	if err := db.Save(&token).Error; err != nil {
		return nil, fmt.Errorf("failed to update price: %w", err)
	}

	log.Printf("[TOKEN] 💰 Price recalculated: Business=%s, $%.2f → $%.2f", businessID, oldPrice, newPrice)
	return &token, nil
}
