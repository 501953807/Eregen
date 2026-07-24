package config

import "os"

type Config struct {
	Port         string
	DatabaseType string // "postgres" or "sqlite"
	DatabaseURL  string
	SQLitePath   string
	RedisURL     string
	JWTSecret    string
}

func Load() *Config {
	dbType := getEnv("DATABASE_TYPE", "sqlite")
	return &Config{
		Port:         getEnv("PORT", "8085"),
		DatabaseType: dbType,
		DatabaseURL:  getEnv("DATABASE_URL", ""),
		SQLitePath:   getEnv("SQLITE_PATH", "eregen.db"),
		RedisURL:     getEnv("REDIS_URL", ""),
		// JWT_SECRET must be set in production
		JWTSecret: getEnv("JWT_SECRET", ""),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
