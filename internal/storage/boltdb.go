package storage

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"time"

	"github.com/GusAguilra/LinuxHealthDoctor/internal/baseline"
	"github.com/GusAguilra/LinuxHealthDoctor/internal/core"
	"github.com/GusAguilra/LinuxHealthDoctor/internal/snapshot"
	bolt "go.etcd.io/bbolt"
)

var (
	metricsBucket    = []byte("metrics")
	checkResultsBucket = []byte("check_results")
)

type BoltStore struct {
	db *bolt.DB
}

func NewBoltStore(path string) (*BoltStore, error) {
	db, err := bolt.Open(path, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, fmt.Errorf("open bolt: %w", err)
	}

	s := &BoltStore{db: db}
	if err := s.init(); err != nil {
		db.Close()
		return nil, fmt.Errorf("init bolt: %w", err)
	}
	return s, nil
}

func (s *BoltStore) init() error {
	return s.db.Update(func(tx *bolt.Tx) error {
		for _, b := range [][]byte{metricsBucket, checkResultsBucket} {
			if _, err := tx.CreateBucketIfNotExists(b); err != nil {
				return fmt.Errorf("create bucket %s: %w", string(b), err)
			}
		}
		return nil
	})
}

func (s *BoltStore) WriteMetric(ctx context.Context, m *core.Metric) error {
	key := timeKey(m.Timestamp)
	data, err := json.Marshal(m)
	if err != nil {
		return fmt.Errorf("marshal metric: %w", err)
	}
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(metricsBucket)
		keyWithName := append([]byte(m.Name+":"), key...)
		return b.Put(keyWithName, data)
	})
}

func (s *BoltStore) QueryMetrics(ctx context.Context, name string, from, to time.Time) ([]*core.Metric, error) {
	prefix := []byte(name + ":")
	startKey := append(prefix, timeKey(from)...)
	endKey := append(prefix, timeKey(to)...)

	var metrics []*core.Metric
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(metricsBucket)
		c := b.Cursor()

		for k, v := c.Seek(startKey); k != nil && string(k[:len(prefix)]) == string(prefix); k, v = c.Next() {
			if string(k) > string(endKey) {
				break
			}
			var m core.Metric
			if err := json.Unmarshal(v, &m); err != nil {
				return fmt.Errorf("unmarshal metric: %w", err)
			}
			metrics = append(metrics, &m)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return metrics, nil
}

func (s *BoltStore) LatestMetric(ctx context.Context, name string) (*core.Metric, error) {
	prefix := []byte(name + ":")
	var metric *core.Metric
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(metricsBucket)
		c := b.Cursor()

		k, _ := c.Seek(prefix)
		if k == nil {
			return nil
		}

		var lastVal []byte
		for k, v := c.Seek(prefix); k != nil && string(k[:len(prefix)]) == string(prefix); k, v = c.Next() {
			lastVal = v
		}
		if lastVal == nil {
			return nil
		}
		var m core.Metric
		if err := json.Unmarshal(lastVal, &m); err != nil {
			return fmt.Errorf("unmarshal metric: %w", err)
		}
		metric = &m
		return nil
	})
	if err != nil {
		return nil, err
	}
	return metric, nil
}

func (s *BoltStore) SaveCheckResult(ctx context.Context, result *core.CheckResult) error {
	data, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("marshal check result: %w", err)
	}
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(checkResultsBucket)
		key := []byte(result.ID + ":" + result.Timestamp.Format(time.RFC3339Nano))
		return b.Put(key, data)
	})
}

func (s *BoltStore) QueryCheckResults(ctx context.Context, filter core.ResultFilter) ([]*core.CheckResult, error) {
	var results []*core.CheckResult
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(checkResultsBucket)
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			var r core.CheckResult
			if err := json.Unmarshal(v, &r); err != nil {
				return fmt.Errorf("unmarshal check result: %w", err)
			}
			if matchFilter(&r, filter) {
				results = append(results, &r)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return results, nil
}

func (s *BoltStore) LatestCheckResult(ctx context.Context, checkID string) (*core.CheckResult, error) {
	prefix := []byte(checkID + ":")
	var result *core.CheckResult
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(checkResultsBucket)
		c := b.Cursor()

		var lastVal []byte
		for k, v := c.Seek(prefix); k != nil && string(k[:len(prefix)]) == string(prefix); k, v = c.Next() {
			lastVal = v
		}
		if lastVal == nil {
			return nil
		}
		var r core.CheckResult
		if err := json.Unmarshal(lastVal, &r); err != nil {
			return fmt.Errorf("unmarshal check result: %w", err)
		}
		result = &r
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *BoltStore) SaveBaseline(ctx context.Context, bl *baseline.Baseline) error {
	return core.ErrNotImplemented
}

func (s *BoltStore) GetBaseline(ctx context.Context, id string) (*baseline.Baseline, error) {
	return nil, core.ErrNotImplemented
}

func (s *BoltStore) ListBaselines(ctx context.Context) ([]*baseline.Baseline, error) {
	return nil, core.ErrNotImplemented
}

func (s *BoltStore) DeleteBaseline(ctx context.Context, id string) error {
	return core.ErrNotImplemented
}

func (s *BoltStore) SaveSnapshot(ctx context.Context, snap *snapshot.Snapshot) error {
	return core.ErrNotImplemented
}

func (s *BoltStore) GetSnapshot(ctx context.Context, id string) (*snapshot.Snapshot, error) {
	return nil, core.ErrNotImplemented
}

func (s *BoltStore) ListSnapshots(ctx context.Context) ([]*snapshot.Snapshot, error) {
	return nil, core.ErrNotImplemented
}

func (s *BoltStore) DeleteSnapshot(ctx context.Context, id string) error {
	return core.ErrNotImplemented
}

func (s *BoltStore) SaveEvent(ctx context.Context, event *core.Event) error {
	return core.ErrNotImplemented
}

func (s *BoltStore) QueryEvents(ctx context.Context, filter core.EventFilter) ([]*core.Event, error) {
	return nil, core.ErrNotImplemented
}

func (s *BoltStore) Health(ctx context.Context) error {
	return s.db.View(func(tx *bolt.Tx) error {
		return nil
	})
}

func (s *BoltStore) Close() error {
	return s.db.Close()
}

func timeKey(t time.Time) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(t.UnixNano()))
	return b
}

func matchFilter(r *core.CheckResult, filter core.ResultFilter) bool {
	if len(filter.CheckIDs) > 0 {
		found := false
		for _, id := range filter.CheckIDs {
			if r.ID == id {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	if len(filter.Components) > 0 {
		found := false
		for _, c := range filter.Components {
			if r.Category == c {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	if len(filter.Statuses) > 0 {
		found := false
		for _, st := range filter.Statuses {
			if r.Status == st {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	if filter.Severity > 0 && r.Severity < filter.Severity {
		return false
	}
	if !filter.Since.IsZero() && r.Timestamp.Before(filter.Since) {
		return false
	}
	if !filter.Until.IsZero() && r.Timestamp.After(filter.Until) {
		return false
	}
	return true
}
