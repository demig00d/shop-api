package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"shop/internal/models"
	"shop/internal/usecase"
	"shop/pkg/logger"
)

// ApiHandler структура для обработки всех API запросов.
type ApiHandler struct {
	userUseCase     usecase.UserUseCaseInterface
	sendCoinUseCase usecase.SendCoinUseCaseInterface
	buyItemUseCase  usecase.BuyItemUseCaseInterface
	authMiddleware  authMiddlewareHandler
	log             *logger.Logger
}

// NewApiHandler создает новый ApiHandler.
func NewApiHandler(
	userUseCase usecase.UserUseCaseInterface,
	sendCoinUseCase usecase.SendCoinUseCaseInterface,
	buyItemUseCase usecase.BuyItemUseCaseInterface,
	log *logger.Logger,
) *ApiHandler {
	return &ApiHandler{
		userUseCase:     userUseCase,
		sendCoinUseCase: sendCoinUseCase,
		buyItemUseCase:  buyItemUseCase,
		authMiddleware:  NewAuthMiddlewareHandler(userUseCase),
		log:             log,
	}
}

// RegisterRoutes регистрирует обработчики для API маршрутов.
func (h *ApiHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/info", h.authMiddleware.AuthMiddleware(h.handleInfo))
	mux.HandleFunc("/api/sendCoin", h.authMiddleware.AuthMiddleware(h.handleSendCoin))
	mux.HandleFunc("/api/buy/", h.authMiddleware.AuthMiddleware(h.handleBuyItem))
	mux.HandleFunc("/api/auth", h.handleAuth)
}

// handleInfo обрабатывает запросы на получение информации о пользователе.
func (h *ApiHandler) handleInfo(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())
	log.Debug("Обработка запроса handleInfo", "path", r.URL.Path, "method", r.Method)

	username := UsernameFromContext(r.Context())

	response, err := h.userUseCase.GetUserInfo(r.Context(), username)
	if err != nil {
		log.Error("Ошибка usecase GetUserInfo", "username", username, "error", err)
		if errors.Is(err, usecase.ErrUserNotFound) {
			RespondWithError(w, http.StatusNotFound, err.Error())
		} else {
			RespondWithError(w, http.StatusInternalServerError, "Внутренняя ошибка сервера.")
		}
		return
	}

	RespondWithJSON(w, http.StatusOK, response)
}

// handleSendCoin обрабатывает запросы на отправку монет.
func (h *ApiHandler) handleSendCoin(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())
	log.Debug("Обработка запроса handleSendCoin", "path", r.URL.Path, "method", r.Method)

	username := UsernameFromContext(r.Context())

	var req models.SendCoinRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error("Ошибка декодирования запроса handleSendCoin", "error", err)
		RespondWithError(w, http.StatusBadRequest, "Неверный запрос.")
		return
	}
	defer r.Body.Close()

	if req.Amount <= 0 {
		RespondWithError(w, http.StatusBadRequest, "Неверный запрос.")
		return
	}

	err := h.sendCoinUseCase.SendCoin(r.Context(), username, req.ToUser, req.Amount)
	if err != nil {
		log.Error("Ошибка usecase SendCoin", "username", username, "error", err)
		if errors.Is(err, usecase.ErrInvalidAmount) ||
			errors.Is(err, usecase.ErrInsufficientFunds) ||
			errors.Is(err, usecase.ErrSelfTransfer) ||
			errors.Is(err, usecase.ErrReceiverNotFound) ||
			errors.Is(err, usecase.ErrUserNotFound) {
			RespondWithError(w, http.StatusBadRequest, err.Error())
		} else {
			RespondWithError(w, http.StatusInternalServerError, "Внутренняя ошибка сервера.")
		}
		return
	}

	RespondWithOK(w)
}

// handleBuyItem обрабатывает запросы на покупку предмета за монеты.
func (h *ApiHandler) handleBuyItem(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())
	log.Debug("Обработка запроса handleBuyItem", "path", r.URL.Path, "method", r.Method)

	itemPath := strings.TrimPrefix(r.URL.Path, "/api/buy/")

	if itemPath == "" {
		RespondWithError(w, http.StatusBadRequest, "Название предмета обязательно в пути /api/buy/{itemName}")
		return
	}

	username := UsernameFromContext(r.Context())

	err := h.buyItemUseCase.BuyItem(r.Context(), username, itemPath)
	if err != nil {
		log.Error("Ошибка usecase BuyItem", "username", username, "item", itemPath, "error", err)
		if errors.Is(err, usecase.ErrItemNotFound) ||
			errors.Is(err, usecase.ErrItemRequired) ||
			errors.Is(err, usecase.ErrNotEnoughCoins) ||
			errors.Is(err, usecase.ErrUserNotFound) {
			RespondWithError(w, http.StatusBadRequest, err.Error())
		} else {
			RespondWithError(w, http.StatusInternalServerError, "Внутренняя ошибка сервера.")
		}
		return
	}
	RespondWithOK(w)
}

// handleAuth обрабатывает запросы аутентификации.
func (h *ApiHandler) handleAuth(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())
	log.Debug("Обработка запроса handleAuth", "path", r.URL.Path, "method", r.Method)

	var req models.AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error("Ошибка декодирования запроса handleAuth", "error", err)
		RespondWithError(w, http.StatusBadRequest, "Неверный запрос.")
		return
	}
	defer r.Body.Close()

	token, err := h.userUseCase.Auth(r.Context(), req.Username, req.Password)
	if err != nil {
		log.Warn("Ошибка аутентификации", "username", req.Username, "error", err)
		if errors.Is(err, usecase.ErrInvalidPassword) {
			RespondWithError(w, http.StatusUnauthorized, err.Error())
		} else {
			RespondWithError(w, http.StatusInternalServerError, "Внутренняя ошибка сервера.")
		}
		return
	}

	response := models.AuthResponse{Token: token}
	RespondWithJSON(w, http.StatusOK, response)
}
