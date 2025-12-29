package service

import (
	"testing"
)

func TestGameService_CheckWord(t *testing.T) {
	tests := []struct {
		name     string
		word     string
		target   string
		expected string
	}{
		{
			name:     "все буквы правильные",
			word:     "слово",
			target:   "слово",
			expected: "22222",
		},
		{
			name:     "все буквы неправильные",
			word:     "абвгд",
			target:   "ийклм",
			expected: "00000",
		},
		{
			name:     "буквы есть но не на месте",
			word:     "овсол",
			target:   "слово",
			expected: "11111",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := checkWordString(tt.word, tt.target)
			if result != tt.expected {
				t.Errorf("CheckWord(%s, %s) = %s, want %s", tt.word, tt.target, result, tt.expected)
			}
		})
	}
}

// checkWordString возвращает результат проверки слова как строку
func checkWordString(word, target string) string {
	wordRunes := []rune(word)
	targetRunes := []rune(target)
	
	result := make([]rune, len(wordRunes))
	targetUsed := make([]bool, len(targetRunes))
	
	// Первый проход: находим точные совпадения (2)
	for i := range wordRunes {
		if i < len(targetRunes) && wordRunes[i] == targetRunes[i] {
			result[i] = '2'
			targetUsed[i] = true
		} else {
			result[i] = '0'
		}
	}
	
	// Второй проход: находим буквы не на своем месте (1)
	for i := range wordRunes {
		if result[i] == '2' {
			continue
		}
		
		for j := range targetRunes {
			if !targetUsed[j] && wordRunes[i] == targetRunes[j] {
				result[i] = '1'
				targetUsed[j] = true
				break
			}
		}
	}
	
	return string(result)
}

func TestGameService_CalculateReward(t *testing.T) {
	commissionRate := 0.05 // 5%
	
	tests := []struct {
		name       string
		bet        float64
		multiplier float64
		expected   float64
	}{
		{
			name:       "базовый расчёт награды",
			bet:        1.0,
			multiplier: 2.0,
			expected:   1.9, // (1.0 * 2.0) * 0.95 = 1.9
		},
		{
			name:       "большая ставка",
			bet:        10.0,
			multiplier: 3.0,
			expected:   28.5, // (10.0 * 3.0) * 0.95 = 28.5
		},
		{
			name:       "маленькая ставка",
			bet:        0.1,
			multiplier: 1.5,
			expected:   0.1425, // (0.1 * 1.5) * 0.95 = 0.1425
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateGameReward(tt.bet, tt.multiplier, commissionRate)
			// Допускаем небольшую погрешность из-за float
			if result < tt.expected-0.001 || result > tt.expected+0.001 {
				t.Errorf("CalculateReward() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func calculateGameReward(bet, multiplier, commissionRate float64) float64 {
	grossReward := bet * multiplier
	commission := grossReward * commissionRate
	return grossReward - commission
}

func TestGameService_ValidateGameParams(t *testing.T) {
	tests := []struct {
		name       string
		word       string
		minBet     float64
		maxBet     float64
		maxTries   int
		wantErr    bool
		errMessage string
	}{
		{
			name:     "валидные параметры",
			word:     "тест",
			minBet:   0.1,
			maxBet:   1.0,
			maxTries: 6,
			wantErr:  false,
		},
		{
			name:       "пустое слово",
			word:       "",
			minBet:     0.1,
			maxBet:     1.0,
			maxTries:   6,
			wantErr:    true,
			errMessage: "word is required",
		},
		{
			name:       "min_bet больше max_bet",
			word:       "тест",
			minBet:     10.0,
			maxBet:     1.0,
			maxTries:   6,
			wantErr:    true,
			errMessage: "min_bet > max_bet",
		},
		{
			name:       "max_tries = 0",
			word:       "тест",
			minBet:     0.1,
			maxBet:     1.0,
			maxTries:   0,
			wantErr:    true,
			errMessage: "max_tries must be positive",
		},
		{
			name:       "отрицательная min_bet",
			word:       "тест",
			minBet:     -0.1,
			maxBet:     1.0,
			maxTries:   6,
			wantErr:    true,
			errMessage: "min_bet must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateGameParams(tt.word, tt.minBet, tt.maxBet, tt.maxTries)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateGameParams() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func validateGameParams(word string, minBet, maxBet float64, maxTries int) error {
	if word == "" {
		return errWordRequired
	}
	if minBet <= 0 {
		return errMinBetPositive
	}
	if maxBet <= 0 {
		return errMaxBetPositive
	}
	if minBet > maxBet {
		return errMinBetGreaterThanMax
	}
	if maxTries <= 0 {
		return errMaxTriesPositive
	}
	return nil
}

// Errors for validation
type validationError string

func (e validationError) Error() string { return string(e) }

const (
	errWordRequired         = validationError("word is required")
	errMinBetPositive       = validationError("min_bet must be positive")
	errMaxBetPositive       = validationError("max_bet must be positive")
	errMinBetGreaterThanMax = validationError("min_bet > max_bet")
	errMaxTriesPositive     = validationError("max_tries must be positive")
)

func TestGameService_GenerateShortID(t *testing.T) {
	// Тест генерации короткого ID
	ids := make(map[string]bool)
	
	for i := 0; i < 100; i++ {
		id := generateShortID() // Используем функцию из game_service.go
		if len(id) != 8 {
			t.Errorf("generateShortID() length = %d, want 8", len(id))
		}
		if ids[id] {
			t.Errorf("generateShortID() generated duplicate: %s", id)
		}
		ids[id] = true
	}
}

// Benchmark тесты
func BenchmarkCheckWordString(b *testing.B) {
	word := "столб"
	target := "слово"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		checkWordString(word, target)
	}
}

func BenchmarkValidateGameParams(b *testing.B) {
	word := "тест"
	minBet := 0.1
	maxBet := 1.0
	maxTries := 6
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validateGameParams(word, minBet, maxBet, maxTries)
	}
}
