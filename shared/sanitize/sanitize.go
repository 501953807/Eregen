package sanitize

import "strings"

// MaskEmail masks an email: u***@domain.com
func MaskEmail(email string) string {
	if email == "" {
		return ""
	}
	parts := strings.SplitN(email, "@", 2)
	if len(parts) != 2 {
		return "***@***.***"
	}
	local := parts[0]
	if len(local) <= 1 {
		return "*" + "@" + parts[1]
	}
	return local[:1] + "***@" + parts[1]
}

// MaskPhone masks a phone: 138****5678
func MaskPhone(phone string) string {
	if len(phone) < 7 {
		return "***"
	}
	if len(phone) == 11 {
		return phone[:3] + "****" + phone[7:]
	}
	return phone[:3] + "****" + phone[len(phone)-4:]
}

// MaskToken masks a JWT-like token by showing only the prefix.
func MaskToken(token string) string {
	if len(token) < 16 {
		return "***"
	}
	return token[:12] + "...***"
}

// SanitizePII recursively sanitizes PII fields in a map[string]interface{}.
var piiKeys = map[string]bool{
	"phone": true, "email": true, "password": true, "PasswordHash": true,
	"token": true, "access_token": true, "refresh_token": true,
	"Password": true, "confirm_password": true, "password_hash": true,
}

func SanitizePII(data interface{}) interface{} {
	switch v := data.(type) {
	case map[string]interface{}:
		result := make(map[string]interface{}, len(v))
		for k, val := range v {
			lower := strings.ToLower(k)
			if piiKeys[lower] || strings.HasSuffix(lower, "_hash") {
				result[k] = "***REDACTED***"
			} else {
				result[k] = SanitizePII(val)
			}
		}
		return result
	case []interface{}:
		result := make([]interface{}, len(v))
		for i, item := range v {
			result[i] = SanitizePII(item)
		}
		return result
	default:
		return v
	}
}
