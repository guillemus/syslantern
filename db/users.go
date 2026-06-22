package db

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
)

func (c *Conn) GetUserByEmail(ctx context.Context, email string) (User, error) {
	row, err := c.GetUserByEmailQuery(ctx, email)
	return row.User, err
}

func (c *Conn) GetUserByID(ctx context.Context, id int64) (User, error) {
	row, err := c.GetUserByIDQuery(ctx, id)
	return row.User, err
}

func (c *Conn) CreateUserAndTeam(ctx context.Context, email, passwordHash string) (User, error) {
	tx, err := c.DB.BeginTx(ctx, nil)
	if err != nil {
		return User{}, err
	}
	defer tx.Rollback()

	q := c.Queries.WithTx(tx)

	agentAPIKey, err := newAgentAPIKey()
	if err != nil {
		return User{}, err
	}

	team, err := q.CreateTeamQuery(ctx, CreateTeamQueryParams{Name: "My Team", AgentApiKey: agentAPIKey})
	if err != nil {
		return User{}, err
	}

	user, err := q.CreateUserQuery(ctx, CreateUserQueryParams{
		TeamID:       team.ID,
		Email:        email,
		PasswordHash: sql.NullString{String: passwordHash, Valid: true}},
	)
	if err != nil {
		return User{}, err
	}

	if err := tx.Commit(); err != nil {
		return User{}, err
	}

	return user, nil
}

func newAgentAPIKey() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate agent api key: %w", err)
	}
	return "sla_" + base64.RawURLEncoding.EncodeToString(b), nil
}
