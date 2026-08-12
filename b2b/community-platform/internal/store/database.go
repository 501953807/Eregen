package store

import "context"

// Store is the unified interface for all database backends.
type Store interface {
	EventStore
	RegistrationStore
	HealthCheckStore
	CarePlanStore
}

// EventStore handles community events.
type EventStore interface {
	CreateEvent(ctx context.Context, evt *CommunityEvent) error
	ListEvents(ctx context.Context, serviceType ServiceType, page, pageSize int) ([]CommunityEvent, int, error)
	GetEventByID(ctx context.Context, id string) (*CommunityEvent, error)
	DeleteEvent(ctx context.Context, id string) error
}

// RegistrationStore handles event registrations.
type RegistrationStore interface {
	RegisterForEvent(ctx context.Context, reg *EventRegistration) error
	GetRegistrationsForEvent(ctx context.Context, eventID string) ([]EventRegistration, error)
	CancelEventRegistration(ctx context.Context, eventID, elderlyID string) error
	ActiveRegistrationsCount(ctx context.Context, eventID string) (int, error)
}

// HealthCheckStore handles health check records.
type HealthCheckStore interface {
	CreateHealthCheck(ctx context.Context, record *HealthCheckRecord) error
	GetHealthChecksForElderly(ctx context.Context, elderlyID string, limit int) ([]HealthCheckRecord, error)
	GetHealthCheckByID(ctx context.Context, id string) (*HealthCheckRecord, error)
	UpdateHealthCheck(ctx context.Context, id string, record *HealthCheckRecord) error
	DeleteHealthCheck(ctx context.Context, id string) error
}

// CarePlanStore handles care plans.
type CarePlanStore interface {
	CreateCarePlan(ctx context.Context, plan *CarePlan) error
	GetCarePlansForElderly(ctx context.Context, elderlyID string) ([]CarePlan, error)
	GetCarePlanByID(ctx context.Context, id string) (*CarePlan, error)
	UpdateCarePlan(ctx context.Context, id string, plan *CarePlan) error
	DeleteCarePlan(ctx context.Context, id string) error
}
