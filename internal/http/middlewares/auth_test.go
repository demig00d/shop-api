package middlewares

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"shop/internal/http/helpers"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	ucmocks "shop/internal/usecase/mocks"
)

func TestAuthMiddleware_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserUseCase := ucmocks.NewMockUserUseCaseInterface(ctrl)
	middlewareHandler := NewAuthMiddlewareHandler(mockUserUseCase)

	// Тестовый обработчик, который будет вызван после middleware.
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Проверяем, что username был извлечен из контекста.
		username := helpers.UsernameFromContext(r.Context())
		assert.Equal(t, "testuser", username, "Username должен быть извлечен из контекста")
		w.WriteHeader(http.StatusOK)
	})

	// Ожидаем вызов VerifyJWTToken с "valid_token".
	mockUserUseCase.EXPECT().VerifyJWTToken("valid_token").Return("testuser", nil)

	req := httptest.NewRequest("GET", "/api/protected", nil)
	// Устанавливаем заголовок Authorization.
	req.Header.Set("Authorization", "Bearer valid_token")
	recorder := httptest.NewRecorder()

	// Применяем middleware к тестовому обработчику.
	middleware := middlewareHandler.AuthMiddleware(testHandler)
	middleware.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code, "Код статуса должен быть 200 OK")
}

func TestAuthMiddleware_NoToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserUseCase := ucmocks.NewMockUserUseCaseInterface(ctrl)
	middlewareHandler := NewAuthMiddlewareHandler(mockUserUseCase)

	// Тестовый обработчик, который *не* должен быть вызван.
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler не должен быть вызван при отсутствии токена")
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/api/protected", nil)
	recorder := httptest.NewRecorder()

	middleware := middlewareHandler.AuthMiddleware(testHandler)
	middleware.ServeHTTP(recorder, req)

	// Проверяем код статуса (401).
	assert.Equal(t, http.StatusUnauthorized, recorder.Code, "Код статуса должен быть 401 Unauthorized")
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserUseCase := ucmocks.NewMockUserUseCaseInterface(ctrl)
	middlewareHandler := NewAuthMiddlewareHandler(mockUserUseCase)

	// Тестовый обработчик, который *не* должен быть вызван.
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler не должен быть вызван при неверном токене")
		w.WriteHeader(http.StatusOK)
	})

	// Ожидаем вызов VerifyJWTToken, который вернет ошибку.
	mockUserUseCase.EXPECT().VerifyJWTToken("invalid_token").Return("", errors.New("invalid token error"))

	req := httptest.NewRequest("GET", "/api/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid_token")
	recorder := httptest.NewRecorder()

	middleware := middlewareHandler.AuthMiddleware(testHandler)
	middleware.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusUnauthorized, recorder.Code, "Код статуса должен быть 401 Unauthorized")
}
