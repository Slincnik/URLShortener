package config

import (
	"log"
	"os"
	"strconv"

	"github.com/subosito/gotenv"
)

type Config struct {
	DBPath               string
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
		DBPath:               getEnvWithDefault("DB_PATH", "file:urls.db?_fk=1"),
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

func getEnv(key string) string {
	value := os.Getenv(key)

	if value == "" {
		log.Fatalf("Missing environment variable: %s", key)
		return ""
	}

	return value
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		intValue, _ := strconv.Atoi(value)

		return intValue
	}
	return defaultValue
}
