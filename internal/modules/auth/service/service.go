package service

import (
	"errors"
	"log"
	"time"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/database"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/auth/dto"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/auth/repo"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserExists         = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidToken       = errors.New("invalid token")
	ErrAccountBlocked     = errors.New("account is blocked")
	ErrWeakPassword       = errors.New("password must be at least 6 characters")
)

// AuthService handles authentication business logic
type AuthService struct {
	repo repo.AuthRepository
}

// NewAuthService creates a new auth service
func NewAuthService(repository repo.AuthRepository) *AuthService {
	return &AuthService{
		repo: repository,
	}
}

// Register registers a new user
func (s *AuthService) Register(req dto.RegisterRequest) (*dto.AuthResponse, error) {
	// Check if user exists
	existingUser, _ := s.repo.FindByEmail(req.Email)
	if existingUser != nil {
		return nil, ErrUserExists
	}

	// Validate password strength
	if len(req.Password) < 6 {
		return nil, ErrWeakPassword
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Create user with home_chef role (MVP default)
	user := &models.User{
		ID:        uuid.New().String(),
		Email:     req.Email,
		Name:      req.Name,
		Password:  string(hashedPassword),
		Role:      models.RoleHomeChef, // Default role for new users
		CreatedAt: time.Now(),
	}

	// Save to database
	if err := s.repo.Create(user); err != nil {
		return nil, err
	}

	// Initialize token bank for new user with 100 tokens welcome bonus
	tokenRepo := &database.TokenBankRepository{}
	if err := tokenRepo.InitializeTokenBankForUser(user.ID); err != nil {
		// Log error but don't block registration
		// User can still use the app, tokens can be allocated manually later
		log.Printf("Warning: Failed to initialize token bank for user %s: %v", user.ID, err)
	}

	// Generate JWT token
	token, err := GenerateToken(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, err
	}

	return &dto.AuthResponse{
		Success: true,
		Token:   token,
		User:    dto.ToUserInfo(user),
		Message: "Registration successful",
	}, nil
}

// Login authenticates a user
func (s *AuthService) Login(req dto.LoginRequest) (*dto.AuthResponse, error) {
	// Find user by email
	user, err := s.repo.FindByEmail(req.Email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	// Check if user is blocked
	if user.Status == models.UserStatusBlocked {
		return nil, ErrAccountBlocked
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	// Update last_login timestamp
	now := time.Now()
	user.LastLogin = &now
	if err := s.repo.UpdateLastLogin(user.ID, now); err != nil {
		// Log error but don't fail login
		log.Printf("Warning: Failed to update last_login for user %s: %v\n", user.ID, err)
	}

	// Generate JWT token
	token, err := GenerateToken(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, err
	}

	return &dto.AuthResponse{
		Success: true,
		Token:   token,
		User:    dto.ToUserInfo(user),
		Message: "Login successful",
	}, nil
}

// VerifyToken verifies JWT token
func (s *AuthService) VerifyToken(req dto.VerifyTokenRequest) (*dto.VerifyTokenResponse, error) {
	// Validate token
	claims, err := ValidateToken(req.Token)
	if err != nil {
		return &dto.VerifyTokenResponse{Valid: false}, nil
	}

	// Get user to verify it still exists
	user, err := s.repo.FindByID(claims.Subject)
	if err != nil {
		return &dto.VerifyTokenResponse{Valid: false}, nil
	}

	return &dto.VerifyTokenResponse{
		Valid:  true,
		UserID: user.ID,
		Role:   user.Role,
		Name:   user.Name,
		Email:  user.Email,
	}, nil
}

// GetCurrentUser retrieves current authenticated user with profile
func (s *AuthService) GetCurrentUser(userID uuid.UUID) (*dto.CurrentUserResponse, error) {
	// Fetch user
	user, err := s.repo.FindByID(userID.String())
	if err != nil {
		return nil, ErrUserNotFound
	}

	// Fetch user profile
	profile, err := s.repo.GetUserProfile(userID)

	response := &dto.CurrentUserResponse{
		ID:            user.ID,
		Email:         user.Email,
		Name:          user.Name,
		Role:          user.Role,
		Status:        user.Status,
		CreatedAt:     user.CreatedAt,
		WalletBalance: 0,
	}

	// If profile exists, add profile data
	if profile != nil {
		response.WalletBalance = int(profile.WalletBalance)
		response.Profile = &dto.UserProfileInfo{
			Level:            profile.Level,
			Stars:            profile.Stars,
			XP:               profile.XP,
			AvatarURL:        profile.AvatarURL,
			Language:         profile.Language,
			CompletedCourses: profile.CompletedCourses,
		}
	}

	return response, nil
}
