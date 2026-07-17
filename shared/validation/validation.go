package validation

import (
	"errors"
	"fmt"
	"net/mail"
	"net/url"
	"regexp"
	"strings"
)

var (
	ErrInvalidEmail     = errors.New("invalid email address")
	ErrInvalidPhone     = errors.New("invalid phone number, must be 11-digit Chinese mobile")
	ErrInvalidPage      = errors.New("page must be >= 1")
	ErrInvalidPageSize  = errors.New("page_size must be >= 1 and <= 100")
	ErrOutOfRange       = errors.New("value out of range")
	ErrInvalidEnum      = errors.New("value not in allowed list")
	ErrInvalidLatitude  = errors.New("latitude must be between -90 and 90")
	ErrInvalidLongitude = errors.New("longitude must be between -180 and 180")
)

// ValidateEmail checks email format.
func ValidateEmail(e string) error {
	_, err := mail.ParseAddress(e)
	if err != nil {
		return ErrInvalidEmail
	}
	return nil
}

// ValidatePhone checks 11-digit Chinese mobile number.
func ValidatePhone(p string) error {
	p = strings.TrimSpace(p)
	matched, _ := regexp.MatchString(`^1[3-9]\d{9}$`, p)
	if !matched {
		return ErrInvalidPhone
	}
	return nil
}

// ValidatePagination clamps and validates pagination params.
// Returns (page, pageSize, error).
func ValidatePagination(page, pageSize, maxPageSize int) (int, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > maxPageSize {
		pageSize = maxPageSize
	}
	if page > 10000 {
		return 0, 0, fmt.Errorf("page %d exceeds maximum 10000", page)
	}
	return page, pageSize, nil
}

// ValidateFloatRange checks val is within [min, max].
func ValidateFloatRange(val, min, max float64) error {
	if val < min || val > max {
		return fmt.Errorf("%w: got %v, expected [%v, %v]", ErrOutOfRange, val, min, max)
	}
	return nil
}

// ValidateEnum checks val is in the allowed list.
func ValidateEnum(val string, allowed []string) error {
	for _, a := range allowed {
		if val == a {
			return nil
		}
	}
	return fmt.Errorf("%w: %q not in %v", ErrInvalidEnum, val, allowed)
}

// ValidateLatitude checks latitude is in [-90, 90].
func ValidateLatitude(lat float64) error {
	return ValidateFloatRange(lat, -90, 90)
}

// ValidateLongitude checks longitude is in [-180, 180].
func ValidateLongitude(lon float64) error {
	return ValidateFloatRange(lon, -180, 180)
}

// SanitizeHTML strips HTML tags from input to prevent XSS.
func SanitizeHTML(s string) string {
	result := strings.Map(func(r rune) rune {
		if r == '<' || r == '>' || r == '"' || r == '\'' {
			return -1
		}
		return r
	}, s)
	return strings.TrimSpace(result)
}

// SanitizeURL validates and returns a safe URL.
func SanitizeURL(raw string) (string, error) {
	parsed, err := url.ParseRequestURI(raw)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}
	if parsed.Scheme != "https" {
		return "", fmt.Errorf("only HTTPS URLs allowed")
	}
	return parsed.String(), nil
}
