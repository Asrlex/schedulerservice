package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	startTime = time.Now()

	Uptime = prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: "scheduler_service_uptime_seconds",
			Help: "Time since the scheduler service started",
		},
		func() float64 {
			return time.Since(startTime).Seconds()
		},
	)
)
