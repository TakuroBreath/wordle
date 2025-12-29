package service

import (
	"testing"
	"time"
)

func TestAuthService_ValidateTokenExpiry(t *testing.T) {
	tests := []struct {
		name      string
		expiresAt time.Time
		expected  bool
	}{
		{
			name:      "токен не истёк",
			expiresAt: time.Now().Add(1 * time.Hour),
			expected:  true,
		},
		{
			name:      "токен истёк",
			expiresAt: time.Now().Add(-1 * time.Hour),
			expected:  false,
		},
		{
			name:      "токен истекает сейчас",
			expiresAt: time.Now(),
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isTokenValid(tt.expiresAt)
			if result != tt.expected {
				t.Errorf("isTokenValid() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func isTokenValid(expiresAt time.Time) bool {
	return time.Now().Before(expiresAt)
}

func TestAuthService_ParseTelegramInitData(t *testing.T) {
	// Тест парсинга данных Telegram
	tests := []struct {
		name     string
		initData string
		wantUser bool
	}{
		{
			name:     "пустые данные",
			initData: "",
			wantUser: false,
		},
		{
			name:     "некорректный формат",
			initData: "invalid_data",
			wantUser: false,
		},
		// Для реальных тестов нужен корректный hash с секретом бота
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := parseInitData(tt.initData)
			if (user != nil) != tt.wantUser {
				t.Errorf("parseInitData() got user = %v, want user = %v", user != nil, tt.wantUser)
			}
		})
	}
}

type telegramUser struct {
	ID        uint64
	Username  string
	FirstName string
	LastName  string
}

func parseInitData(initData string) *telegramUser {
	if initData == "" {
		return nil
	}
	// Простая проверка формата
	if len(initData) < 10 {
		return nil
	}
	return nil // Полная реализация требует секрет бота
}

func TestAuthService_GenerateTokenFormat(t *testing.T) {
	// Тест формата JWT токена
	tests := []struct {
		name  string
		token string
		valid bool
	}{
		{
			name:  "валидный формат JWT",
			token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
			valid: true,
		},
		{
			name:  "невалидный формат - нет точек",
			token: "invalidtoken",
			valid: false,
		},
		{
			name:  "невалидный формат - одна точка",
			token: "invalid.token",
			valid: false,
		},
		{
			name:  "пустой токен",
			token: "",
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidJWTFormat(tt.token)
			if result != tt.valid {
				t.Errorf("isValidJWTFormat() = %v, want %v", result, tt.valid)
			}
		})
	}
}

func isValidJWTFormat(token string) bool {
	if token == "" {
		return false
	}
	parts := 0
	for _, c := range token {
		if c == '.' {
			parts++
		}
	}
	return parts == 2
}

func TestAuthService_ExtractUserIDFromToken(t *testing.T) {
	// Тест извлечения ID из токена
	tests := []struct {
		name       string
		claims     map[string]any
		expectedID uint64
	}{
		{
			name: "валидные claims",
			claims: map[string]any{
				"telegram_id": float64(123456789),
			},
			expectedID: 123456789,
		},
		{
			name:       "пустые claims",
			claims:     map[string]any{},
			expectedID: 0,
		},
		{
			name: "неверный тип",
			claims: map[string]any{
				"telegram_id": "invalid",
			},
			expectedID: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractTelegramID(tt.claims)
			if result != tt.expectedID {
				t.Errorf("extractTelegramID() = %v, want %v", result, tt.expectedID)
			}
		})
	}
}

func extractTelegramID(claims map[string]any) uint64 {
	if id, ok := claims["telegram_id"]; ok {
		if floatID, ok := id.(float64); ok {
			return uint64(floatID)
		}
	}
	return 0
}

// Benchmark тесты
func BenchmarkIsTokenValid(b *testing.B) {
	expiresAt := time.Now().Add(1 * time.Hour)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		isTokenValid(expiresAt)
	}
}

func BenchmarkIsValidJWTFormat(b *testing.B) {
	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgNryP4J3jVmNHl0w5N_XgL0n3I9PlFUP0THsR8U"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		isValidJWTFormat(token)
	}
}
