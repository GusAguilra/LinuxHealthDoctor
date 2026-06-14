package storage

import (
	"context"
	"time"

	"github.com/linuxhealthdoctor/lhd/internal/baseline"
	"github.com/linuxhealthdoctor/lhd/internal/core"
	"github.com/linuxhealthdoctor/lhd/internal/snapshot"
)

type Store interface {
	SaveCheckResult(ctx context.Context, result *core.CheckResult) error
	QueryCheckResults(ctx context.Context, filter core.ResultFilter) ([]*core.CheckResult, error)
	LatestCheckResult(ctx context.Context, checkID string) (*core.CheckResult, error)

	SaveBaseline(ctx context.Context, baseline *baseline.Baseline) error
	GetBaseline(ctx context.Context, id string) (*baseline.Baseline, error)
	ListBaselines(ctx context.Context) ([]*baseline.Baseline, error)
	DeleteBaseline(ctx context.Context, id string) error

	SaveSnapshot(ctx context.Context, snapshot *snapshot.Snapshot) error
	GetSnapshot(ctx context.Context, id string) (*snapshot.Snapshot, error)
	ListSnapshots(ctx context.Context) ([]*snapshot.Snapshot, error)
	DeleteSnapshot(ctx context.Context, id string) error

	WriteMetric(ctx context.Context, m *core.Metric) error
	QueryMetrics(ctx context.Context, name string, from, to time.Time) ([]*core.Metric, error)
	LatestMetric(ctx context.Context, name string) (*core.Metric, error)

	SaveEvent(ctx context.Context, event *core.Event) error
	QueryEvents(ctx context.Context, filter core.EventFilter) ([]*core.Event, error)

	Health(ctx context.Context) error
	Close() error
}
