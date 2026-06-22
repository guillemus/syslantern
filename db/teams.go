package db

import "context"

func (c *Conn) GetTeamByID(ctx context.Context, id TeamID) (Team, error) {
	row, err := c.GetTeamByIDQuery(ctx, id)
	return row.Team, err
}

func (c *Conn) GetTeamByAgentAPIKey(ctx context.Context, agentAPIKey string) (Team, error) {
	row, err := c.GetTeamByAgentAPIKeyQuery(ctx, AgentAPIKey(agentAPIKey))
	return row.Team, err
}
