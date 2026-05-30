package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Port               string
	DBDriver           string
	DatabaseURL        string
	JWTSecret          string
	JWTExpiry          string
	CORSOrigin         string
	MinPlayers         int
	MaxPlayers         int
	MaxUsernameLength  int
	AdminSeedUsernames []string
}

func Load() *Config {
	return &Config{
		Port:               getEnv("PORT", "8080"),
		DBDriver:           getEnv("DB_DRIVER", "sqlite"),
		DatabaseURL:        getEnv("DATABASE_URL", "deckforge.db"),
		JWTSecret:          getEnv("JWT_SECRET", ""),
		JWTExpiry:          getEnv("JWT_EXPIRY", "24h"),
		CORSOrigin:         getEnv("CORS_ORIGIN", "http://localhost:5173"),
		MinPlayers:         getEnvInt("MIN_PLAYERS", 2),
		MaxPlayers:         getEnvInt("MAX_PLAYERS", 8),
		MaxUsernameLength:  getEnvInt("MAX_USERNAME_LENGTH", 15),
		AdminSeedUsernames: parseCSV(getEnv("ADMIN_SEED_USERNAMES", "")),
	}
}

// parseCSV splits a comma-separated string into a trimmed slice, skipping empty entries.
func parseCSV(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
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
