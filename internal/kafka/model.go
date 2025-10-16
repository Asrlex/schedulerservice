package kafka

import (
	"encoding/json"
)

type KafkaMessage struct {
	Id				string `json:"id"`
	Type 			string `json:"type"`
	Payload 	json.RawMessage `json:"payload"`
	Timestamp 	int64  `json:"timestamp"`
}

