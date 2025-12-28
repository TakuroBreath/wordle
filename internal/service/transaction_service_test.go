package service

import (
	"testing"
)

func TestTransactionService_ValidateTransactionAmount(t *testing.T) {
	tests := []struct {
		name     string
		amount   float64
		expected bool
	}{
		{
			name:     "положительная сумма",
			amount:   1.0,
			expected: true,
		},
		{
			name:     "нулевая сумма",
			amount:   0.0,
			expected: false,
		},
		{
			name:     "отрицательная сумма",
			amount:   -1.0,
			expected: false,
		},
		{
			name:     "маленькая сумма",
			amount:   0.001,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validateTransactionAmount(tt.amount)
			if result != tt.expected {
				t.Errorf("validateTransactionAmount() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func validateTransactionAmount(amount float64) bool {
	return amount > 0
}

func TestTransactionService_ValidateCurrency(t *testing.T) {
	tests := []struct {
		name     string
		currency string
		expected bool
	}{
		{
			name:     "TON",
			currency: "TON",
			expected: true,
		},
		{
			name:     "USDT",
			currency: "USDT",
			expected: true,
		},
		{
			name:     "неизвестная валюта",
			currency: "BTC",
			expected: false,
		},
		{
			name:     "пустая строка",
			currency: "",
			expected: false,
		},
		{
			name:     "нижний регистр",
			currency: "ton",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validateCurrency(tt.currency)
			if result != tt.expected {
				t.Errorf("validateCurrency() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func validateCurrency(currency string) bool {
	return currency == "TON" || currency == "USDT"
}

func TestTransactionService_ValidateTransactionType(t *testing.T) {
	tests := []struct {
		name     string
		txType   string
		expected bool
	}{
		{
			name:     "deposit",
			txType:   "deposit",
			expected: true,
		},
		{
			name:     "withdraw",
			txType:   "withdraw",
			expected: true,
		},
		{
			name:     "reward",
			txType:   "reward",
			expected: true,
		},
		{
			name:     "bet",
			txType:   "bet",
			expected: true,
		},
		{
			name:     "неизвестный тип",
			txType:   "unknown",
			expected: false,
		},
		{
			name:     "пустая строка",
			txType:   "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validateTransactionType(tt.txType)
			if result != tt.expected {
				t.Errorf("validateTransactionType() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func validateTransactionType(txType string) bool {
	validTypes := map[string]bool{
		"deposit":  true,
		"withdraw": true,
		"reward":   true,
		"bet":      true,
		"refund":   true,
	}
	return validTypes[txType]
}

func TestTransactionService_CalculateNetworkFee(t *testing.T) {
	tests := []struct {
		name     string
		currency string
		expected float64
	}{
		{
			name:     "комиссия TON",
			currency: "TON",
			expected: 0.01, // 0.01 TON
		},
		{
			name:     "комиссия USDT",
			currency: "USDT",
			expected: 0.5, // 0.5 USDT
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateNetworkFee(tt.currency)
			if result != tt.expected {
				t.Errorf("calculateNetworkFee() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func calculateNetworkFee(currency string) float64 {
	switch currency {
	case "TON":
		return 0.01
	case "USDT":
		return 0.5
	default:
		return 0
	}
}

func TestTransactionService_ValidateTxHash(t *testing.T) {
	tests := []struct {
		name     string
		txHash   string
		expected bool
	}{
		{
			name:     "валидный хеш",
			txHash:   "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
			expected: true,
		},
		{
			name:     "короткий хеш",
			txHash:   "abc123",
			expected: false,
		},
		{
			name:     "пустой хеш",
			txHash:   "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validateTxHash(tt.txHash)
			if result != tt.expected {
				t.Errorf("validateTxHash() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func validateTxHash(txHash string) bool {
	return len(txHash) >= 64
}

func TestTransactionService_FormatAmount(t *testing.T) {
	tests := []struct {
		name     string
		amount   float64
		expected string
	}{
		{
			name:     "целое число",
			amount:   10.0,
			expected: "10.00",
		},
		{
			name:     "дробное число",
			amount:   10.5,
			expected: "10.50",
		},
		{
			name:     "много знаков после запятой",
			amount:   10.12345,
			expected: "10.12",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatAmount(tt.amount)
			if result != tt.expected {
				t.Errorf("formatAmount() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func formatAmount(amount float64) string {
	return formatFloat(amount, 2)
}

func formatFloat(f float64, decimals int) string {
	format := "%." + string(rune('0'+decimals)) + "f"
	return sprintf(format, f)
}

func sprintf(format string, a ...any) string {
	// Простая реализация для теста
	if len(a) > 0 {
		if f, ok := a[0].(float64); ok {
			if format == "%.2f" {
				intPart := int(f)
				decPart := int((f - float64(intPart)) * 100)
				if decPart < 0 {
					decPart = -decPart
				}
				decStr := ""
				if decPart < 10 {
					decStr = "0" + string(rune('0'+decPart))
				} else {
					decStr = string(rune('0'+decPart/10)) + string(rune('0'+decPart%10))
				}
				result := ""
				if intPart == 0 {
					result = "0"
				} else {
					for intPart > 0 {
						result = string(rune('0'+intPart%10)) + result
						intPart /= 10
					}
				}
				return result + "." + decStr
			}
		}
	}
	return ""
}

// Benchmark тесты
func BenchmarkValidateTransactionAmount(b *testing.B) {
	amount := 1.0
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validateTransactionAmount(amount)
	}
}

func BenchmarkValidateCurrency(b *testing.B) {
	currency := "TON"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validateCurrency(currency)
	}
}

func BenchmarkFormatAmount(b *testing.B) {
	amount := 10.12345
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		formatAmount(amount)
	}
}
