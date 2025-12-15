package service

import (
	"testing"
)

func TestRound2(t *testing.T) {
	tests := []struct {
		name     string
		input    float64
		expected float64
	}{
		{
			name:     "Типичная проблема float: 10.350000000000001",
			input:    10.350000000000001,
			expected: 10.35,
		},
		{
			name:     "Уже округлённое число",
			input:    15.99,
			expected: 15.99,
		},
		{
			name:     "Целое число",
			input:    20.0,
			expected: 20.0,
		},
		{
			name:     "Округление вверх: 3.456 → 3.46",
			input:    3.456,
			expected: 3.46,
		},
		{
			name:     "Округление вниз: 7.123 → 7.12",
			input:    7.123,
			expected: 7.12,
		},
		{
			name:     "Реальный пример: 3560 * 0.00581 = 20.6836",
			input:    3560 * 0.00581,
			expected: 20.68,
		},
		{
			name:     "Маленькое число",
			input:    0.01,
			expected: 0.01,
		},
		{
			name:     "Ноль",
			input:    0.0,
			expected: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := round2(tt.input)
			if result != tt.expected {
				t.Errorf("round2(%f) = %f, want %f", tt.input, result, tt.expected)
			}
		})
	}
}

func TestCalculateTotalPrice(t *testing.T) {
	service := &FridgeService{}

	t.Run("Nil price returns nil", func(t *testing.T) {
		result := service.calculateTotalPrice(100, nil)
		if result != nil {
			t.Errorf("Expected nil, got %v", result)
		}
	})

	t.Run("Real example: 3560g * 0.00581 PLN/g = 20.68 PLN", func(t *testing.T) {
		price := 0.00581
		result := service.calculateTotalPrice(3560, &price)
		
		if result == nil {
			t.Fatal("Expected non-nil result")
		}
		
		expected := 20.68
		if *result != expected {
			t.Errorf("calculateTotalPrice(3560, 0.00581) = %.2f, want %.2f", *result, expected)
		}
	})

	t.Run("Float precision issue: should round correctly", func(t *testing.T) {
		// Этот тест воспроизводит проблему 10.350000000000001
		price := 0.00290845
		result := service.calculateTotalPrice(3560, &price)
		
		if result == nil {
			t.Fatal("Expected non-nil result")
		}
		
		// 3560 * 0.00290845 = 10.3540820... → должно округлиться до 10.35
		expected := 10.35
		if *result != expected {
			t.Errorf("calculateTotalPrice(3560, 0.00290845) = %.10f, want %.2f", *result, expected)
		}
	})
}
