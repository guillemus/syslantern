package shared

import "time"

type EventBatch struct {
	BatchID string       `json:"batch_id"`
	Agent   BatchAgent   `json:"agent"`
	Host    BatchHost    `json:"host"`
	SentAt  time.Time    `json:"sent_at"`
	Events  []BatchEvent `json:"events"`
}

type BatchAgent struct {
	ID      string `json:"id"`
	Version string `json:"version"`
}

type BatchHost struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	OS   string `json:"os"`
	Arch string `json:"arch"`
}

type BatchEvent struct {
	ID         string        `json:"id"`
	ObservedAt time.Time     `json:"observed_at"`
	Type       string        `json:"type"`
	Source     string        `json:"source"`
	Payload    MetricPayload `json:"payload"`
}

type MetricPayload struct {
	Name   string         `json:"name"`
	Value  float64        `json:"value"`
	Unit   string         `json:"unit"`
	Fields map[string]any `json:"fields"`
}
