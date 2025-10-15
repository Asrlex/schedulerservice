package jobs

import (
	"github.com/robfig/cron/v3"
)

type Job struct {
    Name     string `json:"name"`
    Cron     string `json:"cron"`
    Endpoint string `json:"endpoint"`
}

type JobName struct {
	Name string `json:"name"`
}

type JobResponse struct {
		Status  string `json:"status"`
		Name    string `json:"name"`
		Message string `json:"message"`
		Job     Job    `json:"job"`
}

type JobListItem struct {
		Name     string `json:"name"`
		Cron     cron.Schedule `json:"cron"`
		Endpoint cron.Job `json:"endpoint"`
}

type JobListResponse struct {
		Status  string `json:"status"`
		Name    string `json:"name"`
		Message string `json:"message"`
		Jobs    []JobListItem  `json:"jobs"`
}
