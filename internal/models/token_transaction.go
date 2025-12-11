package models

import (
	"time"
)

// TokenTransactionType типы транзакций токенов
const (
	// Treasury → User allocations
	TransactionTypeWelcomeBonus       = "WELCOME_BONUS"
	TransactionTypeQuestReward        = "QUEST_REWARD"
	TransactionTypeAchievementReward  = "ACHIEVEMENT_REWARD"
	TransactionTypeAdminAllocation    = "ADMIN_ALLOCATION"
	TransactionTypeDailyBonus         = "DAILY_BONUS"
	TransactionTypeReferralBonus      = "REFERRAL_BONUS"
	
	// User → Treasury spending
	TransactionTypeAIUsage            = "AI_USAGE"
	TransactionTypeMarketplacePurchase = "MARKETPLACE_PURCHASE"
	TransactionTypeRecipeUnlock       = "RECIPE_UNLOCK"
	TransactionTypePremiumFeature     = "PREMIUM_FEATURE"
	
	// Admin operations
	TransactionTypeAdminRevoke        = "ADMIN_REVOKE"
	TransactionTypeAdminAdjustment    = "ADMIN_ADJUSTMENT"
	
	// Special operations
	TransactionTypeTokenBurn          = "TOKEN_BURN"
	TransactionTypeUserTransfer       = "USER_TRANSFER"
)

// TokenTransaction представляет историю операций с токенами
type TokenTransaction struct {
	ID          string                 `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	FromUserID  *string                `gorm:"type:uuid" json:"from_user_id"`                    // NULL = Treasury
	ToUserID    *string                `gorm:"type:uuid" json:"to_user_id"`                      // NULL = Burn/Revoke
	Amount      int64                  `gorm:"not null" json:"amount"`
	Type        string                 `gorm:"type:varchar(50);not null" json:"type"`
	Description string                 `gorm:"type:text" json:"description"`
	Metadata    map[string]interface{} `gorm:"type:jsonb" json:"metadata,omitempty"`             // Дополнительные данные
	CreatedAt   time.Time              `gorm:"default:now()" json:"created_at"`

	// Связи (optional, для удобства запросов)
	FromUser *User `gorm:"foreignKey:FromUserID;references:ID" json:"from_user,omitempty"`
	ToUser   *User `gorm:"foreignKey:ToUserID;references:ID" json:"to_user,omitempty"`
}

// TableName указывает имя таблицы
func (TokenTransaction) TableName() string {
	return "token_transactions"
}

// IsTreasuryAllocation проверяет, является ли это выделением из Treasury
func (t *TokenTransaction) IsTreasuryAllocation() bool {
	return t.FromUserID == nil && t.ToUserID != nil
}

// IsTreasurySpending проверяет, является ли это тратой в Treasury
func (t *TokenTransaction) IsTreasurySpending() bool {
	return t.FromUserID != nil && t.ToUserID == nil
}

// IsUserTransfer проверяет, является ли это переводом между пользователями
func (t *TokenTransaction) IsUserTransfer() bool {
	return t.FromUserID != nil && t.ToUserID != nil
}
