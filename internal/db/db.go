package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"shop/internal/models"
	"shop/pkg/logger"
)

// Интерфейсы для взаимодействия с данными пользователей, товаров и транзакций.
type UserDBInterface interface {
	GetUserByUsername(ctx context.Context, username string) (*models.DBUser, error)
	CreateUser(ctx context.Context, username string, passwordHash string) error
	UpdateUserCoins(ctx context.Context, userID int, coins int) error
	GetUserInventory(ctx context.Context, userID int) ([]models.DBInventoryItem, error)
	UpdateUserInventory(ctx context.Context, userID int, itemType string, quantity int, tx *sql.Tx) error
	GetUserIDByUsername(ctx context.Context, username string) (int, error)
	SetInitialCoins(ctx context.Context, userID int, initialCoins int) error
}

type ItemDBInterface interface {
	GetItemPrice(ctx context.Context, itemName string) (int, error)
}

type TransactionDBInterface interface {
	RecordTransaction(ctx context.Context, senderUserID int, receiverUserID int, amount int, tx *sql.Tx) error
	GetDB() *sql.DB
	GetCoinHistory(ctx context.Context, userID int) (*models.CoinHistory, error)
}

// Реализации для PostgreSQL.
type UserDB struct {
	Db  *sql.DB
	log *logger.Logger
}

type ItemDB struct {
	Db  *sql.DB
	log *logger.Logger
}

type TransactionDB struct {
	Db  *sql.DB
	log *logger.Logger
}

// Функции создания новых экземпляров.
func NewUserDB(db *sql.DB, log *logger.Logger) *UserDB {
	return &UserDB{Db: db, log: log}
}

func NewItemDB(db *sql.DB, log *logger.Logger) *ItemDB {
	return &ItemDB{Db: db, log: log}
}

func NewTransactionDB(db *sql.DB, log *logger.Logger) *TransactionDB {
	return &TransactionDB{Db: db, log: log}
}

// GetDB возвращает базовое соединение sql.DB.
func (tdb *TransactionDB) GetDB() *sql.DB {
	return tdb.Db
}

// GetUserByUsername получает пользователя из базы данных по имени пользователя.
func (udb *UserDB) GetUserByUsername(ctx context.Context, username string) (*models.DBUser, error) {
	udb.log.Debug("GetUserByUsername", "username", username)
	user := &models.DBUser{}
	err := udb.Db.QueryRowContext(ctx, "SELECT id, username, password_hash, coins FROM users WHERE username = $1", username).Scan(&user.ID, &user.Username, &user.PasswordHash, &user.Coins)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Пользователь не найден
		}
		udb.log.Error("Ошибка SQL запроса GetUserByUsername", "username", username, "error", err)
		return nil, fmt.Errorf("ошибка при получении пользователя по имени: %w", err)
	}
	return user, nil
}

// CreateUser создает нового пользователя в базе данных.
func (udb *UserDB) CreateUser(ctx context.Context, username string, passwordHash string) error {
	udb.log.Debug("CreateUser", "username", username)
	_, err := udb.Db.ExecContext(ctx, "INSERT INTO users (username, password_hash, coins) VALUES ($1, $2, 0)", username, passwordHash) // Монеты устанавливаются в 0 при создании
	if err != nil {
		udb.log.Error("Ошибка SQL запроса CreateUser", "username", username, "error", err)
		return fmt.Errorf("ошибка при создании пользователя: %w", err)
	}
	return nil
}

// UpdateUserCoins обновляет баланс монет пользователя в базе данных.
func (udb *UserDB) UpdateUserCoins(ctx context.Context, userID int, coins int) error {
	_, err := udb.Db.ExecContext(ctx, "UPDATE users SET coins = $1 WHERE id = $2", coins, userID)
	udb.log.Debug("UpdateUserCoins", "userID", userID, "coins", coins)
	if err != nil {
		udb.log.Error("Ошибка SQL запроса UpdateUserCoins", "userID", userID, "coins", coins, "error", err)
		return fmt.Errorf("ошибка при обновлении монет пользователя: %w", err)
	}
	return nil
}

