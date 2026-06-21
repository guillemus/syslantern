package db

import (
	"context"
	"database/sql"
	"errors"
	"strings"
)

var ErrDuplicateEmail = errors.New("duplicate email")

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

	team, err := q.CreateTeamQuery(ctx, "My Team")
	if err != nil {
		return User{}, err
	}

	user, err := q.CreateUserQuery(ctx, CreateUserQueryParams{TeamID: team.ID, Email: email, PasswordHash: sql.NullString{String: passwordHash, Valid: true}})
	if err != nil {
		if strings.Contains(err.Error(), "users_email_lower_idx") || strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return User{}, ErrDuplicateEmail
		}
		return User{}, err
	}

	if err := tx.Commit(); err != nil {
		return User{}, err
	}

	return user, nil
}
