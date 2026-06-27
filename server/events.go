package server

type AgentCreatedEvent struct {
	TeamID  int64
	AgentID string
}

type AgentDeletedEvent struct {
	TeamID  int64
	AgentID string
}
