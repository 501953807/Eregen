package store

import (
	"context"
	"fmt"
	"log"
	"time"

	"eregen.dev/pipeline/internal/config"
	"eregen.dev/pipeline/internal/model"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Store wraps PostgreSQL and InfluxDB connections for the pipeline.
type Store struct {
	pgPool       *pgxpool.Pool
	influxClient influxdb2.Client
	queryAPI     api.QueryAPI
	writeAPI     api.WriteAPIBlocking
	cfg          *config.Config
}

// NewStore initializes database connections.
func NewStore(cfg *config.Config) (*Store, error) {
	s := &Store{cfg: cfg}

	pgPool, err := pgxpool.New(context.Background(), cfg.PostgresDSN)
	if err != nil {
		return nil, fmt.Errorf("postgres connect: %w", err)
	}
	if err := pgPool.Ping(context.Background()); err != nil {
		pgPool.Close()
		return nil, fmt.Errorf("postgres ping: %w", err)
	}
	s.pgPool = pgPool

	influxClient := influxdb2.NewClient(cfg.InfluxDBURL, cfg.InfluxDBToken)
	s.influxClient = influxClient
	s.queryAPI = influxClient.QueryAPI(cfg.InfluxDBOrg)
	s.writeAPI = influxClient.WriteAPIBlocking(cfg.InfluxDBOrg, cfg.InfluxDBBucket)

	log.Printf("[store] connected to postgres and influxdb (bucket=%s)", cfg.InfluxDBBucket)
	return s, nil
}

// Close shuts down all connections.
func (s *Store) Close() {
	s.influxClient.Close()
	if s.pgPool != nil {
		s.pgPool.Close()
	}
}

// InsertAnalysisResult saves an anomaly detection result to PostgreSQL.
func (s *Store) InsertAnalysisResult(r *model.AnalysisResult) error {
	query := `INSERT INTO analysis_results (elderly_id, metric, value, baseline, deviation, risk_level, timestamp)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := s.pgPool.Exec(context.Background(), query,
		r.ElderlyID, r.Metric, r.Value, r.Baseline, r.Deviation,
		string(r.RiskLevel), r.Timestamp,
	)
	return err
}

// InsertRiskScore saves the composite risk score.
func (s *Store) InsertRiskScore(r *model.RiskScore) error {
	query := `INSERT INTO risk_scores (elderly_id, composite_score, vitals_deviation,
		medication_adherence, activity_level, sleep_quality, recorded_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := s.pgPool.Exec(context.Background(), query,
		r.ElderlyID, r.CompositeScore, r.VitalsDeviation,
		r.MedicationAdherence, r.ActivityLevel, r.SleepQuality,
		r.RecordedAt,
	)
	return err
}

// InsertLocation stores a GPS location point in InfluxDB.
func (s *Store) InsertLocation(elderlyID string, lat, lon float64) error {
	point := influxdb2.NewPoint(
		"location",
		map[string]string{"elderly_id": elderlyID},
		map[string]interface{}{
			"lat": lat,
			"lon": lon,
		},
		time.Now().UTC(),
	)
	return s.writeAPI.WritePoint(context.Background(), point)
}

// QueryBaseline fetches the rolling average for a metric over N days from InfluxDB.
func (s *Store) QueryBaseline(elderlyID, metric string, days int) (float64, error) {
	from := time.Now().Add(-time.Duration(days) * 24 * time.Hour)
	to := time.Now()

	query := fmt.Sprintf(`
		from(bucket: "%s")
			|> range(start: %s, stop: %s)
			|> filter(fn: (r) => r["_measurement"] == "health")
			|> filter(fn: (r) => r["_field"] == "%s")
			|> filter(fn: (r) => r["elderly_id"] == "%s")
			|> mean()
			|> last()
	`, s.cfg.InfluxDBBucket, from.Format(time.RFC3339), to.Format(time.RFC3339), metric, elderlyID)

	table, err := s.queryAPI.Query(context.Background(), query)
	if err != nil {
		return 0, fmt.Errorf("influx query baseline: %w", err)
	}

	if table.Next() {
		result := table.Record()
		if v, ok := result.Value().(int64); ok {
			return float64(v), nil
		}
	}

	return 0, nil
}

// GetLatestVitalsDeviation returns the latest vitals deviation score (0-100).
func (s *Store) GetLatestVitalsDeviation(elderlyID string, days int) (float64, error) {
	query := `SELECT COALESCE(AVG(deviation), 0) FROM analysis_results
		WHERE elderly_id = $1 AND metric IN ('heart_rate', 'spo2', 'bp_systolic', 'bp_diastolic')
		AND timestamp > now() - interval $2 day
		GROUP BY elderly_id`
	var avg float64
	err := s.pgPool.QueryRow(context.Background(), query, elderlyID, days).Scan(&avg)
	if err != nil {
		return 0, err
	}
	return avg, nil
}

// GetLatestMedAdherence returns medication adherence rate (0-100).
func (s *Store) GetLatestMedAdherence(elderlyID string, days int) (float64, error) {
	query := `SELECT COALESCE(AVG(adherence_rate), 100) FROM medication_adherence
		WHERE elderly_id = $1 AND period_end > now() - interval $2 day
		GROUP BY elderly_id`
	var rate float64
	err := s.pgPool.QueryRow(context.Background(), query, elderlyID, days).Scan(&rate)
	if err != nil {
		return 100, nil
	}
	return rate, nil
}

// GetLatestActivityLevel returns activity level score (0-100).
func (s *Store) GetLatestActivityLevel(elderlyID string, days int) (float64, error) {
	query := `SELECT COALESCE(AVG(CASE WHEN steps > 0 THEN 100 ELSE 0 END), 50)
		FROM health_data WHERE elderly_id = $1 AND timestamp > now() - interval $2 day`
	var level float64
	err := s.pgPool.QueryRow(context.Background(), query, elderlyID, days).Scan(&level)
	if err != nil {
		return 50, nil
	}
	return level, nil
}

// GetLatestSleepQuality returns sleep quality score (0-100).
func (s *Store) GetLatestSleepQuality(elderlyID string, days int) (float64, error) {
	query := `SELECT COALESCE(AVG(CASE WHEN sleep_hours > 0 THEN LEAST(sleep_hours * 10, 100) ELSE 50 END), 50)
		FROM health_data WHERE elderly_id = $1 AND timestamp > now() - interval $2 day`
	var quality float64
	err := s.pgPool.QueryRow(context.Background(), query, elderlyID, days).Scan(&quality)
	if err != nil {
		return 50, nil
	}
	return quality, nil
}
