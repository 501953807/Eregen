package fcm

import (
	"os"
	"testing"
)

func TestGetEnv_ReturnsFallback(t *testing.T) {
	os.Unsetenv("FCM_KEY_PATH_TEST")
	got := getEnv("FCM_KEY_PATH_TEST", "fallback")
	if got != "fallback" {
		t.Errorf("getEnv() = %q, want fallback", got)
	}
}

func TestGetEnv_ReturnsValue(t *testing.T) {
	os.Setenv("FCM_KEY_PATH_TEST", "my-key-path")
	defer os.Unsetenv("FCM_KEY_PATH_TEST")
	got := getEnv("FCM_KEY_PATH_TEST", "fallback")
	if got != "my-key-path" {
		t.Errorf("getEnv() = %q, want my-key-path", got)
	}
}

func TestReadFile_NonExistent(t *testing.T) {
	_, err := readFile("/nonexistent/path/key.json")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}
