package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"

	"sync"
)

var (
	JobsRegisteredTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: string(TotalJobs),
			Help: "Total number of jobs registered",
		},
	)

	JobsActive = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: string(ActiveJobs),
			Help: "Current number of active jobs",
		},
	)

	JobExecutions = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: string(TotalExecutions),
			Help: "Total number of job executions",
		},
		[]string{"job_name"},
	)

	JobFailures = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: string(TotalFailures),
			Help: "Total number of job execution failures",
		},
		[]string{"job_name"},
	)

	JobDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    string(ExecutionDuration),
			Help:    "Job execution time in seconds",
			Buckets: prometheus.LinearBuckets(0.1, 0.5, 10),
		},
		[]string{"job_name"},
	)
)

func Init() {
	sync.OnceFunc(func() {
		prometheus.MustRegister(
			JobsRegisteredTotal,
			JobsActive,
			JobExecutions,
			JobFailures,
			JobDuration,
			Uptime,
			collectors.NewGoCollector(),
			collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		)
	})
}
