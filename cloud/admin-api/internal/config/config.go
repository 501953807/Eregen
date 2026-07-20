package config

import "os"

type Config struct {
	Port        string
	DatabaseURL string
	RedisURL    string
	JWTSecret   string
}

func Load() *Config {
	return &Config{
		Port:        getEnv("PORT", "8085"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://localhost/eregen"),
		RedisURL:    getEnv("REDIS_URL", "redis://localhost:6379"),
		// JWT_SECRET must be set in production
		JWTSecret:   getEnv("JWT_SECRET", ""),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
