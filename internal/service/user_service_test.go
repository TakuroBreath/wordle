package service

import (
	"testing"
)

func TestUserService_ValidateBalance(t *testing.T) {
	tests := []struct {
		name           string
		balance        float64
		requiredAmount float64
		expected       bool
	}{
		{
			name:           "достаточно средств",
			balance:        10.0,
			requiredAmount: 5.0,
			expected:       true,
		},
		{
			name:           "недостаточно средств",
			balance:        5.0,
			requiredAmount: 10.0,
			expected:       false,
		},
		{
			name:           "точно равно",
			balance:        5.0,
			requiredAmount: 5.0,
			expected:       true,
		},
		{
			name:           "нулевой баланс",
			balance:        0.0,
			requiredAmount: 1.0,
			expected:       false,
		},
		{
			name:           "нулевое требование",
			balance:        10.0,
			requiredAmount: 0.0,
			expected:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validateBalance(tt.balance, tt.requiredAmount)
			if result != tt.expected {
				t.Errorf("validateBalance() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func validateBalance(balance, requiredAmount float64) bool {
	return balance >= requiredAmount
}

func TestUserService_CalculateAvailableBalance(t *testing.T) {
	tests := []struct {
		name              string
		balance           float64
		pendingWithdrawal float64
		expected          float64
	}{
		{
			name:              "без pending",
			balance:           100.0,
			pendingWithdrawal: 0.0,
			expected:          100.0,
		},
		{
			name:              "с pending",
			balance:           100.0,
			pendingWithdrawal: 30.0,
			expected:          70.0,
		},
		{
			name:              "pending = balance",
			balance:           50.0,
			pendingWithdrawal: 50.0,
			expected:          0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateAvailableBalance(tt.balance, tt.pendingWithdrawal)
			if result != tt.expected {
				t.Errorf("calculateAvailableBalance() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func calculateAvailableBalance(balance, pendingWithdrawal float64) float64 {
	available := balance - pendingWithdrawal
	if available < 0 {
		return 0
	}
	return available
}

func TestUserService_ValidateWalletAddress(t *testing.T) {
	tests := []struct {
		name     string
		address  string
		expected bool
	}{
		{
			name:     "валидный TON адрес (EQ)",
			address:  "EQDtFpEwcFAEcRe5mLVh2N6C0x-_hJEM7W61_JLnSF74p4q2",
			expected: true,
		},
		{
			name:     "валидный TON адрес (UQ)",
			address:  "UQDtFpEwcFAEcRe5mLVh2N6C0x-_hJEM7W61_JLnSF74p4q2",
			expected: true,
		},
		{
			name:     "пустой адрес",
			address:  "",
			expected: false,
		},
		{
			name:     "короткий адрес",
			address:  "EQDtFpEw",
			expected: false,
		},
		{
			name:     "неправильный префикс",
			address:  "XXDtFpEwcFAEcRe5mLVh2N6C0x-_hJEM7W61_JLnSF74p4q2",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validateWalletAddress(tt.address)
			if result != tt.expected {
				t.Errorf("validateWalletAddress() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func validateWalletAddress(address string) bool {
	if len(address) < 48 {
		return false
	}
	prefix := address[:2]
	return prefix == "EQ" || prefix == "UQ"
}

func TestUserService_ValidateWithdrawalAmount(t *testing.T) {
	minWithdraw := 0.1
	
	tests := []struct {
		name     string
		amount   float64
		balance  float64
		expected bool
	}{
		{
			name:     "валидная сумма",
			amount:   1.0,
			balance:  10.0,
			expected: true,
		},
		{
			name:     "сумма больше баланса",
			amount:   15.0,
			balance:  10.0,
			expected: false,
		},
		{
			name:     "сумма меньше минимума",
			amount:   0.05,
			balance:  10.0,
			expected: false,
		},
		{
			name:     "отрицательная сумма",
			amount:   -1.0,
			balance:  10.0,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validateWithdrawalAmount(tt.amount, tt.balance, minWithdraw)
			if result != tt.expected {
				t.Errorf("validateWithdrawalAmount() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func validateWithdrawalAmount(amount, balance, minWithdraw float64) bool {
	if amount <= 0 {
		return false
	}
	if amount < minWithdraw {
		return false
	}
	if amount > balance {
		return false
	}
	return true
}

// Benchmark тесты
func BenchmarkValidateBalance(b *testing.B) {
	balance := 100.0
	required := 50.0
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validateBalance(balance, required)
	}
}

func BenchmarkValidateWalletAddress(b *testing.B) {
	address := "EQDtFpEwcFAEcRe5mLVh2N6C0x-_hJEM7W61_JLnSF74p4q2"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validateWalletAddress(address)
	}
}
