package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	return r
}

func TestHealthCheck(t *testing.T) {
	router := setupTestRouter()
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response["status"] != "ok" {
		t.Errorf("Expected status 'ok', got '%v'", response["status"])
	}
}

func TestValidateJSONRequest(t *testing.T) {
	router := setupTestRouter()
	router.POST("/test", func(c *gin.Context) {
		var input struct {
			Name  string `json:"name" binding:"required"`
			Value int    `json:"value" binding:"required,min=1"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"name": input.Name, "value": input.Value})
	})

	tests := []struct {
		name           string
		payload        map[string]any
		expectedStatus int
	}{
		{
			name:           "валидный запрос",
			payload:        map[string]any{"name": "test", "value": 10},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "без обязательного поля name",
			payload:        map[string]any{"value": 10},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "без обязательного поля value",
			payload:        map[string]any{"name": "test"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "value меньше минимума",
			payload:        map[string]any{"name": "test", "value": 0},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "пустой запрос",
			payload:        map[string]any{},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest("POST", "/test", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d. Body: %s", tt.expectedStatus, w.Code, w.Body.String())
			}
		})
	}
}

func TestPaginationParams(t *testing.T) {
	router := setupTestRouter()
	router.GET("/list", func(c *gin.Context) {
		limit := 10
		offset := 0

		if limitStr := c.Query("limit"); limitStr != "" {
			if val := parseInt(limitStr); val > 0 {
				limit = val
			}
		}

		if offsetStr := c.Query("offset"); offsetStr != "" {
			if val := parseInt(offsetStr); val >= 0 {
				offset = val
			}
		}

		c.JSON(http.StatusOK, gin.H{"limit": limit, "offset": offset})
	})

	tests := []struct {
		name           string
		query          string
		expectedLimit  int
		expectedOffset int
	}{
		{
			name:           "дефолтные значения",
			query:          "",
			expectedLimit:  10,
			expectedOffset: 0,
		},
		{
			name:           "custom limit",
			query:          "?limit=20",
			expectedLimit:  20,
			expectedOffset: 0,
		},
		{
			name:           "custom offset",
			query:          "?offset=5",
			expectedLimit:  10,
			expectedOffset: 5,
		},
		{
			name:           "custom limit и offset",
			query:          "?limit=50&offset=100",
			expectedLimit:  50,
			expectedOffset: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/list"+tt.query, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
			}

			var response map[string]any
			_ = json.Unmarshal(w.Body.Bytes(), &response)

			if int(response["limit"].(float64)) != tt.expectedLimit {
				t.Errorf("Expected limit %d, got %v", tt.expectedLimit, response["limit"])
			}
			if int(response["offset"].(float64)) != tt.expectedOffset {
				t.Errorf("Expected offset %d, got %v", tt.expectedOffset, response["offset"])
			}
		})
	}
}

func parseInt(s string) int {
	result := 0
	for _, c := range s {
		if c >= '0' && c <= '9' {
			result = result*10 + int(c-'0')
		}
	}
	return result
}

func TestUUIDValidation(t *testing.T) {
	router := setupTestRouter()
	router.GET("/game/:id", func(c *gin.Context) {
		id := c.Param("id")
		if !isValidUUID(id) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid game ID"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"id": id})
	})

	tests := []struct {
		name           string
		id             string
		expectedStatus int
	}{
		{
			name:           "валидный UUID",
			id:             "550e8400-e29b-41d4-a716-446655440000",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "невалидный UUID",
			id:             "invalid-uuid",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "короткий ID (short_id)",
			id:             "ABCD1234",
			expectedStatus: http.StatusBadRequest, // UUID формат ожидается
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/game/"+tt.id, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d. Body: %s", tt.expectedStatus, w.Code, w.Body.String())
			}
		})
	}
}

func isValidUUID(s string) bool {
	// Простая проверка формата UUID: 8-4-4-4-12
	if len(s) != 36 {
		return false
	}
	for i, c := range s {
		if i == 8 || i == 13 || i == 18 || i == 23 {
			if c != '-' {
				return false
			}
		} else {
			if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
				return false
			}
		}
	}
	return true
}

func TestCORSHeaders(t *testing.T) {
	router := setupTestRouter()
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})
	router.GET("/api/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	t.Run("OPTIONS request", func(t *testing.T) {
		req := httptest.NewRequest("OPTIONS", "/api/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNoContent {
			t.Errorf("Expected status %d, got %d", http.StatusNoContent, w.Code)
		}

		if w.Header().Get("Access-Control-Allow-Origin") != "*" {
			t.Error("Missing or incorrect Access-Control-Allow-Origin header")
		}
	})

	t.Run("GET request with CORS", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		if w.Header().Get("Access-Control-Allow-Origin") != "*" {
			t.Error("Missing or incorrect Access-Control-Allow-Origin header")
		}
	})
}

func TestErrorResponse(t *testing.T) {
	router := setupTestRouter()
	router.GET("/error/:code", func(c *gin.Context) {
		code := c.Param("code")
		switch code {
		case "400":
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		case "401":
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		case "403":
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		case "404":
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		case "500":
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		default:
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		}
	})

	tests := []struct {
		code           string
		expectedStatus int
		expectedError  string
	}{
		{"400", http.StatusBadRequest, "bad request"},
		{"401", http.StatusUnauthorized, "unauthorized"},
		{"403", http.StatusForbidden, "forbidden"},
		{"404", http.StatusNotFound, "not found"},
		{"500", http.StatusInternalServerError, "internal server error"},
	}

	for _, tt := range tests {
		t.Run("error_"+tt.code, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/error/"+tt.code, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			var response map[string]any
			_ = json.Unmarshal(w.Body.Bytes(), &response)
			if response["error"] != tt.expectedError {
				t.Errorf("Expected error '%s', got '%v'", tt.expectedError, response["error"])
			}
		})
	}
}

// Benchmark тесты
func BenchmarkHealthCheck(b *testing.B) {
	router := setupTestRouter()
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

func BenchmarkJSONBinding(b *testing.B) {
	router := setupTestRouter()
	router.POST("/test", func(c *gin.Context) {
		var input struct {
			Name  string `json:"name"`
			Value int    `json:"value"`
		}
		_ = c.ShouldBindJSON(&input)
		c.JSON(http.StatusOK, input)
	})

	payload := []byte(`{"name":"test","value":10}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/test", bytes.NewBuffer(payload))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}
