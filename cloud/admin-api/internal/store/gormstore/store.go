package gormstore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"eregen.dev/admin-api/internal/auth"
	"eregen.dev/admin-api/internal/model"
	"eregen.dev/admin-api/internal/store/models"
)

type Store struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Store {
	return &Store{db: db}
}

func NewFromDSN(dbType, dsn string) (*Store, error) {
	var db *gorm.DB
	var err error
	switch dbType {
	case "postgres":
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Info)})
	case "sqlite":
		db, err = gorm.Open(sqlite.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Info)})
	default:
		return nil, fmt.Errorf("unsupported database type: %s", dbType)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	if dbType == "sqlite" {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.SetMaxOpenConns(1)
		}
	}
	return New(db), nil
}

func nullableString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func nullableStringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func ptrString(s *string) *string {
	if s == nil {
		return nil
	}
	return s
}

func (s *Store) AutoMigrate() error {
	return s.db.AutoMigrate(
		&models.User{}, &models.ElderlyProfile{}, &models.Device{},
		&models.HealthRecord{}, &models.MedicationRule{}, &models.LocationHistory{},
		&models.Alert{}, &models.Subscription{}, &models.FirmwareRelease{},
		&models.Person{}, &models.HospitalAdmission{}, &models.MedicalWristbandPatient{},
		&models.RegulatoryFenceConfig{}, &models.AlertRule{},
		&models.APIKey{}, &models.SystemSetting{}, &models.OTAJob{},
	)
}

func (s *Store) DB() *gorm.DB { return s.db }

func (s *Store) CreateUser(ctx context.Context, name, email, phone, role, password string) (string, error) {
	hash, err := auth.HashPassword(password)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	id := uuid.New().String()
	user := &models.User{BaseModel: models.BaseModel{ID: id}, Name: name, Email: email, Phone: phone, Role: role, PasswordHash: hash}
	if err := s.db.WithContext(ctx).Create(user).Error; err != nil {
		return "", fmt.Errorf("create user: %w", err)
	}
	return id, nil
}

func (s *Store) GetUserByCredential(ctx context.Context, method, credential, secret string) (*model.UserLogin, error) {
	var user models.User
	var err error
	switch method {
	case "email":
		err = s.db.WithContext(ctx).Where("email = ?", credential).First(&user).Error
	case "phone":
		err = s.db.WithContext(ctx).Where("phone = ?", credential).First(&user).Error
	default:
		return nil, errors.New("invalid method")
	}
	if err != nil {
		return nil, err
	}
	if !auth.ComparePassword(secret, user.PasswordHash) {
		return nil, errors.New("invalid credentials")
	}
	return &model.UserLogin{ID: user.ID, Name: user.Name, Role: user.Role}, nil
}

func (s *Store) ListUsers(ctx context.Context, page, pageSize int, role string) ([]model.UserSummary, error) {
	var users []models.User
	query := s.db.WithContext(ctx).Model(&models.User{})
	if role != "" {
		query = query.Where("role = ?", role)
	}
	query = query.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize)
	if err := query.Find(&users).Error; err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	result := make([]model.UserSummary, len(users))
	for i, u := range users {
		var deviceCount int64
		s.db.WithContext(ctx).Model(&models.Device{}).Where("owner_user_id = ?", u.ID).Count(&deviceCount)
		result[i] = model.UserSummary{ID: u.ID, Name: u.Name, Role: u.Role, CreatedAt: u.CreatedAt, Devices: int(deviceCount)}
	}
	return result, nil
}

func (s *Store) UpdateUser(ctx context.Context, id, name, email, phone, role string) error {
	return s.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", id).Updates(map[string]interface{}{"name": name, "email": email, "phone": phone, "role": role}).Error
}

func (s *Store) SetUserRole(ctx context.Context, userID, role string) error {
	return s.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", userID).Update("role", role).Error
}

func (s *Store) DeleteUser(ctx context.Context, id string) error {
	return s.db.WithContext(ctx).Delete(&models.User{}, "id = ?", id).Error
}

func (s *Store) CreateAlert(ctx context.Context, a *model.AlertSummary) error {
	id := uuid.New().String()
	alert := &models.Alert{BaseModel: models.BaseModel{ID: id}, ElderlyID: a.ElderlyID, AlertType: a.AlertType, Severity: a.Severity, Status: a.Status, DeviceID: a.DeviceID}
	if err := s.db.WithContext(ctx).Create(alert).Error; err != nil {
		return fmt.Errorf("create alert: %w", err)
	}
	a.ID = id
	a.CreatedAt = alert.CreatedAt
	return nil
}

func (s *Store) ListAlerts(ctx context.Context, severity, status string, limit int) ([]model.AlertSummary, error) {
	var alerts []models.Alert
	query := s.db.WithContext(ctx).Model(&models.Alert{})
	if severity != "" {
		query = query.Where("severity = ?", severity)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	query = query.Order("created_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if err := query.Find(&alerts).Error; err != nil {
		return nil, fmt.Errorf("list alerts: %w", err)
	}
	result := make([]model.AlertSummary, len(alerts))
	for i, a := range alerts {
		result[i] = model.AlertSummary{ID: a.ID, ElderlyID: a.ElderlyID, AlertType: a.AlertType, Severity: a.Severity, Status: a.Status, CreatedAt: a.CreatedAt, DeviceID: a.DeviceID}
	}
	return result, nil
}

func (s *Store) ResolveAlert(ctx context.Context, alertID string) error {
	return s.db.WithContext(ctx).Model(&models.Alert{}).Where("id = ?", alertID).Updates(map[string]interface{}{"status": "resolved", "updated_at": time.Now()}).Error
}

func (s *Store) UpdateAlertStatus(ctx context.Context, alertID, status string) error {
	return s.db.WithContext(ctx).Model(&models.Alert{}).Where("id = ?", alertID).Update("status", status).Error
}

func (s *Store) CreateElderly(ctx context.Context, name, birthDate, userID string, healthTiers []string, avatarURL string) (*model.ElderlyProfile, error) {
	tiersJSON, _ := json.Marshal(healthTiers)
	id := uuid.New().String()
	ep := &models.ElderlyProfile{BaseModel: models.BaseModel{ID: id}, UserID: userID, Name: name, BirthDate: parseTime(birthDate), AvatarURL: &avatarURL, HealthTiers: string(tiersJSON)}
	if err := s.db.WithContext(ctx).Create(ep).Error; err != nil {
		return nil, fmt.Errorf("create elderly: %w", err)
	}
	var tiers []string
	json.Unmarshal([]byte(ep.HealthTiers), &tiers)
	return &model.ElderlyProfile{ID: ep.ID, UserID: ep.UserID, Name: ep.Name, BirthDate: ep.BirthDate, AvatarURL: ep.AvatarURL, HealthTiers: tiers, CreatedAt: ep.CreatedAt, UpdatedAt: ep.UpdatedAt}, nil
}

func (s *Store) GetElderly(ctx context.Context, id string) (*model.ElderlyProfile, error) {
	var ep models.ElderlyProfile
	if err := s.db.WithContext(ctx).Where("id = ?", id).First(&ep).Error; err != nil {
		return nil, err
	}
	var tiers []string
	json.Unmarshal([]byte(ep.HealthTiers), &tiers)
	return &model.ElderlyProfile{ID: ep.ID, UserID: ep.UserID, Name: ep.Name, BirthDate: ep.BirthDate, AvatarURL: ep.AvatarURL, HealthTiers: tiers, CreatedAt: ep.CreatedAt, UpdatedAt: ep.UpdatedAt}, nil
}

func (s *Store) ListElderly(ctx context.Context, page, pageSize int) ([]model.ElderlyProfile, error) {
	var eps []models.ElderlyProfile
	query := s.db.WithContext(ctx).Model(&models.ElderlyProfile{}).Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize)
	if err := query.Find(&eps).Error; err != nil {
		return nil, fmt.Errorf("list elderly: %w", err)
	}
	result := make([]model.ElderlyProfile, len(eps))
	for i, ep := range eps {
		var tiers []string
		json.Unmarshal([]byte(ep.HealthTiers), &tiers)
		result[i] = model.ElderlyProfile{ID: ep.ID, UserID: ep.UserID, Name: ep.Name, BirthDate: ep.BirthDate, AvatarURL: ep.AvatarURL, HealthTiers: tiers, CreatedAt: ep.CreatedAt, UpdatedAt: ep.UpdatedAt}
	}
	return result, nil
}

func (s *Store) UpdateElderly(ctx context.Context, id, name, birthDate, userID string, healthTiers []string, avatarURL string) (*model.ElderlyProfile, error) {
	tiersJSON, _ := json.Marshal(healthTiers)
	if err := s.db.WithContext(ctx).Model(&models.ElderlyProfile{}).Where("id = ?", id).Updates(map[string]interface{}{"name": name, "birth_date": birthDate, "user_id": userID, "health_tiers": string(tiersJSON), "avatar_url": avatarURL, "updated_at": time.Now()}).Error; err != nil {
		return nil, fmt.Errorf("update elderly: %w", err)
	}
	return s.GetElderly(ctx, id)
}

func (s *Store) DeleteElderly(ctx context.Context, id string) error {
	return s.db.WithContext(ctx).Delete(&models.ElderlyProfile{}, "id = ?", id).Error
}

func (s *Store) ListDevices(ctx context.Context, page, pageSize int, status, devType, tier string) ([]model.DeviceSummary, error) {
	var devices []models.Device
	query := s.db.WithContext(ctx).Model(&models.Device{})
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if devType != "" {
		query = query.Where("device_type = ?", devType)
	}
	if tier != "" {
		query = query.Where("tier = ?", tier)
	}
	query = query.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize)
	if err := query.Find(&devices).Error; err != nil {
		return nil, fmt.Errorf("list devices: %w", err)
	}
	result := make([]model.DeviceSummary, len(devices))
	for i, d := range devices {
		result[i] = model.DeviceSummary{ID: d.ID, DeviceID: d.DeviceID, Type: d.DeviceType, Tier: d.Tier, Status: d.Status, LastSeen: optionalTime(d.LastSeen)}
	}
	return result, nil
}

func (s *Store) GetDeviceByID(ctx context.Context, id string) (*model.DeviceDetail, error) {
	var d models.Device
	if err := s.db.WithContext(ctx).Where("id = ?", id).First(&d).Error; err != nil {
		return nil, fmt.Errorf("get device: %w", err)
	}
	return &model.DeviceDetail{ID: d.ID, DeviceID: d.DeviceID, Type: d.DeviceType, Tier: d.Tier, Status: d.Status, LastSeen: optionalTime(d.LastSeen), OwnerName: "", FirmwareVer: extractFirmwareVer(d.Settings)}, nil
}

func (s *Store) UpdateDeviceConfig(ctx context.Context, deviceID string, config map[string]interface{}) error {
	settingsJSON, _ := json.Marshal(config)
	return s.db.WithContext(ctx).Model(&models.Device{}).Where("id = ?", deviceID).Update("settings", string(settingsJSON)).Error
}

func (s *Store) CreateHealthRecord(ctx context.Context, r *model.HealthRecordRow) error {
	id := uuid.New().String()
	record := &models.HealthRecord{BaseModel: models.BaseModel{ID: id}, ElderlyID: r.ElderlyID, HR: r.HR, SpO2: r.SpO2, Steps: r.Steps, SleepHours: r.SleepHours, Timestamp: r.Timestamp}
	if err := s.db.WithContext(ctx).Create(record).Error; err != nil {
		return fmt.Errorf("create health record: %w", err)
	}
	r.ID = id
	return nil
}

func (s *Store) GetElderlyHealthRecords(ctx context.Context, elderlyID string, limit int) ([]model.HealthRecordRow, error) {
	var records []models.HealthRecord
	query := s.db.WithContext(ctx).Where("elderly_id = ?", elderlyID).Order("timestamp DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if err := query.Find(&records).Error; err != nil {
		return nil, fmt.Errorf("get health records: %w", err)
	}
	result := make([]model.HealthRecordRow, len(records))
	for i, r := range records {
		result[i] = model.HealthRecordRow{ID: r.ID, ElderlyID: r.ElderlyID, HR: r.HR, SpO2: r.SpO2, Steps: r.Steps, SleepHours: r.SleepHours, Timestamp: r.Timestamp}
	}
	return result, nil
}

func (s *Store) CreateMedicationRule(ctx context.Context, elderlyID string, rule *model.MedicationRuleRow) error {
	id := uuid.New().String()
	r := &models.MedicationRule{BaseModel: models.BaseModel{ID: id}, ElderlyID: elderlyID, ScheduleTime: rule.ScheduleTime, PillType: rule.PillType, DoseCount: rule.DoseCount, Active: rule.Active}
	if err := s.db.WithContext(ctx).Create(r).Error; err != nil {
		return fmt.Errorf("create medication rule: %w", err)
	}
	rule.ID = id
	return nil
}

func (s *Store) GetElderlyMedicationRules(ctx context.Context, elderlyID string) ([]model.MedicationRuleRow, error) {
	var rules []models.MedicationRule
	if err := s.db.WithContext(ctx).Where("elderly_id = ?", elderlyID).Find(&rules).Error; err != nil {
		return nil, fmt.Errorf("get medication rules: %w", err)
	}
	result := make([]model.MedicationRuleRow, len(rules))
	for i, r := range rules {
		result[i] = model.MedicationRuleRow{ID: r.ID, ElderlyID: r.ElderlyID, ScheduleTime: r.ScheduleTime, PillType: r.PillType, DoseCount: r.DoseCount, Active: r.Active}
	}
	return result, nil
}

func (s *Store) UpdateMedicationRule(ctx context.Context, elderlyID, ruleID string, updates map[string]interface{}) error {
	return s.db.WithContext(ctx).Model(&models.MedicationRule{}).Where("elderly_id = ? AND id = ?", elderlyID, ruleID).Updates(updates).Error
}

func (s *Store) DeleteMedicationRule(ctx context.Context, elderlyID, ruleID string) error {
	return s.db.WithContext(ctx).Delete(&models.MedicationRule{}, "id = ? AND elderly_id = ?", ruleID, elderlyID).Error
}

func (s *Store) CreateLocation(ctx context.Context, loc *model.LocationPoint) error {
	id := uuid.New().String()
	l := &models.LocationHistory{BaseModel: models.BaseModel{ID: id}, ElderlyID: loc.ElderlyID, Lat: loc.Lat, Lng: loc.Lon, Accuracy: loc.Accuracy, Timestamp: loc.Timestamp}
	if err := s.db.WithContext(ctx).Create(l).Error; err != nil {
		return fmt.Errorf("create location: %w", err)
	}
	loc.ID = id
	return nil
}

func (s *Store) GetElderlyLocationHistory(ctx context.Context, elderlyID string, limit int) ([]model.LocationPoint, error) {
	var locations []models.LocationHistory
	query := s.db.WithContext(ctx).Where("elderly_id = ?", elderlyID).Order("timestamp DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if err := query.Find(&locations).Error; err != nil {
		return nil, fmt.Errorf("get location history: %w", err)
	}
	result := make([]model.LocationPoint, len(locations))
	for i, l := range locations {
		result[i] = model.LocationPoint{ID: l.ID, ElderlyID: l.ElderlyID, Lat: l.Lat, Lon: l.Lng, Accuracy: l.Accuracy, Timestamp: l.Timestamp}
	}
	return result, nil
}

func (s *Store) GetElderlyHealthStats(ctx context.Context, elderlyID string) (*model.HealthStats, error) {
	var stats model.HealthStats
	stats.ElderlyID = elderlyID
	// TODO: Use Raw SQL for aggregate query (GORM has limited aggregate support)
	return &stats, nil
}

func (s *Store) GetElderlyDevices(ctx context.Context, elderlyID string) ([]model.DeviceSummaryRow, error) {
	var devices []models.Device
	if err := s.db.WithContext(ctx).Table("devices").
		Joins("JOIN elderly_devices ON devices.id = elderly_devices.device_id").
		Where("elderly_devices.elderly_id = ?", elderlyID).
		Order("devices.last_seen DESC").
		Find(&devices).Error; err != nil {
		return nil, fmt.Errorf("list elderly devices: %w", err)
	}
	result := make([]model.DeviceSummaryRow, len(devices))
	for i, d := range devices {
		result[i] = model.DeviceSummaryRow{
			ID: d.ID, DeviceID: d.DeviceID, Type: d.DeviceType, Tier: d.Tier,
			Status: d.Status, FirmwareVer: extractFirmwareVer(d.Settings),
			LastSeen: optionalTime(d.LastSeen),
		}
	}
	return result, nil
}

