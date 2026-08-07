// Package store defines the data access interfaces for admin operations.
package store

// Store is the composite interface that PostgresStore and SqliteStore implement.
// It combines all domain interfaces for the top-level dependency.
type Store interface {
	DashboardStore
	DeviceStore
	UserStore
	AuthStore
	AlertStore
	ElderlyStore
	FirmwareStore
	SettingsStore
	PatientStore
	WristbandStore
	ClinicalStore
	AdmissionStore
	RegulatoryStore
	CommunityWBStore
	SubscriptionStore
	InstitutionStore
}
