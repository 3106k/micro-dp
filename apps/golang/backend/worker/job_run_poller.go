package worker

import (
	"context"
	"log"
	"time"

	"github.com/user/micro-dp/domain"
	"github.com/user/micro-dp/internal/observability"
)

type JobRunPoller struct {
	jobRuns  domain.JobRunRepository
	queue    domain.JobRunQueue
	metrics  *observability.JobRunMetrics
	interval time.Duration
}

func NewJobRunPoller(
	jobRuns domain.JobRunRepository,
	queue domain.JobRunQueue,
	metrics *observability.JobRunMetrics,
	interval time.Duration,
) *JobRunPoller {
	return &JobRunPoller{
		jobRuns:  jobRuns,
		queue:    queue,
		metrics:  metrics,
		interval: interval,
	}
}

func (p *JobRunPoller) Run(ctx context.Context) {
	log.Println("job_run_poller started")

	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("job_run_poller stopped")
			return
		case <-ticker.C:
			p.poll(ctx)
		}
	}
}

func (p *JobRunPoller) poll(ctx context.Context) {
	runs, err := p.jobRuns.ListReady(ctx)
	if err != nil {
		log.Printf("job_run_poller: list ready error: %v", err)
		return
	}

	for _, jr := range runs {
		msg := &domain.JobRunMessage{
			JobRunID: jr.ID,
			TenantID: jr.TenantID,
		}
		if err := p.queue.Enqueue(ctx, msg); err != nil {
			log.Printf("job_run_poller: enqueue error job_run_id=%s: %v", jr.ID, err)
			continue
		}
		p.metrics.DispatchedTotal.Add(ctx, 1)
		log.Printf("job_run_poller: enqueued job_run_id=%s", jr.ID)
	}
}
