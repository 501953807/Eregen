package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"eregen.dev/api-server/internal/model"
	"eregen.dev/api-server/internal/store"

	"go.uber.org/zap"
)

// ReportPeriod maps report type to days window.
var ReportPeriod = map[string]int{
	"weekly":  7,
	"monthly": 30,
	"annual":  365,
}

// ChronicReportService builds periodic health reports.
type ChronicReportService struct {
	store *store.ChronicStore
	log   *zap.Logger
}

// NewChronicReportService creates a new report service.
func NewChronicReportService(svc *store.ChronicStore, log *zap.Logger) *ChronicReportService {
	return &ChronicReportService{store: svc, log: log}
}

// GenerateReport creates a new chronic health report for the given elderly and report type.
// reportType must be one of: weekly, monthly, annual.
func (s *ChronicReportService) GenerateReport(ctx context.Context, elderlyID, reportType string) (*model.ChronicHealthReport, error) {
	days, ok := ReportPeriod[reportType]
	if !ok {
		return nil, fmt.Errorf("invalid report type %q, expected weekly/monthly/annual", reportType)
	}

	periodEnd := time.Now()
	periodStart := periodEnd.AddDate(0, 0, -days)

	// Fetch raw data for each metric
	glucoseRecords, err := s.store.ListGlucoseRecords(ctx, elderlyID, days)
	if err != nil {
		return nil, fmt.Errorf("fetch glucose records: %w", err)
	}
	uricRecords, err := s.store.ListUricAcidRecords(ctx, elderlyID, days)
	if err != nil {
		return nil, fmt.Errorf("fetch uric acid records: %w", err)
	}
	bpRecords, err := s.store.ListBPRecords(ctx, elderlyID, days)
	if err != nil {
		return nil, fmt.Errorf("fetch BP records: %w", err)
	}

	// Compute stats
	glucoseSummary := computeGlucoseSummary(glucoseRecords)
	uricSummary := computeUricSummary(uricRecords)
	bpSummary := computeBPSummary(bpRecords)

	// Build data_summary
	dataSummary := map[string]any{
		"report_type": reportType,
		"period_start": periodStart.Format("2006-01-02"),
		"period_end":   periodEnd.Format("2006-01-02"),
		"glucose":      glucoseSummary,
		"uric_acid":    uricSummary,
		"blood_pressure": bpSummary,
		"record_counts": map[string]int{
			"glucose": len(glucoseRecords),
			"uric_acid": len(uricRecords),
			"bp": len(bpRecords),
		},
	}
	dataSummaryBytes, err := json.Marshal(dataSummary)
	if err != nil {
		return nil, fmt.Errorf("marshal data_summary: %w", err)
	}

	// Build AI recommendations (rules-based)
	recommendations := buildRecommendations(glucoseSummary, uricSummary, bpSummary)
	recommendationsBytes, err := json.Marshal(recommendations)
	if err != nil {
		return nil, fmt.Errorf("marshal ai_recommendations: %w", err)
	}

	// Persist report
	report := &model.ChronicHealthReport{
		ElderlyID:        elderlyID,
		ReportType:       reportType,
		PeriodStart:      periodStart,
		PeriodEnd:        periodEnd,
		DataSummary:      toStringPtr(string(dataSummaryBytes)),
		AIRecommendations: toStringPtr(string(recommendationsBytes)),
	}
	if err := s.store.SaveHealthReport(ctx, report); err != nil {
		return nil, fmt.Errorf("save health report: %w", err)
	}

	s.log.Info("report generated",
		zap.String("elderly_id", elderlyID),
		zap.String("type", reportType),
		zap.Int("glucose_records", len(glucoseRecords)),
		zap.Int("uric_records", len(uricRecords)),
		zap.Int("bp_records", len(bpRecords)),
	)
	return report, nil
}

// GetReport fetches an existing report by elderly ID and type.
func (s *ChronicReportService) GetReport(ctx context.Context, elderlyID, reportType string) (*model.ChronicHealthReport, error) {
	reports, err := s.store.ListHealthReports(ctx, elderlyID)
	if err != nil {
		return nil, fmt.Errorf("list health reports: %w", err)
	}
	for _, r := range reports {
		if r.ReportType == reportType {
			return &r, nil
		}
	}
	return nil, fmt.Errorf("report not found for type %q", reportType)
}

// ─── helpers ────────────────────────────────────────────────────────────────