func (s *Store) GetElderlyAlertHistory(ctx context.Context, elderlyID string, limit int) ([]model.AlertSummaryRow, error) {
	var alerts []models.Alert
	query := s.db.WithContext(ctx).Where("elderly_id = ?", elderlyID).Order("created_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if err := query.Find(&alerts).Error; err != nil {
		return nil, fmt.Errorf("list elderly alerts: %w", err)
	}
	result := make([]model.AlertSummaryRow, len(alerts))
	for i, a := range alerts {
		result[i] = model.AlertSummaryRow{ID: a.ID, ElderlyID: a.ElderlyID, AlertType: a.AlertType, Severity: a.Severity, Status: a.Status, CreatedAt: a.CreatedAt}
	}
	return result, nil
}

func (s *Store) TriggerOTA(ctx context.Context, deviceID, firmwareURL, sha256Hash string) error {
	return s.db.WithContext(ctx).Model(&models.Device{}).Where("device_id = ?", deviceID).
		Updates(map[string]interface{}{"ota_url": firmwareURL, "ota_hash": sha256Hash, "ota_status": "pending"}).Error
}

func (s *Store) UnbindDevice(ctx context.Context, deviceID string) error {
	if err := s.db.WithContext(ctx).Table("elderly_devices").Where("device_id = ?", deviceID).Delete(&models.Device{}).Error; err != nil {
		return fmt.Errorf("unbind elderly devices: %w", err)
	}
	return s.db.WithContext(ctx).Model(&models.Device{}).Where("device_id = ?", deviceID).Update("owner_user_id", nil).Error
}

func (s *Store) BatchTriggerOTA(ctx context.Context, deviceIDs, firmwareURL, sha256Hash []string) error {
	for i, id := range deviceIDs {
		url := firmwareURL[i%len(firmwareURL)]
		hash := sha256Hash[i%len(sha256Hash)]
		if err := s.TriggerOTA(ctx, id, url, hash); err != nil {
			return fmt.Errorf("batch OTA device %s: %w", id, err)
		}
	}
	return nil
}

func parseTime(s string) *time.Time {
	if s == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return nil
	}
	return &t
}

func optionalTime(t *time.Time) time.Time {
	if t == nil {
		return time.Time{}
	}
	return *t
}

func extractFirmwareVer(settings string) string {
	var cfg map[string]interface{}
	if err := json.Unmarshal([]byte(settings), &cfg); err != nil {
		return "v0.1"
	}
	if ver, ok := cfg["fw_version"].(string); ok {
		return ver
	}
	return "v0.1"
}

func (s *Store) CreateFirmwareVersion(ctx context.Context, v *model.FirmwareVersion) error {
	fr := &models.FirmwareRelease{
		DeviceType: v.DeviceType, Tier: v.Tier, Version: v.Version,
		URL: v.DownloadURL, Changelog: v.Changelog, ReleasedAt: time.Now(),
	}
	return s.db.WithContext(ctx).Create(fr).Error
}

func (s *Store) ListFirmwareVersions(ctx context.Context) ([]model.FirmwareVersion, error) {
	var releases []models.FirmwareRelease
	if err := s.db.WithContext(ctx).Order("created_at DESC").Find(&releases).Error; err != nil {
		return nil, fmt.Errorf("list firmware: %w", err)
	}
	result := make([]model.FirmwareVersion, len(releases))
	for i, r := range releases {
		result[i] = model.FirmwareVersion{ID: r.ID, DeviceType: r.DeviceType, Tier: r.Tier, Version: r.Version,
			DownloadURL: r.URL, Changelog: r.Changelog, ReleaseDate: r.ReleasedAt}
	}
	return result, nil
}

func (s *Store) DeleteFirmwareVersion(ctx context.Context, id string) error {
	return s.db.WithContext(ctx).Model(&models.FirmwareRelease{}).Where("id = ?", id).Delete(&models.FirmwareRelease{}).Error
}

func (s *Store) PushOTAJob(ctx context.Context, firmwareID string, deviceIDs []string) (string, error) {
	devicesJSON, _ := json.Marshal(deviceIDs)
	job := &models.OTAJob{FirmwareID: firmwareID, TargetDevices: string(devicesJSON), Progress: "{}"}
	if err := s.db.WithContext(ctx).Create(job).Error; err != nil {
		return "", err
	}
	return job.ID, nil
}

func (s *Store) GetOTAJob(ctx context.Context, jobID string) (*model.OTAJob, error) {
	var gormJob models.OTAJob
	if err := s.db.WithContext(ctx).Where("id = ?", jobID).First(&gormJob).Error; err != nil {
		return nil, err
	}
	return &model.OTAJob{
		ID:            gormJob.ID,
		FirmwareID:    gormJob.FirmwareID,
		TargetDevices: json.RawMessage(gormJob.TargetDevices),
		Progress:      json.RawMessage(gormJob.Progress),
		CreatedAt:     gormJob.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     gormJob.UpdatedAt.Format(time.RFC3339),
	}, nil
}

func (s *Store) GetNotificationSettings(ctx context.Context) (map[string]any, error) {
	var setting models.SystemSetting
	if err := s.db.WithContext(ctx).Where("key = 'notification'").First(&setting).Error; err != nil {
		return map[string]any{}, nil
	}
	var result map[string]any
	json.Unmarshal([]byte(setting.SettingValue), &result)
	if result == nil {
		result = map[string]any{}
	}
	return result, nil
}

func (s *Store) UpdateNotificationSettings(ctx context.Context, data map[string]any) error {
	value, _ := json.Marshal(data)
	return s.db.WithContext(ctx).Model(&models.SystemSetting{}).
		Where("key = ?", "notification").
		Updates(map[string]interface{}{"setting_value": string(value)}).Error
}

func (s *Store) ListAPIKeys(ctx context.Context) ([]model.APIKeySummary, error) {
	var keys []models.APIKey
	if err := s.db.WithContext(ctx).Order("created_at DESC").Find(&keys).Error; err != nil {
		return nil, fmt.Errorf("list api keys: %w", err)
	}
	result := make([]model.APIKeySummary, len(keys))
	for i, k := range keys {
		result[i] = model.APIKeySummary{ID: k.ID, Name: k.Name, Active: k.Active, CreatedAt: k.CreatedAt}
	}
	return result, nil
}

func (s *Store) CreateAPIKey(ctx context.Context, name, keyHash string, expiresAt *time.Time) (string, error) {
	id := uuid.New().String()
	key := &models.APIKey{BaseModel: models.BaseModel{ID: id}, Name: name, KeyHash: keyHash, ExpiresAt: expiresAt, Active: true}
	if err := s.db.WithContext(ctx).Create(key).Error; err != nil {
		return "", fmt.Errorf("create api key: %w", err)
	}
	return id, nil
}

func (s *Store) RevokeAPIKey(ctx context.Context, id string) error {
	return s.db.WithContext(ctx).Model(&models.APIKey{}).Where("id = ?", id).Update("active", false).Error
}

func (s *Store) ChangeAdminPassword(ctx context.Context, userID, hash string) error {
	return s.db.WithContext(ctx).Model(&models.User{}).Where("id = ? AND role = 'admin'", userID).Update("password_hash", hash).Error
}

func (s *Store) GetDashboardStats(ctx context.Context) (*model.DashboardStats, error) {
	var stats model.DashboardStats
	var count int64
	s.db.WithContext(ctx).Model(&models.Device{}).Where("status = 'online'").Count(&count)
	stats.OnlineDevices = int(count)
	s.db.WithContext(ctx).Model(&models.Device{}).Count(&count)
	stats.TotalDevices = int(count)
	s.db.WithContext(ctx).Model(&models.Alert{}).Where("status = 'pending'").Count(&count)
	stats.ActiveAlerts = int(count)
	s.db.WithContext(ctx).Model(&models.Alert{}).Where("severity = 'P0' AND status = 'pending'").Count(&count)
	stats.P0Alerts = int(count)
	s.db.WithContext(ctx).Model(&models.Alert{}).Where("severity = 'P1' AND status = 'pending'").Count(&count)
	stats.P1Alerts = int(count)
	s.db.WithContext(ctx).Model(&models.Alert{}).Where("severity = 'P2' AND status = 'pending'").Count(&count)
	stats.P2Alerts = int(count)
	s.db.WithContext(ctx).Model(&models.User{}).Count(&count)
	stats.TotalUsers = int(count)
	s.db.WithContext(ctx).Model(&models.Subscription{}).Where("status = 'active'").Count(&count)
	stats.ActiveSubscriptions = int(count)
	return &stats, nil
}

func (s *Store) GetSubscriptionStats(ctx context.Context) ([]model.SubscriptionStat, error) {
	type statRow struct {
		Tier  string `gorm:"column:plan_tier"`
		Count int64  `gorm:"column:count"`
	}
	var rows []statRow
	if err := s.db.WithContext(ctx).Model(&models.Subscription{}).
		Select("plan_tier, COUNT(*) as count").Group("plan_tier").Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("subscription stats: %w", err)
	}
	var total int64
	s.db.WithContext(ctx).Model(&models.Subscription{}).Count(&total)
	result := make([]model.SubscriptionStat, len(rows))
	for i, r := range rows {
		pct := float64(0)
		if total > 0 {
			pct = float64(r.Count) * 100.0 / float64(total)
		}
		result[i] = model.SubscriptionStat{Tier: r.Tier, Count: int(r.Count), Pct: pct}
	}
	return result, nil
}

func (s *Store) GetAlertTrend(ctx context.Context, days int) ([]model.AlertTrendPoint, error) {
	// Use Raw SQL for date aggregation (GORM has limited support for DATE() formatting)
	type trendRow struct {
		Date          string `gorm:"column:date"`
		BraceletCount int    `gorm:"column:bracelet_count"`
		PillboxCount  int    `gorm:"column:pillbox_count"`
	}
	var rows []trendRow
	sql := `SELECT DATE(a.created_at) AS date,
			   SUM(CASE WHEN d.device_type = 'bracelet' THEN 1 ELSE 0 END) AS bracelet_count,
			   SUM(CASE WHEN d.device_type = 'pillbox' THEN 1 ELSE 0 END) AS pillbox_count
			FROM alerts a LEFT JOIN devices d ON a.elderly_id = d.id
			WHERE a.created_at >= ?
			GROUP BY DATE(a.created_at) ORDER BY date`
	if err := s.db.Raw(sql, time.Now().AddDate(0, 0, -days)).Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("alert trend: %w", err)
	}
	result := make([]model.AlertTrendPoint, len(rows))
	for i, r := range rows {
		result[i] = model.AlertTrendPoint{Date: r.Date, BraceletCount: r.BraceletCount, PillboxCount: r.PillboxCount}
	}
	return result, nil
}

func (s *Store) GetAlertDistribution(ctx context.Context) ([]model.AlertDistributionItem, error) {
	type distRow struct {
		AlertType string `gorm:"column:alert_type"`
		Count     int64  `gorm:"column:count"`
	}
	var rows []distRow
	if err := s.db.WithContext(ctx).Model(&models.Alert{}).
		Select("alert_type, COUNT(*) as count").Group("alert_type").Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("alert distribution: %w", err)
	}
	colors := map[string]string{
		"sos": "#ff4d4f", "fall": "#fa541c", "med_missed": "#faad14",
		"device_offline": "#1890ff", "geofence_breach": "#722ed1",
	}
	result := make([]model.AlertDistributionItem, len(rows))
	for i, r := range rows {
		result[i] = model.AlertDistributionItem{Name: r.AlertType, Value: int(r.Count), Color: colors[r.AlertType]}
	}
	return result, nil
}

func (s *Store) GetUserGrowth(ctx context.Context, months int) ([]model.UserGrowthPoint, error) {
	// Use Raw SQL for month formatting
	type growthRow struct {
		Month    string `gorm:"column:month"`
		NewUsers int64  `gorm:"column:new_users"`
	}
	var rows []growthRow
	sql := `SELECT strftime('%Y-%m', created_at) AS month, COUNT(*) AS new_users
			FROM users GROUP BY strftime('%Y-%m', created_at)
			ORDER BY month DESC LIMIT ?`
	if err := s.db.Raw(sql, months).Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("user growth: %w", err)
	}
	result := make([]model.UserGrowthPoint, len(rows))
	for i, r := range rows {
		result[i] = model.UserGrowthPoint{Month: r.Month, NewUsers: int(r.NewUsers)}
	}
	return result, nil
}

func (s *Store) ListSubscriptions(ctx context.Context, page, pageSize int, status, planTier string) ([]model.SubscriptionItem, error) {
	var subs []models.Subscription
	query := s.db.WithContext(ctx).Model(&models.Subscription{})
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if planTier != "" {
		query = query.Where("plan_tier = ?", planTier)
	}
	query = query.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize)
	if err := query.Find(&subs).Error; err != nil {
		return nil, fmt.Errorf("list subscriptions: %w", err)
	}
	result := make([]model.SubscriptionItem, len(subs))
	for i, s := range subs {
		startDate := ""
		endDate := ""
		if s.StartsAt != nil {
			startDate = s.StartsAt.Format("2006-01-02")
		}
		if s.ExpiresAt != nil {
			endDate = s.ExpiresAt.Format("2006-01-02")
		}
		result[i] = model.SubscriptionItem{
			ID: s.ID, UserID: s.UserID, PlanTier: s.PlanTier, Status: s.Status,
			BillingCycle: s.BillingCycle, StartDate: startDate, EndDate: endDate,
		}
	}
	return result, nil
}

func (s *Store) GetSubscription(ctx context.Context, id string) (*model.SubscriptionItem, error) {
	var sub models.Subscription
	if err := s.db.WithContext(ctx).Where("id = ?", id).First(&sub).Error; err != nil {
		return nil, err
	}
	startDate := ""
	endDate := ""
	if sub.StartsAt != nil {
		startDate = sub.StartsAt.Format("2006-01-02")
	}
	if sub.ExpiresAt != nil {
		endDate = sub.ExpiresAt.Format("2006-01-02")
	}
	return &model.SubscriptionItem{ID: sub.ID, UserID: sub.UserID, PlanTier: sub.PlanTier,
		Status: sub.Status, BillingCycle: sub.BillingCycle, StartDate: startDate, EndDate: endDate}, nil
}

func (s *Store) CreateSubscription(ctx context.Context, sub *model.SubscriptionItem) error {
	id := uuid.New().String()
	var startsAt, expiresAt *time.Time
	if sub.StartDate != "" {
		if t, err := time.Parse("2006-01-02", sub.StartDate); err == nil {
			startsAt = &t
		}
	}
	if sub.EndDate != "" {
		if t, err := time.Parse("2006-01-02", sub.EndDate); err == nil {
			expiresAt = &t
		}
	}
	subs := &models.Subscription{BaseModel: models.BaseModel{ID: id}, UserID: sub.UserID, PlanTier: sub.PlanTier,
		Status: sub.Status, BillingCycle: sub.BillingCycle, StartsAt: startsAt, ExpiresAt: expiresAt}
	return s.db.WithContext(ctx).Create(subs).Error
}

func (s *Store) UpdateSubscription(ctx context.Context, id string, updates map[string]any) error {
	return s.db.WithContext(ctx).Model(&models.Subscription{}).Where("id = ?", id).Updates(updates).Error
}

func (s *Store) RenewSubscription(ctx context.Context, id, endDate string) error {
	var expiresAt *time.Time
	if endDate != "" {
		if t, err := time.Parse("2006-01-02", endDate); err == nil {
			expiresAt = &t
		}
	}
	return s.db.WithContext(ctx).Model(&models.Subscription{}).Where("id = ?", id).
		Updates(map[string]interface{}{"expires_at": expiresAt, "status": "active"}).Error
}

// ====== Admission Store Methods ======

func (s *Store) CreateAdmission(ctx context.Context, a *model.HospitalAdmission) error {
	admission := &models.HospitalAdmission{
		BaseModel:           models.BaseModel{ID: a.ID},
		PatientID:           a.PatientID,
		AdmissionNo:         a.AdmissionNo,
		BedNo:               a.BedNo,
		Department:          a.Department,
		Diagnosis:           a.Diagnosis,
		EmergencyContact:    a.EmergencyContact,
		Allergies:           a.Allergies,
		AdmittedAt:          time.Now(),
		ExpectedDischargeAt: a.ExpectedDischargeAt,
		DischargeType:       a.DischargeType,
		TransferredTo:       a.TransferredTo,
		Notes:               a.Notes,
	}
	return s.db.WithContext(ctx).Create(admission).Error
}

func (s *Store) GetAdmission(ctx context.Context, id string) (*model.HospitalAdmission, error) {
	var a models.HospitalAdmission
	if err := s.db.WithContext(ctx).Where("id = ?", id).First(&a).Error; err != nil {
		return nil, err
	}
	return &model.HospitalAdmission{
		ID:                  a.ID,
		PatientID:           a.PatientID,
		AdmissionNo:         a.AdmissionNo,
		BedNo:               a.BedNo,
		Department:          a.Department,
		Diagnosis:           a.Diagnosis,
		EmergencyContact:    a.EmergencyContact,
		Allergies:           a.Allergies,
		AdmittedAt:          a.AdmittedAt,
		ExpectedDischargeAt: a.ExpectedDischargeAt,
		DischargedAt:        a.DischargedAt,
		DischargeType:       a.DischargeType,
		TransferredTo:       a.TransferredTo,
		Notes:               a.Notes,
	}, nil
}

