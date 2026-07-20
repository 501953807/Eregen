package validation

import (
	"testing"
)

func TestDeviceID_Valid(t *testing.T) {
	validIDs := []string{"BR-0001", "PX-ABCD", "BR-FFFF", "PX-0000"}
	for _, id := range validIDs {
		if err := DeviceID(id); err != nil {
			t.Errorf("DeviceID(%q) unexpected error: %v", id, err)
		}
	}
}

func TestDeviceID_Invalid(t *testing.T) {
	invalidIDs := []string{
		"",
		"XX-0001",
		"BR-000",
		"BR-000G",
		"BR-00001",
		"bracelet-0001",
	}
	for _, id := range invalidIDs {
		if err := DeviceID(id); err == nil {
			t.Errorf("DeviceID(%q) expected error", id)
		}
	}
}

func TestName_Valid(t *testing.T) {
	if err := Name("Alice", 50); err != nil {
		t.Errorf("Name('Alice') unexpected error: %v", err)
	}
}

func TestName_Empty(t *testing.T) {
	if err := Name("", 50); err == nil {
		t.Error("Name('') expected error")
	}
}

func TestName_TooLong(t *testing.T) {
	longName := "A"
	for len(longName) < 100 {
		longName += "A"
	}
	if err := Name(longName, 50); err == nil {
		t.Error("Name(long) expected error")
	}
}

func TestElderlyName_Valid(t *testing.T) {
	if err := ElderlyName("Grandma Wang"); err != nil {
		t.Errorf("ElderlyName unexpected error: %v", err)
	}
}

func TestUserName_Valid(t *testing.T) {
	if err := UserName("John Doe"); err != nil {
		t.Errorf("UserName unexpected error: %v", err)
	}
}

func TestPhone_Valid(t *testing.T) {
	validPhones := []string{"13800138000", "15900001111", "18612345678"}
	for _, phone := range validPhones {
		if err := Phone(phone); err != nil {
			t.Errorf("Phone(%q) unexpected error: %v", phone, err)
		}
	}
}

func TestPhone_Invalid(t *testing.T) {
	invalidPhones := []string{
		"",
		"12345678901",
		"1380013800",
		"138001380000",
		"03800138000",
	}
	for _, phone := range invalidPhones {
		if err := Phone(phone); err == nil {
			t.Errorf("Phone(%q) expected error", phone)
		}
	}
}

func TestEmail_Valid(t *testing.T) {
	validEmails := []string{"user@example.com", "test.user@domain.org"}
	for _, email := range validEmails {
		if err := Email(email); err != nil {
			t.Errorf("Email(%q) unexpected error: %v", email, err)
		}
	}
}

func TestEmail_Invalid(t *testing.T) {
	invalidEmails := []string{
		"",
		"invalid",
		"@example.com",
		"user@",
	}
	for _, email := range invalidEmails {
		if err := Email(email); err == nil {
			t.Errorf("Email(%q) expected error", email)
		}
	}
}

func TestPassword_Valid(t *testing.T) {
	if err := Password("password123", 8); err != nil {
		t.Errorf("Password() unexpected error: %v", err)
	}
}

func TestPassword_TooShort(t *testing.T) {
	if err := Password("short", 8); err == nil {
		t.Error("Password(short) expected error")
	}
}

func TestStrongPassword_Valid(t *testing.T) {
	if err := StrongPassword("Password123"); err != nil {
		t.Errorf("StrongPassword() unexpected error: %v", err)
	}
}

func TestStrongPassword_TooShort(t *testing.T) {
	if err := StrongPassword("Pass1"); err == nil {
		t.Error("StrongPassword(TooShort) expected error")
	}
}

func TestTimeString_Valid(t *testing.T) {
	validTimes := []string{"08:00", "12:30", "23:59", "00:00"}
	for _, time := range validTimes {
		if err := TimeString(time); err != nil {
			t.Errorf("TimeString(%q) unexpected error: %v", time, err)
		}
	}
}

