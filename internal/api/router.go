package api

import (
    "encoding/json"
    "net/http"

		"github.com/prometheus/client_golang/prometheus/promhttp"
		"github.com/Asrlex/schedulerservice/internal/jobs"
)

var jobManager = jobs.NewJobManager()

func NewRouter() *http.ServeMux {
    mux := http.NewServeMux()
    mux.HandleFunc("/healthz", healthHandler)
		mux.HandleFunc("/jobs/register", jobRegisterHandler)
		mux.HandleFunc("/jobs/list", jobListHandler)
    mux.Handle("/metrics", promhttp.Handler())
    return loggingMiddleware(mux).(*http.ServeMux)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
        http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{
        "status": "ok",
        "service": "schedulerservice",
    })
}

func jobRegisterHandler(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
        http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
        return
    }

		var job jobs.Job
		if err := json.NewDecoder(r.Body).Decode(&job); err != nil {
				http.Error(w, "invalid request body", http.StatusBadRequest)
				return
		}

		if err := jobManager.Register(job); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(map[string]string{"status": "registered"})
}

func jobListHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(jobManager.List())
}
