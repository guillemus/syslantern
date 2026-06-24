package db

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

func (c *Conn) GetUserByEmail(ctx context.Context, email string) (User, error) {
	row, err := c.GetUserByEmailQuery(ctx, email)
	return row.User, err
}

func (c *Conn) GetUserByID(ctx context.Context, id UserID) (User, error) {
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

	team, err := q.CreateTeamQuery(ctx, "My Team")
	if err != nil {
		return User{}, err
	}

	user, err := q.CreateUserQuery(ctx, CreateUserQueryParams{
		TeamID:       team.ID,
		Email:        email,
		PasswordHash: passwordHash,
	})
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
