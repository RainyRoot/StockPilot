package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port         int
	DBPath       string
	CacheTTLSecs int
	FrontendURL  string
}

func Load() *Config {
	return &Config{
		Port:         getEnvInt("PORT", 8080),
		DBPath:       getEnv("DB_PATH", "stockpilot.db"),
		CacheTTLSecs: getEnvInt("CACHE_TTL_SECS", 60),
		FrontendURL:  getEnv("FRONTEND_URL", "http://localhost:5173"),
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
