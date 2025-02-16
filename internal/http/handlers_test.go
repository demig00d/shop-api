package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"shop/internal/models"
	"shop/internal/usecase"
	ucmocks "shop/internal/usecase/mocks"
	"shop/pkg/logger"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

var (
	// Моки usecase'ов
	mockUserUseCase     *ucmocks.MockUserUseCaseInterface
	mockSendCoinUseCase *ucmocks.MockSendCoinUseCaseInterface
	mockBuyItemUseCase  *ucmocks.MockBuyItemUseCaseInterface
	// Обработчик API
	handler *ApiHandler
	// Контроллер для моков
	ctrl *gomock.Controller
	// Логгер
	log *logger.Logger
)

// Функция установки окружения для тестирования обработчиков.
func setupHandlerTest(t *testing.T) {
	ctrl = gomock.NewController(t)
	log = logger.NewTestLogger()

	mockUserUseCase = ucmocks.NewMockUserUseCaseInterface(ctrl)
	mockSendCoinUseCase = ucmocks.NewMockSendCoinUseCaseInterface(ctrl)
	mockBuyItemUseCase = ucmocks.NewMockBuyItemUseCaseInterface(ctrl)
	handler = NewApiHandler(mockUserUseCase, mockSendCoinUseCase, mockBuyItemUseCase, log)
}

// Функция завершения окружения для тестирования обработчиков.
func teardownHandlerTest() {
	ctrl.Finish()
}

func TestApiHandler_handleInfo_Success(t *testing.T) {
	setupHandlerTest(t)
	defer teardownHandlerTest()

	// Ожидаемый ответ.
	expectedResponse := &models.InfoResponse{
		Coins:     100,
		Inventory: []models.InventoryItem{{Type: "sword", Quantity: 1}},
		CoinHistory: models.CoinHistory{
			Received: []models.Transaction{},
			Sent:     []models.Transaction{},
		},
	}

	// Ожидаем вызов метода GetUserInfo usecase'а с любым контекстом и именем пользователя "testuser".
	mockUserUseCase.EXPECT().GetUserInfo(gomock.Any(), "testuser").Return(expectedResponse, nil)

	// Создаем тестовый запрос.
	req := httptest.NewRequest("GET", "/api/info", nil)
	// Добавляем имя пользователя в контекст запроса.
	reqCtx := context.WithValue(req.Context(), "username", "testuser")
	req = req.WithContext(reqCtx)
	// Создаем ResponseRecorder для записи ответа.
	recorder := httptest.NewRecorder()

	// Вызываем тестируемый обработчик.
	handler.handleInfo(recorder, req)

	// Проверяем код статуса и Content-Type.
	assert.Equal(t, http.StatusOK, recorder.Code, "Код статуса должен быть 200 OK")
	assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"), "Content-Type должен быть application/json")

	// Декодируем JSON-ответ и проверяем его содержимое.
	var actualResponse models.InfoResponse
	err := json.NewDecoder(recorder.Body).Decode(&actualResponse)
	assert.NoError(t, err, "Декодирование JSON ответа не должно завершаться с ошибкой")
	assert.Equal(t, expectedResponse, &actualResponse, "Тело ответа должно соответствовать ожидаемому")
}

func TestApiHandler_handleInfo_UserNotFound(t *testing.T) {
	setupHandlerTest(t)
	defer teardownHandlerTest()

	// Ожидаем, что GetUserInfo вернет ошибку ErrUserNotFound.
	mockUserUseCase.EXPECT().GetUserInfo(gomock.Any(), "testuser").Return(nil, usecase.ErrUserNotFound)

	req := httptest.NewRequest("GET", "/api/info", nil)
	reqCtx := context.WithValue(req.Context(), "username", "testuser")
	req = req.WithContext(reqCtx)
	recorder := httptest.NewRecorder()

	handler.handleInfo(recorder, req)

	// Проверяем код статуса (404).
	assert.Equal(t, http.StatusNotFound, recorder.Code, "Код статуса должен быть 404 Not Found")

	// Проверяем сообщение об ошибке.
	var errorResponse models.ErrorResponse
	err := json.NewDecoder(recorder.Body).Decode(&errorResponse)
	assert.NoError(t, err, "Декодирование JSON ответа об ошибке не должно завершаться с ошибкой")
	assert.Contains(t, errorResponse.Errors, "пользователь не найден", "Сообщение об ошибке должно быть корректным")
}

func TestApiHandler_handleSendCoin_Success(t *testing.T) {
	setupHandlerTest(t)
	defer teardownHandlerTest()

	// Ожидаем вызов метода SendCoin.
	mockSendCoinUseCase.EXPECT().SendCoin(gomock.Any(), "senderUser", "receiverUser", 50).Return(nil)

	// Подготавливаем тело запроса.
	requestBody := models.SendCoinRequest{
		ToUser: "receiverUser",
		Amount: 50,
	}
	jsonBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/api/sendCoin", bytes.NewBuffer(jsonBody))
	reqCtx := context.WithValue(req.Context(), "username", "senderUser")
	req = req.WithContext(reqCtx)
	recorder := httptest.NewRecorder()

	handler.handleSendCoin(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code, "Код статуса должен быть 200 OK")
}

