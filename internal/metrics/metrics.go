package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
)

var (
		JobsRegistered = prometheus.NewCounter(
        prometheus.CounterOpts{
            Name: "scheduler_jobs_registered_total",
            Help: "Total number of jobs registered",
        },
    )

    JobExecutions = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "scheduler_job_executions_total",
            Help: "Total number of job executions",
        },
        []string{"job_name"},
    )

    JobFailures = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "scheduler_job_failures_total",
            Help: "Total number of job execution failures",
        },
        []string{"job_name"},
    )

		JobDuration = prometheus.NewHistogramVec(
				prometheus.HistogramOpts{
						Name:    "scheduler_job_duration_seconds",
						Help:    "Job execution time in seconds",
						Buckets: prometheus.LinearBuckets(0.1, 0.5, 10),
				},
				[]string{"job_name"},
		)
)

func Init() {
    prometheus.MustRegister(JobsRegistered, JobExecutions, JobFailures, JobDuration)
}
