package jobs

import (
    "fmt"
    "log"
    "sync"
		"time"

		"github.com/Asrlex/schedulerservice/internal/metrics"

    "github.com/robfig/cron/v3"
)

type JobManager struct {
    mu    sync.Mutex
    cron  *cron.Cron
    jobs  map[string]cron.EntryID
}

func NewJobManager() *JobManager {
    c := cron.New()
    c.Start()
    return &JobManager{
        cron: c,
        jobs: make(map[string]cron.EntryID),
    }
}

func (jm *JobManager) Register(job Job) error {
    jm.mu.Lock()
    defer jm.mu.Unlock()

    if _, exists := jm.jobs[job.Name]; exists {
        return fmt.Errorf("job %q already exists", job.Name)
    }

    id, err := jm.cron.AddFunc(job.Cron, func() {
        start := time.Now()
        log.Printf("[JOB] Executing %s -> %s", job.Name, job.Endpoint)
				metrics.JobExecutions.WithLabelValues(job.Name).Inc()
        duration := time.Since(start).Seconds()
        metrics.JobDuration.WithLabelValues(job.Name).Observe(duration)
    })

    if err != nil {
        return fmt.Errorf("invalid cron: %w", err)
    }

    jm.jobs[job.Name] = id
		metrics.JobsRegistered.Inc()
    log.Printf("[JOB] Registered %s (%s)", job.Name, job.Cron)
    return nil
}

func (jm *JobManager) List() []string {
    jm.mu.Lock()
    defer jm.mu.Unlock()

    names := make([]string, 0, len(jm.jobs))
    for name := range jm.jobs {
        names = append(names, name)
    }
    return names
}
