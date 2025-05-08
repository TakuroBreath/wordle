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

	"github.com/TakuroBreath/wordle/internal/logger"
	"github.com/TakuroBreath/wordle/internal/models"
	"github.com/TakuroBreath/wordle/internal/repository"
	"github.com/dgrijalva/jwt-go"
	initdata "github.com/telegram-mini-apps/init-data-golang"
	"go.uber.org/zap"
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
	logger    *zap.Logger
}

// NewAuthService создает новый экземпляр AuthServiceImpl.
// jwtSecret и botToken должны передаваться из конфигурации.
func NewAuthService(userRepo models.UserRepository, redisRepo repository.RedisRepository, jwtSecret, botToken string) models.AuthService {
	log := logger.GetLogger(zap.String("service", "auth"))
	log.Info("Creating new AuthService",
		zap.String("jwtSecret_length", fmt.Sprintf("%d", len(jwtSecret))),
		zap.String("botToken_length", fmt.Sprintf("%d", len(botToken))))

	return &AuthServiceImpl{
		userRepo:  userRepo,
		redisRepo: redisRepo,
		jwtSecret: jwtSecret,
		botToken:  botToken,
		logger:    log,
	}
}

// InitAuth инициализирует аутентификацию пользователя на основе данных от Telegram Mini App
func (s *AuthServiceImpl) InitAuth(ctx context.Context, initDataStr string) (string, error) {
	log := s.logger.With(zap.String("method", "InitAuth"))
	log.Info("Initializing authentication from Telegram Mini App data",
		zap.Int("init_data_length", len(initDataStr)))

	// Валидируем данные Telegram Mini App
	log.Debug("Validating Telegram Mini App data")
	if err := initdata.Validate(initDataStr, s.botToken, time.Hour); err != nil {
		log.Error("Invalid init data", zap.Error(err),
			zap.String("botToken_length", fmt.Sprintf("%d", len(s.botToken))))
		return "", fmt.Errorf("invalid init data: %w", err)
	}

	// Парсим инициализационные данные
	log.Debug("Parsing initialization data")
	data, err := initdata.Parse(initDataStr)
	if err != nil {
		log.Error("Failed to parse init data", zap.Error(err))
		return "", fmt.Errorf("failed to parse init data: %w", err)
	}

	// Получаем данные пользователя из инициализационных данных
	if data.User.ID == 0 {
		log.Error("User data not found in init data")
		return "", fmt.Errorf("user data not found in init data")
	}

	// Преобразуем строковый ID в uint64
	telegramID := uint64(data.User.ID)
	if telegramID == 0 {
		log.Error("Invalid telegram ID", zap.Int64("telegram_id_int64", int64(data.User.ID)))
		return "", fmt.Errorf("invalid telegram ID")
	}

	log.Info("Getting user by Telegram ID", zap.Uint64("telegram_id", telegramID))

	// Проверяем, существует ли пользователь
	user, err := s.userRepo.GetByTelegramID(ctx, telegramID)
	if err != nil {
		// Если пользователь не найден, создаем нового
		if errors.Is(err, models.ErrUserNotFound) {
			log.Info("User not found, creating new user",
				zap.Uint64("telegram_id", telegramID),
				zap.String("username", data.User.Username))

			newUser := &models.User{
				TelegramID: telegramID,
				Username:   data.User.Username,
				FirstName:  data.User.FirstName,
				LastName:   data.User.LastName,
			}

			log.Debug("Creating new user",
				zap.Uint64("telegram_id", newUser.TelegramID),
				zap.String("username", newUser.Username),
				zap.String("first_name", newUser.FirstName),
				zap.String("last_name", newUser.LastName))

			err = s.userRepo.Create(ctx, newUser)
			if err != nil {
				log.Error("Failed to create user", zap.Error(err),
					zap.Uint64("telegram_id", telegramID))
				return "", fmt.Errorf("failed to create user: %w", err)
			}
			user = newUser
			log.Info("User created successfully", zap.Uint64("telegram_id", telegramID))
		} else {
			log.Error("Failed to get user", zap.Error(err),
				zap.Uint64("telegram_id", telegramID))
			return "", fmt.Errorf("failed to get user: %w", err)
		}
	} else {
		log.Info("User found",
			zap.Uint64("telegram_id", user.TelegramID),
			zap.String("username", user.Username))
	}

	// Обновляем информацию о пользователе, если она изменилась
	if user.Username != data.User.Username ||
		user.FirstName != data.User.FirstName ||
		user.LastName != data.User.LastName {

		log.Info("Updating user information",
			zap.Uint64("telegram_id", user.TelegramID),
			zap.String("old_username", user.Username),
			zap.String("new_username", data.User.Username))

		user.Username = data.User.Username
		user.FirstName = data.User.FirstName
		user.LastName = data.User.LastName

		err = s.userRepo.Update(ctx, user)
		if err != nil {
			log.Error("Failed to update user", zap.Error(err),
				zap.Uint64("telegram_id", user.TelegramID))
			return "", fmt.Errorf("failed to update user: %w", err)
		}
		log.Info("User updated successfully", zap.Uint64("telegram_id", user.TelegramID))
	}

	// Генерируем JWT токен для пользователя
	log.Debug("Generating JWT token", zap.Uint64("telegram_id", user.TelegramID))
	token, err := s.GenerateToken(ctx, *user)
	if err != nil {
		log.Error("Failed to generate token", zap.Error(err),
			zap.Uint64("telegram_id", user.TelegramID))
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	log.Info("Token generated successfully",
		zap.Uint64("telegram_id", user.TelegramID),
		zap.Int("token_length", len(token)))
	return token, nil
}

// VerifyAuth проверяет аутентификацию по токену
func (s *AuthServiceImpl) VerifyAuth(ctx context.Context, tokenString string) (*models.User, error) {
	log := s.logger.With(zap.String("method", "VerifyAuth"))
	log.Info("Verifying authentication token", zap.Int("token_length", len(tokenString)))

	// Валидируем JWT токен
	user, err := s.ValidateToken(ctx, tokenString)
	if err != nil {
		log.Error("Token validation failed", zap.Error(err))
		return nil, err
	}

	log.Info("Token verification successful", zap.Uint64("telegram_id", user.TelegramID))
	return user, nil
}

// GenerateToken создает JWT токен для пользователя
func (s *AuthServiceImpl) GenerateToken(ctx context.Context, user models.User) (string, error) {
	log := s.logger.With(zap.String("method", "GenerateToken"),
		zap.Uint64("telegram_id", user.TelegramID))
	log.Debug("Generating token")

	expirationTime := time.Now().Add(24 * time.Hour)
	log.Debug("Token expiration set", zap.Time("expires_at", expirationTime))

	claims := &Claims{
		TelegramID: user.TelegramID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		log.Error("Failed to sign token", zap.Error(err))
		return "", err
	}

	sessionKey := fmt.Sprintf("token:%d", user.TelegramID)
	expirationSeconds := int((24 * time.Hour).Seconds())

	log.Debug("Saving token to Redis",
		zap.String("session_key", sessionKey),
		zap.Int("expiration_seconds", expirationSeconds))

	if err := s.redisRepo.SetSession(ctx, sessionKey, tokenString, expirationSeconds); err != nil {
		log.Error("Failed to save token to Redis", zap.Error(err),
			zap.String("session_key", sessionKey))
	} else {
		log.Debug("Token saved to Redis successfully")
	}

	log.Info("Token generated successfully")
	return tokenString, nil
}

// ValidateToken проверяет валидность JWT токена
func (s *AuthServiceImpl) ValidateToken(ctx context.Context, tokenString string) (*models.User, error) {
	log := s.logger.With(zap.String("method", "ValidateToken"))
	log.Debug("Validating JWT token")

	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			log.Error("Invalid token signature")
			return nil, errors.New(ErrAuthInvalidToken)
		}
		log.Error("Failed to parse token", zap.Error(err))
		return nil, err
	}

	if !token.Valid {
		log.Error("Token is invalid")
		return nil, errors.New(ErrAuthInvalidToken)
	}

	// Проверяем срок действия токена
	if claims.ExpiresAt < time.Now().Unix() {
		log.Error("Token has expired",
			zap.Int64("expires_at", claims.ExpiresAt),
			zap.Int64("now", time.Now().Unix()))
		return nil, errors.New(ErrAuthTokenExpired)
	}

	log.Debug("Checking token in Redis", zap.Uint64("telegram_id", claims.TelegramID))
	sessionKey := fmt.Sprintf("token:%d", claims.TelegramID)
	savedToken, err := s.redisRepo.GetSession(ctx, sessionKey)

	// Замените repository.ErrRedisNil на вашу реальную ошибку, определенную в пакете repository
	if err != nil && !errors.Is(err, repository.ErrRedisNil) {
		log.Error("Redis error when validating token", zap.Error(err),
			zap.String("session_key", sessionKey))
	} else if err == nil && savedToken != tokenString {
		log.Error("Token has been invalidated or is not the latest",
			zap.Uint64("telegram_id", claims.TelegramID))
		return nil, errors.New("token has been invalidated or is not the latest")
	}

	// Получаем пользователя по telegram ID из токена
	log.Debug("Getting user by Telegram ID from token",
		zap.Uint64("telegram_id", claims.TelegramID))
	user, err := s.userRepo.GetByTelegramID(ctx, claims.TelegramID)
	if err != nil {
		log.Error("User not found", zap.Error(err),
			zap.Uint64("telegram_id", claims.TelegramID))
		return nil, errors.New(ErrAuthUserNotFound)
	}

	log.Info("Token validated successfully", zap.Uint64("telegram_id", user.TelegramID))
	return user, nil
}