func TestTimeString_Invalid(t *testing.T) {
	invalidTimes := []string{
		"25:00",
		"12:60",
		"abc",
		"12:00:00",
		"",
	}
	for _, time := range invalidTimes {
		if err := TimeString(time); err == nil {
			t.Errorf("TimeString(%q) expected error", time)
		}
	}
}

func TestDateString_Valid(t *testing.T) {
	if err := DateString("2000-01-01"); err != nil {
		t.Errorf("DateString() unexpected error: %v", err)
	}
}

func TestDateString_Future(t *testing.T) {
	if err := DateString("2099-01-01"); err == nil {
		t.Error("DateString(future) expected error")
	}
}

func TestDateString_InvalidFormat(t *testing.T) {
	if err := DateString("01/01/2000"); err == nil {
		t.Error("DateString(invalid format) expected error")
	}
}

func TestHealthData_HeartRate(t *testing.T) {
	if err := HealthData("hr", 72); err != nil {
		t.Errorf("HealthData(hr, 72) unexpected error: %v", err)
	}
	if err := HealthData("hr", 10); err == nil {
		t.Error("HealthData(hr, 10) expected error")
	}
	if err := HealthData("hr", 350); err == nil {
		t.Error("HealthData(hr, 350) expected error")
	}
}

func TestHealthData_SpO2(t *testing.T) {
	if err := HealthData("spo2", 98); err != nil {
		t.Errorf("HealthData(spo2, 98) unexpected error: %v", err)
	}
	if err := HealthData("spo2", 40); err == nil {
		t.Error("HealthData(spo2, 40) expected error")
	}
	if err := HealthData("spo2", 101); err == nil {
		t.Error("HealthData(spo2, 101) expected error")
	}
}

func TestHealthData_Steps(t *testing.T) {
	if err := HealthData("steps", 10000); err != nil {
		t.Errorf("HealthData(steps, 10000) unexpected error: %v", err)
	}
	if err := HealthData("steps", -1); err == nil {
		t.Error("HealthData(steps, -1) expected error")
	}
	if err := HealthData("steps", 300000); err == nil {
		t.Error("HealthData(steps, 300000) expected error")
	}
}

func TestHealthData_SleepHours(t *testing.T) {
	if err := HealthData("sleep_hours", 8.5); err != nil {
		t.Errorf("HealthData(sleep_hours, 8.5) unexpected error: %v", err)
	}
	if err := HealthData("sleep_hours", -1); err == nil {
		t.Error("HealthData(sleep_hours, -1) expected error")
	}
	if err := HealthData("sleep_hours", 25); err == nil {
		t.Error("HealthData(sleep_hours, 25) expected error")
	}
}

func TestHealthData_BloodPressure(t *testing.T) {
	if err := HealthData("bp_systolic", 120); err != nil {
		t.Errorf("HealthData(bp_systolic, 120) unexpected error: %v", err)
	}
	if err := HealthData("bp_systolic", 30); err == nil {
		t.Error("HealthData(bp_systolic, 30) expected error")
	}
	if err := HealthData("bp_diastolic", 80); err != nil {
		t.Errorf("HealthData(bp_diastolic, 80) unexpected error: %v", err)
	}
	if err := HealthData("bp_diastolic", 10); err == nil {
		t.Error("HealthData(bp_diastolic, 10) expected error")
	}
}

func TestHealthData_UnknownMetric(t *testing.T) {
	if err := HealthData("unknown", 100); err == nil {
		t.Error("HealthData(unknown) expected error")
	}
}

func TestLocation_Valid(t *testing.T) {
	if err := Location(31.2304, 121.4737); err != nil {
		t.Errorf("Location() unexpected error: %v", err)
	}
}

func TestLocation_Invalid(t *testing.T) {
	if err := Location(91, 0); err == nil {
		t.Error("Location(91, 0) expected error")
	}
	if err := Location(-91, 0); err == nil {
		t.Error("Location(-91, 0) expected error")
	}
	if err := Location(0, 181); err == nil {
		t.Error("Location(0, 181) expected error")
	}
	if err := Location(0, -181); err == nil {
		t.Error("Location(0, -181) expected error")
	}
}

