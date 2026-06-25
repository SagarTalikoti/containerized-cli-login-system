package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	DBPath            string
	SessionTimeout    time.Duration
	MaxFailedAttempts int
	LockoutDuration   time.Duration
}

func Load() Config {
	return Config{
		DBPath:            envString("DB_PATH", "./data/app.db"),
		SessionTimeout:    time.Duration(envInt("SESSION_TIMEOUT_MINUTES", 30)) * time.Minute,
		MaxFailedAttempts: envInt("MAX_FAILED_ATTEMPTS", 5),
		LockoutDuration:   time.Duration(envInt("LOCKOUT_MINUTES", 15)) * time.Minute,
	}
}

func envString(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func envInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return fallback
	}

	return parsed
}
