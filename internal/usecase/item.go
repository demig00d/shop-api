// ./internal/usecase/item.go
package usecase

import (
	"context"
	"fmt"

	"shop/internal/db"
	"shop/pkg/logger"
)

// Ошибки
var (
	ErrItemNotFound   = fmt.Errorf("%w: товар не найден", ErrNotFound)
	ErrItemRequired   = fmt.Errorf("%w: название предмета обязательно", ErrInvalidRequest)
	ErrNotEnoughCoins = fmt.Errorf("%w: недостаточно монет", ErrInvalidRequest)
)

// BuyItemUseCaseInterface интерфейс для use case'а покупки предмета.
type BuyItemUseCaseInterface interface {
	BuyItem(ctx context.Context, username string, itemName string) error
}

// BuyItemUseCase реализует BuyItemUseCaseInterface.
type BuyItemUseCase struct {
	userDB        db.UserDBInterface
	itemDB        db.ItemDBInterface
	transactionDB db.TransactionDBInterface
	log           *logger.Logger
}

// NewBuyItemUseCase создает новый BuyItemUseCase.
func NewBuyItemUseCase(userDB db.UserDBInterface, itemDB db.ItemDBInterface, transactionDB db.TransactionDBInterface, log *logger.Logger) *BuyItemUseCase {
	return &BuyItemUseCase{
		userDB:        userDB,
		itemDB:        itemDB,
		transactionDB: transactionDB,
		log:           log,
	}
}

// BuyItem обрабатывает бизнес-логику покупки предмета.
func (uc *BuyItemUseCase) BuyItem(ctx context.Context, username string, item string) error {
	uc.log.Debug("BuyItem", "username", username, "item", item)

	if item == "" {
		uc.log.Warn("Название предмета не указано")
		return ErrItemRequired
	}

	price, err := uc.itemDB.GetItemPrice(ctx, item)
	if err != nil {
		uc.log.Error("Ошибка GetItemPrice", "item", item, "error", err)
		return ErrItemNotFound
	}

	user, err := uc.userDB.GetUserByUsername(ctx, username)
	if err != nil {
		uc.log.Error("Ошибка GetUserByUsername", "username", username, "error", err)
		return fmt.Errorf("ошибка при получении пользователя: %w", err)
	}
	if user == nil {
		uc.log.Warn("Пользователь не найден", "username", username)
		return ErrUserNotFound
	}
	uc.log.Debug("Пользователь найден", "username", username, "userID", user.ID)

	if user.Coins < price {
		uc.log.Warn("Недостаточно монет", "username", username, "coins", user.Coins, "price", price, "item", item)
		return ErrNotEnoughCoins
	}

	tx, err := uc.transactionDB.GetDB().BeginTx(ctx, nil)
	if err != nil {
		uc.log.Error("Ошибка начала транзакции", "error", err)
		return fmt.Errorf("ошибка начала транзакции: %w", err)
	}
	defer func() {
		if p := recover(); p != nil {
			if err := tx.Rollback(); err != nil {
				uc.log.Error("Ошибка отката транзакции", "error", err)
			}
			uc.log.Error("Паника во время транзакции, rollback", "panic", p)
			panic(p) // Re-panic after rollback.
		} else if err != nil {
			if err := tx.Rollback(); err != nil {
				uc.log.Error("Ошибка отката транзакции", "error", err)
			}
			uc.log.Error("Транзакция отменена из-за ошибки", "error", err)
		} else {
			err = tx.Commit()
			if err != nil {
				uc.log.Error("Ошибка коммита транзакции", "error", err)
			}
		}
	}()

	err = uc.userDB.UpdateUserCoins(ctx, user.ID, user.Coins-price)
	if err != nil {
		uc.log.Error("Ошибка UpdateUserCoins", "userID", user.ID, "price", price, "error", err)
		return err
	}

	err = uc.userDB.UpdateUserInventory(ctx, user.ID, item, 1, tx)
	if err != nil {
		uc.log.Error("Ошибка UpdateUserInventory", "userID", user.ID, "item", item, "error", err)
		return err
	}

	return nil
}
