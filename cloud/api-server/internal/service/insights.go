package service

import (
	"context"
	"math"
	"strconv"
	"time"

	"eregen.dev/api-server/internal/model"

	"go.uber.org/zap"
)

// InsightEngine analyzes health data and generates personalized recommendations.
type InsightEngine struct {
	store healthStore
	log   *zap.Logger
}

// NewInsightEngine creates a new insight engine.
func NewInsightEngine(store healthStore, log *zap.Logger) *InsightEngine {
	return &InsightEngine{store: store, log: log}
}

// HealthScore represents a composite health score (0-100).
type HealthScore struct {
	Score       int        `json:"score"`
	Grade       string     `json:"grade"` // A/B/C/D/F
	Timestamp   time.Time  `json:"timestamp"`
	DailyTrend  float64    `json:"daily_trend"` // % change from yesterday
	Component   []ScoreComp `json:"components"`
}

type ScoreComp struct {
	Name   string  `json:"name"`
	Score  int     `json:"score"`
	Weight float64 `json:"weight"`
}

// DailyInsight is a personalized daily health insight.
type DailyInsight struct {
	Date      string   `json:"date"`
	Score     HealthScore `json:"score"`
	Alerts    []string `json:"alerts"`
	Suggestions []string `json:"suggestions"`
	Trends    []TrendInfo `json:"trends"`
}

// TrendInfo shows a metric's direction and anomaly status.
type TrendInfo struct {
	Metric    string  `json:"metric"`
	Value     float64 `json:"value"`
	Avg7Day   float64 `json:"avg_7day"`
	Trend     string  `json:"trend"` // up/down/stable
	Anomaly   bool    `json:"anomaly"`
	Threshold string  `json:"threshold,omitempty"`
}

// GetDailyInsight generates a comprehensive daily health report.
func (e *InsightEngine) GetDailyInsight(ctx context.Context, elderlyID string, day time.Time) (*DailyInsight, error) {
	insight := &DailyInsight{
		Date: day.Format("2006-01-02"),
	}

	// Fetch 7-day history for trend analysis
	history, err := e.store.GetHealthHistory(ctx, elderlyID, 7)
	if err != nil {
		return nil, err
	}

	// Calculate composite score
	score := e.calculateHealthScore(history)
	insight.Score = score

	// Detect anomalies and generate suggestions
	e.analyzeTrends(history, insight)

	e.log.Info("generated daily insight",
		zap.String("elderly_id", elderlyID),
		zap.Int("score", score.Score),
		zap.String("grade", score.Grade),
	)

	return insight, nil
}

// calculateHealthScore computes a weighted composite score from all metrics.
func (e *InsightEngine) calculateHealthScore(records []model.HealthRecord) HealthScore {
	now := time.Now()
	today := now.Format("2006-01-02")
	yesterday := now.AddDate(0, 0, -1).Format("2006-01-02")

	var todayRecords, yestRecords []model.HealthRecord
	for _, r := range records {
		day := r.Timestamp.Format("2006-01-02")
		if day == today {
			todayRecords = append(todayRecords, r)
		} else if day == yesterday {
			yestRecords = append(yestRecords, r)
		}
	}

	// Score each component (0-100)
	components := []ScoreComp{
		{Name: "心率", Score: e.scoreHeartRate(todayRecords), Weight: 0.25},
		{Name: "血氧", Score: e.scoreSPO2(todayRecords), Weight: 0.20},
		{Name: "活动量", Score: e.scoreActivity(todayRecords), Weight: 0.20},
		{Name: "睡眠", Score: e.scoreSleep(todayRecords), Weight: 0.20},
		{Name: "血压", Score: e.scoreBloodPressure(todayRecords), Weight: 0.15},
	}

	// Calculate weighted total
	total := 0.0
	for _, c := range components {
		total += float64(c.Score) * c.Weight
	}

	// Calculate daily trend
	dailyTrend := 0.0
	if len(yestRecords) > 0 && len(todayRecords) > 0 {
		yestAvg := avgHR(yestRecords)
		todayAvg := avgHR(todayRecords)
		if yestAvg > 0 {
			dailyTrend = ((float64(todayAvg) - float64(yestAvg)) / float64(yestAvg)) * 100
		}
	}

	// Determine grade
	grade := "F"
	score := int(total)
	switch {
	case score >= 90:
		grade = "A+"
	case score >= 80:
		grade = "B"
	case score >= 70:
		grade = "C"
	case score >= 60:
		grade = "D"
	}

	return HealthScore{
		Score:      score,
		Grade:      grade,
		Timestamp:  now,
		DailyTrend: math.Round(dailyTrend*100) / 100,
		Component:  components,
	}
}

