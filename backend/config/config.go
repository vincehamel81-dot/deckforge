package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port              string
	DBDriver          string
	DatabaseURL       string
	JWTSecret         string
	JWTExpiry         string
	CORSOrigin        string
	MinPlayers        int
	MaxPlayers        int
	MaxUsernameLength int
	AdminSeedUsername string
}

func Load() *Config {
	return &Config{
		Port:              getEnv("PORT", "8080"),
		DBDriver:          getEnv("DB_DRIVER", "sqlite"),
		DatabaseURL:       getEnv("DATABASE_URL", "deckforge.db"),
		JWTSecret:         getEnv("JWT_SECRET", ""),
		JWTExpiry:         getEnv("JWT_EXPIRY", "24h"),
		CORSOrigin:        getEnv("CORS_ORIGIN", "http://localhost:5173"),
		MinPlayers:        getEnvInt("MIN_PLAYERS", 2),
		MaxPlayers:        getEnvInt("MAX_PLAYERS", 8),
		MaxUsernameLength: getEnvInt("MAX_USERNAME_LENGTH", 15),
		AdminSeedUsername: getEnv("ADMIN_SEED_USERNAME", ""),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}
