package service

import (
	"errors"
	"time"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/wallet/dto"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/wallet/repo"
	"github.com/google/uuid"
)

var (
	ErrInsufficientBalance = errors.New("insufficient balance")
	ErrInvalidAmount       = errors.New("invalid amount")
	ErrUserNotFound        = errors.New("user not found")
)

// WalletService handles wallet business logic
type WalletService struct {
	repo repo.WalletRepository
}

// NewWalletService creates a new wallet service
func NewWalletService(repository repo.WalletRepository) *WalletService {
	return &WalletService{
		repo: repository,
	}
}

// GetBalance retrieves user's wallet balance
func (s *WalletService) GetBalance(userID uuid.UUID) (*dto.BalanceResponse, error) {
	balance, err := s.repo.GetBalance(userID)
	if err != nil {
		return nil, err
	}

	return &dto.BalanceResponse{
		UserID:  userID,
		Balance: balance,
	}, nil
}

// PurchaseTokens adds tokens to user's wallet
func (s *WalletService) PurchaseTokens(userID uuid.UUID, req dto.PurchaseRequest) (*dto.TransactionResponse, error) {
	// Validate amount
	if req.Amount <= 0 {
		return nil, ErrInvalidAmount
	}

	// Get current balance
	currentBalance, err := s.repo.GetBalance(userID)
	if err != nil {
		return nil, err
	}

	// Calculate new balance
	newBalance := currentBalance + req.Amount

	// Update balance
	if err := s.repo.UpdateBalance(userID, newBalance); err != nil {
		return nil, err
	}

	// Create transaction record
	transaction := repo.WalletTransaction{
		ID:            uuid.New(),
		UserID:        userID,
		Amount:        req.Amount,
		Type:          "purchase",
		Status:        "completed",
		PaymentMethod: req.PaymentMethod,
		Description:   "Token purchase",
		CreatedAt:     time.Now(),
	}

	if err := s.repo.CreateTransaction(transaction); err != nil {
		// Rollback balance update
		s.repo.UpdateBalance(userID, currentBalance)
		return nil, err
	}

	return &dto.TransactionResponse{
		ID:              transaction.ID,
		UserID:          userID,
		Amount:          req.Amount,
		Type:            transaction.Type,
		Status:          transaction.Status,
		NewBalance:      newBalance,
		PreviousBalance: currentBalance,
		CreatedAt:       transaction.CreatedAt,
	}, nil
}

// SpendTokens deducts tokens from user's wallet
func (s *WalletService) SpendTokens(userID uuid.UUID, req dto.SpendRequest) (*dto.TransactionResponse, error) {
	// Validate amount
	if req.Amount <= 0 {
		return nil, ErrInvalidAmount
	}

	// Get current balance
	currentBalance, err := s.repo.GetBalance(userID)
	if err != nil {
		return nil, err
	}

	// Check sufficient balance
	if currentBalance < req.Amount {
		return nil, ErrInsufficientBalance
	}

	// Calculate new balance
	newBalance := currentBalance - req.Amount

	// Update balance
	if err := s.repo.UpdateBalance(userID, newBalance); err != nil {
		return nil, err
	}

	// Create transaction record
	transaction := repo.WalletTransaction{
		ID:          uuid.New(),
		UserID:      userID,
		Amount:      -req.Amount, // Negative for spending
		Type:        "spend",
		Status:      "completed",
		RelatedID:   req.RelatedID,
		Description: req.Description,
		CreatedAt:   time.Now(),
	}

	if err := s.repo.CreateTransaction(transaction); err != nil {
		// Rollback balance update
		s.repo.UpdateBalance(userID, currentBalance)
		return nil, err
	}

	return &dto.TransactionResponse{
		ID:              transaction.ID,
		UserID:          userID,
		Amount:          req.Amount,
		Type:            transaction.Type,
		Status:          transaction.Status,
		NewBalance:      newBalance,
		PreviousBalance: currentBalance,
		CreatedAt:       transaction.CreatedAt,
	}, nil
}

// GetTransactionHistory retrieves user's transaction history
func (s *WalletService) GetTransactionHistory(userID uuid.UUID, limit int) ([]dto.TransactionResponse, error) {
	if limit <= 0 {
		limit = 50
	}

	transactions, err := s.repo.GetTransactions(userID, limit)
	if err != nil {
		return nil, err
	}

	// Convert to response DTOs
	result := make([]dto.TransactionResponse, len(transactions))
	for i, tx := range transactions {
		result[i] = dto.TransactionResponse{
			ID:        tx.ID,
			UserID:    tx.UserID,
			Amount:    tx.Amount,
			Type:      tx.Type,
			Status:    tx.Status,
			CreatedAt: tx.CreatedAt,
		}
	}

	return result, nil
}

// GrantWelcomeTokens grants initial tokens to new users
func (s *WalletService) GrantWelcomeTokens(userID uuid.UUID, amount int) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}

	// Get current balance
	currentBalance, err := s.repo.GetBalance(userID)
	if err != nil {
		return err
	}

	// Update balance
	newBalance := currentBalance + amount
	if err := s.repo.UpdateBalance(userID, newBalance); err != nil {
		return err
	}

	// Create transaction record
	transaction := repo.WalletTransaction{
		ID:          uuid.New(),
		UserID:      userID,
		Amount:      amount,
		Type:        "grant",
		Status:      "completed",
		Description: "Welcome bonus",
		CreatedAt:   time.Now(),
	}

	return s.repo.CreateTransaction(transaction)
}
