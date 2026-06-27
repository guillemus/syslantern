package db

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"syslantern/shared"
)

type AgentStatus string

const (
	AgentStatusCreated  AgentStatus = "created"
	AgentStatusRunning  AgentStatus = "running"
	AgentStatusDeleted  AgentStatus = "deleted"
	AgentStatusPaused   AgentStatus = "paused"
	AgentStatusResuming AgentStatus = "resuming"
)

func (s AgentStatus) ToShared() shared.AgentStatus {
	return shared.AgentStatus(s)
}

func newAgentID() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		panic(err)
	}
	return string(hex.EncodeToString(buf))
}

func newApiKey() string {
	buf := make([]byte, 24)
	if _, err := rand.Read(buf); err != nil {
		panic(err)
	}
	return string(hex.EncodeToString(buf))
}

func (c *Conn) CreateAgent(
	ctx context.Context, teamID int64, name string, version string,
) (Agent, error) {
	id := newAgentID()
	return c.upsertAgent(ctx, upsertAgentParams{
		ID:      id,
		TeamID:  teamID,
		Name:    name,
		Version: version,
		Status:  AgentStatusCreated,
		ApiKey:  newApiKey(),
	})
}

type DeleteAgentParams struct {
	ID     string
	TeamID int64
}

func (c *Conn) DeleteAgent(ctx context.Context, arg DeleteAgentParams) error {
	return c.setAgentStatus(ctx, setAgentStatusParams{
		Status: AgentStatusDeleted,
		ID:     arg.ID,
		TeamID: arg.TeamID,
	})
}

func (c *Conn) UpdateAgentHostID(ctx context.Context, agentID string, hostID string) error {
	return c.updateAgentHostID(ctx, updateAgentHostIDParams{
		ID:     agentID,
		HostID: sql.NullString{String: hostID, Valid: true},
	})
}
