package main

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
	"shop/internal/config"
	"shop/internal/db"
	"shop/internal/http"
	uc "shop/internal/usecase"
	"shop/pkg/logger"
)

func main() {
	// init config
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Ошибка загрузки конфигурации: %v\n", err)
		os.Exit(1)
	}

	// init logger
	level, err := logger.ParseLogLevel(cfg.LogLevel)
	log := logger.New(level)

	if err != nil {
		log.Warn("Неверный уровень логгирования, используется уровень по умолчанию Info", "error", err, "LogLevel", cfg.LogLevel)
	}

	log.Info("Конфигурация загружена", "config", cfg)

	// init storage
	database, err := connectDB(cfg.Database, log)
	if err != nil {
		log.Error("Ошибка подключения к базе данных", "error", err)
		os.Exit(1)
	}

	userDB := db.NewUserDB(database, log)
	itemDB := db.NewItemDB(database, log)
	transactionDB := db.NewTransactionDB(database, log)

	userInfoUseCase := uc.NewUserInfoUseCase(cfg.JWT.SecretKey, userDB, transactionDB, log)
	sendCoinUseCase := uc.NewSendCoinUseCase(userDB, transactionDB, log)
	buyItemUseCase := uc.NewBuyItemUseCase(userDB, itemDB, transactionDB, log)

	srv := http.NewServer(cfg.Server.Port, userInfoUseCase, sendCoinUseCase, buyItemUseCase, log)
	log.Info("Сервер запущен", "address", srv.Addr)
	if err := srv.ListenAndServe(); err != nil {
		log.Error("Ошибка сервера", "error", err)
		os.Exit(1)
	}

	err = database.Close()
	if err != nil {
		log.Error("Ошибка закрытия соединения с базой данных", "error", err)
	}
}

func connectDB(dbCfg config.DatabaseConfig, log *logger.Logger) (*sql.DB, error) {
	// Строка подключения к базе данных PostgreSQL
	connStr := fmt.Sprintf(
		"port=%s user=%s password=%s dbname=%s host=%s sslmode=disable",
		dbCfg.Port, dbCfg.User, dbCfg.Password, dbCfg.Name, dbCfg.Host,
	)

	database, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("ошибка подключения к базе данных: %w", err)
	}

	// Проверка соединения с базой данных
	err = database.Ping()
	if err != nil {
		return nil, fmt.Errorf("ошибка проверки соединения с базой данных: %w", err)
	}
	log.Info("Успешное подключение к базе данных")
	return database, nil
}
