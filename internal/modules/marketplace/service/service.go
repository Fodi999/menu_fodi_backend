package service

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/marketplace/dto"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/marketplace/repo"
)

const (
	PlatformCommissionRate = 0.10 // 10% platform commission
)

// MarketplaceService handles marketplace business logic
type MarketplaceService interface {
	GetMarketRecipes(filters dto.MarketplaceFilters) (*dto.MarketplaceResponse, error)
	PurchaseRecipe(req dto.PurchaseRequest) (*dto.PurchaseResponse, error)
	GetUserPurchases(userID uuid.UUID) ([]dto.UserPurchase, error)
	GetSellerStats(sellerID uuid.UUID) (*dto.SellerStats, error)
	GetLeaderboard(sortBy, language string, limit int) (*dto.LeaderboardResponse, error)
}

type marketplaceService struct {
	repo repo.MarketplaceRepository
}

// NewMarketplaceService creates new marketplace service
func NewMarketplaceService(repo repo.MarketplaceRepository) MarketplaceService {
	return &marketplaceService{repo: repo}
}

func (s *marketplaceService) GetMarketRecipes(filters dto.MarketplaceFilters) (*dto.MarketplaceResponse, error) {
	recipes, err := s.repo.GetMarketRecipes(filters)
	if err != nil {
		return nil, err
	}

	return &dto.MarketplaceResponse{
		Recipes: recipes,
		Total:   len(recipes),
	}, nil
}

func (s *marketplaceService) PurchaseRecipe(req dto.PurchaseRequest) (*dto.PurchaseResponse, error) {
	recipeID, err := uuid.Parse(req.RecipeID)
	if err != nil {
		return nil, fmt.Errorf("invalid recipe ID")
	}

	buyerID, err := uuid.Parse(req.BuyerID)
	if err != nil {
		return nil, fmt.Errorf("invalid buyer ID")
	}

	// Get recipe
	recipe, err := s.repo.GetRecipeByID(recipeID)
	if err != nil {
		return nil, err
	}

	// Check: cannot buy own recipe
	if recipe.UserID == buyerID {
		return nil, repo.ErrCannotBuyOwnRecipe
	}

	// Check: already purchased?
	exists, err := s.repo.CheckPurchaseExists(recipeID, buyerID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, repo.ErrAlreadyPurchased
	}

	// Get buyer and seller profiles
	buyerProfile, err := s.repo.GetUserProfile(buyerID)
	if err != nil {
		return nil, fmt.Errorf("buyer profile not found")
	}

	_, err = s.repo.GetUserProfile(recipe.UserID)
	if err != nil {
		return nil, fmt.Errorf("seller profile not found")
	}

	// Check balance
	if buyerProfile.WalletBalance < recipe.Price {
		return nil, repo.ErrInsufficientFunds
	}

	// Calculate amounts
	commission := recipe.Price * PlatformCommissionRate
	netAmount := recipe.Price - commission

	// Create purchase record
	purchase := &models.RecipePurchase{
		RecipeID:   req.RecipeID,
		BuyerID:    req.BuyerID,
		SellerID:   recipe.UserID.String(),
		Price:      recipe.Price,
		Commission: commission,
		NetAmount:  netAmount,
	}

	if err := s.repo.CreatePurchase(purchase); err != nil {
		return nil, fmt.Errorf("failed to create purchase: %w", err)
	}

	// Update wallet balances
	if err := s.repo.UpdateWalletBalance(buyerID, -recipe.Price); err != nil {
		return nil, fmt.Errorf("failed to deduct from buyer: %w", err)
	}

	if err := s.repo.UpdateWalletBalance(recipe.UserID, netAmount); err != nil {
		return nil, fmt.Errorf("failed to credit seller: %w", err)
	}

	// Increment recipe purchases counter
	if err := s.repo.IncrementPurchases(recipeID); err != nil {
		return nil, fmt.Errorf("failed to update purchases: %w", err)
	}

	// Create wallet transactions
	purchaseUUID, _ := uuid.Parse(purchase.ID)

	buyerTx := &models.WalletTransaction{
		UserID:      buyerID,
		Amount:      -recipe.Price,
		Type:        "purchase",
		Description: fmt.Sprintf("Recipe purchase: %s", recipe.Title),
		RelatedID:   purchaseUUID,
	}
	s.repo.CreateWalletTransaction(buyerTx)

	sellerTx := &models.WalletTransaction{
		UserID:      recipe.UserID,
		Amount:      netAmount,
		Type:        "sale",
		Description: fmt.Sprintf("Recipe sale: %s", recipe.Title),
		RelatedID:   purchaseUUID,
	}
	s.repo.CreateWalletTransaction(sellerTx)

	// Return response with updated buyer balance
	buyerProfile.WalletBalance -= recipe.Price

	return &dto.PurchaseResponse{
		PurchaseID:     purchase.ID,
		Recipe:         recipe.Title,
		Price:          recipe.Price,
		Commission:     commission,
		SellerReceived: netAmount,
		BuyerBalance:   buyerProfile.WalletBalance,
	}, nil
}

func (s *marketplaceService) GetUserPurchases(userID uuid.UUID) ([]dto.UserPurchase, error) {
	return s.repo.GetUserPurchases(userID)
}

func (s *marketplaceService) GetSellerStats(sellerID uuid.UUID) (*dto.SellerStats, error) {
	return s.repo.GetSellerStats(sellerID)
}

func (s *marketplaceService) GetLeaderboard(sortBy, language string, limit int) (*dto.LeaderboardResponse, error) {
	// Default sort by XP
	if sortBy == "" {
		sortBy = "xp"
	}

	// Default limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	leaders, err := s.repo.GetLeaderboard(sortBy, language, limit)
	if err != nil {
		return nil, err
	}

	return &dto.LeaderboardResponse{
		Leaders: leaders,
		Total:   len(leaders),
		SortBy:  sortBy,
	}, nil
}
