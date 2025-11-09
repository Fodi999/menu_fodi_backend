package repo

import (
	"time"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/database"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// WalletTransaction represents a wallet transaction record
type WalletTransaction struct {
	ID            uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID        uuid.UUID `gorm:"type:uuid;not null" json:"userId"`
	Amount        int       `gorm:"not null" json:"amount"`
	Type          string    `gorm:"type:varchar(50);not null" json:"type"` // purchase, spend, grant
	Status        string    `gorm:"type:varchar(50);not null" json:"status"`
	PaymentMethod string    `gorm:"type:varchar(50)" json:"paymentMethod,omitempty"`
	RelatedID     string    `gorm:"type:varchar(255)" json:"relatedId,omitempty"`
	Description   string    `gorm:"type:text" json:"description"`
	CreatedAt     time.Time `gorm:"autoCreateTime" json:"createdAt"`
}

// TableName specifies the table name
func (WalletTransaction) TableName() string {
	return "WalletTransaction"
}

// WalletRepository handles wallet database operations
type WalletRepository interface {
	GetBalance(userID uuid.UUID) (int, error)
	UpdateBalance(userID uuid.UUID, newBalance int) error
	CreateTransaction(transaction WalletTransaction) error
	GetTransactions(userID uuid.UUID, limit int) ([]WalletTransaction, error)
}

// walletRepository implements WalletRepository
type walletRepository struct {
	db *gorm.DB
}

// NewWalletRepository creates a new wallet repository
func NewWalletRepository() WalletRepository {
	return &walletRepository{
		db: database.DB,
	}
}

// GetBalance retrieves user's wallet balance from UserProfile
func (r *walletRepository) GetBalance(userID uuid.UUID) (int, error) {
	var profile struct {
		WalletBalance int
	}

	err := r.db.Table("\"UserProfile\"").
		Select("wallet_balance").
		Where("user_id = ?", userID).
		First(&profile).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, nil
		}
		return 0, err
	}

	return profile.WalletBalance, nil
}

// UpdateBalance updates user's wallet balance
func (r *walletRepository) UpdateBalance(userID uuid.UUID, newBalance int) error {
	return r.db.Table("\"UserProfile\"").
		Where("user_id = ?", userID).
		Update("wallet_balance", newBalance).Error
}

// CreateTransaction creates a new wallet transaction record
func (r *walletRepository) CreateTransaction(transaction WalletTransaction) error {
	return r.db.Create(&transaction).Error
}

// GetTransactions retrieves user's transaction history
func (r *walletRepository) GetTransactions(userID uuid.UUID, limit int) ([]WalletTransaction, error) {
	var transactions []WalletTransaction

	err := r.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Find(&transactions).Error

	return transactions, err
}
