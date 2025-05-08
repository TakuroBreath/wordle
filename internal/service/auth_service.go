package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/TakuroBreath/wordle/internal/models"
	"github.com/TakuroBreath/wordle/internal/repository"
	"github.com/dgrijalva/jwt-go"
	initdata "github.com/telegram-mini-apps/init-data-golang"
)

const (
	// ErrAuthInvalidToken определяет ошибку невалидного токена
	ErrAuthInvalidToken = "invalid auth token"
	// ErrAuthTokenExpired определяет ошибку истекшего токена
	ErrAuthTokenExpired = "auth token expired"
	// ErrAuthUserNotFound определяет ошибку отсутствующего пользователя
	ErrAuthUserNotFound = "user not found"
)

// Claims представляет структуру JWT токена
type Claims struct {
	TelegramID uint64 `json:"telegram_id"`
	jwt.StandardClaims
}

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

// InitAuth инициализирует аутентификацию пользователя на основе данных от Telegram Mini App
func (s *AuthServiceImpl) InitAuth(ctx context.Context, initDataStr string) (string, error) {
	// Валидируем данные Telegram Mini App
	if err := initdata.Validate(initDataStr, s.botToken, time.Hour); err != nil {
		return "", fmt.Errorf("invalid init data: %w", err)
	}

	// Парсим инициализационные данные
	data, err := initdata.Parse(initDataStr)
	if err != nil {
		return "", fmt.Errorf("failed to parse init data: %w", err)
	}

	// Получаем данные пользователя из инициализационных данных
	if data.User.ID == 0 {
		return "", fmt.Errorf("user data not found in init data")
	}

	// Преобразуем строковый ID в uint64
	telegramID := uint64(data.User.ID)
	if telegramID == 0 {
		return "", fmt.Errorf("invalid telegram ID")
	}

	// Проверяем, существует ли пользователь
	user, err := s.userRepo.GetByTelegramID(ctx, telegramID)
	if err != nil {
		// Если пользователь не найден, создаем нового
		if errors.Is(err, models.ErrUserNotFound) {
			newUser := &models.User{
				TelegramID: telegramID,
				Username:   data.User.Username,
				FirstName:  data.User.FirstName,
				LastName:   data.User.LastName,
			}
			err = s.userRepo.Create(ctx, newUser)
			if err != nil {
				return "", fmt.Errorf("failed to create user: %w", err)
			}
			user = newUser
		} else {
			return "", fmt.Errorf("failed to get user: %w", err)
		}
	}

	// Обновляем информацию о пользователе, если она изменилась
	if user.Username != data.User.Username ||
		user.FirstName != data.User.FirstName ||
		user.LastName != data.User.LastName {

		user.Username = data.User.Username
		user.FirstName = data.User.FirstName
		user.LastName = data.User.LastName

		err = s.userRepo.Update(ctx, user)
		if err != nil {
			return "", fmt.Errorf("failed to update user: %w", err)
		}
	}

	// Генерируем JWT токен для пользователя
	token, err := s.GenerateToken(ctx, *user)
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	return token, nil
}

// VerifyAuth проверяет аутентификацию по токену
func (s *AuthServiceImpl) VerifyAuth(ctx context.Context, tokenString string) (*models.User, error) {
	// Валидируем JWT токен
	user, err := s.ValidateToken(ctx, tokenString)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// GenerateToken создает JWT токен для пользователя
func (s *AuthServiceImpl) GenerateToken(ctx context.Context, user models.User) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)

	claims := &Claims{
		TelegramID: user.TelegramID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", err
	}

	sessionKey := fmt.Sprintf("token:%d", user.TelegramID)
	expirationSeconds := int((24 * time.Hour).Seconds())
	if err := s.redisRepo.SetSession(ctx, sessionKey, tokenString, expirationSeconds); err != nil {
		fmt.Printf("failed to save token to redis: %v\n", err) // Логгируем
	}

	return tokenString, nil
}

// ValidateToken проверяет валидность JWT токена
func (s *AuthServiceImpl) ValidateToken(ctx context.Context, tokenString string) (*models.User, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			return nil, errors.New(ErrAuthInvalidToken)
		}
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New(ErrAuthInvalidToken)
	}

	// Проверяем срок действия токена
	if claims.ExpiresAt < time.Now().Unix() {
		return nil, errors.New(ErrAuthTokenExpired)
	}

	sessionKey := fmt.Sprintf("token:%d", claims.TelegramID)
	savedToken, err := s.redisRepo.GetSession(ctx, sessionKey)
	// Замените repository.ErrRedisNil на вашу реальную ошибку, определенную в пакете repository
	if err != nil && !errors.Is(err, repository.ErrRedisNil) {
		fmt.Printf("redis error when validating token: %v\n", err)
	} else if err == nil && savedToken != tokenString {
		return nil, errors.New("token has been invalidated or is not the latest")
	}

	// Получаем пользователя по telegram ID из токена
	user, err := s.userRepo.GetByTelegramID(ctx, claims.TelegramID)
	if err != nil {
		return nil, errors.New(ErrAuthUserNotFound)
	}

	return user, nil
}

// Logout выполняет выход пользователя
func (s *AuthServiceImpl) Logout(ctx context.Context, tokenString string) error {
	// В JWT нет стандартного механизма отзыва токенов
	// Здесь можно добавить токен в blacklist или использовать Redis для хранения отозванных токенов
	// Для простоты просто возвращаем nil
	return nil
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
