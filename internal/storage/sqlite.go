package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/GusAguilra/LinuxHealthDoctor/internal/baseline"
	"github.com/GusAguilra/LinuxHealthDoctor/internal/core"
	"github.com/GusAguilra/LinuxHealthDoctor/internal/snapshot"

	_ "modernc.org/sqlite"
)

type SQLiteStore struct {
	db *sql.DB
}

func NewSQLiteStore(path string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite", "file:"+path+"?mode=rwc")
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	db.SetMaxOpenConns(1)

	s := &SQLiteStore{db: db}
	if err := s.init(); err != nil {
		db.Close()
		return nil, fmt.Errorf("init sqlite: %w", err)
	}
	return s, nil
}

func (s *SQLiteStore) init() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS check_results (
			id TEXT PRIMARY KEY,
			status INTEGER NOT NULL,
			severity INTEGER NOT NULL DEFAULT 0,
			category TEXT NOT NULL DEFAULT '',
			message TEXT NOT NULL DEFAULT '',
			details TEXT NOT NULL DEFAULT '{}',
			metrics TEXT NOT NULL DEFAULT '{}',
			timestamp DATETIME NOT NULL,
			duration_ns INTEGER NOT NULL DEFAULT 0,
			check_error TEXT NOT NULL DEFAULT '',
			remediation TEXT NOT NULL DEFAULT '[]',
			evidence TEXT NOT NULL DEFAULT '[]'
		)`,
		`CREATE INDEX IF NOT EXISTS idx_check_results_timestamp ON check_results(timestamp)`,
		`CREATE INDEX IF NOT EXISTS idx_check_results_category ON check_results(category)`,
		`CREATE INDEX IF NOT EXISTS idx_check_results_status ON check_results(status)`,

		`CREATE TABLE IF NOT EXISTS baselines (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL DEFAULT '',
			data TEXT NOT NULL DEFAULT '{}',
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL
		)`,

		`CREATE TABLE IF NOT EXISTS snapshots (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL DEFAULT '',
			data TEXT NOT NULL DEFAULT '{}',
			created_at DATETIME NOT NULL
		)`,

		`CREATE TABLE IF NOT EXISTS events (
			id TEXT PRIMARY KEY,
			type TEXT NOT NULL,
			source TEXT NOT NULL DEFAULT '',
			severity INTEGER NOT NULL DEFAULT 0,
			message TEXT NOT NULL DEFAULT '',
			data TEXT NOT NULL DEFAULT '{}',
			timestamp DATETIME NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_events_type ON events(type)`,
		`CREATE INDEX IF NOT EXISTS idx_events_timestamp ON events(timestamp)`,

		`CREATE TABLE IF NOT EXISTS metrics (
			name TEXT NOT NULL,
			value REAL NOT NULL,
			unit TEXT NOT NULL DEFAULT '',
			timestamp DATETIME NOT NULL,
			labels TEXT NOT NULL DEFAULT '{}'
		)`,
		`CREATE INDEX IF NOT EXISTS idx_metrics_name ON metrics(name)`,
		`CREATE INDEX IF NOT EXISTS idx_metrics_timestamp ON metrics(timestamp)`,
	}

	for _, q := range queries {
		if _, err := s.db.Exec(q); err != nil {
			return fmt.Errorf("exec %q: %w", q[:60], err)
		}
	}
	return nil
}

func (s *SQLiteStore) SaveCheckResult(ctx context.Context, result *core.CheckResult) error {
	details, err := json.Marshal(result.Details)
	if err != nil {
		return fmt.Errorf("marshal details: %w", err)
	}
	metrics, err := json.Marshal(result.Metrics)
	if err != nil {
		return fmt.Errorf("marshal metrics: %w", err)
	}
	remediation, err := json.Marshal(result.Remediation)
	if err != nil {
		return fmt.Errorf("marshal remediation: %w", err)
	}
	evidence, err := json.Marshal(result.Evidence)
	if err != nil {
		return fmt.Errorf("marshal evidence: %w", err)
	}

	var checkErr string
	if result.Error != nil {
		checkErr = result.Error.Error()
	}

	_, err = s.db.ExecContext(ctx, `
		INSERT INTO check_results (id, status, severity, category, message, details, metrics, timestamp, duration_ns, check_error, remediation, evidence)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			status=excluded.status,
			severity=excluded.severity,
			category=excluded.category,
			message=excluded.message,
			details=excluded.details,
			metrics=excluded.metrics,
			timestamp=excluded.timestamp,
			duration_ns=excluded.duration_ns,
			check_error=excluded.check_error,
			remediation=excluded.remediation,
			evidence=excluded.evidence
	`, result.ID, result.Status, result.Severity, result.Category, result.Message,
		string(details), string(metrics), result.Timestamp.UTC(),
		result.Duration.Nanoseconds(), checkErr,
		string(remediation), string(evidence))
	if err != nil {
		return fmt.Errorf("save check result: %w", err)
	}
	return nil
}

func (s *SQLiteStore) QueryCheckResults(ctx context.Context, filter core.ResultFilter) ([]*core.CheckResult, error) {
	q := "SELECT id, status, severity, category, message, details, metrics, timestamp, duration_ns, check_error, remediation, evidence FROM check_results WHERE 1=1"
	args := []interface{}{}

	if len(filter.CheckIDs) > 0 {
		placeholders := make([]string, len(filter.CheckIDs))
		for i, id := range filter.CheckIDs {
			placeholders[i] = "?"
			args = append(args, id)
		}
		q += " AND id IN (" + join(placeholders, ",") + ")"
	}
	if len(filter.Components) > 0 {
		placeholders := make([]string, len(filter.Components))
		for i, c := range filter.Components {
			placeholders[i] = "?"
			args = append(args, string(c))
		}
		q += " AND category IN (" + join(placeholders, ",") + ")"
	}
	if len(filter.Statuses) > 0 {
		placeholders := make([]string, len(filter.Statuses))
		for i, st := range filter.Statuses {
			placeholders[i] = "?"
			args = append(args, int(st))
		}
		q += " AND status IN (" + join(placeholders, ",") + ")"
	}
	if filter.Severity > 0 {
		q += " AND severity >= ?"
		args = append(args, int(filter.Severity))
	}
	if !filter.Since.IsZero() {
		q += " AND timestamp >= ?"
		args = append(args, filter.Since.UTC())
	}
	if !filter.Until.IsZero() {
		q += " AND timestamp <= ?"
		args = append(args, filter.Until.UTC())
	}
	q += " ORDER BY timestamp DESC"
	if filter.Limit > 0 {
		q += " LIMIT ?"
		args = append(args, filter.Limit)
	}
	if filter.Offset > 0 {
		q += " OFFSET ?"
		args = append(args, filter.Offset)
	}

	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("query check results: %w", err)
	}
	defer rows.Close()

	var results []*core.CheckResult
	for rows.Next() {
		r, err := scanCheckResult(rows)
		if err != nil {
			return nil, err
		}
		results = append(results, r)
	}
	return results, rows.Err()
}

func (s *SQLiteStore) LatestCheckResult(ctx context.Context, checkID string) (*core.CheckResult, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, status, severity, category, message, details, metrics, timestamp, duration_ns, check_error, remediation, evidence
		FROM check_results WHERE id = ? ORDER BY timestamp DESC LIMIT 1
	`, checkID)

	r, err := scanCheckResult(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("latest check result: %w", err)
	}
	return r, nil
}

