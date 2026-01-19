package utils

import "testing"

func TestCapitalize(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "lowercase english",
			input:    "bacon",
			expected: "Bacon",
		},
		{
			name:     "lowercase russian",
			input:    "бекон",
			expected: "Бекон",
		},
		{
			name:     "lowercase polish with diacritics",
			input:    "łosoś",
			expected: "Łosoś",
		},
		{
			name:     "already capitalized",
			input:    "Salmon",
			expected: "Salmon",
		},
		{
			name:     "all uppercase (keeps rest as-is)",
			input:    "EGGS",
			expected: "EGGS",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "whitespace only",
			input:    "   ",
			expected: "",
		},
		{
			name:     "with leading/trailing spaces",
			input:    "  test  ",
			expected: "Test",
		},
		{
			name:     "single character",
			input:    "a",
			expected: "A",
		},
		{
			name:     "russian with yo",
			input:    "ёлка",
			expected: "Ёлка",
		},
		{
			name:     "polish l with stroke",
			input:    "łódka",
			expected: "Łódka",
		},
		{
			name:     "mixed case keeps original",
			input:    "iPhone",
			expected: "IPhone",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Capitalize(tt.input)
			if result != tt.expected {
				t.Errorf("Capitalize(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestCapitalizeWords(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "two words english",
			input:    "olive oil",
			expected: "Olive Oil",
		},
		{
			name:     "two words russian",
			input:    "красный перец",
			expected: "Красный Перец",
		},
		{
			name:     "already capitalized",
			input:    "Sea Salt",
			expected: "Sea Salt",
		},
		{
			name:     "single word",
			input:    "bacon",
			expected: "Bacon",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "multiple spaces",
			input:    "sea   salt",
			expected: "Sea Salt",
		},
		{
			name:     "with leading/trailing spaces",
			input:    "  olive oil  ",
			expected: "Olive Oil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CapitalizeWords(tt.input)
			if result != tt.expected {
				t.Errorf("CapitalizeWords(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}