func TestGeofence_Valid(t *testing.T) {
	if err := Geofence("Home", 31.2304, 121.4737, 1000); err != nil {
		t.Errorf("Geofence() unexpected error: %v", err)
	}
}

func TestGeofence_RadiusTooSmall(t *testing.T) {
	if err := Geofence("Home", 31.2304, 121.4737, 10); err == nil {
		t.Error("Geofence(radius=10) expected error")
	}
}

func TestGeofence_RadiusTooLarge(t *testing.T) {
	if err := Geofence("Home", 31.2304, 121.4737, 60000); err == nil {
		t.Error("Geofence(radius=60000) expected error")
	}
}

func TestMedication_Valid(t *testing.T) {
	if err := Medication(2, "capsule"); err != nil {
		t.Errorf("Medication() unexpected error: %v", err)
	}
}

func TestMedication_DoseCountOutOfRange(t *testing.T) {
	if err := Medication(0, "capsule"); err == nil {
		t.Error("Medication(dose=0) expected error")
	}
	if err := Medication(21, "capsule"); err == nil {
		t.Error("Medication(dose=21) expected error")
	}
}

func TestMedication_EmptyPillType(t *testing.T) {
	if err := Medication(1, ""); err == nil {
		t.Error("Medication(pillType='') expected error")
	}
}

func TestDaysOfWeek_Valid(t *testing.T) {
	if err := DaysOfWeek([]int{1, 2, 3, 4, 5}); err != nil {
		t.Errorf("DaysOfWeek() unexpected error: %v", err)
	}
}

func TestDaysOfWeek_Empty(t *testing.T) {
	if err := DaysOfWeek(nil); err == nil {
		t.Error("DaysOfWeek(nil) expected error")
	}
}

func TestDaysOfWeek_OutOfRange(t *testing.T) {
	if err := DaysOfWeek([]int{0}); err == nil {
		t.Error("DaysOfWeek([0]) expected error")
	}
	if err := DaysOfWeek([]int{8}); err == nil {
		t.Error("DaysOfWeek([8]) expected error")
	}
}

func TestDaysOfWeek_Duplicate(t *testing.T) {
	if err := DaysOfWeek([]int{1, 1}); err == nil {
		t.Error("DaysOfWeek([1,1]) expected error")
	}
}

func TestOTP_Valid(t *testing.T) {
	if err := OTP("123456"); err != nil {
		t.Errorf("OTP() unexpected error: %v", err)
	}
}

func TestOTP_Invalid(t *testing.T) {
	invalidOTPs := []string{
		"",
		"12345",
		"1234567",
		"abcdef",
		"12345a",
	}
	for _, otp := range invalidOTPs {
		if err := OTP(otp); err == nil {
			t.Errorf("OTP(%q) expected error", otp)
		}
	}
}

func TestAlertType_Valid(t *testing.T) {
	validTypes := []string{"sos", "fall", "med_missed", "device_offline", "geofence_breach"}
	for _, alertType := range validTypes {
		if err := AlertType(alertType); err != nil {
			t.Errorf("AlertType(%q) unexpected error: %v", alertType, err)
		}
	}
}

func TestAlertType_Invalid(t *testing.T) {
	if err := AlertType("invalid"); err == nil {
		t.Error("AlertType('invalid') expected error")
	}
}

func TestHealthTier_Valid(t *testing.T) {
	validTiers := []string{"低风险", "中风险", "高风险", "low", "medium", "high"}
	for _, tier := range validTiers {
		if err := HealthTier(tier); err != nil {
			t.Errorf("HealthTier(%q) unexpected error: %v", tier, err)
		}
	}
}

func TestHealthTier_Invalid(t *testing.T) {
	if err := HealthTier("invalid"); err == nil {
		t.Error("HealthTier('invalid') expected error")
	}
}