func (s *Store) ListAdmissions(ctx context.Context, page, pageSize int, department, status string) ([]model.HospitalAdmission, error) {
	var admissions []models.HospitalAdmission
	query := s.db.WithContext(ctx).Model(&models.HospitalAdmission{})
	if department != "" {
		query = query.Where("department = ?", department)
	}
	if status != "" {
		query = query.Where("status != ?", status)
	}
	query = query.Order("admitted_at DESC").Offset((page - 1) * pageSize).Limit(pageSize)
	if err := query.Find(&admissions).Error; err != nil {
		return nil, fmt.Errorf("list admissions: %w", err)
	}
	result := make([]model.HospitalAdmission, len(admissions))
	for i, a := range admissions {
		result[i] = model.HospitalAdmission{
			ID: a.ID, PatientID: a.PatientID, AdmissionNo: a.AdmissionNo, BedNo: a.BedNo,
			Department: a.Department, Diagnosis: a.Diagnosis, EmergencyContact: a.EmergencyContact,
			Allergies: a.Allergies, AdmittedAt: a.AdmittedAt, ExpectedDischargeAt: a.ExpectedDischargeAt,
			DischargedAt: a.DischargedAt, DischargeType: a.DischargeType, TransferredTo: a.TransferredTo, Notes: a.Notes,
		}
	}
	return result, nil
}

func (s *Store) CompleteAdmission(ctx context.Context, id, dischargeType, notes, transferredTo string) error {
	now := time.Now()
	if err := s.db.WithContext(ctx).Model(&models.HospitalAdmission{}).Where("id = ?", id).
		Updates(map[string]interface{}{"discharged_at": now, "discharge_type": dischargeType, "notes": notes, "transferred_to": transferredTo}).Error; err != nil {
		return err
	}
	var patientID string
	s.db.WithContext(ctx).Model(&models.HospitalAdmission{}).Where("id = ?", id).Select("patient_id").First(&models.HospitalAdmission{}).Scan(&patientID)
	if patientID != "" {
		s.db.WithContext(ctx).Model(&models.MedicalPatient{}).Where("id = ?", patientID).Update("status", "discharged")
		s.db.WithContext(ctx).Model(&models.MedicalBinding{}).Where("patient_id = ? AND unbound_at IS NULL", patientID).Update("unbound_at", now)
	}
	return nil
}

func (s *Store) CreateWardRound(ctx context.Context, w *model.WardRoundEntry) error {
	entry := &models.WardRoundEntry{
		BaseModel:     models.BaseModel{ID: w.ID},
		PatientID:     w.PatientID,
		NurseID:       w.NurseID,
		BloodPressure: w.BloodPressure,
		HeartRate:     &w.HeartRate,
		SpO2:          &w.SpO2,
		Temperature:   &w.Temperature,
		Weight:        &w.Weight,
		Notes:         w.Notes,
		Observations:  w.Observations,
		CompletedAt:   time.Now(),
	}
	return s.db.WithContext(ctx).Create(entry).Error
}

func (s *Store) ListWardRounds(ctx context.Context, patientID string) ([]model.WardRoundEntry, error) {
	var entries []models.WardRoundEntry
	if err := s.db.WithContext(ctx).Where("patient_id = ?", patientID).Order("completed_at DESC").Find(&entries).Error; err != nil {
		return nil, fmt.Errorf("list ward rounds: %w", err)
	}
	result := make([]model.WardRoundEntry, len(entries))
	for i, e := range entries {
		heartRate := 0
		spo2 := 0
		temperature := 0.0
		weight := 0.0
		if e.HeartRate != nil {
			heartRate = *e.HeartRate
		}
		if e.SpO2 != nil {
			spo2 = *e.SpO2
		}
		if e.Temperature != nil {
			temperature = *e.Temperature
		}
		if e.Weight != nil {
			weight = *e.Weight
		}
		result[i] = model.WardRoundEntry{
			ID: e.ID, PatientID: e.PatientID, NurseID: e.NurseID, BloodPressure: e.BloodPressure,
			HeartRate: heartRate, SpO2: spo2, Temperature: temperature, Weight: weight,
			Notes: e.Notes, Observations: e.Observations, CompletedAt: e.CompletedAt,
		}
	}
	return result, nil
}

func (s *Store) EvaluateRegulatoryRules(ctx context.Context, event string, data map[string]string) ([]*model.RegulatoryRuleResult, error) {
	var results []*model.RegulatoryRuleResult
	now := time.Now().UTC()
	patientID := data["patient_id"]
	switch event {
	case "patient_admitted":
		var count int64
		s.db.WithContext(ctx).Model(&models.MedicalBinding{}).Where("patient_id = ? AND unbound_at IS NULL", patientID).Count(&count)
		if count == 0 {
			results = append(results, &model.RegulatoryRuleResult{RuleCode: "R01", Severity: "P1", PatientID: patientID, Message: "Patient admitted without active wristband binding", TriggeredAt: now})
		}
	case "verification_scan":
		if data["scan_type"] == "medication" {
			var count int64
			s.db.WithContext(ctx).Table("medical_bindings mb").Joins("JOIN medical_wristband_patients p ON p.id=mb.patient_id").
				Where("p.id=? AND p.status='admitted' AND mb.unbound_at IS NULL", patientID).Count(&count)
			if count == 0 {
				results = append(results, &model.RegulatoryRuleResult{RuleCode: "R05", Severity: "P2", PatientID: patientID, Message: "Medication verification without active wristband binding", TriggeredAt: now})
			}
		}
	}
	return results, nil
}

// ====== Patient Store Methods ======

func (s *Store) CreatePatient(ctx context.Context, p *model.MedicalPatient) error {
	tagsJSON, _ := json.Marshal(p.TagIDs)
	patient := &models.MedicalPatient{
		BaseModel:         models.BaseModel{ID: p.ID},
		AdmissionNo:       p.AdmissionNo,
		Name:              p.Name,
		Gender:            p.Gender,
		Age:               &p.Age,
		Department:        p.Department,
		BedNumber:         p.BedNumber,
		BloodType:         p.BloodType,
		Allergies:         p.Allergies,
		SpecialConditions: p.SpecialConditions,
		TagIDs:            string(tagsJSON),
		Status:            p.Status,
	}
	return s.db.WithContext(ctx).Create(patient).Error
}

func (s *Store) GetPatient(ctx context.Context, id string) (*model.MedicalPatient, error) {
	var p models.MedicalPatient
	if err := s.db.WithContext(ctx).Where("id = ?", id).First(&p).Error; err != nil {
		return nil, err
	}
	var tags []string
	json.Unmarshal([]byte(p.TagIDs), &tags)
	age := 0
	if p.Age != nil {
		age = *p.Age
	}
	return &model.MedicalPatient{
		ID: p.ID, AdmissionNo: p.AdmissionNo, Name: p.Name, Gender: p.Gender, Age: age,
		Department: p.Department, BedNumber: p.BedNumber, BloodType: p.BloodType,
		Allergies: p.Allergies, SpecialConditions: p.SpecialConditions, TagIDs: tags, Status: p.Status,
		CreatedAt: p.CreatedAt, UpdatedAt: p.UpdatedAt,
	}, nil
}

func (s *Store) ListPatients(ctx context.Context, page, pageSize int, status string) ([]model.MedicalPatient, error) {
	var patients []models.MedicalPatient
	query := s.db.WithContext(ctx).Model(&models.MedicalPatient{})
	if status != "" {
		query = query.Where("status = ?", status)
	}
	query = query.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize)
	if err := query.Find(&patients).Error; err != nil {
		return nil, fmt.Errorf("list patients: %w", err)
	}
	result := make([]model.MedicalPatient, len(patients))
	for i, p := range patients {
		var tags []string
		json.Unmarshal([]byte(p.TagIDs), &tags)
		age := 0
		if p.Age != nil {
			age = *p.Age
		}
		result[i] = model.MedicalPatient{
			ID: p.ID, AdmissionNo: p.AdmissionNo, Name: p.Name, Gender: p.Gender, Age: age,
			Department: p.Department, BedNumber: p.BedNumber, BloodType: p.BloodType,
			Allergies: p.Allergies, SpecialConditions: p.SpecialConditions, TagIDs: tags, Status: p.Status,
			CreatedAt: p.CreatedAt, UpdatedAt: p.UpdatedAt,
		}
	}
	return result, nil
}

func (s *Store) UpdatePatient(ctx context.Context, p *model.MedicalPatient) error {
	tagsJSON, _ := json.Marshal(p.TagIDs)
	age := p.Age
	return s.db.WithContext(ctx).Model(&models.MedicalPatient{}).Where("id = ?", p.ID).
		Updates(map[string]interface{}{
			"admission_no": p.AdmissionNo, "name": p.Name, "gender": p.Gender, "age": age,
			"department": p.Department, "bed_number": p.BedNumber, "blood_type": p.BloodType,
			"allergies": p.Allergies, "special_conditions": p.SpecialConditions,
			"tag_ids": string(tagsJSON), "status": p.Status,
		}).Error
}

func (s *Store) DeletePatient(ctx context.Context, id string) error {
	return s.db.WithContext(ctx).Model(&models.MedicalPatient{}).Where("id = ?", id).Update("status", "discharged").Error
}

func (s *Store) GetPatientByAdmissionNo(ctx context.Context, admissionNo string) (*model.MedicalPatient, error) {
	var p models.MedicalPatient
	if err := s.db.WithContext(ctx).Where("admission_no = ?", admissionNo).First(&p).Error; err != nil {
		return nil, err
	}
	var tags []string
	json.Unmarshal([]byte(p.TagIDs), &tags)
	age := 0
	if p.Age != nil {
		age = *p.Age
	}
	return &model.MedicalPatient{
		ID: p.ID, AdmissionNo: p.AdmissionNo, Name: p.Name, Gender: p.Gender, Age: age,
		Department: p.Department, BedNumber: p.BedNumber, BloodType: p.BloodType,
		Allergies: p.Allergies, SpecialConditions: p.SpecialConditions, TagIDs: tags, Status: p.Status,
		CreatedAt: p.CreatedAt, UpdatedAt: p.UpdatedAt,
	}, nil
}

func (s *Store) BatchImportPatients(ctx context.Context, patients []model.MedicalPatient) error {
	for _, p := range patients {
		if err := s.CreatePatient(ctx, &p); err != nil {
			return fmt.Errorf("import patient %s: %w", p.Name, err)
		}
	}
	return nil
}

func (s *Store) GetPatientHistory(ctx context.Context, patientID string) (*model.MedicalPatientHistory, error) {
	entries, err := s.ListDailyEntries(ctx, patientID, "")
	if err != nil {
		return nil, err
	}
	return &model.MedicalPatientHistory{DailyEntries: entries}, nil
}

func (s *Store) ListDailyEntries(ctx context.Context, patientID string, date string) ([]model.MedicalDailyEntry, error) {
	var entries []models.MedicalDailyEntry
	query := s.db.WithContext(ctx).Where("patient_id = ?", patientID).Order("entry_date DESC, created_at DESC")
	if date != "" {
		query = query.Where("entry_date = ?", date)
	}
	if err := query.Find(&entries).Error; err != nil {
		return nil, fmt.Errorf("list daily entries: %w", err)
	}
	result := make([]model.MedicalDailyEntry, len(entries))
	for i, e := range entries {
		result[i] = model.MedicalDailyEntry{
			ID: e.ID, PatientID: e.PatientID, EntryDate: e.EntryDate,
			EntryType: e.EntryType, Content: e.Content, NurseID: e.NurseID,
			CreatedAt: e.CreatedAt, UpdatedAt: e.UpdatedAt,
		}
	}
	return result, nil
}

// ====== Wristband Store Methods ======

func (s *Store) BindWristband(ctx context.Context, patientID, deviceID string) error {
	binding := &models.MedicalBinding{PatientID: patientID, DeviceID: deviceID, BoundAt: time.Now()}
	if err := s.db.WithContext(ctx).Create(binding).Error; err != nil {
		return err
	}
	return s.db.WithContext(ctx).Model(&models.MedicalWristbandDevice{}).Where("device_id = ?", deviceID).
		Updates(map[string]interface{}{"bound_patient_id": patientID, "status": "bound"}).Error
}

func (s *Store) UnbindWristband(ctx context.Context, bindingID string) error {
	now := time.Now()
	if err := s.db.WithContext(ctx).Model(&models.MedicalBinding{}).Where("id = ? AND unbound_at IS NULL", bindingID).
		Update("unbound_at", now).Error; err != nil {
		return err
	}
	return s.db.WithContext(ctx).Model(&models.MedicalWristbandDevice{}).Where("id IN (SELECT device_id FROM medical_bindings WHERE id=?)", bindingID).
		Updates(map[string]interface{}{"bound_patient_id": nil, "status": "idle"}).Error
}

func (s *Store) ClearWristband(ctx context.Context, deviceID string) error {
	return s.db.WithContext(ctx).Model(&models.MedicalWristbandDevice{}).Where("device_id = ?", deviceID).
		Updates(map[string]interface{}{"bound_patient_id": nil, "status": "cleared"}).Error
}

func (s *Store) ListWristbands(ctx context.Context, page, pageSize int, status string) ([]model.MedicalWristbandDevice, error) {
	var devices []models.MedicalWristbandDevice
	query := s.db.WithContext(ctx).Model(&models.MedicalWristbandDevice{})
	if status != "" {
		query = query.Where("status = ?", status)
	}
	query = query.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize)
	if err := query.Find(&devices).Error; err != nil {
		return nil, fmt.Errorf("list wristbands: %w", err)
	}
	result := make([]model.MedicalWristbandDevice, len(devices))
	for i, d := range devices {
		boundID := ""
		if d.BoundPatientID != nil {
			boundID = *d.BoundPatientID
		}
		result[i] = model.MedicalWristbandDevice{
			ID: d.ID, DeviceID: d.DeviceID, FirmwareVersion: d.FirmwareVersion,
			Status: d.Status, BoundPatientID: boundID, CreatedAt: d.CreatedAt, UpdatedAt: d.UpdatedAt,
		}
	}
	return result, nil
}

func (s *Store) GetWristbandFirmware(ctx context.Context, deviceID string) (string, error) {
	var d models.MedicalWristbandDevice
	if err := s.db.WithContext(ctx).Where("device_id = ?", deviceID).First(&d).Error; err != nil {
		return "", err
	}
	return d.FirmwareVersion, nil
}

func (s *Store) WriteToWristband(ctx context.Context, deviceID, data string) error {
	return s.db.WithContext(ctx).Model(&models.MedicalWristbandDevice{}).Where("device_id = ?", deviceID).
		Update("firmware_version", data).Error
}

// ====== Audit Store Methods ======

func (s *Store) CreateAuditLog(ctx context.Context, log *model.AuditLog) error {
	detailsJSON, _ := json.Marshal(log.Details)
	audit := &models.AuditLog{
		BaseModel: models.BaseModel{ID: log.ID},
		UserID:    log.UserID,
		Action:    log.Action,
		Resource:  log.Resource,
		ResourceID: log.ResourceID,
		Details:   string(detailsJSON),
		IPAddress: log.IP,
		UserAgent: log.UserAgent,
	}
	if audit.ID == "" {
		audit.ID = uuid.New().String()
	}
	return s.db.WithContext(ctx).Create(audit).Error
}

func (s *Store) ListAuditLogs(ctx context.Context, limit int) ([]model.AuditLog, error) {
	if limit <= 0 || limit > 500 {
		limit = 50
	}
	var logs []models.AuditLog
	if err := s.db.WithContext(ctx).Order("created_at DESC").Limit(limit).Find(&logs).Error; err != nil {
		return nil, err
	}
	result := make([]model.AuditLog, len(logs))
	for i, l := range logs {
		var details map[string]interface{}
		json.Unmarshal([]byte(l.Details), &details)
		result[i] = model.AuditLog{
			ID: l.ID, UserID: l.UserID, Action: l.Action, Resource: l.Resource,
			ResourceID: l.ResourceID, Details: details, IP: l.IPAddress, UserAgent: l.UserAgent,
			CreatedAt: l.CreatedAt,
		}
	}
	return result, nil
}

