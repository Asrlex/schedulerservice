package main

import (
	"log"
	"net/http"

	"schedulerservice/internal/api"
	"schedulerservice/internal/metrics"
	"schedulerservice/internal/jobs"
	"schedulerservice/internal/kafka"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	metrics.Init()
	jm := jobs.GetJobManager()
	kafka.InitKafka(jm)

	router := api.NewRouter()

	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatalf("Could not start server: %s\n", err.Error())
	}
}
