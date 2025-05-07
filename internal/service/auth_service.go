package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/TakuroBreath/wordle/internal/models"
	"github.com/TakuroBreath/wordle/internal/repository"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// AuthServiceImpl представляет собой реализацию сервиса для аутентификации пользователей
type AuthServiceImpl struct {
	userRepo  models.UserRepository
	redisRepo repository.RedisRepository
	jwtSecret string
	botToken  string
}

func NewAuthService(userRepo models.UserRepository, redisRepo repository.RedisRepository) AuthService {
	// В реальном приложении секреты должны загружаться из переменных окружения или конфигурационного файла
	return &AuthServiceImpl{
		userRepo:  userRepo,
		redisRepo: redisRepo,
		jwtSecret: "your-jwt-secret", // Заменить на реальный секрет
		botToken:  "your-bot-token",  // Заменить на реальный токен бота
	}
}

// Инициализация AuthServiceImpl происходит в service.go через функцию NewAuthService

// InitAuth инициализирует аутентификацию пользователя на основе данных от Telegram
func (s *AuthServiceImpl) InitAuth(ctx context.Context, initData string) (string, error) {
	// Проверка валидности initData
	if initData == "" {
		return "", errors.New("init data cannot be empty")
	}

	// Парсинг initData
	params, err := url.ParseQuery(initData)
	if err != nil {
		return "", fmt.Errorf("failed to parse init data: %w", err)
	}

	// Проверка подписи
	if !s.validateInitData(params) {
		return "", errors.New("invalid init data signature")
	}

	// Получение данных пользователя
	userStr := params.Get("user")
	if userStr == "" {
		return "", errors.New("user data not found")
	}

	// Парсинг данных пользователя
	var userData struct {
		ID        int64  `json:"id"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name,omitempty"`
		Username  string `json:"username,omitempty"`
	}

	err = json.Unmarshal([]byte(userStr), &userData)
	if err != nil {
		return "", fmt.Errorf("failed to parse user data: %w", err)
	}

	// Проверка существования пользователя
	user, err := s.userRepo.GetByTelegramID(ctx, userData.ID)
	if err != nil {
		// Пользователь не найден, создаем нового
		username := userData.Username
		if username == "" {
			username = fmt.Sprintf("%s %s", userData.FirstName, userData.LastName)
		}

		user = &models.User{
			ID:         uuid.New(),
			TelegramID: userData.ID,
			Username:   username,
			Balance:    0,
			Wins:       0,
			Losses:     0,
		}

		err = s.userRepo.Create(ctx, user)
		if err != nil {
			return "", fmt.Errorf("failed to create user: %w", err)
		}
	}

	// Генерация JWT токена
	token, err := s.GenerateToken(ctx, *user)
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	return token, nil
}

// VerifyAuth проверяет аутентификацию пользователя
func (s *AuthServiceImpl) VerifyAuth(ctx context.Context, token string) (models.User, error) {
	// Проверка валидности токена
	user, err := s.ValidateToken(ctx, token)
	if err != nil {
		return models.User{}, fmt.Errorf("invalid token: %w", err)
	}

	return user, nil
}

// GenerateToken генерирует JWT токен для пользователя
func (s *AuthServiceImpl) GenerateToken(ctx context.Context, user models.User) (string, error) {
	// Создание JWT токена
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":     user.ID.String(),
		"telegram_id": user.TelegramID,
		"username":    user.Username,
		"exp":         time.Now().Add(24 * time.Hour).Unix(), // Токен действителен 24 часа
	})

	// Подписание токена
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	// Сохранение токена в Redis для возможности инвалидации
	err = s.redisRepo.SetSession(ctx, fmt.Sprintf("token:%s", user.ID.String()), tokenString, 24*60*60) // 24 часа
	if err != nil {
		return "", fmt.Errorf("failed to save token: %w", err)
	}

	return tokenString, nil
}

// ValidateToken проверяет валидность JWT токена
func (s *AuthServiceImpl) ValidateToken(ctx context.Context, tokenString string) (models.User, error) {
	// Парсинг токена
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Проверка метода подписи
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return models.User{}, fmt.Errorf("failed to parse token: %w", err)
	}

	// Проверка валидности токена
	if !token.Valid {
		return models.User{}, errors.New("invalid token")
	}

	// Получение данных из токена
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return models.User{}, errors.New("invalid token claims")
	}

	// Получение ID пользователя
	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		return models.User{}, errors.New("user ID not found in token")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return models.User{}, fmt.Errorf("invalid user ID: %w", err)
	}

	// Проверка, что токен не был инвалидирован
	savedToken, err := s.redisRepo.GetSession(ctx, fmt.Sprintf("token:%s", userIDStr))
	if err != nil || savedToken != tokenString {
		return models.User{}, errors.New("token has been invalidated")
	}

	// Получение пользователя из базы данных
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return models.User{}, fmt.Errorf("failed to get user: %w", err)
	}

	return *user, nil
}

// validateInitData проверяет подпись initData от Telegram
func (s *AuthServiceImpl) validateInitData(params url.Values) bool {
	// Получение хеша
	hash := params.Get("hash")
	if hash == "" {
		return false
	}

	// Удаление хеша из параметров для проверки
	params.Del("hash")

	// Сортировка параметров
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Формирование строки для проверки
	var dataCheckString strings.Builder
	for _, k := range keys {
		for _, v := range params[k] {
			dataCheckString.WriteString(k)
			dataCheckString.WriteString("=")
			dataCheckString.WriteString(v)
			dataCheckString.WriteString("\n")
		}
	}

	// Вычисление HMAC-SHA256
	h := hmac.New(sha256.New, []byte("WebAppData"))
	h.Write([]byte(dataCheckString.String()))
	signature := hex.EncodeToString(h.Sum(nil))

	// Сравнение хешей
	return signature == hash
}

// validateAuthTimestamp проверяет, что временная метка не устарела
func (s *AuthServiceImpl) validateAuthTimestamp(authDate string) bool {
	// Парсинг временной метки
	timestamp, err := strconv.ParseInt(authDate, 10, 64)
	if err != nil {
		return false
	}

	// Проверка, что временная метка не старше 24 часов
	return time.Now().Unix()-timestamp < 24*60*60
}
