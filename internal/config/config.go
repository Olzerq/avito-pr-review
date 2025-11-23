package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
)

type Config struct {
	AppPort string

	DBHost string
	DBPort string
	DBUser string
	DBPass string
	DBName string
}

func LoadConfig() Config {
	return Config{
		AppPort: validatePort(getEnv("APP_PORT", ""), "8080"),

		DBHost: getEnv("DB_HOST", "localhost"),
		DBPort: validatePort(getEnv("DB_PORT", ""), "5432"),
		DBUser: getEnv("DB_USER", "avito_user"),
		DBPass: getEnv("DB_PASS", "avito_password"),
		DBName: getEnv("DB_NAME", "avito_prs"),
	}
}

func (c Config) DBConnStr() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		c.DBUser, c.DBPass, c.DBHost, c.DBPort, c.DBName,
	)
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func validatePort(port string, def string) string {
	if port == "" {
		return def
	}

	portNum, err := strconv.Atoi(port)
	if err != nil {
		log.Printf("Ошибка: значение порта '%s' не является числом. Используется порт по умолчанию: %s",
			port, def)
		return def
	}

	// Диапазон 1-65535
	if portNum < 1 || portNum > 65535 {
		log.Printf("Ошибка: значение порта '%d' должно быть в диапазоне от 1 до 65535. Используется порт по умолчанию: %s",
			portNum, def)
		return def
	}

	if portNum < 1024 {
		log.Printf("Предупреждение: порт %d — системный. Рекомендуется использовать порт > 1024.", portNum)
	}

	return port
}
