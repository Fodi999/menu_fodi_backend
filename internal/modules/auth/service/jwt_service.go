package service

import (
	"errors"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims структура для JWT токена
type Claims struct {
	Email   string `json:"email"`
	Role    string `json:"role"`
	HasRole bool   `json:"hasRole"`
	jwt.RegisteredClaims // Contains Subject (sub), ExpiresAt (exp), IssuedAt (iat)
}

// JWTService handles JWT token generation and validation
type JWTService struct{}

// NewJWTService creates a new JWT service
func NewJWTService() *JWTService {
	// Log JWT_SECRET configuration at startup
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		log.Printf("⚠️  JWT_SECRET not set, using fallback secret (INSECURE!)")
	} else {
		log.Printf("✅ JWT_SECRET loaded from environment (length: %d)", len(secret))
	}
	return &JWTService{}
}

// GenerateToken генерирует JWT токен для пользователя
func (s *JWTService) GenerateToken(userID, email, role string) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "your-secret-key-change-this-in-production"
		log.Printf("⚠️  GenerateToken: Using fallback secret")
	} else {
		log.Printf("🔑 GenerateToken: Using JWT_SECRET from env (len=%d)", len(secret))
	}

	claims := &Claims{
		Email:   email,
		Role:    role,
		HasRole: true,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID, // 🔴 RFC 7519: sub field для user ID
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ValidateToken проверяет и парсит JWT токен
func (s *JWTService) ValidateToken(tokenString string) (*Claims, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "your-secret-key-change-this-in-production"
		log.Printf("⚠️  ValidateToken: Using fallback secret")
	} else {
		log.Printf("🔑 ValidateToken: Using JWT_SECRET from env (len=%d)", len(secret))
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	if err != nil {
		log.Printf("❌ JWT parse error: %v", err)
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// Legacy helper functions for backward compatibility
// GenerateToken генерирует JWT токен для пользователя
func GenerateToken(userID, email, role string) (string, error) {
	svc := NewJWTService()
	return svc.GenerateToken(userID, email, role)
}

// ValidateToken проверяет и парсит JWT токен
func ValidateToken(tokenString string) (*Claims, error) {
	svc := NewJWTService()
	return svc.ValidateToken(tokenString)
}
