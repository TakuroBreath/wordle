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
)

// AuthServiceImpl представляет собой реализацию сервиса для аутентификации пользователей
type AuthServiceImpl struct {
	userRepo  models.UserRepository
	redisRepo repository.RedisRepository
	jwtSecret string
	botToken  string
}

// NewAuthService создает новый экземпляр AuthServiceImpl.
// jwtSecret и botToken должны передаваться из конфигурации.
func NewAuthService(userRepo models.UserRepository, redisRepo repository.RedisRepository, jwtSecret, botToken string) models.AuthService {
	return &AuthServiceImpl{
		userRepo:  userRepo,
		redisRepo: redisRepo,
		jwtSecret: jwtSecret,
		botToken:  botToken,
	}
}

// InitAuth инициализирует аутентификацию пользователя на основе данных от Telegram
func (s *AuthServiceImpl) InitAuth(ctx context.Context, initData string) (string, error) {
	if initData == "" {
		return "", errors.New("init data cannot be empty")
	}

	params, err := url.ParseQuery(initData)
	if err != nil {
		return "", fmt.Errorf("failed to parse init data: %w", err)
	}

	authDateStr := params.Get("auth_date")
	if !s.validateAuthTimestamp(authDateStr) {
		return "", errors.New("authentication data is outdated")
	}

	if !s.validateInitData(params, s.botToken) {
		return "", errors.New("invalid init data signature")
	}

	userStr := params.Get("user")
	if userStr == "" {
		return "", errors.New("user data not found in initData")
	}

	var userData struct {
		ID        int64  `json:"id"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name,omitempty"`
		Username  string `json:"username,omitempty"`
	}

	if err := json.Unmarshal([]byte(userStr), &userData); err != nil {
		return "", fmt.Errorf("failed to parse user data: %w", err)
	}

	telegramID := uint64(userData.ID)
	user, err := s.userRepo.GetByTelegramID(ctx, telegramID)
	if err != nil {
		// Замените models.ErrUserNotFound на вашу реальную ошибку, определенную в пакете models
		if !errors.Is(err, models.ErrUserNotFound) {
			return "", fmt.Errorf("error fetching user: %w", err)
		}
		// Пользователь не найден, создаем нового
		username := userData.Username
		if username == "" {
			username = strings.TrimSpace(fmt.Sprintf("%s %s", userData.FirstName, userData.LastName))
		}

		newUser := &models.User{
			TelegramID:  telegramID,
			Username:    username,
			FirstName:   userData.FirstName,
			LastName:    userData.LastName,
			Wallet:      "",
			BalanceTon:  0.0,
			BalanceUsdt: 0.0,
			Wins:        0,
			Losses:      0,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		if errCreate := s.userRepo.Create(ctx, newUser); errCreate != nil {
			return "", fmt.Errorf("failed to create user: %w", errCreate)
		}
		user = newUser
	} else {
		changed := false
		if user.Username != userData.Username && userData.Username != "" {
			user.Username = userData.Username
			changed = true
		}
		if user.FirstName != userData.FirstName {
			user.FirstName = userData.FirstName
			changed = true
		}
		if user.LastName != userData.LastName {
			user.LastName = userData.LastName
			changed = true
		}
		if changed {
			user.UpdatedAt = time.Now()
			if errUpdate := s.userRepo.Update(ctx, user); errUpdate != nil {
				fmt.Printf("failed to update user data: %v\n", errUpdate) // Логгируем, не блокируем вход
			}
		}
	}

	return s.GenerateToken(ctx, *user)
}

func (s *AuthServiceImpl) VerifyAuth(ctx context.Context, tokenString string) (*models.User, error) {
	return s.ValidateToken(ctx, tokenString)
}

func (s *AuthServiceImpl) GenerateToken(ctx context.Context, user models.User) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"telegram_id": user.TelegramID,
		"username":    user.Username,
		"exp":         time.Now().Add(24 * time.Hour).Unix(),
	})

	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	sessionKey := fmt.Sprintf("token:%d", user.TelegramID)
	expirationSeconds := int((24 * time.Hour).Seconds())
	if err := s.redisRepo.SetSession(ctx, sessionKey, tokenString, expirationSeconds); err != nil {
		fmt.Printf("failed to save token to redis: %v\n", err) // Логгируем
	}

	return tokenString, nil
}

func (s *AuthServiceImpl) ValidateToken(ctx context.Context, tokenString string) (*models.User, error) {
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	telegramIDFloat, ok := claims["telegram_id"].(float64)
	if !ok {
		return nil, errors.New("telegram_id not found or invalid in token claims")
	}
	telegramID := uint64(telegramIDFloat)

	sessionKey := fmt.Sprintf("token:%d", telegramID)
	savedToken, err := s.redisRepo.GetSession(ctx, sessionKey)
	// Замените repository.ErrRedisNil на вашу реальную ошибку, определенную в пакете repository
	if err != nil && !errors.Is(err, repository.ErrRedisNil) {
		fmt.Printf("redis error when validating token: %v\n", err)
	} else if err == nil && savedToken != tokenString {
		return nil, errors.New("token has been invalidated or is not the latest")
	}

	user, err := s.userRepo.GetByTelegramID(ctx, telegramID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by telegram_id: %w", err)
	}

	return user, nil
}

func (s *AuthServiceImpl) Logout(ctx context.Context, tokenString string) error {
	claims, ok := s.parseClaims(tokenString)
	if !ok {
		return errors.New("invalid token claims for logout")
	}

	telegramIDFloat, ok := claims["telegram_id"].(float64)
	if !ok {
		return errors.New("telegram_id not found in token for logout")
	}
	telegramID := uint64(telegramIDFloat)

	sessionKey := fmt.Sprintf("token:%d", telegramID)
	return s.redisRepo.DeleteSession(ctx, sessionKey)
}

func (s *AuthServiceImpl) validateInitData(params url.Values, botToken string) bool {
	hashToCompare := params.Get("hash")
	if hashToCompare == "" {
		return false
	}

	var dataCheckArr []string
	for k, v := range params {
		if k == "hash" {
			continue
		}
		dataCheckArr = append(dataCheckArr, fmt.Sprintf("%s=%s", k, v[0]))
	}
	sort.Strings(dataCheckArr)
	dataCheckString := strings.Join(dataCheckArr, "\n")

	secretKeyHMAC := hmac.New(sha256.New, []byte("WebAppData"))
	secretKeyHMAC.Write([]byte(botToken))
	secretKey := secretKeyHMAC.Sum(nil)

	calculatedHashHMAC := hmac.New(sha256.New, secretKey)
	calculatedHashHMAC.Write([]byte(dataCheckString))
	calculatedHash := hex.EncodeToString(calculatedHashHMAC.Sum(nil))

	return calculatedHash == hashToCompare
}

func (s *AuthServiceImpl) validateAuthTimestamp(authDateStr string) bool {
	if authDateStr == "" {
		return false
	}
	authDate, err := strconv.ParseInt(authDateStr, 10, 64)
	if err != nil {
		return false
	}
	return (time.Now().Unix() - authDate) < 86400 // 24 часа
}

func (s *AuthServiceImpl) parseClaims(tokenString string) (jwt.MapClaims, bool) {
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return nil, false
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	return claims, ok
}