func (s *SQLiteStore) SaveBaseline(ctx context.Context, bl *baseline.Baseline) error {
	data, err := json.Marshal(bl)
	if err != nil {
		return fmt.Errorf("marshal baseline: %w", err)
	}
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO baselines (id, name, data, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			name=excluded.name,
			data=excluded.data,
			updated_at=excluded.updated_at
	`, bl.ID, bl.Name, string(data), bl.CreatedAt.UTC(), bl.UpdatedAt.UTC())
	if err != nil {
		return fmt.Errorf("save baseline: %w", err)
	}
	return nil
}

func (s *SQLiteStore) GetBaseline(ctx context.Context, id string) (*baseline.Baseline, error) {
	row := s.db.QueryRowContext(ctx, "SELECT data FROM baselines WHERE id = ?", id)
	var data string
	if err := row.Scan(&data); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get baseline: %w", err)
	}
	var bl baseline.Baseline
	if err := json.Unmarshal([]byte(data), &bl); err != nil {
		return nil, fmt.Errorf("unmarshal baseline: %w", err)
	}
	return &bl, nil
}

func (s *SQLiteStore) ListBaselines(ctx context.Context) ([]*baseline.Baseline, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT data FROM baselines ORDER BY created_at DESC")
	if err != nil {
		return nil, fmt.Errorf("list baselines: %w", err)
	}
	defer rows.Close()

	var baselines []*baseline.Baseline
	for rows.Next() {
		var data string
		if err := rows.Scan(&data); err != nil {
			return nil, fmt.Errorf("scan baseline: %w", err)
		}
		var bl baseline.Baseline
		if err := json.Unmarshal([]byte(data), &bl); err != nil {
			return nil, fmt.Errorf("unmarshal baseline: %w", err)
		}
		baselines = append(baselines, &bl)
	}
	return baselines, rows.Err()
}

func (s *SQLiteStore) DeleteBaseline(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM baselines WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete baseline: %w", err)
	}
	return nil
}

func (s *SQLiteStore) SaveSnapshot(ctx context.Context, snap *snapshot.Snapshot) error {
	data, err := json.Marshal(snap)
	if err != nil {
		return fmt.Errorf("marshal snapshot: %w", err)
	}
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO snapshots (id, name, data, created_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			name=excluded.name,
			data=excluded.data
	`, snap.ID, snap.Name, string(data), snap.CreatedAt.UTC())
	if err != nil {
		return fmt.Errorf("save snapshot: %w", err)
	}
	return nil
}

