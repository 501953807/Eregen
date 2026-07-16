package model

import "time"

// CommunityRole describes what a user can do in the community platform.
type CommunityRole string

const (
	RoleAdmin       CommunityRole = "admin"
	RoleCaregiver   CommunityRole = "caregiver"
	RoleVolunteer   CommunityRole = "volunteer"
	RoleElderly     CommunityRole = "elderly"
)

// ServiceType categorizes community health services.
type ServiceType string

const (
	ServiceHomeVisit  ServiceType = "home_visit"
	ServiceHealthCheck ServiceType = "health_check"
	ServiceMedDelivery ServiceType = "med_delivery"
	ServiceEmergency  ServiceType = "emergency_response"
	ServiceRehab      ServiceType = "rehabilitation"
)

// CommunityEvent represents an organized activity at a community center.
type CommunityEvent struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	ServiceType ServiceType `json:"service_type"`
	Location    string     `json:"location"`
	StartTime   time.Time  `json:"start_time"`
	EndTime     time.Time  `json:"end_time"`
	MaxParticipants int    `json:"max_participants"`
	Status      string     `json:"status"` // scheduled, in_progress, completed, cancelled
	CreatedAt   time.Time  `json:"created_at"`
}

// EventRegistration tracks who signed up for an event.
type EventRegistration struct {
	ID         string    `json:"id"`
	EventID    string    `json:"event_id"`
	ElderlyID  string    `json:"elderly_id"`
	CaregiverID *string   `json:"caregiver_id,omitempty"` // accompanying caregiver
	Status     string    `json:"status"` // confirmed, attended, cancelled
	RegisteredAt time.Time `json:"registered_at"`
}

// HealthCheckRecord stores results from community health checks.
type HealthCheckRecord struct {
	ID          string      `json:"id"`
	ElderlyID   string      `json:"elderly_elderly_id"`
	CheckDate   time.Time   `json:"check_date"`
	BP_Systolic *float64    `json:"bp_systolic,omitempty"`
	BP_Diastolic *float64  `json:"bp_diastolic,omitempty"`
	HR          *float64    `json:"hr,omitempty"`
	SPO2        *float64    `json:"spo2,omitempty"`
	Weight      *float64    `json:"weight,omitempty"`
	Height      *float64    `json:"height,omitempty"`
	Glucose     *float64    `json:"glucose,omitempty"`
	Notes       string      `json:"notes,omitempty"`
	CheckedBy   string      `json:"checked_by"`
}

// CarePlan is a personalized care plan for an elderly person managed by the community.
type CarePlan struct {
	ID           string        `json:"id"`
	ElderlyID    string        `json:"elderly_id"`
	Title        string        `json:"title"`
	Description  string        `json:"description"`
	Tasks        []CareTask    `json:"tasks"`
	AssignedTo   string        `json:"assigned_to"` // caregiver ID
	Status       string        `json:"status"` // active, completed, paused
	StartDate    time.Time     `json:"start_date"`
	EndDate      *time.Time    `json:"end_date,omitempty"`
	CreatedAt    time.Time     `json:"created_at"`
}

// CareTask is a single task within a care plan.
type CareTask struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Type      ServiceType `json:"type"`
	Schedule  string    `json:"schedule"` // e.g. "daily", "weekly", "mon_wed_fri"
	DueTime   *time.Time `json:"due_time,omitempty"`
	Completed bool      `json:"completed"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}
