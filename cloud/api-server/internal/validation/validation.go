package validation

import (
	"errors"
	"fmt"
	"math"
	"net/mail"
	"regexp"
	"strings"
	"time"
)

var (
	errEmpty       = errors.New("field is required")
	errTooShort    = fmt.Errorf("field is too short")
	errTooLong     = fmt.Errorf("field is too long")
	errInvalidFmt  = fmt.Errorf("invalid format")
	errOutOfRange  = fmt.Errorf("value out of range")
)

// DeviceID matches BR-XXXX or PX-XXXX where X is uppercase hex.
var deviceIDRe = regexp.MustCompile(`^(BR|PX)-[0-9A-F]{4}$`)

func DeviceID(id string) error {
	if id == "" {
		return errEmpty
	}
	if len(id) > 16 {
		return fmt.Errorf("%w: device ID exceeds 16 chars", errTooLong)
	}
	if !deviceIDRe.MatchString(id) {
		return fmt.Errorf("%w: device ID must be BR-XXXX or PX-XXXX (hex)", errInvalidFmt)
	}
	return nil
}

// Name validates a human name: non-empty, trim whitespace, max length.
func Name(n string, maxLen int) error {
	n = strings.TrimSpace(n)
	if n == "" {
		return errEmpty
	}
	if len(n) > maxLen {
		return fmt.Errorf("%w: name exceeds %d chars", errTooLong, maxLen)
	}
	return nil
}

func ElderlyName(n string) error { return Name(n, 50) }
func UserName(n string) error   { return Name(n, 100) }

// Phone validates a Chinese mobile phone number (11 digits starting with 1).
var phoneRe = regexp.MustCompile(`^1[3-9]\d{9}$`)

func Phone(p string) error {
	p = strings.TrimSpace(p)
	if p == "" {
		return errEmpty
	}
	if !phoneRe.MatchString(p) {
		return fmt.Errorf("%w: invalid Chinese phone number", errInvalidFmt)
	}
	return nil
}

// Email validates email format.
func Email(e string) error {
	e = strings.TrimSpace(e)
	if e == "" {
		return errEmpty
	}
	_, err := mail.ParseAddress(e)
	if err != nil {
		return fmt.Errorf("%w: invalid email", errInvalidFmt)
	}
	return nil
}

// Password enforces minimum length.
func Password(p string, minLen int) error {
	if len(p) < minLen {
		return fmt.Errorf("%w: password must be at least %d chars", errTooShort, minLen)
	}
	return nil
}

func StrongPassword(p string) error { return Password(p, 8) }

// TimeString validates HH:MM format.
func TimeString(t string) error {
	_, err := time.Parse("15:04", t)
	if err != nil {
		return fmt.Errorf("%w: schedule_time must be HH:MM format", errInvalidFmt)
	}
	return nil
}

// DateString validates YYYY-MM-DD format and ensures date is reasonable.
func DateString(d string) error {
	date, err := time.Parse("2006-01-02", d)
	if err != nil {
		return fmt.Errorf("%w: date must be YYYY-MM-DD", errInvalidFmt)
	}
	now := time.Now()
	if date.After(now) {
		return fmt.Errorf("%w: date cannot be in the future", errOutOfRange)
	}
	// Born more than 130 years ago is unlikely
	if date.Before(now.Add(-130 * 365 * 24 * time.Hour)) {
		return fmt.Errorf("%w: date seems unreasonably old", errOutOfRange)
	}
	return nil
}

