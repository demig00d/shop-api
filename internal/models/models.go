package models

import "time"

// InfoResponse соответствует components/schemas/InfoResponse в swagger спецификации.
type InfoResponse struct {
	Coins       int             `json:"coins"`
	Inventory   []InventoryItem `json:"inventory"`
	CoinHistory CoinHistory     `json:"coinHistory"`
}

// InventoryItem описывает предмет инвентаря.
type InventoryItem struct {
	Type     string `json:"type"`
	Quantity int    `json:"quantity"`
}

// CoinHistory описывает историю транзакций монет пользователя.
type CoinHistory struct {
	Received []Transaction `json:"received"`
	Sent     []Transaction `json:"sent"`
}

// Transaction описывает детали транзакции монет
type Transaction struct {
	FromUser string `json:"fromUser,omitempty"`
	ToUser   string `json:"toUser,omitempty"`
	Amount   int    `json:"amount"`
}

// ErrorResponse соответствует components/schemas/ErrorResponse в swagger спецификации.
type ErrorResponse struct {
	Errors string `json:"errors"`
}

// AuthRequest соответствует components/schemas/AuthRequest в swagger спецификации.
type AuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// AuthResponse соответствует components/schemas/AuthResponse в swagger спецификации.
type AuthResponse struct {
	Token string `json:"token"`
}

// SendCoinRequest соответствует components/schemas/SendCoinRequest в swagger спецификации.
type SendCoinRequest struct {
	ToUser string `json:"toUser"`
	Amount int    `json:"amount"`
}

// DBUser модель пользователя для базы данных.
type DBUser struct {
	ID           int    `json:"id"`
	Username     string `json:"username"`
	PasswordHash string `json:"password_hash"`
	Coins        int    `json:"coins"`
}

// DBInventoryItem модель предмета инвентаря в базе данных.
type DBInventoryItem struct {
	ID       int    `json:"id"`
	UserID   int    `json:"user_id"`
	ItemType string `json:"item_type"`
	Quantity int    `json:"quantity"`
}

// DBTransaction модель транзакции для базы данных.
type DBTransaction struct {
	ID              int       `json:"id"`
	SenderUserID    int       `json:"sender_user_id"`
	ReceiverUserID  int       `json:"receiver_user_id"`
	Amount          int       `json:"amount"`
	TransactionDate time.Time `json:"transaction_date"`
}

// DBItem модель товара для продажи.
type DBItem struct {
	ID       int    `json:"id"`
	ItemName string `json:"item_name"`
	Price    int    `json:"price"`
}
