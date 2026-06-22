package db

import "context"

func (c *Conn) RegisterAgentForTeam(ctx context.Context, teamID TeamID, id AgentID, name string, version string) (Agent, error) {
	return c.UpsertAgentForTeamQuery(ctx, UpsertAgentForTeamQueryParams{
		ID:      id,
		TeamID:  teamID,
		Name:    name,
		Version: version,
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
