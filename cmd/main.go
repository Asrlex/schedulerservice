package main

import (
	"log"
	"net/http"
	"context"
	"os"
	"os/signal"
	"syscall"

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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go kafka.InitKafka(ctx, jm)

	router := api.NewRouter()

	gracefulShutdown(cancel)

	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatalf("Could not start server: %s\n", err.Error())
	}
}

// gracefulShutdown listens for OS signals and cancels the context to allow for graceful shutdown.
func gracefulShutdown(cancel context.CancelFunc) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	go func() {
			<-sig
			log.Println("shutdown signal received, cancelling background work")
			cancel()
	}()
}
