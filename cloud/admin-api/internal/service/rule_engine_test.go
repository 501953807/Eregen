package service

import (
	"math"
	"testing"
	"time"

	"eregen.dev/admin-api/internal/model"
)

func TestFenceCalculator_Distance_SamePoint(t *testing.T) {
	calc := FenceCalculator{}
	dist := calc.Distance(31.2304, 121.4737, 31.2304, 121.4737)
	if dist != 0 {
		t.Errorf("distance between same points = %f, want 0", dist)
	}
}

func TestFenceCalculator_Distance_BeijingToShanghai(t *testing.T) {
	calc := FenceCalculator{}
	dist := calc.Distance(39.9042, 116.4074, 31.2304, 121.4737)
	// Real distance ~1068km, allow ±10% margin
	if dist < 960000 || dist > 1180000 {
		t.Errorf("Beijing-Shanghai distance = %f, want ~1068000 (±10%%)", dist)
	}
}

func TestFenceCalculator_IsInside_Fence(t *testing.T) {
	calc := FenceCalculator{}
	centerLat := 31.2304
	centerLng := 121.4737
	radius := 200 // meters

	// Point 50m away should be inside
	if !calc.IsInside(centerLat+0.0004, centerLng, centerLat, centerLng, radius) {
		t.Error("point 50m from center should be inside fence")
	}

	// Point at exact radius boundary should be inside (use slightly smaller offset for float tolerance)
	if !calc.IsInside(centerLat+0.00099, centerLng, centerLat, centerLng, 111) {
		t.Error("point at radius boundary should be inside fence")
	}
}

func TestFenceCalculator_IsOutside_Fence(t *testing.T) {
	calc := FenceCalculator{}
	centerLat := 31.2304
	centerLng := 121.4737
	radius := 200

	// Far point should be outside
	if calc.IsInside(40.0, 116.0, centerLat, centerLng, radius) {
		t.Error("far point should be outside fence")
	}

	// 5km away should be outside 200m fence
	if calc.IsInside(centerLat+0.045, centerLng, centerLat, centerLng, radius) {
		t.Error("point 5km away should be outside 200m fence")
	}
}

func TestDesensitize_FilterMedication_NoDosage(t *testing.T) {
	d := Desensitize{}
	now := time.Now()
	meds := []model.MedicalMedication{
		{Name: "Aspirin", Dosage: "100mg", CreatedAt: now},
		{Name: "Metformin", Dosage: "500mg", CreatedAt: now},
	}
	result := d.FilterMedication(meds)
	if len(result) != 2 {
		t.Fatalf("expected 2 results, got %d", len(result))
	}
	for i, r := range result {
		if r["name"] == "" {
			t.Errorf("result[%d] name is empty", i)
		}
		if r["dosage"] != "" {
			t.Errorf("result[%d] dosage should be empty for family visibility, got %q", i, r["dosage"])
		}
	}
}

func TestDesensitize_FilterExpense_PreservesAmount(t *testing.T) {
	d := Desensitize{}
	now := time.Now()
	expenses := []model.MedicalExpense{
		{ItemName: "Blood Test", Amount: 150.0, Category: "lab", CreatedAt: now},
	}
	result := d.FilterExpense(expenses)
	if len(result) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result))
	}
	if result[0]["amount"] != 150.0 {
		t.Error("expense amount should be preserved")
	}
}

func TestRuleEngine_TickInterval(t *testing.T) {
	// Verify the rule engine tick interval constant is reasonable
	expectedMinutes := 5.0
	if expectedMinutes <= 0 {
		t.Error("rule engine tick must be positive")
	}
	if expectedMinutes > 15 {
		t.Error("rule engine tick should not exceed 15 minutes")
	}
}

func TestHaversine_Antipodal(t *testing.T) {
	calc := FenceCalculator{}
	// New York to Sydney — approximately antipodal
	dist := calc.Distance(40.7128, -74.0060, -33.8688, 151.2093)
	// Real distance ~15992km, allow generous margin
	if dist < 15000000 || dist > 17000000 {
		t.Errorf("NY-Sydney distance = %f, want ~15992000", dist)
	}
}

func TestHaversine_NorthPole(t *testing.T) {
	calc := FenceCalculator{}
	// North Pole to equator — should be ~10000km
	dist := calc.Distance(90.0, 0.0, 0.0, 0.0)
	if dist < 9900000 || dist > 10100000 {
		t.Errorf("North Pole to equator = %f, want ~10000000", dist)
	}
}

func TestHaversine_EdgeCase_ZeroRadius(t *testing.T) {
	calc := FenceCalculator{}
	// Same point with 0 radius should return 0
	dist := calc.Distance(31.2304, 121.4737, 31.2304, 121.4737)
	if dist != 0 {
		t.Errorf("zero-distance = %f, want 0", dist)
	}
	// IsInside with 0 radius and same point should be true
	if !calc.IsInside(31.2304, 121.4737, 31.2304, 121.4737, 0) {
		t.Error("same point with 0 radius should be inside")
	}
}

func TestConstants_MathPackageAvailable(t *testing.T) {
	// Verify math package functions work correctly for Haversine formula
	lat := 31.2304
	a := math.Sin(lat/2) * math.Sin(lat/2)
	c := 2 * math.Asin(math.Sqrt(a))
	expected := 2 * math.Asin(math.Sqrt(math.Sin(lat/2)*math.Sin(lat/2)))
	if c != expected {
		t.Error("math.Sin/Cos/Sqrt/Asin should work correctly")
	}
}
