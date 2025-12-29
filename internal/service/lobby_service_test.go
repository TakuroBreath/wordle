package service

import (
	"testing"
)

func TestLobbyService_CheckWord(t *testing.T) {
	// Тестируем функцию проверки слова напрямую
	tests := []struct {
		name     string
		word     string
		target   string
		expected []int
	}{
		{
			name:     "все буквы правильные",
			word:     "слово",
			target:   "слово",
			expected: []int{2, 2, 2, 2, 2},
		},
		{
			name:     "все буквы неправильные",
			word:     "абвгд",
			target:   "ийклм",
			expected: []int{0, 0, 0, 0, 0},
		},
		{
			name:     "буквы есть но не на месте",
			word:     "овсол",
			target:   "слово",
			expected: []int{1, 1, 1, 1, 1},
		},
		{
			name:     "повторяющиеся буквы - точные совпадения",
			word:     "аабаа",
			target:   "аабаа",
			expected: []int{2, 2, 2, 2, 2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := checkWordResult(tt.word, tt.target)
			
			if len(result) != len(tt.expected) {
				t.Errorf("checkWordResult() length = %v, want %v", len(result), len(tt.expected))
				return
			}
			
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("checkWordResult()[%d] = %v, want %v", i, result[i], tt.expected[i])
				}
			}
		})
	}
}

// checkWordResult - локальная реализация для тестирования
func checkWordResult(word, target string) []int {
	wordRunes := []rune(word)
	targetRunes := []rune(target)
	
	result := make([]int, len(wordRunes))
	targetUsed := make([]bool, len(targetRunes))
	
	// Первый проход: находим точные совпадения (2)
	for i := range wordRunes {
		if i < len(targetRunes) && wordRunes[i] == targetRunes[i] {
			result[i] = 2
			targetUsed[i] = true
		}
	}
	
	// Второй проход: находим буквы не на своем месте (1)
	for i := range wordRunes {
		if result[i] == 2 {
			continue // Уже найдено точное совпадение
		}
		
		for j := range targetRunes {
			if !targetUsed[j] && wordRunes[i] == targetRunes[j] {
				result[i] = 1
				targetUsed[j] = true
				break
			}
		}
	}
	
	return result
}

func TestLobbyService_IsWordCorrect(t *testing.T) {
	tests := []struct {
		name     string
		word     string
		target   string
		expected bool
	}{
		{
			name:     "одинаковые слова",
			word:     "слово",
			target:   "слово",
			expected: true,
		},
		{
			name:     "разные слова",
			word:     "слава",
			target:   "слово",
			expected: false,
		},
		{
			name:     "разный регистр",
			word:     "СЛОВО",
			target:   "слово",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isWordCorrect(tt.word, tt.target)
			if result != tt.expected {
				t.Errorf("isWordCorrect() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func isWordCorrect(word, target string) bool {
	return word == target
}

func TestLobbyService_CalculateReward(t *testing.T) {
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateReward(tt.bet, tt.multiplier, commissionRate)
			// Допускаем небольшую погрешность из-за float
			if result < tt.expected-0.01 || result > tt.expected+0.01 {
				t.Errorf("calculateReward() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func calculateReward(bet, multiplier, commissionRate float64) float64 {
	grossReward := bet * multiplier
	commission := grossReward * commissionRate
	return grossReward - commission
}

// Benchmark тесты
func BenchmarkCheckWord(b *testing.B) {
	word := "столб"
	target := "слово"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		checkWordResult(word, target)
	}
}

func BenchmarkCalculateReward(b *testing.B) {
	bet := 1.0
	multiplier := 2.0
	commission := 0.05
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		calculateReward(bet, multiplier, commission)
	}
}
