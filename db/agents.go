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
	return c.UpsertAgentForTeamQuery(ctx, UpsertAgentForTeamQueryParams{
		ID:      id,
		TeamID:  teamID,
		Name:    name,
		Version: version,
		Status:  AgentStatusCreated,
		ApiKey:  newApiKey(),
	})
}

func (c *Conn) ListAgentsForTeam(ctx context.Context, teamID TeamID) ([]Agent, error) {
	rows, err := c.ListAgentsForTeamQuery(ctx, teamID)
	if err != nil {
		return nil, err
	}

	agents := make([]Agent, 0, len(rows))
	for _, row := range rows {
		agents = append(agents, row.Agent)
	}
	return agents, nil
}

func (c *Conn) GetAgentByApiKey(ctx context.Context, apikey string) (Agent, error) {
	row, err := c.GetAgentByAPIKeyQuery(ctx, AgentAPIKey(apikey))
	if err != nil {
		return Agent{}, err
	}
	return row.Agent, nil
}

func (c *Conn) UpdateAgentHostID(ctx context.Context, agentID AgentID, hostID string) error {
	return c.UpdateAgentHostIDQuery(ctx, UpdateAgentHostIDQueryParams{
		ID:     agentID,
		HostID: sql.NullString{String: hostID, Valid: true},
	})
}
