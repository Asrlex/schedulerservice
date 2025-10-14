package main

import (
    "log"
    "net/http"

    "github.com/Asrlex/schedulerservice/internal/api"
    "github.com/Asrlex/schedulerservice/internal/metrics"
)

func main() {
	metrics.Init()
	router := api.NewRouter()

	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatalf("Could not start server: %s\n", err.Error())
	}
}