func (s *Store) ListAuditLogsByUser(ctx context.Context, userID string, limit int) ([]model.AuditLog, error) {
	if limit <= 0 || limit > 500 {
		limit = 50
	}
	var logs []models.AuditLog
	if err := s.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at DESC").Limit(limit).Find(&logs).Error; err != nil {
		return nil, err
	}
	result := make([]model.AuditLog, len(logs))
	for i, l := range logs {
		var details map[string]interface{}
		json.Unmarshal([]byte(l.Details), &details)
		result[i] = model.AuditLog{
			ID: l.ID, UserID: l.UserID, Action: l.Action, Resource: l.Resource,
			ResourceID: l.ResourceID, Details: details, IP: l.IPAddress, UserAgent: l.UserAgent,
			CreatedAt: l.CreatedAt,
		}
	}
	return result, nil
}

func (s *Store) ListAuditLogsByAction(ctx context.Context, action string, limit int) ([]model.AuditLog, error) {
	if limit <= 0 || limit > 500 {
		limit = 50
	}
	var logs []models.AuditLog
	if err := s.db.WithContext(ctx).Where("action = ?", action).Order("created_at DESC").Limit(limit).Find(&logs).Error; err != nil {
		return nil, err
	}
	result := make([]model.AuditLog, len(logs))
	for i, l := range logs {
		var details map[string]interface{}
		json.Unmarshal([]byte(l.Details), &details)
		result[i] = model.AuditLog{
			ID: l.ID, UserID: l.UserID, Action: l.Action, Resource: l.Resource,
			ResourceID: l.ResourceID, Details: details, IP: l.IPAddress, UserAgent: l.UserAgent,
			CreatedAt: l.CreatedAt,
		}
	}
	return result, nil
}

// ====== Institution Methods ======

func (s *Store) CreateInstitution(ctx context.Context, i *model.InstitutionSummary) error {
	i.ID = uuid.New().String()
	i.CreatedAt = time.Now()
	i.UpdatedAt = time.Now()
	return s.db.WithContext(ctx).Create(&models.Institution{BaseModel: models.BaseModel{ID: i.ID, CreatedAt: i.CreatedAt, UpdatedAt: i.UpdatedAt}, Name: i.Name, Type: i.Type, Code: i.Code, ContactName: i.ContactName, ContactPhone: i.ContactPhone, AccessLevel: i.AccessLevel, Status: i.Status}).Error
}

func (s *Store) GetInstitution(ctx context.Context, id string) (*model.InstitutionSummary, error) {
	var i models.Institution
	if err := s.db.WithContext(ctx).Where("id = ?", id).First(&i).Error; err != nil {
		return nil, err
	}
	return &model.InstitutionSummary{ID: i.ID, Name: i.Name, Type: i.Type, Code: i.Code, ContactName: i.ContactName, ContactPhone: i.ContactPhone, AccessLevel: i.AccessLevel, Status: i.Status, CreatedAt: i.CreatedAt, UpdatedAt: i.UpdatedAt, APIKeyCount: int(i.APIKeyCount)}, nil
}

func (s *Store) ListInstitutions(ctx context.Context, page, pageSize int, name, typ, status string) ([]model.InstitutionSummary, error) {
	var institutions []models.Institution
	query := s.db.WithContext(ctx).Model(&models.Institution{})
	if name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}
	if typ != "" {
		query = query.Where("type = ?", typ)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	query = query.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize)
	if err := query.Find(&institutions).Error; err != nil {
		return nil, fmt.Errorf("list institutions: %w", err)
	}
	result := make([]model.InstitutionSummary, len(institutions))
	for i, inst := range institutions {
		result[i] = model.InstitutionSummary{ID: inst.ID, Name: inst.Name, Type: inst.Type, Code: inst.Code, ContactName: inst.ContactName, ContactPhone: inst.ContactPhone, AccessLevel: inst.AccessLevel, Status: inst.Status, CreatedAt: inst.CreatedAt, UpdatedAt: inst.UpdatedAt, APIKeyCount: int(inst.APIKeyCount)}
	}
	return result, nil
}

func (s *Store) UpdateInstitution(ctx context.Context, id string, updates map[string]any) error {
	return s.db.WithContext(ctx).Model(&models.Institution{}).Where("id = ?", id).Updates(updates).Error
}

func (s *Store) DeleteInstitution(ctx context.Context, id string) error {
	return s.db.WithContext(ctx).Delete(&models.Institution{}, "id = ?", id).Error
}

func (s *Store) CreateInstitutionAPIKey(ctx context.Context, institutionID, name string) (string, error) {
	id := uuid.New().String()
	key := &models.InstitutionAPIKey{BaseModel: models.BaseModel{ID: id}, InstitutionID: institutionID, Name: name, KeyHash: uuid.New().String(), Active: true}
	if err := s.db.WithContext(ctx).Create(key).Error; err != nil {
		return "", fmt.Errorf("create api key: %w", err)
	}
	return id, nil
}

func (s *Store) RevokeInstitutionAPIKey(ctx context.Context, institutionID, keyID string) error {
	return s.db.WithContext(ctx).Model(&models.InstitutionAPIKey{}).Where("id = ? AND institution_id = ?", keyID, institutionID).Update("active", false).Error
}

// ====== Person Store Methods ======

func (s *Store) CreatePerson(ctx context.Context, p *model.Person) error {
	p.ID = uuid.New().String()
	p.CreatedAt = time.Now()
	p.UpdatedAt = time.Now()
	person := &models.Person{BaseModel: models.BaseModel{ID: p.ID}, IDCard: p.IDCard, Name: p.Name, Gender: p.Gender, BirthDate: p.BirthDate, Phone: p.Phone, EmergencyContact: p.EmergencyContact, Address: p.Address, AvatarURL: p.AvatarURL, Status: p.Status}
	return s.db.WithContext(ctx).Create(person).Error
}

func (s *Store) GetPerson(ctx context.Context, id string) (*model.Person, error) {
	var p models.Person
	if err := s.db.WithContext(ctx).Where("id = ?", id).First(&p).Error; err != nil {
		return nil, err
	}
	return &model.Person{ID: p.ID, IDCard: p.IDCard, Name: p.Name, Gender: p.Gender, BirthDate: p.BirthDate, Phone: p.Phone, EmergencyContact: p.EmergencyContact, Address: p.Address, AvatarURL: p.AvatarURL, Status: p.Status, CreatedAt: p.CreatedAt, UpdatedAt: p.UpdatedAt}, nil
}

func (s *Store) GetPersonByIDCard(ctx context.Context, idCard string) (*model.Person, error) {
	var p models.Person
	if err := s.db.WithContext(ctx).Where("id_card = ?", idCard).First(&p).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &model.Person{ID: p.ID, IDCard: p.IDCard, Name: p.Name, Gender: p.Gender, BirthDate: p.BirthDate, Phone: p.Phone, EmergencyContact: p.EmergencyContact, Address: p.Address, AvatarURL: p.AvatarURL, Status: p.Status, CreatedAt: p.CreatedAt, UpdatedAt: p.UpdatedAt}, nil
}

func (s *Store) ListPersons(ctx context.Context, page, pageSize int, businessChain, status string) ([]model.Person, error) {
	var persons []models.Person
	query := s.db.WithContext(ctx).Model(&models.Person{})
	if businessChain != "" {
		query = query.Joins("JOIN person_profiles ON person_profiles.person_id = persons.id").Where("person_profiles.business_chain = ?", businessChain)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	query = query.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize)
	if err := query.Find(&persons).Error; err != nil {
		return nil, fmt.Errorf("list persons: %w", err)
	}
	result := make([]model.Person, len(persons))
	for i, p := range persons {
		result[i] = model.Person{ID: p.ID, IDCard: p.IDCard, Name: p.Name, Gender: p.Gender, BirthDate: p.BirthDate, Phone: p.Phone, EmergencyContact: p.EmergencyContact, Address: p.Address, AvatarURL: p.AvatarURL, Status: p.Status, CreatedAt: p.CreatedAt, UpdatedAt: p.UpdatedAt}
	}
	return result, nil
}

func (s *Store) UpdatePerson(ctx context.Context, id string, updates map[string]any) error {
	if len(updates) == 0 {
		return nil
	}
	return s.db.WithContext(ctx).Model(&models.Person{}).Where("id = ?", id).Updates(updates).Error
}

func (s *Store) DeletePerson(ctx context.Context, id string) error {
	return s.db.WithContext(ctx).Delete(&models.Person{}, "id = ?", id).Error
}

// ====== Person Profile Methods ======

func (s *Store) CreateProfile(ctx context.Context, pp *model.PersonProfile) error {
	pp.CreatedAt = time.Now()
	pp.UpdatedAt = time.Now()
	profile := &models.PersonProfile{PersonID: pp.PersonID, BusinessChain: pp.BusinessChain, SubscriptionTier: pp.SubscriptionTier, SubscriptionStatus: pp.SubscriptionStatus, SubscriptionStart: pp.SubscriptionStart, SubscriptionEnd: pp.SubscriptionEnd, HealthRiskLevel: pp.HealthRiskLevel, AdmissionNo: pp.AdmissionNo, Department: pp.Department, BedNumber: pp.BedNumber, BloodType: pp.BloodType, AttendingDoctor: pp.AttendingDoctor, Diagnosis: pp.Diagnosis, AdmissionDate: pp.AdmissionDate, ExpectedDischarge: pp.ExpectedDischarge, DischargeDate: pp.DischargeDate, DischargeType: pp.DischargeType, HospitalID: pp.HospitalID, HospitalIDCommunity: pp.HospitalIDCommunity, MinzhengCertified: pp.MinzhengCertified, SubsidyType: pp.SubsidyType, CertificationDate: pp.CertificationDate, CertificationDoc: pp.CertificationDoc, NextReviewDate: pp.NextReviewDate, LinkedPersonID: pp.LinkedPersonID}
	return s.db.WithContext(ctx).Create(profile).Error
}

func (s *Store) GetProfile(ctx context.Context, personID string, chain model.BusinessChain) (*model.PersonProfile, error) {
	var pp models.PersonProfile
	if err := s.db.WithContext(ctx).Where("person_id = ? AND business_chain = ?", personID, chain).First(&pp).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("get profile: %w", err)
	}
	return &model.PersonProfile{PersonID: pp.PersonID, BusinessChain: pp.BusinessChain, SubscriptionTier: pp.SubscriptionTier, SubscriptionStatus: pp.SubscriptionStatus, SubscriptionStart: pp.SubscriptionStart, SubscriptionEnd: pp.SubscriptionEnd, HealthRiskLevel: pp.HealthRiskLevel, AdmissionNo: pp.AdmissionNo, Department: pp.Department, BedNumber: pp.BedNumber, BloodType: pp.BloodType, AttendingDoctor: pp.AttendingDoctor, Diagnosis: pp.Diagnosis, AdmissionDate: pp.AdmissionDate, ExpectedDischarge: pp.ExpectedDischarge, DischargeDate: pp.DischargeDate, DischargeType: pp.DischargeType, HospitalID: pp.HospitalID, HospitalIDCommunity: pp.HospitalIDCommunity, MinzhengCertified: pp.MinzhengCertified, SubsidyType: pp.SubsidyType, CertificationDate: pp.CertificationDate, CertificationDoc: pp.CertificationDoc, NextReviewDate: pp.NextReviewDate, LinkedPersonID: pp.LinkedPersonID}, nil
}

func (s *Store) ListProfiles(ctx context.Context, chain model.BusinessChain) ([]model.PersonProfile, error) {
	var profiles []models.PersonProfile
	if err := s.db.WithContext(ctx).Where("business_chain = ?", chain).Order("created_at DESC").Find(&profiles).Error; err != nil {
		return nil, fmt.Errorf("list profiles: %w", err)
	}
	result := make([]model.PersonProfile, len(profiles))
	for i, p := range profiles {
		result[i] = model.PersonProfile{PersonID: p.PersonID, BusinessChain: p.BusinessChain, SubscriptionTier: p.SubscriptionTier, SubscriptionStatus: p.SubscriptionStatus, SubscriptionStart: p.SubscriptionStart, SubscriptionEnd: p.SubscriptionEnd, HealthRiskLevel: p.HealthRiskLevel, AdmissionNo: p.AdmissionNo, Department: p.Department, BedNumber: p.BedNumber, BloodType: p.BloodType, AttendingDoctor: p.AttendingDoctor, Diagnosis: p.Diagnosis, AdmissionDate: p.AdmissionDate, ExpectedDischarge: p.ExpectedDischarge, DischargeDate: p.DischargeDate, DischargeType: p.DischargeType, HospitalID: p.HospitalID, HospitalIDCommunity: p.HospitalIDCommunity, MinzhengCertified: p.MinzhengCertified, SubsidyType: p.SubsidyType, CertificationDate: p.CertificationDate, CertificationDoc: p.CertificationDoc, NextReviewDate: p.NextReviewDate, LinkedPersonID: p.LinkedPersonID}
	}
	return result, nil
}

func (s *Store) UpdateProfile(ctx context.Context, pp *model.PersonProfile) error {
	pp.UpdatedAt = time.Now()
	return s.db.WithContext(ctx).Model(&models.PersonProfile{}).Where("person_id = ? AND business_chain = ?", pp.PersonID, pp.BusinessChain).Updates(map[string]interface{}{"subscription_tier": pp.SubscriptionTier, "subscription_status": pp.SubscriptionStatus, "subscription_start": pp.SubscriptionStart, "subscription_end": pp.SubscriptionEnd, "health_risk_level": pp.HealthRiskLevel, "admission_no": pp.AdmissionNo, "department": pp.Department, "bed_number": pp.BedNumber, "blood_type": pp.BloodType, "attending_doctor": pp.AttendingDoctor, "diagnosis": pp.Diagnosis, "admission_date": pp.AdmissionDate, "expected_discharge": pp.ExpectedDischarge, "discharge_date": pp.DischargeDate, "discharge_type": pp.DischargeType, "hospital_id": pp.HospitalID, "hospital_id_community": pp.HospitalIDCommunity, "minzheng_certified": pp.MinzhengCertified, "subsidy_type": pp.SubsidyType, "certification_date": pp.CertificationDate, "certification_doc": pp.CertificationDoc, "next_review_date": pp.NextReviewDate, "linked_person_id": pp.LinkedPersonID, "updated_at": pp.UpdatedAt}).Error
}

// ====== Person Welfare Tag Methods ======

func (s *Store) AssignPersonWelfareTag(ctx context.Context, wt *model.PersonWelfareTag) error {
	existing := &models.PersonWelfareTag{}
	result := s.db.WithContext(ctx).Where("person_id = ? AND tag_code = ?", wt.PersonID, wt.TagCode).First(existing)
	if result.Error == nil {
		existing.ValidFrom = wt.ValidFrom
		existing.ValidTo = wt.ValidTo
		return s.db.WithContext(ctx).Save(existing).Error
	}
	tag := &models.PersonWelfareTag{PersonID: wt.PersonID, TagCode: wt.TagCode, ValidFrom: wt.ValidFrom, ValidTo: wt.ValidTo}
	return s.db.WithContext(ctx).Create(tag).Error
}

func (s *Store) RevokePersonWelfareTag(ctx context.Context, personID, tagCode string) error {
	return s.db.WithContext(ctx).Where("person_id = ? AND tag_code = ?", personID, tagCode).Delete(&models.PersonWelfareTag{}).Error
}

func (s *Store) ListPersonWelfareTags(ctx context.Context, personID string) ([]model.PersonWelfareTag, error) {
	var tags []models.PersonWelfareTag
	if err := s.db.WithContext(ctx).Where("person_id = ?", personID).Find(&tags).Error; err != nil {
		return nil, fmt.Errorf("list welfare tags: %w", err)
	}
	result := make([]model.PersonWelfareTag, len(tags))
	for i, t := range tags {
		result[i] = model.PersonWelfareTag{PersonID: t.PersonID, TagCode: t.TagCode, ValidFrom: t.ValidFrom, ValidTo: t.ValidTo}
	}
	return result, nil
}

// ====== Lifecycle Methods ======

func (s *Store) TransitionStatus(ctx context.Context, personID string, chain model.BusinessChain, newStatus, reason string) error {
	var pp models.PersonProfile
	if err := s.db.WithContext(ctx).Where("person_id = ? AND business_chain = ?", personID, chain).First(&pp).Error; err != nil {
		return fmt.Errorf("profile not found: %w", err)
	}
	return s.db.WithContext(ctx).Model(&pp).Update("status", newStatus).Error
}

func (s *Store) GetPersonStatus(ctx context.Context, personID string, chain model.BusinessChain) (string, error) {
	var pp models.PersonProfile
	err := s.db.WithContext(ctx).Where("person_id = ? AND business_chain = ?", personID, chain).Select("status").First(&pp).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", nil
	}
	return pp.Status, err
}