func (s *SQLiteStore) GetSnapshot(ctx context.Context, id string) (*snapshot.Snapshot, error) {
	row := s.db.QueryRowContext(ctx, "SELECT data FROM snapshots WHERE id = ?", id)
	var data string
	if err := row.Scan(&data); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get snapshot: %w", err)
	}
	var snap snapshot.Snapshot
	if err := json.Unmarshal([]byte(data), &snap); err != nil {
		return nil, fmt.Errorf("unmarshal snapshot: %w", err)
	}
	return &snap, nil
}

func (s *SQLiteStore) ListSnapshots(ctx context.Context) ([]*snapshot.Snapshot, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT data FROM snapshots ORDER BY created_at DESC")
	if err != nil {
		return nil, fmt.Errorf("list snapshots: %w", err)
	}
	defer rows.Close()

	var snapshots []*snapshot.Snapshot
	for rows.Next() {
		var data string
		if err := rows.Scan(&data); err != nil {
			return nil, fmt.Errorf("scan snapshot: %w", err)
		}
		var snap snapshot.Snapshot
		if err := json.Unmarshal([]byte(data), &snap); err != nil {
			return nil, fmt.Errorf("unmarshal snapshot: %w", err)
		}
		snapshots = append(snapshots, &snap)
	}
	return snapshots, rows.Err()
}

func (s *SQLiteStore) DeleteSnapshot(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM snapshots WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete snapshot: %w", err)
	}
	return nil
}

func (s *SQLiteStore) WriteMetric(ctx context.Context, m *core.Metric) error {
	labels, err := json.Marshal(m.Labels)
	if err != nil {
		return fmt.Errorf("marshal metric labels: %w", err)
	}
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO metrics (name, value, unit, timestamp, labels)
		VALUES (?, ?, ?, ?, ?)
	`, m.Name, m.Value, m.Unit, m.Timestamp.UTC(), string(labels))
	if err != nil {
		return fmt.Errorf("write metric: %w", err)
	}
	return nil
}

func (s *SQLiteStore) QueryMetrics(ctx context.Context, name string, from, to time.Time) ([]*core.Metric, error) {
	q := "SELECT name, value, unit, timestamp, labels FROM metrics WHERE name = ? AND timestamp >= ? AND timestamp <= ? ORDER BY timestamp ASC"
	rows, err := s.db.QueryContext(ctx, q, name, from.UTC(), to.UTC())
	if err != nil {
		return nil, fmt.Errorf("query metrics: %w", err)
	}
	defer rows.Close()

	var metrics []*core.Metric
	for rows.Next() {
		m, err := scanMetric(rows)
		if err != nil {
			return nil, err
		}
		metrics = append(metrics, m)
	}
	return metrics, rows.Err()
}

func (s *SQLiteStore) LatestMetric(ctx context.Context, name string) (*core.Metric, error) {
	row := s.db.QueryRowContext(ctx, "SELECT name, value, unit, timestamp, labels FROM metrics WHERE name = ? ORDER BY timestamp DESC LIMIT 1", name)
	m, err := scanMetric(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("latest metric: %w", err)
	}
	return m, nil
}

func (s *SQLiteStore) SaveEvent(ctx context.Context, event *core.Event) error {
	data, err := json.Marshal(event.Data)
	if err != nil {
		return fmt.Errorf("marshal event data: %w", err)
	}
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO events (id, type, source, severity, message, data, timestamp)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO NOTHING
	`, event.ID, event.Type, event.Source, event.Severity, event.Message, string(data), event.Timestamp.UTC())
	if err != nil {
		return fmt.Errorf("save event: %w", err)
	}
	return nil
}

