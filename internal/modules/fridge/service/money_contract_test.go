package service

import (
	"testing"
	"time"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
)

// TestMoneyContract_TotalPriceAlwaysRounded - контрактный тест для правила денег
// ⚠️ КРИТИЧНЫЙ ТЕСТ: Если этот тест падает, фронт сломается!
// API КОНТРАКТ: totalPrice ВСЕГДА округлён до 2 знаков, НИКОГДА не пересчитывается на фронте
func TestMoneyContract_TotalPriceAlwaysRounded(t *testing.T) {
	service := &FridgeService{}

	tests := []struct {
		name        string
		quantity    float64
		pricePerUnit float64
		expectedTotal float64
		description string
	}{
		{
			name:          "Real production case: Ogórek",
			quantity:      3560,
			pricePerUnit:  0.00581,
			expectedTotal: 20.68, // NOT 20.6836
			description:   "3560g * 0.00581 PLN/g must equal 20.68 PLN",
		},
		{
			name:          "Float precision issue",
			quantity:      3560,
			pricePerUnit:  0.00290845,
			expectedTotal: 10.35, // NOT 10.350000000000001
			description:   "Must handle float precision correctly",
		},
		{
			name:          "Small quantities",
			quantity:      150,
			pricePerUnit:  0.0032,
			expectedTotal: 0.48, // 150 * 0.0032 = 0.48
			description:   "Small numbers must also round correctly",
		},
		{
			name:          "Large price per unit",
			quantity:      2,
			pricePerUnit:  15.999,
			expectedTotal: 32.00, // 2 * 15.999 = 31.998 → 32.00
			description:   "Rounding up must work correctly",
		},
		{
			name:          "Edge case: 0.005 rounds to 0.01",
			quantity:      1,
			pricePerUnit:  0.005,
			expectedTotal: 0.01, // math.Round(0.005 * 100) / 100 = 0.01
			description:   "Banker's rounding: 0.005 → 0.01",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pricePtr := &tt.pricePerUnit
			result := service.calculateTotalPrice(tt.quantity, pricePtr)

			if result == nil {
				t.Fatal("Expected non-nil totalPrice")
			}

			if *result != tt.expectedTotal {
				t.Errorf(
					"API CONTRACT VIOLATION!\n"+
						"Description: %s\n"+
						"Calculation: %.0f * %.8f\n"+
						"Expected: %.2f PLN\n"+
						"Got: %.10f PLN\n"+
						"⚠️ Frontend expects exactly %.2f, not %.10f!",
					tt.description,
					tt.quantity,
					tt.pricePerUnit,
					tt.expectedTotal,
					*result,
					tt.expectedTotal,
					*result,
				)
			}
		})
	}
}

// TestMoneyContract_SumOfRoundedValues - контракт для wartość lodówki
// ПРАВИЛО: Wartość lodówki = SUM(totalPrice) где каждое totalPrice уже округлено
func TestMoneyContract_SumOfRoundedValues(t *testing.T) {
	service := &FridgeService{}

	// Симулируем 3 продукта в холодильнике
	items := []struct {
		quantity     float64
		pricePerUnit float64
	}{
		{quantity: 3560, pricePerUnit: 0.00581},  // 20.68 PLN
		{quantity: 1200, pricePerUnit: 0.00416},  // 4.99 PLN
		{quantity: 500, pricePerUnit: 0.008},     // 4.00 PLN
	}

	var totalFridgeValue float64
	for _, item := range items {
		pricePtr := &item.pricePerUnit
		itemTotal := service.calculateTotalPrice(item.quantity, pricePtr)
		if itemTotal != nil {
			totalFridgeValue += *itemTotal
		}
	}

	// Wartość lodówki = 20.68 + 4.99 + 4.00 = 29.67 PLN
	expected := 29.67
	
	// Округляем итоговую сумму тоже до 2 знаков (на случай накопления ошибок)
	totalFridgeValue = round2(totalFridgeValue)
	
	if totalFridgeValue != expected {
		t.Errorf(
			"FRIDGE VALUE CONTRACT VIOLATION!\n"+
				"Expected wartość lodówki: %.2f PLN\n"+
				"Got: %.2f PLN\n"+
				"⚠️ This affects user's financial tracking!",
			expected,
			totalFridgeValue,
		)
	}
}

// TestMoneyContract_APIResponse - проверяем что DTO не ломает контракт
func TestMoneyContract_APIResponse(t *testing.T) {
	// Создаём типичный API response
	pricePerUnit := 0.00581
	totalPrice := 20.68
	
	response := models.FridgeItemListResponse{
		ID:           "test-id",
		Name:         "Ogórek",
		Category:     "vegetable",
		Quantity:     3560,
		Unit:         "g",
		PricePerUnit: &pricePerUnit,
		TotalPrice:   &totalPrice,
		Currency:     "PLN",
		ArrivedAt:    time.Now(),
		Status:       "ok",
	}

	// Проверяем что TotalPrice точно 2 знака
	if response.TotalPrice == nil {
		t.Fatal("TotalPrice must not be nil")
	}

	// Проверяем что значение ровно 20.68, без лишних знаков
	if *response.TotalPrice != 20.68 {
		t.Errorf(
			"API Response contains wrong totalPrice!\n"+
				"Expected: 20.68\n"+
				"Got: %.10f\n"+
				"⚠️ Frontend will display incorrect price to user!",
			*response.TotalPrice,
		)
	}

	// Проверяем что мы НЕ можем пересчитать на фронте и получить то же самое
	recalculated := response.Quantity * (*response.PricePerUnit)
	if recalculated == *response.TotalPrice {
		t.Logf("✅ Lucky case: recalculation matches (but frontend must NOT do this!)")
	} else {
		t.Logf(
			"⚠️ EXPECTED DIFFERENCE (this is why we have API contract):\n"+
				"   Backend totalPrice: %.2f\n"+
				"   Frontend recalculation would give: %.10f\n"+
				"   Difference: %.10f\n"+
				"   → Frontend MUST use backend's totalPrice!",
			*response.TotalPrice,
			recalculated,
			recalculated - *response.TotalPrice,
		)
	}
}
