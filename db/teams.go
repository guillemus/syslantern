package db

import "context"

func (c *Conn) GetTeamByID(ctx context.Context, id int64) (Team, error) {
	row, err := c.GetTeamByIDQuery(ctx, id)
	return row.Team, err
}
