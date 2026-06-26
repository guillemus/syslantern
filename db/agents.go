package db

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
)

const (
	AgentStatusCreated  AgentStatus = "created"
	AgentStatusRunning  AgentStatus = "running"
	AgentStatusDeleted  AgentStatus = "deleted"
	AgentStatusPaused   AgentStatus = "paused"
	AgentStatusResuming AgentStatus = "resuming"
)

func newAgentID() AgentID {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		panic(err)
	}
	return AgentID(hex.EncodeToString(buf))
}

func newApiKey() AgentAPIKey {
	buf := make([]byte, 24)
	if _, err := rand.Read(buf); err != nil {
		panic(err)
	}
	return AgentAPIKey(hex.EncodeToString(buf))
}

func (c *Conn) CreateAgentForTeam(
	ctx context.Context, teamID TeamID, name string, version string,
) (Agent, error) {
	id := newAgentID()
	return c.upsertAgentForTeam(ctx, upsertAgentForTeamParams{
		ID:      id,
		TeamID:  teamID,
		Name:    name,
		Version: version,
		Status:  AgentStatusCreated,
		ApiKey:  newApiKey(),
	})
}

func (c *Conn) UpdateAgentHostID(ctx context.Context, agentID AgentID, hostID string) error {
	return c.updateAgentHostID(ctx, updateAgentHostIDParams{
		ID:     agentID,
		HostID: sql.NullString{String: hostID, Valid: true},
	})
}
