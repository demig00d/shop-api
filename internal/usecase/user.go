// ./internal/usecase/user.go
package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"shop/internal/db"
	"shop/internal/models"
	"shop/pkg/logger"
)

// Ошибки
var (
	ErrInvalidRequest  = errors.New("неверный запрос")
	ErrNotFound        = errors.New("не найдено")
	ErrUnauthorized    = errors.New("не авторизован")
	ErrUserNotFound    = fmt.Errorf("%w: пользователь не найден", ErrNotFound)
	ErrInvalidPassword = fmt.Errorf("%w: неверный пароль", ErrUnauthorized)
)

// UserUseCaseInterface интерфейс для use case'ов информации о пользователе и аутентификации.
type UserUseCaseInterface interface {
	GetUserInfo(ctx context.Context, username string) (*models.InfoResponse, error)
	Auth(ctx context.Context, username string, password string) (string, error)
	GenerateJWTToken(username string) (string, error)
	VerifyJWTToken(tokenString string) (string, error)
}

// UserUseCase реализует UserInfoUseCaseInterface.
type UserUseCase struct {
	userDB        db.UserDBInterface
	transactionDB db.TransactionDBInterface
	jwtSecret     []byte
	log           *logger.Logger
}

// NewUserInfoUseCase создает новый UserUseCase.
func NewUserInfoUseCase(jwtSecretString string, userDB db.UserDBInterface, transactionDB db.TransactionDBInterface, log *logger.Logger) *UserUseCase {
	return &UserUseCase{
		userDB:        userDB,
		transactionDB: transactionDB,
		jwtSecret:     []byte(jwtSecretString),
		log:           log,
	}
}

// GetUserInfo получает информацию о пользователе.
func (uc *UserUseCase) GetUserInfo(ctx context.Context, username string) (*models.InfoResponse, error) {
	uc.log.Debug("GetUserInfo", "username", username)

	user, err := uc.userDB.GetUserByUsername(ctx, username)
	if err != nil {
		uc.log.Error("Ошибка GetUserByUsername в GetUserInfo", "username", username, "error", err)
		return nil, fmt.Errorf("ошибка при получении пользователя: %w", err)
	}
	if user == nil {
		uc.log.Warn("Пользователь не найден в GetUserInfo", "username", username)
		return nil, ErrUserNotFound
	}
	uc.log.Debug("Пользователь найден", "username", username, "userID", user.ID)

	inventoryDB, err := uc.userDB.GetUserInventory(ctx, user.ID)
	if err != nil {
		uc.log.Error("Ошибка GetUserInventory в GetUserInfo", "userID", user.ID, "error", err)
		return nil, fmt.Errorf("ошибка при получении инвентаря пользователя: %w", err)
	}

	inventory := []models.InventoryItem{}
	for _, item := range inventoryDB {
		inventory = append(inventory, models.InventoryItem{Type: item.ItemType, Quantity: item.Quantity})
	}

	history, err := uc.transactionDB.GetCoinHistory(ctx, user.ID)
	if err != nil {
		uc.log.Error("Ошибка GetCoinHistory в GetUserInfo", "userID", user.ID, "error", err)
		return nil, fmt.Errorf("ошибка при получении истории транзакций: %w", err)
	}

	response := &models.InfoResponse{
		Coins:       user.Coins,
		Inventory:   inventory,
		CoinHistory: *history,
	}
	return response, nil
}

// Auth аутентифицирует пользователя и возвращает JWT токен.
func (uc *UserUseCase) Auth(ctx context.Context, username string, password string) (string, error) {
	uc.log.Debug("Auth", "username", username)

	user, err := uc.userDB.GetUserByUsername(ctx, username)
	if err != nil {
		uc.log.Error("Ошибка GetUserByUsername в Auth", "username", username, "error", err)
		return "", fmt.Errorf("ошибка сервера при поиске пользователя: %w", err)
	}

	uc.log.Debug("Пользователь после GetUserByUsername", "username", username, "user", user)

	if user == nil {
		// Пользователь не найден, создаем нового (логика регистрации).
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			uc.log.Error("Ошибка bcrypt.GenerateFromPassword в Auth", "username", username, "error", err)
			return "", fmt.Errorf("ошибка сервера при хешировании пароля: %w", err)
		}
		err = uc.userDB.CreateUser(ctx, username, string(hashedPassword))
		if err != nil {
			uc.log.Error("Ошибка CreateUser в Auth", "username", username, "error", err)
			return "", fmt.Errorf("ошибка сервера при создании пользователя: %w", err)
		}
		user, err = uc.userDB.GetUserByUsername(ctx, username)
		if err != nil {
			uc.log.Error("Ошибка GetUserByUsername после создания в Auth", "username", username, "error", err)
			return "", fmt.Errorf("ошибка сервера после создания пользователя: %w", err)
		}
		// Устанавливаем начальное количество монет для нового пользователя.
		err = uc.userDB.SetInitialCoins(ctx, user.ID, 1000)
		if err != nil {
			uc.log.Error("Ошибка SetInitialCoins в Auth", "userID", user.ID, "error", err)
			return "", fmt.Errorf("ошибка сервера при установке начальных монет: %w", err)
		}
	} else {
		// Пользователь существует, проверяем пароль.
		err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
		if err != nil {
			uc.log.Error("Ошибка bcrypt.CompareHashAndPassword", "username", username, "error", err)
			return "", ErrInvalidPassword
		}
	}

	token, err := uc.GenerateJWTToken(username)
	if err != nil {
		uc.log.Error("Ошибка GenerateJWTToken в Auth", "username", username, "error", err)
		return "", fmt.Errorf("ошибка сервера при генерации токена: %w", err)
	}
	return token, nil
}

// GenerateJWTToken генерирует JWT токен для заданного имени пользователя.
func (uc *UserUseCase) GenerateJWTToken(username string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": username,
	})

	tokenString, err := token.SignedString(uc.jwtSecret)
	if err != nil {
		return "", fmt.Errorf("ошибка подписи токена: %w", err)
	}
	return tokenString, nil
}

// VerifyJWTToken проверяет JWT токен и возвращает имя пользователя, если токен действителен.
func (uc *UserUseCase) VerifyJWTToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("неожиданный метод подписи: %v", token.Header["alg"])
		}
		return uc.jwtSecret, nil
	})

	if err != nil {
		return "", fmt.Errorf("ошибка парсинга токена: %w", err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		username, ok := claims["username"].(string)
		if !ok {
			return "", fmt.Errorf("неверное имя пользователя в токене")
		}
		return username, nil
	}
	return "", fmt.Errorf("неверный токен")
}
