package service

import (
	"errors"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/database"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
)

// TokenBankService интерфейс для работы с токенами
type TokenBankService interface {
	// Balance Operations
	GetUserBalance(userID string) (int64, error)
	CheckBalance(userID string, requiredAmount int64) (bool, error)
	
	// Spending Operations (tokens go back to Treasury)
	SpendTokens(userID string, amount int64) error
	SpendTokensForAIRequest(userID string, cost int64) error
	SpendTokensForMarketplace(userID string, productPrice int64) error
	SpendTokensForPremiumFeature(userID string, featureCost int64) error
	
	// Earning Operations (tokens come from Treasury)
	EarnTokens(userID string, amount int64, reason string) error
	
	// Info Operations
	GetTokenBank(userID string) (*models.TokenBank, error)
	GetAllTokenBanks() ([]models.TokenBank, error)
}

type tokenBankService struct {
	repo *database.TokenBankRepository
}

// NewTokenBankService создаёт новый экземпляр сервиса токен-банка
func NewTokenBankService() TokenBankService {
	return &tokenBankService{
		repo: &database.TokenBankRepository{},
	}
}

// ============================================
// Balance Operations
// ============================================

// GetUserBalance возвращает текущий баланс пользователя
func (s *tokenBankService) GetUserBalance(userID string) (int64, error) {
	tokenBank, err := s.repo.FindByUserID(userID)
	if err != nil {
		return 0, err
	}
	return tokenBank.Balance, nil
}

// CheckBalance проверяет, достаточно ли токенов у пользователя
func (s *tokenBankService) CheckBalance(userID string, requiredAmount int64) (bool, error) {
	return s.repo.CheckUserBalance(userID, requiredAmount)
}

// ============================================
// Spending Operations
// ============================================

// SpendTokens списывает токены у пользователя и возвращает их в Treasury
func (s *tokenBankService) SpendTokens(userID string, amount int64) error {
	return s.repo.SpendTokens(userID, amount)
}

// SpendTokensForAIRequest списывает токены за AI-запрос
// Проверяет баланс перед списанием и возвращает понятную ошибку
func (s *tokenBankService) SpendTokensForAIRequest(userID string, cost int64) error {
	if cost <= 0 {
		return errors.New("AI request cost must be positive")
	}

	// Проверяем, что токен-банк пользователя существует
	userBank, err := s.repo.FindByUserID(userID)
	if err != nil {
		if err.Error() == "token bank not found" {
			return errors.New("user token bank not initialized")
		}
		return err
	}

	// Проверяем достаточность средств
	if userBank.Balance < cost {
		return errors.New("not enough tokens to process AI request")
	}

	// Списываем токены и возвращаем их в Treasury
	return s.repo.SpendTokens(userID, cost)
}

// SpendTokensForMarketplace списывает токены за покупку в маркетплейсе
func (s *tokenBankService) SpendTokensForMarketplace(userID string, productPrice int64) error {
	if productPrice <= 0 {
		return errors.New("product price must be positive")
	}

	// Проверяем баланс
	userBank, err := s.repo.FindByUserID(userID)
	if err != nil {
		return err
	}

	if userBank.Balance < productPrice {
		return errors.New("not enough tokens to purchase this product")
	}

	// Списываем токены
	return s.repo.SpendTokens(userID, productPrice)
}

// SpendTokensForPremiumFeature списывает токены за премиум-функцию
func (s *tokenBankService) SpendTokensForPremiumFeature(userID string, featureCost int64) error {
	if featureCost <= 0 {
		return errors.New("feature cost must be positive")
	}

	// Проверяем баланс
	userBank, err := s.repo.FindByUserID(userID)
	if err != nil {
		return err
	}

	if userBank.Balance < featureCost {
		return errors.New("not enough tokens to unlock this feature")
	}

	// Списываем токены
	return s.repo.SpendTokens(userID, featureCost)
}

// ============================================
// Earning Operations
// ============================================

// EarnTokens начисляет токены пользователю из Treasury (с указанием причины)
func (s *tokenBankService) EarnTokens(userID string, amount int64, reason string) error {
	if amount <= 0 {
		return errors.New("earning amount must be positive")
	}

	// Определяем тип начисления по причине
	switch reason {
	case "quest_completion":
		return s.repo.AllocateQuestReward(userID, "manual", amount)
	case "achievement":
		return s.repo.AllocateAchievementReward(userID, "manual", amount)
	case "welcome_bonus":
		return s.repo.AllocateWelcomeBonus(userID, amount)
	default:
		// Для всех остальных случаев используем стандартное начисление из Treasury
		return s.repo.AllocateFromTreasury(userID, amount)
	}
}

// ============================================
// Info Operations
// ============================================

// GetTokenBank возвращает полную информацию о токен-банке пользователя
func (s *tokenBankService) GetTokenBank(userID string) (*models.TokenBank, error) {
	return s.repo.FindByUserID(userID)
}

// GetAllTokenBanks возвращает все токен-банки (для админа)
func (s *tokenBankService) GetAllTokenBanks() ([]models.TokenBank, error) {
	return s.repo.FindAll()
}

// ============================================
// Helper Functions
// ============================================

// CalculateAICost рассчитывает стоимость AI-запроса в зависимости от параметров
func CalculateAICost(requestLength int, complexity string) int64 {
	baseCost := int64(1)

	// Стоимость в зависимости от длины запроса
	if requestLength > 500 {
		baseCost = 5
	} else if requestLength > 200 {
		baseCost = 3
	}

	// Множитель в зависимости от сложности
	var multiplier int64 = 1
	switch complexity {
	case "basic":
		multiplier = 1
	case "pro":
		multiplier = 2
	case "advanced":
		multiplier = 3
	default:
		multiplier = 1
	}

	return baseCost * multiplier
}

// GetTokenPricing возвращает информацию о ценах на различные операции
func GetTokenPricing() map[string]int64 {
	return map[string]int64{
		"ai_request_basic":    1,
		"ai_request_pro":      3,
		"ai_request_advanced": 5,
		"marketplace_unlock":  10,
		"premium_recipe":      25,
		"cooking_assistant":   2,
		"nutrition_analysis":  3,
	}
}
