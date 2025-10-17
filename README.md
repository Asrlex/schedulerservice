# Go Scheduler Service

A simple scheduler service written in Go.

## Features

- Job registration and deregistration
- Cron-based scheduling
- Metrics collection with Prometheus
- Health checks
- Graceful shutdown
- SQLite database integration
- Docker support

## Requirements

- Go 1.16 or higher
- SQLite3
- Docker (optional, for containerization)
- Prometheus (optional, for metrics collection)

## Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/Asrlex/schedulerservice.git & cd schedulerservice
   ```

2. Install dependencies:
   ```bash
   go mod tidy
   ```

3. Build the application:
   ```bash
   go build -o schedulerservice
   ```

4. Run the application:
   ```bash
   ./schedulerservice
   ```

## Go Commands

- To initialize a new module:
  ```bash
  go mod init github.com/Asrlex/schedulerservice
  ```
- To download dependencies:
  ```bash
  go mod download
  ```
- To run tests:
  ```bash
  go test ./...
  ```
- To format code:
  ```bash
  go fmt ./...
  ```
- To lint code (requires `golangci-lint`):
  ```bash
  golangci-lint run
  ```
- To install dependencies:
  ```bash
  go mod tidy
  ```

## Testing

You can test the service using the provided commands. Make sure to set the `API_KEY` environment variable before running the tests.
```bash
curl -X POST localhost:8080/jobs/register \
   -H "Content-Type: application/json" \
   -H "X-API-KEY: your-secret-api-key" \
   -d '{"name":"ping","cron":"*/10 * * * * *","endpoint":"http://localhost:3000/ping"}'

curl -X GET localhost:8080/jobs/list \
   -H "X-API-KEY: your-secret-api-key"

curl -X POST localhost:8080/jobs/deregister \
   -H "Content-Type: application/json" \
   -H "X-API-KEY: your-secret-api-key" \
   -d '{"name":"ping"}'
```
