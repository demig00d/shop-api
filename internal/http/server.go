package http

import (
	"log/slog"
	"net/http"
	"time"

	uc "shop/internal/usecase"
	"shop/pkg/logger"
)

// NewServer создает и настраивает новый HTTP сервер.
func NewServer(
	userUseCase uc.UserUseCaseInterface,
	sendCoinUseCase uc.SendCoinUseCaseInterface,
	buyItemUseCase uc.BuyItemUseCaseInterface,
	log *logger.Logger,
) *http.Server {
	mux := http.NewServeMux()

	apiHandler := NewApiHandler(userUseCase, sendCoinUseCase, buyItemUseCase, log)
	apiHandler.RegisterRoutes(mux)

	swaggerDir := "./swagger"
	swaggerHandler := http.FileServer(http.Dir(swaggerDir))

	mux.Handle("/docs/", http.StripPrefix("/docs/", swaggerHandler))
	mux.Handle("/schema.json", swaggerHandler)

	serverAddress := "http://localhost:8080"
	slog.Info("Сервер запущен", slog.String("address", serverAddress))
	slog.Info("Swagger UI доступен", slog.String("address", "http://localhost:8080/docs/"))
	server := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	return server
}
