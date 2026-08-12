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

func (s *Store) PushOTAJob(ctx context.Context, firmwareID string, deviceIDs []string) error {
	devicesJSON, _ := json.Marshal(deviceIDs)
	job := &models.OTAJob{FirmwareID: firmwareID, TargetDevices: string(devicesJSON), Progress: "{}"}
	return s.db.WithContext(ctx).Create(job).Error
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
