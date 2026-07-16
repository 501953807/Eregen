package analyzer

import "time"

// now returns current time (overridable for testing).
func now() time.Time {
	return time.Now().UTC()
}