func computeGlucoseSummary(records []model.ChronicGlucoseRecord) map[string]any {
	if len(records) == 0 {
		return map[string]any{"avg": 0, "min": 0, "max": 0, "count": 0, "in_range_pct": 0}
	}
	var sum, minV, maxV float64
	inRange := 0
	for _, r := range records {
		sum += r.Value
		if r.Value < minV || sum == r.Value {
			minV = r.Value
		}
		if r.Value > maxV {
			maxV = r.Value
		}
		if r.Value >= 3.9 && r.Value <= 7.8 {
			inRange++
		}
	}
	return map[string]any{
		"avg":          round2(sum / float64(len(records))),
		"min":          minV,
		"max":          maxV,
		"count":        len(records),
		"in_range_pct": round100(float64(inRange) / float64(len(records)) * 100),
	}
}

func computeUricSummary(records []model.ChronicUricAcidRecord) map[string]any {
	if len(records) == 0 {
		return map[string]any{"avg": 0, "min": 0, "max": 0, "count": 0, "high_pct": 0}
	}
	var sum, minV, maxV float64
	high := 0
	for _, r := range records {
		sum += r.Value
		if sum == r.Value || r.Value < minV {
			minV = r.Value
		}
		if r.Value > maxV {
			maxV = r.Value
		}
		if r.Value > 420 { // μmol/L upper limit for men; conservative for elderly
			high++
		}
	}
	return map[string]any{
		"avg":      round2(sum / float64(len(records))),
		"min":      minV,
		"max":      maxV,
		"count":    len(records),
		"high_pct": round100(float64(high) / float64(len(records)) * 100),
	}
}

func computeBPSummary(records []model.ChronicBPRecord) map[string]any {
	if len(records) == 0 {
		return map[string]any{"avg_systolic": 0, "avg_diastolic": 0, "count": 0, "high_systolic_pct": 0}
	}
	var sumSys, sumDia float64
	highSys := 0
	for _, r := range records {
		sumSys += float64(r.Systolic)
		sumDia += float64(r.Diastolic)
		if r.Systolic >= 140 {
			highSys++
		}
	}
	n := float64(len(records))
	return map[string]any{
		"avg_systolic":   int(sumSys / n),
		"avg_diastolic":  int(sumDia / n),
		"count":          len(records),
		"high_systolic_pct": round100(float64(highSys) / n * 100),
	}
}

// buildRecommendations produces a simple list of rule-based advice strings.
func buildRecommendations(glucose, uric, bp map[string]any) []map[string]any {
	var recs []map[string]any
	add := func(level, title, detail string) {
		recs = append(recs, map[string]any{
			"level": level, // "info", "warning", "danger"
			"title": title,
			"detail": detail,
		})
	}

	// Glucose
	if g, ok := glucose["avg"].(float64); ok && g > 0 {
		switch {
		case g > 10.0:
			add("danger", "血糖显著偏高", fmt.Sprintf("近%d次检测平均血糖 %.1f mmol/L，建议尽快就医调整方案", int(glucose["count"].(int)), g))
		case g > 7.8:
			add("warning", "血糖偏高", fmt.Sprintf("平均血糖 %.1f mmol/L，高于推荐范围，建议控制饮食并增加运动", g))
		case g < 3.9:
			add("warning", "血糖偏低", fmt.Sprintf("平均血糖 %.1f mmol/L，注意防范低血糖，随身携带糖果", g))
		default:
			add("info", "血糖控制良好", fmt.Sprintf("平均血糖 %.1f mmol/L，维持在正常范围 (3.9-7.8 mmol/L)", g))
		}
	}

	// Uric acid
	if u, ok := uric["avg"].(float64); ok && u > 0 {
		if u > 420 {
			add("warning", "尿酸偏高", fmt.Sprintf("平均尿酸 %.1f μmol/L，建议减少高嘌呤食物摄入，多饮水", u))
		} else {
			add("info", "尿酸水平正常", fmt.Sprintf("平均尿酸 %.1f μmol/L，保持在理想范围", u))
		}
	}

	// BP
	if sys, ok := bp["avg_systolic"].(int); ok && sys > 0 {
		if sys >= 160 {
			add("danger", "收缩压显著偏高", fmt.Sprintf("平均收缩压 %d mmHg，建议尽快就医", sys))
		} else if sys >= 140 {
			add("warning", "收缩压偏高", fmt.Sprintf("平均收缩压 %d mmHg，建议低盐饮食并监测血压", sys))
		} else if sys >= 130 {
			add("info", "收缩压临界偏高", fmt.Sprintf("平均收缩压 %d mmHg，建议定期监测，控制盐分摄入", sys))
		} else {
			add("info", "血压控制良好", fmt.Sprintf("平均收缩压 %d mmHg，维持在正常范围", sys))
		}
	}

	if len(recs) == 0 {
		add("info", "暂无足够数据", "本报告周期内数据记录不足，请确保按时记录健康指标")
	}
	return recs
}

func round2(v float64) float64 {
	return float64(int(v*100+0.5)) / 100
}

func round100(v float64) float64 {
	return float64(int(v*100+0.5)) / 100
}

func toStringPtr(s string) *string {
	return &s
}
