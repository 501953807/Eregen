package store

import (
	"context"
	"eregen.dev/admin-api/internal/model"
	"fmt"
)


// GetDashboardStats returns the top-level metrics for the admin dashboard.
func (s *SqliteStore) GetDashboardStats(ctx context.Context) (*model.DashboardStats, error) {
	var stats model.DashboardStats
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM devices WHERE status='online'`).Scan(&stats.OnlineDevices); err != nil {
		return nil, fmt.Errorf("online devices: %w", err)
	}
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM devices`).Scan(&stats.TotalDevices); err != nil {
		return nil, fmt.Errorf("total devices: %w", err)
	}
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM alerts WHERE status='pending'`).Scan(&stats.ActiveAlerts); err != nil {
		return nil, fmt.Errorf("active alerts: %w", err)
	}
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM alerts WHERE severity='P0' AND status='pending'`).Scan(&stats.P0Alerts); err != nil {
		return nil, fmt.Errorf("p0 alerts: %w", err)
	}
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM alerts WHERE severity='P1' AND status='pending'`).Scan(&stats.P1Alerts); err != nil {
		return nil, fmt.Errorf("p1 alerts: %w", err)
	}
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM alerts WHERE severity='P2' AND status='pending'`).Scan(&stats.P2Alerts); err != nil {
		return nil, fmt.Errorf("p2 alerts: %w", err)
	}
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users`).Scan(&stats.TotalUsers); err != nil {
		return nil, fmt.Errorf("total users: %w", err)
	}
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM subscriptions WHERE status='active'`).Scan(&stats.ActiveSubscriptions); err != nil {
		return nil, fmt.Errorf("active subscriptions: %w", err)
	}
	return &stats, nil
}

// GetSubscriptionStats returns a per-tier subscription count breakdown.
func (s *SqliteStore) GetSubscriptionStats(ctx context.Context) ([]model.SubscriptionStat, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT plan_tier, COUNT(*),
		       ROUND(COUNT(*) * 100.0 / (SELECT COUNT(*) FROM subscriptions), 1)
		FROM subscriptions GROUP BY plan_tier ORDER BY COUNT(*) DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("subscription stats: %w", err)
	}
	defer rows.Close()

	var stats []model.SubscriptionStat
	for rows.Next() {
		var st model.SubscriptionStat
		if err := rows.Scan(&st.Tier, &st.Count, &st.Pct); err != nil {
			return nil, fmt.Errorf("scan subscription stat: %w", err)
		}
		stats = append(stats, st)
	}
	return stats, rows.Err()
}

// GetAlertTrend returns alert counts grouped by date and device type.
func (s *SqliteStore) GetAlertTrend(ctx context.Context, days int) ([]model.AlertTrendPoint, error) {
	query := `SELECT DATE(a.created_at) AS alert_date,
		   SUM(CASE WHEN a.business_chain = 'self' THEN 1 ELSE 0 END) AS bracelet_count,
		   SUM(CASE WHEN a.business_chain = 'hospital' THEN 1 ELSE 0 END) AS pillbox_count
		FROM alerts a
		WHERE a.created_at >= datetime('now', '-' || ? || ' days')
		GROUP BY DATE(a.created_at) ORDER BY alert_date`
	rows, err := s.db.QueryContext(ctx, query, days)
	if err != nil {
		return nil, fmt.Errorf("alert trend: %w", err)
	}
	defer rows.Close()

	var result []model.AlertTrendPoint
	for rows.Next() {
		var p model.AlertTrendPoint
		if err := rows.Scan(&p.Date, &p.BraceletCount, &p.PillboxCount); err != nil {
			return nil, fmt.Errorf("scan alert trend: %w", err)
		}
		result = append(result, p)
	}
	return result, rows.Err()
}

// GetAlertDistribution returns alert counts by type.
func (s *SqliteStore) GetAlertDistribution(ctx context.Context) ([]model.AlertDistributionItem, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT alert_type, COUNT(*) FROM alerts GROUP BY alert_type ORDER BY COUNT(*) DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("alert distribution: %w", err)
	}
	defer rows.Close()

	colors := map[string]string{
		"sos":              "#ff4d4f",
		"fall":             "#fa541c",
		"med_missed":       "#faad14",
		"device_offline":   "#1890ff",
		"geofence_breach":  "#722ed1",
	}

	var result []model.AlertDistributionItem
	for rows.Next() {
		var item model.AlertDistributionItem
		if err := rows.Scan(&item.Name, &item.Value); err != nil {
			return nil, fmt.Errorf("scan alert dist: %w", err)
		}
		item.Color = colors[item.Name]
		result = append(result, item)
	}
	return result, rows.Err()
}

// GetUserGrowth returns new user counts grouped by month.
func (s *SqliteStore) GetUserGrowth(ctx context.Context, months int) ([]model.UserGrowthPoint, error) {
	query := `SELECT strftime('%Y-%m', created_at) AS month,
	       COUNT(*) AS new_users
		FROM users GROUP BY strftime('%Y-%m', created_at)
		ORDER BY month DESC LIMIT ?`
	rows, err := s.db.QueryContext(ctx, query, months)
	if err != nil {
		return nil, fmt.Errorf("user growth: %w", err)
	}
	defer rows.Close()

	var result []model.UserGrowthPoint
	for rows.Next() {
		var p model.UserGrowthPoint
		if err := rows.Scan(&p.Month, &p.NewUsers); err != nil {
			return nil, fmt.Errorf("scan user growth: %w", err)
		}
		result = append(result, p)
	}
	return result, rows.Err()
}
