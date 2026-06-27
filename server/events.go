package server

type EventSnapshotProcessed struct {
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
