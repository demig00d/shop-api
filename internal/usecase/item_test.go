package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	dbmocks "shop/internal/db/mocks"
	"shop/internal/models"
	"shop/pkg/logger"
)

var (
	log = logger.NewTestLogger()
)

func TestBuyItemUseCase_BuyItem_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserDB := dbmocks.NewMockUserDBInterface(ctrl)
	mockItemDB := dbmocks.NewMockItemDBInterface(ctrl)
	mockTransactionDB := dbmocks.NewMockTransactionDBInterface(ctrl)
	uc := NewBuyItemUseCase(mockUserDB, mockItemDB, mockTransactionDB, log)

	// Данные пользователя и цена товара.
	user := &models.DBUser{ID: 1, Username: "testuser", Coins: 100}
	itemPrice := 50

	// Ожидаем получение цены товара.
	mockItemDB.
		EXPECT().
		GetItemPrice(gomock.Any(), "pen").
		Return(itemPrice, nil)

	// Ожидаем получение данных пользователя.
	mockUserDB.
		EXPECT().
		GetUserByUsername(gomock.Any(), "testuser").
		Return(user, nil)

	// Создаем мок базы данных.
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("не удалось создать sqlmock: %v", err)
	}
	defer db.Close()

	// Ожидаем транзакцию (Begin, Commit).
	sqlMock.ExpectBegin()
	sqlMock.ExpectCommit()

	// Ожидаем вызовы методов БД.
	mockTransactionDB.
		EXPECT().
		GetDB().
		Return(db)

	mockUserDB.
		EXPECT().
		UpdateUserCoins(gomock.Any(), 1, 50).
		Return(nil)

	mockUserDB.
		EXPECT().
		UpdateUserInventory(gomock.Any(), 1, "pen", 1, gomock.Any()).
		Return(nil)

	err = uc.BuyItem(context.Background(), "testuser", "pen")
	assert.NoError(t, err)

	if err := sqlMock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestBuyItemUseCase_BuyItem_ItemNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserDB := dbmocks.NewMockUserDBInterface(ctrl)
	mockItemDB := dbmocks.NewMockItemDBInterface(ctrl)
	mockTransactionDB := dbmocks.NewMockTransactionDBInterface(ctrl)
	uc := NewBuyItemUseCase(mockUserDB, mockItemDB, mockTransactionDB, log)

	// Ожидаем, что GetItemPrice вернет ошибку.
	mockItemDB.
		EXPECT().
		GetItemPrice(gomock.Any(), "nonexistent_item").
		Return(0, errors.New("item not found"))

	// Проверяем, что метод возвращает ошибку.  Используем .Contains, чтобы проверить часть сообщения об ошибке.
	err := uc.BuyItem(context.Background(), "testuser", "nonexistent_item")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), ErrNotFound.Error(), "Error message")
}

func TestBuyItemUseCase_BuyItem_NotEnoughCoins(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserDB := dbmocks.NewMockUserDBInterface(ctrl)
	mockItemDB := dbmocks.NewMockItemDBInterface(ctrl)
	mockTransactionDB := dbmocks.NewMockTransactionDBInterface(ctrl)
	uc := NewBuyItemUseCase(mockUserDB, mockItemDB, mockTransactionDB, log)

	// У пользователя недостаточно монет.
	user := &models.DBUser{ID: 1, Username: "testuser", Coins: 30}
	itemPrice := 50

	mockItemDB.
		EXPECT().
		GetItemPrice(gomock.Any(), "pen").
		Return(itemPrice, nil)

	mockUserDB.
		EXPECT().
		GetUserByUsername(gomock.Any(), "testuser").
		Return(user, nil)

	// Проверяем ошибку ErrNotEnoughCoins.
	err := uc.BuyItem(context.Background(), "testuser", "pen")
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrNotEnoughCoins))
}

func TestBuyItemUseCase_BuyItem_ItemRequired(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserDB := dbmocks.NewMockUserDBInterface(ctrl)
	mockItemDB := dbmocks.NewMockItemDBInterface(ctrl)
	mockTransactionDB := dbmocks.NewMockTransactionDBInterface(ctrl)
	uc := NewBuyItemUseCase(mockUserDB, mockItemDB, mockTransactionDB, log)

	// Проверяем ошибку ErrItemRequired, если не указано название товара.
	err := uc.BuyItem(context.Background(), "testuser", "")
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrItemRequired))
}
