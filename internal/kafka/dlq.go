package kafka

import (
    "context"
    "log"
    "os"
    "strconv"
    "time"

    kafka "github.com/segmentio/kafka-go"
)

const maxRetries = 3
const attemptsHeader = "attempts"
const errorHeader = "error-reason"

// sendToDLQ writes the original message and an error reason to a DLQ topic.
// It includes original payload and simple metadata in headers.
func sendToDLQ(ctx context.Context, orig kafka.Message, reason string) error {
    brokers := os.Getenv("KAFKA_BROKERS")
    dlqTopic := os.Getenv("KAFKA_DLQ_TOPIC")
    if dlqTopic == "" {
        log.Printf("[DLQ] no DLQ topic configured; dropping message id/key=%s reason=%s", string(orig.Key), reason)
        return nil
    }

    w := kafka.NewWriter(kafka.WriterConfig{
        Brokers: []string{brokers},
        Topic:   dlqTopic,
    })
    defer w.Close()

    headers := append(orig.Headers,
        kafka.Header{Key: errorHeader, Value: []byte(reason)},
        kafka.Header{Key: "dlq-timestamp", Value: []byte(strconv.FormatInt(time.Now().Unix(), 10))},
    )

    msg := kafka.Message{
        Key:     orig.Key,
        Value:   orig.Value,
        Headers: headers,
    }

    return w.WriteMessages(ctx, msg)
}

// getAttempts reads the attempts header from a kafka.Message.
func getAttempts(msg kafka.Message) int {
    for _, h := range msg.Headers {
        if h.Key == attemptsHeader {
            if v, err := strconv.Atoi(string(h.Value)); err == nil {
                return v
            }
        }
    }
    return 0
}

// setAttempts sets/updates the attempts header.
func setAttempts(msg *kafka.Message, attempts int) {
    seen := false
    for i := range msg.Headers {
        if msg.Headers[i].Key == attemptsHeader {
            msg.Headers[i].Value = []byte(strconv.Itoa(attempts))
            seen = true
            break
        }
    }
    if !seen {
        msg.Headers = append(msg.Headers, kafka.Header{Key: attemptsHeader, Value: []byte(strconv.Itoa(attempts))})
    }
}

// requeueMessage writes the message back to either the original topic or a retry topic.
// You can add delay by sleeping or producing to time-partitioned retry topics.
func requeueMessage(ctx context.Context, orig kafka.Message) error {
    brokers := os.Getenv("KAFKA_BROKERS")
		time.Sleep(10 * time.Second)
    topic := os.Getenv("KAFKA_TOPIC")
    if topic == "" {
        topic = orig.Topic
    }

    w := kafka.NewWriter(kafka.WriterConfig{
        Brokers: []string{brokers},
        Topic:   topic,
    })
    defer w.Close()

    return w.WriteMessages(ctx, orig)
}
