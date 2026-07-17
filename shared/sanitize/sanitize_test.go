package sanitize

import (
	"encoding/json"
	"testing"
)

func TestMaskEmail(t *testing.T) {
	if got := MaskEmail("user@example.com"); got != "u***@example.com" {
		t.Errorf("MaskEmail = %q, want u***@example.com", got)
	}
	if got := MaskEmail(""); got != "" {
		t.Errorf("MaskEmail(\"\") = %q, want empty", got)
	}
}

func TestMaskPhone(t *testing.T) {
	if got := MaskPhone("13800138000"); got != "138****8000" {
		t.Errorf("MaskPhone = %q, want 138****8000", got)
	}
	if got := MaskPhone("123"); got == "123" {
		t.Error("short phone should be masked")
	}
}

func TestMaskToken(t *testing.T) {
	token := "eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgNryP4J3jVmNHl0w5N_XgL0n3I9PlFUP0THsR8U"
	if got := MaskToken(token); got == token {
		t.Error("long token should be masked")
	}
}

func TestSanitizePII(t *testing.T) {
	data := map[string]interface{}{
		"name":         "John",
		"phone":        "13800138000",
		"email":        "john@example.com",
		"access_token": "some_jwt_token_here",
		"settings": map[string]interface{}{
			"theme": "dark",
		},
	}
	sanitized := SanitizePII(data).(map[string]interface{})
	if sanitized["name"] != "John" {
		t.Error("non-PII field should be preserved")
	}
	if sanitized["phone"] != "***REDACTED***" {
		t.Errorf("PII field not redacted: %v", sanitized["phone"])
	}
	if sanitized["access_token"] != "***REDACTED***" {
		t.Errorf("token not redacted: %v", sanitized["access_token"])
	}
}

func TestSanitizePIINestedArray(t *testing.T) {
	data := map[string]interface{}{
		"users": []interface{}{
			map[string]interface{}{"phone": "13800138000", "name": "A"},
			map[string]interface{}{"email": "b@test.com", "name": "B"},
		},
	}
	sanitized := SanitizePII(data).(map[string]interface{})
	users := sanitized["users"].([]interface{})
	if users[0].(map[string]interface{})["phone"] != "***REDACTED***" {
		t.Error("nested PII not redacted")
	}
	if users[1].(map[string]interface{})["email"] != "***REDACTED***" {
		t.Error("nested PII not redacted")
	}
}

func TestSanitizePIIGenericHashKey(t *testing.T) {
	data := map[string]interface{}{
		"api_key_hash": "secret123",
		"signature":    "valid_sig",
	}
	sanitized := SanitizePII(data).(map[string]interface{})
	if sanitized["api_key_hash"] != "***REDACTED***" {
		t.Errorf("_hash suffix not redacted: %v", sanitized["api_key_hash"])
	}
	if sanitized["signature"] != "valid_sig" {
		t.Error("non-hash key incorrectly redacted")
	}
}

func TestSanitizePIIJSONRoundTrip(t *testing.T) {
	data := map[string]interface{}{
		"user":   "alice",
		"phone":  "13800138000",
		"token":  "my_secret_token_value",
		"levels": []interface{}{
			map[string]interface{}{"password": "hunter2", "role": "admin"},
		},
	}
	sanitized := SanitizePII(data).(map[string]interface{})
	b, err := json.Marshal(sanitized)
	if err != nil {
		t.Fatalf("JSON marshal failed: %v", err)
	}
	// Verify sensitive values are gone from JSON output
	var result map[string]interface{}
	if err := json.Unmarshal(b, &result); err != nil {
		t.Fatalf("JSON unmarshal failed: %v", err)
	}
	if result["phone"] != "***REDACTED***" {
		t.Error("phone should be redacted in JSON output")
	}
	if result["token"] != "***REDACTED***" {
		t.Error("token should be redacted in JSON output")
	}
	if result["user"] != "alice" {
		t.Error("user should be preserved in JSON output")
	}
}
