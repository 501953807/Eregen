package store

import (
	"context"
	"time"

	"eregen.dev/api-server/internal/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Domain interfaces for handler dependency injection.

type AuthStore interface {
	CreateUser(ctx context.Context, u *model.User) error
	GetUserByID(ctx context.Context, id string) (*model.User, error)
	GetUserByPhone(ctx context.Context, phone string) (*model.User, error)
	GetUserByEmail(ctx context.Context, email string) (*model.User, error)
	GetUserByOpenID(ctx context.Context, openID string) (*model.User, error)
	CreateDevice(ctx context.Context, d *model.Device) error
}

type DeviceStore interface {
	ListDevices(ctx context.Context, ownerID string, deviceType *string, page, pageSize int) ([]model.Device, int, error)
	GetDevice(ctx context.Context, deviceID string) (*model.Device, error)
	UpdateDeviceSettings(ctx context.Context, deviceID string, settings map[string]any) error
	DeleteDevice(ctx context.Context, deviceID, ownerID string) error
	AdminDeleteDevice(ctx context.Context, deviceID string) error
	BindDevice(ctx context.Context, deviceID, ownerUserID, deviceType, tier string) (*model.Device, error)
	AdminDeviceList(ctx context.Context, deviceType, tier, status string, page, pageSize int) ([]model.Device, int, error)
	CreateHealthRecord(ctx context.Context, r *model.HealthRecord) error
	CreateLocationRecord(ctx context.Context, r *model.LocationRecord) error
	CreateMedStatusRecord(ctx context.Context, r *model.MedStatusRecord) error
	CreateFirmwareRelease(ctx context.Context, r *model.FirmwareRelease) error
	ListFirmwareReleases(ctx context.Context, deviceType, tier string) ([]model.FirmwareRelease, error)
	GetFirmwareRelease(ctx context.Context, id string) (*model.FirmwareRelease, error)
	CreateOTAJob(ctx context.Context, j *model.OTAJob) error
	GetOTAJob(ctx context.Context, id string) (*model.OTAJob, error)
	UpdateOTAJobProgress(ctx context.Context, jobID string, fn UpdateOTAJobProgressFn) error
	Pool() *pgxpool.Pool
}

type ProfileStore interface {
	GetUserByID(ctx context.Context, id string) (*model.User, error)
	UpdateUser(ctx context.Context, id string, name, phone, email *string) error
	ListElderlyProfiles(ctx context.Context, userID string, page, pageSize int) ([]model.ElderlyProfile, int, error)
	CreateElderlyProfile(ctx context.Context, ep *model.ElderlyProfile) error
	GetElderlyProfile(ctx context.Context, elderlyID string) (*model.ElderlyProfile, error)
	UpdateElderlyProfile(ctx context.Context, elderlyID string, req *model.UpdateElderlyRequest) error
	GetDevice(ctx context.Context, deviceID string) (*model.Device, error)
	UpdateDeviceSettings(ctx context.Context, deviceID string, settings map[string]any) error
	Pool() *pgxpool.Pool
}

type AlertStore interface {
	ListAlerts(ctx context.Context, elderIDs []string, filter *model.AlertFilter, page, pageSize int) ([]model.Alert, int, error)
	GetAlert(ctx context.Context, alertID string) (*model.Alert, error)
}

type SessionStore interface {
	SetRefreshToken(ctx context.Context, token string, userID string, ttl time.Duration) error
	ValidateRefreshToken(ctx context.Context, token string) (string, error)
	InvalidateRefreshToken(ctx context.Context, token string) error
	SetOTP(ctx context.Context, phoneOrEmail, code string, ttl time.Duration) error
	VerifyOTP(ctx context.Context, phoneOrEmail, code string) error
	DelByPattern(ctx context.Context, pattern string) error
	SetResetToken(ctx context.Context, token, userID string, ttl time.Duration) error
	Store(ctx context.Context, key, value string, ttl time.Duration) error
	Verify(ctx context.Context, key, value string) error
}

type DeviceCacheStore interface {
	SetDeviceOnline(ctx context.Context, deviceID string) error
	IsDeviceOnline(ctx context.Context, deviceID string) (bool, error)
	InvalidateDevice(ctx context.Context, deviceID string) error
}
