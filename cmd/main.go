package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"schedulerservice/internal/api"
	"schedulerservice/internal/db"
	"schedulerservice/internal/jobs"
	"schedulerservice/internal/auth"
	"schedulerservice/internal/kafka"
	"schedulerservice/internal/metrics"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	go func() {
		registerService()
	}()
	metrics.Init()
	jm := jobs.GetJobManager()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go kafka.InitKafka(ctx, jm)

	router := api.NewRouter()

	gracefulShutdown(cancel, jm)

	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatalf("Could not start server: %s\n", err.Error())
	}
}

// gracefulShutdown listens for OS signals and cancels the context to allow for graceful shutdown.
func gracefulShutdown(cancel context.CancelFunc, jm *jobs.JobManager) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sig)

	go func() {
			<-sig
			log.Println("shutdown signal received, cancelling background work")
			cancel()
			
			if err := jm.ShutDown(); err != nil {
				log.Printf("Error during JobManager shutdown: %v", err)
			}
			if err := db.GetDB().Close(); err != nil {
				log.Printf("Error closing database connection: %v", err)
			}
			if err := deregisterService(); err != nil {
				log.Printf("Error deregistering service: %v", err)
			}

			log.Println("shutdown complete")
			os.Exit(0)
	}()
}

// registerService registers the service with the service registry
func registerService() error {
	registryEndpoint := os.Getenv("SERVICE_REGISTRY_URL") + "/register"
	req, callErr := http.NewRequest(http.MethodGet, registryEndpoint, nil)
	if callErr != nil {
		return callErr
	}
	auth.AddAPIKeyToRequest(req)
	client := &http.Client{}
	resp, callErr := client.Do(req)
	if callErr != nil {
		return callErr
	}
	defer resp.Body.Close()
	return nil
}

// deregisterService deregisters the service from the service registry
func deregisterService() error {
	registryEndpoint := os.Getenv("SERVICE_REGISTRY_URL") + "/deregister"
	req, callErr := http.NewRequest(http.MethodGet, registryEndpoint, nil)
	if callErr != nil {
		return callErr
	}
	auth.AddAPIKeyToRequest(req)
	client := &http.Client{}
	resp, callErr := client.Do(req)
	if callErr != nil {
		return callErr
	}
	defer resp.Body.Close()
	return nil
}