// GetUserInventory получает инвентарь пользователя из базы данных.
func (udb *UserDB) GetUserInventory(ctx context.Context, userID int) ([]models.DBInventoryItem, error) {
	rows, err := udb.Db.QueryContext(ctx, "SELECT id, user_id, item_type, quantity FROM inventory WHERE user_id = $1", userID)
	if err != nil {
		udb.log.Error("Ошибка SQL запроса GetUserInventory", "userID", userID, "error", err)
		return nil, fmt.Errorf("ошибка при получении инвентаря пользователя: %w", err)
	}
	defer rows.Close()

	inventory := []models.DBInventoryItem{}
	for rows.Next() {
		item := models.DBInventoryItem{}
		if err := rows.Scan(&item.ID, &item.UserID, &item.ItemType, &item.Quantity); err != nil {
			udb.log.Error("Ошибка сканирования строки GetUserInventory", "userID", userID, "error", err)
			return nil, fmt.Errorf("ошибка при сканировании элемента инвентаря: %w", err)
		}
		inventory = append(inventory, item)
	}
	if err := rows.Err(); err != nil {
		udb.log.Error("Ошибка итерации строк GetUserInventory", "userID", userID, "error", err)
		return nil, fmt.Errorf("ошибка при итерации строк инвентаря: %w", err)
	}
	return inventory, nil
}

// UpdateUserInventory обновляет инвентарь пользователя в базе данных.
func (udb *UserDB) UpdateUserInventory(ctx context.Context, userID int, itemType string, quantity int, tx *sql.Tx) error {
	var existingQuantity int
	err := tx.QueryRowContext(ctx, "SELECT quantity FROM inventory WHERE user_id = $1 AND item_type = $2", userID, itemType).Scan(&existingQuantity)
	if err == nil { // Элемент существует, обновляем количество
		udb.log.Debug("UpdateUserInventory: Element exists, updating quantity", "userID", userID, "itemType", itemType, "quantity", quantity)
		_, err := tx.ExecContext(ctx, "UPDATE inventory SET quantity = $1 WHERE user_id = $2 AND item_type = $3", existingQuantity+quantity, userID, itemType)
		if err != nil {
			udb.log.Error("Ошибка SQL запроса UpdateUserInventory (update existing)", "userID", userID, "itemType", itemType, "quantity", quantity, "error", err)
			return fmt.Errorf("ошибка при обновлении существующего элемента инвентаря: %w", err)
		}
	} else if err == sql.ErrNoRows { // Элемент не существует, добавляем новый
		udb.log.Debug("UpdateUserInventory: Element does not exist, adding new", "userID", userID, "itemType", itemType, "quantity", quantity)
		_, err := tx.ExecContext(ctx, "INSERT INTO inventory (user_id, item_type, quantity) VALUES ($1, $2, $3)", userID, itemType, quantity)
		if err != nil {
			udb.log.Error("Ошибка SQL запроса UpdateUserInventory (insert new)", "userID", userID, "itemType", itemType, "quantity", quantity, "error", err)
			return fmt.Errorf("ошибка при добавлении нового элемента инвентаря: %w", err)
		}
	} else if err != nil {
		udb.log.Error("Ошибка проверки существования элемента инвентаря", "userID", userID, "itemType", itemType, "error", err)
		return fmt.Errorf("ошибка при проверке существующего элемента инвентаря: %w", err)
	}
	return nil
}

// GetItemPrice получает цену товара из базы данных.
func (idb *ItemDB) GetItemPrice(ctx context.Context, itemName string) (int, error) {
	idb.log.Debug("GetItemPrice", "itemName", itemName)
	var price int
	err := idb.Db.QueryRowContext(ctx, "SELECT price FROM items WHERE item_name = $1", itemName).Scan(&price)
	if err != nil {
		if err == sql.ErrNoRows {
			idb.log.Warn("Товар не найден", "itemName", itemName)
			return 0, fmt.Errorf("товар '%s' не найден", itemName)
		}
		idb.log.Error("Ошибка SQL запроса GetItemPrice", "itemName", itemName, "error", err)
		return 0, fmt.Errorf("ошибка при получении цены товара: %w", err)
	}
	return price, nil
}

// RecordTransaction записывает транзакцию монет в базу данных.
func (tdb *TransactionDB) RecordTransaction(ctx context.Context, senderUserID int, receiverUserID int, amount int, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, "INSERT INTO coin_transactions (sender_user_id, receiver_user_id, amount, transaction_date) VALUES ($1, $2, $3, $4)", senderUserID, receiverUserID, amount, time.Now())
	tdb.log.Debug("RecordTransaction", "senderUserID", senderUserID, "receiverUserID", receiverUserID, "amount", amount)
	if err != nil {
		tdb.log.Error("Ошибка SQL запроса RecordTransaction", "senderUserID", senderUserID, "receiverUserID", receiverUserID, "amount", amount, "error", err)
		return fmt.Errorf("ошибка при записи транзакции: %w", err)
	}
	return nil
}

