package transparencia

import (
	"testing"
)

func TestParseMoney(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected float64
	}{
		{"Empty string", "", 0},
		{"Formatted BRL", "R$ 1.234,56", 1234.56},
		{"Simple number", "100,50", 100.50},
		{"Large number", "R$ 1.000.000,00", 1000000.00},
		{"Zero BRL", "R$ 0,00", 0.00},
		{"Invalid format", "invalid", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseMoney(tt.input)
			if got != tt.expected {
				t.Errorf("ParseMoney(%q) = %v; want %v", tt.input, got, tt.expected)
			}
		})
	}
}