// Score ranges for elderly (60-85):
// Heart rate: 60-100 normal, 50-60 or 100-110 mild concern, <50 or >110 alert
func (e *InsightEngine) scoreHeartRate(records []model.HealthRecord) int {
	if len(records) == 0 || records[0].HR == nil {
		return 50 // unknown
	}
	hr := float64(*records[0].HR)
	switch {
	case hr >= 60 && hr <= 100:
		return 100
	case hr >= 50 && hr < 60 || hr > 100 && hr <= 110:
		return 70
	case hr >= 40 && hr < 50 || hr > 110 && hr <= 130:
		return 40
	default:
		return 20
	}
}

// SPO2: >=95% normal, 90-94% mild, <90% alert
func (e *InsightEngine) scoreSPO2(records []model.HealthRecord) int {
	if len(records) == 0 || records[0].SPO2 == nil {
		return 50
	}
	spo2 := float64(*records[0].SPO2)
	switch {
	case spo2 >= 95:
		return 100
	case spo2 >= 90:
		return 65
	case spo2 >= 85:
		return 40
	default:
		return 20
	}
}

// Steps: >6000 active, 3000-6000 moderate, <3000 sedentary
func (e *InsightEngine) scoreActivity(records []model.HealthRecord) int {
	if len(records) == 0 || records[0].Steps == nil {
		return 50
	}
	steps := float64(*records[0].Steps)
	switch {
	case steps >= 6000:
		return 100
	case steps >= 3000:
		return 70
	case steps >= 1000:
		return 40
	default:
		return 20
	}
}

// Sleep: 7-9h normal, 6-7 or 9-10 mild, <6 or >10 alert
func (e *InsightEngine) scoreSleep(records []model.HealthRecord) int {
	if len(records) == 0 || records[0].SleepHours == nil {
		return 50
	}
	hours := *records[0].SleepHours
	switch {
	case hours >= 7 && hours <= 9:
		return 100
	case hours >= 6 && hours < 7 || hours > 9 && hours <= 10:
		return 70
	case hours >= 5 && hours < 6 || hours > 10 && hours <= 12:
		return 40
	default:
		return 20
	}
}

// BP: systolic <140 normal-ish for elderly, 140-160 mild, >160 alert
func (e *InsightEngine) scoreBloodPressure(records []model.HealthRecord) int {
	if len(records) == 0 || records[0].BPSystolic == nil {
		return 50
	}
	sys := float64(*records[0].BPSystolic)
	switch {
	case sys < 140:
		return 100
	case sys < 160:
		return 60
	case sys < 180:
		return 35
	default:
		return 15
	}
}

func avgHR(records []model.HealthRecord) int {
	if len(records) == 0 || records[0].HR == nil {
		return 0
	}
	sum := 0
	count := 0
	for _, r := range records {
		if r.HR != nil {
			sum += *r.HR
			count++
		}
	}
	if count == 0 {
		return 0
	}
	return sum / count
}

