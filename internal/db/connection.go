package db

import (
	"database/sql"
	"fmt"
	"shop/internal/config"
)

func ConnectDB(dbCfg config.DatabaseConfig) (*sql.DB, error) {
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
	return database, nil
}
