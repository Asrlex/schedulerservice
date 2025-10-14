package jobs

type Job struct {
    Name     string `json:"name"`
    Cron     string `json:"cron"`
    Endpoint string `json:"endpoint"`
}
