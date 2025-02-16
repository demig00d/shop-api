package usecase

import (
	"context"
	"errors"
	"testing"

	dbmocks "shop/internal/db/mocks"
	"shop/internal/models"
	"shop/pkg/logger"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestUserUseCase_GetUserInfo_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserDB := dbmocks.NewMockUserDBInterface(ctrl)
	mockTransactionDB := dbmocks.NewMockTransactionDBInterface(ctrl)
	log := logger.NewTestLogger()
	uc := NewUserInfoUseCase("secret", mockUserDB, mockTransactionDB, log)

	// Ожидаемый ответ.
	expectedResponse := &models.InfoResponse{
		Coins: 100,
		Inventory: []models.InventoryItem{
			{Type: "pen", Quantity: 1},
		},
		CoinHistory: models.CoinHistory{
			Received: []models.Transaction{},
			Sent:     []models.Transaction{},
		},
	}

	// Ожидаемые данные из БД.
	expectedUser := &models.DBUser{ID: 1, Username: "testuser", Coins: 100}
	expectedInventory := []models.DBInventoryItem{{ItemType: "pen", Quantity: 1}}
	expectedHistory := &models.CoinHistory{Received: []models.Transaction{}, Sent: []models.Transaction{}}

	// Ожидаемые вызовы методов БД.
	mockUserDB.EXPECT().GetUserByUsername(gomock.Any(), "testuser").Return(expectedUser, nil)
	mockUserDB.EXPECT().GetUserInventory(gomock.Any(), 1).Return(expectedInventory, nil)
	mockTransactionDB.EXPECT().GetCoinHistory(gomock.Any(), 1).Return(expectedHistory, nil)

	// Вызываем тестируемый метод.
	response, err := uc.GetUserInfo(context.Background(), "testuser")
	assert.NoError(t, err)
	assert.Equal(t, expectedResponse, response)
}

func TestUserUseCase_GetUserInfo_UserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserDB := dbmocks.NewMockUserDBInterface(ctrl)
	mockTransactionDB := dbmocks.NewMockTransactionDBInterface(ctrl)
	log := logger.NewTestLogger()
	uc := NewUserInfoUseCase("secret", mockUserDB, mockTransactionDB, log)

	// Ожидаем, что GetUserByUsername вернет nil, nil (пользователь не найден).
	mockUserDB.EXPECT().GetUserByUsername(gomock.Any(), "testuser").Return(nil, nil)

	// Проверяем, что возвращается ошибка ErrUserNotFound.
	response, err := uc.GetUserInfo(context.Background(), "testuser")
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.True(t, errors.Is(err, ErrUserNotFound))
}

func TestUserUseCase_Auth_Success_ExistingUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserDB := dbmocks.NewMockUserDBInterface(ctrl)
	mockTransactionDB := dbmocks.NewMockTransactionDBInterface(ctrl)
	log := logger.NewTestLogger()
	uc := NewUserInfoUseCase("secret", mockUserDB, mockTransactionDB, log)

	// Хэш пароля.
	validPasswordHashBytes, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	validPasswordHash := string(validPasswordHashBytes)

	// Ожидаемый пользователь.
	expectedUser := &models.DBUser{
		ID:           1,
		Username:     "testuser",
		PasswordHash: validPasswordHash,
		Coins:        100,
	}
	// Ожидаем вызов GetUserByUsername.
	mockUserDB.EXPECT().GetUserByUsername(gomock.Any(), "testuser").Return(expectedUser, nil)

	// Вызываем Auth и проверяем, что токен сгенерирован.
	token, err := uc.Auth(context.Background(), "testuser", "password")
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// Проверяем токен.
	username, verifyErr := uc.VerifyJWTToken(token)
	assert.NoError(t, verifyErr)
	assert.Equal(t, "testuser", username)
}

func TestUserUseCase_Auth_Success_NewUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserDB := dbmocks.NewMockUserDBInterface(ctrl)
	mockTransactionDB := dbmocks.NewMockTransactionDBInterface(ctrl)
	log := logger.NewTestLogger()
	uc := NewUserInfoUseCase("secret", mockUserDB, mockTransactionDB, log)

	// Ожидаем вызов GetUserByUsername, который вернет nil, nil (пользователь не найден)
	// Ожидаем вызов CreateUser для создания пользователя.
	// Ожидаем повторный вызов GetUserByUsername, который вернет уже созданного пользователя
	// Ожидаем вызов SetInitialCoins для установки начального количества монет.
	mockUserDB.EXPECT().GetUserByUsername(gomock.Any(), "newuser").Return(nil, nil)
	mockUserDB.EXPECT().CreateUser(gomock.Any(), "newuser", gomock.Any()).Return(nil)
	mockUserDB.EXPECT().GetUserByUsername(gomock.Any(), "newuser").Return(&models.DBUser{ID: 2, Username: "newuser", Coins: 0}, nil)
	mockUserDB.EXPECT().SetInitialCoins(gomock.Any(), 2, 1000).Return(nil)

	// Вызываем Auth, проверяем, что токен сгенерирован.
	token, err := uc.Auth(context.Background(), "newuser", "password")
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// Проверяем токен.
	username, verifyErr := uc.VerifyJWTToken(token)
	assert.NoError(t, verifyErr)
	assert.Equal(t, "newuser", username)
}

func TestUserUseCase_Auth_InvalidPassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserDB := dbmocks.NewMockUserDBInterface(ctrl)
	mockTransactionDB := dbmocks.NewMockTransactionDBInterface(ctrl)
	log := logger.NewTestLogger()
	uc := NewUserInfoUseCase("secret", mockUserDB, mockTransactionDB, log)

	validPasswordHashBytes, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	validPasswordHash := string(validPasswordHashBytes)

	expectedUser := &models.DBUser{
		ID:           1,
		Username:     "testuser",
		PasswordHash: validPasswordHash,
		Coins:        100,
	}
	mockUserDB.EXPECT().GetUserByUsername(gomock.Any(), "testuser").Return(expectedUser, nil)

	// Проверяем, что возвращается ошибка ErrInvalidPassword, если пароль неверный
	token, err := uc.Auth(context.Background(), "testuser", "wrong_password")
	assert.Error(t, err)
	assert.Empty(t, token)
	assert.True(t, errors.Is(err, ErrInvalidPassword))
}

func TestUserUseCase_GenerateJWTToken_VerifyJWTToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserDB := dbmocks.NewMockUserDBInterface(ctrl)
	mockTransactionDB := dbmocks.NewMockTransactionDBInterface(ctrl)
	log := logger.NewTestLogger()
	uc := NewUserInfoUseCase("secret", mockUserDB, mockTransactionDB, log)

	// Генерация и проверка токена.
	username := "testuser"
	token, err := uc.GenerateJWTToken(username)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	verifiedUsername, err := uc.VerifyJWTToken(token)
	assert.NoError(t, err)
	assert.Equal(t, username, verifiedUsername)
}
