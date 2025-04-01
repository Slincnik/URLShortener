package config

import (
	"log"
	"os"
	"strconv"

	"github.com/subosito/gotenv"
)

const (
	EnvDev  = "DEV"
	EnvProd = "PROD"
)

type Config struct {
	DBType               string // sqlite, postgres
	DBPath               string // only sqlite
	Env                  string
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

	env := getEnvWithDefault("ENV", EnvProd)
	if !isValidEnv(env) {
		log.Printf("Invalid environment variable: %s. Defaulting to %s", env, EnvProd)
		env = EnvProd
	}

	return &Config{
		DBType:               getEnvWithDefault("DB_TYPE", "sqlite"),
		Env:                  env,
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

func isValidEnv(env string) bool {
	return env == EnvDev || env == EnvProd
}
