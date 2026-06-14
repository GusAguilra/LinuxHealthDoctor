package snapshot

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/linuxhealthdoctor/lhd/internal/core"
)

type Engine struct {
	store Store
}

type Store interface {
	SaveSnapshot(ctx context.Context, s *Snapshot) error
	LoadSnapshot(ctx context.Context, id string) (*Snapshot, error)
	ListSnapshots(ctx context.Context) ([]*Snapshot, error)
	DeleteSnapshot(ctx context.Context, id string) error
}

func NewEngine(store Store) *Engine {
	return &Engine{store: store}
}

type Snapshot struct {
	ID          string
	Name        string
	Timestamp   time.Time
	CreatedAt   time.Time
	Distro      string
	Kernel      string
	Hostname    string
	Categories  map[core.Component]CategoryData
	Manifest    *Manifest
	Size        int64
	Compressed  bool
	Checksum    string
}

type CategoryData struct {
	Component core.Component
	Data      map[string]interface{}
	FileCount int
	TotalSize int64
}

type Manifest struct {
	Files    []FileEntry
	Packages []PackageEntry
	Services []ServiceEntry
}

type FileEntry struct {
	Path   string
	Size   int64
	Mode   os.FileMode
	SHA256 string
}

type PackageEntry struct {
	Name    string
	Version string
}

type ServiceEntry struct {
	Name   string
	Status string
}

func (e *Engine) Create(ctx context.Context, name string, result *core.AggregatedResult) (*Snapshot, error) {
	snap := &Snapshot{
		ID:         fmt.Sprintf("snap-%d", time.Now().UnixNano()),
		Name:       name,
		Timestamp:  time.Now(),
		Categories: make(map[core.Component]CategoryData),
	}

	for component, results := range result.Results {
		cat := CategoryData{
			Component: component,
			Data:      make(map[string]interface{}),
			FileCount: len(results),
		}

		for _, r := range results {
			cat.Data[r.ID] = map[string]interface{}{
				"status":   r.Status.String(),
				"severity": r.Severity.String(),
				"message":  r.Message,
				"metrics":  r.Metrics,
				"details":  r.Details,
			}
			cat.TotalSize += int64(len(r.Message) + len(r.Details))
		}

		snap.Categories[component] = cat
	}

	snap.Manifest = &Manifest{}

	return snap, nil
}
