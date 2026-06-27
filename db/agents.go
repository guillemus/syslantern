package db

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
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

func (c *Conn) GetAgentFromAPIKey(
	ctx context.Context, apiKey string,
) (agent Agent, notFound bool, err error) {
	agent, err = c.getAgentFromAPIKey(ctx, apiKey)
	if errors.Is(err, sql.ErrNoRows) {
		return agent, true, nil
	}
	if err != nil {
		return agent, false, err
	}

	return agent, false, nil
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
