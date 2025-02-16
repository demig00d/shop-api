package http

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	uc "shop/internal/usecase"
	"shop/pkg/logger"
)

// NewServer создает и настраивает новый HTTP сервер.
func NewServer(
	port string,
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

	serverAddress := fmt.Sprintf(":%s", port)

	addr := fmt.Sprintf("http://localhost:%s", port)
	slog.Info("Сервер запущен", slog.String("address", addr))
	slog.Info("Swagger UI доступен", slog.String("address", fmt.Sprintf("http://localhost:%s/docs/", port)))
	server := &http.Server{
		Addr:         serverAddress,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	return server
}
