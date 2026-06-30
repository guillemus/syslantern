package server

type SnapshotProcessedType string

const (
	SnapshotProcessedTypeLogs    SnapshotProcessedType = "logs"
	SnapshotProcessedTypeMetrics SnapshotProcessedType = "metrics"
)

type EventSnapshotProcessed struct {
	Type    SnapshotProcessedType
	TeamID  int64
	AgentID string
}

type EventAgentCreated struct {
	TeamID  int64
	AgentID string
}

type EventAgentDeleted struct {
	TeamID  int64
	AgentID string
}
