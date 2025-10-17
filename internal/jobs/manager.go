package jobs

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"schedulerservice/internal/auth"
	"schedulerservice/internal/db"
	"schedulerservice/internal/metrics"

	"github.com/robfig/cron/v3"
)

var (
	defaultManager *JobManager
	managerOnce    sync.Once
)

// GetJobManager returns the singleton JobManager instance
func GetJobManager() *JobManager {
	managerOnce.Do(func() {
		defaultManager = NewJobManager()
		if err := defaultManager.LoadJobs(); err != nil {
			log.Printf("[ERROR] Failed to load jobs: %v", err)
		}
	})
	return defaultManager
}

// NewJobManager creates a new JobManager instance
func NewJobManager() *JobManager {
	db.InitDBConnection()
	LoadMetricsFromDB()
	c := cron.New()
	c.Start()
	return &JobManager{
		cron: c,
		jobs: make(map[string]cron.EntryID),
	}
}

// LoadJobs loads jobs from the database and registers them
func (jm *JobManager) LoadJobs() error {
	rows, err := db.GetDB().Query("SELECT name, cron, endpoint FROM jobs")
	if err != nil {
		return fmt.Errorf("failed to load jobs: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var job Job
		if err := rows.Scan(&job.Name, &job.Cron, &job.Endpoint); err != nil {
			return fmt.Errorf("failed to scan job: %w", err)
		}

		if err := jm.Register(job); err != nil {
			log.Printf("[WARN] Failed to register job %s: %v", job.Name, err)
		}
	}
	return nil
}

func LoadMetricsFromDB() {
	rows, err := db.GetDB().Query("SELECT metric_name, metric_value FROM metrics")
	if err != nil {
		log.Printf("[ERROR] Failed to load metrics from database: %v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var metricName string
		var metricValue float64
		if err := rows.Scan(&metricName, &metricValue); err != nil {
			log.Printf("[ERROR] Failed to scan metric: %v", err)
			continue
		}

		switch metricName {
		case string(metrics.TotalJobs):
			metrics.JobsRegisteredTotal.Add(metricValue)
		case string(metrics.ActiveJobs):
			metrics.JobsActive.Set(metricValue)
		case string(metrics.TotalExecutions):
			// Assuming you have job-specific labels, you may need to handle this differently
		case string(metrics.TotalFailures):
			// Handle job-specific labels
		case string(metrics.ExecutionDuration):
			// Handle histogram buckets
		default:
			log.Printf("[WARN] Unknown metric %s", metricName)
		}
	}
}

// Register adds a new job to the manager
func (jm *JobManager) Register(job Job) error {
	jm.mu.Lock()
	defer jm.mu.Unlock()

	if _, exists := jm.jobs[job.Name]; exists {
		return fmt.Errorf("job %q already exists", job.Name)
	}

	id, dbErr := jm.cron.AddFunc(job.Cron, func() {
		start := time.Now()
		log.Printf("[JOB] Executing %s -> %s", job.Name, job.Endpoint)
		if err := handleJobRequest(job); err != nil {
			log.Printf("[ERROR] Failed to execute job %s: %v", job.Name, err)
			metrics.JobFailures.WithLabelValues(job.Name).Inc()
			db.UpdateMetric(metrics.TotalFailures, 1, job.Name)
			return
		}
		duration := time.Since(start).Seconds()
		metrics.JobExecutions.WithLabelValues(job.Name).Inc()
		metrics.JobDuration.WithLabelValues(job.Name).Observe(duration)
		db.UpdateGlobalMetric(metrics.TotalExecutions, 1)
		db.UpdateGlobalMetric(metrics.ExecutionDuration, duration)
	})

	if dbErr != nil {
		return fmt.Errorf("invalid cron: %w", dbErr)
	}

	_, dbErr = db.GetDB().Exec(
		"INSERT INTO jobs (name, cron, endpoint) VALUES (?, ?, ?)",
		job.Name, job.Cron, job.Endpoint,
	)
	if dbErr != nil {
		jm.cron.Remove(id)
		return fmt.Errorf("failed to save job in database: %w", dbErr)
	}

	jm.jobs[job.Name] = id
	metrics.JobsRegisteredTotal.Inc()
	metrics.JobsActive.Inc()
	db.UpdateGlobalMetric(metrics.TotalJobs, 1)
	db.UpdateGlobalMetric(metrics.ActiveJobs, 1)
	log.Printf("[JOB] Registered %s (%s)", job.Name, job.Cron)
	return nil
}

// handleJobRequest makes the HTTP request for the job
func handleJobRequest(job Job) error {
	req, callErr := http.NewRequest(http.MethodGet, job.Endpoint, nil)
	if callErr != nil {
		return callErr
	}
	auth.AddAPIKeyToRequest(req)

	client := &http.Client{}
	resp, callErr := client.Do(req)
	if callErr != nil {
		return callErr
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Printf("[ERROR] Job %s returned non-2xx status: %d", job.Name, resp.StatusCode)
		metrics.JobFailures.WithLabelValues(job.Name).Inc()
		db.UpdateMetric(metrics.TotalFailures, 1, job.Name)
		return callErr
	}

	return nil
}

// Deregister removes a job from the manager
func (jm *JobManager) Deregister(name string) error {
	jm.mu.Lock()
	defer jm.mu.Unlock()

	id, exists := jm.jobs[name]
	if !exists {
		return fmt.Errorf("job %q does not exist", name)
	}

	_, dbErr := db.GetDB().Exec("DELETE FROM jobs WHERE name = ?", name)
	if dbErr != nil {
		return fmt.Errorf("failed to delete job from database: %w", dbErr)
	}
	db.UpdateGlobalMetric(metrics.ActiveJobs, -1)

	jm.cron.Remove(id)
	delete(jm.jobs, name)
	metrics.JobsActive.Dec()
	log.Printf("[JOB] Deregistered %s", name)
	return nil
}

// List returns a list of all registered jobs
func (jm *JobManager) List() []JobListItem {
	jm.mu.Lock()
	defer jm.mu.Unlock()

	jobs := make([]JobListItem, 0, len(jm.jobs))
	for name, id := range jm.jobs {
		jobs = append(jobs, JobListItem{
			Name:     name,
			Cron:     jm.cron.Entry(id).Schedule,
			Endpoint: jm.cron.Entry(id).Job.(cron.FuncJob),
		})
	}
	return jobs
}

// ShutDown stops the cron scheduler
func (jm *JobManager) ShutDown() error {
	if defaultManager != nil {
		jm.cron.Stop()
	}
	return nil
}