// analyzeTrends detects anomalies and generates suggestions.
func (e *InsightEngine) analyzeTrends(records []model.HealthRecord, insight *DailyInsight) {
	// Group by date for averages
	dailyAvgs := make(map[string]model.HealthRecord)
	for _, r := range records {
		day := r.Timestamp.Format("2006-01-02")
		dailyAvgs[day] = r
	}

	days := []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}
	now := time.Now()
	for i := 6; i >= 0; i-- {
		day := now.AddDate(0, 0, -i)
		dayStr := day.Format("2006-01-02")
		shortDay := days[day.Weekday()]

		if rec, ok := dailyAvgs[dayStr]; ok {
			// Heart rate trend
			if rec.HR != nil {
				trend := e.calcTrend(dailyAvgs, func(r model.HealthRecord) float64 {
					if r.HR != nil {
						return float64(*r.HR)
					}
					return 0
				})
				info := TrendInfo{Metric: shortDay + " 心率", Value: float64(*rec.HR), Avg7Day: trend.avg, Trend: trend.dir}
				if *rec.HR < 50 || *rec.HR > 110 {
					info.Anomaly = true
					info.Threshold = "<50 或 >110 bpm"
					insight.Alerts = append(insight.Alerts, "⚠️ "+shortDay+" 心率异常: "+strconv.Itoa(*rec.HR)+" bpm")
				}
				insight.Trends = append(insight.Trends, info)
			}

			// SPO2 trend
			if rec.SPO2 != nil {
				trend := e.calcTrend(dailyAvgs, func(r model.HealthRecord) float64 {
					if r.SPO2 != nil {
						return float64(*r.SPO2)
					}
					return 0
				})
				info := TrendInfo{Metric: shortDay + " 血氧", Value: float64(*rec.SPO2), Avg7Day: trend.avg, Trend: trend.dir}
				if *rec.SPO2 < 90 {
					info.Anomaly = true
					info.Threshold = "<90%"
					insight.Alerts = append(insight.Alerts, "⚠️ "+shortDay+" 血氧过低: "+strconv.Itoa(*rec.SPO2)+"%")
				}
				insight.Trends = append(insight.Trends, info)
			}

			// Steps trend
			if rec.Steps != nil {
				trend := e.calcTrend(dailyAvgs, func(r model.HealthRecord) float64 {
					if r.Steps != nil {
						return float64(*r.Steps)
					}
					return 0
				})
				info := TrendInfo{Metric: shortDay + " 步数", Value: float64(*rec.Steps), Avg7Day: trend.avg, Trend: trend.dir}
				insight.Trends = append(insight.Trends, info)
			}

			// Sleep trend
			if rec.SleepHours != nil {
				trend := e.calcTrend(dailyAvgs, func(r model.HealthRecord) float64 {
					if r.SleepHours != nil {
						return *r.SleepHours
					}
					return 0
				})
				info := TrendInfo{Metric: shortDay + " 睡眠", Value: *rec.SleepHours, Avg7Day: trend.avg, Trend: trend.dir}
				if *rec.SleepHours < 5 {
					info.Anomaly = true
					info.Threshold = "<5小时"
					insight.Suggestions = append(insight.Suggestions, "⚡ "+shortDay+" 睡眠不足 "+strconv.FormatFloat(*rec.SleepHours, 'f', 1, 6)+"小时，建议咨询医生")
				}
				insight.Trends = append(insight.Trends, info)
			}

			// BP trend
			if rec.BPSystolic != nil {
				trend := e.calcTrend(dailyAvgs, func(r model.HealthRecord) float64 {
					if r.BPSystolic != nil {
						return float64(*r.BPSystolic)
					}
					return 0
				})
				info := TrendInfo{Metric: shortDay + " 收缩压", Value: float64(*rec.BPSystolic), Avg7Day: trend.avg, Trend: trend.dir}
				if *rec.BPSystolic > 160 {
					info.Anomaly = true
					info.Threshold = ">160 mmHg"
					insight.Alerts = append(insight.Alerts, "⚠️ "+shortDay+" 血压偏高: "+strconv.Itoa(*rec.BPSystolic)+"/"+strconv.Itoa(*rec.BPDiastolic)+" mmHg")
				}
				insight.Trends = append(insight.Trends, info)
			}
		}
	}
}

type trendResult struct {
	dir string
	avg float64
}

func (e *InsightEngine) calcTrend(avgs map[string]model.HealthRecord, extract func(model.HealthRecord) float64) trendResult {
	var vals []float64
	for _, r := range avgs {
		v := extract(r)
		if v > 0 {
			vals = append(vals, v)
		}
	}
	if len(vals) < 2 {
		return trendResult{dir: "stable", avg: 0}
	}

	// Simple linear regression slope
	n := float64(len(vals))
	sumX := 0.0
	sumY := 0.0
	for i, v := range vals {
		sumX += float64(i)
		sumY += v
	}
	meanX := sumX / n
	meanY := sumY / n

	num := 0.0
	den := 0.0
	for i, v := range vals {
		dx := float64(i) - meanX
		num += dx * (v - meanY)
		den += dx * dx
	}

	if den == 0 {
		return trendResult{dir: "stable", avg: meanY}
	}

	slope := num / den
	mean := sumY / n

	// Classify trend
	dir := "stable"
	if slope > mean*0.02 {
		dir = "up"
	} else if slope < -mean*0.02 {
		dir = "down"
	}

	return trendResult{dir: dir, avg: mean}
}
