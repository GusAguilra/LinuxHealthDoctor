package plugin

import (
	"context"

	"github.com/linuxhealthdoctor/lhd/internal/core"
)

type Checker interface {
	ID() string
	Name() string
	Description() string
	Category() core.Component
	Check(ctx context.Context, req *CheckRequest) (*core.CheckResult, error)
	Dependencies() []string
	Tags() []string
}

type CheckRequest struct {
	Baseline  interface{}
	Threshold *core.Threshold
	Options   map[string]interface{}
}
