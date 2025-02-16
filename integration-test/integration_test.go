package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"shop/internal/config"
	"shop/internal/db"
	http2 "shop/internal/http"
	"shop/internal/models"
	uc "shop/internal/usecase"
	"shop/pkg/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	_ "github.com/lib/pq"
)

var (
	testConfig config.Config
	testDB     *sql.DB
	log        *logger.Logger
)

func TestMain(m *testing.M) {
	var err error
	testConfig, err = config.LoadConfigFrom("test.env")
	if err != nil {
		fmt.Printf("Не удалось загрузить тестовую конфигурацию: %v\n", err)
		os.Exit(1)
	}

	log = logger.NewTestLogger()

	testDB, err = db.ConnectDB(testConfig.Database)
	if err != nil {
		log.Error("Не удалось подключиться к тестовой базе данных", "ошибка", err)
		os.Exit(1)
	}
	defer testDB.Close()

	os.Exit(m.Run())
}

// setupTestServer настраивает тестовый HTTP-сервер.
func setupTestServer() *httptest.Server {
	userDB := db.NewUserDB(testDB, log)
	itemDB := db.NewItemDB(testDB, log)
	transactionDB := db.NewTransactionDB(testDB, log)

	userInfoUseCase := uc.NewUserInfoUseCase(testConfig.JWT.SecretKey, userDB, transactionDB, log)
	sendCoinUseCase := uc.NewSendCoinUseCase(userDB, transactionDB, log)
	buyItemUseCase := uc.NewBuyItemUseCase(userDB, itemDB, transactionDB, log)

	server := http2.NewServer(testConfig.Server.Port, userInfoUseCase, sendCoinUseCase, buyItemUseCase, log)
	return httptest.NewServer(server.Handler)
}

// clearTestData очищает тестовые данные и заново создает тестовых пользователей.
func clearTestData(t *testing.T) {
	t.Helper()
	_, err := testDB.Exec(`
		DELETE FROM coin_transactions;
		DELETE FROM inventory;
		DELETE FROM users;
	`)
	require.NoError(t, err, "Не удалось очистить тестовые данные")
	require.NoError(t, createTestUsers(testDB), "Не удалось создать тестовых пользователей")
}

// createTestUsers создает тестовых пользователей в базе данных.
func createTestUsers(database *sql.DB) error {
	users := []struct {
		username string
		password string
		coins    int
	}{
		{"alice", "password", 1000},
		{"bob", "password", 1000},
		{"charlie", "password", 10},
	}

	userDB := db.NewUserDB(database, log)
	for _, u := range users {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.password), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("не удалось захэшировать пароль: %w", err)
		}

		if err := userDB.CreateUser(context.Background(), u.username, string(hashedPassword)); err != nil {
			return fmt.Errorf("не удалось создать пользователя: %w", err)
		}

		userID, err := userDB.GetUserIDByUsername(context.Background(), u.username)
		if err != nil {
			return fmt.Errorf("не удалось получить ID пользователя: %w", err)
		}

		if err := userDB.SetInitialCoins(context.Background(), userID, u.coins); err != nil {
			return fmt.Errorf("не удалось установить начальное количество монет: %w", err)
		}
	}
	return nil
}

// newTestClient создает новый HTTP-клиент для тестирования.
func newTestClient() *http.Client {
	return &http.Client{Timeout: 5 * time.Second}
}

// newAuthenticatedRequest создает новый HTTP-запрос с аутентификацией.
func newAuthenticatedRequest(t *testing.T, method, url, token string, body interface{}) *http.Request {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		require.NoError(t, json.NewEncoder(&buf).Encode(body))
	}

	req, err := http.NewRequest(method, url, &buf)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return req
}

// doRequest выполняет HTTP-запрос и проверяет код состояния ответа.
func doRequest(t *testing.T, client *http.Client, req *http.Request, expectedStatus int) *http.Response {
	t.Helper()
	resp, err := client.Do(req)
	require.NoError(t, err)
	require.Equal(t, expectedStatus, resp.StatusCode, "Неожиданный код состояния")
	return resp
}

// decodeResponse декодирует JSON-ответ.
func decodeResponse(t *testing.T, resp *http.Response, target interface{}) {
	t.Helper()
	defer resp.Body.Close()
	require.NoError(t, json.NewDecoder(resp.Body).Decode(target))
}

// getAuthToken получает токен аутентификации для пользователя.
func getAuthToken(t *testing.T, serverURL, username, password string) string {
	t.Helper()
	authReq := models.AuthRequest{
		Username: username,
		Password: password,
	}

	req := newAuthenticatedRequest(t, "POST", serverURL+"/api/auth", "", authReq)
	resp := doRequest(t, newTestClient(), req, http.StatusOK)

	var authResp models.AuthResponse
	decodeResponse(t, resp, &authResp)
	return authResp.Token
}

