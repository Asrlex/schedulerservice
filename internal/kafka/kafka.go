package kafka

import (
	"fmt"
	"os"
	"context"
	"log"
	"encoding/json"

	kafka "github.com/segmentio/kafka-go"

	"schedulerservice/internal/jobs"
)

// KafkaInit initializes the Kafka consumer and starts processing messages.
func InitKafka(jr jobs.JobRegistrar) {
	brokers := os.Getenv("KAFKA_BROKERS")
	topic := os.Getenv("KAFKA_TOPIC")
	groupID := os.Getenv("KAFKA_GROUP_ID")
	if brokers == "" || topic == "" || groupID == "" {
		fmt.Println("KAFKA_BROKERS, KAFKA_TOPIC, and KAFKA_GROUP_ID environment variables must be set")
		return
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{brokers},
		Topic:   topic,
		GroupID: groupID,
	})
	defer reader.Close()
	fmt.Println("Kafka consumer initialized")

	for {
		m, err := reader.ReadMessage(context.Background())
		if err != nil {
			fmt.Println("Error reading message:", err)
			continue
		}
		ProcessMessage(m, jr)
	}
}

// ProcessMessage processes a single Kafka message and performs the corresponding job operation.
func ProcessMessage(msg kafka.Message, jr jobs.JobRegistrar) {
	var km KafkaMessage
	if err := json.Unmarshal(msg.Value, &km); err != nil {
			log.Printf("[ERROR] invalid kafka message JSON: %v", err)
			return
	}

	switch km.Type {
		case "REGISTER":
			var job jobs.Job
			if err := json.Unmarshal(km.Payload, &job); err != nil {
				log.Printf("[ERROR] invalid job JSON: %v", err)
				return
			}
			jr.Register(job)
		case "UNREGISTER":
			var name jobs.JobName
			if err := json.Unmarshal(km.Payload, &name); err != nil {
				log.Printf("[ERROR] invalid job name JSON: %v", err)
				return
			}
			jr.Unregister(name.Name)
		default:
			log.Printf("[WARN] unknown kafka message type: %s", km.Type)
	}
}
