package http

import (
	"context"
	"encoding/json"
	"net/http"

	"shop/internal/models"
)

// RespondWithError отправляет JSON ответ с ошибкой и указанным статус кодом.
func RespondWithError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	resp := models.ErrorResponse{Errors: message}
	_ = json.NewEncoder(w).Encode(resp)
}

// RespondWithOK отправляет ответ с кодом 200 OK.
func RespondWithOK(w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
}

// RespondWithJSON отправляет JSON ответ с указанным статус кодом и полезной нагрузкой.
func RespondWithJSON(w http.ResponseWriter, statusCode int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload)
}

// UsernameFromContext извлекает имя пользователя из контекста запроса.
func UsernameFromContext(ctx context.Context) string {
	val := ctx.Value("username")
	if username, ok := val.(string); ok {
		return username
	}
	return ""
}

// ContextKey тип для ключей контекста, чтобы избежать коллизий.
type ContextKey string
