package fcm

import (
	"os"
)

func getEnv(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}

func readFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}
