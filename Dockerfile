# -------- Go Service Build Stage --------
FROM golang:1.25-alpine AS builder

WORKDIR /app

RUN apk add --no-cache gcc musl-dev

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o go-service ./cmd/main.go

# -------- Runtime Stage --------
FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/schedulerservice .

EXPOSE 8080

CMD ["./schedulerservice"]
