package db

import "context"

func (c *Conn) ListAgentsForUser(ctx context.Context, userID int64) ([]Agent, error) {
	rows, err := c.ListAgentsForUserQuery(ctx, userID)
	if err != nil {
		return nil, err
	}

	agents := make([]Agent, 0, len(rows))
	for _, row := range rows {
		agents = append(agents, row.Agent)
	}
	return agents, nil
}
