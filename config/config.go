package config

import (
	"log"
	"os"
	"strconv"

	"github.com/subosito/gotenv"
)

type Config struct {
	DBType               string // sqlite, postgres
	DBPath               string // only sqlite
	DBHost               string
	DBPort               int
	DBUser               string
	DBPassword           string
	DBName               string
	MaxDBConns           int
	IdleDBConns          int
	MaxAttemptsCreateKey int
}

func LoadConfig(path string) (config *Config) {
	err := gotenv.Load()

	if err != nil && !os.IsNotExist(err) {
		log.Fatalf("Failed to load environment variables: %v", err)
	}

	return &Config{
		DBType:               getEnvWithDefault("DB_TYPE", "sqlite"),
		DBPath:               getEnvWithDefault("DB_PATH", "file:urls.db?_fk=1"),
		DBHost:               getEnvWithDefault("DB_HOST", "localhost"),
		DBPort:               getEnvInt("DB_PORT", 5432),
		DBUser:               getEnvWithDefault("DB_USER", "postgres"),
		DBPassword:           getEnvWithDefault("DB_PASSWORD", "postgres"),
		DBName:               getEnvWithDefault("DB_NAME", "shortener"),
		MaxDBConns:           getEnvInt("MAX_DB_CONNS", 20),
		IdleDBConns:          getEnvInt("IDLE_DB_CONNS", 10),
		MaxAttemptsCreateKey: getEnvInt("MAX_ATTEMPTS_CREATE_KEY", 5),
	}
}

func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		intValue, _ := strconv.Atoi(value)

		return intValue
	}
	return defaultValue
}
