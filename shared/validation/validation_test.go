package validation

import (
	"strings"
	"testing"
)

func TestValidateEmail(t *testing.T) {
	tests := []struct{ input, errVal string }{
		{"user@example.com", ""},
		{"invalid", ErrInvalidEmail.Error()},
		{"@no.com", ErrInvalidEmail.Error()},
	}
	for _, tt := range tests {
		err := ValidateEmail(tt.input)
		if tt.errVal == "" && err != nil {
			t.Errorf("ValidateEmail(%q) unexpected error: %v", tt.input, err)
		} else if tt.errVal != "" && err == nil {
			t.Errorf("ValidateEmail(%q) expected error %q", tt.input, tt.errVal)
		}
	}
}

func TestValidatePhone(t *testing.T) {
	if err := ValidatePhone("13800138000"); err != nil {
		t.Errorf("valid phone rejected: %v", err)
	}
	if err := ValidatePhone("12345"); err == nil {
		t.Error("invalid phone accepted")
	}
}

func TestValidatePagination(t *testing.T) {
	page, size, err := ValidatePagination(0, 200, 100)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if page != 1 || size != 100 {
		t.Errorf("expected page=1, size=100, got page=%d, size=%d", page, size)
	}
}

func TestValidateFloatRange(t *testing.T) {
	if err := ValidateFloatRange(5.0, 0, 10); err != nil {
		t.Errorf("valid range rejected: %v", err)
	}
	if err := ValidateFloatRange(100, 0, 10); err == nil {
		t.Error("out of range accepted")
	}
}

func TestValidateEnum(t *testing.T) {
	allowed := []string{"P0", "P1", "P2"}
	if err := ValidateEnum("P0", allowed); err != nil {
		t.Errorf("valid enum rejected: %v", err)
	}
	if err := ValidateEnum("P3", allowed); err == nil {
		t.Error("invalid enum accepted")
	}
}

func TestSanitizeHTML(t *testing.T) {
	input := `<script>alert("xss")</script>Hello`
	got := SanitizeHTML(input)
	if strings.Contains(got, "<") || strings.Contains(got, ">") {
		t.Errorf("tags not stripped: %s", got)
	}
}

func TestSanitizeURL(t *testing.T) {
	if _, err := SanitizeURL("https://safe.example.com/path"); err != nil {
		t.Errorf("valid HTTPS URL rejected: %v", err)
	}
	if _, err := SanitizeURL("http://insecure.com/path"); err == nil {
		t.Error("HTTP URL accepted")
	}
}
