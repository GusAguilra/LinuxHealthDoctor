package report

import (
	"context"
	"time"

	"github.com/linuxhealthdoctor/lhd/internal/baseline"
	"github.com/linuxhealthdoctor/lhd/internal/core"
	"github.com/linuxhealthdoctor/lhd/internal/snapshot"
)

type ReportData struct {
	Title       string
	GeneratedAt time.Time
	Hostname    string
	Distro      string
	Kernel      string
	Uptime      string

	HealthScore *HealthScore
	Results     *core.AggregatedResult
	Analysis    interface{}
	Baseline    *baseline.Baseline
	Snapshot    *snapshot.Snapshot
	Comparison  *baseline.ComparisonResult

	Sections []ReportSection
}

type ReportSection struct {
	Title   string
	Content string
	Level   int
}

type HealthScore struct {
	Overall     float64
	Category    map[core.Component]float64
	Passed      int
	Failed      int
	Total       int
}

type Engine struct {
	formatters map[string]Formatter
}

type Formatter interface {
	Format(ctx context.Context, data *ReportData, opts *FormatOptions) ([]byte, error)
	Extension() string
	MIMEType() string
}

type FormatOptions struct {
	Color     bool
	Verbose   bool
	Template  string
	IncludeSections []string
	ExcludeSections []string
}

func NewEngine() *Engine {
	return &Engine{
		formatters: make(map[string]Formatter),
	}
}

func (e *Engine) RegisterFormatter(name string, f Formatter) {
	e.formatters[name] = f
}

func (e *Engine) Generate(ctx context.Context, data *ReportData, format string, opts *FormatOptions) ([]byte, error) {
	f, ok := e.formatters[format]
	if !ok {
		return nil, core.ErrNotImplemented
	}
	return f.Format(ctx, data, opts)
}

func (e *Engine) SupportedFormats() []string {
	formats := make([]string, 0, len(e.formatters))
	for name := range e.formatters {
		formats = append(formats, name)
	}
	return formats
}
