package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/services"
	"github.com/gorilla/mux"
)

var tokenService = services.NewTokenService()

// CreateTokenRequest запрос на создание нового токена
type CreateTokenRequest struct {
	Symbol       string  `json:"symbol"`
	InitialPrice float64 `json:"initialPrice"`
	TotalSupply  int64   `json:"totalSupply"`
}

// MintTokensRequest запрос на создание токенов
type MintTokensRequest struct {
	Amount int64  `json:"amount"`
	Reason string `json:"reason"`
}

// BurnTokensRequest запрос на сжигание токенов
type BurnTokensRequest struct {
	Amount int64  `json:"amount"`
	Reason string `json:"reason"`
}

// 🆕 POST /api/businesses/{id}/tokens
func CreateBusinessToken(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	businessID := vars["id"]

	var input CreateTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Валидация
	if input.Symbol == "" {
		http.Error(w, "Symbol is required", http.StatusBadRequest)
		return
	}

	if input.TotalSupply <= 0 {
		http.Error(w, "Total supply must be positive", http.StatusBadRequest)
		return
	}

	if input.InitialPrice <= 0 {
		input.InitialPrice = 19.0 // Дефолтная цена
	}

	// Создание токена через сервис
	token, err := tokenService.CreateToken(businessID, input.Symbol, input.InitialPrice, input.TotalSupply)
	if err != nil {
		log.Printf("[TOKEN] ❌ Error creating token: %v", err)
		http.Error(w, "Failed to create token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("[TOKEN] 🆕 Created token for business %s: Symbol=%s, Supply=%d, Price=$%.2f",
		businessID, input.Symbol, input.TotalSupply, input.InitialPrice)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "✅ Token created successfully",
		"token":   token,
	})
}

// 🪙 POST /api/businesses/{id}/tokens/mint
func MintBusinessTokens(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	businessID := vars["id"]

	var input MintTokensRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Валидация
	if input.Amount <= 0 {
		http.Error(w, "Amount must be positive", http.StatusBadRequest)
		return
	}

	if input.Reason == "" {
		input.Reason = "Manual mint"
	}

	// Mint токенов
	token, err := tokenService.MintTokens(businessID, input.Amount, input.Reason)
	if err != nil {
		log.Printf("[TOKEN] ❌ Error minting tokens: %v", err)
		http.Error(w, "Failed to mint tokens: "+err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("[TOKEN] ✅ Minted %d tokens for business %s", input.Amount, businessID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "✅ Tokens minted successfully",
		"token":   token,
		"minted":  input.Amount,
	})
}

// 🔥 POST /api/businesses/{id}/tokens/burn
func BurnBusinessTokens(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	businessID := vars["id"]

	var input BurnTokensRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Валидация
	if input.Amount <= 0 {
		http.Error(w, "Amount must be positive", http.StatusBadRequest)
		return
	}

	if input.Reason == "" {
		input.Reason = "Manual burn"
	}

	// Burn токенов
	token, err := tokenService.BurnTokens(businessID, input.Amount, input.Reason)
	if err != nil {
		log.Printf("[TOKEN] ❌ Error burning tokens: %v", err)
		http.Error(w, "Failed to burn tokens: "+err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("[TOKEN] 🔥 Burned %d tokens for business %s", input.Amount, businessID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "🔥 Tokens burned successfully",
		"token":   token,
		"burned":  input.Amount,
	})
}

// 📊 GET /api/businesses/{id}/tokens
func GetBusinessTokens(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	businessID := vars["id"]

	// Получение токена
	token, err := tokenService.GetBusinessToken(businessID)
	if err != nil {
		log.Printf("[TOKEN] ❌ Token not found for business %s: %v", businessID, err)
		http.Error(w, "Token not found", http.StatusNotFound)
		return
	}

	log.Printf("[TOKEN] ✅ Retrieved token for business %s: Supply=%d, Price=$%.2f",
		businessID, token.TotalSupply, token.Price)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(token)
}

// 💰 POST /api/businesses/{id}/tokens/recalculate-price
func RecalculateTokenPrice(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	businessID := vars["id"]

	// Пересчет цены
	token, err := tokenService.RecalculatePrice(businessID)
	if err != nil {
		log.Printf("[TOKEN] ❌ Error recalculating price: %v", err)
		http.Error(w, "Failed to recalculate price: "+err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("[TOKEN] 💰 Price recalculated for business %s: $%.2f", businessID, token.Price)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "💰 Price recalculated successfully",
		"token":   token,
	})
}