func TestApiHandler_handleSendCoin_InvalidAmount(t *testing.T) {
	setupHandlerTest(t)
	defer teardownHandlerTest()

	// Подготавливаем тело запроса с неверной суммой.
	requestBody := models.SendCoinRequest{
		ToUser: "receiverUser",
		Amount: 0,
	}
	jsonBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/api/sendCoin", bytes.NewBuffer(jsonBody))
	reqCtx := context.WithValue(req.Context(), "username", "senderUser")
	req = req.WithContext(reqCtx)
	recorder := httptest.NewRecorder()

	handler.handleSendCoin(recorder, req)

	// Проверяем код статуса (400) и сообщение об ошибке.
	assert.Equal(t, http.StatusBadRequest, recorder.Code, "Код статуса должен быть 400 Bad Request")
	var errorResponse models.ErrorResponse
	json.NewDecoder(recorder.Body).Decode(&errorResponse)
	assert.Contains(t, errorResponse.Errors, "Неверный запрос.", "Сообщение об ошибке должно быть корректным")
}

func TestApiHandler_handleBuyItem_Success(t *testing.T) {
	setupHandlerTest(t)
	defer teardownHandlerTest()

	// Ожидаем вызов метода BuyItem
	mockBuyItemUseCase.EXPECT().BuyItem(gomock.Any(), "testuser", "pen").Return(nil)

	req := httptest.NewRequest("POST", "/api/buy/pen", nil)
	reqCtx := context.WithValue(req.Context(), "username", "testuser")
	req = req.WithContext(reqCtx)
	recorder := httptest.NewRecorder()

	handler.handleBuyItem(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code, "Код статуса должен быть 200 OK")
}

func TestApiHandler_handleBuyItem_ItemNotFound(t *testing.T) {
	setupHandlerTest(t)
	defer teardownHandlerTest()

	// Ожидаем вызов метода BuyItem, который вернет ошибку ErrItemNotFound
	mockBuyItemUseCase.EXPECT().BuyItem(gomock.Any(), gomock.Any(), "nonexistent_item").Return(usecase.ErrItemNotFound)

	req := httptest.NewRequest("POST", "/api/buy/nonexistent_item", nil)
	reqCtx := context.WithValue(req.Context(), "username", "testuser")
	req = req.WithContext(reqCtx)
	recorder := httptest.NewRecorder()

	handler.handleBuyItem(recorder, req)

	assert.Equal(t, http.StatusBadRequest, recorder.Code, "Код статуса должен быть 400 Bad Request")
	var errorResponse models.ErrorResponse
	json.NewDecoder(recorder.Body).Decode(&errorResponse)
	assert.Contains(t, errorResponse.Errors, "товар не найден", "Сообщение об ошибке должно быть корректным")
}

func TestApiHandler_handleAuth_Success(t *testing.T) {
	setupHandlerTest(t)
	defer teardownHandlerTest()

	// Ожидаемый токен.
	expectedToken := "test_jwt_token"
	// Ожидаем вызов метода Auth.
	mockUserUseCase.EXPECT().Auth(gomock.Any(), "testuser", "password").Return(expectedToken, nil)

	// Подготавливаем тело запроса.
	requestBody := models.AuthRequest{
		Username: "testuser",
		Password: "password",
	}
	jsonBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/api/auth", bytes.NewBuffer(jsonBody))
	recorder := httptest.NewRecorder()

	handler.handleAuth(recorder, req)

	// Проверяем код статуса и токен в ответе.
	assert.Equal(t, http.StatusOK, recorder.Code, "Код статуса должен быть 200 OK")
	var response models.AuthResponse
	json.NewDecoder(recorder.Body).Decode(&response)
	assert.Equal(t, expectedToken, response.Token, "Токен в ответе должен соответствовать ожидаемому")
}

func TestApiHandler_handleAuth_InvalidPassword(t *testing.T) {
	setupHandlerTest(t)
	defer teardownHandlerTest()

	// Ожидаем, что Auth вернет ошибку ErrInvalidPassword.
	mockUserUseCase.EXPECT().Auth(gomock.Any(), "testuser", "wrong_password").Return("", usecase.ErrInvalidPassword)

	requestBody := models.AuthRequest{
		Username: "testuser",
		Password: "wrong_password",
	}
	jsonBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/api/auth", bytes.NewBuffer(jsonBody))
	recorder := httptest.NewRecorder()

	handler.handleAuth(recorder, req)

	// Проверяем код статуса (401) и сообщение об ошибке
	assert.Equal(t, http.StatusUnauthorized, recorder.Code, "Код статуса должен быть 401 Unauthorized")
	var errorResponse models.ErrorResponse
	json.NewDecoder(recorder.Body).Decode(&errorResponse)
	assert.Contains(t, errorResponse.Errors, "неверный пароль", "Сообщение об ошибке должно быть корректным")
}