// HealthData validates health metric ranges.
func HealthData(metric string, value float64) error {
	switch metric {
	case "hr":
		if value < 20 || value > 300 {
			return fmt.Errorf("%w: heart rate must be 20-300 bpm", errOutOfRange)
		}
	case "spo2":
		if value < 50 || value > 100 {
			return fmt.Errorf("%w: SpO2 must be 50-100%%", errOutOfRange)
		}
	case "steps":
		if value < 0 || value > 200000 {
			return fmt.Errorf("%w: steps must be 0-200000", errOutOfRange)
		}
	case "sleep_hours":
		if value < 0 || value > 24 {
			return fmt.Errorf("%w: sleep hours must be 0-24", errOutOfRange)
		}
	case "bp_systolic":
		if value < 40 || value > 300 {
			return fmt.Errorf("%w: systolic BP must be 40-300 mmHg", errOutOfRange)
		}
	case "bp_diastolic":
		if value < 20 || value > 200 {
			return fmt.Errorf("%w: diastolic BP must be 20-200 mmHg", errOutOfRange)
		}
	default:
		return fmt.Errorf("%w: unknown metric %q", errInvalidFmt, metric)
	}
	return nil
}

// Location validates GPS coordinates.
func Location(lat, lon float64) error {
	if math.Abs(lat) > 90 {
		return fmt.Errorf("%w: latitude must be -90 to 90", errOutOfRange)
	}
	if math.Abs(lon) > 180 {
		return fmt.Errorf("%w: longitude must be -180 to 180", errOutOfRange)
	}
	return nil
}

// Geofence validates geofence parameters.
func Geofence(name string, lat, lon float64, radius int) error {
	if err := Name(name, 100); err != nil {
		return fmt.Errorf("name: %w", err)
	}
	if err := Location(lat, lon); err != nil {
		return fmt.Errorf("coordinates: %w", err)
	}
	if radius < 50 || radius > 50000 {
		return fmt.Errorf("%w: radius must be 50-50000 meters", errOutOfRange)
	}
	return nil
}

// Medication validates medication rule fields.
func Medication(doseCount int, pillType string) error {
	if doseCount < 1 || doseCount > 20 {
		return fmt.Errorf("%w: dose_count must be 1-20", errOutOfRange)
	}
	if pillType == "" {
		return errEmpty
	}
	if len(pillType) > 100 {
		return fmt.Errorf("%w: pill_type exceeds 100 chars", errTooLong)
	}
	return nil
}

// DaysOfWeek validates day-of-week values (1=Mon..7=Sun).
func DaysOfWeek(days []int) error {
	if len(days) == 0 {
		return errEmpty
	}
	seen := make(map[int]bool, len(days))
	for _, d := range days {
		if d < 1 || d > 7 {
			return fmt.Errorf("%w: day_of_week must be 1-7", errOutOfRange)
		}
		if seen[d] {
			return fmt.Errorf("duplicate day_of_week: %d", d)
		}
		seen[d] = true
	}
	return nil
}

// OTP validates a 6-digit numeric code.
var otpRe = regexp.MustCompile(`^\d{6}$`)

func OTP(code string) error {
	if !otpRe.MatchString(code) {
		return fmt.Errorf("%w: OTP must be exactly 6 digits", errInvalidFmt)
	}
	return nil
}

// AlertType validates alert type string.
func AlertType(t string) error {
	valid := map[string]bool{
		"sos": true, "fall": true, "med_missed": true,
		"device_offline": true, "geofence_breach": true,
	}
	if !valid[t] {
		return fmt.Errorf("%w: invalid alert_type %q", errInvalidFmt, t)
	}
	return nil
}

// HealthTier validates health tier label.
func HealthTier(tier string) error {
	valid := map[string]bool{
		"低风险": true, "中风险": true, "高风险": true,
		"low": true, "medium": true, "high": true,
	}
	if !valid[tier] {
		return fmt.Errorf("%w: health_tier must be one of 低风险/中风险/高风险 or low/medium/high", errInvalidFmt)
	}
	return nil
}

// Pagination validates page parameters.
func Pagination(page, pageSize int) error {
	if page < 1 {
		return fmt.Errorf("%w: page must be >= 1", errOutOfRange)
	}
	if pageSize < 1 || pageSize > 100 {
		return fmt.Errorf("%w: page_size must be 1-100", errOutOfRange)
	}
	return nil
}
