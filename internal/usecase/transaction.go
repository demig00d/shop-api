// ./internal/usecase/transaction.go
package usecase

import (
	"context"
	"fmt"

	"shop/internal/db"
	"shop/pkg/logger"
)

// Ошибки
var (
	ErrInsufficientFunds = fmt.Errorf("%w: недостаточно монет для перевода", ErrInvalidRequest)
	ErrSelfTransfer      = fmt.Errorf("%w: нельзя отправить монеты самому себе", ErrInvalidRequest)
	ErrReceiverNotFound  = fmt.Errorf("%w: получатель не найден", ErrInvalidRequest)
	ErrInvalidAmount     = fmt.Errorf("%w: сумма перевода должна быть положительной", ErrInvalidRequest)
)

// SendCoinUseCaseInterface интерфейс для use case'а отправки монет.
type SendCoinUseCaseInterface interface {
	SendCoin(ctx context.Context, senderUsername string, receiverUsername string, amount int) error
}

// SendCoinUseCase реализует SendCoinUseCaseInterface.
type SendCoinUseCase struct {
	userDB        db.UserDBInterface
	transactionDB db.TransactionDBInterface
	log           *logger.Logger
}

// NewSendCoinUseCase создает новый SendCoinUseCase.
func NewSendCoinUseCase(userDB db.UserDBInterface, transactionDB db.TransactionDBInterface, log *logger.Logger) *SendCoinUseCase {
	return &SendCoinUseCase{
		userDB:        userDB,
		transactionDB: transactionDB,
		log:           log,
	}
}

// SendCoin обрабатывает бизнес-логику перевода монет.
func (uc *SendCoinUseCase) SendCoin(ctx context.Context, senderUsername string, receiverUsername string, amount int) error {
	uc.log.Debug("SendCoin", "senderUsername", senderUsername, "receiverUsername", receiverUsername, "amount", amount)

	if amount <= 0 {
		uc.log.Warn("Неверная сумма перевода", "amount", amount)
		return ErrInvalidAmount
	}

	senderUser, err := uc.userDB.GetUserByUsername(ctx, senderUsername)
	if err != nil {
		uc.log.Error("Ошибка GetUserByUsername (sender)", "senderUsername", senderUsername, "error", err)
		return fmt.Errorf("ошибка при получении отправителя: %w", err)
	}
	if senderUser == nil {
		uc.log.Warn("Отправитель не найден", "senderUsername", senderUsername)
		return ErrUserNotFound
	}

	receiverUser, err := uc.userDB.GetUserByUsername(ctx, receiverUsername)
	if err != nil {
		uc.log.Error("Ошибка GetUserByUsername (receiver)", "receiverUsername", receiverUsername, "error", err)
		return fmt.Errorf("ошибка при получении получателя: %w", err)
	}
	if receiverUser == nil {
		uc.log.Warn("Получатель не найден", "receiverUsername", receiverUsername)
		return ErrReceiverNotFound
	}
	uc.log.Debug("Пользователи найдены", "senderUsername", senderUsername, "receiverUsername", receiverUsername)

	if senderUser.ID == receiverUser.ID {
		uc.log.Warn("Попытка отправить монеты самому себе", "senderUsername", senderUsername)
		return ErrSelfTransfer
	}

	if senderUser.Coins < amount {
		uc.log.Warn("Недостаточно монет для перевода", "senderUsername", senderUsername, "coins", senderUser.Coins, "amount", amount)
		return ErrInsufficientFunds
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

	err = uc.userDB.UpdateUserCoins(ctx, senderUser.ID, senderUser.Coins-amount)
	if err != nil {
		uc.log.Error("Ошибка UpdateUserCoins (sender)", "senderUserID", senderUser.ID, "amount", amount, "error", err)
		return err
	}
	err = uc.userDB.UpdateUserCoins(ctx, receiverUser.ID, receiverUser.Coins+amount)
	if err != nil {
		uc.log.Error("Ошибка UpdateUserCoins (receiver)", "receiverUserID", receiverUser.ID, "amount", amount, "error", err)
		return err
	}

	err = uc.transactionDB.RecordTransaction(ctx, senderUser.ID, receiverUser.ID, amount, tx)
	if err != nil {
		uc.log.Error("Ошибка RecordTransaction", "senderUserID", senderUser.ID, "receiverUserID", receiverUser.ID, "amount", amount, "error", err)
		return err
	}

	return nil
}
