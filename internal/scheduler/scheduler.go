package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
)

type Job struct {
	ID       string
	Name     string
	Schedule string
	Func     func(context.Context) error
	Enabled  bool
}

type Scheduler struct {
	mu       sync.Mutex
	cron     *cron.Cron
	jobs     map[string]Job
	running  bool
}

func New() *Scheduler {
	return &Scheduler{
		cron: cron.New(cron.WithParser(cron.NewParser(
			cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor,
		))),
		jobs: make(map[string]Job),
	}
}

func (s *Scheduler) AddJob(job Job) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, err := cron.ParseStandard(job.Schedule); err != nil {
		return fmt.Errorf("invalid cron schedule %q: %w", job.Schedule, err)
	}

	s.jobs[job.ID] = job
	return nil
}

func (s *Scheduler) RemoveJob(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.jobs, id)
}

func (s *Scheduler) Start(ctx context.Context) {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}
	s.running = true

	for _, job := range s.jobs {
		if job.Enabled {
			j := job
			_, _ = s.cron.AddFunc(j.Schedule, func() {
				ctx, cancel := context.WithTimeout(ctx, 30*time.Minute)
				defer cancel()
				_ = j.Func(ctx)
			})
		}
	}
	s.mu.Unlock()

	s.cron.Start()
	<-ctx.Done()
	s.cron.Stop()
}

func (s *Scheduler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cron.Stop()
	s.running = false
}

func (s *Scheduler) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}

func (s *Scheduler) ListJobs() []Job {
	s.mu.Lock()
	defer s.mu.Unlock()
	jobs := make([]Job, 0, len(s.jobs))
	for _, j := range s.jobs {
		jobs = append(jobs, j)
	}
	return jobs
}
