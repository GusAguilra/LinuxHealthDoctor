package repository

import (
	"context"
	"fmt"

	"github.com/linuxhealthdoctor/lhd/internal/core"
	"github.com/linuxhealthdoctor/lhd/internal/storage"
)

type CheckResultRepository struct {
	store storage.Store
}

func NewCheckResultRepository(store storage.Store) *CheckResultRepository {
	return &CheckResultRepository{store: store}
}

func (r *CheckResultRepository) Save(ctx context.Context, result *core.CheckResult) error {
	if result.ID == "" {
		return fmt.Errorf("check result ID is required")
	}
	return r.store.SaveCheckResult(ctx, result)
}

func (r *CheckResultRepository) FindByID(ctx context.Context, id string) (*core.CheckResult, error) {
	return r.store.LatestCheckResult(ctx, id)
}

func (r *CheckResultRepository) FindByFilter(ctx context.Context, filter core.ResultFilter) ([]*core.CheckResult, error) {
	return r.store.QueryCheckResults(ctx, filter)
}

func (r *CheckResultRepository) LatestByCheckID(ctx context.Context, checkID string) (*core.CheckResult, error) {
	return r.store.LatestCheckResult(ctx, checkID)
}

func (r *CheckResultRepository) FindByComponent(ctx context.Context, component core.Component, limit int) ([]*core.CheckResult, error) {
	filter := core.ResultFilter{
		Components: []core.Component{component},
		Limit:      limit,
	}
	return r.store.QueryCheckResults(ctx, filter)
}

func (r *CheckResultRepository) FindByStatus(ctx context.Context, status core.CheckStatus, limit int) ([]*core.CheckResult, error) {
	filter := core.ResultFilter{
		Statuses: []core.CheckStatus{status},
		Limit:    limit,
	}
	return r.store.QueryCheckResults(ctx, filter)
}

func (r *CheckResultRepository) FindRecent(ctx context.Context, limit int) ([]*core.CheckResult, error) {
	filter := core.ResultFilter{
		Limit: limit,
	}
	return r.store.QueryCheckResults(ctx, filter)
}

func (r *CheckResultRepository) CountByStatus(ctx context.Context) (pass, fail, errs, skip int, e error) {
	results, err := r.store.QueryCheckResults(ctx, core.ResultFilter{})
	if err != nil {
		return 0, 0, 0, 0, err
	}
	for _, res := range results {
		switch res.Status {
		case core.StatusPass:
			pass++
		case core.StatusFail:
			fail++
		case core.StatusError:
			errs++
		case core.StatusSkip:
			skip++
		}
	}
	return
}
