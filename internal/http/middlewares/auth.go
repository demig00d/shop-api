package middlewares

import (
	"context"
	"net/http"

	"shop/internal/http/helpers"
	"shop/internal/usecase"
	"shop/pkg/logger"
	"strings"
)

type AuthMiddlewareHandler struct {
	userUseCase usecase.UserUseCaseInterface
}

func NewAuthMiddlewareHandler(uc usecase.UserUseCaseInterface) AuthMiddlewareHandler {
	return AuthMiddlewareHandler{userUseCase: uc}
}

// AuthMiddleware middleware функция для проверки JWT токена авторизации.
func (h AuthMiddlewareHandler) AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := logger.FromContext(r.Context())
		log.Debug("Проверка авторизации", "path", r.URL.Path, "method", r.Method)

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			log.Warn("Отсутствует токен авторизации")
			helpers.RespondWithError(w, http.StatusUnauthorized, "Не авторизован: отсутствует токен")

			return
		}
		// Извлекаем токен из заголовка, предполагая схему "Bearer {token}".
		tokenString := strings.Replace(authHeader, "Bearer ", "", 1)
		username, err := h.userUseCase.VerifyJWTToken(tokenString)
		if err != nil {
			log.Warn("JWT верификация не удалась", "error", err)
			helpers.RespondWithError(w, http.StatusUnauthorized, "Не авторизован: "+err.Error())
			return
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, "username", username)

		// Add logger to context
		ctx = logger.WithLogger(ctx, log.With("username", username))

		next.ServeHTTP(w, r.WithContext(ctx))
	}
}