// GetCoinHistory получает историю транзакций монет для пользователя.
func (tdb *TransactionDB) GetCoinHistory(ctx context.Context, userID int) (*models.CoinHistory, error) {
	history := &models.CoinHistory{
		Received: []models.Transaction{},
		Sent:     []models.Transaction{},
	}

	// Полученные транзакции
	rows, err := tdb.Db.QueryContext(ctx, `
        SELECT ct.amount, u_sender.username
        FROM coin_transactions ct
        INNER JOIN users u_sender ON ct.sender_user_id = u_sender.id
        WHERE ct.receiver_user_id = $1
        ORDER BY ct.transaction_date DESC`, userID)
	if err != nil {
		tdb.log.Error("Ошибка SQL запроса GetCoinHistory (received)", "userID", userID, "error", err)
		return nil, fmt.Errorf("ошибка при получении полученных транзакций: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var transaction models.Transaction
		var senderUsername string
		if err := rows.Scan(&transaction.Amount, &senderUsername); err != nil {
			tdb.log.Error("Ошибка сканирования строки GetCoinHistory (received)", "userID", userID, "error", err)
			continue
		}
		transaction.FromUser = senderUsername
		history.Received = append(history.Received, transaction)
	}
	if err = rows.Err(); err != nil {
		tdb.log.Error("Ошибка итерации строк GetCoinHistory (received)", "userID", userID, "error", err)
		return nil, fmt.Errorf("ошибка при итерации строк полученных транзакций: %w", err)
	}

	// Отправленные транзакции
	rows, err = tdb.Db.QueryContext(ctx, `
        SELECT ct.amount, u_receiver.username
        FROM coin_transactions ct
        INNER JOIN users u_receiver ON ct.receiver_user_id = u_receiver.id
        WHERE ct.sender_user_id = $1
        ORDER BY ct.transaction_date DESC`, userID)
	if err != nil {
		tdb.log.Error("Ошибка SQL запроса GetCoinHistory (sent)", "userID", userID, "error", err)
		return nil, fmt.Errorf("ошибка при получении отправленных транзакций: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var transaction models.Transaction
		var receiverUsername string
		if err := rows.Scan(&transaction.Amount, &receiverUsername); err != nil {
			tdb.log.Error("Ошибка сканирования строки GetCoinHistory (sent)", "userID", userID, "error", err)
			continue
		}
		transaction.ToUser = receiverUsername
		history.Sent = append(history.Sent, transaction)
	}
	if err = rows.Err(); err != nil {
		tdb.log.Error("Ошибка итерации строк GetCoinHistory (sent)", "userID", userID, "error", err)
		return nil, fmt.Errorf("ошибка при итерации строк отправленных транзакций: %w", err)
	}

	return history, nil
}

// GetUserIDByUsername получает ID пользователя из базы данных по имени пользователя.
func (udb *UserDB) GetUserIDByUsername(ctx context.Context, username string) (int, error) {
	udb.log.Debug("GetUserIDByUsername", "username", username)
	var userID int
	err := udb.Db.QueryRowContext(ctx, "SELECT id FROM users WHERE username = $1", username).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			udb.log.Warn("Пользователь не найден", "username", username)
			return 0, fmt.Errorf("пользователь не найден")
		}
		udb.log.Error("Ошибка SQL запроса GetUserIDByUsername", "username", username, "error", err)
		return 0, fmt.Errorf("ошибка при получении ID пользователя по имени: %w", err)
	}
	return userID, nil
}

// SetInitialCoins устанавливает начальный баланс монет для пользователя.
func (udb *UserDB) SetInitialCoins(ctx context.Context, userID int, initialCoins int) error {
	_, err := udb.Db.ExecContext(ctx, "UPDATE users SET coins = $1 WHERE id = $2", initialCoins, userID)
	udb.log.Debug("SetInitialCoins", "userID", userID, "initialCoins", initialCoins)
	if err != nil {
		udb.log.Error("Ошибка SQL запроса SetInitialCoins", "userID", userID, "initialCoins", initialCoins, "error", err)
		return fmt.Errorf("ошибка при установке начального количества монет для пользователя: %w", err)
	}
	return nil
}
