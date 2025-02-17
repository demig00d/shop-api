package config

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type (
	// Config содержит конфигурацию приложения.
	Config struct {
		Database DatabaseConfig
		JWT      JWTConfig
		LogLevel string `env:"LOG_LEVEL" env-default:"INFO"`
	}

	// DatabaseConfig содержит конфигурацию базы данных.
	DatabaseConfig struct {
		Host     string `env:"DATABASE_HOST"`
		Port     string `env:"DATABASE_PORT" env-default:"5432"`
		User     string `env:"DATABASE_USER" env-default:"shop"`
		Password string `env:"DATABASE_PASSWORD" env-default:"shop"`
		Name     string `env:"DATABASE_NAME" env-default:"shop"`
	}

	// JWTConfig содержит конфигурацию JWT.
	JWTConfig struct {
		SecretKey string `env:"JWT_SECRET_KEY" env-default:"secret"`
	}
)

// LoadConfig загружает конфигурацию из переменных окружения и .env файла.
func LoadConfig() (Config, error) {
	var errFile error

	cfg := &Config{}
	// Читаем конфигурацию из переменных окружения
	errEnv := cleanenv.ReadEnv(cfg)

	if errEnv == nil {
		return *cfg, nil
	}

	// Если переменные окружения не заданы, читаем из .env файла
	if errEnv != nil {
		errFile = cleanenv.ReadConfig(".env", cfg)
	}

	if errFile != nil {
		return *cfg, errors.Join(errEnv, errFile)
	}

	return *cfg, nil
}

// LoadConfigFrom загружает конфигурацию из указанного файла .env.
func LoadConfigFrom(envFile string) (Config, error) {
	cfg := &Config{}
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(dir)
	err = cleanenv.ReadConfig(envFile, cfg)
	if err != nil {
		return *cfg, fmt.Errorf("ошибка чтения конфигурации из файла: %w", err)
	}
	return *cfg, nil
}