func TestBuyItem(t *testing.T) {
	t.Run("SuccessfulPurchase", func(t *testing.T) {
		clearTestData(t)
		server := setupTestServer()
		defer server.Close()

		token := getAuthToken(t, server.URL, "alice", "password")
		client := newTestClient()

		// Покупка товара
		req := newAuthenticatedRequest(t, "POST", server.URL+"/api/buy/hoody", token, nil)
		doRequest(t, client, req, http.StatusOK)

		// Проверка покупки
		req = newAuthenticatedRequest(t, "GET", server.URL+"/api/info", token, nil)
		resp := doRequest(t, client, req, http.StatusOK)

		var info models.InfoResponse
		decodeResponse(t, resp, &info)

		assert.Equal(t, 700, info.Coins)
		assert.Len(t, info.Inventory, 1)
		assert.Equal(t, "hoody", info.Inventory[0].Type)
		assert.Equal(t, 1, info.Inventory[0].Quantity)
	})

	t.Run("InsufficientFunds", func(t *testing.T) {
		clearTestData(t)
		server := setupTestServer()
		defer server.Close()

		// Перевод монет, чтобы оставить недостаточно средств
		senderToken := getAuthToken(t, server.URL, "charlie", "password")
		client := newTestClient()

		transferReq := newAuthenticatedRequest(t, "POST", server.URL+"/api/sendCoin", senderToken, models.SendCoinRequest{
			ToUser: "alice",
			Amount: 5,
		})
		doRequest(t, client, transferReq, http.StatusOK)

		// Попытка покупки
		buyReq := newAuthenticatedRequest(t, "POST", server.URL+"/api/buy/pen", senderToken, nil)
		resp := doRequest(t, client, buyReq, http.StatusBadRequest)

		var errorResp models.ErrorResponse
		decodeResponse(t, resp, &errorResp)
		assert.Contains(t, errorResp.Errors, "недостаточно монет")
	})

	t.Run("ItemNotFound", func(t *testing.T) {
		clearTestData(t)
		server := setupTestServer()
		defer server.Close()

		client := newTestClient()

		token := getAuthToken(t, server.URL, "alice", "password")
		req := newAuthenticatedRequest(t, "POST", server.URL+"/api/buy/nonexistentitem", token, nil)
		resp := doRequest(t, client, req, http.StatusBadRequest)

		var errorResp models.ErrorResponse
		decodeResponse(t, resp, &errorResp)
		assert.Contains(t, errorResp.Errors, "товар не найден")
	})
}

func TestSendCoins(t *testing.T) {
	t.Run("SuccessfulTransfer", func(t *testing.T) {
		clearTestData(t)
		server := setupTestServer()
		defer server.Close()

		senderToken := getAuthToken(t, server.URL, "alice", "password")
		receiverToken := getAuthToken(t, server.URL, "bob", "password")
		client := newTestClient()

		// Отправка монет
		sendReq := newAuthenticatedRequest(t, "POST", server.URL+"/api/sendCoin", senderToken, models.SendCoinRequest{
			ToUser: "bob",
			Amount: 50,
		})
		doRequest(t, client, sendReq, http.StatusOK)

		// Проверка баланса отправителя
		senderInfoReq := newAuthenticatedRequest(t, "GET", server.URL+"/api/info", senderToken, nil)
		resp := doRequest(t, client, senderInfoReq, http.StatusOK)

		var senderInfo models.InfoResponse
		decodeResponse(t, resp, &senderInfo)
		assert.Equal(t, 950, senderInfo.Coins)

		// Проверка баланса получателя
		receiverInfoReq := newAuthenticatedRequest(t, "GET", server.URL+"/api/info", receiverToken, nil)
		resp = doRequest(t, client, receiverInfoReq, http.StatusOK)

		var receiverInfo models.InfoResponse
		decodeResponse(t, resp, &receiverInfo)
		assert.Equal(t, 1050, receiverInfo.Coins)
	})

	t.Run("InsufficientFunds", func(t *testing.T) {
		clearTestData(t)
		server := setupTestServer()
		defer server.Close()

		token := getAuthToken(t, server.URL, "alice", "password")
		req := newAuthenticatedRequest(t, "POST", server.URL+"/api/sendCoin", token, models.SendCoinRequest{
			ToUser: "bob",
			Amount: 1001,
		})

		resp := doRequest(t, newTestClient(), req, http.StatusBadRequest)
		var errorResp models.ErrorResponse
		decodeResponse(t, resp, &errorResp)
		assert.Contains(t, errorResp.Errors, "недостаточно монет")
	})

	t.Run("NonExistentRecipient", func(t *testing.T) {
		clearTestData(t)
		server := setupTestServer()
		defer server.Close()

		token := getAuthToken(t, server.URL, "alice", "password")
		req := newAuthenticatedRequest(t, "POST", server.URL+"/api/sendCoin", token, models.SendCoinRequest{
			ToUser: "nonexistent",
			Amount: 50,
		})

		resp := doRequest(t, newTestClient(), req, http.StatusBadRequest)
		var errorResp models.ErrorResponse
		decodeResponse(t, resp, &errorResp)
		assert.Contains(t, errorResp.Errors, "получатель не найден")
	})
}

func TestAuth(t *testing.T) {
	t.Run("SuccessfulAuthentication", func(t *testing.T) {
		clearTestData(t)
		server := setupTestServer()
		defer server.Close()

		req := newAuthenticatedRequest(t, "POST", server.URL+"/api/auth", "", models.AuthRequest{
			Username: "alice",
			Password: "password",
		})

		resp := doRequest(t, newTestClient(), req, http.StatusOK)
		var authResp models.AuthResponse
		decodeResponse(t, resp, &authResp)
		assert.NotEmpty(t, authResp.Token)
	})

	t.Run("InvalidCredentials", func(t *testing.T) {
		clearTestData(t)
		server := setupTestServer()
		defer server.Close()

		req := newAuthenticatedRequest(t, "POST", server.URL+"/api/auth", "", models.AuthRequest{
			Username: "alice",
			Password: "wrong",
		})

		doRequest(t, newTestClient(), req, http.StatusUnauthorized)
	})
}
