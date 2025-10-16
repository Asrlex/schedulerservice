package metrics

type MetricName string

const (
	TotalJobs         MetricName = "jobs_registered_total"
	ActiveJobs        MetricName = "jobs_active"
	TotalExecutions   MetricName = "jobs_total_executions"
	TotalFailures     MetricName = "jobs_total_failures"
	ExecutionDuration MetricName = "jobs_execution_duration"
)
