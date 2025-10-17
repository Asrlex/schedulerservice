package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"schedulerservice/internal/api"
	"schedulerservice/internal/auth"
	"schedulerservice/internal/db"
	"schedulerservice/internal/jobs"
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

type Service struct {
	Name      string `json:"name"`
	URL       string `json:"url"`
	HealthURL string `json:"health_url"`
}

// registerService registers the service with the service registry
func registerService() error {
	registryEndpoint := os.Getenv("SERVICE_REGISTRY_URL") + "/register"
	req, callErr := http.NewRequest(http.MethodPost, registryEndpoint, nil)
	if callErr != nil {
		return callErr
	}
	auth.AddAPIKeyToRequest(req)
	DefineService(req)
	client := &http.Client{}
	var retries = 3
	resp, callErr := client.Do(req)
	for callErr != nil && retries > 0 {
		log.Printf("Failed to register service, retrying... (%d retries left)", retries)
		retries--
		time.Sleep(10 * time.Second)
		resp, callErr = client.Do(req)
	}
	defer resp.Body.Close()
	return nil
}

// deregisterService deregisters the service from the service registry
func deregisterService() error {
	registryEndpoint := os.Getenv("SERVICE_REGISTRY_URL") + "/deregister"
	req, callErr := http.NewRequest(http.MethodPost, registryEndpoint, nil)
	if callErr != nil {
		return callErr
	}
	auth.AddAPIKeyToRequest(req)
	DefineService(req)
	client := &http.Client{}
	resp, callErr := client.Do(req)
	if callErr != nil {
		return callErr
	}
	defer resp.Body.Close()
	return nil
}

// DefineService defines the service details in the request body
func DefineService(req *http.Request) error {
	var s = &Service{
		Name:      "schedulerservice",
		URL:       "schedulerservice:8080",
		HealthURL: "schedulerservice:8080/health",
	}
	var buf bytes.Buffer
	var callErr = json.NewEncoder(&buf).Encode(s)
	if callErr != nil {
		return callErr
	}
	req.Body = io.NopCloser(&buf)
	return nil
}
