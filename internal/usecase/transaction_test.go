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

func TestSendCoinUseCase_SendCoin_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserDB := dbmocks.NewMockUserDBInterface(ctrl)
	mockTransactionDB := dbmocks.NewMockTransactionDBInterface(ctrl)
	log := logger.NewTestLogger()
	uc := NewSendCoinUseCase(mockUserDB, mockTransactionDB, log)

	senderUser := &models.DBUser{ID: 1, Username: "sender", Coins: 100}
	receiverUser := &models.DBUser{ID: 2, Username: "receiver", Coins: 50}

	mockUserDB.
		EXPECT().
		GetUserByUsername(gomock.Any(), "sender").
		Return(senderUser, nil)
	mockUserDB.
		EXPECT().
		GetUserByUsername(gomock.Any(), "receiver").
		Return(receiverUser, nil)

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
		UpdateUserCoins(gomock.Any(), 1, 50). // У отправителя вычитаются монеты.
		Return(nil)
	mockUserDB.
		EXPECT().
		UpdateUserCoins(gomock.Any(), 2, 100). // Получателю добавляются монеты.
		Return(nil)
	mockTransactionDB.
		EXPECT().
		RecordTransaction(gomock.Any(), 1, 2, 50, gomock.Any()). // Запись транзакции.
		Return(nil)

	// Вызываем тестируемый метод.
	err = uc.SendCoin(context.Background(), "sender", "receiver", 50)
	assert.NoError(t, err)

	// Проверяем, что все ожидания sqlmock были удовлетворены.
	if err := sqlMock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestSendCoinUseCase_SendCoin_InsufficientFunds(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserDB := dbmocks.NewMockUserDBInterface(ctrl)
	mockTransactionDB := dbmocks.NewMockTransactionDBInterface(ctrl)
	log := logger.NewTestLogger()
	uc := NewSendCoinUseCase(mockUserDB, mockTransactionDB, log)

	// У отправителя недостаточно монет.
	senderUser := &models.DBUser{ID: 1, Username: "sender", Coins: 30}
	receiverUser := &models.DBUser{ID: 2, Username: "receiver", Coins: 50}

	mockUserDB.
		EXPECT().
		GetUserByUsername(gomock.Any(), "sender").
		Return(senderUser, nil)
	mockUserDB.
		EXPECT().
		GetUserByUsername(gomock.Any(), "receiver").
		Return(receiverUser, nil)

		// Проверяем ошибку ErrInsufficientFunds
	err := uc.SendCoin(context.Background(), "sender", "receiver", 50)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrInsufficientFunds))
}

func TestSendCoinUseCase_SendCoin_SelfTransfer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserDB := dbmocks.NewMockUserDBInterface(ctrl)
	mockTransactionDB := dbmocks.NewMockTransactionDBInterface(ctrl)
	log := logger.NewTestLogger()
	uc := NewSendCoinUseCase(mockUserDB, mockTransactionDB, log)

	senderUser := &models.DBUser{ID: 1, Username: "sender", Coins: 100}

	// Отправитель и получатель - один и тот же пользователь.
	mockUserDB.
		EXPECT().
		GetUserByUsername(gomock.Any(), "sender").
		Return(senderUser, nil)
	mockUserDB.
		EXPECT().
		GetUserByUsername(gomock.Any(), "sender").
		Return(senderUser, nil)

		// Проверяем ошибку ErrSelfTransfer.
	err := uc.SendCoin(context.Background(), "sender", "sender", 50)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrSelfTransfer))
}

func TestSendCoinUseCase_SendCoin_ReceiverNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserDB := dbmocks.NewMockUserDBInterface(ctrl)
	mockTransactionDB := dbmocks.NewMockTransactionDBInterface(ctrl)
	log := logger.NewTestLogger()
	uc := NewSendCoinUseCase(mockUserDB, mockTransactionDB, log)

	senderUser := &models.DBUser{ID: 1, Username: "sender", Coins: 100}

	// Получатель не найден.
	mockUserDB.
		EXPECT().
		GetUserByUsername(gomock.Any(), "sender").
		Return(senderUser, nil)
	mockUserDB.
		EXPECT().
		GetUserByUsername(gomock.Any(), "receiver").
		Return(nil, nil)

		// Проверяем ошибку ErrReceiverNotFound
	err := uc.SendCoin(context.Background(), "sender", "receiver", 50)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrReceiverNotFound))
}

func TestSendCoinUseCase_SendCoin_InvalidAmount(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserDB := dbmocks.NewMockUserDBInterface(ctrl)
	mockTransactionDB := dbmocks.NewMockTransactionDBInterface(ctrl)
	log := logger.NewTestLogger()
	uc := NewSendCoinUseCase(mockUserDB, mockTransactionDB, log)

	// Неверная сумма (0).
	err := uc.SendCoin(context.Background(), "sender", "receiver", 0)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrInvalidAmount))
}