func (s *Store) LinkPersons(ctx context.Context, personID1, personID2 string, chain1, chain2 model.BusinessChain) error {
	if err := s.db.WithContext(ctx).Model(&models.PersonProfile{}).Where("person_id = ? AND business_chain = ?", personID1, chain1).Update("linked_person_id", personID2).Error; err != nil {
		return fmt.Errorf("link person 1: %w", err)
	}
	return s.db.WithContext(ctx).Model(&models.PersonProfile{}).Where("person_id = ? AND business_chain = ?", personID2, chain2).Update("linked_person_id", personID1).Error
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func intBool(i int) bool {
	return i != 0
}

func (s *Store) CreateExpense(ctx context.Context, e *model.MedicalExpense) error {
	expense := &models.MedicalExpense{BaseModel: models.BaseModel{ID: e.ID}, PatientID: e.PatientID, ItemName: e.ItemName, Category: e.Category, Amount: e.Amount, Quantity: e.Quantity, UnitPrice: e.UnitPrice, Notes: e.Notes}
	if expense.ID == "" {
		expense.ID = uuid.New().String()
	}
	return s.db.WithContext(ctx).Create(expense).Error
}

func (s *Store) ListExpenses(ctx context.Context, patientID string, page, pageSize int) ([]model.MedicalExpense, error) {
	var expenses []models.MedicalExpense
	query := s.db.WithContext(ctx).Where("patient_id = ?", patientID).Order("created_at DESC")
	if page > 0 && pageSize > 0 {
		query = query.Offset((page - 1) * pageSize).Limit(pageSize)
	}
	if err := query.Find(&expenses).Error; err != nil {
		return nil, fmt.Errorf("list expenses: %w", err)
	}
	result := make([]model.MedicalExpense, len(expenses))
	for i, e := range expenses {
		result[i] = model.MedicalExpense{ID: e.ID, PatientID: e.PatientID, ItemName: e.ItemName, Category: e.Category, Amount: e.Amount, Quantity: e.Quantity, UnitPrice: e.UnitPrice, Notes: e.Notes, CreatedAt: e.CreatedAt, UpdatedAt: e.UpdatedAt}
	}
	return result, nil
}

func (s *Store) CreateMedication(ctx context.Context, m *model.MedicalMedication) error {
	rec := &models.MedicalMedication{BaseModel: models.BaseModel{ID: m.ID}, PatientID: m.PatientID, Name: m.Name, Dosage: m.Dosage, Frequency: m.Frequency, Duration: m.Duration, Route: m.Route, Notes: m.Notes}
	if rec.ID == "" {
		rec.ID = uuid.New().String()
	}
	return s.db.WithContext(ctx).Create(rec).Error
}

func (s *Store) ListMedications(ctx context.Context, patientID string) ([]model.MedicalMedication, error) {
	var items []models.MedicalMedication
	if err := s.db.WithContext(ctx).Where("patient_id = ?", patientID).Order("created_at DESC").Find(&items).Error; err != nil {
		return nil, fmt.Errorf("list medications: %w", err)
	}
	result := make([]model.MedicalMedication, len(items))
	for i, m := range items {
		result[i] = model.MedicalMedication{ID: m.ID, PatientID: m.PatientID, Name: m.Name, Dosage: m.Dosage, Frequency: m.Frequency, Duration: m.Duration, Route: m.Route, Notes: m.Notes}
	}
	return result, nil
}

func (s *Store) CreateTestResult(ctx context.Context, r *model.MedicalTestResult) error {
	rec := &models.MedicalTestResult{BaseModel: models.BaseModel{ID: r.ID}, PatientID: r.PatientID, TestName: r.TestName, Result: r.Result, ReferenceRange: r.ReferenceRange, Unit: r.Unit, Notes: r.Notes}
	if rec.ID == "" {
		rec.ID = uuid.New().String()
	}
	if r.CollectedAt != nil {
		rec.CollectedAt = r.CollectedAt
	}
	if r.ReportedAt != nil {
		rec.ReportedAt = r.ReportedAt
	}
	return s.db.WithContext(ctx).Create(rec).Error
}

func (s *Store) ListTestResults(ctx context.Context, patientID string) ([]model.MedicalTestResult, error) {
	var items []models.MedicalTestResult
	if err := s.db.WithContext(ctx).Where("patient_id = ?", patientID).Order("collected_at DESC").Find(&items).Error; err != nil {
		return nil, fmt.Errorf("list test results: %w", err)
	}
	result := make([]model.MedicalTestResult, len(items))
	for i, t := range items {
		result[i] = model.MedicalTestResult{ID: t.ID, PatientID: t.PatientID, TestName: t.TestName, Result: t.Result, ReferenceRange: t.ReferenceRange, Unit: t.Unit, Notes: t.Notes, CollectedAt: t.CollectedAt, ReportedAt: t.ReportedAt}
	}
	return result, nil
}

func (s *Store) CreateDailyEntry(ctx context.Context, e *model.MedicalDailyEntry) error {
	rec := &models.MedicalDailyEntry{BaseModel: models.BaseModel{ID: e.ID}, PatientID: e.PatientID, EntryDate: e.EntryDate, EntryType: e.EntryType, Content: e.Content, NurseID: e.NurseID}
	if rec.ID == "" {
		rec.ID = uuid.New().String()
	}
	return s.db.WithContext(ctx).Create(rec).Error
}

func (s *Store) CreateVerification(ctx context.Context, v *model.MedicalVerification) error {
	rec := &models.MedicalVerification{BaseModel: models.BaseModel{ID: v.ID}, DeviceID: v.DeviceID, VerificationType: v.VerificationType, Result: v.Result, Matched: v.Matched, VerifiedBy: v.VerifiedBy, Notes: v.Notes}
	if rec.ID == "" {
		rec.ID = uuid.New().String()
	}
	if v.PatientID != nil {
		rec.PatientID = v.PatientID
	}
	if v.VerifiedAt != nil {
		rec.VerifiedAt = v.VerifiedAt
	}
	return s.db.WithContext(ctx).Create(rec).Error
}

func (s *Store) ListVerifications(ctx context.Context, page, pageSize int) ([]model.MedicalVerification, error) {
	query := s.db.WithContext(ctx).Order("created_at DESC")
	if page > 0 && pageSize > 0 {
		query = query.Offset((page - 1) * pageSize).Limit(pageSize)
	}
	var items []models.MedicalVerification
	if err := query.Find(&items).Error; err != nil {
		return nil, fmt.Errorf("list verifications: %w", err)
	}
	result := make([]model.MedicalVerification, len(items))
	for i, v := range items {
		result[i] = model.MedicalVerification{ID: v.ID, DeviceID: v.DeviceID, VerificationType: v.VerificationType, Result: v.Result, Matched: v.Matched, VerifiedBy: v.VerifiedBy, VerifiedAt: v.VerifiedAt, Notes: v.Notes, CreatedAt: v.CreatedAt}
		if v.PatientID != nil {
			result[i].PatientID = v.PatientID
		}
	}
	return result, nil
}

func (s *Store) UpdateVerificationStatus(ctx context.Context, id, status string) error {
	return s.db.WithContext(ctx).Model(&models.MedicalVerification{}).Where("id = ?", id).Update("status", status).Error
}

func (s *Store) GetTodayVerificationStats(ctx context.Context) (*model.MedicalVerificationStats, error) {
	var total, matched, unmatched int64
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	s.db.WithContext(ctx).Model(&models.MedicalVerification{}).Where("verified_at >= ?", todayStart).Count(&total)
	s.db.WithContext(ctx).Model(&models.MedicalVerification{}).Where("verified_at >= ? AND matched = ?", todayStart, true).Count(&matched)
	s.db.WithContext(ctx).Model(&models.MedicalVerification{}).Where("verified_at >= ? AND matched = ?", todayStart, false).Count(&unmatched)
	return &model.MedicalVerificationStats{Total: int(total), Matched: int(matched), Unmatched: int(unmatched)}, nil
}

func (s *Store) GetMedicalStatsOverview(ctx context.Context) (*model.MedicalStatsOverview, error) {
	var overview model.MedicalStatsOverview
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	var yesterdayStart time.Time
	if now.Hour() >= 1 {
		yesterdayStart = todayStart.AddDate(0, 0, -1)
	} else {
		yesterdayStart = todayStart
	}
	var activePatients int64
	s.db.WithContext(ctx).Model(&models.MedicalWristbandPatient{}).Where("status = ?", "admitted").Count(&activePatients)
	overview.ActivePatients = int(activePatients)
	var todayAdmitted int64
	s.db.WithContext(ctx).Model(&models.MedicalWristbandPatient{}).Where("created_at >= ?", todayStart).Count(&todayAdmitted)
	overview.TodayAdmitted = int(todayAdmitted)
	var todayDischarged int64
	s.db.WithContext(ctx).Model(&models.MedicalWristbandPatient{}).Where("updated_at >= ? AND status = ?", yesterdayStart, "discharged").Count(&todayDischarged)
	overview.TodayDischarged = int(todayDischarged)
	var boundDevices int64
	s.db.WithContext(ctx).Model(&models.MedicalBinding{}).Where("unbound_at IS NULL").Count(&boundDevices)
	overview.BoundDevices = int(boundDevices)
	var totalDevices int64
	s.db.WithContext(ctx).Model(&models.MedicalWristbandDevice{}).Count(&totalDevices)
	overview.TotalDevices = int(totalDevices)
	return &overview, nil
}

func (s *Store) CreateAlertTagConfig(ctx context.Context, c *model.MedicalAlertTagConfig) error {
	rec := &models.MedicalAlertTagConfig{BaseModel: models.BaseModel{ID: c.ID}, TagName: c.TagName, TagColor: c.TagColor, TagIcon: c.TagIcon, Enabled: c.Enabled}
	if rec.ID == "" {
		rec.ID = uuid.New().String()
	}
	return s.db.WithContext(ctx).Create(rec).Error
}

func (s *Store) ListAlertTagConfigs(ctx context.Context) ([]model.MedicalAlertTagConfig, error) {
	var items []models.MedicalAlertTagConfig
	if err := s.db.WithContext(ctx).Order("tag_name ASC").Find(&items).Error; err != nil {
		return nil, fmt.Errorf("list alert tag configs: %w", err)
	}
	result := make([]model.MedicalAlertTagConfig, len(items))
	for i, c := range items {
		result[i] = model.MedicalAlertTagConfig{ID: c.ID, TagName: c.TagName, TagColor: c.TagColor, TagIcon: c.TagIcon, Enabled: c.Enabled, CreatedAt: c.CreatedAt, UpdatedAt: c.UpdatedAt}
	}
	return result, nil
}

func (s *Store) CreateCommunityElder(ctx context.Context, e *model.CommunityElder) error {
	if e.ID == "" {
		e.ID = fmt.Sprintf("ce_%d", time.Now().UnixNano())
	}
	e.CreatedAt = time.Now()
	e.UpdatedAt = time.Now()
	rec := &models.CommunityElder{BaseModel: models.BaseModel{ID: e.ID, CreatedAt: e.CreatedAt, UpdatedAt: e.UpdatedAt}, Name: e.Name, IDCard: e.IDCard, Gender: e.Gender, Age: e.Age, Address: e.Address, EmergencyContact: e.EmergencyContact, BankAccount: func(s string) string { if e.BankAccount != nil { return *e.BankAccount }; return s }(""), HospitalID: e.HospitalID, Status: e.Status, DeactivatedAt: e.DeactivatedAt, DeactivatedReason: func(s string) string { if e.DeactivatedReason != nil { return *e.DeactivatedReason }; return s }("")}
	return s.db.WithContext(ctx).Create(rec).Error
}

func (s *Store) GetCommunityElder(ctx context.Context, id string) (*model.CommunityElder, error) {
	var rec models.CommunityElder
	if err := s.db.WithContext(ctx).Where("id = ?", id).First(&rec).Error; err != nil {
		return nil, err
	}
	return &model.CommunityElder{ID: rec.ID, Name: rec.Name, IDCard: rec.IDCard, Gender: rec.Gender, Age: rec.Age, Address: rec.Address, EmergencyContact: ptrString(rec.EmergencyContact), BankAccount: ptrString(&rec.BankAccount), HospitalID: rec.HospitalID, Status: rec.Status, CreatedAt: rec.CreatedAt, UpdatedAt: rec.UpdatedAt, DeactivatedAt: rec.DeactivatedAt, DeactivatedReason: ptrString(&rec.DeactivatedReason)}, nil
}

func (s *Store) ListCommunityElders(ctx context.Context, page, pageSize int, status string) ([]model.CommunityElder, error) {
	query := s.db.WithContext(ctx).Model(&models.CommunityElder{}).Order("created_at DESC")
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if page > 0 && pageSize > 0 {
		query = query.Offset((page - 1) * pageSize).Limit(pageSize)
	}
	var items []models.CommunityElder
	if err := query.Find(&items).Error; err != nil {
		return nil, fmt.Errorf("list community elders: %w", err)
	}
	result := make([]model.CommunityElder, len(items))
	for i, e := range items {
		result[i] = model.CommunityElder{ID: e.ID, Name: e.Name, IDCard: e.IDCard, Gender: e.Gender, Age: e.Age, Address: e.Address, EmergencyContact: ptrString(e.EmergencyContact), BankAccount: ptrString(&e.BankAccount), HospitalID: e.HospitalID, Status: e.Status, CreatedAt: e.CreatedAt, UpdatedAt: e.UpdatedAt, DeactivatedAt: e.DeactivatedAt, DeactivatedReason: ptrString(&e.DeactivatedReason)}
	}
	return result, nil
}

func (s *Store) UpdateCommunityElder(ctx context.Context, e *model.CommunityElder) error {
	e.UpdatedAt = time.Now()
	return s.db.WithContext(ctx).Model(&models.CommunityElder{}).Where("id = ?", e.ID).Updates(map[string]interface{}{"name": e.Name, "id_card": e.IDCard, "gender": e.Gender, "age": e.Age, "address": e.Address, "emergency_contact": e.EmergencyContact, "bank_account": e.BankAccount, "hospital_id": e.HospitalID, "status": e.Status, "updated_at": e.UpdatedAt, "deactivated_at": e.DeactivatedAt, "deactivated_reason": e.DeactivatedReason}).Error
}

func (s *Store) DeleteCommunityElder(ctx context.Context, id string) error {
	return s.db.WithContext(ctx).Model(&models.CommunityElder{}).Where("id = ?", id).Updates(map[string]interface{}{"status": "deactivated", "deactivated_at": time.Now(), "deactivated_reason": "deleted"}).Error
}

func (s *Store) BulkUpsertCommunityElders(ctx context.Context, elders []model.CommunityElder) error {
	for i := range elders {
		e := &elders[i]
		if e.ID == "" {
			e.ID = fmt.Sprintf("ce_%d", time.Now().UnixNano())
		}
		e.CreatedAt = time.Now()
		e.UpdatedAt = time.Now()
		rec := &models.CommunityElder{BaseModel: models.BaseModel{ID: e.ID, CreatedAt: e.CreatedAt, UpdatedAt: e.UpdatedAt}, Name: e.Name, IDCard: e.IDCard, Gender: e.Gender, Age: e.Age, Address: e.Address, EmergencyContact: e.EmergencyContact, BankAccount: func(s string) string { if e.BankAccount != nil { return *e.BankAccount }; return s }(""), HospitalID: e.HospitalID, Status: e.Status}
		if err := s.db.WithContext(ctx).Save(rec).Error; err != nil {
			return fmt.Errorf("bulk upsert elder: %w", err)
		}
	}
	return nil
}

func (s *Store) GetCommunityElderStats(ctx context.Context) (*model.CommunityElderStats, error) {
	stats := &model.CommunityElderStats{}
	var totalElders int64
	s.db.WithContext(ctx).Model(&models.CommunityElder{}).Count(&totalElders)
	stats.TotalElders = int(totalElders)
	var activeElders int64
	s.db.WithContext(ctx).Model(&models.CommunityElder{}).Where("status = ?", "active").Count(&activeElders)
	stats.ActiveElders = int(activeElders)
	today := time.Now().Format("2006-01-02")
	var signinCount int64
	s.db.WithContext(ctx).Model(&models.CommunitySigninRecord{}).Where("strftime('%Y-%m-%d', signin_time) = ?", today).Count(&signinCount)
	stats.TodaySignins = int(signinCount)
	var pharmacyCount int64
	s.db.WithContext(ctx).Model(&models.CommunityPharmacyLog{}).Where("strftime('%Y-%m-%d', dispense_time) = ?", today).Count(&pharmacyCount)
	stats.TodayDispenses = int(pharmacyCount)
	var activeWelfareTags int64
	s.db.WithContext(ctx).Model(&models.CommunityElderWelfare{}).Where("revoked_at IS NULL").Count(&activeWelfareTags)
	stats.ActiveWelfareTags = int(activeWelfareTags)
	return stats, nil
}

func (s *Store) CreateCommunityDevice(ctx context.Context, d *model.CommunityWristbandDevice) error {
	if d.ID == "" {
		d.ID = fmt.Sprintf("cd_%d", time.Now().UnixNano())
	}
	d.CreatedAt = time.Now()
	d.UpdatedAt = time.Now()
	rec := &models.CommunityWristbandDevice{BaseModel: models.BaseModel{ID: d.ID, CreatedAt: d.CreatedAt, UpdatedAt: d.UpdatedAt}, DeviceID: d.DeviceID, FirmwareVersion: d.FirmwareVersion, Mode: d.Mode, Status: d.Status, LastSeen: d.LastSeen}
	return s.db.WithContext(ctx).Create(rec).Error
}

func (s *Store) GetCommunityDevice(ctx context.Context, deviceID string) (*model.CommunityWristbandDevice, error) {
	var rec models.CommunityWristbandDevice
	if err := s.db.WithContext(ctx).Where("device_id = ?", deviceID).First(&rec).Error; err != nil {
		return nil, err
	}
	return &model.CommunityWristbandDevice{ID: rec.ID, DeviceID: rec.DeviceID, FirmwareVersion: rec.FirmwareVersion, Mode: rec.Mode, Status: rec.Status, LastSeen: rec.LastSeen, CreatedAt: rec.CreatedAt, UpdatedAt: rec.UpdatedAt}, nil
}

func (s *Store) ListCommunityDevices(ctx context.Context, page, pageSize int, status string) ([]model.CommunityWristbandDevice, error) {
	query := s.db.WithContext(ctx).Model(&models.CommunityWristbandDevice{}).Order("created_at DESC")
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if page > 0 && pageSize > 0 {
		query = query.Offset((page - 1) * pageSize).Limit(pageSize)
	}
	var items []models.CommunityWristbandDevice
	if err := query.Find(&items).Error; err != nil {
		return nil, fmt.Errorf("list community devices: %w", err)
	}
	result := make([]model.CommunityWristbandDevice, len(items))
	for i, d := range items {
		result[i] = model.CommunityWristbandDevice{ID: d.ID, DeviceID: d.DeviceID, FirmwareVersion: d.FirmwareVersion, Mode: d.Mode, Status: d.Status, LastSeen: d.LastSeen, CreatedAt: d.CreatedAt, UpdatedAt: d.UpdatedAt}
	}
	return result, nil
}

func (s *Store) UpdateCommunityDevice(ctx context.Context, d *model.CommunityWristbandDevice) error {
	d.UpdatedAt = time.Now()
	return s.db.WithContext(ctx).Model(&models.CommunityWristbandDevice{}).Where("id = ?", d.ID).Updates(map[string]interface{}{"firmware_version": d.FirmwareVersion, "status": d.Status, "last_seen": d.LastSeen, "updated_at": d.UpdatedAt}).Error
}

func (s *Store) BindCommunityElderDevice(ctx context.Context, elderID, deviceID string) error {
	id := fmt.Sprintf("cb_%d", time.Now().UnixNano())
	binding := &models.CommunityElderBinding{BaseModel: models.BaseModel{ID: id}, ElderID: elderID, DeviceID: deviceID, BoundAt: time.Now()}
	return s.db.WithContext(ctx).Create(binding).Error
}

func (s *Store) UnbindCommunityElderDevice(ctx context.Context, bindingID string) error {
	return s.db.WithContext(ctx).Model(&models.CommunityElderBinding{}).Where("id = ?", bindingID).Update("unbound_at", time.Now()).Error
}

func (s *Store) CreateWelfareTagConfig(ctx context.Context, c *model.CommunityWelfareTagConfig) error {
	if c.ID == "" {
		c.ID = fmt.Sprintf("wtc_%d", time.Now().UnixNano())
	}
	c.CreatedAt = time.Now()
	c.UpdatedAt = time.Now()
	rec := &models.CommunityWelfareTagConfig{BaseModel: models.BaseModel{ID: c.ID, CreatedAt: c.CreatedAt, UpdatedAt: c.UpdatedAt}, TagCode: c.TagCode, TagName: c.TagName, Issuer: c.Issuer, RenewalPeriodDays: c.RenewalPeriodDays, BenefitAmount: c.BenefitAmount, Enabled: c.Enabled}
	return s.db.WithContext(ctx).Create(rec).Error
}

func (s *Store) UpdateWelfareTagConfig(ctx context.Context, c *model.CommunityWelfareTagConfig) error {
	c.UpdatedAt = time.Now()
	return s.db.WithContext(ctx).Model(&models.CommunityWelfareTagConfig{}).Where("tag_code = ?", c.TagCode).Updates(map[string]interface{}{"tag_name": c.TagName, "issuer": c.Issuer, "renewal_period_days": c.RenewalPeriodDays, "benefit_amount": c.BenefitAmount, "enabled": c.Enabled, "updated_at": c.UpdatedAt}).Error
}

func (s *Store) ListWelfareTagConfigs(ctx context.Context) ([]model.CommunityWelfareTagConfig, error) {
	var items []models.CommunityWelfareTagConfig
	if err := s.db.WithContext(ctx).Order("tag_code ASC").Find(&items).Error; err != nil {
		return nil, fmt.Errorf("list welfare tag configs: %w", err)
	}
	result := make([]model.CommunityWelfareTagConfig, len(items))
	for i, c := range items {
		result[i] = model.CommunityWelfareTagConfig{ID: c.ID, TagCode: c.TagCode, TagName: c.TagName, Issuer: c.Issuer, RenewalPeriodDays: c.RenewalPeriodDays, BenefitAmount: c.BenefitAmount, Enabled: c.Enabled, CreatedAt: c.CreatedAt, UpdatedAt: c.UpdatedAt}
	}
	return result, nil
}

func (s *Store) GetWelfareTagConfig(ctx context.Context, tagCode string) (*model.CommunityWelfareTagConfig, error) {
	var rec models.CommunityWelfareTagConfig
	if err := s.db.WithContext(ctx).Where("tag_code = ?", tagCode).First(&rec).Error; err != nil {
		return nil, err
	}
	return &model.CommunityWelfareTagConfig{ID: rec.ID, TagCode: rec.TagCode, TagName: rec.TagName, Issuer: rec.Issuer, RenewalPeriodDays: rec.RenewalPeriodDays, BenefitAmount: rec.BenefitAmount, Enabled: rec.Enabled}, nil
}

func (s *Store) AssignWelfareTag(ctx context.Context, a *model.CommunityElderWelfare) error {
	if a.ID == "" {
		a.ID = fmt.Sprintf("ewf_%d", time.Now().UnixNano())
	}
	a.EffectiveAt = time.Now()
	rec := &models.CommunityElderWelfare{BaseModel: models.BaseModel{ID: a.ID}, ElderID: a.ElderID, TagCode: a.TagCode, ValidFrom: a.ValidFrom, ValidTo: a.ValidTo, CertifiedBy: a.CertifiedBy, CertificationDoc: a.CertificationDoc, EffectiveAt: a.EffectiveAt}
	return s.db.WithContext(ctx).Create(rec).Error
}

func (s *Store) RevokeWelfareTag(ctx context.Context, elderID, tagCode string) error {
	return s.db.WithContext(ctx).Model(&models.CommunityElderWelfare{}).Where("elder_id = ? AND tag_code = ? AND revoked_at IS NULL", elderID, tagCode).Update("revoked_at", time.Now()).Error
}

func (s *Store) ListElderWelfareTags(ctx context.Context, elderID string) ([]model.CommunityElderWelfare, error) {
	var items []models.CommunityElderWelfare
	if err := s.db.WithContext(ctx).Where("elder_id = ? AND revoked_at IS NULL", elderID).Find(&items).Error; err != nil {
		return nil, fmt.Errorf("list elder welfare tags: %w", err)
	}
	result := make([]model.CommunityElderWelfare, len(items))
	for i, t := range items {
		result[i] = model.CommunityElderWelfare{ID: t.ID, ElderID: t.ElderID, TagCode: t.TagCode, ValidFrom: t.ValidFrom, ValidTo: t.ValidTo, CertifiedBy: t.CertifiedBy, CertificationDoc: t.CertificationDoc, EffectiveAt: t.EffectiveAt, RevokedAt: t.RevokedAt}
	}
	return result, nil
}

func (s *Store) CreateSigninRecord(ctx context.Context, r *model.CommunitySigninRecord) error {
	if r.ID == "" {
		r.ID = fmt.Sprintf("sr_%d", time.Now().UnixNano())
	}
	r.SigninTime = time.Now()
	rec := &models.CommunitySigninRecord{BaseModel: models.BaseModel{ID: r.ID}, ElderID: r.ElderID, DeviceID: r.DeviceID, HospitalID: r.HospitalID, PharmacistID: r.PharmacistID, SigninTime: r.SigninTime, Period: r.Period, IDCard: r.IDCard, ActivatedTags: r.ActivatedTags, IsMedicalSignin: r.IsMedicalSignin, IsWelfareSignin: r.IsWelfareSignin, Notes: r.Notes}
	return s.db.WithContext(ctx).Create(rec).Error
}

func (s *Store) ListSigninRecords(ctx context.Context, elderID, period, hospitalID string, page, pageSize int) ([]model.CommunitySigninRecord, error) {
	query := s.db.WithContext(ctx).Model(&models.CommunitySigninRecord{}).Order("signin_time DESC")
	if elderID != "" {
		query = query.Where("elder_id = ?", elderID)
	}
	if period != "" {
		query = query.Where("period = ?", period)
	}
	if hospitalID != "" {
		query = query.Where("hospital_id = ?", hospitalID)
	}
	if page > 0 && pageSize > 0 {
		query = query.Offset((page - 1) * pageSize).Limit(pageSize)
	}
	var items []models.CommunitySigninRecord
	if err := query.Find(&items).Error; err != nil {
		return nil, fmt.Errorf("list signin records: %w", err)
	}
	result := make([]model.CommunitySigninRecord, len(items))
	for i, r := range items {
		result[i] = model.CommunitySigninRecord{ID: r.ID, ElderID: r.ElderID, DeviceID: r.DeviceID, HospitalID: r.HospitalID, PharmacistID: r.PharmacistID, SigninTime: r.SigninTime, Period: r.Period, IDCard: r.IDCard, ActivatedTags: r.ActivatedTags, IsMedicalSignin: r.IsMedicalSignin, IsWelfareSignin: r.IsWelfareSignin, Notes: r.Notes}
	}
	return result, nil
}

func (s *Store) GetSigninSummary(ctx context.Context, elderID, period string) (*model.CommunitySigninRecord, error) {
	var rec models.CommunitySigninRecord
	if err := s.db.WithContext(ctx).Where("elder_id = ? AND period = ?", elderID, period).Order("signin_time DESC").First(&rec).Error; err != nil {
		return nil, err
	}
	return &model.CommunitySigninRecord{ID: rec.ID, ElderID: rec.ElderID, DeviceID: rec.DeviceID, HospitalID: rec.HospitalID, PharmacistID: rec.PharmacistID, SigninTime: rec.SigninTime, Period: rec.Period, IDCard: rec.IDCard, ActivatedTags: rec.ActivatedTags, IsMedicalSignin: rec.IsMedicalSignin, IsWelfareSignin: rec.IsWelfareSignin, Notes: rec.Notes}, nil
}

func (s *Store) CreatePharmacyLog(ctx context.Context, p *model.CommunityPharmacyLog) error {
	if p.ID == "" {
		p.ID = fmt.Sprintf("pl_%d", time.Now().UnixNano())
	}
	p.DispenseTime = time.Now()
	rec := &models.CommunityPharmacyLog{BaseModel: models.BaseModel{ID: p.ID}, ElderID: p.ElderID, DeviceID: p.DeviceID, HospitalID: p.HospitalID, PharmacistID: p.PharmacistID, DispenseTime: p.DispenseTime, Period: p.Period, Items: p.Items, TotalCost: p.TotalCost, InsuranceCovered: p.InsuranceCovered, SelfPay: p.SelfPay, Notes: p.Notes}
	return s.db.WithContext(ctx).Create(rec).Error
}

func (s *Store) ListPharmacyLogs(ctx context.Context, elderID, period string, page, pageSize int) ([]model.CommunityPharmacyLog, error) {
	query := s.db.WithContext(ctx).Model(&models.CommunityPharmacyLog{}).Order("dispense_time DESC")
	if elderID != "" {
		query = query.Where("elder_id = ?", elderID)
	}
	if period != "" {
		query = query.Where("period = ?", period)
	}
	if page > 0 && pageSize > 0 {
		query = query.Offset((page - 1) * pageSize).Limit(pageSize)
	}
	var items []models.CommunityPharmacyLog
	if err := query.Find(&items).Error; err != nil {
		return nil, fmt.Errorf("list pharmacy logs: %w", err)
	}
	result := make([]model.CommunityPharmacyLog, len(items))
	for i, p := range items {
		result[i] = model.CommunityPharmacyLog{ID: p.ID, ElderID: p.ElderID, DeviceID: p.DeviceID, HospitalID: p.HospitalID, PharmacistID: p.PharmacistID, DispenseTime: p.DispenseTime, Period: p.Period, Items: p.Items, TotalCost: p.TotalCost, InsuranceCovered: p.InsuranceCovered, SelfPay: p.SelfPay, Notes: p.Notes}
	}
	return result, nil
}

func (s *Store) CreateMinzhengSync(ctx context.Context, m *model.CommunityMinzhengSync) error {
	if m.ID == "" {
		m.ID = fmt.Sprintf("ms_%d", time.Now().UnixNano())
	}
	m.CreatedAt = time.Now()
	rec := &models.CommunityMinzhengSync{BaseModel: models.BaseModel{ID: m.ID, CreatedAt: m.CreatedAt}, Source: m.Source, Filename: m.Filename, ImportedCount: m.ImportedCount, MatchedCount: m.MatchedCount, PendingReviewCount: m.PendingReviewCount, ErrorCount: m.ErrorCount, Status: m.Status, CompletedAt: m.CompletedAt}
	return s.db.WithContext(ctx).Create(rec).Error
}

func (s *Store) ListMinzhengSync(ctx context.Context, page, pageSize int) ([]model.CommunityMinzhengSync, error) {
	query := s.db.WithContext(ctx).Model(&models.CommunityMinzhengSync{}).Order("created_at DESC")
	if page > 0 && pageSize > 0 {
		query = query.Offset((page - 1) * pageSize).Limit(pageSize)
	}
	var items []models.CommunityMinzhengSync
	if err := query.Find(&items).Error; err != nil {
		return nil, fmt.Errorf("list minzheng sync: %w", err)
	}
	result := make([]model.CommunityMinzhengSync, len(items))
	for i, m := range items {
		result[i] = model.CommunityMinzhengSync{ID: m.ID, Source: m.Source, Filename: m.Filename, ImportedCount: m.ImportedCount, MatchedCount: m.MatchedCount, PendingReviewCount: m.PendingReviewCount, ErrorCount: m.ErrorCount, Status: m.Status, CreatedAt: m.CreatedAt, CompletedAt: m.CompletedAt}
	}
	return result, nil
}

func (s *Store) GetLatestMinzhengSync(ctx context.Context) (*model.CommunityMinzhengSync, error) {
	var rec models.CommunityMinzhengSync
	if err := s.db.WithContext(ctx).Order("created_at DESC").First(&rec).Error; err != nil {
		return nil, err
	}
	return &model.CommunityMinzhengSync{ID: rec.ID, Source: rec.Source, Filename: rec.Filename, ImportedCount: rec.ImportedCount, MatchedCount: rec.MatchedCount, PendingReviewCount: rec.PendingReviewCount, ErrorCount: rec.ErrorCount, Status: rec.Status, CreatedAt: rec.CreatedAt, CompletedAt: rec.CompletedAt}, nil
}

func (s *Store) CreateBatchPayment(ctx context.Context, p *model.CommunityBatchPayment) error {
	if p.ID == "" {
		p.ID = fmt.Sprintf("bp_%d", time.Now().UnixNano())
	}
	p.CreatedAt = time.Now()
	rec := &models.CommunityBatchPayment{BaseModel: models.BaseModel{ID: p.ID, CreatedAt: p.CreatedAt}, BatchID: p.BatchID, Period: p.Period, PayType: p.PayType, ElderID: p.ElderID, Amount: p.Amount, BankAccount: p.BankAccount, Status: p.Status, FailureReason: p.FailureReason, ExecutedAt: p.ExecutedAt}
	return s.db.WithContext(ctx).Create(rec).Error
}

func (s *Store) BulkCreateBatchPayments(ctx context.Context, payments []model.CommunityBatchPayment) error {
	now := time.Now()
	for i := range payments {
		p := &payments[i]
		if p.ID == "" {
			p.ID = fmt.Sprintf("bp_%d", now.UnixNano())
		}
		p.CreatedAt = now
		rec := &models.CommunityBatchPayment{BaseModel: models.BaseModel{ID: p.ID, CreatedAt: p.CreatedAt}, BatchID: p.BatchID, Period: p.Period, PayType: p.PayType, ElderID: p.ElderID, Amount: p.Amount, BankAccount: p.BankAccount, Status: p.Status, FailureReason: p.FailureReason, ExecutedAt: p.ExecutedAt}
		if err := s.db.WithContext(ctx).Create(rec).Error; err != nil {
			return fmt.Errorf("bulk create batch payment: %w", err)
		}
	}
	return nil
}

func (s *Store) ListBatchPayments(ctx context.Context, batchID string, page, pageSize int) ([]model.CommunityBatchPayment, error) {
	query := s.db.WithContext(ctx).Model(&models.CommunityBatchPayment{}).Order("created_at DESC")
	if batchID != "" {
		query = query.Where("batch_id = ?", batchID)
	}
	if page > 0 && pageSize > 0 {
		query = query.Offset((page - 1) * pageSize).Limit(pageSize)
	}
	var items []models.CommunityBatchPayment
	if err := query.Find(&items).Error; err != nil {
		return nil, fmt.Errorf("list batch payments: %w", err)
	}
	result := make([]model.CommunityBatchPayment, len(items))
	for i, p := range items {
		result[i] = model.CommunityBatchPayment{ID: p.ID, BatchID: p.BatchID, Period: p.Period, PayType: p.PayType, ElderID: p.ElderID, Amount: p.Amount, BankAccount: p.BankAccount, Status: p.Status, FailureReason: p.FailureReason, ExecutedAt: p.ExecutedAt, CreatedAt: p.CreatedAt}
	}
	return result, nil
}

func (s *Store) UpdateBatchPaymentStatus(ctx context.Context, id, status string, failureReason string) error {
	return s.db.WithContext(ctx).Model(&models.CommunityBatchPayment{}).Where("id = ?", id).Updates(map[string]interface{}{"status": status, "failure_reason": failureReason, "executed_at": time.Now()}).Error
}

func (s *Store) CountPendingPayments(ctx context.Context) (int64, error) {
	var count int64
	err := s.db.WithContext(ctx).Model(&models.CommunityBatchPayment{}).Where("status = ?", "pending").Count(&count).Error
	return count, err
}

func (s *Store) ListMedicationRules(ctx context.Context, personID string, chain model.BusinessChain) ([]model.MedicationRuleRow, error) {
	query := s.db.WithContext(ctx).Model(&models.MedicationRuleV2{}).Where("person_id = ?", personID)
	if chain != "" {
		query = query.Where("business_chain = ?", chain)
	}
	var items []models.MedicationRuleV2
	if err := query.Order("created_at DESC").Find(&items).Error; err != nil {
		return nil, fmt.Errorf("list medication rules: %w", err)
	}
	result := make([]model.MedicationRuleRow, len(items))
	for i, r := range items {
		result[i] = model.MedicationRuleRow{
			ID: r.ID, PersonID: r.PersonID, BusinessChain: r.BusinessChain,
			SourceType: r.SourceType, SourceID: r.SourceID, DrugName: r.DrugName,
			GenericName: r.GenericName, DrugCategory: r.DrugCategory, Dosage: r.Dosage,
			Frequency: r.Frequency, Route: r.Route, ScheduleTime1: r.ScheduleTime1,
			ScheduleTime2: r.ScheduleTime2, ScheduleTime3: r.ScheduleTime3, DaysOfWeek: r.DaysOfWeek,
			Duration: r.Duration, PreMeal: boolToInt(r.PreMeal), PostMeal: boolToInt(r.PostMeal),
			SpecialInstructions: r.SpecialInstructions, PrescribedBy: r.PrescribedBy,
			PrescribedAt: r.PrescribedAt, Active: r.Active, CreatedAt: r.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}
	return result, nil
}

func (s *Store) CreateMedicationRuleV2(ctx context.Context, r *model.MedicationRuleRow) error {
	rec := &models.MedicationRuleV2{
		BaseModel: models.BaseModel{ID: r.ID}, PersonID: r.PersonID, BusinessChain: r.BusinessChain,
		SourceType: r.SourceType, SourceID: r.SourceID, DrugName: r.DrugName,
		GenericName: r.GenericName, DrugCategory: r.DrugCategory, Dosage: r.Dosage,
		Frequency: r.Frequency, Route: r.Route, ScheduleTime1: r.ScheduleTime1,
		ScheduleTime2: r.ScheduleTime2, ScheduleTime3: r.ScheduleTime3, DaysOfWeek: r.DaysOfWeek,
		Duration: r.Duration, PreMeal: intBool(r.PreMeal), PostMeal: intBool(r.PostMeal),
		SpecialInstructions: r.SpecialInstructions, PrescribedBy: r.PrescribedBy,
		PrescribedAt: r.PrescribedAt, Active: true,
	}
	if rec.ID == "" {
		rec.ID = uuid.New().String()
	}
	return s.db.WithContext(ctx).Create(rec).Error
}

func (s *Store) UpdateMedicationRuleV2(ctx context.Context, id string, updates map[string]any) error {
	if len(updates) == 0 {
		return nil
	}
	updates["updated_at"] = time.Now()
	return s.db.WithContext(ctx).Model(&models.MedicationRuleV2{}).Where("id = ?", id).Updates(updates).Error
}

func (s *Store) DeleteMedicationRuleV2(ctx context.Context, id string) error {
	return s.db.WithContext(ctx).Delete(&models.MedicationRuleV2{}, "id = ?", id).Error
}

func (s *Store) CreateMedicationExecution(ctx context.Context, e *model.MedicationExecution) error {
	rec := &models.MedicationExecution{
		BaseModel: models.BaseModel{ID: e.ID}, PersonID: e.PersonID, BusinessChain: e.BusinessChain,
		RuleID: e.RuleID, ScheduledTime: e.ScheduledTime, ActualTime: e.ActualTime,
		Status: e.Status, TakenBy: e.TakenBy, DeviceID: e.DeviceID,
		VerificationMethod: e.VerificationMethod, Notes: e.Notes,
	}
	if rec.ID == "" {
		rec.ID = uuid.New().String()
	}
	return s.db.WithContext(ctx).Create(rec).Error
}

func (s *Store) ListMedicationExecutions(ctx context.Context, personID string, chain model.BusinessChain, limit int) ([]model.MedicationExecution, error) {
	query := s.db.WithContext(ctx).Model(&models.MedicationExecution{}).Where("person_id = ?", personID)
	if chain != "" {
		query = query.Where("business_chain = ?", chain)
	}
	if limit > 0 {
		query = query.Limit(limit)
	}
	query = query.Order("scheduled_time DESC")
	var items []models.MedicationExecution
	if err := query.Find(&items).Error; err != nil {
		return nil, fmt.Errorf("list medication executions: %w", err)
	}
	result := make([]model.MedicationExecution, len(items))
	for i, e := range items {
		result[i] = model.MedicationExecution{
			ID: e.ID, PersonID: e.PersonID, BusinessChain: e.BusinessChain,
			RuleID: e.RuleID, ScheduledTime: e.ScheduledTime, ActualTime: e.ActualTime,
			Status: e.Status, TakenBy: e.TakenBy, DeviceID: e.DeviceID,
			VerificationMethod: e.VerificationMethod, Notes: e.Notes, CreatedAt: e.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}
	return result, nil
}

func (s *Store) AssignRole(ctx context.Context, binding *model.PersonRoleBinding) error {
	binding.ID = uuid.New().String()
	binding.CreatedAt = time.Now()
	rec := &models.UserRoleBinding{
		BaseModel: models.BaseModel{ID: binding.ID, CreatedAt: binding.CreatedAt},
		UserID: binding.UserID, BusinessChain: binding.BusinessChain, Role: binding.Role,
		InstitutionID: binding.InstitutionID, GrantedBy: binding.GrantedBy,
		ExpiresAt: binding.ExpiresAt, Active: binding.Active != 0,
	}
	return s.db.WithContext(ctx).Create(rec).Error
}

func (s *Store) ListRoles(ctx context.Context, userID string) ([]model.PersonRoleBinding, error) {
	var items []models.UserRoleBinding
	if err := s.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at DESC").Find(&items).Error; err != nil {
		return nil, fmt.Errorf("list roles: %w", err)
	}
	result := make([]model.PersonRoleBinding, len(items))
	for i, b := range items {
		result[i] = model.PersonRoleBinding{
			ID: b.ID, UserID: b.UserID, BusinessChain: b.BusinessChain, Role: b.Role,
			InstitutionID: b.InstitutionID, GrantedBy: b.GrantedBy, ExpiresAt: b.ExpiresAt,
			Active: boolToInt(b.Active), CreatedAt: b.CreatedAt,
		}
	}
	return result, nil
}

func (s *Store) ListRolesByChain(ctx context.Context, chain model.BusinessChain) ([]model.PersonRoleBinding, error) {
	var items []models.UserRoleBinding
	if err := s.db.WithContext(ctx).Where("business_chain = ? AND active = ?", chain, true).Order("created_at DESC").Find(&items).Error; err != nil {
		return nil, fmt.Errorf("list roles by chain: %w", err)
	}
	result := make([]model.PersonRoleBinding, len(items))
	for i, b := range items {
		result[i] = model.PersonRoleBinding{
			ID: b.ID, UserID: b.UserID, BusinessChain: b.BusinessChain, Role: b.Role,
			InstitutionID: b.InstitutionID, GrantedBy: b.GrantedBy, ExpiresAt: b.ExpiresAt,
			Active: boolToInt(b.Active), CreatedAt: b.CreatedAt,
		}
	}
	return result, nil
}

func (s *Store) RevokeRole(ctx context.Context, bindingID string) error {
	return s.db.WithContext(ctx).Model(&models.UserRoleBinding{}).Where("id = ?", bindingID).Update("active", false).Error
}

func (s *Store) GetEffectiveRole(ctx context.Context, userID string, chain model.BusinessChain) (string, bool) {
	var role string
	err := s.db.WithContext(ctx).Model(&models.UserRoleBinding{}).
		Where("user_id = ? AND business_chain = ? AND active = ?", userID, chain, true).
		Order("granted_at DESC").
		Select("role").First(&role).Error
	if err != nil {
		return "", false
	}
	return role, true
}

func (s *Store) CreateAlertRule(ctx context.Context, r *model.AlertRule) error {
	rec := &models.AlertRuleGorm{
		BaseModel: models.BaseModel{ID: r.ID}, Name: r.Name, BusinessChain: r.BusinessChain,
		AlertType: r.AlertType, Severity: r.Severity, ConditionField: r.ConditionField,
		ConditionOperator: r.ConditionOperator, NotifyRoles: r.NotifyRoles,
		NotifyChannels: r.NotifyChannels, EscalationTimeoutMin: r.EscalationTimeoutMin,
		Enabled: r.Active != 0,
	}
	if rec.ID == "" {
		rec.ID = uuid.New().String()
	}
	if r.ConditionThreshold != nil {
		t := int(*r.ConditionThreshold)
		rec.ConditionThreshold = &t
	}
	if r.ConditionDurationMin != nil {
		d := int(*r.ConditionDurationMin)
		rec.ConditionDurationMin = &d
	}
	return s.db.WithContext(ctx).Create(rec).Error
}

func (s *Store) GetAlertRule(ctx context.Context, id string) (*model.AlertRule, error) {
	var rec models.AlertRuleGorm
	if err := s.db.WithContext(ctx).Where("id = ?", id).First(&rec).Error; err != nil {
		return nil, err
	}
	r := &model.AlertRule{
		ID: rec.ID, Name: rec.Name, BusinessChain: rec.BusinessChain,
		AlertType: rec.AlertType, Severity: rec.Severity, ConditionField: rec.ConditionField,
		ConditionOperator: rec.ConditionOperator, NotifyRoles: rec.NotifyRoles,
		NotifyChannels: rec.NotifyChannels, EscalationTimeoutMin: rec.EscalationTimeoutMin,
		Active: boolToInt(rec.Enabled), CreatedAt: rec.CreatedAt, UpdatedAt: rec.UpdatedAt,
	}
	if rec.ConditionThreshold != nil {
		v := int(*rec.ConditionThreshold)
		r.ConditionThreshold = &v
	}
	if rec.ConditionDurationMin != nil {
		v := int(*rec.ConditionDurationMin)
		r.ConditionDurationMin = &v
	}
	return r, nil
}

func (s *Store) ListAlertRules(ctx context.Context, chain model.BusinessChain) ([]model.AlertRule, error) {
	var items []models.AlertRuleGorm
	if err := s.db.WithContext(ctx).Where("business_chain = ?", chain).Order("created_at DESC").Find(&items).Error; err != nil {
		return nil, fmt.Errorf("list alert rules: %w", err)
	}
	result := make([]model.AlertRule, len(items))
	for i, r := range items {
		item := model.AlertRule{
			ID: r.ID, Name: r.Name, BusinessChain: r.BusinessChain,
			AlertType: r.AlertType, Severity: r.Severity, ConditionField: r.ConditionField,
			ConditionOperator: r.ConditionOperator, NotifyRoles: r.NotifyRoles,
			NotifyChannels: r.NotifyChannels, EscalationTimeoutMin: r.EscalationTimeoutMin,
			Active: boolToInt(r.Enabled), CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt,
		}
		if r.ConditionThreshold != nil {
			v := int(*r.ConditionThreshold)
			item.ConditionThreshold = &v
		}
		if r.ConditionDurationMin != nil {
			v := int(*r.ConditionDurationMin)
			item.ConditionDurationMin = &v
		}
		result[i] = item
	}
	return result, nil
}

func (s *Store) UpdateAlertRule(ctx context.Context, id string, updates map[string]any) error {
	if len(updates) == 0 {
		return nil
	}
	updates["updated_at"] = time.Now()
	return s.db.WithContext(ctx).Model(&models.AlertRuleGorm{}).Where("id = ?", id).Updates(updates).Error
}

func (s *Store) DeleteAlertRule(ctx context.Context, id string) error {
	return s.db.WithContext(ctx).Delete(&models.AlertRuleGorm{}, "id = ?", id).Error
}

func (s *Store) CreateGuidanceRule(ctx context.Context, r *model.HealthGuidanceRule) error {
	rec := &models.HealthGuidanceRule{
		BaseModel: models.BaseModel{ID: r.ID}, Name: r.Name, BusinessChain: r.BusinessChain,
		TriggerCondition: r.TriggerCondition, ConditionField: r.ConditionField,
		ConditionOp: r.ConditionOp, ConditionThresh: r.ConditionThresh,
		GuidanceType: r.GuidanceType, Title: r.Title, Content: r.Content,
		Priority: r.Priority, Enabled: r.Enabled != 0,
	}
	if rec.ID == "" {
		rec.ID = uuid.New().String()
	}
	return s.db.WithContext(ctx).Create(rec).Error
}

func (s *Store) ListGuidanceRules(ctx context.Context, chain model.BusinessChain, enabledOnly bool) ([]model.HealthGuidanceRule, error) {
	query := s.db.WithContext(ctx).Model(&models.HealthGuidanceRule{}).Where("business_chain = ?", chain)
	if enabledOnly {
		query = query.Where("enabled = ?", true)
	}
	query = query.Order("priority DESC, created_at DESC")
	var items []models.HealthGuidanceRule
	if err := query.Find(&items).Error; err != nil {
		return nil, fmt.Errorf("list guidance rules: %w", err)
	}
	result := make([]model.HealthGuidanceRule, len(items))
	for i, r := range items {
		result[i] = model.HealthGuidanceRule{
			ID: r.ID, Name: r.Name, BusinessChain: r.BusinessChain,
			TriggerCondition: r.TriggerCondition, ConditionField: r.ConditionField,
			ConditionOp: r.ConditionOp, ConditionThresh: r.ConditionThresh,
			GuidanceType: r.GuidanceType, Title: r.Title, Content: r.Content,
			Priority: r.Priority, Enabled: boolToInt(r.Enabled), CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt,
		}
	}
	return result, nil
}

func (s *Store) EvaluateGuidanceRules(ctx context.Context, personID string, chain model.BusinessChain, healthData map[string]any) ([]model.HealthGuidanceRule, error) {
	rules, err := s.ListGuidanceRules(ctx, chain, true)
	if err != nil {
		return nil, err
	}
	return rules, nil
}

func (s *Store) CreateGuidanceDelivery(ctx context.Context, d *model.HealthGuidanceDelivery) error {
	rec := &models.HealthGuidanceDelivery{
		BaseModel: models.BaseModel{ID: d.ID}, PersonID: d.PersonID, BusinessChain: d.BusinessChain,
		RuleID: d.RuleID, GuidanceType: d.GuidanceType, Title: d.Title,
		Content: d.Content, Channel: d.Channel, DeliveredAt: d.DeliveredAt,
		ReadStatus: d.ReadStatus, Feedback: d.Feedback,
	}
	if rec.ID == "" {
		rec.ID = uuid.New().String()
	}
	return s.db.WithContext(ctx).Create(rec).Error
}

func (s *Store) ListGuidanceDeliveries(ctx context.Context, personID string, chain model.BusinessChain, limit int) ([]model.HealthGuidanceDelivery, error) {
	query := s.db.WithContext(ctx).Model(&models.HealthGuidanceDelivery{}).
		Where("person_id = ? AND business_chain = ?", personID, chain).
		Order("delivered_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	var items []models.HealthGuidanceDelivery
	if err := query.Find(&items).Error; err != nil {
		return nil, fmt.Errorf("list guidance deliveries: %w", err)
	}
	result := make([]model.HealthGuidanceDelivery, len(items))
	for i, d := range items {
		result[i] = model.HealthGuidanceDelivery{
			ID: d.ID, PersonID: d.PersonID, BusinessChain: d.BusinessChain,
			RuleID: d.RuleID, GuidanceType: d.GuidanceType, Title: d.Title,
			Content: d.Content, Channel: d.Channel, DeliveredAt: d.DeliveredAt,
			ReadStatus: d.ReadStatus, Feedback: d.Feedback,
		}
	}
	return result, nil
}

func (s *Store) CreateReportTemplate(ctx context.Context, t *model.HealthReportTemplate) error {
	rec := &models.HealthReportTemplate{
		BaseModel: models.BaseModel{ID: t.ID}, Name: t.Name,
		BusinessChain: t.BusinessChain, Frequency: t.Frequency, TemplateType: t.TemplateType,
	}
	if rec.ID == "" {
		rec.ID = uuid.New().String()
	}
	return s.db.WithContext(ctx).Create(rec).Error
}

func (s *Store) ListReportTemplates(ctx context.Context, chain model.BusinessChain) ([]model.HealthReportTemplate, error) {
	var items []models.HealthReportTemplate
	if err := s.db.WithContext(ctx).Where("business_chain = ?", chain).Order("created_at DESC").Find(&items).Error; err != nil {
		return nil, fmt.Errorf("list report templates: %w", err)
	}
	result := make([]model.HealthReportTemplate, len(items))
	for i, t := range items {
		result[i] = model.HealthReportTemplate{
			ID: t.ID, Name: t.Name, BusinessChain: t.BusinessChain,
			Frequency: t.Frequency, TemplateType: t.TemplateType,
		}
	}
	return result, nil
}

func (s *Store) CreateReport(ctx context.Context, r *model.HealthReport) error {
	if r.ID == "" {
		r.ID = uuid.New().String()
	}
	if r.Status == "" {
		r.Status = "generated"
	}
	rec := &models.HealthReport{
		BaseModel: models.BaseModel{ID: r.ID}, PersonID: r.PersonID,
		BusinessChain: r.BusinessChain, TemplateID: r.TemplateID,
		ReportPeriod: r.ReportPeriodStart.Format("2006-01-02") + " to " + r.ReportPeriodEnd.Format("2006-01-02"),
		Content: r.ReportData,
	}
	return s.db.WithContext(ctx).Create(rec).Error
}

func (s *Store) ListReports(ctx context.Context, personID string, chain model.BusinessChain, limit int) ([]model.HealthReport, error) {
	query := s.db.WithContext(ctx).Model(&models.HealthReport{}).Where("person_id = ? AND business_chain = ?", personID, chain)
	if limit > 0 {
		query = query.Limit(limit)
	}
	query = query.Order("created_at DESC")
	var items []models.HealthReport
	if err := query.Find(&items).Error; err != nil {
		return nil, fmt.Errorf("list reports: %w", err)
	}
	result := make([]model.HealthReport, len(items))
	for i, r := range items {
		result[i] = model.HealthReport{
			ID: r.ID, PersonID: r.PersonID, BusinessChain: r.BusinessChain,
			TemplateID: r.TemplateID,
		}
	}
	return result, nil
}

func (s *Store) CreateComplianceRule(ctx context.Context, r *model.ComplianceRule) error {
	rec := &models.ComplianceRule{
		BaseModel: models.BaseModel{ID: r.ID}, RuleCode: r.RuleCode,
		Name: r.Name, Description: r.Description, BusinessChain: r.BusinessChain,
		Condition: r.ConditionSQL, Action: r.ActionRequired, Enabled: r.Enabled != 0,
	}
	if rec.ID == "" {
		rec.ID = uuid.New().String()
	}
	return s.db.WithContext(ctx).Create(rec).Error
}

func (s *Store) ListComplianceRules(ctx context.Context, chain model.BusinessChain) ([]model.ComplianceRule, error) {
	var items []models.ComplianceRule
	if err := s.db.WithContext(ctx).Where("business_chain = ?", chain).Order("created_at DESC").Find(&items).Error; err != nil {
		return nil, fmt.Errorf("list compliance rules: %w", err)
	}
	result := make([]model.ComplianceRule, len(items))
	for i, r := range items {
		result[i] = model.ComplianceRule{
			ID: r.ID, RuleCode: r.RuleCode, Name: r.Name, Description: r.Description,
			BusinessChain: r.BusinessChain,
			ActionRequired: r.Action, Enabled: boolToInt(r.Enabled),
			CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt,
		}
	}
	return result, nil
}

func (s *Store) RunComplianceCheck(ctx context.Context, ruleCode string, personID string) (*model.ComplianceCheck, error) {
	check := &model.ComplianceCheck{
		ID: uuid.New().String(), RuleID: ruleCode, PersonID: personID,
		CheckTime: time.Now(), CreatedAt: time.Now(),
	}
	rec := &models.ComplianceCheck{
		BaseModel: models.BaseModel{ID: check.ID}, RuleID: check.RuleID,
		PersonID: check.PersonID, Violated: check.Violated != 0,
		Result: check.ViolationDetails, Notes: check.ViolationDetails,
	}
	if err := s.db.WithContext(ctx).Create(rec).Error; err != nil {
		return nil, err
	}
	return check, nil
}

func (s *Store) ListComplianceChecks(ctx context.Context, personID string, limit int) ([]model.ComplianceCheck, error) {
	query := s.db.WithContext(ctx).Model(&models.ComplianceCheck{}).Where("person_id = ?", personID)
	if limit > 0 {
		query = query.Limit(limit)
	}
	query = query.Order("created_at DESC")
	var items []models.ComplianceCheck
	if err := query.Find(&items).Error; err != nil {
		return nil, fmt.Errorf("list compliance checks: %w", err)
	}
	result := make([]model.ComplianceCheck, len(items))
	for i, c := range items {
		result[i] = model.ComplianceCheck{
			ID: c.ID, RuleID: c.RuleID, PersonID: c.PersonID,
			CheckTime: c.CreatedAt, Violated: boolToInt(c.Violated),
			ViolationDetails: c.Result, ReviewedBy: c.ReviewerID,
			ActionTaken: c.Notes,
		}
	}
	return result, nil
}

func (s *Store) ReviewCheck(ctx context.Context, checkID string, reviewerID string, result string, notes string) error {
	return s.db.WithContext(ctx).Model(&models.ComplianceCheck{}).Where("id = ?", checkID).Updates(map[string]interface{}{
		"reviewer_id": reviewerID, "result": result, "notes": notes,
	}).Error
}

func (s *Store) BindDevice(ctx context.Context, binding *model.DeviceBinding) error {
	binding.ID = uuid.New().String()
	binding.BoundAt = time.Now()
	rec := &models.DeviceBinding{
		BaseModel: models.BaseModel{ID: binding.ID}, DeviceID: binding.DeviceID,
		PersonID: binding.PersonID, BusinessChain: binding.BusinessChain,
	}
	return s.db.WithContext(ctx).Create(rec).Error
}

func (s *Store) ListDeviceBindings(ctx context.Context, personID string, chain model.BusinessChain) ([]model.DeviceBinding, error) {
	query := s.db.WithContext(ctx).Model(&models.DeviceBinding{}).Where("person_id = ?", personID)
	if chain != "" {
		query = query.Where("business_chain = ?", chain)
	}
	var items []models.DeviceBinding
	if err := query.Order("created_at DESC").Find(&items).Error; err != nil {
		return nil, fmt.Errorf("list device bindings: %w", err)
	}
	result := make([]model.DeviceBinding, len(items))
	for i, b := range items {
		result[i] = model.DeviceBinding{
			ID: b.ID, DeviceID: b.DeviceID, PersonID: b.PersonID,
			BusinessChain: b.BusinessChain,
		}
	}
	return result, nil
}

func (s *Store) ListDevicesByPerson(ctx context.Context, personID string) ([]model.DeviceSummary, error) {
	var devices []models.Device
	if err := s.db.WithContext(ctx).Joins("JOIN device_bindings db ON devices.id = db.device_id").
		Where("db.person_id = ?", personID).Find(&devices).Error; err != nil {
		return nil, fmt.Errorf("list devices by person: %w", err)
	}
	result := make([]model.DeviceSummary, len(devices))
	for i, d := range devices {
		result[i] = model.DeviceSummary{
			ID: d.ID, DeviceID: d.DeviceID, Type: d.DeviceType,
			Tier: d.Tier, Status: d.Status, FirmwareVer: d.OTAURL,
		}
	}
	return result, nil
}

func (s *Store) CreateNotificationTemplate(ctx context.Context, t *model.NotificationTemplate) error {
	rec := &models.NotificationTemplate{
		BaseModel: models.BaseModel{ID: t.ID}, Name: t.Name,
		BusinessChain: t.BusinessChain, Channel: t.Channel,
		Subject: t.Subject, Content: t.BodyTemplate, Enabled: t.Enabled != 0,
	}
	if rec.ID == "" {
		rec.ID = uuid.New().String()
	}
	return s.db.WithContext(ctx).Create(rec).Error
}

func (s *Store) ListNotificationTemplates(ctx context.Context, chain model.BusinessChain) ([]model.NotificationTemplate, error) {
	var items []models.NotificationTemplate
	if err := s.db.WithContext(ctx).Where("business_chain = ?", chain).Order("created_at DESC").Find(&items).Error; err != nil {
		return nil, fmt.Errorf("list notification templates: %w", err)
	}
	result := make([]model.NotificationTemplate, len(items))
	for i, t := range items {
		result[i] = model.NotificationTemplate{
			ID: t.ID, Name: t.Name, BusinessChain: t.BusinessChain,
			Channel: t.Channel, Subject: t.Subject, BodyTemplate: t.Content,
			Enabled: boolToInt(t.Enabled), CreatedAt: t.CreatedAt,
		}
	}
	return result, nil
}

func (s *Store) CreateNotificationLog(ctx context.Context, l *model.NotificationLog) error {
	rec := &models.NotificationLog{
		BaseModel: models.BaseModel{ID: l.ID}, PersonID: l.PersonID,
		BusinessChain: l.BusinessChain, TemplateID: l.TemplateID,
		Channel: l.Channel, Status: l.Status,
	}
	if rec.ID == "" {
		rec.ID = uuid.New().String()
	}
	return s.db.WithContext(ctx).Create(rec).Error
}

func (s *Store) UpdateNotificationStatus(ctx context.Context, logID string, status string, sentAt, readAt *time.Time) error {
	updates := map[string]interface{}{"status": status}
	if sentAt != nil {
		updates["sent_at"] = sentAt
	}
	if readAt != nil {
		updates["read_at"] = readAt
	}
	return s.db.WithContext(ctx).Model(&models.NotificationLog{}).Where("id = ?", logID).Updates(updates).Error
}

func (s *Store) ListNotificationLogs(ctx context.Context, personID string, chain model.BusinessChain, limit int) ([]model.NotificationLog, error) {
	query := s.db.WithContext(ctx).Model(&models.NotificationLog{}).Where("person_id = ? AND business_chain = ?", personID, chain)
	if limit > 0 {
		query = query.Limit(limit)
	}
	query = query.Order("created_at DESC")
	var items []models.NotificationLog
	if err := query.Find(&items).Error; err != nil {
		return nil, fmt.Errorf("list notification logs: %w", err)
	}
	result := make([]model.NotificationLog, len(items))
	for i, l := range items {
		result[i] = model.NotificationLog{
			ID: l.ID, PersonID: l.PersonID, BusinessChain: l.BusinessChain,
			TemplateID: l.TemplateID, Channel: l.Channel, Status: l.Status,
			SentAt: l.SentAt, ReadAt: l.ReadAt,
			CreatedAt: l.CreatedAt,
		}
	}
	return result, nil
}

func (s *Store) CreateHealthRecordV2(ctx context.Context, r *model.HealthRecordV2) error {
	rec := &models.HealthRecordV2{
		BaseModel: models.BaseModel{ID: r.ID}, PersonID: r.PersonID,
		BusinessChain: r.BusinessChain, RecordType: r.RecordType,
		Source: r.Source, DeviceID: r.DeviceID, RecordedAt: r.RecordedAt,
		HeartRate: r.HeartRate, BloodPressureSys: r.BloodPressureSys,
		BloodPressureDia: r.BloodPressureDia, SpO2: r.SpO2,
		Temperature: r.Temperature, RespiratoryRate: r.RespiratoryRate,
		PulseRate: r.PulseRate, GlucoseFasting: r.GlucoseFasting,
		UricAcid: r.UricAcid, Steps: r.Steps, SleepHours: r.SleepHours,
		Content: r.Notes,
	}
	if rec.ID == "" {
		rec.ID = uuid.New().String()
	}
	return s.db.WithContext(ctx).Create(rec).Error
}

func (s *Store) ListHealthRecordsV2(ctx context.Context, personID string, chain model.BusinessChain, recordType string, limit int) ([]model.HealthRecordV2, error) {
	query := s.db.WithContext(ctx).Model(&models.HealthRecordV2{}).Where("person_id = ?", personID)
	if chain != "" {
		query = query.Where("business_chain = ?", chain)
	}
	if recordType != "" {
		query = query.Where("record_type = ?", recordType)
	}
	if limit > 0 {
		query = query.Limit(limit)
	}
	query = query.Order("recorded_at DESC")
	var items []models.HealthRecordV2
	if err := query.Find(&items).Error; err != nil {
		return nil, fmt.Errorf("list health records v2: %w", err)
	}
	result := make([]model.HealthRecordV2, len(items))
	for i, r := range items {
		result[i] = model.HealthRecordV2{
			ID: r.ID, PersonID: r.PersonID, BusinessChain: r.BusinessChain,
			RecordType: r.RecordType, Source: r.Source, DeviceID: r.DeviceID,
			RecordedAt: r.RecordedAt, HeartRate: r.HeartRate, BloodPressureSys: r.BloodPressureSys,
			BloodPressureDia: r.BloodPressureDia, SpO2: r.SpO2, Temperature: r.Temperature,
			RespiratoryRate: r.RespiratoryRate, PulseRate: r.PulseRate,
			GlucoseFasting: r.GlucoseFasting, UricAcid: r.UricAcid,
			Steps: r.Steps, SleepHours: r.SleepHours, Notes: r.Content, CreatedAt: r.CreatedAt,
		}
	}
	return result, nil
}

func (s *Store) GetHealthSummaryV2(ctx context.Context, personID string, chain model.BusinessChain) (*model.PersonHealthSummary, error) {
	var rec models.PersonHealthSummary
	if err := s.db.WithContext(ctx).Where("person_id = ? AND business_chain = ?", personID, chain).First(&rec).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &model.PersonHealthSummary{
		PersonID: rec.PersonID, BusinessChain: rec.BusinessChain,
		LatestHR: rec.LatestHR, LatestSpO2: rec.LatestSpO2,
		LatestBPSys: rec.LatestBPSys, LatestBPDia: rec.LatestBPDia,
		LatestGlucoseFasting: rec.LatestGlucoseFasting, LatestUricAcid: rec.LatestUricAcid,
		LatestSteps: rec.LatestSteps, LatestSleepHours: rec.LatestSleepHours,
		RiskScore: rec.RiskScore, TrendDirection: rec.TrendDirection,
		LastUpdated: time.Now(), ARecommendation: rec.Recommendation,
	}, nil
}

func (s *Store) UpdateHealthSummaryV2(ctx context.Context, s2 *model.PersonHealthSummary) error {
	rec := &models.PersonHealthSummary{
		PersonID: s2.PersonID, BusinessChain: s2.BusinessChain,
		LatestHR: s2.LatestHR, LatestSpO2: s2.LatestSpO2,
		LatestBPSys: s2.LatestBPSys, LatestBPDia: s2.LatestBPDia,
		LatestGlucoseFasting: s2.LatestGlucoseFasting, LatestUricAcid: s2.LatestUricAcid,
		LatestSteps: s2.LatestSteps, LatestSleepHours: s2.LatestSleepHours,
		RiskScore: s2.RiskScore, TrendDirection: s2.TrendDirection,
		Recommendation: s2.ARecommendation,
	}
	return s.db.WithContext(ctx).Where("person_id = ? AND business_chain = ?", s2.PersonID, s2.BusinessChain).
		Assign(rec).FirstOrCreate(rec).Error
}