func (s *SQLiteStore) QueryEvents(ctx context.Context, filter core.EventFilter) ([]*core.Event, error) {
	q := "SELECT id, type, source, severity, message, data, timestamp FROM events WHERE 1=1"
	args := []interface{}{}

	if len(filter.Types) > 0 {
		placeholders := make([]string, len(filter.Types))
		for i, t := range filter.Types {
			placeholders[i] = "?"
			args = append(args, string(t))
		}
		q += " AND type IN (" + join(placeholders, ",") + ")"
	}
	if filter.Severity > 0 {
		q += " AND severity >= ?"
		args = append(args, int(filter.Severity))
	}
	if !filter.Since.IsZero() {
		q += " AND timestamp >= ?"
		args = append(args, filter.Since.UTC())
	}
	if !filter.Until.IsZero() {
		q += " AND timestamp <= ?"
		args = append(args, filter.Until.UTC())
	}
	q += " ORDER BY timestamp DESC"
	if filter.Limit > 0 {
		q += " LIMIT ?"
		args = append(args, filter.Limit)
	}
	if filter.Offset > 0 {
		q += " OFFSET ?"
		args = append(args, filter.Offset)
	}

	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("query events: %w", err)
	}
	defer rows.Close()

	var events []*core.Event
	for rows.Next() {
		e, err := scanEvent(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, rows.Err()
}

func (s *SQLiteStore) Health(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

func (s *SQLiteStore) Close() error {
	return s.db.Close()
}

func join(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	b := make([]byte, 0, len(strs)*2)
	for i, s := range strs {
		if i > 0 {
			b = append(b, sep...)
		}
		b = append(b, s...)
	}
	return string(b)
}

type scannable interface {
	Scan(dest ...interface{}) error
}

func scanCheckResult(row scannable) (*core.CheckResult, error) {
	var (
		id, message, detailsStr, metricsStr, remediationStr, evidenceStr, checkErr string
		status, severity                                                            int
		category                                                                    string
		timestamp                                                                   time.Time
		durationNs                                                                  int64
	)
	if err := row.Scan(&id, &status, &severity, &category, &message, &detailsStr, &metricsStr, &timestamp, &durationNs, &checkErr, &remediationStr, &evidenceStr); err != nil {
		return nil, err
	}

	r := &core.CheckResult{
		ID:        id,
		Status:    core.CheckStatus(status),
		Severity:  core.Severity(severity),
		Category:  core.Component(category),
		Message:   message,
		Timestamp: timestamp,
		Duration:  time.Duration(durationNs),
	}
	if checkErr != "" {
		r.Error = fmt.Errorf("%s", checkErr)
	}
	json.Unmarshal([]byte(detailsStr), &r.Details)
	json.Unmarshal([]byte(metricsStr), &r.Metrics)
	json.Unmarshal([]byte(remediationStr), &r.Remediation)
	json.Unmarshal([]byte(evidenceStr), &r.Evidence)
	return r, nil
}

func scanMetric(row scannable) (*core.Metric, error) {
	var (
		name, unit, labelsStr string
		value                 float64
		timestamp             time.Time
	)
	if err := row.Scan(&name, &value, &unit, &timestamp, &labelsStr); err != nil {
		return nil, err
	}
	m := &core.Metric{
		Name:      name,
		Value:     value,
		Unit:      unit,
		Timestamp: timestamp,
	}
	json.Unmarshal([]byte(labelsStr), &m.Labels)
	return m, nil
}

func scanEvent(row scannable) (*core.Event, error) {
	var (
		id, etype, source, message, dataStr string
		severity                            int
		timestamp                           time.Time
	)
	if err := row.Scan(&id, &etype, &source, &severity, &message, &dataStr, &timestamp); err != nil {
		return nil, err
	}
	e := &core.Event{
		ID:        id,
		Type:      core.EventType(etype),
		Source:    source,
		Severity:  core.Severity(severity),
		Message:   message,
		Timestamp: timestamp,
	}
	json.Unmarshal([]byte(dataStr), &e.Data)
	return e, nil
}
