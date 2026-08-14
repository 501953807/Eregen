package store

import (
	"context"

	"eregen.dev/b2b-community-platform/internal/model"
)

// Store is the unified interface for all database backends.
type Store interface {
	EventStore
	RegistrationStore
	HealthCheckStore
	CarePlanStore
}

// EventStore handles community events.
type EventStore interface {
	CreateEvent(ctx context.Context, evt *model.CommunityEvent) error
	ListEvents(ctx context.Context, serviceType model.ServiceType, page, pageSize int) ([]model.CommunityEvent, int, error)
	GetEventByID(ctx context.Context, id string) (*model.CommunityEvent, error)
	DeleteEvent(ctx context.Context, id string) error
}

// RegistrationStore handles event registrations.
type RegistrationStore interface {
	RegisterForEvent(ctx context.Context, reg *model.EventRegistration) error
	GetRegistrationsForEvent(ctx context.Context, eventID string) ([]model.EventRegistration, error)
	CancelEventRegistration(ctx context.Context, eventID, elderlyID string) error
	ActiveRegistrationsCount(ctx context.Context, eventID string) (int, error)
}

// HealthCheckStore handles health check records.
type HealthCheckStore interface {
	CreateHealthCheck(ctx context.Context, record *model.HealthCheckRecord) error
	GetHealthChecksForElderly(ctx context.Context, elderlyID string, limit int) ([]model.HealthCheckRecord, error)
	GetHealthCheckByID(ctx context.Context, id string) (*model.HealthCheckRecord, error)
	UpdateHealthCheck(ctx context.Context, id string, record *model.HealthCheckRecord) error
	DeleteHealthCheck(ctx context.Context, id string) error
}

// CarePlanStore handles care plans.
type CarePlanStore interface {
	CreateCarePlan(ctx context.Context, plan *model.CarePlan) error
	GetCarePlansForElderly(ctx context.Context, elderlyID string) ([]model.CarePlan, error)
	GetCarePlanByID(ctx context.Context, id string) (*model.CarePlan, error)
	UpdateCarePlan(ctx context.Context, id string, plan *model.CarePlan) error
	DeleteCarePlan(ctx context.Context, id string) error
}