// Logout выполняет выход пользователя
func (s *AuthServiceImpl) Logout(ctx context.Context, tokenString string) error {
	log := s.logger.With(zap.String("method", "Logout"))
	log.Info("Processing logout request")

	// В JWT нет стандартного механизма отзыва токенов
	// Для простоты просто возвращаем nil
	log.Info("Logout completed")
	return nil
}

func (s *AuthServiceImpl) validateInitData(params url.Values, botToken string) bool {
	log := s.logger.With(zap.String("method", "validateInitData"))
	log.Debug("Validating init data")

	hashToCompare := params.Get("hash")
	if hashToCompare == "" {
		log.Error("Hash parameter is missing")
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

	isValid := calculatedHash == hashToCompare
	if !isValid {
		log.Error("Hash validation failed",
			zap.String("expected", calculatedHash),
			zap.String("actual", hashToCompare))
	} else {
		log.Debug("Hash validation successful")
	}

	return isValid
}

func (s *AuthServiceImpl) validateAuthTimestamp(authDateStr string) bool {
	log := s.logger.With(zap.String("method", "validateAuthTimestamp"))
	log.Debug("Validating auth timestamp", zap.String("auth_date", authDateStr))

	if authDateStr == "" {
		log.Error("Auth date is missing")
		return false
	}

	authDate, err := strconv.ParseInt(authDateStr, 10, 64)
	if err != nil {
		log.Error("Failed to parse auth date", zap.Error(err),
			zap.String("auth_date", authDateStr))
		return false
	}

	now := time.Now().Unix()
	age := now - authDate
	isValid := age < 86400 // 24 часа

	if !isValid {
		log.Error("Auth date is too old",
			zap.Int64("auth_date", authDate),
			zap.Int64("now", now),
			zap.Int64("age_seconds", age),
			zap.Int64("max_age", 86400))
	} else {
		log.Debug("Auth timestamp is valid",
			zap.Int64("auth_date", authDate),
			zap.Int64("now", now),
			zap.Int64("age_seconds", age))
	}

	return isValid
}

func (s *AuthServiceImpl) parseClaims(tokenString string) (jwt.MapClaims, bool) {
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return nil, false
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	return claims, ok
}
